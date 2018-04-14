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
	"fmt"
)

// Command describes a CLI command
type Command struct {
	Name            string
	Help            string
	SubCommands     map[string][]string
	CustomCompleter func(input string, position int)
	Handle          func(*Request) error
}

var commands []*Command
var commandMap map[string]*Command

// FindCommand finds command handler for a command string
func FindCommand(name string) *Command {
	return commandMap[name]
}

// AllCommands returns all available commands
func AllCommands() []*Command {
	return commands
}

// AddCommand adds a command to internal list
func AddCommand(cmd *Command) {
	commands = append(commands, cmd)
	if commandMap == nil {
		commandMap = make(map[string]*Command)
	}
	commandMap[cmd.Name] = cmd
}

// PrintUsage prints help usage for a command
func PrintUsage() {
	commandHelp := ""
	for _, cmd := range commands {
		commandHelp += fmt.Sprintf("%s\t\t%s\n", cmd.Name, cmd.Help)
	}
	fmt.Printf(`usage: cmk [options] [commands]

Command Line Interface for Apache CloudStack

default commands:
%s

Try cmk [help]`, commandHelp)
}
