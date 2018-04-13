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
	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
	"os"
	"path"
	"strconv"
)

const (
	Json  = "json"
	Xml   = "xml"
	Table = "table"
	Text  = "text"
)

type ServerProfile struct {
	Url        string `ini:"url"`
	Username   string `ini:"username"`
	Password   string `ini:"password"`
	Domain     string `ini:"domain"`
	ApiKey     string `ini:"apikey"`
	SecretKey  string `ini:"secretkey"`
	VerifyCert bool   `ini:"verifycert"`
}

type Core struct {
	AsyncBlock    bool           `ini:"asyncblock"`
	Timeout       int            `ini:"timeout"`
	Output        string         `ini:"output"`
	ProfileName   string         `ini:"profile"`
	ActiveProfile *ServerProfile `ini:"-"`
}

type Config struct {
	Dir         string
	ConfigFile  string
	HistoryFile string
	CacheFile   string
	LogFile     string
	Core        *Core
}

func getDefaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".cmk")
}

func defaultConfig() *Config {
	configDir := getDefaultConfigDir()
	return &Config{
		Dir:         configDir,
		ConfigFile:  path.Join(configDir, "config"),
		CacheFile:   path.Join(configDir, "cache"),
		HistoryFile: path.Join(configDir, "history"),
		LogFile:     path.Join(configDir, "log"),
		Core: &Core{
			AsyncBlock:  false,
			Timeout:     1800,
			Output:      Json,
			ProfileName: "local",
			ActiveProfile: &ServerProfile{
				Url:        "http://localhost:8080/client/api",
				Username:   "admin",
				Password:   "password",
				Domain:     "/",
				ApiKey:     "",
				SecretKey:  "",
				VerifyCert: false,
			},
		},
	}
}

func reloadConfig(cfg *Config) *Config {

	if _, err := os.Stat(cfg.Dir); err != nil {
		os.Mkdir(cfg.Dir, 0700)
	}

	// Save on missing config
	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		defaultConf := defaultConfig()
		conf := ini.Empty()
		conf.Section(ini.DEFAULT_SECTION).ReflectFrom(defaultConf.Core)
		conf.Section(cfg.Core.ProfileName).ReflectFrom(defaultConf.Core.ActiveProfile)
		conf.SaveTo(cfg.ConfigFile)
	}

	// Read config
	conf, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, cfg.ConfigFile)

	if err != nil {
		fmt.Printf("Fail to read config file: %v", err)
		os.Exit(1)
	}

	core, err := conf.GetSection(ini.DEFAULT_SECTION)
	if core == nil {
		section, _ := conf.NewSection(ini.DEFAULT_SECTION)
		section.ReflectFrom(&defaultConfig().Core)
	} else {
		// Write
		if cfg.Core != nil {
			conf.Section(ini.DEFAULT_SECTION).ReflectFrom(&cfg.Core)
		}
		// Update
		core := new(Core)
		conf.Section(ini.DEFAULT_SECTION).MapTo(core)
		cfg.Core = core
	}

	profile, err := conf.GetSection(cfg.Core.ProfileName)
	if profile == nil {
		section, _ := conf.NewSection(cfg.Core.ProfileName)
		section.ReflectFrom(&defaultConfig().Core.ActiveProfile)
	} else {
		// Write
		if cfg.Core.ActiveProfile != nil {
			conf.Section(cfg.Core.ProfileName).ReflectFrom(&cfg.Core.ActiveProfile)
		}
		// Update
		profile := new(ServerProfile)
		conf.Section(cfg.Core.ProfileName).MapTo(profile)
		cfg.Core.ActiveProfile = profile
	}
	// Save
	conf.SaveTo(cfg.ConfigFile)

	fmt.Println("Updating config to:", cfg.Core, cfg.Core.ActiveProfile)
	return cfg
}

func (c *Config) UpdateGlobalConfig(key string, value string) {
	c.UpdateConfig("", key, value)
}

func (c *Config) UpdateConfig(namespace string, key string, value string) {
	switch key {
	case "asyncblock":
		c.Core.AsyncBlock = value == "true"
	case "output":
		c.Core.Output = value
	case "timeout":
		intValue, _ := strconv.Atoi(value)
		c.Core.Timeout = intValue
	case "profile":
		c.Core.ProfileName = value
		c.Core.ActiveProfile = nil
	case "url":
		c.Core.ActiveProfile.Url = value
	case "username":
		c.Core.ActiveProfile.Username = value
	case "password":
		c.Core.ActiveProfile.Password = value
	case "domain":
		c.Core.ActiveProfile.Domain = value
	case "apikey":
		c.Core.ActiveProfile.ApiKey = value
	case "secretkey":
		c.Core.ActiveProfile.SecretKey = value
	case "verifycert":
		c.Core.ActiveProfile.VerifyCert = value == "true"
	default:
		return
	}

	reloadConfig(c)
}

func NewConfig() *Config {
	defaultConf := defaultConfig()
	defaultConf.Core = nil
	cfg := reloadConfig(defaultConf)
	LoadCache(cfg)
	return cfg
}
