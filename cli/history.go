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
	"bufio"
	"fmt"
	"os"
)

func initHistory() {
	if _, err := os.Stat(cfg.HistoryFile); os.IsNotExist(err) {
		os.OpenFile(cfg.HistoryFile, os.O_RDONLY|os.O_CREATE, 0600)
	}
}

func readHistory() []string {
	initHistory()
	file, err := os.Open(cfg.HistoryFile)
	if err != nil {
		fmt.Println("Failed to open history file:", err)
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Println("Failed to read history:", scanner.Err())
	}
	return lines
}

func writeHistory(in string) {
	file, err := os.OpenFile(cfg.HistoryFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Failed to open history file:", err)
	}
	defer file.Close()

	if _, err = file.WriteString(in + "\n"); err != nil {
		fmt.Println("Failed to write history:", err)
	}
}
