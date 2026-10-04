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
	"reflect"
	"strings"
	"testing"

	"github.com/apache/cloudstack-cloudmonkey/config"
)

func TestFindAutocompleteAPIRelatedNounMatch(t *testing.T) {
	arg := &config.APIArg{
		Name: "domainid=",
		Related: []string{
			"createDomain",
			"listDomains",
			"updateDomain",
		},
	}

	apiFound := &config.API{
		Name: "listVirtualMachines",
		Verb: "list",
		Noun: "virtualmachines",
	}

	apiMap := map[string][]*config.API{
		"list": {
			{
				Name: "listDomains",
				Noun: "domains",
			},
		},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result == nil {
		t.Fatal("expected API, got nil")
	}

	if result.Name != "listDomains" {
		t.Fatalf("expected listDomains, got %s", result.Name)
	}
}

func TestFindAutocompleteAPIRelatedFallback(t *testing.T) {
	arg := &config.APIArg{
		Name: "domainid=",
		Related: []string{
			"listDomainChildren",
		},
	}

	apiFound := &config.API{
		Name: "listVirtualMachines",
		Verb: "list",
		Noun: "virtualmachines",
	}

	apiMap := map[string][]*config.API{
		"list": {
			{
				Name: "listDomainChildren",
				Noun: "domainchildren",
			},
		},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result == nil {
		t.Fatal("expected API, got nil")
	}

	if result.Name != "listDomainChildren" {
		t.Fatalf("expected listDomainChildren, got %s", result.Name)
	}
}

func TestFindAutocompleteAPIEmptyRelatedFallsBackToHeuristic(t *testing.T) {
	arg := &config.APIArg{
		Name: "zoneid=",
	}

	apiFound := &config.API{
		Name: "listVirtualMachines",
		Verb: "list",
		Noun: "virtualmachines",
	}

	apiMap := map[string][]*config.API{
		"list": {
			{
				Name: "listZones",
				Noun: "zones",
			},
		},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result == nil {
		t.Fatal("expected API, got nil")
	}

	if result.Name != "listZones" {
		t.Fatalf("expected listZones, got %s", result.Name)
	}
}

func TestFindAutocompleteAPINonListRelatedFallsBackToHeuristic(t *testing.T) {
	arg := &config.APIArg{
		Name: "zoneid=",
		Related: []string{
			"createZone",
			"updateZone",
		},
	}

	apiFound := &config.API{
		Name: "listVirtualMachines",
		Verb: "list",
		Noun: "virtualmachines",
	}

	apiMap := map[string][]*config.API{
		"list": {
			{
				Name: "listZones",
				Noun: "zones",
			},
		},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result == nil {
		t.Fatal("expected API, got nil")
	}

	if result.Name != "listZones" {
		t.Fatalf("expected listZones, got %s", result.Name)
	}
}

func TestFindAutocompleteAPIMapTypeReturnsNil(t *testing.T) {
	arg := &config.APIArg{
		Type: "map",
	}

	apiFound := &config.API{
		Name: "listVirtualMachines",
		Verb: "list",
		Noun: "virtualmachines",
	}

	apiMap := map[string][]*config.API{
		"list": {},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestFindAutocompleteAPIHeuristicWinsOverRelated(t *testing.T) {
	// registerIso's projectid arg lists many related APIs; the noun heuristic
	// must keep winning so the completion stays listProjects.
	arg := &config.APIArg{
		Name: "projectid=",
		Related: []string{
			"listProjectAccounts",
			"listProjects",
		},
	}

	apiFound := &config.API{
		Name: "registerIso",
		Verb: "register",
		Noun: "iso",
	}

	apiMap := map[string][]*config.API{
		"list": {
			{
				Name: "listProjectAccounts",
				Noun: "projectaccounts",
			},
			{
				Name: "listProjects",
				Noun: "projects",
			},
		},
	}

	result := findAutocompleteAPI(arg, apiFound, apiMap)

	if result == nil {
		t.Fatal("expected API, got nil")
	}

	if result.Name != "listProjects" {
		t.Fatalf("expected listProjects, got %s", result.Name)
	}
}

func TestMapFieldsFromDescription(t *testing.T) {
	description := "Example: tags[0].key=value&tags[0].value=value"

	fields := mapFieldsFromDescription(description)

	expected := []string{
		"tags[0].key=",
		"tags[0].value=",
	}

	if len(fields) != len(expected) {
		t.Fatalf("expected %d fields, got %d: %v", len(expected), len(fields), fields)
	}

	for i := range expected {
		if fields[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], fields[i])
		}
	}
}

func TestMapFieldsFromDescriptionWithRealMap(t *testing.T) {
	description := "VMware only: used to specify network mapping of a vApp VMware template registered \"as-is\". Example nicnetworklist[0].ip=Nic-101&nicnetworklist[0].network=uuid"

	fields := mapFieldsFromDescription(description)

	expected := []string{
		"nicnetworklist[0].ip=",
		"nicnetworklist[0].network=",
	}

	if len(fields) != len(expected) {
		t.Fatalf("expected %d fields, got %d: %v", len(expected), len(fields), fields)
	}

	for i := range expected {
		if fields[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], fields[i])
		}
	}
}

func TestMapFieldsFromDescriptionWithMultipleFields(t *testing.T) {
	description := "Disk offering details. Example: datadisksdetails[0].diskofferingid=id&datadisksdetails[0].size=10&datadisksdetails[0].miniops=100&datadisksdetails[0].maxiops=200"

	fields := mapFieldsFromDescription(description)

	expected := []string{
		"datadisksdetails[0].diskofferingid=",
		"datadisksdetails[0].size=",
		"datadisksdetails[0].miniops=",
		"datadisksdetails[0].maxiops=",
	}

	if len(fields) != len(expected) {
		t.Fatalf("expected %d fields, got %d: %v", len(expected), len(fields), fields)
	}

	for i := range expected {
		if fields[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], fields[i])
		}
	}
}

