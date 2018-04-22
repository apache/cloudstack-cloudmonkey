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

// Output formats
const (
	CSV   = "csv"
	JSON  = "json"
	XML   = "xml"
	TABLE = "table"
	TEXT  = "text"
)

// ServerProfile describes a management server
type ServerProfile struct {
	URL        string `ini:"url"`
	Username   string `ini:"username"`
	Password   string `ini:"password"`
	Domain     string `ini:"domain"`
	APIKey     string `ini:"apikey"`
	SecretKey  string `ini:"secretkey"`
	VerifyCert bool   `ini:"verifycert"`
}

// Core block describes common options for the CLI
type Core struct {
	AsyncBlock  bool   `ini:"asyncblock"`
	Timeout     int    `ini:"timeout"`
	Output      string `ini:"output"`
	ProfileName string `ini:"profile"`
}

// Config describes CLI config file and default options
type Config struct {
	Dir           string
	ConfigFile    string
	HistoryFile   string
	CacheFile     string
	LogFile       string
	Core          *Core
	ActiveProfile *ServerProfile
}

func getDefaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".cmk")
}

func defaultCoreConfig() Core {
	return Core{
		AsyncBlock:  false,
		Timeout:     1800,
		Output:      JSON,
		ProfileName: "local",
	}
}

func defaultProfile() ServerProfile {
	return ServerProfile{
		URL:        "http://localhost:8080/client/api",
		Username:   "admin",
		Password:   "password",
		Domain:     "/",
		APIKey:     "",
		SecretKey:  "",
		VerifyCert: false,
	}
}

func defaultConfig() *Config {
	configDir := getDefaultConfigDir()
	defaultCoreConfig := defaultCoreConfig()
	defaultProfile := defaultProfile()
	return &Config{
		Dir:           configDir,
		ConfigFile:    path.Join(configDir, "config"),
		CacheFile:     path.Join(configDir, "cache"),
		HistoryFile:   path.Join(configDir, "history"),
		LogFile:       path.Join(configDir, "log"),
		Core:          &defaultCoreConfig,
		ActiveProfile: &defaultProfile,
	}
}

var profiles []string

// GetProfiles returns list of available profiles
func GetProfiles() []string {
	return profiles
}

func reloadConfig(cfg *Config) *Config {

	if _, err := os.Stat(cfg.Dir); err != nil {
		os.Mkdir(cfg.Dir, 0700)
	}

	// Save on missing config
	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		defaultCoreConfig := defaultCoreConfig()
		defaultProfile := defaultProfile()
		conf := ini.Empty()
		conf.Section(ini.DEFAULT_SECTION).ReflectFrom(&defaultCoreConfig)
		conf.Section(defaultCoreConfig.ProfileName).ReflectFrom(&defaultProfile)
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
		defaultCore := defaultCoreConfig()
		section, _ := conf.NewSection(ini.DEFAULT_SECTION)
		section.ReflectFrom(&defaultCore)
		cfg.Core = &defaultCore
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
		activeProfile := defaultProfile()
		section, _ := conf.NewSection(cfg.Core.ProfileName)
		section.ReflectFrom(&activeProfile)
		cfg.ActiveProfile = &activeProfile
	} else {
		// Write
		if cfg.ActiveProfile != nil {
			conf.Section(cfg.Core.ProfileName).ReflectFrom(&cfg.ActiveProfile)
		}
		// Update
		profile := new(ServerProfile)
		conf.Section(cfg.Core.ProfileName).MapTo(profile)
		cfg.ActiveProfile = profile
	}
	// Save
	conf.SaveTo(cfg.ConfigFile)

	// Update available profiles list
	profiles = []string{}
	for _, profile := range conf.Sections() {
		if profile.Name() == ini.DEFAULT_SECTION {
			continue
		}
		profiles = append(profiles, profile.Name())
	}

	return cfg
}

// UpdateConfig updates and saves config
func (c *Config) UpdateConfig(key string, value string) {
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
		c.ActiveProfile = nil
	case "url":
		c.ActiveProfile.URL = value
	case "username":
		c.ActiveProfile.Username = value
	case "password":
		c.ActiveProfile.Password = value
	case "domain":
		c.ActiveProfile.Domain = value
	case "apikey":
		c.ActiveProfile.APIKey = value
	case "secretkey":
		c.ActiveProfile.SecretKey = value
	case "verifycert":
		c.ActiveProfile.VerifyCert = value == "true"
	}

	reloadConfig(c)
}

// NewConfig creates or reload config and loads API cache
func NewConfig() *Config {
	defaultConf := defaultConfig()
	defaultConf.Core = nil
	defaultConf.ActiveProfile = nil
	cfg := reloadConfig(defaultConf)
	LoadCache(cfg)
	return cfg
}
