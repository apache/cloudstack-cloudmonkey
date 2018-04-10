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

	"../cmd"
	"../config"

	"github.com/rhtyd/readline/runes"
)

type CliCompleter struct {
	Config *config.Config
}

var completer *CliCompleter

func TrimSpaceLeft(in []rune) []rune {
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

func (t *CliCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {

	line = TrimSpaceLeft(line[:pos])
	lineLen := len(line)

	apiCache := t.Config.GetCache()
	apiMap := make(map[string][]*config.Api)
	for api := range apiCache {
		verb := apiCache[api].Verb
		apiMap[verb] = append(apiMap[verb], apiCache[api])
	}

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

	var verbs []string
	for verb := range apiMap {
		verbs = append(verbs, verb)
		sort.Slice(apiMap[verb], func(i, j int) bool {
			return apiMap[verb][i].Name < apiMap[verb][j].Name
		})
	}
	sort.Strings(verbs)

	var verbsFound []string
	for _, verb := range verbs {
		search := verb + " "
		if !runes.HasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, lineLen, []rune(search))
			newLine = append(newLine, sLine...)
			offset = sOffset
		} else {
			verbsFound = append(verbsFound, verb)
		}
	}

	apiArg := false
	for _, verbFound := range verbsFound {
		search := verbFound + " "

		nLine := TrimSpaceLeft(line[len(search):])
		offset = lineLen - len(verbFound) - 1

		for _, api := range apiMap[verbFound] {
			resource := strings.TrimPrefix(strings.ToLower(api.Name), verbFound)
			search = resource + " "

			if runes.HasPrefix(nLine, []rune(search)) {
				// FIXME: handle params to API here with = stuff
				for _, arg := range api.Args {
					opt := arg.Name + "="
					newLine = append(newLine, []rune(opt))
				}
				if string(nLine[len(nLine)-1]) == "=" {
					apiArg = true
				}
				offset = lineLen - len(verbFound) - len(resource) - 1
			} else {
				sLine, _ := doInternal(nLine, pos, len(nLine), []rune(search))
				newLine = append(newLine, sLine...)
			}
		}
	}

	// FIXME: pass selector uuid options
	if apiArg {
		fmt.Println()
		option := ShowSelector()
		// show only one option in autocompletion
		newLine = [][]rune{[]rune(option)}
		offset = 0
	}

	return newLine, offset
}

func NewCompleter(cfg *config.Config) *CliCompleter {
	completer = &CliCompleter{
		Config: cfg,
	}
	return completer
}
