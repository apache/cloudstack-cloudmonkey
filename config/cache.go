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

package config

import (
	_ "embed" // for embedding apis.json
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"unicode"
)

// FAKE is used for fake CLI only options like filter=
const (
	FAKE        = "fake"
	FilePathArg = "filepath="
)

//go:embed apis.json
var bundledAPICache []byte

// APIArg are the args passable to an API
type APIArg struct {
	Name        string
	Type        string
	Related     []string
	Description string
	Required    bool
	Length      int
}

// API describes a CloudStack API
type API struct {
	Name         string
	Verb         string
	Noun         string
	Args         []*APIArg
	RequiredArgs []string
	FakeArgs     []string
	Related      []string
	Async        bool
	Description  string
	ResponseKeys []string
}

var apiCache map[string]*API
var apiVerbMap map[string][]*API

// GetAPIVerbMap returns API cache by verb
func (c *Config) GetAPIVerbMap() map[string][]*API {
	if apiVerbMap != nil {
		return apiVerbMap
	}
	apiSplitMap := make(map[string][]*API)
	for api := range apiCache {
		verb := apiCache[api].Verb
		apiSplitMap[verb] = append(apiSplitMap[verb], apiCache[api])
	}
	return apiSplitMap
}

// GetCache returns API cache by full API name
func (c *Config) GetCache() map[string]*API {
	if apiCache == nil {
		// read from disk?
		return make(map[string]*API)
	}
	return apiCache
}

// LoadCache loads cache using the default cache file
func LoadCache(c *Config) interface{} {
	cacheFile := c.CacheFile()
	Debug("Trying to read API cache from:", cacheFile)
	cache, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loaded in-built API cache. Failed to read API cache, please run 'sync'.\n")
		cache = bundledAPICache
	}
	var data map[string]interface{}
	_ = json.Unmarshal(cache, &data)
	return c.UpdateCache(data)
}

// SaveCache saves received auto-discovery data to cache file
func (c *Config) SaveCache(response map[string]interface{}) {
	output, _ := json.Marshal(response)
	ioutil.WriteFile(c.CacheFile(), output, 0600)
}

// UpdateCache uses auto-discovery data to update internal API cache
func (c *Config) UpdateCache(response map[string]interface{}) interface{} {
	apiCache = make(map[string]*API)
	apiVerbMap = nil

	count := response["count"]
	apiList := response["api"].([]interface{})

	for _, node := range apiList {
		api, valid := node.(map[string]interface{})
		if !valid {
			fmt.Println("Error, moving on...")
			continue
		}
		apiName := api["name"].(string)
		isAsync := api["isasync"].(bool)
		description := api["description"].(string)

		idx := 0
		for _, chr := range apiName {
			if unicode.IsLower(chr) {
				idx++
			} else {
				break
			}
		}
		verb := apiName[:idx]
		noun := strings.ToLower(apiName[idx:])

		var apiArgs []*APIArg
		for _, argNode := range api["params"].([]interface{}) {
			apiArg, _ := argNode.(map[string]interface{})
			related := []string{}
			if apiArg["related"] != nil {
				related = strings.Split(apiArg["related"].(string), ",")
				sort.Strings(related)
			}
			apiArgs = append(apiArgs, &APIArg{
				Name:        apiArg["name"].(string) + "=",
				Type:        apiArg["type"].(string),
				Required:    apiArg["required"].(bool),
				Related:     related,
				Description: apiArg["description"].(string),
			})
		}

		// Add filter arg
		fakeArg := &APIArg{
			Name:        "filter=",
			Type:        FAKE,
			Description: "cloudmonkey specific response key filtering",
		}
		apiArgs = append(apiArgs, fakeArg)
		fakeArgs := []string{fakeArg.Name}

		// Add exclude arg
		fakeArg = &APIArg{
			Name:        "exclude=",
			Type:        FAKE,
			Description: "cloudmonkey specific response key to exlude when filtering",
		}
		apiArgs = append(apiArgs, fakeArg)
		fakeArgs = append(fakeArgs, fakeArg.Name)

		if IsFileUploadAPI(apiName) {
			fakeArg = &APIArg{
				Name:        FilePathArg,
				Type:        FAKE,
				Description: "cloudmonkey specific key to upload files for supporting APIs. Comma-separated list of file paths can be provided",
			}
			apiArgs = append(apiArgs, fakeArg)
			fakeArgs = append(fakeArgs, fakeArg.Name)
		}

		sort.Slice(apiArgs, func(i, j int) bool {
			return apiArgs[i].Name < apiArgs[j].Name
		})

		var responseKeys []string
		for _, respNode := range api["response"].([]interface{}) {
			if resp, ok := respNode.(map[string]interface{}); ok {
				if resp == nil || resp["name"] == nil {
					continue
				}
				responseKeys = append(responseKeys, fmt.Sprintf("%v,", resp["name"]))
			}
		}
		sort.Strings(responseKeys)

		var requiredArgs []string
		for _, arg := range apiArgs {
			if arg.Required {
				requiredArgs = append(requiredArgs, arg.Name)
			}
		}

		apiCache[strings.ToLower(apiName)] = &API{
			Name:         apiName,
			Verb:         verb,
			Noun:         noun,
			Args:         apiArgs,
			RequiredArgs: requiredArgs,
			FakeArgs:     fakeArgs,
			Async:        isAsync,
			Description:  description,
			ResponseKeys: responseKeys,
		}
	}
	return count
}

// IsFileUploadAPI checks if the provided API name corresponds to a file upload-related API.
// It returns true if the API name matches one of the following (case-insensitive):
// "getUploadParamsForIso", "getUploadParamsForVolume", or "getUploadParamsForTemplate".
func IsFileUploadAPI(api string) bool {
	apiName := strings.ToLower(api)
	return apiName == "getuploadparamsforiso" ||
		apiName == "getuploadparamsforvolume" ||
		apiName == "getuploadparamsfortemplate"
}
