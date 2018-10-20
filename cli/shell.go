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

	"github.com/apache/cloudstack-cloudmonkey/config"
	"github.com/c-bata/go-prompt"
)

var cfg *config.Config

func executor(in string) {
	if err := ExecLine(in); err != nil {
		fmt.Println("🙈 Error:", err)
	}
}

func prefix() (string, bool) {
	return cfg.GetPrompt(), true
}

// ExecShell starts a shell
func ExecShell(sysArgs []string) {
	cfg = config.NewConfig()

	if len(sysArgs) > 0 {
		err := ExecCmd(sysArgs)
		if err != nil {
			fmt.Println("🙈 Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg.HasShell = true
	cfg.PrintHeader()

	lines, _ := readLines(cfg.HistoryFile)

	shell := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("cloudmonkey"),
		prompt.OptionPrefix(cfg.GetPrompt()),
		prompt.OptionLivePrefix(prefix),
		prompt.OptionMaxSuggestion(8),
		prompt.OptionHistory(lines),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionPreviewSuggestionTextColor(prompt.DarkBlue),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkBlue),
		prompt.OptionSelectedDescriptionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.Black),
		prompt.OptionSuggestionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.LightGray),
		prompt.OptionScrollbarThumbColor(prompt.DarkBlue),
		prompt.OptionScrollbarBGColor(prompt.LightGray),
	)
	shell.Run()
}
