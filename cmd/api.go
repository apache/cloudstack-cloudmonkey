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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var apiCommand *Command

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

			api := r.Config.GetCache()[apiName]
			if api == nil {
				return errors.New("unknown command or API requested")
			}

			if strings.Contains(strings.Join(apiArgs, " "), "-h") {
				fmt.Println("=== Help docs ===")
				fmt.Printf("\033[34m%s\033[0m [async=%v] %s\n", api.Name, api.Async, api.Description)
				if len(api.RequiredArgs) > 0 {
					fmt.Println("Required params:", strings.Join(api.RequiredArgs, ", "))
				}
				if len(api.Args) > 0 {
					fmt.Println("API params:")
				}
				for _, arg := range api.Args {
					fmt.Printf("\033[35m%-24s\033[0m \033[36m%-12s\033[0m %s\n", arg.Name, arg.Type, arg.Description)
				}
				return nil
			}

			var missingArgs []string
			for _, required := range api.RequiredArgs {
				provided := false
				for _, arg := range apiArgs {
					if strings.HasPrefix(arg, required+"=") {
						provided = true
					}
				}
				if !provided {
					missingArgs = append(missingArgs, required)
				}
			}

			if len(missingArgs) > 0 {
				fmt.Println("💩 Missing required arguments: ", strings.Join(missingArgs, ", "))
				return nil
			}

			b, _ := NewAPIRequest(r, api.Name, apiArgs)
			response, _ := json.MarshalIndent(b, "", "  ")

			// Implement various output formats
			fmt.Println(string(response))
			return nil
		},
	}
	AddCommand(apiCommand)
}
