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
	"reflect"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/config"
)

func init() {
	AddCommand(&Command{
		Name: "set",
		Help: "Configures options for cmk",
		SubCommands: map[string][]string{
			"prompt":       {"🐵", "🐱", "random"},
			"asyncblock":   {"true", "false"},
			"timeout":      {"600", "1800", "3600"},
			"output":       config.GetOutputFormats(),
			"profile":      {},
			"url":          {},
			"username":     {},
			"password":     {},
			"domain":       {},
			"apikey":       {},
			"secretkey":    {},
			"verifycert":   {"true", "false"},
			"debug":        {"true", "false"},
			"autocomplete": {"true", "false"},
			"postrequest":  {"true", "false"},
		},
		Handle: func(r *Request) error {
			if len(r.Args) < 1 {
				fmt.Println("Please provide one of the sub-commands: ", reflect.ValueOf(r.Command.SubCommands).MapKeys())
				return nil
			}
			subCommand := r.Args[0]
			value := strings.Trim(strings.Join(r.Args[1:], " "), " ")
			config.Debug("Set command received:", subCommand, " values:", value)
			if r.Args[len(r.Args)-1] == "-h" {
				fmt.Println("Usage: set <subcommand> <option>. Press tab-tab to see available subcommands and options.")
				return nil
			}
			if subCommand == "display" {
				subCommand = "output"
			}
			validArgs := r.Command.SubCommands[subCommand]
			if len(validArgs) != 0 && subCommand != "timeout" {
				if !config.CheckIfValuePresent(validArgs, value) {
					return errors.New("Invalid value set for " + subCommand + ". Supported values: " + strings.Join(validArgs, ", "))
				}
			}
			r.Config.UpdateConfig(subCommand, value, true)

			if subCommand == "profile" && r.Config.HasShell {
				fmt.Println("Loaded server profile:", r.Config.Core.ProfileName)
				fmt.Println("Url:        ", r.Config.ActiveProfile.URL)
				fmt.Println("Username:   ", r.Config.ActiveProfile.Username)
				fmt.Println("Domain:     ", r.Config.ActiveProfile.Domain)
				fmt.Println("API Key:    ", r.Config.ActiveProfile.APIKey)
				fmt.Println("Total APIs: ", len(r.Config.GetCache()))

				fmt.Println()
			}
			return nil
		},
	})
}
