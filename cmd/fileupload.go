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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/apache/cloudstack-cloudmonkey/config"
	"github.com/briandowns/spinner"
)

const (
	uploadingMessage  = "Uploading files, please wait..."
	progressCharCount = 24
)

// ValidateAndGetFileList parses a comma-separated string of file paths, trims them,
// checks for existence, and returns a slice of valid file paths or an error if any are missing.
func ValidateAndGetFileList(filePaths string) ([]string, error) {
	filePathsList := strings.FieldsFunc(filePaths, func(r rune) bool { return r == ',' })

	var missingFiles []string
	var validFiles []string
	for _, filePath := range filePathsList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, filePath)
		} else {
			validFiles = append(validFiles, filePath)
		}
	}
	if len(missingFiles) > 0 {
		return nil, fmt.Errorf("file(s) do not exist or are not accessible: %s", strings.Join(missingFiles, ", "))
	}
	return validFiles, nil
}

// PromptAndUploadFilesIfNeeded prompts the user to provide file paths for upload and the API is getUploadParamsFor*
func PromptAndUploadFilesIfNeeded(r *Request, api string, response map[string]interface{}) {
	if !r.Config.HasShell {
		return
	}
	apiName := strings.ToLower(api)
	if !config.IsFileUploadAPI(apiName) {
		return
	}
	fmt.Print("Enter path of the file(s) to upload (comma-separated), leave empty to skip: ")
	var filePaths string
	fmt.Scanln(&filePaths)
	if filePaths == "" {
		return
	}
	validFiles, err := ValidateAndGetFileList(filePaths)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(validFiles) == 0 {
		fmt.Println("No valid files to upload.")
		return
	}
	UploadFiles(r, api, response, validFiles)
}

// UploadFiles uploads files to a remote server using parameters from the API response.
// Shows progress for each file and reports any failures.
func UploadFiles(r *Request, api string, response map[string]interface{}, validFiles []string) {
	paramsRaw, ok := response["getuploadparams"]
	if !ok || reflect.TypeOf(paramsRaw).Kind() != reflect.Map {
		fmt.Println("Invalid response format for getuploadparams.")
		return
	}
	params := paramsRaw.(map[string]interface{})
	requiredKeys := []string{"postURL", "metadata", "signature", "expires"}
	for _, key := range requiredKeys {
		if _, ok := params[key]; !ok {
			fmt.Printf("Missing required key '%s' in getuploadparams response.\n", key)
			return
		}
	}
	postURL, _ := params["postURL"].(string)
	signature, _ := params["signature"].(string)
	expires, _ := params["expires"].(string)
	metadata, _ := params["metadata"].(string)

	fmt.Println("Uploading files for", api, ":", validFiles)

	var spinner *spinner.Spinner
	if r.Config.HasShell {
		spinner = r.Config.StartSpinner(uploadingMessage)
	}

	errored := 0
	for i, filePath := range validFiles {
		if spinner != nil {
			spinner.Suffix = fmt.Sprintf(" uploading %d/%d %s...", i+1, len(validFiles), filepath.Base(filePath))
		}

		if err := uploadFile(i, len(validFiles), postURL, filePath, signature, expires, metadata, spinner); err != nil {
			if spinner != nil {
				spinner.Stop()
			}

			fmt.Println("Error uploading", filePath, ":", err)
			errored++

			if spinner != nil {
				spinner.Suffix = fmt.Sprintf(" %s", uploadingMessage)
				spinner.Start()
			}
		}
	}

	r.Config.StopSpinner(spinner)
	if errored > 0 {
		fmt.Printf("🙈 %d out of %d files failed to upload.\n", errored, len(validFiles))
	} else {
		fmt.Println("All files uploaded successfully.")
	}
}

// progressReader streams file data and updates progress as bytes are read.
type progressBody struct {
	f      *os.File
	read   int64
	total  int64
	update func(int)
}

func (pb *progressBody) Read(p []byte) (int, error) {
	n, err := pb.f.Read(p)
	if n > 0 {
		pb.read += int64(n)
		pct := int(float64(pb.read) * 100 / float64(pb.total))
		pb.update(pct)
	}
	return n, err
}
func (pb *progressBody) Close() error { return pb.f.Close() }

func barArrow(pct int) string {
	width := progressCharCount
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	pos := (pct * width) / 100
	// 100%: full bar, no head
	if pos >= width {
		return fmt.Sprintf("[%s]",
			strings.Repeat("=", width))
	}
	left := strings.Repeat("=", pos) + ">"
	right := strings.Repeat(" ", width-pos-1)

	return fmt.Sprintf("[%s%s]", left, right)
}

// uploadFile streams a large file to the server with progress updates.
func uploadFile(index, count int, postURL, filePath, signature, expires, metadata string, spn *spinner.Spinner) error {
	fileName := filepath.Base(filePath)
	in, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = in.Stat()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "multipart-body-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	mw := multipart.NewWriter(tmp)
	part, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, in); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}
	size, err := tmp.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", postURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("x-signature", signature)
	req.Header.Set("x-expires", expires)
	req.Header.Set("x-metadata", metadata)
	req.ContentLength = size
	pb := &progressBody{
		f:     tmp,
		total: size,
		update: func(pct int) {
			if spn != nil {
				spn.Suffix = fmt.Sprintf(" [%d/%d] %s\t%s %d%%", index+1, count, fileName, barArrow(pct), pct)
			}
		},
	}
	req.Body = pb
	req.GetBody = func() (io.ReadCloser, error) {
		f, err := os.Open(tmp.Name())
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	client := &http.Client{
		Timeout: 24 * time.Hour,
		Transport: &http.Transport{
			ExpectContinueTimeout: 0,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("[%d/%d] %s\tupload failed: %s", index+1, count, fileName, string(b))
	}

	if spn != nil {
		spn.Stop()
	}
	fmt.Printf("[%d/%d] %s\t%s ✅\n", index+1, count, fileName, barArrow(100))
	if spn != nil {
		spn.Suffix = fmt.Sprintf(" %s", uploadingMessage)
		spn.Start()
	}
	return nil
}
