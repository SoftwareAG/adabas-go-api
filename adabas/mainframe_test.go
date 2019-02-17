/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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

func ExampleConnection_readMainframe() {
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
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[11100301:11100303]")
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

func ExampleConnection_readBorderMainframe() {
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
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=(11100301:11100303)")
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

func ExampleConnection_readNoMinimumMainframe() {
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
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=(11100301:11100303]")
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

func ExampleConnection_readNoMaximumMainframe() {
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
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[1100301:11100303)")
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

func ExampleConnection_periodGroupMfPart() {
	f, _ := initLogWithFile("connection.log")
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[54711,4]")
	if cerr != nil {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error open connection", cerr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	request.QueryFields("PERSONNEL-ID,INCOME")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100303:11100304]")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = result.DumpValues()
	if err != nil {
		fmt.Println("Error dump values", err)
	}

	// Output: Dump all result values
	// Record Isn: 0252
	//   PERSONNEL-ID = > 11100303 <
	//   INCOME = [ 3 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 21846 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 1717 <
	//     BONUS[01,02] = > 3000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 21025 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1538 <
	//    CURR-CODE[03] = > EUR <
	//    SALARY[03] = > 20307 <
	//    BONUS[03] = [ 1 ]
	//     BONUS[03,01] = > 1282 <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   INCOME = [ 2 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 25230 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 2256 <
	//     BONUS[01,02] = > 2000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 24102 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1948 <

}
