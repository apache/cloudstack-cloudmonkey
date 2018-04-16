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

	"cloudmonkey/cmd"
	"cloudmonkey/config"

	"github.com/chzyer/readline/runes"
)

type autoCompleter struct {
	Config *config.Config
}

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

func doInternal(line []rune, pos int, lineLen int, argName []rune) (newLine [][]rune, offset int) {
	offset = lineLen
	if lineLen >= len(argName) {
		if runes.HasPrefix(line, argName) {
			if lineLen == len(argName) {
				newLine = append(newLine, []rune{' '})
			} else {
				newLine = append(newLine, argName)
			}
			offset = offset - len(argName) - 1
		}
	} else {
		if runes.HasPrefix(argName, line) {
			newLine = append(newLine, argName[offset:])
		}
	}
	return
}

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
		if !runes.HasPrefix(line, []rune(search)) {
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
		if !runes.HasPrefix(line, []rune(search)) {
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
		if !runes.HasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, len(line), []rune(search))
			options = append(options, sLine...)
			offset = sOffset
		} else {
			if arg.Type == "boolean" {
				options = [][]rune{[]rune("true "), []rune("false ")}
				offset = 0
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

			r := cmd.NewRequest(nil, completer.Config, nil, nil)
			autocompleteAPIArgs := []string{"listall=true"}
			if autocompleteAPI.Noun == "templates" {
				autocompleteAPIArgs = append(autocompleteAPIArgs, "templatefilter=all")
			}
			fmt.Printf("\nFetching options, please wait...")
			response, _ := cmd.NewAPIRequest(r, autocompleteAPI.Name, autocompleteAPIArgs)
			fmt.Printf("\r")

			var autocompleteOptions []selectOption
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
						opt := selectOption{}
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

						autocompleteOptions = append(autocompleteOptions, opt)
					}
					break
				}
			}

			var selected string
			if len(autocompleteOptions) > 1 {
				sort.Slice(autocompleteOptions, func(i, j int) bool {
					return autocompleteOptions[i].Name < autocompleteOptions[j].Name
				})
				selectedOption := showSelector(autocompleteOptions)
				if strings.HasSuffix(arg.Name, "id=") || strings.HasSuffix(arg.Name, "ids=") {
					selected = selectedOption.ID
				} else {
					selected = selectedOption.Name
				}
			} else {
				if len(autocompleteOptions) == 1 {
					selected = autocompleteOptions[0].ID
				}
			}
			options = [][]rune{[]rune(selected + " ")}
			offset = 0
		}
	}

	return options, offset
}
