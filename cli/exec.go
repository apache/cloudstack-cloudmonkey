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
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/cmd"
	"github.com/mattn/go-shellwords"
)

func ExecLine(line string) error {
	shellwords.ParseEnv = true
	parser := shellwords.NewParser()
	args, err := parser.Parse(line)
	if err != nil {
		fmt.Println("🙈 Failed to parse line:", err)
		return err
	}

	if parser.Position > 0 {
		line = fmt.Sprintf("shell %s %v", cfg.Name(), line)
		args = strings.Split(line, " ")
	}

	return ExecCmd(args)
}

// ExecCmd executes a single provided command
func ExecCmd(args []string) error {
	if len(args) < 1 {
		return nil
	}

	command := cmd.FindCommand(args[0])
	if command != nil {
		return command.Handle(cmd.NewRequest(command, cfg, args[1:]))
	}

	catchAllHandler := cmd.GetAPIHandler()
	return catchAllHandler.Handle(cmd.NewRequest(catchAllHandler, cfg, args))
}
