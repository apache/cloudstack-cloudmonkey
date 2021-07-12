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

package app

import (
	"net/http"

	"github.com/apache/cloudstack-cloudmonkey/internal/config"
)

// Request describes a command request
type Request struct {
	Command *Command
	Config  *config.Config
	Args    []string
}

// Client method returns the http Client for the current server profile
func (r *Request) Client() *http.Client {
	return r.Config.ActiveProfile.Client
}

// NewRequest creates a new request from a command
func NewRequest(cmd *Command, cfg *config.Config, args []string) *Request {
	return &Request{
		Command: cmd,
		Config:  cfg,
		Args:    args,
	}
}
