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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/apache/cloudstack-cloudmonkey/config"
)

func TestUploadFilesNonInteractive(t *testing.T) {
	file, err := os.CreateTemp("", "cloudmonkey-upload-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	const fileContent = "test upload content"

	if _, err := file.WriteString(fileContent); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	var uploaded bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %s, want POST", r.Method)
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		uploadedFile, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get uploaded file: %v", err)
		}
		defer uploadedFile.Close()

		data, err := io.ReadAll(uploadedFile)
		if err != nil {
			t.Fatalf("failed to read uploaded file: %v", err)
		}

		if string(data) != fileContent {
			t.Errorf("uploaded content = %q, want %q", string(data), fileContent)
		}

		uploaded = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		HasShell: false,
	}

	request := &Request{
		Config: cfg,
	}

	response := map[string]interface{}{
		"getuploadparams": map[string]interface{}{
			"postURL":   server.URL,
			"metadata":  "metadata",
			"signature": "signature",
			"expires":   "expires",
		},
	}

	UploadFiles(request, "getUploadParamsForTemplate", response, []string{file.Name()})

	if !uploaded {
		t.Fatal("file was not uploaded")
	}
}
