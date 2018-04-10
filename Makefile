# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

build:
	go build -ldflags='-s -w' -o cmk cmk.go

run:
	go run cmk.go

test:
	go test

install: build
	@echo Copied to ~/bin
	@cp cmk ~/bin

debug:
	go build -gcflags='-N -l' -o cmk cmk.go &&  dlv --listen=:2345 --headless=true --api-version=2 exec ./cmk

dist:
	rm -fr dist
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -ldflags='-s -w' -o dist/cmk-linux-amd64 cmk.go
	GOOS=linux   GOARCH=386   go build -ldflags='-s -w' -o dist/cmk-linux-i386 cmk.go
	GOOS=linux   GOARCH=arm64 go build -ldflags='-s -w' -o dist/cmk-linux-arm64 cmk.go
	GOOS=linux   GOARCH=arm   go build -ldflags='-s -w' -o dist/cmk-linux-arm cmk.go
	GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o dist/cmk-x64.exe cmk.go
	GOOS=windows GOARCH=386   go build -ldflags='-s -w' -o dist/cmk-x32.exe cmk.go
	GOOS=darwin  GOARCH=amd64 go build -ldflags='-s -w' -o dist/cmk-mac64.bin cmk.go
	GOOS=darwin  GOARCH=386   go build -ldflags='-s -w' -o dist/cmk-mac32.bin cmk.go

clean:
	@rm -f cmk
	@rm -fr dist

deps:
	go get -u github.com/rhtyd/readline
	go get -u github.com/mitchellh/go-homedir
	go get -u github.com/mattn/go-shellwords
	go get -u github.com/manifoldco/promptui

