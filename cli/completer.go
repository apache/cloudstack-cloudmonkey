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

type CliCompleter struct {
	Config *config.Config
}

var completer *CliCompleter

func buildApiCacheMap(apiMap map[string][]*config.Api) map[string][]*config.Api {
	for _, cmd := range cmd.AllCommands() {
		verb := cmd.Name
		if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
			for _, scmd := range cmd.SubCommands {
				dummyApi := &config.Api{
					Name: scmd,
					Verb: verb,
				}
				apiMap[verb] = append(apiMap[verb], dummyApi)
			}
		} else {
			dummyApi := &config.Api{
				Name: "",
				Verb: verb,
			}
			apiMap[verb] = append(apiMap[verb], dummyApi)
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

func (t *CliCompleter) Do(line []rune, pos int) (options [][]rune, offset int) {

	apiMap := buildApiCacheMap(t.Config.GetApiVerbMap())

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
	var apiFound *config.Api
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
		search := arg.Name + "="
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

			var autocompleteApi *config.Api
			var relatedNoun string
			if arg.Name == "id" || arg.Name == "ids" {
				relatedNoun = apiFound.Noun
				if apiFound.Verb != "list" {
					relatedNoun += "s"
				}
			} else if arg.Name == "account" {
				relatedNoun = "accounts"
			} else {
				relatedNoun = strings.Replace(strings.Replace(arg.Name, "ids", "", -1), "id", "", -1) + "s"
			}
			for _, related := range apiMap["list"] {
				if relatedNoun == related.Noun {
					autocompleteApi = related
					break
				}
			}

			if autocompleteApi == nil {
				return nil, 0
			}

			r := cmd.NewRequest(nil, config.NewConfig(), nil, nil)
			autocompleteApiArgs := []string{"listall=true"}
			if autocompleteApi.Noun == "templates" {
				autocompleteApiArgs = append(autocompleteApiArgs, "templatefilter=all")
			}
			response, _ := cmd.NewAPIRequest(r, autocompleteApi.Name, autocompleteApiArgs)

			var autocompleteOptions []SelectOption
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
						opt := SelectOption{}
						if resource["id"] != nil {
							opt.Id = resource["id"].(string)
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
				fmt.Println()
				selectedOption := ShowSelector(autocompleteOptions)
				if strings.HasSuffix(arg.Name, "id") || strings.HasSuffix(arg.Name, "ids") {
					selected = selectedOption.Id
				} else {
					selected = selectedOption.Name
				}
			} else {
				if len(autocompleteOptions) == 1 {
					selected = autocompleteOptions[0].Id
				}
			}
			options = [][]rune{[]rune(selected + " ")}
			offset = 0
		}
	}

	return options, offset
}

func NewCompleter(cfg *config.Config) *CliCompleter {
	completer = &CliCompleter{
		Config: cfg,
	}
	return completer
}
