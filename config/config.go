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
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gofrs/flock"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

// Output formats
const (
	COLUMN = "column"
	CSV    = "csv"
	JSON   = "json"
	TABLE  = "table"
	TEXT   = "text"
)

// ServerProfile describes a management server
type ServerProfile struct {
	URL       string       `ini:"url"`
	Username  string       `ini:"username"`
	Password  string       `ini:"password"`
	Domain    string       `ini:"domain"`
	APIKey    string       `ini:"apikey"`
	SecretKey string       `ini:"secretkey"`
	Client    *http.Client `ini:"-"`
}

// Core block describes common options for the CLI
type Core struct {
	Prompt      string `ini:"prompt"`
	AsyncBlock  bool   `ini:"asyncblock"`
	Timeout     int    `ini:"timeout"`
	Output      string `ini:"output"`
	VerifyCert  bool   `ini:"verifycert"`
	ProfileName string `ini:"profile"`
}

// Config describes CLI config file and default options
type Config struct {
	Dir           string
	ConfigFile    string
	HistoryFile   string
	CacheFile     string
	LogFile       string
	HasShell      bool
	Core          *Core
	ActiveProfile *ServerProfile
}

func getDefaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cmkHome := path.Join(home, ".cmk")
	if _, err := os.Stat(cmkHome); os.IsNotExist(err) {
		err := os.Mkdir(cmkHome, 0700)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	return cmkHome
}

func defaultCoreConfig() Core {
	return Core{
		Prompt:      "🐱",
		AsyncBlock:  true,
		Timeout:     1800,
		Output:      JSON,
		VerifyCert:  true,
		ProfileName: "localcloud",
	}
}

func defaultProfile() ServerProfile {
	return ServerProfile{
		URL:       "http://localhost:8080/client/api",
		Username:  "admin",
		Password:  "password",
		Domain:    "/",
		APIKey:    "",
		SecretKey: "",
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
		HasShell:      false,
		Core:          &defaultCoreConfig,
		ActiveProfile: &defaultProfile,
	}
}

var profiles []string

// GetProfiles returns list of available profiles
func GetProfiles() []string {
	return profiles
}

func newHTTPClient(cfg *Config) *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !cfg.Core.VerifyCert},
		},
	}
	client.Timeout = time.Duration(time.Duration(cfg.Core.Timeout) * time.Second)
	return client
}

func reloadConfig(cfg *Config) *Config {
	fileLock := flock.New(path.Join(getDefaultConfigDir(), "lock"))
	err := fileLock.Lock()
	if err != nil {
		fmt.Println("Failed to grab config file lock, please try again")
		return cfg
	}
	cfg = saveConfig(cfg)
	fileLock.Unlock()
	return cfg
}

func saveConfig(cfg *Config) *Config {
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

	cfg.ActiveProfile.Client = newHTTPClient(cfg)
	return cfg
}

// UpdateConfig updates and saves config
func (c *Config) UpdateConfig(key string, value string, update bool) {
	switch key {
	case "prompt":
		c.Core.Prompt = value
	case "asyncblock":
		c.Core.AsyncBlock = value == "true"
	case "display":
		fallthrough
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
		c.Core.VerifyCert = value == "true"
	case "debug":
		if value == "true" {
			EnableDebugging()
		} else {
			DisableDebugging()
		}
	}

	Debug("UpdateConfig key:", key, " value:", value, " update:", update)

	if update {
		reloadConfig(c)
	}
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
