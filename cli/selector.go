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
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rhtyd/readline"
)

type SelectOptions struct {
	Name   string
	Id     string
	Detail string
}

func ShowSelector() string {
	options := []SelectOptions{
		{Name: "Option1", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option2", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option3", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option4", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option5", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option6", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option7", Id: "some-uuid", Detail: "Some Detail"},
		{Name: "Option8", Id: "some-uuid", Detail: "Some Detail"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "🐵 {{ .Name | cyan }} ({{ .Id | red }})",
		Inactive: "  {{ .Name | cyan }} ({{ .Id | red }})",
		Selected: "Selected: {{ .Name | cyan }} ({{ .Id | red }})",
		Details: `
--------- Current Selection ----------
{{ "Name:" | faint }} {{ .Name }}
{{ "Id:" | faint }}  {{ .Id }}
{{ "Detail:" | faint }}  {{ .Detail }}`,
	}

	searcher := func(input string, index int) bool {
		pepper := options[index]
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search",
		Items:             options,
		Templates:         templates,
		Size:              5,
		Searcher:          searcher,
		StartInSearchMode: true,
		Keys: &promptui.SelectKeys{
			Prev:     promptui.Key{Code: readline.CharPrev, Display: "↑"},
			Next:     promptui.Key{Code: readline.CharNext, Display: "↓"},
			PageUp:   promptui.Key{Code: readline.CharBackward, Display: "←"},
			PageDown: promptui.Key{Code: readline.CharForward, Display: "→"},
			Search:   promptui.Key{Code: '/', Display: "/"},
		},
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return options[i].Id
}
