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

var helpCommand *Command

func init() {
	helpCommand = &Command{
		Name: "help",
		Help: "Help",
		Handle: func(r *Request) error {
			if len(r.Args) < 1 || r.Args[0] == "-h" {
				PrintUsage()
				return nil
			}

			api := r.Config.GetCache()[strings.ToLower(r.Args[0])]
			if api == nil {
				return errors.New("unknown command or API requested")
			}

			fmt.Printf("\033[34m%s\033[0m: %s\n", api.Name, api.Description)
			if api.Async {
				fmt.Println("This API is \033[35masynchronous\033[0m.")
			}
			if len(api.RequiredArgs) > 0 {
				fmt.Println("Required params:", strings.Join(api.RequiredArgs, ", "))
			}
			if len(api.Args) > 0 {
				fmt.Printf("%-24s %-8s %s\n", "API Params", "Type", "Description")
				fmt.Printf("%-24s %-8s %s\n", "==========", "====", "===========")
			}
			for _, arg := range api.Args {
				fmt.Printf("\033[36m%-24s\033[0m \033[32m%-8s\033[0m ", arg.Name, arg.Type)
				info := []rune(arg.Description)
				for i, r := range info {
					fmt.Printf("%s", string(r))
					if i > 0 && i%40 == 0 {
						fmt.Println()
						for i := 0; i < 34; i++ {
							fmt.Printf(" ")
						}
					}
				}
				fmt.Println()
			}
			return nil
		},
	}
	AddCommand(helpCommand)
}
