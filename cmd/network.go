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
	"net/url"
	"sort"
	"strings"
)

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

func NewAPIRequest(r *Request, api string, args []string) (map[string]interface{}, error) {
	params := make(url.Values)
	params.Add("command", api)
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		if len(parts) == 2 {
			params.Add(parts[0], parts[1])
		}
	}

	apiKey := r.Config.Core.ActiveProfile.ApiKey
	secretKey := r.Config.Core.ActiveProfile.SecretKey

	if len(apiKey) > 0 {
		params.Add("apiKey", apiKey)
	}

	params.Add("response", "json")
	encodedParams := encodeRequestParams(params)

	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(strings.Replace(strings.ToLower(encodedParams), "+", "%20", -1)))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	encodedParams = encodedParams + fmt.Sprintf("&signature=%s", url.QueryEscape(signature))

	apiUrl := fmt.Sprintf("%s?%s", r.Config.Core.ActiveProfile.Url, encodedParams)

	//fmt.Println("[debug] Requesting: ", apiUrl)
	response, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(response.Body)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(body), &data)

	for k := range data {
		if strings.HasSuffix(k, "response") {
			return data[k].(map[string]interface{}), nil
		}
	}
	return nil, errors.New("failed to decode response")
}
