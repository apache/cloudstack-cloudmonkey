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
	"strings"
)

var apiCommand *Command

// GetAPIHandler returns a catchall command handler
func GetAPIHandler() *Command {
	return apiCommand
}

func init() {
	apiCommand = &Command{
		Name: "api",
		Help: "Runs a provided API",
		Handle: func(r *Request) error {
			if len(r.Args) == 0 {
				return errors.New("please provide an API to execute")
			}

			apiName := strings.ToLower(r.Args[0])
			apiArgs := r.Args[1:]
			if r.Config.GetCache()[apiName] == nil && len(r.Args) > 1 {
				apiName = strings.ToLower(strings.Join(r.Args[:2], ""))
				apiArgs = r.Args[2:]
			}

			for _, arg := range r.Args {
				if arg == "-h" {
					r.Args[0] = apiName
					return helpCommand.Handle(r)
				}
			}

			api := r.Config.GetCache()[apiName]
			if api == nil {
				return errors.New("unknown command or API requested")
			}

			var missingArgs []string
			for _, required := range api.RequiredArgs {
				required = strings.ReplaceAll(required, "=", "")
				provided := false
				for _, arg := range apiArgs {
					if strings.Contains(arg, "=") && strings.HasPrefix(arg, required) {
						provided = true
					}
				}
				if !provided {
					missingArgs = append(missingArgs, strings.Replace(required, "=", "", -1))
				}
			}

			if len(missingArgs) > 0 {
				fmt.Println("💩 Missing required parameters: ", strings.Join(missingArgs, ", "))
				return nil
			}

			response, err := NewAPIRequest(r, api.Name, apiArgs, api.Async)
			if err != nil {
				if strings.HasSuffix(err.Error(), "context canceled") {
					return nil
				} else if response != nil {
					printResult(r.Config.Core.Output, response, nil, nil)
				}
				return err
			}

			var filterKeys []string
			for _, arg := range apiArgs {
				if strings.HasPrefix(arg, "filter=") {
					for _, filterKey := range strings.Split(strings.Split(arg, "=")[1], ",") {
						if len(strings.TrimSpace(filterKey)) > 0 {
							filterKeys = append(filterKeys, strings.TrimSpace(filterKey))
						}
					}
				}
			}

			var excludeKeys []string
			for _, arg := range apiArgs {
				if strings.HasPrefix(arg, "exclude=") {
					for _, excludeKey := range strings.Split(strings.Split(arg, "=")[1], ",") {
						if len(strings.TrimSpace(excludeKey)) > 0 {
							excludeKeys = append(excludeKeys, strings.TrimSpace(excludeKey))
						}
					}
				}
			}

			if len(response) > 0 {
				printResult(r.Config.Core.Output, response, filterKeys, excludeKeys)
			}

			return nil
		},
	}
}
