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
	"io"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/config"
	"github.com/chzyer/readline"
)

// CLI config instance
var cfg *config.Config

// SetConfig allows to set a config.Config object to cli
func SetConfig(c *config.Config) {
	cfg = c
}

var completer *autoCompleter
var shell *readline.Instance

// ExecPrompt starts a CLI prompt
func ExecPrompt() {
	completer = &autoCompleter{
		Config: cfg,
	}
	shell, err := readline.NewEx(&readline.Config{
		Prompt:            cfg.GetPrompt(),
		HistoryFile:       cfg.HistoryFile,
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		VimMode:           false,
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
	})

	if err != nil {
		panic(err)
	}
	defer shell.Close()

	cfg.HasShell = true
	cfg.PrintHeader()

	for {
		shell.SetPrompt(cfg.GetPrompt())
		line, err := shell.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) < 1 {
			continue
		}

		if err = ExecLine(line); err != nil {
			fmt.Println("🙈 Error:", err)
		}
	}

}
