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
)

func ExampleReadRequest_ReadLogicalWithCursoring() {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println("Error initializing log", err)
		return
	}
	defer f.Close()

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
	request.QueryFields("NAME,PERSONNEL-ID")
	request.Limit = 0
	fmt.Println("Init cursor data...")
	col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100110:11100115]")
	if cerr != nil {
		fmt.Println("Error reading logical with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
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

func ExampleReadRequest_ReadLogicalWithCursoringLimit() {
	f, ferr := initLogWithFile("connection_map.log")
	if ferr != nil {
		fmt.Println("Error initializing log", ferr)
		return
	}
	defer f.Close()

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
	request.QueryFields("NAME,PERSONNEL-ID")
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
