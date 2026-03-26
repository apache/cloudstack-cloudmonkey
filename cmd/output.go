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
	"encoding/csv"
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

func jsonify(value interface{}, format string) string {
	if value == nil {
		return ""
	}
	if reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice {
		var jsonStr []byte
		var err error
		if format == "text" {
			jsonStr, err = json.MarshalIndent(value, "", "")
		} else {
			jsonStr, err = json.Marshal(value)
		}
		if err == nil {
			value = string(jsonStr)
		}
	}
	switch value.(type) {
	case float64, float32:
		return fmt.Sprintf("%.f", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func printJSON(response map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(response)
}

func getItemsFromValue(v interface{}) ([]interface{}, reflect.Kind, bool) {
	valueKind := reflect.TypeOf(v).Kind()
	if valueKind == reflect.Slice {
		sliceItems, ok := v.([]interface{})
		if !ok {
			return nil, valueKind, false
		}
		return sliceItems, valueKind, true
	} else if valueKind == reflect.Map {
		mapItem, ok := v.(map[string]interface{})
		if !ok {
			return nil, valueKind, false
		}
		return []interface{}{mapItem}, valueKind, true
	}
	return nil, valueKind, false
}

func printText(response map[string]interface{}) {
	format := "text"
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, _, ok := getItemsFromValue(v)
			if ok {
				fmt.Printf("%v:\n", k)
				for idx, item := range items {
					if idx > 0 {
						fmt.Println("================================================================================")
					}
					row, isMap := item.(map[string]interface{})
					if isMap {
						for field, value := range row {
							fmt.Printf("%s = %v\n", field, jsonify(value, format))
						}
					} else {
						fmt.Printf("%v\n", item)
					}
				}
				return
			}
		}
		fmt.Printf("%v = %v\n", k, jsonify(v, format))
	}
}

func printTable(response map[string]interface{}, filter []string) {
	format := "table"
	table := tablewriter.NewWriter(os.Stdout)
	for k, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, _, ok := getItemsFromValue(v)
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
					rowArray = append(rowArray, jsonify(row[field], format))
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
	format := "column"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	for _, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, _, ok := getItemsFromValue(v)
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
					values = append(values, jsonify(row[strings.ToLower(key)], format))
				}
				fmt.Fprintln(w, strings.Join(values, "\t"))
			}
		}
	}
	w.Flush()
}

func printCsv(response map[string]interface{}, filter []string) {
	format := "csv"
	enc := csv.NewWriter(os.Stdout)
	for _, v := range response {
		valueType := reflect.TypeOf(v)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, _, ok := getItemsFromValue(v)
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
					enc.Write(header)
				}
				var values []string
				for _, key := range header {
					values = append(values, jsonify(row[key], format))
				}
				enc.Write(values)
			}
		}
	}
	enc.Flush()
}

func filterResponse(response map[string]interface{}, filter []string, excludeFilter []string, outputType string) map[string]interface{} {
	if len(filter) == 0 && len(excludeFilter) == 0 {
		return response
	}

	excludeSet := make(map[string]struct{}, len(excludeFilter))
	for _, key := range excludeFilter {
		excludeSet[key] = struct{}{}
	}

	filterSet := make(map[string]struct{}, len(filter))
	for _, key := range filter {
		filterSet[key] = struct{}{}
	}

	filteredResponse := make(map[string]interface{})

	for key, value := range response {
		valueType := reflect.TypeOf(value)
		if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Map {
			items, originalKind, ok := getItemsFromValue(value)
			if !ok {
				continue
			}
			var filteredRows []interface{}
			for _, item := range items {
				row, ok := item.(map[string]interface{})
				if !ok || len(row) == 0 {
					continue
				}

				filteredRow := make(map[string]interface{})

				if len(filter) > 0 {
					// Include only keys that exist in filterSet
					for filterKey := range filterSet {
						if val, exists := row[filterKey]; exists {
							filteredRow[filterKey] = val
						} else if outputType == config.COLUMN || outputType == config.CSV || outputType == config.TABLE {
							filteredRow[filterKey] = "" // Ensure all filter keys exist in row
						}
					}
				} else {
					// Exclude keys from excludeFilter
					for field, val := range row {
						if _, excluded := excludeSet[field]; !excluded {
							filteredRow[field] = val
						}
					}
				}

				// Skip completely empty rows after filtering
				if len(filteredRow) == 0 {
					continue
				}

				filteredRows = append(filteredRows, filteredRow)
			}

			// If no non-empty rows remain for this key, omit the key entirely
			if len(filteredRows) == 0 {
				continue
			}

			if originalKind == reflect.Map && len(filteredRows) > 0 {
				filteredResponse[key] = filteredRows[0]
			} else {
				filteredResponse[key] = filteredRows
			}
		} else {
			filteredResponse[key] = value
		}
	}

	return filteredResponse
}

func printResult(outputType string, response map[string]interface{}, filter []string, excludeFilter []string) {
	response = filterResponse(response, filter, excludeFilter, outputType)
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
	case config.DEFAULT:
		printJSON(response)
	default:
		fmt.Println("Invalid output type configured, please fix that!")
	}
}
