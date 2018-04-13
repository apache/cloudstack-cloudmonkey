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

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
)

type selectOption struct {
	ID     string
	Name   string
	Detail string
}

type selector struct {
	InUse bool
}

var optionSelector selector

func init() {
	optionSelector = selector{
		InUse: false,
	}
}

func (s selector) lock() {
	s.InUse = true
}

func (s selector) unlock() {
	s.InUse = false
}

func showSelector(options []selectOption) selectOption {
	if optionSelector.InUse {
		return selectOption{}
	}
	optionSelector.lock()
	defer optionSelector.unlock()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▶ {{ .Name | cyan }} ({{ .ID | red }})",
		Inactive: "  {{ .Name | cyan }} ({{ .ID | red }})",
		Selected: "👊Selected: {{ .Name | cyan }} ({{ .ID | red }})",
		Details: `
--------- Current Selection ----------
{{ "ID:" | faint }}  {{ .ID }}
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
		return selectOption{}
	}

	return options[i]
}
