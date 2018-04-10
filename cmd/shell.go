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
	"os/exec"
	"strings"
)

func init() {
	AddCommand(&Command{
		Name: "shell",
		Help: "Drops into a shell",
		Handle: func(r *Request) error {
			cmd := strings.TrimSpace(strings.Join(r.Args, " "))
			if len(cmd) < 1 {
				return errors.New("no shell command provided")
			}
			out, err := exec.Command("bash", "-c", cmd).Output()
			if err == nil {
				fmt.Println(string(out))
				return nil
			}
			return errors.New("failed to execute command, " + err.Error())
		},
	})
}
