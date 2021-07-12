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
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/internal/app"
	"github.com/apache/cloudstack-cloudmonkey/internal/config"
	"github.com/google/shlex"
)

// ExecLine executes a line of command
func ExecLine(line string) error {
	config.Debug("ExecLine line:", line)
	args, err := shlex.Split(line)
	if err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		for _, arg := range args {
			if arg == "|" {
				result, err := exec.Command("bash", "-c", fmt.Sprintf("%s %v", os.Args[0], line)).Output()
				fmt.Print(string(result))
				return err
			}
		}
	}

	return ExecCmd(args)
}

// ExecCmd executes a single provided command
func ExecCmd(args []string) error {
	config.Debug("ExecCmd args: ", strings.Join(args, ", "))
	if len(args) < 1 {
		return nil
	}

	command := app.FindCommand(args[0])
	if command != nil {
		return command.Handle(app.NewRequest(command, cfg, args[1:]))
	}

	catchAllHandler := app.GetAPIHandler()
	return catchAllHandler.Handle(app.NewRequest(catchAllHandler, cfg, args))
}
