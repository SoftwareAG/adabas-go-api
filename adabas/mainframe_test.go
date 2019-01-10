/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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
	"os"
)

func ExampleReadRequest_fileMf() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
	if err != nil {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("AA=[11100301:11100303]", nil, result)
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0251
	// Record Quantity: 0003
	//   AA = > 11100301 <
	//   AB = [ 1 ]
	//    AC = > HANS                 <
	//    AE = > BERGMANN             <
	//    AD = > WILHELM              <
	// Record Isn: 0383
	// Record Quantity: 0003
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <

}

func ExampleReadRequest_fileMfBorder() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
	if err != nil {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("AA=(11100301:11100303)", nil, result)
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0383
	// Record Quantity: 0001
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <

}

func ExampleReadRequest_fileMfNoMinimum() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
	if err != nil {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("AA=(11100301:11100303]", nil, result)
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0383
	// Record Quantity: 0002
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <
	// Record Isn: 0252
	// Record Quantity: 0002
	//   AA = > 11100303 <
	//   AB = [ 1 ]
	//    AC = > KRISTINA             <
	//    AE = > FALTER               <
	//    AD = > MARIA                <

}

func ExampleReadRequest_fileMfNoMaximum() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
	if err != nil {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("AA=[1100301:11100303)", nil, result)
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0204
	// Record Quantity: 0017
	//   AA = > 11100102 <
	//   AB = [ 1 ]
	//    AC = > EDGAR                <
	//    AE = > SCHINDLER            <
	//    AD = > PETER                <
	// Record Isn: 0205
	// Record Quantity: 0017
	//   AA = > 11100105 <
	//   AB = [ 1 ]
	//    AC = > CHRISTIAN            <
	//    AE = > SCHIRM               <
	//    AD = >                      <

}
