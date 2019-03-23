#
# Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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

PKGS            = $(or $(PKG),$(shell cd $(CURDIR) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "^vendor/"))
TESTPKGS        = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
CGO_CFLAGS      = $(if $(ACLDIR),-I$(ACLDIR)/inc,)
CGO_LDFLAGS     = $(if $(ACLDIR),-L$(ACLDIR)/lib -ladalnkx,)
CGO_EXT_LDFLAGS = $(if $(ACLDIR),-lsagsmp2 -lsagxts3 -ladazbuf,)
GO_TAGS         = $(if $(ACLDIR),"release adalnk","release")
GO_FLAGS        = $(if $(debug),"-x",) -tags $(GO_TAGS)

#export GOPATH

GO           = go
GODOC        = godoc
GOARCH      ?= $(shell $(GO) env GOARCH)
GOOS        ?= $(shell $(GO) env GOOS)
TIMEOUT      = 2000
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: prepare generate fmt lint lib $(EXECS) test-build

lib: $(LIBS) $(CEXEC)

exec: $(EXECS)

prepare: $(LOGPATH) $(CURLOGPATH) $(BIN)
	@echo "Build architecture ${GOARCH} ${GOOS} network=${WCPHOST} GOFLAGS=$(GO_FLAGS)"

$(LIBS): ; $(info $(M) building libraries…) @ ## Build program binary
	$Q cd $(CURDIR) && \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-buildmode=c-shared \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $(BIN)/$(GOOS)/$@.so $@.go

$(EXECS): ; $(info $(M) building executable $(@:$(BIN)/%=%)…) @ ## Build program binary
	$Q cd $(CURDIR) &&  \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $@ ./$(@:$(BIN)/%=%)

cleanModules:  ; $(info $(M) cleaning modules) @ ## Build program binary
ifneq ("$(wildcard $(GOPATH)/pkg/mod)","")
	$Q cd $(CURDIR) &&  \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) clean -modcache -cache ./...
endif

$(LOGPATH):
	@mkdir -p $@

$(CURLOGPATH):
	@mkdir -p $@

$(BIN):
	@mkdir -p $@
$(BIN)/%: $(BIN); $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GO111MODULE=off GOPATH=$$tmp CGO_CFLAGS= CGO_LDFLAGS= \
		go get $(REPOSITORY) && cp $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		# (GOPATH=$$tmp go clean -modcache ./...); \
		rm -rf $$tmp ; exit $$ret

# Tools
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

test-build: ; $(info $(M) building $(NAME:%=% )tests…) @ ## Build tests
	$Q cd $(CURDIR) && for pkg in $(TESTDIR); do echo "Build $$pkg in $(CURDIR)"; \
	LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
		DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) REFERENCES=$(REFERENCES) \
	    $(GO) test -c -tags $(GO_TAGS) ./$$pkg; done

TEST_TARGETS := test-default test-bench test-short test-verbose test-race test-sanitizer
.PHONY: $(TEST_TARGETS) check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
test-sanitizer:  ARGS=-msan      ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt lint ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(CURDIR) && LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
		DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) REFERENCES=$(REFERENCES) \
	    $(GO) test -timeout $(TIMEOUT)s -v -tags $(GO_TAGS) $(ARGS) ./...

TEST_XML_TARGETS := test-xml-bench
.PHONY: $(TEST_XML_TARGETS) test-xml
test-xml-bench:     ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
$(TEST_XML_TARGETS): NAME=$(MAKECMDGOALS:test-xml-%=%)
$(TEST_XML_TARGETS): test-xml
test-xml: prepare fmt lint $(TESTOUTPUT) | $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	sh $(CURDIR)/sh/evaluateQueues.sh
	$Q cd $(CURDIR) && 2>&1 TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) \
	    REFERENCES=$(REFERENCES) LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    ENABLE_DEBUG=$(ENABLE_DEBUG) WCPHOST=$(WCPHOST) ADATCPHOST=$(ADATCPHOST) ADAMFDBID=$(ADAMFDBID) \
	    $(GO) test -timeout $(TIMEOUT)s -count=1 $(GO_FLAGS) -v $(ARGS) ./... | tee $(TESTOUTPUT)/tests.output
	sh $(CURDIR)/sh/evaluateQueues.sh
	$(GO2XUNIT) -input $(TESTOUTPUT)/tests.output -output $(TESTOUTPUT)/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage
test-coverage: fmt lint test-coverage-tools ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q echo "Work on test packages: $(TESTPKGS)"
	$Q cd $(CURDIR) && for pkg in $(TESTPKGS); do echo "Coverage for $$pkg"; \
		TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) \
	    REFERENCES=$(REFERENCES) LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    ENABLE_DEBUG=$(ENABLE_DEBUG) WCPHOST=$(WCPHOST) ADATCPHOST=$(ADATCPHOST) ADAMFDBID=$(ADAMFDBID) \
		$(GO) test -count=1 \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) -timeout $(TIMEOUT)s $(GO_FLAGS) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q echo "Start coverage analysis"
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: | $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(CURDIR) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running fmt…) @ ## Run go fmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GO) fmt  $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

.PHONY: generate
generate: ; $(info $(M) generating messages…) @ ## Generate message go code
	$Q cd $(CURDIR)/generate && 2>&1 CURDIR=$(CURDIR) GO_ADA_MESSAGES=$(MESSAGES) \
	                      $(GO) generate -v $(GO_FLAGS)

# Misc
.PHONY: clean
clean: cleanModules; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH) $(CURDIR)/adabas/vendor
	@rm -rf $(BIN) $(CURDIR)/pkg $(CURDIR)/logs $(CURDIR)/test
	@rm -rf test/tests.* test/coverage.*
	@rm -f $(CURDIR)/adabas.test $(CURDIR)/adatypes.test $(CURDIR)/*.log $(CURDIR)/*.output

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: doc
doc: ; $(info $(M) running GODOC…) @ ## Run go doc on all source files
	$Q cd $(CURDIR) && echo "Open http://localhost:6060/pkg/github.com/SoftwareAG/adabas-go-api/" && \
	   GOPATH=$(GOPATH) \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GODOC) -http=:6060 -v -src
#	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) doc $(PACKAGE)

.PHONY: vendor-update
vendor-update:
	@echo "Uses GO modules"

.PHONY: version
version:
	@echo $(VERSION)
