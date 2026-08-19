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
		Name: "switch",
		Help: "Switches profile",
		SubCommands: map[string][]string{
			"profile": {},
		},
		Handle: func(r *Request) error {
			if len(r.Args) < 1 {
				fmt.Println("Please provide one of the sub-commands: ", reflect.ValueOf(r.Command.SubCommands).MapKeys())
				return nil
			}
			subCommand := r.Args[0]
			value := strings.Trim(strings.Join(r.Args[1:], " "), " ")
			config.Debug("Switch command received:", subCommand, " values:", value)
			if r.Args[len(r.Args)-1] == "-h" {
				fmt.Println("Usage: switch <subcommand> <option>. Press tab-tab to see available subcommands and options.")
				return nil
			}
			validArgs := r.Command.SubCommands[subCommand]
			if len(validArgs) != 0 {
				if !config.CheckIfValuePresent(validArgs, value) {
					return errors.New("Invalid value for " + subCommand + ". Supported values: " + strings.Join(validArgs, ", "))
				}
			}

			if subCommand == "profile" {
				if err := r.Config.LoadProfile(value, false); err != nil {
					if r.Config.HasShell {
						fmt.Printf("Failed to switch to server profile: %s due to: %v\n", value, err)
						return nil
					}
					return err
				}
				if r.Config.HasShell {
					ap := r.Config.ActiveProfile
					fmt.Printf("Loaded server profile: %s\nUrl:        %s\nUsername:   %s\nDomain:     %s\nAPI Key:    %s\nTotal APIs: %d\n\n",
						r.Config.Core.ProfileName, ap.URL, ap.Username, ap.Domain, ap.APIKey, len(r.Config.GetCache()))
				}
			}
			return nil
		},
	})
}
