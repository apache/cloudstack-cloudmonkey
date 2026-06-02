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
	"crypto/sha512"
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

func getLoginResponse(responseBody []byte) (map[string]interface{}, error) {
	var responseMap map[string]interface{}
	err := json.Unmarshal(responseBody, &responseMap)
	if err != nil {
		return nil, errors.New("failed to parse login response: " + err.Error())
	}
	loginRespRaw, ok := responseMap["loginresponse"]
	if !ok {
		return nil, errors.New("failed to parse login response, expected 'loginresponse' key not found")
	}
	loginResponse, ok := loginRespRaw.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to parse login response, expected 'loginresponse' to be a map")
	}
	return loginResponse, nil
}

func getResponseBooleanValue(response map[string]interface{}, key string) (bool, bool) {
	v, found := response[key]
	if !found {
		return false, false
	}
	switch value := v.(type) {
	case bool:
		return true, value
	case string:
		return true, strings.ToLower(value) == "true"
	case float64:
		return true, value != 0
	default:
		return true, false
	}
}

func checkLogin2FAPromptAndValidate(r *Request, response map[string]interface{}, sessionKey string) error {
	if !r.Config.HasShell {
		return nil
	}
	config.Debug("Checking if 2FA is enabled and verified for the user ", response)
	found, is2faEnabled := getResponseBooleanValue(response, "is2faenabled")
	if !found || !is2faEnabled {
		config.Debug("2FA is not enabled for the user, skipping 2FA validation")
		return nil
	}
	found, is2faVerified := getResponseBooleanValue(response, "is2faverified")
	if !found || is2faVerified {
		config.Debug("2FA is already verified for the user, skipping 2FA validation")
		return nil
	}
	activeSpinners := r.Config.PauseActiveSpinners()
	fmt.Print("Enter 2FA code: ")
	var code string
	fmt.Scanln(&code)
	if activeSpinners > 0 {
		r.Config.ResumePausedSpinners()
	}
	params := make(url.Values)
	params.Add("command", "validateUserTwoFactorAuthenticationCode")
	params.Add("codefor2fa", code)
	params.Add("sessionkey", sessionKey)

	msURL, _ := url.Parse(r.Config.ActiveProfile.URL)

	config.Debug("Validating 2FA with POST URL:", msURL, params)
	spinner := r.Config.StartSpinner("trying to validate 2FA...")
	resp, err := r.Client().PostForm(msURL.String(), params)
	r.Config.StopSpinner(spinner)
	if err != nil {
		return errors.New("failed to failed to validate 2FA code: " + err.Error())
	}
	config.Debug("ValidateUserTwoFactorAuthenticationCode POST response status code:", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		r.Client().Jar, _ = cookiejar.New(nil)
		return errors.New("failed to validate 2FA code, please check the code. Invalidating session")
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

	body, _ := ioutil.ReadAll(resp.Body)
	config.Debug("Login response body:", string(body))
	loginResponse, err := getLoginResponse(body)
	if err != nil {
		return "", err
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
	if err := checkLogin2FAPromptAndValidate(r, loginResponse, sessionKey); err != nil {
		return "", err
	}
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
		escaped := url.QueryEscape(value)
		// we need to ensure + (representing a space) is encoded as %20
		escaped = strings.Replace(escaped, "+", "%20", -1)
		// we need to ensure * is not escaped
		escaped = strings.Replace(escaped, "%2A", "*", -1)
		buf.WriteString(escaped)
	}
	return buf.String()
}

func cloneRequestParams(params url.Values) url.Values {
	cloned := make(url.Values)
	for key, values := range params {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

func buildAPIRequestParams(r *Request, api string, args []string) url.Values {
	params := make(url.Values)
	params.Add("command", api)
	apiData := r.Config.GetCache()[api]
	for _, arg := range args {
		if apiData != nil {
			skip := false
			for _, fakeArg := range apiData.FakeArgs {
				if strings.HasPrefix(arg, fakeArg) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

		}
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
	signatureversion := "3"
	expiresKey := "expires"
	params.Add("response", "json")
	params.Add("signatureversion", signatureversion)
	params.Add(expiresKey, time.Now().UTC().Add(15*time.Minute).Format(time.RFC3339))
	return params
}

func signRequest(unsignedRequest, secretKey, algorithm string) (string, error) {
	signatureAlgorithm, err := config.NormalizeSignatureAlgorithm(algorithm)
	if err != nil {
		return "", err
	}

	var signature []byte
	switch signatureAlgorithm {
	case config.SignatureAlgorithmHmacSHA1:
		mac := hmac.New(sha1.New, []byte(secretKey))
		mac.Write([]byte(strings.ToLower(unsignedRequest)))
		signature = mac.Sum(nil)
	case config.SignatureAlgorithmHmacSHA512:
		mac := hmac.New(sha512.New, []byte(secretKey))
		mac.Write([]byte(strings.ToLower(unsignedRequest)))
		signature = mac.Sum(nil)
	default:
		return "", errors.New("signature algorithm must be concrete")
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func executeSignedAPIRequest(r *Request, unsignedParams url.Values, algorithm string) (*http.Response, error) {
	params := cloneRequestParams(unsignedParams)
	encodedParams := encodeRequestParams(params)

	signature, err := signRequest(encodedParams, r.Config.ActiveProfile.SecretKey, algorithm)
	if err != nil {
		return nil, err
	}
	if r.Config.Core.PostRequest {
		params.Add("signature", signature)
	} else {
		encodedParams = encodedParams + fmt.Sprintf("&signature=%s", url.QueryEscape(signature))
		params = nil
	}

	requestURL := fmt.Sprintf("%s?%s", r.Config.ActiveProfile.URL, encodedParams)
	config.Debug("NewAPIRequest API request URL:", requestURL)
	return executeRequest(r, requestURL, params)
}

func parseAPIResponse(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errors.New("failed to decode response")
	}

	if apiResponse := getResponseData(data); apiResponse != nil {
		if _, ok := apiResponse["errorcode"]; ok {
			return nil, fmt.Errorf("(HTTP %v, error code %v) %v", apiResponse["errorcode"], apiResponse["cserrorcode"], apiResponse["errortext"])
		}
		return apiResponse, nil
	}

	return nil, errors.New("failed to decode response")
}

func isAuthenticationFailure(statusCode int, err error) bool {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return true
	}
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	for _, marker := range []string{"signature", "authenticate", "authentication", "credential", "unauthoriz", "api key", "apikey"} {
		if strings.Contains(errText, marker) {
			return true
		}
	}
	return false
}

func persistDetectedSignatureAlgorithm(r *Request, algorithm string) {
	r.Config.ActiveProfile.SignatureAlgorithm = algorithm
	if r.CredentialsSupplied {
		config.Debug("Credentials supplied on command-line, not persisting detected signature algorithm")
		return
	}
	r.Config.UpdateConfig("signaturealgorithm", algorithm, true)
}

func detectSignatureAlgorithm(r *Request) (string, error) {
	attempts := []string{config.SignatureAlgorithmHmacSHA512, config.SignatureAlgorithmHmacSHA1}
	var lastErr error

	for _, algorithm := range attempts {
		config.Debug("Trying API signature algorithm probe:", algorithm)
		params := buildAPIRequestParams(r, "listApis", []string{"listall=true"})
		params.Add("apiKey", r.Config.ActiveProfile.APIKey)
		response, err := executeSignedAPIRequest(r, params, algorithm)
		if err != nil {
			config.Debug("API signature algorithm probe failed before response for ", algorithm, ": ", err)
			lastErr = err
			continue
		}

		body, _ := ioutil.ReadAll(response.Body)
		config.Debug("Signature algorithm probe response body:", string(body))
		if _, err := parseAPIResponse(body); err == nil {
			config.Debug("Selected API signature algorithm:", algorithm)
			persistDetectedSignatureAlgorithm(r, algorithm)
			return algorithm, nil
		} else {
			lastErr = err
			if isAuthenticationFailure(response.StatusCode, err) {
				config.Debug("API signature algorithm probe failed authentication for ", algorithm, ": ", err)
			} else {
				config.Debug("API signature algorithm probe failed with non-authentication error for ", algorithm, ": ", err)
			}
		}
	}

	config.Debug("Signature algorithm autodetection failed; attempted algorithms:", strings.Join(attempts, ", "))
	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("failed to detect signature algorithm")
}

func activeSignatureAlgorithm(r *Request) (string, error) {
	signatureAlgorithm, err := config.NormalizeSignatureAlgorithm(r.Config.ActiveProfile.SignatureAlgorithm)
	if err != nil {
		return "", err
	}
	if signatureAlgorithm == config.SignatureAlgorithmAuto {
		return detectSignatureAlgorithm(r)
	}
	return signatureAlgorithm, nil
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
	timeout := time.NewTimer(time.Duration(float64(r.Config.Core.Timeout)) * time.Second)
	ticker := time.NewTicker(time.Duration(2 * time.Second))
	defer ticker.Stop()
	defer timeout.Stop()

	spinner := r.Config.StartSpinner("polling for async API result")
	defer r.Config.StopSpinner(spinner)

	for {
		select {
		case <-r.Config.C:
			return nil, errors.New("async API job polling interrupted")

		case <-timeout.C:
			return nil, errors.New("async API job query timed out")

		case <-ticker.C:
			args := []string{"jobid=" + jobID}
			if r.Args != nil {
				for _, arg := range r.Args {
					if strings.HasPrefix(strings.ToLower(arg), "filter=") {
						args = append(args, arg)
						break
					}
				}
			}
			queryResult, queryError := NewAPIRequest(r, "queryAsyncJobResult", args, false)
			if queryError != nil {
				return queryResult, queryError
			}
			jobStatus := queryResult["jobstatus"].(float64)

			switch jobStatus {
			case 0:
				continue

			case 1:
				return queryResult["jobresult"].(map[string]interface{}), nil

			case 2:
				return queryResult, errors.New("async API failed for job " + jobID)
			}
		}
	}
}

// NewAPIRequest makes an API request to configured management server
func NewAPIRequest(r *Request, api string, args []string, isAsync bool) (map[string]interface{}, error) {
	params := buildAPIRequestParams(r, api, args)

	var encodedParams string
	var err error
	usingSessionAuth := false

	if len(r.Config.ActiveProfile.APIKey) > 0 && len(r.Config.ActiveProfile.SecretKey) > 0 {
		apiKey := r.Config.ActiveProfile.APIKey
		if len(apiKey) > 0 {
			params.Add("apiKey", apiKey)
		}
		signatureAlgorithm, err := activeSignatureAlgorithm(r)
		if err != nil {
			return nil, err
		}
		response, err := executeSignedAPIRequest(r, params, signatureAlgorithm)
		if err != nil {
			return nil, err
		}
		return processAPIResponse(r, response, isAsync)
	} else if len(r.Config.ActiveProfile.Username) > 0 && len(r.Config.ActiveProfile.Password) > 0 {
		usingSessionAuth = true
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

	var response *http.Response
	response, err = executeRequest(r, requestURL, params)
	if err != nil {
		return nil, err
	}
	config.Debug("NewAPIRequest response status code:", response.StatusCode)

	if r.CredentialsSupplied {
		config.Debug("Credentials supplied on command-line, not falling back to login")
	}

	if usingSessionAuth && response.StatusCode == http.StatusUnauthorized && !r.CredentialsSupplied {
		r.Client().Jar, _ = cookiejar.New(nil)
		sessionKey, err := Login(r)
		if err != nil {
			return nil, err
		}
		params.Del("sessionkey")
		params.Add("sessionkey", sessionKey)
		requestURL = fmt.Sprintf("%s?%s", r.Config.ActiveProfile.URL, encodeRequestParams(params))
		config.Debug("NewAPIRequest API request URL:", requestURL)

		response, err = executeRequest(r, requestURL, params)
		if err != nil {
			return nil, err
		}
	}

	return processAPIResponse(r, response, isAsync)
}

func processAPIResponse(r *Request, response *http.Response, isAsync bool) (map[string]interface{}, error) {
	body, _ := ioutil.ReadAll(response.Body)
	config.Debug("NewAPIRequest response body:", string(body))

	apiResponse, err := parseAPIResponse(body)
	if err != nil {
		return nil, err
	}

	if isAsync && r.Config.Core.AsyncBlock {
		if apiResponse["jobid"] != nil {
			jobID := apiResponse["jobid"].(string)
			return pollAsyncJob(r, jobID)
		}
	}

	return apiResponse, nil
}

// we can implement further conditions to do POST or GET (or other http commands) here
func executeRequest(r *Request, requestURL string, params url.Values) (*http.Response, error) {
	config.SetupContext(r.Config)
	if params.Has("password") || params.Has("userdata") || r.Config.Core.PostRequest {
		requestURL = r.Config.ActiveProfile.URL
		config.Debug("Using HTTP POST for the request: ", requestURL)
		return r.Client().PostForm(requestURL, params)
	}
	config.Debug("Using HTTP GET for the request: ", requestURL)
	req, _ := http.NewRequestWithContext(*r.Config.Context, "GET", requestURL, nil)
	return r.Client().Do(req)
}
