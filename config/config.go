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
	"os"
	"path"
)

type OutputFormat string

const (
	Json  OutputFormat = "json"
	Xml   OutputFormat = "xml"
	Table OutputFormat = "table"
	Text  OutputFormat = "text"
)

type Profile struct {
	Name       string
	Url        string
	VerifyCert bool
	Username   string
	Password   string
	Domain     string
	ApiKey     string
	SecretKey  string
}

type Config struct {
	Dir           string
	ConfigFile    string
	HistoryFile   string
	CacheFile     string
	LogFile       string
	Output        OutputFormat
	AsyncBlock    bool
	ActiveProfile Profile
}

func NewConfig() *Config {
	return loadConfig()
}

func defaultConfig() *Config {
	configDir := getDefaultConfigDir()
	return &Config{
		Dir:         configDir,
		ConfigFile:  path.Join(configDir, "config"),
		HistoryFile: path.Join(configDir, "history"),
		CacheFile:   path.Join(configDir, "cache"),
		LogFile:     path.Join(configDir, "log"),
		Output:      Json,
		AsyncBlock:  true,
		ActiveProfile: Profile{
			Name:       "local",
			Url:        "http://192.168.1.10:8080/client/api",
			VerifyCert: false,
			Username:   "admin",
			Password:   "password",
			Domain:     "/",
			// TODO: remove test data
			ApiKey:    "IgrUOA_46IVoBNzAR_Th2JbdbgIs2lMW1kGe9A80F9X0uOnfGO0Su23IqOSqbdzZW3To95PNrcdWsk60ieXYBQ",
			SecretKey: "E7NRSv5d_1VhqXUHJEqvAsm7htR_V_vtPJZsCPkgPKSgkiS3sh4SOrIqMm_eWhSFoL6RHRIlxtA_viQAt7EDVA",
		},
	}
}

func loadConfig() *Config {
	cfg := defaultConfig()

	if _, err := os.Stat(cfg.Dir); err != nil {
		os.Mkdir(cfg.Dir, 0700)
	}

	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		// FIXME: write default cfg
	} else {
		//load config?
	}

	LoadCache(cfg)

	return cfg
}
