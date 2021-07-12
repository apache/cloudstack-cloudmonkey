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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/internal/app"
	"github.com/apache/cloudstack-cloudmonkey/internal/cli"
	"github.com/apache/cloudstack-cloudmonkey/internal/config"
)

// GitSHA holds the git SHA
var GitSHA string

// BuildDate holds the build datetime
var BuildDate string

func init() {
	flag.Usage = func() {
		app.PrintUsage()
	}
}

func main() {
	outputFormat := flag.String("o", "", "output format: json, text, table, column, csv")
	showVersion := flag.Bool("v", false, "show version")
	debug := flag.Bool("d", false, "enable debug mode")
	profile := flag.String("p", "", "server profile")
	configFilePath := flag.String("c", "", "config file path")

	flag.Parse()

	cfg := config.NewConfig(configFilePath)

	if *showVersion {
		fmt.Printf("%s %s (build: %s, %s)\n", cfg.Name(), cfg.Version(), GitSHA, BuildDate)
		os.Exit(0)
	}

	if *debug {
		config.EnableDebugging()
	}

	if *outputFormat != "" {
		cfg.UpdateConfig("output", *outputFormat, false)
	}

	if *profile != "" {
		cfg.LoadProfile(*profile)
	}

	cli.SetConfig(cfg)
	args := flag.Args()
	config.Debug("appline args:", strings.Join(os.Args, ", "))
	if len(args) > 0 {
		if err := cli.ExecCmd(args); err != nil {
			fmt.Println("🙈 Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	cli.ExecPrompt()
}
