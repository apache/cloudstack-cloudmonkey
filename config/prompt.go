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
	"math/rand"
	"runtime"
	"strings"
	"time"
)

var emojis []string

func init() {
	rand.Seed(time.Now().Unix())
	emojis = strings.Split("🐶 🐹 🐰 🐻 🐼 🐨 🐯 🦁 🐷 🐙 🙈 🙉 🙊 🐒 🐔 🐧 🐦 🐤 🐣 🐥 🐺 🐗 🐴 🦄 🐝 🐛 🐌 🐞 🐜 🕷 🦂 🦀 🐍 🐢 🐠 🐟 🐡 🐬 🐳 🐋 🐅 🐃 🐂 🐄 🐘 🐐 🐑 🐎 🐖 🐀 🐓 🦃 🕊 🐕 🐩 🐈 🐇 🐿  🐲 🌵 🦍 🦊 🦌 🦏 🦇 🦅 🦆 🦉 🦈 🦐 🦑 🦋 🌴 🍀 🍂 🍁 🍄 🌍 ⛅️", " ")
}

func emoji() string {
	return emojis[rand.Intn(len(emojis)-1)]
}

func renderPrompt(prompt string) string {
	if runtime.GOOS == "windows" {
		return "cmk"
	}
	if prompt == "random" {
		return emoji()
	}
	return prompt
}

// GetPrompt returns prompt that the CLI should use
func (c *Config) GetPrompt() string {
	return fmt.Sprintf("(%s) %s > ", c.Core.ProfileName, renderPrompt(c.Core.Prompt))
}
