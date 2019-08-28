/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */
package adabas

import (
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func BenchmarkConnection_noMultifetch(b *testing.B) {
	ferr := initLogWithFile("connection_bench.log")
	if ferr != nil {
		fmt.Println("Error creating log")
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", b.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(b, err) {
		return
	}
	if !assert.NotNil(b, connection) {
		return
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(b, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 1

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(b, qErr)
	result, rerr := readRequest.ReadPhysicalSequence()
	assert.NoError(b, rerr)
	if assert.NotNil(b, result) {
		assert.Equal(b, 1107, len(result.Values))
	}
}

func BenchmarkConnection_Multifetch(b *testing.B) {
	ferr := initLogWithFile("connection_bench.log")
	if ferr != nil {
		fmt.Println("Error creating log")
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", b.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(b, err) {
		return
	}
	if !assert.NotNil(b, connection) {
		return
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(b, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 10

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(b, qErr)
	var result *Response
	result, err = readRequest.ReadPhysicalSequence()
	assert.NoError(b, err)
	assert.Equal(b, 1107, len(result.Values))
}
