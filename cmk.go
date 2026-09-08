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
	"strconv"
	"strings"

	"github.com/apache/cloudstack-cloudmonkey/cli"
	"github.com/apache/cloudstack-cloudmonkey/cmd"
	"github.com/apache/cloudstack-cloudmonkey/config"
)

// GitSHA holds the git SHA
var GitSHA string

// BuildDate holds the build datetime
var BuildDate string

func init() {
	flag.Usage = func() {
		cmd.PrintUsage()
	}
}

func main() {
	validFormats := strings.Join(config.GetOutputFormats(), ",")
	outputFormat := flag.String("o", "", "output format: "+validFormats)
	showVersion := flag.Bool("v", false, "show version")
	debug := flag.Bool("d", false, "enable debug mode")
	profile := flag.String("p", "", "server profile")
	configFilePath := flag.String("c", "", "config file path")
	acsURL := flag.String("u", config.DefaultACSAPIEndpoint, "cloudStack's API endpoint URL")
	apiKey := flag.String("k", "", "cloudStack user's API key")
	secretKey := flag.String("s", "", "cloudStack user's secret key")
	flag.Parse()
	args := flag.Args()

	// Fall back to environment variables for flags not passed on the
	// command line; CLI flags take precedence over environment variables.
	passedFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		passedFlags[f.Name] = true
	})
	fallbackToEnvVar := func(flagName string, flagValue *string, envVar string) {
		if !passedFlags[flagName] {
			if value := os.Getenv(envVar); value != "" {
				*flagValue = value
			}
		}
	}
	fallbackToEnvVar("c", configFilePath, config.ConfigFileEnvVar)
	fallbackToEnvVar("p", profile, config.ProfileEnvVar)
	fallbackToEnvVar("u", acsURL, config.URLEnvVar)
	fallbackToEnvVar("k", apiKey, config.APIKeyEnvVar)
	fallbackToEnvVar("s", secretKey, config.SecretKeyEnvVar)
	fallbackToEnvVar("o", outputFormat, config.OutputEnvVar)
	if !passedFlags["d"] {
		if value, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(config.DebugEnvVar))); err == nil {
			*debug = value
		}
	}

	cfg := config.NewConfig(configFilePath)

	if *showVersion {
		fmt.Printf("%s %s (build: %s, %s)\n", cfg.Name(), cfg.Version(), GitSHA, BuildDate)
		os.Exit(0)
	}

	if *debug {
		config.EnableDebugging()
	}

	if *outputFormat != "" {
		if !config.CheckIfValuePresent(config.GetOutputFormats(), *outputFormat) {
			fmt.Println("Invalid value set for output format. Supported values: " + validFormats)
			os.Exit(1)
		}
		cfg.UpdateConfig("output", *outputFormat, false)
	}

	if *acsURL != config.DefaultACSAPIEndpoint {
		cfg.UpdateConfig("url", *acsURL, false)
	}

	if *apiKey != "" {
		cfg.UpdateConfig("apikey", *apiKey, false)
	}

	if *secretKey != "" {
		cfg.UpdateConfig("secretkey", *secretKey, false)
	}

	if *profile != "" {
		cfg.LoadProfile(*profile, true)
	}
	config.LoadCache(cfg)
	cli.SetConfig(cfg)

	config.Debug("cmdline args:", strings.Join(os.Args, ", "))
	if len(args) > 0 {
		if err := cli.ExecCmd(args, (*apiKey != "" || *secretKey != "")); err != nil {
			fmt.Println("🙈 Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	cli.ExecPrompt()
}
