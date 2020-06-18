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
	"github.com/olekukonko/tablewriter"
)

func jsonify(value interface{}) string {
	if value == nil {
		return ""
	}
	if reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice {
		jsonStr, err := json.Marshal(value)
		if err == nil {
			value = string(jsonStr)
		}
	}
	return fmt.Sprintf("%v", value)
}

func printJSON(response map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(response)
}

func printText(response map[string]interface{}) {
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice {
			fmt.Printf("%v:\n", k)
			for idx, item := range v.([]interface{}) {
				if idx > 0 {
					fmt.Println("================================================================================")
				}
				row, isMap := item.(map[string]interface{})
				if isMap {
					for field, value := range row {
						fmt.Printf("%s = %v\n", field, jsonify(value))
					}
				} else {
					fmt.Printf("%v\n", item)
				}
			}
		} else {
			fmt.Printf("%v = %v\n", k, jsonify(v))
		}
	}
}

func printTable(response map[string]interface{}, filter []string) {
	table := tablewriter.NewWriter(os.Stdout)
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice {
			items, ok := v.([]interface{})
			if !ok {
				continue
			}
			fmt.Printf("%v:\n", k)
			var header []string
			for _, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) < 1 {
					continue
				}
				if len(header) == 0 {
					if len(filter) > 0 {
						header = filter
					} else {
						for field := range row {
							header = append(header, field)
						}
						sort.Strings(header)
					}
					table.SetHeader(header)
				}
				var rowArray []string
				for _, field := range header {
					rowArray = append(rowArray, jsonify(row[field]))
				}
				table.Append(rowArray)
			}
		} else {
			fmt.Printf("%v = %v\n", k, v)
		}
	}
	table.Render()
}

func printColumn(response map[string]interface{}, filter []string) {
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
					if len(filter) > 0 {
						header = filter
					} else {
						for rk := range row {
							header = append(header, strings.ToUpper(rk))
						}
						sort.Strings(header)
					}
					fmt.Fprintln(w, strings.Join(header, "\t"))
				}
				var values []string
				for _, key := range header {
					values = append(values, jsonify(row[strings.ToLower(key)]))
				}
				fmt.Fprintln(w, strings.Join(values, "\t"))
			}
		}
	}
	w.Flush()
}

func printCsv(response map[string]interface{}, filter []string) {
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
					if len(filter) > 0 {
						header = filter
					} else {
						for rk := range row {
							header = append(header, rk)
						}
						sort.Strings(header)
					}
					fmt.Println(strings.Join(header, ","))
				}
				var values []string
				for _, key := range header {
					values = append(values, jsonify(row[key]))
				}
				fmt.Println(strings.Join(values, ","))
			}
		}
	}
}

func filterResponse(response map[string]interface{}, filter []string, outputType string) map[string]interface{} {
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
					if outputType == config.COLUMN || outputType == config.CSV || outputType == config.TABLE {
						if _, ok := filteredRow[filterKey]; !ok {
							filteredRow[filterKey] = ""
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
	response = filterResponse(response, filter, outputType)
	switch outputType {
	case config.JSON:
		printJSON(response)
	case config.TEXT:
		printText(response)
	case config.COLUMN:
		printColumn(response, filter)
	case config.CSV:
		printCsv(response, filter)
	case config.TABLE:
		printTable(response, filter)
	default:
		fmt.Println("Invalid output type configured, please fix that!")
	}
}
