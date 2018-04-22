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
	"github.com/olekukonko/tablewriter"
	"os"
	"reflect"
	"sort"
	"strings"

	"cloudmonkey/config"
)

var apiCommand *Command

// GetAPIHandler returns a catchall command handler
func GetAPIHandler() *Command {
	return apiCommand
}

func printText(itemMap map[string]interface{}) {
	for k, v := range itemMap {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice {
			fmt.Printf("%s:\n", k)
			for _, item := range v.([]interface{}) {
				row, isMap := item.(map[string]interface{})
				if isMap {
					for field, value := range row {
						fmt.Printf("%s = %v\n", field, value)
					}
				} else {
					fmt.Printf("%v\n", item)
				}
				fmt.Println("================================================================================")
			}
		} else {
			fmt.Printf("%s = %v\n", k, v)
		}
	}
}

func printResult(outputType string, response map[string]interface{}, filter []string) {
	switch outputType {
	case config.TABLE:
		table := tablewriter.NewWriter(os.Stdout)
		for k, v := range response {
			valueType := reflect.TypeOf(v)
			if valueType.Kind() == reflect.Slice {
				items, ok := v.([]interface{})
				if !ok {
					continue
				}
				fmt.Printf("%s:\n", k)
				var header []string
				for _, item := range items {
					row, ok := item.(map[string]interface{})
					if !ok || len(row) < 1 {
						continue
					}

					if len(header) == 0 {
						for field, _ := range row {
							if filter != nil && len(filter) > 0 {
								for _, filterKey := range filter {
									if filterKey == field {
										header = append(header, field)
									}
								}
								continue
							}
							header = append(header, field)
						}
						sort.Strings(header)
						table.SetHeader(header)
					}
					var rowArray []string
					for _, field := range header {
						rowArray = append(rowArray, fmt.Sprintf("%v", row[field]))
					}
					table.Append(rowArray)
				}
			} else {
				fmt.Printf("%s = %v\n", k, v)
			}
		}
		table.Render()
	case config.TEXT:
		printText(response)
	default:
		jsonOutput, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(jsonOutput))
	}
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
				provided := false
				for _, arg := range apiArgs {
					if strings.HasPrefix(arg, required) {
						provided = true
					}
				}
				if !provided {
					missingArgs = append(missingArgs, strings.Replace(required, "=", "", -1))
				}
			}

			if len(missingArgs) > 0 {
				fmt.Println("💩 Missing required arguments: ", strings.Join(missingArgs, ", "))
				return nil
			}

			response, err := NewAPIRequest(r, api.Name, apiArgs)
			if err != nil {
				return err
			}

			var filterKeys []string
			for _, arg := range apiArgs {
				if strings.HasPrefix(arg, "filter=") {
					filterKeys = strings.Split(strings.Split(arg, "=")[1], ",")
				}
			}

			printResult(r.Config.Core.Output, response, filterKeys)

			return nil
		},
	}
}
