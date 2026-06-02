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

package config

import (
	"path/filepath"
	"strings"
	"testing"

	ini "gopkg.in/ini.v1"
)

func testCore() *Core {
	return &Core{
		Prompt:       "cmk",
		AsyncBlock:   true,
		Timeout:      1800,
		Output:       JSON,
		VerifyCert:   true,
		ProfileName:  "localcloud",
		AutoComplete: true,
		PostRequest:  true,
	}
}

func TestDefaultProfileSignatureAlgorithmAuto(t *testing.T) {
	profile := defaultProfile()
	if profile.SignatureAlgorithm != SignatureAlgorithmAuto {
		t.Fatalf("default signature algorithm = %q, want %q", profile.SignatureAlgorithm, SignatureAlgorithmAuto)
	}
}

func TestNormalizeSignatureAlgorithmCanonicalizesInput(t *testing.T) {
	tests := map[string]string{
		"":            SignatureAlgorithmAuto,
		" AUTO ":      SignatureAlgorithmAuto,
		"hmacsha1":    SignatureAlgorithmHmacSHA1,
		"HMACSHA512":  SignatureAlgorithmHmacSHA512,
		"HmacSHA512":  SignatureAlgorithmHmacSHA512,
		" HmacSHA1  ": SignatureAlgorithmHmacSHA1,
	}

	for input, want := range tests {
		got, err := NormalizeSignatureAlgorithm(input)
		if err != nil {
			t.Fatalf("NormalizeSignatureAlgorithm(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("NormalizeSignatureAlgorithm(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSaveConfigMigratesMissingSignatureAlgorithmToAuto(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config")
	conf := ini.Empty()
	conf.Section(ini.DEFAULT_SECTION).ReflectFrom(testCore())
	conf.Section("localcloud").ReflectFrom(&ServerProfile{
		URL:       DefaultACSAPIEndpoint,
		Username:  "admin",
		Password:  "password",
		Domain:    "/",
		APIKey:    "api-key",
		SecretKey: "secret-key",
	})
	conf.Section("localcloud").DeleteKey("signaturealgorithm")
	if err := conf.SaveTo(configFile); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Dir:         dir,
		ConfigFile:  configFile,
		HistoryFile: filepath.Join(dir, "history"),
		Core:        testCore(),
	}
	saveConfig(cfg)

	updated := readConfig(cfg)
	got := updated.Section("localcloud").Key("signaturealgorithm").String()
	if got != SignatureAlgorithmAuto {
		t.Fatalf("persisted signaturealgorithm = %q, want %q", got, SignatureAlgorithmAuto)
	}
	if cfg.ActiveProfile.SignatureAlgorithm != SignatureAlgorithmAuto {
		t.Fatalf("active profile signature algorithm = %q, want %q", cfg.ActiveProfile.SignatureAlgorithm, SignatureAlgorithmAuto)
	}
}

func TestPromptAddsFIPSIndicatorForHmacSHA512(t *testing.T) {
	cfg := &Config{
		Core: testCore(),
		ActiveProfile: &ServerProfile{
			SignatureAlgorithm: SignatureAlgorithmHmacSHA512,
		},
	}

	if got := cfg.GetPrompt(); !strings.Contains(got, "(localcloud-fips)") {
		t.Fatalf("prompt = %q, want FIPS profile indicator", got)
	}
}
