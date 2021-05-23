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
BINTESTS        = $(CURDIR)/bin/tests/$(GOOS)_$(GOARCH)
DYLD_LIBRARY_PATH = $(ACLDIR)/lib:/lib:/usr/lib:$(ACLDIR)/../common/security/openssl/lib
GOEXE          ?= $(shell $(GO) env GOEXE)

export DYLD_LIBRARY_PATH

GO           = go
GODOC        = godoc
TIMEOUT      = 2000
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

export TIMEOUT GO CGO_CFLAGS CGO_LDFLAGS GO_TAGS

.PHONY: all
all: prepare generate fmt lint lib $(EXECS) test-build

lib: $(LIBS) $(CEXEC)

exec: $(EXECS)

prepare: $(LOGPATH) $(BIN)
	@echo "Build architecture ${GOARCH} ${GOOS} network=${WCPHOST} GOFLAGS=$(GO_FLAGS)"

$(LIBS): ; $(info $(M) building libraries…) @ ## Build program binary
	$Q cd $(CURDIR) && \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-buildmode=c-shared \
		-ldflags '-X $(PACKAGE)/adabas.Version=$(VERSION) -X $(PACKAGE)/adabas.BuildDate=$(DATE)' \
		-o $(BIN)/$(GOOS)/$@.so $@.go

$(EXECS): $(OBJECTS) ; $(info $(M) building executable $(@:$(BIN)/%=%)…) @ ## Build program binary
	$Q cd $(CURDIR) &&  \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) build $(GO_FLAGS) \
		-ldflags '-X $(PACKAGE)/adabas.Version=$(VERSION) -X $(PACKAGE)/adabas.BuildDate=$(DATE)' \
		-o $@$(GOEXE) ./$(@:$(BIN)/%=%)

cleanModules:  ; $(info $(M) cleaning modules) @ ## Build program binary
ifneq ("$(wildcard $(GOPATH)/pkg/mod)","")
	$Q cd $(CURDIR) &&  \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) clean -modcache -cache ./...
endif

$(LOGPATH):
	@mkdir -p $@

$(BIN):
	@mkdir -p $@

$(BIN)/%: $(BIN); $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GO111MODULE=off GOPATH=$$tmp CGO_CFLAGS= CGO_LDFLAGS= \
		go get $(REPOSITORY) && cp -r $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		# (GOPATH=$$tmp go clean -modcache ./...); \
		rm -rf $$tmp ; exit $$ret

$(BINTOOLS):
	@mkdir -p $@
$(BINTOOLS)/%: $(BINTOOLS) ; $(info $(M) building tool $(BINTOOLS) on $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp CGO_CFLAGS= CGO_LDFLAGS= \
		go get $(REPOSITORY) && find $$tmp/bin -type f -exec cp {} $(BINTOOLS)/. \;) || ret=$$?; \
		(GOPATH=$$tmp go clean -modcache ./...); \
		rm -rf $$tmp ; exit $$ret

$(BINTESTS):
	@mkdir -p $@

# Tools
GOLINT = $(BINTOOLS)/golint
$(BINTOOLS)/golint: REPOSITORY=golang.org/x/lint/golint

GOCILINT = $(BINTOOLS)/golangci-lint
$(BINTOOLS)/golangci-lint: REPOSITORY=github.com/golangci/golangci-lint/cmd/golangci-lint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: REPOSITORY=github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: REPOSITORY=github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: REPOSITORY=github.com/AlekSi/gocov-xml

GOTESTSUM = $(BIN)/gotestsum
$(BIN)/gotestsum: REPOSITORY=gotest.tools/gotestsum

# Tests
$(TESTOUTPUT):
	mkdir $(TESTOUTPUT)

test-build: $(OBJECTS) $(BINTESTS) ; $(info $(M) building $(NAME:%=% )tests…) @ ## Build tests
	$Q cd $(CURDIR) && for pkg in $(TESTPKGSDIR); do echo "Build $$pkg in $(CURDIR)"; \
	LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
		DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) REFERENCES=$(REFERENCES) \
	    $(GO) test -c -o $(BINTESTS)/$$pkg.test -tags $(GO_TAGS) ./$$pkg; done

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
	    $(GO) test -timeout $(TIMEOUT)s -count=1 -v -tags $(GO_TAGS) $(ARGS) ./...

