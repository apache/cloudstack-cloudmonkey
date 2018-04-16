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
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
)

func init() {
	AddCommand(&Command{
		Name: "login",
		Help: "Log in to your account",
		Handle: func(r *Request) error {
			if len(r.Args) > 0 {
				return errors.New("this does not accept any additional arguments")
			}

			validate := func(input string) error {
				if len(input) < 1 {
					return errors.New("You have not entered anything")
				}
				return nil
			}

			// username
			prompt := promptui.Prompt{
				Label:    "Username",
				Validate: validate,
				Default:  "",
			}
			username, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return nil
			}

			//password
			prompt = promptui.Prompt{
				Label:    "Password",
				Validate: validate,
				Mask:     '*',
			}
			password, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return nil
			}

			// domain
			prompt = promptui.Prompt{
				Label:    "Domain",
				Validate: validate,
				Default:  "/",
			}
			domain, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return nil
			}

			r.Config.ActiveProfile.Username = username
			r.Config.ActiveProfile.Password = password
			r.Config.ActiveProfile.Domain = domain

			client, _, err := Login(r)
			if client == nil || err != nil {
				fmt.Println("Failed to login, check credentials")
			} else {
				fmt.Println("Success!")
			}

			return nil
		},
	})
}
