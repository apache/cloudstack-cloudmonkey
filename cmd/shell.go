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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RunShellCommand executes a local OS shell command and streams output to stdout/stderr
func RunShellCommand(command string) error {
	command = strings.TrimSpace(command)
	if len(command) == 0 {
		return errors.New("please provide a shell command to execute")
	}

	var shellCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		shellCmd = exec.Command("cmd", "/C", command)
	} else {
		shellBin := os.Getenv("SHELL")
		if len(shellBin) == 0 {
			shellBin = "sh"
		}
		shellCmd = exec.Command(shellBin, "-c", command)
	}

	shellCmd.Stdin = os.Stdin
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	return shellCmd.Run()
}

func init() {
	AddCommand(&Command{
		Name: "shell",
		Help: "Runs a local shell command (or prefix with '!')",
		Handle: func(r *Request) error {
			if len(r.Args) == 0 {
				return errors.New("please provide a shell command to execute")
			}
			if r.Args[0] == "-h" {
				fmt.Println("Usage: shell <command> or !<command>")
				return nil
			}
			return RunShellCommand(strings.Join(r.Args, " "))
		},
	})
}
