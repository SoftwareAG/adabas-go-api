#
# Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
#
# SPDX-License-Identifier: Apache-2.0
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#

PACKAGE     = github.com/SoftwareAG/adabas-go-api
DATE       ?= $(shell date +%FT%T%z)
VERSION    ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
TMPPATH    ?= /tmp
GOPATH      = $(TMPPATH)/tmp_gopath.$(shell id -u)
BIN         = $(CURDIR)/bin
LOGPATH     = $(CURDIR)/logs
TESTFILES   = $(CURDIR)/files
REFERENCES  = $(TESTFILES)/references
MESSAGES    = $(CURDIR)/messages
CURLOGPATH  = $(CURDIR)/logs
WCPHOST    ?= wcphost:30011
ADATCPHOST ?= tcphost:60177
TESTOUTPUT  = $(CURDIR)/test
EXECS       = tests/employee_client tests/testsuite tests/simple_read
LIBS        = slib/adaapi
BASE        = $(GOPATH)/src/$(PACKAGE)
BASESRC     = $(CURDIR)
PKGS        = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "^vendor/"))
TESTPKGS    = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
CGO_CFLAGS  = $(if $(ACLDIR),-I$(ACLDIR)/inc,)
CGO_LDFLAGS = $(if $(ACLDIR),-L$(ACLDIR)/lib -ladalnkx,)
CGO_EXT_LDFLAGS = $(if $(ACLDIR),-lsagsmp2 -lsagxts3 -ladazbuf,)
GO_TAGS     = $(if $(ACLDIR),"release adalnk","release")
GO_FLAGS    = $(if $(debug),"-x",) -tags $(GO_TAGS)

export GOPATH

GO      = go
GODOC   = godoc
GOFMT   = gofmt
GOARCH   ?= $(shell $(GO) env GOARCH)
GOOS     ?= $(shell $(GO) env GOOS)
TIMEOUT = 60
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: prepare generate fmt lint vendor $(EXECS)

lib: all $(LIBS)

prepare: $(LOGPATH) $(CURLOGPATH) $(BIN) $(BASE)
	@echo "Build architecture ${GOARCH} ${GOOS} network=${WCPHOST} GOFLAGS=$(GO_FLAGS)"

$(LIBS): | $(BASE) ; $(info $(M) building libraries…) @ ## Build program binary
	$Q cd $(BASE) && \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-buildmode=c-shared \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $(BIN)/$@ $@.go

$(EXECS): | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && \
	echo $(BIN)/$@ && \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $(BIN)/$@ $@.go

$(BASE): ; $(info $(M) setting GOPATH…)
	@mkdir -p $(dir $@)
#	cp -r $(CURDIR)/$(PACKAGE) $@
	ln -sf $(CURDIR) $@

# Tools

$(LOGPATH):
	@mkdir -p $@

$(CURLOGPATH):
	@mkdir -p $@

$(BIN):
	@mkdir -p $@
$(BIN)/%: $(BIN) | $(BASE) ; $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp go get $(REPOSITORY) && cp $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		rm -rf $$tmp ; exit $$ret

GODEP = $(BIN)/dep
$(BIN)/dep: REPOSITORY=github.com/golang/dep/cmd/dep

GOLINT = $(BIN)/golint
$(BIN)/golint: REPOSITORY=golang.org/x/lint/golint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: REPOSITORY=github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: REPOSITORY=github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: REPOSITORY=github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit: REPOSITORY=github.com/tebeka/go2xunit

# Tests
$(TESTOUTPUT):
	mkdir $(TESTOUTPUT)

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt lint vendor | $(BASE) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(BASE) && LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(ACLDIR)/lib" \
		DYLD_LIBRARY_PATH="$DYLD_LIBRARY_PATH:$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) REFERENCES=$(REFERENCES) \
	    $(GO) test -timeout $(TIMEOUT)s -v -tags $(GO_TAGS) $(ARGS) $(TESTPKGS)

TEST_XML_TARGETS := test-xml-bench
.PHONY: $(TEST_XML_TARGETS) test-xml
test-xml-bench:     ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
$(TEST_XML_TARGETS): NAME=$(MAKECMDGOALS:test-xml-%=%)
$(TEST_XML_TARGETS): test-xml
test-xml: prepare fmt lint vendor $(TESTOUTPUT) | $(BASE) $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	sh $(CURDIR)/sh/evaluteQueues.sh
	$Q cd $(BASE) && 2>&1 TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) \
	    REFERENCES=$(REFERENCES) LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    ENABLE_DEBUG=0 WCPHOST=$(WCPHOST) ADATCPHOST=$(ADATCPHOST) \
	    $(GO) test -timeout $(TIMEOUT)s $(GO_FLAGS) -v $(ARGS) $(TESTPKGS) | tee $(TESTOUTPUT)/tests.output
	sh $(CURDIR)/sh/evaluteQueues.sh
	$(GO2XUNIT) -input $(TESTOUTPUT)/tests.output -output $(TESTOUTPUT)/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint vendor test-coverage-tools | $(BASE) ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE) && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) -timeout $(TIMEOUT)s \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

vendor: Gopkg.toml Gopkg.lock | $(BASE) $(GODEP) ; $(info $(M) retrieving dependencies…)
	$Q cd $(GOPATH)/src/$(PACKAGE) && $(GODEP) ensure -v
	@echo "$(GOPATH)"
#	@ln -nsf . vendor/src
	@touch $@

.PHONY: vendor-update
vendor-update: vendor | $(BASE) $(GODEP)
ifeq "$(origin PKG)" "command line"
	$(info $(M) updating $(PKG) dependency…)
	$Q cd $(BASE) && $(GODEP) ensure -update $(PKG)
else
	$(info $(M) updating all dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure -update
endif
#	@ln -nsf . vendor/src
	@touch vendor

.PHONY: generate
generate: ; $(info $(M) generating…) @ ## Generate message go repository
	$Q cd $(BASESRC)/generate && 2>&1 CURDIR=$(CURDIR) GO_ADA_MESSAGES=$(MESSAGES) \
	                      $(GO) generate -v $(GO_FLAGS)

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH) $(CURDIR)/adabas/vendor
	@rm -rf $(BIN) $(CURDIR)/pkg $(CURDIR)/logs $(CURDIR)/test
	@rm -rf test/tests.* test/coverage.*
	@rm -rf $(BASESRC)/vendor $(BASESRC)/.vendor-new $(CURDIR)/vendor
#	$(GO) clean -cache

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: doc
doc: ; $(info $(M) running GODOC…) @ ## Run go doc on all source files
	$Q cd $(BASESRC) && \
	   GOPATH=$(GOPATH) \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GODOC) -http=:6060 -v
#	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) doc $(PACKAGE)

.PHONY: version
version:
	@echo $(VERSION)
