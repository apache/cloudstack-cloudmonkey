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
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"
)

var cursor = []string{"\r⣷ 😸", "\r⣯ 😹", "\r⣟ 😺", "\r⡿ 😻", "\r⢿ 😼", "\r⣻ 😽", "\r⣽ 😾", "\r⣾ 😻"}

func init() {
	if runtime.GOOS == "windows" {
		cursor = []string{"|", "/", "-", "\\"}
	}
}

func encodeRequestParams(params url.Values) string {
	if params == nil {
		return ""
	}

	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, key := range keys {
		value := params.Get(key)
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(value))
	}
	return buf.String()
}

// Login logs in a user based on provided request and returns http client and session key
func Login(r *Request) (*http.Client, string, error) {
	params := make(url.Values)
	params.Add("command", "login")
	params.Add("username", r.Config.ActiveProfile.Username)
	params.Add("password", r.Config.ActiveProfile.Password)
	params.Add("domain", r.Config.ActiveProfile.Domain)
	params.Add("response", "json")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !r.Config.Core.VerifyCert},
		},
	}

	sessionKey := ""
	resp, err := client.PostForm(r.Config.ActiveProfile.URL, params)
	if err != nil {
		return client, sessionKey, errors.New("failed to authenticate with the CloudStack server, please check the settings: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		e := errors.New("failed to authenticate, please check the credentials")
		if err != nil {
			e = errors.New("failed to authenticate due to " + err.Error())
		}
		return client, sessionKey, e
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessionkey" {
			sessionKey = cookie.Value
			break
		}
	}
	return client, sessionKey, nil
}

func getResponseData(data map[string]interface{}) map[string]interface{} {
	for k := range data {
		if strings.HasSuffix(k, "response") {
			return data[k].(map[string]interface{})
		}
	}
	return nil
}

func pollAsyncJob(r *Request, jobID string) (map[string]interface{}, error) {
	for timeout := float64(r.Config.Core.Timeout); timeout > 0.0; {
		startTime := time.Now()
		s := spinner.New(cursor, 200*time.Millisecond)
		s.Color("blue", "bold")
		s.Suffix = " polling for async API job result"
		s.Start()
		queryResult, queryError := NewAPIRequest(r, "queryAsyncJobResult", []string{"jobid=" + jobID}, false)
		diff := time.Duration(1*time.Second).Nanoseconds() - time.Now().Sub(startTime).Nanoseconds()
		if diff > 0 {
			time.Sleep(time.Duration(diff) * time.Nanosecond)
		}
		s.Stop()
		timeout = timeout - time.Now().Sub(startTime).Seconds()
		if queryError != nil {
			return queryResult, queryError
		}
		jobStatus := queryResult["jobstatus"].(float64)
		if jobStatus == 0 {
			continue
		}
		if jobStatus == 1 {
			return queryResult["jobresult"].(map[string]interface{}), nil

		}
		if jobStatus == 2 {
			return queryResult, errors.New("async API failed for job " + jobID)
		}
	}
	return nil, errors.New("async API job query timed out")
}

// NewAPIRequest makes an API request to configured management server
func NewAPIRequest(r *Request, api string, args []string, isAsync bool) (map[string]interface{}, error) {
	params := make(url.Values)
	params.Add("command", api)
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		if len(parts) == 2 {
			params.Add(parts[0], parts[1])
		}
	}
	params.Add("response", "json")

	var client *http.Client
	var encodedParams string
	var err error

	if len(r.Config.ActiveProfile.APIKey) > 0 && len(r.Config.ActiveProfile.SecretKey) > 0 {
		apiKey := r.Config.ActiveProfile.APIKey
		secretKey := r.Config.ActiveProfile.SecretKey

		if len(apiKey) > 0 {
			params.Add("apiKey", apiKey)
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !r.Config.Core.VerifyCert},
			},
		}
		encodedParams = encodeRequestParams(params)

		mac := hmac.New(sha1.New, []byte(secretKey))
		mac.Write([]byte(strings.Replace(strings.ToLower(encodedParams), "+", "%20", -1)))
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		encodedParams = encodedParams + fmt.Sprintf("&signature=%s", url.QueryEscape(signature))
	} else if len(r.Config.ActiveProfile.Username) > 0 && len(r.Config.ActiveProfile.Password) > 0 {
		var sessionKey string
		client, sessionKey, err = Login(r)
		if err != nil {
			return nil, err
		}
		params.Add("sessionkey", sessionKey)
		encodedParams = encodeRequestParams(params)
	} else {
		fmt.Println("Please provide either apikey/secretkey or username/password to make an API call")
		return nil, errors.New("failed to authenticate to make API call")
	}

	client.Timeout = time.Duration(time.Duration(r.Config.Core.Timeout) * time.Second)
	response, err := client.Get(fmt.Sprintf("%s?%s", r.Config.ActiveProfile.URL, encodedParams))
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(response.Body)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(body), &data)

	if isAsync && r.Config.Core.AsyncBlock {
		if jobResponse := getResponseData(data); jobResponse != nil && jobResponse["jobid"] != nil {
			jobID := jobResponse["jobid"].(string)
			return pollAsyncJob(r, jobID)
		}
	}

	if apiResponse := getResponseData(data); apiResponse != nil {
		if _, ok := apiResponse["errorcode"]; ok {
			return nil, fmt.Errorf("(HTTP %v, error code %v) %v", apiResponse["errorcode"], apiResponse["cserrorcode"], apiResponse["errortext"])
		}
		return apiResponse, nil
	}

	return nil, errors.New("failed to decode response")
}
