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
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

var name = "cloudmonkey"
var version = "6.0.0-alpha1"

func getDefaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".cmk")
}

func (c *Config) Name() string {
	return name
}

func (c *Config) Version() string {
	return version
}

func (c *Config) PrintHeader() {
	fmt.Printf("Apache CloudStack 🐵 cloudmonkey %s.\n", version)
	fmt.Printf("Type \"help\" for details, \"sync\" to update API cache or press tab to list commands.\n\n")
}

func (c *Config) GetPrompt() string {
	return fmt.Sprintf("(%s) \033[34m🐵\033[0m > ", c.ActiveProfile.Name)
}

func (c *Config) UpdateGlobalConfig(key string, value string) {
	c.UpdateConfig("", key, value)
}

func (c *Config) UpdateConfig(namespace string, key string, value string) {
	fmt.Println("Updating for key", key, ", value=", value, ", in ns=", namespace)
	if key == "profile" {
		//FIXME
		c.ActiveProfile.Name = value
	}
}
