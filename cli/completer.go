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

func lastString(array []string) string {
	return array[len(array)-1]
}

type argOption struct {
	Value  string
	Detail string
}

func buildArgOptions(response map[string]interface{}, hasID bool) []argOption {
	argOptions := []argOption{}
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
				var id, name, detail string
				if resource["id"] != nil {
					id = resource["id"].(string)
				}
				if resource["name"] != nil {
					name = resource["name"].(string)
				} else if resource["username"] != nil {
					name = resource["username"].(string)
				} else if resource["hypervisor"] != nil && resource["hypervisorversion"] != nil {
					name = fmt.Sprintf("%s %s", resource["hypervisor"].(string), resource["hypervisorversion"].(string))
					if resource["osdisplayname"] != nil {
						name = fmt.Sprintf("%s; %s", resource["osdisplayname"].(string), name)
					}
				}
				if resource["displaytext"] != nil {
					detail = resource["displaytext"].(string)
				}
				if len(detail) == 0 && resource["description"] != nil {
					detail = resource["description"].(string)
				}
				if len(detail) == 0 && resource["ipaddress"] != nil {
					detail = resource["ipaddress"].(string)
				}
				var opt argOption
				if hasID {
					opt.Value = id
					opt.Detail = name
					if len(name) == 0 {
						opt.Detail = detail
					}
				} else {
					opt.Value = name
					opt.Detail = detail
					if len(name) == 0 {
						opt.Value = detail
					}
				}
				argOptions = append(argOptions, opt)
			}
			break
		}
	}
	return argOptions
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

func findAPI(apiMap map[string][]*config.API, relatedNoun string) *config.API {
	var autocompleteAPI *config.API
	for _, listAPI := range apiMap["list"] {
		if relatedNoun == listAPI.Noun {
			autocompleteAPI = listAPI
			break
		}
	}
	return autocompleteAPI
}

func findAutocompleteAPI(arg *config.APIArg, apiFound *config.API, apiMap map[string][]*config.API) *config.API {
	if arg.Type == "map" {
		return nil
	}

	var autocompleteAPI *config.API
	argName := strings.Replace(arg.Name, "=", "", -1)
	relatedNoun := argName
	switch {
	case argName == "id" || argName == "ids":
		// Heuristic: user is trying to autocomplete for id/ids arg for a list API
		relatedNoun = apiFound.Noun
		if apiFound.Verb != "list" {
			relatedNoun += "s"
		}
	case argName == "account":
		// Heuristic: user is trying to autocomplete for accounts
		relatedNoun = "accounts"
	case argName == "ipaddressid":
		// Heuristic: user is trying to autocomplete for ip addresses
		relatedNoun = "publicipaddresses"
	case argName == "storageid":
		relatedNoun = "storagepools"
	case argName == "associatednetworkid":
		relatedNoun = "networks"
	default:
		// Heuristic: autocomplete for the arg for which a list<Arg without id/ids>s API exists
		// For example, for zoneid arg, listZones API exists
		cutIdx := len(argName)
		if strings.HasSuffix(argName, "id") {
			cutIdx -= 2
		} else if strings.HasSuffix(argName, "ids") {
			cutIdx -= 3
		} else {
		}
		relatedNoun = argName[:cutIdx] + "s"
	}

	config.Debug("Possible related noun for the arg: ", relatedNoun, " and type: ", arg.Type)
	autocompleteAPI = findAPI(apiMap, relatedNoun)

	if autocompleteAPI == nil {
		if strings.Contains(strings.ToLower(relatedNoun), "storage") {
			relatedNoun = "storagepools"
			autocompleteAPI = findAPI(apiMap, relatedNoun)
		}
	}

	if autocompleteAPI != nil {
		config.Debug("Autocomplete: API found using heuristics: ", autocompleteAPI.Name)
	}

	if strings.HasSuffix(relatedNoun, "s") {
		relatedNoun = relatedNoun[:len(relatedNoun)-1]
	}

	// Heuristic: find any list API that contains the arg name
	if autocompleteAPI == nil {
		config.Debug("Finding possible API that have: ", argName, " related APIs: ", arg.Related)
		possibleAPIs := []*config.API{}
		for _, listAPI := range apiMap["list"] {
			if strings.Contains(listAPI.Noun, argName) {
				config.Debug("Found possible API: ", listAPI.Name)
				possibleAPIs = append(possibleAPIs, listAPI)
			}
		}
		if len(possibleAPIs) == 1 {
			autocompleteAPI = possibleAPIs[0]
		}
	}

	return autocompleteAPI
}

