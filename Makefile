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

PACKAGE       = github.com/SoftwareAG/adabas-go-api
TESTPKGSDIR   = adabas adatypes

GOARCH      ?= $(shell $(GO) env GOARCH)
GOOS        ?= $(shell $(GO) env GOOS)

DATE         ?= $(shell date +%FT%T%z)
VERSION      ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN           = $(CURDIR)/bin/$(GOOS)_$(GOARCH)
LOGPATH       = $(CURDIR)/logs
TESTFILES     = $(CURDIR)/files
REFERENCES    = $(TESTFILES)/references
MESSAGES      = $(CURDIR)/messages
CURLOGPATH    = $(CURDIR)/logs

# Test parameter
WCPHOST      ?= wcphost:30011
ADATCPHOST   ?= tcphost:60177
ADAMFDBID    ?= 54712
TESTOUTPUT   ?= $(CURDIR)/test
ENABLE_DEBUG ?= 0

# Executables
EXECS         = $(BIN)/tests/employee_client $(BIN)/tests/testsuite $(BIN)/tests/simple_read $(BIN)/tests/query \
    $(BIN)/tests/querym $(BIN)/tests/lobload $(BIN)/tests/clear_map_reference
LIBS          = 

include $(CURDIR)/make/common.mk
