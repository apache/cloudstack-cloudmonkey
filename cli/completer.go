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
	"unicode"

	"github.com/apache/cloudstack-cloudmonkey/cmd"
	"github.com/apache/cloudstack-cloudmonkey/config"
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

func trimSpaceLeft(in []rune) []rune {
	firstIndex := len(in)
	for i, r := range in {
		if unicode.IsSpace(r) == false {
			firstIndex = i
			break
		}
	}
	return in[firstIndex:]
}

func equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func hasPrefix(r, prefix []rune) bool {
	if len(r) < len(prefix) {
		return false
	}
	return equal(r[:len(prefix)], prefix)
}

func inArray(s string, array []string) bool {
	for _, item := range array {
		if s == item {
			return true
		}
	}
	return false
}

type autoCompleter struct {
	Config *config.Config
}

type selectOption struct {
	ID     string
	Name   string
	Detail string
}

func doInternal(line []rune, pos int, lineLen int, argName []rune) (newLine [][]rune, offset int) {
	offset = lineLen
	if lineLen >= len(argName) {
		if hasPrefix(line, argName) {
			if lineLen == len(argName) {
				newLine = append(newLine, []rune{' '})
			} else {
				newLine = append(newLine, argName)
			}
			offset = offset - len(argName) - 1
		}
	} else {
		if hasPrefix(argName, line) {
			newLine = append(newLine, argName[offset:])
		}
	}
	return
}

// FIXME; use cached response
var cachedResponse map[string]interface{}

func (t *autoCompleter) Do(line []rune, pos int) (options [][]rune, offset int) {
	apiMap := buildAPICacheMap(t.Config.GetAPIVerbMap())

	var verbs []string
	for verb := range apiMap {
		verbs = append(verbs, verb)
		sort.Slice(apiMap[verb], func(i, j int) bool {
			return apiMap[verb][i].Name < apiMap[verb][j].Name
		})
	}
	sort.Strings(verbs)

	line = trimSpaceLeft(line[:pos])

	// Auto-complete verb
	var verbFound string
	for _, verb := range verbs {
		search := verb + " "
		if !hasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, len(line), []rune(search))
			options = append(options, sLine...)
			offset = sOffset
		} else {
			verbFound = verb
			break
		}
	}
	if len(verbFound) == 0 {
		return
	}

	// Auto-complete noun
	var nounFound string
	line = trimSpaceLeft(line[len(verbFound):])
	for _, api := range apiMap[verbFound] {
		search := api.Noun + " "
		if !hasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, len(line), []rune(search))
			options = append(options, sLine...)
			offset = sOffset
		} else {
			nounFound = api.Noun
			break
		}
	}
	if len(nounFound) == 0 {
		return
	}

	// Find API
	var apiFound *config.API
	for _, api := range apiMap[verbFound] {
		if api.Noun == nounFound {
			apiFound = api
			break
		}
	}
	if apiFound == nil {
		return
	}

	// Auto-complete api args
	splitLine := strings.Split(string(line), " ")
	line = trimSpaceLeft([]rune(splitLine[len(splitLine)-1]))
	for _, arg := range apiFound.Args {
		search := arg.Name
		if !hasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, len(line), []rune(search))
			options = append(options, sLine...)
			offset = sOffset
		} else {
			if arg.Type == "boolean" {
				options = [][]rune{[]rune("true "), []rune("false ")}
				offset = 0
				return
			}
			if arg.Type == config.FAKE && arg.Name == "filter=" {
				options = [][]rune{}
				offset = 0
				for _, key := range apiFound.ResponseKeys {
					options = append(options, []rune(key))
				}
				return
			}

			argName := strings.Replace(arg.Name, "=", "", -1)
			var autocompleteAPI *config.API
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
					autocompleteAPI = related
					break
				}
			}

			if autocompleteAPI == nil {
				return nil, 0
			}

			r := cmd.NewRequest(nil, completer.Config, nil)
			autocompleteAPIArgs := []string{"listall=true"}
			if autocompleteAPI.Noun == "templates" {
				autocompleteAPIArgs = append(autocompleteAPIArgs, "templatefilter=executable")
			}

			spinner := t.Config.StartSpinner("fetching options, please wait...")
			response, _ := cmd.NewAPIRequest(r, autocompleteAPI.Name, autocompleteAPIArgs, false)
			t.Config.StopSpinner(spinner)

			selectOptions := []selectOption{}
			for _, v := range response {
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
						var opt selectOption
						if resource["id"] != nil {
							opt.ID = resource["id"].(string)
						}
						if resource["name"] != nil {
							opt.Name = resource["name"].(string)
						} else if resource["username"] != nil {
							opt.Name = resource["username"].(string)
						}
						if resource["displaytext"] != nil {
							opt.Detail = resource["displaytext"].(string)
						}

						selectOptions = append(selectOptions, opt)
					}
					break
				}
			}

			sort.Slice(selectOptions, func(i, j int) bool {
				return selectOptions[i].Name < selectOptions[j].Name
			})

			hasID := strings.HasSuffix(arg.Name, "id=") || strings.HasSuffix(arg.Name, "ids=")
			if len(selectOptions) > 1 {
				for _, item := range selectOptions {
					var option string
					if hasID {
						option = fmt.Sprintf("%v (%v)", item.ID, item.Name)
					} else {
						if len(item.Detail) == 0 {
							option = fmt.Sprintf("%v ", item.Name)
						} else {
							option = fmt.Sprintf("%v (%v)", item.Name, item.Detail)
						}
					}
					options = append(options, []rune(option))
				}
			} else {
				option := ""
				if len(selectOptions) == 1 {
					if hasID {
						option = selectOptions[0].ID
					} else {
						option = selectOptions[0].Name
					}
				}
				options = [][]rune{[]rune(option + " ")}
			}
			offset = 0
			return
		}
	}

	return options, offset
}
