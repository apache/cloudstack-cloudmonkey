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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofrs/flock"
	homedir "github.com/mitchellh/go-homedir"
	ini "gopkg.in/ini.v1"
)

// Output formats
const (
	COLUMN  = "column"
	CSV     = "csv"
	JSON    = "json"
	TABLE   = "table"
	TEXT    = "text"
	DEFAULT = "default"
)

const DEFAULT_ACS_API_ENDPOINT = "http://localhost:8080/client/api"

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
	Prompt       string `ini:"prompt"`
	AsyncBlock   bool   `ini:"asyncblock"`
	Timeout      int    `ini:"timeout"`
	Output       string `ini:"output"`
	VerifyCert   bool   `ini:"verifycert"`
	ProfileName  string `ini:"profile"`
	AutoComplete bool   `ini:"autocomplete"`
	PostRequest  bool   `ini:postrequest`
}

// Config describes CLI config file and default options
type Config struct {
	Dir           string
	ConfigFile    string
	HistoryFile   string
	LogFile       string
	HasShell      bool
	Core          *Core
	ActiveProfile *ServerProfile
	Context       *context.Context
	Cancel        context.CancelFunc
	C             chan bool
}

func GetOutputFormats() []string {
	return []string{"column", "csv", "json", "table", "text", "default"}
}

func CheckIfValuePresent(dataset []string, element string) bool {
	for _, arg := range dataset {
		if arg == element {
			return true
		}
	}
	return false
}

// CacheFile returns the path to the cache file for a server profile
func (c Config) CacheFile() string {
	cacheDir := path.Join(c.Dir, "profiles")
	cacheFileName := "cache"
	if c.Core != nil && len(c.Core.ProfileName) > 0 {
		cacheFileName = c.Core.ProfileName + ".cache"
	}
	checkAndCreateDir(cacheDir)
	return path.Join(cacheDir, cacheFileName)
}

func hasAccess(path string) bool {
	status := true
	file, err := os.Open(path)
	if err != nil {
		status = false
	}
	file.Close()
	return status
}

func checkAndCreateDir(path string) string {
	if fileInfo, err := os.Stat(path); os.IsNotExist(err) || !fileInfo.IsDir() {
		err := os.Mkdir(path, 0700)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	isAccsessible := hasAccess(path)
	if !isAccsessible {
		fmt.Println("User isn't authorized to access the config file")
		os.Exit(1)
	}
	return path
}

func getDefaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return checkAndCreateDir(path.Join(home, ".cmk"))
}

func defaultCoreConfig() Core {
	return Core{
		Prompt:       "🐱",
		AsyncBlock:   true,
		Timeout:      1800,
		Output:       JSON,
		VerifyCert:   true,
		ProfileName:  "localcloud",
		AutoComplete: true,
		PostRequest:  true,
	}
}

func defaultProfile() ServerProfile {
	return ServerProfile{
		URL:       DEFAULT_ACS_API_ENDPOINT,
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

func SetupContext(cfg *Config) {
	cfg.C = make(chan bool)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	cfg.Context = &ctx
	cfg.Cancel = cancel
	go func() {
		<-signals
		cfg.Cancel()
		cfg.C <- true
	}()
}

func newHTTPClient(cfg *Config) *http.Client {
	SetupContext(cfg)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !cfg.Core.VerifyCert},
		},
	}
	client.Timeout = time.Duration(time.Duration(cfg.Core.Timeout) * time.Second)
	return client
}

func setActiveProfile(cfg *Config, profile *ServerProfile) {
	cfg.ActiveProfile = profile
	cfg.ActiveProfile.Client = newHTTPClient(cfg)
}

func reloadConfig(cfg *Config, loadCache bool) *Config {
	fileLock := flock.New(path.Join(getDefaultConfigDir(), "lock"))
	err := fileLock.Lock()
	if err != nil {
		fmt.Println("Failed to grab config file lock, please try again")
		return cfg
	}
	cfg = saveConfig(cfg)
	fileLock.Unlock()
	if loadCache {
		LoadCache(cfg)
	}
	return cfg
}

func readConfig(cfg *Config) *ini.File {
	conf, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, cfg.ConfigFile)

	if err != nil {
		fmt.Printf("Fail to read config file: %v", err)
		os.Exit(1)
	}
	return conf
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

	conf := readConfig(cfg)

	core, err := conf.GetSection(ini.DEFAULT_SECTION)
	if core == nil || err != nil {
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
		if !conf.Section(ini.DEFAULT_SECTION).HasKey("autocomplete") {
			core.AutoComplete = true
			core.Output = JSON
		}
		if !conf.Section(ini.DEFAULT_SECTION).HasKey("postrequest") {
			core.PostRequest = true
		}
		cfg.Core = core
	}

	profile, err := conf.GetSection(cfg.Core.ProfileName)
	if profile == nil {
		activeProfile := defaultProfile()
		section, _ := conf.NewSection(cfg.Core.ProfileName)
		section.ReflectFrom(&activeProfile)
		setActiveProfile(cfg, &activeProfile)
	} else {
		// Write
		if cfg.ActiveProfile != nil {
			conf.Section(cfg.Core.ProfileName).ReflectFrom(&cfg.ActiveProfile)
		}
		// Update
		profile := new(ServerProfile)
		conf.Section(cfg.Core.ProfileName).MapTo(profile)
		setActiveProfile(cfg, profile)
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

// LoadProfile loads an existing profile
func (c *Config) LoadProfile(name string) {
	Debug("Trying to load profile: " + name)
	conf := readConfig(c)
	section, err := conf.GetSection(name)
	if err != nil || section == nil {
		fmt.Printf("Unable to load profile '%s': %v", name, err)
		os.Exit(1)
	}
	profile := new(ServerProfile)
	conf.Section(name).MapTo(profile)
	setActiveProfile(c, profile)
	c.Core.ProfileName = name
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
		intValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("Error caught while setting timeout:", err)
			return
		}
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
		if value == "true" || value == "on" {
			EnableDebugging()
		} else {
			DisableDebugging()
		}
	case "autocomplete":
		c.Core.AutoComplete = value == "true"
	default:
		fmt.Println("Invalid option provided:", key)
		return
	}

	Debug("UpdateConfig key:", key, " value:", value, " update:", update)

	if update {
		reloadConfig(c, true)
	}
}

// NewConfig creates or reload config and loads API cache
func NewConfig(configFilePath *string) *Config {
	defaultConf := defaultConfig()
	defaultConf.Core = nil
	defaultConf.ActiveProfile = nil
	if *configFilePath != "" {
		defaultConf.ConfigFile, _ = filepath.Abs(*configFilePath)
		if _, err := os.Stat(defaultConf.ConfigFile); os.IsNotExist(err) {
			fmt.Println("Config file doesn't exist.")
			os.Exit(1)
		}
	}
	cfg := reloadConfig(defaultConf, false)
	return cfg
}
