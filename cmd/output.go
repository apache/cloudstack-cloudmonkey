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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/apache/cloudstack-cloudmonkey/config"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/olekukonko/tablewriter"
)

func printJSON(response map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(response)
}

func printTable(response map[string]interface{}) {
	table := tablewriter.NewWriter(os.Stdout)
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			fmt.Printf("%s:\n", k)
			var header []string
			for _, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}

				if len(header) == 0 {
					for field := range row {
						header = append(header, field)
					}
					sort.Strings(header)
					table.SetHeader(header)
				}
				var rowArray []string
				for _, field := range header {
					rowArray = append(rowArray, fmt.Sprintf("%v", row[field]))
				}
				table.Append(rowArray)
			}
		} else {
			fmt.Printf("%s = %v\n", k, v)
		}
	}
	table.Render()
}

func printText(response map[string]interface{}) {
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice {
			fmt.Printf("%s:\n", k)
			for idx, item := range v.([]interface{}) {
				if idx > 0 {
					fmt.Println("================================================================================")
				}
				row, isMap := item.(map[string]interface{})
				if isMap {
					for field, value := range row {
						fmt.Printf("%s = %v\n", field, value)
					}
				} else {
					fmt.Printf("%v\n", item)
				}
			}
		} else {
			fmt.Printf("%s = %v\n", k, v)
		}
	}
}

func printColumn(response map[string]interface{}) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	for _, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			var header []string
			for idx, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}

				if idx == 0 {
					for rk := range row {
						header = append(header, strings.ToUpper(rk))
					}
					sort.Strings(header)
					fmt.Fprintln(w, strings.Join(header, "\t"))
				}
				var values []string
				for _, key := range header {
					values = append(values, fmt.Sprintf("%v", row[strings.ToLower(key)]))
				}
				fmt.Fprintln(w, strings.Join(values, "\t"))
			}
		}
	}
	w.Flush()
}

func printCsv(response map[string]interface{}) {
	for _, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			var header []string
			for idx, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}

				if idx == 0 {
					for rk := range row {
						header = append(header, rk)
					}
					sort.Strings(header)
					fmt.Println(strings.Join(header, ","))
				}
				var values []string
				for _, key := range header {
					values = append(values, fmt.Sprintf("%v", row[key]))
				}
				fmt.Println(strings.Join(values, ","))
			}

		}
	}
}

func filterResponse(response map[string]interface{}, filter []string) map[string]interface{} {
	if filter == nil || len(filter) == 0 {
		return response
	}
	filteredResponse := make(map[string]interface{})
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			var filteredRows []interface{}
			for _, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}
				filteredRow := make(map[string]interface{})
				for _, filterKey := range filter {
					for field := range row {
						if filterKey == field {
							filteredRow[field] = row[field]
						}
					}
				}
				filteredRows = append(filteredRows, filteredRow)
			}
			filteredResponse[k] = filteredRows
		} else {
			filteredResponse[k] = v
			continue
		}

	}
	return filteredResponse
}

func printResult(outputType string, response map[string]interface{}, filter []string) {
	response = filterResponse(response, filter)
	switch outputType {
	case config.JSON:
		printJSON(response)
	case config.TABLE:
		printTable(response)
	case config.TEXT:
		printText(response)
	case config.COLUMN:
		printColumn(response)
	case config.CSV:
		printCsv(response)
	default:
		fmt.Println("Invalid output type configured, please fix that!")
	}
}

func queryResponse(response map[string]interface{}, query []string) map[string]interface{} {
	if query == nil || len(query) == 0 {
		return response
	}
	queriedResponse := make(map[string]interface{})
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			var queriedRows []interface{}
			for _, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}
				queriedRow := make(map[string]interface{})
				for _, queryKey := range query {
					for field := range row {
						if queryKey == field {
							queriedRow[field] = row[field]
						}
					}
				}
				queriedRows = append(queriedRows, queriedRow)
			}
			queriedResponse[k] = queriedRows
		} else {
			queriedResponse[k] = v
			continue
		}

	}
	return queriedResponse
}

func printResultJmespath(outputType string, response map[string]interface{}, query []string) {
	response = queryResponse(response, query)
	switch outputType {
	case config.JSON:
		printJSON(response)
	case config.TABLE:
		printTable(response)
	case config.TEXT:
		printText(response)
	case config.COLUMN:
		printColumn(response)
	case config.CSV:
		printCsv(response)
	default:
		fmt.Println("Invalid output type configured, please fix that!")
	}
}
