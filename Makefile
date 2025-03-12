#
# Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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

PACKAGE       = github.com/SoftwareAG/adabas-go-api
TESTPKGSDIR   = adabas adatypes

GOARCH      ?= $(shell $(GO) env GOARCH)
GOOS        ?= $(shell $(GO) env GOOS)
GOPATH      ?= $(shell $(GO) env GOPATH)

DATE         ?= $(shell date +%FT%T%z)
VERSION      ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN           = $(CURDIR)/bin/$(GOOS)_$(GOARCH)
BINTOOLS      = $(CURDIR)/bin/tools/$(GOOS)_$(GOARCH)
INSTALLTOOLS  = $(GOPATH)/bin
LOGPATH       = $(CURDIR)/logs
TESTFILES     = $(CURDIR)/files
OBJECTS       = adabas/*.go adatypes/*.go examples/*/*.go
REFERENCES    = $(TESTFILES)/references
MESSAGES      = $(CURDIR)/messages

# Test parameter
WCPHOST      ?= wcphost:30011
ADATCPHOST   ?= tcphost:60177
ADAMFDBID    ?= 
TESTOUTPUT   ?= $(CURDIR)/test
ENABLE_DEBUG ?= 0
DOCKER_GO     = 1.12

# Executables
EXECS         = $(BIN)/tools/employee_client $(BIN)/tools/testsuite $(BIN)/tools/simple_read $(BIN)/tools/query \
    $(BIN)/tools/querym $(BIN)/tools/lobload $(BIN)/tools/clear_map_reference \
	$(BIN)/tools/betacluster $(BIN)/examples/employees $(BIN)/examples/employees_map $(BIN)/examples/employees_struct
LIBS          = 

include $(CURDIR)/make/common.mk

docker: docker/Dockerfile .FORCE
	cd docker; echo "Create docker image $(DOCKER_GO)"; docker build -t adago:$(DOCKER_GO) .; \
       docker run -v `pwd`/..:/data adago:$(DOCKER_GO) make

.FORCE:
