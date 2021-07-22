/*
* Copyright © 2020-2021 Software AG, Darmstadt, Germany and/or its licensors
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

	"github.com/stretchr/testify/assert"
)

func TestStreamValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "stream.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("LOBPICTURE")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	// Read all data at once as reference
	rerr = request.QueryFields("Picture")
	if !assert.NoError(t, rerr) {
		return
	}
	record, err := request.ReadLogicalWith("Filename=p1.jpg")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, record.NrRecords())
	refValue, err := record.Values[0].SearchValue("Picture")
	refData := refValue.Bytes()
	assert.Equal(t, 769996, len(refData))

	// Now read stream and compare parts
	request, rerr = connection.CreateMapReadRequest("LOBPICTURE")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	cursor, derr := request.ReadLobStream("Filename=p1.jpg", "Picture")
	if !assert.NoError(t, derr) {
		fmt.Println("Error read LOB segment", derr)
		return
	}
	if !assert.NotNil(t, cursor) {
		return
	}
	blockCount := 0
	for cursor.HasNextRecord() {
		record, err := cursor.NextRecord()
		if !assert.NoError(t, err, fmt.Sprintf("invalid on block count %d", blockCount)) {
			fmt.Println("Error read next LOB segment", err)
			return
		}
		blockCount++
		f := record.HashFields["Picture"]
		if !assert.NotNil(t, f, fmt.Sprintf("Picture field not found on block count %d", blockCount)) {
			fmt.Println("Hashfields:", record.HashFields)
			return
		}
		data := f.Bytes()
		if !assert.NotNil(t, data, fmt.Sprintf("invalid on block count %d", blockCount)) {
			return
		}
		if !assert.True(t, len(data) == defaultBlockSize, fmt.Sprintf("Invalid len = %d on block count %d", len(data), blockCount)) {
			return
		}
		if !assert.Equal(t, refData[(blockCount-1)*defaultBlockSize:blockCount*defaultBlockSize], data, "Data not correct") {
			return
		}
	}
	assert.Equal(t, 13, blockCount)
}

func TestLOBSegment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "stream.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("LOBPICTURE")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	data, derr := request.ReadLOBSegment(4, "Picture", 8096)
	if !assert.NoError(t, derr) {
		fmt.Println("Error read LOB segment", derr)
		return
	}
	assert.NotNil(t, data)
	assert.True(t, len(data) == 8096, fmt.Sprintf("Invalid len = %d", len(data)))
	data2, derr2 := request.ReadLOBSegment(4, "Picture", 8096)
	if !assert.NoError(t, derr2) {
		fmt.Println("Error read LOB segment", derr)
		return
	}
	assert.NotEqual(t, data, data2)
}