type autoCompleter struct {
	Config *config.Config
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

	// Auto-complete API arg
	splitLine := strings.Split(string(line), " ")
	line = trimSpaceLeft([]rune(splitLine[len(splitLine)-1]))
	for _, arg := range apiFound.Args {
		search := arg.Name
		if !hasPrefix(line, []rune(search)) {
			sLine, sOffset := doInternal(line, pos, len(line), []rune(search))
			options = append(options, sLine...)
			offset = sOffset
		} else {
			words := strings.Split(string(line), "=")
			argInput := lastString(words)
			if arg.Type == "boolean" {
				for _, search := range []string{"true ", "false "} {
					offset = 0
					if strings.HasPrefix(search, argInput) {
						options = append(options, []rune(search[len(argInput):]))
						offset = len(argInput)
					}
				}
				return
			}
			if arg.Type == config.FAKE && arg.Name == "filter=" {
				offset = 0
				filterInputs := strings.Split(strings.Replace(argInput, ",", ",|", -1), "|")
				lastFilterInput := lastString(filterInputs)
				for _, key := range apiFound.ResponseKeys {
					if inArray(key, filterInputs) {
						continue
					}
					if strings.HasPrefix(key, lastFilterInput) {
						options = append(options, []rune(key[len(lastFilterInput):]))
						offset = len(lastFilterInput)
					}
				}
				return
			}

			autocompleteAPI := findAutocompleteAPI(arg, apiFound, apiMap)
			if autocompleteAPI == nil {
				return nil, 0
			}

			completeArgs := t.Config.Core.AutoComplete
			autocompleteAPIArgs := []string{}
			argOptions := []argOption{}
			if completeArgs {
				autocompleteAPIArgs = []string{"listall=true"}
				if autocompleteAPI.Noun == "templates" {
					autocompleteAPIArgs = append(autocompleteAPIArgs, "templatefilter=executable")
				}

				if apiFound.Name != "provisionCertificate" && autocompleteAPI.Name == "listHosts" {
					autocompleteAPIArgs = append(autocompleteAPIArgs, "type=Routing")
				} else if apiFound.Name == "migrateSystemVm" {
					autocompleteAPI.Name = "listSystemVms"
				} else if apiFound.Name == "associateIpAddress" {
					autocompleteAPIArgs = append(autocompleteAPIArgs, "state=Free")
				}

				spinner := t.Config.StartSpinner("fetching options, please wait...")
				request := cmd.NewRequest(nil, completer.Config, nil)
				response, _ := cmd.NewAPIRequest(request, autocompleteAPI.Name, autocompleteAPIArgs, false)
				t.Config.StopSpinner(spinner)

				hasID := strings.HasSuffix(arg.Name, "id=") || strings.HasSuffix(arg.Name, "ids=")
				argOptions = buildArgOptions(response, hasID)
			}

			filteredOptions := []argOption{}
			if len(argOptions) > 0 {
				sort.Slice(argOptions, func(i, j int) bool {
					return argOptions[i].Value < argOptions[j].Value
				})
				for _, item := range argOptions {
					if strings.HasPrefix(item.Value, argInput) {
						filteredOptions = append(filteredOptions, item)
					}
				}
			}
			offset = 0
			if len(filteredOptions) == 0 {
				options = [][]rune{[]rune("")}
			}
			for _, item := range filteredOptions {
				option := item.Value + " "
				if len(filteredOptions) > 1 && len(item.Detail) > 0 {
					option += fmt.Sprintf("(%v)", item.Detail)
				}
				if strings.HasPrefix(option, argInput) {
					options = append(options, []rune(option[len(argInput):]))
					offset = len(argInput)
				}
			}
			return
		}
	}

	return options, offset
}
