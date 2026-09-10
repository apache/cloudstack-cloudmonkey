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
