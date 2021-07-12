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

# Makefile referenced from github.com/vincentbernat/hellogopher
PACKAGE  = cmk
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN      = $(CURDIR)/bin
BASE     = $(CURDIR)
PKGS     = $(or $(PKG),$(shell $(GO) list ./... | grep -v "^$(PACKAGE)/vendor/"))
TESTPKGS = $(shell $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
GIT_SHA  = $(shell git rev-parse --short HEAD)

GO      = CGO_ENABLED=0 go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m ")

.PHONY: all
all: fmt ; $(info $(M) Building executable… $(GIT_SHA)) @ ## Build program binary
	$Q $(GO) build -mod=vendor \
		-tags release \
		-ldflags '-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' \
		-o bin/$(PACKAGE) cmd/main.go
	$(info $(M) Done!) @

run: all
	./bin/cmk

debug:
	$(GO) build -mod=vendor -gcflags='-N -l' -o cmk &&  dlv --listen=:2345 --headless=true --api-version=2 exec ./cmk

dist-mkdir: all
	rm -fr dist
	mkdir -p dist

dist-linux: dist-mkdir
	GOOS=linux   GOARCH=amd64 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.linux.x86-64 cmd/main.go
	GOOS=linux   GOARCH=386 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.linux.x86 cmd/main.go
	GOOS=linux   GOARCH=arm $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.linux.arm32 cmd/main.go
	GOOS=linux   GOARCH=arm64 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.linux.arm64 cmd/main.go


dist: dist-linux
	GOOS=windows GOARCH=386 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.windows.x86.exe cmd/main.go
	GOOS=windows GOARCH=amd64 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.windows.x86-64.exe cmd/main.go
	GOOS=darwin  GOARCH=amd64 $(GO) build -mod=vendor -ldflags='-s -w -X main.GitSHA=$(GIT_SHA) -X main.BuildDate=$(DATE)' -o dist/cmk.darwin.x86-64 cmd/main.go

# Tools

GOLINT = $(BIN)/golint
$(BIN)/golint:  $(BASE) ; $(info $(M) Building golint…)
	$Q go get github.com/golang/lint/golint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: | $(BASE) ; $(info $(M) building gocovmerge…)
	$Q go get github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: | $(BASE) ; $(info $(M) building gocov…)
	$Q go get github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: | $(BASE) ; $(info $(M) building gocov-xml…)
	$Q go get github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit: | $(BASE) ; $(info $(M) Building go2xunit…)
	$Q go get github.com/tebeka/go2xunit

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt lint vendor | $(BASE) ; $(info $(M) Running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(BASE) && $(GO) test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

test-xml: fmt lint vendor | $(BASE) $(GO2XUNIT) ; $(info $(M) Running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	$Q cd $(BASE) && 2>&1 $(GO) test -timeout 20s -v $(TESTPKGS) | tee test/tests.output
	$(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint vendor test-coverage-tools | $(BASE) ; $(info $(M) Running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE) && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) Running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) Running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -mod=vendor -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Misc

.PHONY: clean
clean: ; $(info $(M) Cleaning…)	@
	@rm -rf bin dist cloudstack-cloudmonkey
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo cloudmonkey-$(VERSION)
