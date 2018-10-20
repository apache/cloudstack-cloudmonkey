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
	"sort"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/cmd"
	"github.com/apache/cloudstack-cloudmonkey/config"
	prompt "github.com/c-bata/go-prompt"
)

func buildAPICacheMap(apiMap map[string][]*config.API) map[string][]*config.API {
	for _, cmd := range cmd.AllCommands() {
		verb := cmd.Name
		if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
			for command, opts := range cmd.SubCommands {
				var args []*config.APIArg
				options := opts
				if command == "profile" {
					options = config.GetProfiles()
				}
				for _, opt := range options {
					args = append(args, &config.APIArg{
						Name: opt,
					})
				}
				apiMap[verb] = append(apiMap[verb], &config.API{
					Name: command,
					Verb: verb,
					Noun: command,
					Args: args,
				})
			}
		} else {
			dummyAPI := &config.API{
				Name: "",
				Verb: verb,
			}
			apiMap[verb] = append(apiMap[verb], dummyAPI)
		}
	}
	return apiMap
}

func inArray(s string, array []string) bool {
	for _, item := range array {
		if s == item {
			return true
		}
	}
	return false
}

var cachedResponse map[string]interface{}

func completer(in prompt.Document) []prompt.Suggest {
	if in.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}
	args := strings.Split(strings.TrimLeft(in.TextBeforeCursor(), " "), " ")

	for i := range args {
		if args[i] == "|" {
			return []prompt.Suggest{}
		}
	}

	w := in.GetWordBeforeCursor()
	s := []prompt.Suggest{}
	apiMap := buildAPICacheMap(cfg.GetAPIVerbMap())

	if len(args) <= 1 {
		for verb := range apiMap {
			s = append(s, prompt.Suggest{
				Text: verb,
			})
		}
	} else if len(args) == 2 {
		for _, api := range apiMap[args[0]] {
			s = append(s, prompt.Suggest{
				Text:        api.Noun,
				Description: api.Description,
			})
		}
	} else {
		var apiFound *config.API
		for _, api := range apiMap[args[0]] {
			if api.Noun == args[1] {
				apiFound = api
				break
			}
		}
		opts := []string{}
		for _, arg := range args[2:] {
			if strings.Contains(arg, "=") {
				opts = append(opts, strings.Split(arg, "=")[0])
			}
		}
		if apiFound != nil {
			if strings.HasSuffix(w, "=") {
				var argFound *config.APIArg
				for _, arg := range apiFound.Args {
					if arg.Name+"=" == w {
						argFound = arg
						break
					}
				}
				if argFound != nil {
					switch argFound.Type {
					case "boolean":
						s = append(s, prompt.Suggest{
							Text: "true",
						})
						s = append(s, prompt.Suggest{
							Text: "false",
						})
					case config.FAKE:
						// No suggestions for filter
					default:
						argName := argFound.Name
						var optionsAPI *config.API
						var relatedNoun string
						if argName == "id" || argName == "ids" {
							relatedNoun = apiFound.Noun
							if apiFound.Verb != "list" {
								relatedNoun += "s"
							}
						} else if argName == "account" {
							relatedNoun = "accounts"
						} else {
							relatedNoun = strings.Replace(strings.Replace(argName, "ids", "", -1), "id", "", -1) + "s"
						}

						for _, related := range apiMap["list"] {
							if relatedNoun == related.Noun {
								optionsAPI = related
								break
							}
						}
						if optionsAPI != nil {
							r := cmd.NewRequest(nil, cfg, nil)
							optionsArgs := []string{"listall=true"}
							if optionsAPI.Noun == "templates" {
								optionsArgs = append(optionsArgs, "templatefilter=executable")
							}

							if cachedResponse == nil {
								fmt.Println("")
								spinner := cfg.StartSpinner("fetching options, please wait...")
								cachedResponse, _ = cmd.NewAPIRequest(r, optionsAPI.Name, optionsArgs, false)
								cfg.StopSpinner(spinner)
							}

							for _, v := range cachedResponse {
								switch obj := v.(type) {
								case []interface{}:
									if obj == nil {
										break
									}
									for _, item := range obj {
										resource, ok := item.(map[string]interface{})
										if !ok {
											continue
										}
										opt := prompt.Suggest{}
										if resource["id"] != nil {
											opt.Text = resource["id"].(string)
										}
										if resource["name"] != nil {
											opt.Description = resource["name"].(string)
										} else if resource["username"] != nil {
											opt.Description = resource["username"].(string)
										}
										if opt.Text == "" {
											opt.Text = opt.Description
										}
										s = append(s, opt)
									}
									break
								}
							}

						}
					}
					for idx, es := range s {
						s[idx].Text = w + es.Text
					}
					return s
				}

			} else {
				for _, arg := range apiFound.Args {
					if inArray(arg.Name, opts) {
						continue
					}
					s = append(s, prompt.Suggest{
						Text:        arg.Name,
						Description: arg.Description,
					})
				}
				cachedResponse = nil
			}
		}
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Text < s[j].Text
	})

	return prompt.FilterHasPrefix(s, w, true)
}
