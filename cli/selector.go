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

type SelectOption struct {
	Id     string
	Name   string
	Detail string
}

type Selector struct {
	InUse bool
}

var selector Selector

func init() {
	selector = Selector{
		InUse: false,
	}
}

func (s Selector) lock() {
	s.InUse = true
}

func (s Selector) unlock() {
	s.InUse = false
}

func ShowSelector(options []SelectOption) SelectOption {
	if selector.InUse {
		return SelectOption{}
	}
	selector.lock()
	defer selector.unlock()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▶ {{ .Name | cyan }} ({{ .Id | red }})",
		Inactive: "  {{ .Name | cyan }} ({{ .Id | red }})",
		Selected: "👊Selected: {{ .Name | cyan }} ({{ .Id | red }})",
		Details: `
--------- Current Selection ----------
{{ "Id:" | faint }}  {{ .Id }}
{{ "Name:" | faint }} {{ .Name }}
{{ "Info:" | faint }}  {{ .Detail }}`,
	}

	searcher := func(input string, index int) bool {
		pepper := options[index]
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Use the arrow keys to navigate: ↓ ↑ → ← press / to toggle 🔍search",
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
		return SelectOption{}
	}

	return options[i]
}
