// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/apache/cloudstack-cloudmonkey/config"
)

const (
	testAPIKey    = "api-key"
	testSecretKey = "secret-key"
)

type signatureAttempt struct {
	Command   string
	Algorithm string
}

type signatureTestServer struct {
	server   *httptest.Server
	attempts []signatureAttempt
	mu       sync.Mutex
}

func newSignatureTestServer(t *testing.T, allowedAlgorithms map[string]bool) *signatureTestServer {
	t.Helper()
	testServer := &signatureTestServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "failed to parse request form", http.StatusBadRequest)
			return
		}

		command := req.Form.Get("command")
		algorithm := identifySignatureAlgorithm(req.Form)
		testServer.mu.Lock()
		testServer.attempts = append(testServer.attempts, signatureAttempt{
			Command:   command,
			Algorithm: algorithm,
		})
		testServer.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if !allowedAlgorithms[algorithm] {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"errorresponse":{"errorcode":401,"cserrorcode":9999,"errortext":"unable to verify request signature"}}`)
			return
		}

		fmt.Fprint(w, successResponse(command))
	}))
	testServer.server = server
	t.Cleanup(server.Close)
	return testServer
}

func identifySignatureAlgorithm(form url.Values) string {
	unsigned := cloneRequestParams(form)
	signature := unsigned.Get("signature")
	unsigned.Del("signature")

	unsignedRequest := encodeRequestParams(unsigned)
	for _, algorithm := range []string{config.SignatureAlgorithmHmacSHA512, config.SignatureAlgorithmHmacSHA1} {
		expected, err := signRequest(unsignedRequest, testSecretKey, algorithm)
		if err != nil {
			return "unknown"
		}
		if signature == expected {
			return algorithm
		}
	}
	return "unknown"
}

func successResponse(command string) string {
	responseKey := strings.ToLower(command) + "response"
	if strings.EqualFold(command, "listApis") {
		return `{"listapisresponse":{"count":0,"api":[]}}`
	}
	return fmt.Sprintf(`{"%s":{"success":true}}`, responseKey)
}

func newTestRequest(t *testing.T, serverURL string, signatureAlgorithm string) *Request {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		Dir:         dir,
		ConfigFile:  filepath.Join(dir, "config"),
		HistoryFile: filepath.Join(dir, "history"),
		Core: &config.Core{
			Prompt:       "cmk",
			AsyncBlock:   true,
			Timeout:      30,
			Output:       config.JSON,
			VerifyCert:   true,
			ProfileName:  "localcloud",
			AutoComplete: true,
			PostRequest:  true,
		},
		ActiveProfile: &config.ServerProfile{
			URL:                serverURL,
			Domain:             "/",
			APIKey:             testAPIKey,
			SecretKey:          testSecretKey,
			SignatureAlgorithm: signatureAlgorithm,
			Client:             http.DefaultClient,
		},
	}
	return NewRequest(GetAPIHandler(), cfg, nil, false)
}

func (s *signatureTestServer) snapshotAttempts() []signatureAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempts := make([]signatureAttempt, len(s.attempts))
	copy(attempts, s.attempts)
	return attempts
}

func TestSignRequestHmacSHA1(t *testing.T) {
	got, err := signRequest("apikey=abc&command=listZones&response=json", "secret", config.SignatureAlgorithmHmacSHA1)
	if err != nil {
		t.Fatal(err)
	}
	if got != "tcMI1Kpm20pLhrrVYtCCcualuBU=" {
		t.Fatalf("HmacSHA1 signature = %q", got)
	}
}

func TestSignRequestHmacSHA512(t *testing.T) {
	got, err := signRequest("apikey=abc&command=listZones&response=json", "secret", config.SignatureAlgorithmHmacSHA512)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/JUUf9SPvNd0sjdbfIxsCxYItNUlGauI+T71cOhHd5fYffHuVAXIba9RzFBK+empezGCzlhw4+R9LFri3CG+oQ==" {
		t.Fatalf("HmacSHA512 signature = %q", got)
	}
}

func TestExplicitHmacSHA512DoesNotTryHmacSHA1(t *testing.T) {
	server := newSignatureTestServer(t, map[string]bool{
		config.SignatureAlgorithmHmacSHA1: true,
	})
	request := newTestRequest(t, server.server.URL, config.SignatureAlgorithmHmacSHA512)

	if _, err := NewAPIRequest(request, "listZones", nil, false); err == nil {
		t.Fatal("expected explicit HmacSHA512 request to fail")
	}

	attempts := server.snapshotAttempts()
	if len(attempts) != 1 {
		t.Fatalf("attempt count = %d, want 1", len(attempts))
	}
	if attempts[0].Algorithm != config.SignatureAlgorithmHmacSHA512 {
		t.Fatalf("attempted algorithm = %q, want %q", attempts[0].Algorithm, config.SignatureAlgorithmHmacSHA512)
	}
}

func TestAutoPersistsHmacSHA512WhenProbeSucceeds(t *testing.T) {
	server := newSignatureTestServer(t, map[string]bool{
		config.SignatureAlgorithmHmacSHA512: true,
	})
	request := newTestRequest(t, server.server.URL, config.SignatureAlgorithmAuto)

	if _, err := NewAPIRequest(request, "listZones", nil, false); err != nil {
		t.Fatal(err)
	}

	if got := request.Config.ActiveProfile.SignatureAlgorithm; got != config.SignatureAlgorithmHmacSHA512 {
		t.Fatalf("persisted signature algorithm = %q, want %q", got, config.SignatureAlgorithmHmacSHA512)
	}
	attempts := server.snapshotAttempts()
	if len(attempts) != 2 {
		t.Fatalf("attempt count = %d, want 2", len(attempts))
	}
	if attempts[0] != (signatureAttempt{Command: "listApis", Algorithm: config.SignatureAlgorithmHmacSHA512}) {
		t.Fatalf("first attempt = %+v, want listApis with HmacSHA512", attempts[0])
	}
	if attempts[1] != (signatureAttempt{Command: "listZones", Algorithm: config.SignatureAlgorithmHmacSHA512}) {
		t.Fatalf("second attempt = %+v, want listZones with HmacSHA512", attempts[1])
	}
}

func TestAutoFallsBackAndPersistsHmacSHA1(t *testing.T) {
	server := newSignatureTestServer(t, map[string]bool{
		config.SignatureAlgorithmHmacSHA1: true,
	})
	request := newTestRequest(t, server.server.URL, config.SignatureAlgorithmAuto)

	if _, err := NewAPIRequest(request, "listZones", nil, false); err != nil {
		t.Fatal(err)
	}

	if got := request.Config.ActiveProfile.SignatureAlgorithm; got != config.SignatureAlgorithmHmacSHA1 {
		t.Fatalf("persisted signature algorithm = %q, want %q", got, config.SignatureAlgorithmHmacSHA1)
	}
	attempts := server.snapshotAttempts()
	if len(attempts) != 3 {
		t.Fatalf("attempt count = %d, want 3", len(attempts))
	}
	if attempts[0] != (signatureAttempt{Command: "listApis", Algorithm: config.SignatureAlgorithmHmacSHA512}) {
		t.Fatalf("first attempt = %+v, want listApis with HmacSHA512", attempts[0])
	}
	if attempts[1] != (signatureAttempt{Command: "listApis", Algorithm: config.SignatureAlgorithmHmacSHA1}) {
		t.Fatalf("second attempt = %+v, want listApis with HmacSHA1", attempts[1])
	}
	if attempts[2] != (signatureAttempt{Command: "listZones", Algorithm: config.SignatureAlgorithmHmacSHA1}) {
		t.Fatalf("third attempt = %+v, want listZones with HmacSHA1", attempts[2])
	}
}

func TestAutoDoesNotRetryUserCommandDirectly(t *testing.T) {
	server := newSignatureTestServer(t, map[string]bool{
		config.SignatureAlgorithmHmacSHA1: true,
	})
	request := newTestRequest(t, server.server.URL, config.SignatureAlgorithmAuto)

	if _, err := NewAPIRequest(request, "deployVirtualMachine", []string{"serviceofferingid=1"}, false); err != nil {
		t.Fatal(err)
	}

	userCommandAttempts := 0
	for _, attempt := range server.snapshotAttempts() {
		if attempt.Command == "deployVirtualMachine" {
			userCommandAttempts++
		}
	}
	if userCommandAttempts != 1 {
		t.Fatalf("deployVirtualMachine attempt count = %d, want 1", userCommandAttempts)
	}
}