TEST_XML_TARGETS := test-xml-bench
.PHONY: $(TEST_XML_TARGETS) test-xml
test-xml-bench:     ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
$(TEST_XML_TARGETS): NAME=$(MAKECMDGOALS:test-xml-%=%)
$(TEST_XML_TARGETS): test-xml
test-xml: prepare fmt lint $(TESTOUTPUT) | $(GOTESTSUM) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	sh $(CURDIR)/scripts/evaluateQueues.sh
	$Q cd $(CURDIR) && 2>&1 TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) \
	    REFERENCES=$(REFERENCES) LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    ENABLE_DEBUG=$(ENABLE_DEBUG) WCPHOST=$(WCPHOST) ADATCPHOST=$(ADATCPHOST) ADAMFDBID=$(ADAMFDBID) \
	    $(GOTESTSUM) --junitfile $(TESTOUTPUT)/tests.xml --raw-command -- $(CURDIR)/scripts/test.sh $(ARGS) ||:
	sh $(CURDIR)/scripts/evaluateQueues.sh

TEST_PPROF_TARGETS := test-adatypes-pprof test-adabas-pprof
.PHONY: $(TEST_PPROF_TARGETS) test-pprof
test-adatypes-pprof: ARGS= ## Run pprof for adatypes
test-adabas-pprof:   ARGS= ## Run pprof for adabas
$(TEST_PPROF_TARGETS): NAME=$(MAKECMDGOALS:test-%-pprof=%)
$(TEST_PPROF_TARGETS): test-pprof
test-pprof: prepare $(TESTOUTPUT) | ; $(info $(M) running $(NAME:%=% )pprof…) @ ## Run pprof
	$Q cd $(CURDIR) && 2>&1 TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) \
	    REFERENCES=$(REFERENCES) LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    ENABLE_DEBUG=$(ENABLE_DEBUG) WCPHOST=$(WCPHOST) ADATCPHOST=$(ADATCPHOST) ADAMFDBID=$(ADAMFDBID) \
	    $(GO) test -timeout $(TIMEOUT)s -count=1 -memprofile $(NAME)-memprofile.out -cpuprofile $(NAME)-profile.out $(GO_FLAGS) -v $(ARGS) ./$(NAME) 2>&1 | tee $(TESTOUTPUT)/tests.$(HOST).output

TEST_SINGLE_TARGETS := test-adabas-single test-adatypes-single
.PHONY: $(TEST_SINGLE_TARGETS) test-single
test-adabas-signle:   ARGS=-run=$(TEST)  ## Run single adabas tests
$(TEST_SINGLE_TARGETS): NAME=$(MAKECMDGOALS:test-%-single=%)
$(TEST_SINGLE_TARGETS): test-single
test-single: ; $(info $(M) running $(NAME:%=% )single test $(TEST)…) @ ## Run single tests
	$Q cd $(CURDIR) && LD_LIBRARY_PATH="$(LD_LIBRARY_PATH):$(ACLDIR)/lib" \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" \
	    TESTFILES=$(TESTFILES) GO_ADA_MESSAGES=$(MESSAGES) LOGPATH=$(LOGPATH) REFERENCES=$(REFERENCES) \
	    $(GO) test -timeout $(TIMEOUT)s -count=1 -v -tags $(GO_TAGS) $(ARGS) ./$(NAME)

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_COB_XML = $(COVERAGE_DIR)/coverage-cobertura.xml
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
	$Q cd $(CURDIR) && $(GOLINT) -set_exit_status ./...

.PHONY: cilint
cilint: | $(GOCILINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(CURDIR) && $(GOCILINT) run ./...

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
	@rm -rf $(CURDIR)/adabas/vendor $(CURDIR)/vendor
	@rm -rf $(CURDIR)/bin $(CURDIR)/pkg $(CURDIR)/logs $(CURDIR)/test
	@rm -rf test/tests.* test/coverage.*
	@rm -f $(BINTESTS) $(CURDIR)/*.log $(CURDIR)/*.output


.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/^[^\:]*://g' | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: doc
doc: ; $(info $(M) running GODOC…) @ ## Run go doc on all source files
	$Q cd $(CURDIR) && echo "Open http://localhost:6060/pkg/github.com/SoftwareAG/adabas-go-api/" && \
	   GOPATH=$(GOPATH) \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GODOC) -http=:6060 -v
#	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS) $(CGO_EXT_LDFLAGS)" $(GO) doc $(PACKAGE)

.PHONY: vendor-update
vendor-update:
	@echo "Uses GO modules"

.PHONY: version
version:
	@echo $(VERSION)