func TestMapFieldSuggestions(t *testing.T) {
	arg := &config.APIArg{
		Name:        "tags=",
		Type:        "map",
		Description: "Example: tags[0].key=value&tags[0].value=value",
	}

	suggestions := mapFieldSuggestions(arg)

	expected := []string{
		"tags[0].key=",
		"tags[0].value=",
	}

	if len(suggestions) != len(expected) {
		t.Fatalf("expected %d suggestions, got %d: %v", len(expected), len(suggestions), suggestions)
	}

	for i := range expected {
		if suggestions[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], suggestions[i])
		}
	}
}

func TestMapFieldSuggestionsWithGenericMap(t *testing.T) {
	arg := &config.APIArg{
		Name:        "tags",
		Type:        "map",
		Description: "List resources by tags (key/value pairs)",
	}

	expected := []string{
		"tags[0].key=",
		"tags[0].value=",
	}

	got := mapFieldSuggestions(arg)

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestMapAutocompleteOptions(t *testing.T) {
	arg := &config.APIArg{
		Name:        "tags",
		Type:        "map",
		Description: "List resources by tags (key/value pairs)",
	}

	line := "tags"

	suggestions := mapFieldSuggestions(arg)
	var options [][]rune

	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, line) {
			options = append(options, []rune(suggestion[len(line):]))
		}
	}

	expected := [][]rune{
		[]rune("[0].key="),
		[]rune("[0].value="),
	}

	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected %v, got %v", expected, options)
	}
}

func TestMapFieldsFromDescriptionWithMultipleIndexes(t *testing.T) {
	description := "Example: details[0].key=accesskey&details[0].value=value&details[1].key=secretkey&details[1].value=value"

	expected := []string{
		"details[0].key=",
		"details[0].value=",
		"details[1].key=",
		"details[1].value=",
	}

	fields := mapFieldsFromDescription(description)

	if !reflect.DeepEqual(fields, expected) {
		t.Fatalf("expected %v, got %v", expected, fields)
	}
}

func TestMapFieldSuggestionsWithNonKeyValueMap(t *testing.T) {
	arg := &config.APIArg{
		Name:        "storagecapabilities",
		Type:        "map",
		Description: "the storage pool capabilities",
	}

	if got := mapFieldSuggestions(arg); got != nil {
		t.Fatalf("expected no suggestions, got %v", got)
	}
}
