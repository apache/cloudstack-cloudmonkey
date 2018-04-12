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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode"
)

type ApiArg struct {
	Name        string
	Type        string
	Related     []string
	Description string
	Required    bool
	Length      int
}

type Api struct {
	Name         string
	Verb         string
	Noun         string
	Args         []*ApiArg
	RequiredArgs []string
	Related      []string
	Async        bool
	Description  string
	ResponseName string
}

var apiCache map[string]*Api
var apiVerbMap map[string][]*Api

func (c *Config) GetApiVerbMap() map[string][]*Api {
	if apiVerbMap != nil {
		return apiVerbMap
	}
	apiSplitMap := make(map[string][]*Api)
	for api := range apiCache {
		verb := apiCache[api].Verb
		apiSplitMap[verb] = append(apiSplitMap[verb], apiCache[api])
	}
	return apiSplitMap
}

func (c *Config) GetCache() map[string]*Api {
	if apiCache == nil {
		// read from disk?
		return make(map[string]*Api)
	}
	return apiCache
}

func LoadCache(c *Config) {
	cache, err := ioutil.ReadFile(c.CacheFile)
	if err != nil {
		fmt.Println("Please run sync. Failed to read the cache file: " + c.CacheFile)
		return
	}
	var data map[string]interface{}
	_ = json.Unmarshal(cache, &data)
	c.UpdateCache(data)
}

func (c *Config) SaveCache(response map[string]interface{}) {
	output, _ := json.Marshal(response)
	ioutil.WriteFile(c.CacheFile, output, 0600)
}

func (c *Config) UpdateCache(response map[string]interface{}) interface{} {
	apiCache = make(map[string]*Api)
	apiVerbMap = nil

	count := response["count"]
	apiList := response["api"].([]interface{})

	for _, node := range apiList {
		api, valid := node.(map[string]interface{})
		if !valid {
			//fmt.Println("Errro, moving on")
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

		var apiArgs []*ApiArg
		for _, argNode := range api["params"].([]interface{}) {
			apiArg, _ := argNode.(map[string]interface{})
			related := []string{}
			if apiArg["related"] != nil {
				related = strings.Split(apiArg["related"].(string), ",")
				sort.Strings(related)
			}
			apiArgs = append(apiArgs, &ApiArg{
				Name:        apiArg["name"].(string),
				Type:        apiArg["type"].(string),
				Required:    apiArg["required"].(bool),
				Related:     related,
				Description: apiArg["description"].(string),
			})
		}

		sort.Slice(apiArgs, func(i, j int) bool {
			return apiArgs[i].Name < apiArgs[j].Name
		})

		var requiredArgs []string
		for _, arg := range apiArgs {
			if arg.Required {
				requiredArgs = append(requiredArgs, arg.Name)
			}
		}

		apiCache[strings.ToLower(apiName)] = &Api{
			Name:         apiName,
			Verb:         verb,
			Noun:         noun,
			Args:         apiArgs,
			RequiredArgs: requiredArgs,
			Async:        isAsync,
			Description:  description,
		}
	}
	return count
}
