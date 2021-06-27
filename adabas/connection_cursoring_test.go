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

func ExampleReadRequest_ReadLogicalWithCursoring() {
	err := initLogWithFile("connection_cursoring.log")
	if err != nil {
		fmt.Println("Error initializing log", err)
		return
	}

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Limit query data:")
	err = request.QueryFields("NAME,PERSONNEL-ID")
	if err != nil {
		fmt.Println("Error query fields", err)
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100110:11100115]")
	if cerr != nil {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		if rerr != nil {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		fmt.Println("Record received:")
		record.DumpValues()
		fmt.Println("Read next cursor record...")
		counter++
		if counter >= 7 {
			fmt.Println("Error index about 7")
			return
		}
	}
	fmt.Println("Last cursor record read")

	// Output: Limit query data:
	// Init cursor data...
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100110 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > BUNGERT              <
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100111 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THIELE               <
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100112 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THOMA                <
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100113 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > TREIBER              <
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100114 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > UNGER                <
	// Read next cursor record...
	// Record received:
	// Dump all record values
	//   PERSONNEL-ID = > 11100115 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > VETTER               <
	// Read next cursor record...
	// Last cursor record read

}

func ExampleReadRequest_readLogicalWithCursoringLimit() {
	ferr := initLogWithFile("connection_cursoring.log")
	if ferr != nil {
		fmt.Println("Error initializing log", ferr)
		return
	}

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error creating read request", cerr)
		return
	}
	// Define fields to be part of the request
	err := request.QueryFields("NAME,PERSONNEL-ID")
	if err != nil {
		fmt.Println("Error query field read request", err)
		return
	}
	// Define chunks of cursoring requests
	request.Limit = 5

	// Init cursoring using search
	col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100110:11100120]")
	if cerr != nil {
		fmt.Println("Error init cursoring", cerr)
		return
	}
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if rerr != nil {
			fmt.Println("Error getting next record", rerr)
			return
		}
		fmt.Printf("New record received: ISN=%d\n", record.Isn)
		record.DumpValues()
	}

	// Output: New record received: ISN=210
	// Dump all record values
	//   PERSONNEL-ID = > 11100110 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > BUNGERT              <
	// New record received: ISN=211
	// Dump all record values
	//   PERSONNEL-ID = > 11100111 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THIELE               <
	// New record received: ISN=212
	// Dump all record values
	//   PERSONNEL-ID = > 11100112 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THOMA                <
	// New record received: ISN=213
	// Dump all record values
	//   PERSONNEL-ID = > 11100113 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > TREIBER              <
	// New record received: ISN=214
	// Dump all record values
	//   PERSONNEL-ID = > 11100114 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > UNGER                <
	// New record received: ISN=1102
	// Dump all record values
	//   PERSONNEL-ID = > 11100115 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > VETTER               <
	// New record received: ISN=215
	// Dump all record values
	//   PERSONNEL-ID = > 11100116 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > VOGEL                <
	// New record received: ISN=216
	// Dump all record values
	//   PERSONNEL-ID = > 11100117 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > WABER                <
	// New record received: ISN=217
	// Dump all record values
	//   PERSONNEL-ID = > 11100118 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > WAGNER               <

}

func TestReadLogicalWithCursoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Limit query data:")
	rerr = request.QueryFields("NAME,PERSONNEL-ID")
	if !assert.NoError(t, rerr) {
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[0:9]")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		adatypes.Central.Log.Debugf("Read next cursor record...%d", counter)
	}
	assert.Equal(t, 1107, counter)
	fmt.Println("Last cursor record read")

}

func TestSearchAndReadWithCursoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Limit query data:")
	rerr = request.QueryFields("NAME,PERSONNEL-ID")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read query fields", rerr)
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.SearchAndOrderWithCursoring("PERSONNEL-ID=[1:5]", "NAME")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		switch counter {
		case 1:
			assert.Equal(t, "30000231", record.HashFields["PERSONNEL-ID"].String())
			assert.Equal(t, "ACHIESON            ", record.HashFields["NAME"].String())
		case 10:
			assert.Equal(t, "11300313", record.HashFields["PERSONNEL-ID"].String())
			assert.Equal(t, "AECKERLE            ", record.HashFields["NAME"].String())
		case 100:
			assert.Equal(t, "11100330", record.HashFields["PERSONNEL-ID"].String())
			assert.Equal(t, "BUSH                ", record.HashFields["NAME"].String())
		case 800:
			assert.Equal(t, "30021107", record.HashFields["PERSONNEL-ID"].String())
			assert.Equal(t, "WORTH               ", record.HashFields["NAME"].String())
		}
		adatypes.Central.Log.Debugf("Read next cursor record...%d", counter)
	}
	assert.Equal(t, 807, counter)
	fmt.Println("Last cursor record read")

}

func TestSearchAndReadWithCursoringEmplStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	err := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, err) {
		return
	}
	err = copyAdabasFile(t, "*", adabasModDBIDs, 11, 16)
	if !assert.NoError(t, err) {
		return
	}

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest((*Employees)(nil))
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Limit query data:")
	err = request.QueryFields("Name,ID")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.SearchAndOrderWithCursoring("ID=[500041:500050]", "Name")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		entry, rerr := col.NextData()
		if !assert.NotNil(t, entry) {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		e := entry.(*Employees)
		switch counter {
		case 1:
			assert.Equal(t, "50004900", e.ID)
			assert.Equal(t, "CAOUDAL             ", e.Name)
		case 5:
			assert.Equal(t, "50004600", e.ID)
			assert.Equal(t, "VERDIE              ", e.Name)
		}
		adatypes.Central.Log.Debugf("Read next cursor record...%d", counter)
	}
	assert.Equal(t, 5, counter)
	fmt.Println("Last cursor record read")

}

func TestSearchAndReadWithCursoringEmplStructEmptyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	_ = refreshFile(adabasModDBIDs, 16)

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest((*Employees)(nil))
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Limit query data:")
	cerr = request.QueryFields("Name,ID")
	if !assert.NoError(t, cerr) {
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.SearchAndOrderWithCursoring("ID=[1:5]", "Name")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		entry, rerr := col.NextData()
		if !assert.NotNil(t, entry) {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		e := entry.(*Employees)
		switch counter {
		case 1:
			assert.Equal(t, "30000231", e.ID)
			assert.Equal(t, "ACHIESON            ", e.Name)
		case 10:
			assert.Equal(t, "11300313", e.ID)
			assert.Equal(t, "AECKERLE            ", e.Name)
		case 100:
			assert.Equal(t, "11100330", e.ID)
			assert.Equal(t, "BUSH                ", e.Name)
		case 800:
			assert.Equal(t, "30021107", e.ID)
			assert.Equal(t, "WORTH               ", e.Name)
		}
		adatypes.Central.Log.Debugf("Read next cursor record...%d", counter)
	}
	assert.Equal(t, 0, counter)
	fmt.Println("Last cursor record read")

}

func TestHistogramWithCursoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.HistogramWithCursoring("NAME=[A:B]")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		switch counter {
		case 1:
			assert.Equal(t, "ABELLAN             ", record.HashFields["NAME"].String())
			assert.Equal(t, uint64(1), record.Quantity)
		case 3:
			assert.Equal(t, "ADAM                ", record.HashFields["NAME"].String())
			assert.Equal(t, uint64(1), record.Quantity)
		case 10:
			assert.Equal(t, "ALESTIA             ", record.HashFields["NAME"].String())
			assert.Equal(t, uint64(1), record.Quantity)
		}
	}
	assert.Equal(t, 10, counter)
	fmt.Println("Last cursor record read")

}

func TestPhysicalWithCursoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	err := request.QueryFields("PERSONNEL-ID,NAME")
	if !assert.NoError(t, err) {
		fmt.Println("Error define query fields:", err)
		return
	}
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.ReadPhysicalWithCursoring()
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading physical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading physical with using cursoring", rerr)
			return
		}
		switch counter {
		case 1:
			assert.Equal(t, "ADAM                ", record.HashFields["NAME"].String())
			assert.Equal(t, adatypes.Isn(1), record.Isn)
		case 3:
			assert.Equal(t, "BLOND               ", record.HashFields["NAME"].String())
			assert.Equal(t, adatypes.Isn(3), record.Isn)
		case 10:
			assert.Equal(t, "MONTASSIER          ", record.HashFields["NAME"].String())
			assert.Equal(t, adatypes.Isn(10), record.Isn)
		case 300:
			assert.Equal(t, "YALCIN              ", record.HashFields["NAME"].String())
			assert.Equal(t, adatypes.Isn(0x12c), record.Isn)
		}
	}
	assert.Equal(t, 1048, counter)
	fmt.Println("Last cursor record read")

}
