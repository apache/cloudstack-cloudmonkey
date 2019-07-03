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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/apache/cloudstack-cloudmonkey/config"
)

func findSessionCookie(cookies []*http.Cookie) *http.Cookie {
	if cookies == nil {
		return nil
	}
	for _, cookie := range cookies {
		if cookie.Name == "sessionkey" {
			return cookie
		}
	}
	return nil
}

// Login logs in a user based on provided request and returns http client and session key
func Login(r *Request) (string, error) {
	params := make(url.Values)
	params.Add("command", "login")
	params.Add("username", r.Config.ActiveProfile.Username)
	params.Add("password", r.Config.ActiveProfile.Password)
	params.Add("domain", r.Config.ActiveProfile.Domain)
	params.Add("response", "json")

	msURL, _ := url.Parse(r.Config.ActiveProfile.URL)
	if sessionCookie := findSessionCookie(r.Client().Jar.Cookies(msURL)); sessionCookie != nil {
		return sessionCookie.Value, nil
	}

	config.Debug("Login POST URL:", msURL, params)
	spinner := r.Config.StartSpinner("trying to log in...")
	resp, err := r.Client().PostForm(msURL.String(), params)
	r.Config.StopSpinner(spinner)

	if err != nil {
		return "", errors.New("failed to authenticate with the CloudStack server, please check the settings: " + err.Error())
	}

	config.Debug("Login POST response status code:", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		e := errors.New("failed to authenticate, please check the credentials")
		if err != nil {
			e = errors.New("failed to authenticate due to " + err.Error())
		}
		return "", e
	}

	var sessionKey string
	curTime := time.Now()
	expiryDuration := 15 * time.Minute
	for _, cookie := range resp.Cookies() {
		if cookie.Expires.After(curTime) {
			expiryDuration = cookie.Expires.Sub(curTime)
		}
		if cookie.Name == "sessionkey" {
			sessionKey = cookie.Value
		}
	}
	go func() {
		time.Sleep(expiryDuration)
		r.Client().Jar, _ = cookiejar.New(nil)
	}()

	config.Debug("Login sessionkey:", sessionKey)
	return sessionKey, nil
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
		spinner := r.Config.StartSpinner("polling for async API result")
		queryResult, queryError := NewAPIRequest(r, "queryAsyncJobResult", []string{"jobid=" + jobID}, false)
		diff := time.Duration(1*time.Second).Nanoseconds() - time.Now().Sub(startTime).Nanoseconds()
		if diff > 0 {
			time.Sleep(time.Duration(diff) * time.Nanosecond)
		}
		timeout = timeout - time.Now().Sub(startTime).Seconds()
		r.Config.StopSpinner(spinner)
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
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}
			if strings.HasPrefix(value, "@") {
				possibleFileName := value[1:]
				if fileInfo, err := os.Stat(possibleFileName); err == nil && !fileInfo.IsDir() {
					bytes, err := ioutil.ReadFile(possibleFileName)
					config.Debug()
					if err == nil {
						value = string(bytes)
						config.Debug("Content for argument ", key, " read from file: ", possibleFileName, " is: ", value)
					}
				}
			}
			params.Add(key, value)
		}
	}
	params.Add("response", "json")

	var encodedParams string
	var err error

	if len(r.Config.ActiveProfile.APIKey) > 0 && len(r.Config.ActiveProfile.SecretKey) > 0 {
		apiKey := r.Config.ActiveProfile.APIKey
		secretKey := r.Config.ActiveProfile.SecretKey

		if len(apiKey) > 0 {
			params.Add("apiKey", apiKey)
		}
		encodedParams = encodeRequestParams(params)

		mac := hmac.New(sha1.New, []byte(secretKey))
		mac.Write([]byte(strings.Replace(strings.ToLower(encodedParams), "+", "%20", -1)))
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		encodedParams = encodedParams + fmt.Sprintf("&signature=%s", url.QueryEscape(signature))
	} else if len(r.Config.ActiveProfile.Username) > 0 && len(r.Config.ActiveProfile.Password) > 0 {
		sessionKey, err := Login(r)
		if err != nil {
			return nil, err
		}
		params.Add("sessionkey", sessionKey)
		encodedParams = encodeRequestParams(params)
	} else {
		fmt.Println("Please provide either apikey/secretkey or username/password to make an API call")
		return nil, errors.New("failed to authenticate to make API call")
	}

	requestURL := fmt.Sprintf("%s?%s", r.Config.ActiveProfile.URL, encodedParams)
	config.Debug("NewAPIRequest API request URL:", requestURL)

	response, err := r.Client().Get(requestURL)
	if err != nil {
		return nil, err
	}
	config.Debug("NewAPIRequest response status code:", response.StatusCode)

	if response != nil && response.StatusCode == http.StatusUnauthorized {
		r.Client().Jar, _ = cookiejar.New(nil)
		sessionKey, err := Login(r)
		if err != nil {
			return nil, err
		}
		params.Del("sessionkey")
		params.Add("sessionkey", sessionKey)
		requestURL = fmt.Sprintf("%s?%s", r.Config.ActiveProfile.URL, encodeRequestParams(params))
		config.Debug("NewAPIRequest API request URL:", requestURL)
		response, err = r.Client().Get(requestURL)
		if err != nil {
			return nil, err
		}
	}

	body, _ := ioutil.ReadAll(response.Body)
	config.Debug("NewAPIRequest response body:", string(body))

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
