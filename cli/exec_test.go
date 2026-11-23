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

package cli

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	runErr := fn()

	w.Close()
	os.Stdout = origStdout

	out, readErr := ioutil.ReadAll(r)
	r.Close()
	if readErr != nil {
		t.Fatalf("failed to read stdout pipe: %v", readErr)
	}
	return string(out), runErr
}

func TestExecLineBangShellCommand(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return ExecLine("!echo cmk-bang-test")
	})
	if err != nil {
		t.Fatalf("ExecLine(!echo) returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "cmk-bang-test") {
		t.Fatalf("expected stdout to contain %q, got %q", "cmk-bang-test", out)
	}
}

func TestExecLineShellCommand(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return ExecLine("shell echo cmk-shell-test")
	})
	if err != nil {
		t.Fatalf("ExecLine(shell echo) returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "cmk-shell-test") {
		t.Fatalf("expected stdout to contain %q, got %q", "cmk-shell-test", out)
	}
}

func TestExecLineEmptyShellCommandReturnsError(t *testing.T) {
	if err := ExecLine("shell"); err == nil {
		t.Fatal("expected empty 'shell' command to return an error")
	}
	if err := ExecLine("!"); err != nil {
		t.Fatalf("expected bare '!' to be a no-op, got error: %v", err)
	}
}
