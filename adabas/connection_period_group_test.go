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

func ExampleConnection_periodGroup2() {
	initLogWithFile("connection.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
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

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
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
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > KRISTINA             <
	//    NAME = > FALTER               <
	//    MIDDLE-I = > M <
	//   MAR-STAT = > M <
	//   SEX = > F <
	//   BIRTH = > 1961/07/08 <
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 3 ]
	//     ADDRESS-LINE[01] = > C/O CLAASEN          <
	//     ADDRESS-LINE[02] = > WIESENGRUND 10       <
	//     ADDRESS-LINE[03] = > 6100 DARMSTADT       <
	//    CITY = > DARMSTADT            <
	//    ZIP = > 6100       <
	//    COUNTRY = > D   <
	//   TELEPHONE = [ 1 ]
	//    AREA-CODE = > 06151  <
	//    PHONE = > 453897          <
	//   DEPT = > FINA21 <
	//   JOB-TITLE = > TYPISTIN                  <
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
	//   LEAVE-DATA = [ 1 ]
	//    LEAVE-DUE = > 30 <
	//    LEAVE-TAKEN = > 3 <
	//   LEAVE-BOOKED = [ 1 ]
	//    LEAVE-START[01] = > 19980520 <
	//    LEAVE-END[01] = > 19980523 <
	//   LANG = [ 1 ]
	//    LANG[01] = > GER <
	//   LEAVE-LEFT = > 3003 <
	//   DEPARTMENT = > FINA <
	//   DEPT-PERSON = > FINA21FALTER               <
	//   CURRENCY-SALARY = >  <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > WOLFGANG             <
	//    NAME = > SCHMIDT              <
	//    MIDDLE-I = > J <
	//   MAR-STAT = > M <
	//   SEX = > M <
	//   BIRTH = > 1943/05/04 <
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 3 ]
	//     ADDRESS-LINE[01] = > POSTFACH 67          <
	//     ADDRESS-LINE[02] = > MANDELA-WEG 8        <
	//     ADDRESS-LINE[03] = > 6000 FRANKFURT       <
	//    CITY = > FRANKFURT            <
	//    ZIP = > 6000       <
	//    COUNTRY = > D   <
	//   TELEPHONE = [ 1 ]
	//    AREA-CODE = > 069    <
	//    PHONE = > 549987          <
	//   DEPT = > FINA21 <
	//   JOB-TITLE = > SACHBEARBEITER            <
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
	//   LEAVE-DATA = [ 1 ]
	//    LEAVE-DUE = > 30 <
	//    LEAVE-TAKEN = > 0 <
	//   LEAVE-BOOKED = [ 0 ]
	//   LANG = [ 2 ]
	//    LANG[01] = > GER <
	//    LANG[02] = > ENG <
	//   LEAVE-LEFT = > 3000 <
	//   DEPARTMENT = > FINA <
	//   DEPT-PERSON = > FINA21SCHMIDT              <
	//   CURRENCY-SALARY = >  <

}

func ExampleConnection_periodGroupPart() {
	initLogWithFile("connection.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
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

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
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

func ExampleConnection_periodGroupLastEntry() {
	initLogWithFile("connection.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
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

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = request.QueryFields("PERSONNEL-ID,INCOME[N]")
	if err != nil {
		fmt.Println("Query fields error", err)
		return
	}
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
	//   INCOME = [ 1 ]
	//    CURR-CODE[ N] = > EUR <
	//    SALARY[ N] = > 20307 <
	//    BONUS[ N] = [ 1 ]
	//     BONUS[ N,01] = > 1282 <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   INCOME = [ 1 ]
	//    CURR-CODE[ N] = > EUR <
	//    SALARY[ N] = > 24102 <
	//    BONUS[ N] = [ 1 ]
	//     BONUS[ N,01] = > 1948 <

}

func ExampleConnection_multiplefieldIndex() {
	initLogWithFile("connection.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
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

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = request.QueryFields("PERSONNEL-ID,ADDRESS-LINE[2]")
	if err != nil {
		fmt.Println("Query fields error", err)
		return
	}
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
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 1 ]
	//     ADDRESS-LINE[02] = > WIESENGRUND 10       <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 1 ]
	//     ADDRESS-LINE[02] = > MANDELA-WEG 8        <

}

func XExampleConnectionSingleIndex() {
	initLogWithFile("connection.log")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
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

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = request.QueryFields("PERSONNEL-ID,BONUS[03,01]")
	if err != nil {
		fmt.Println("Query fields error", err)
		return
	}
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
	//   INCOME = [ 1 ]
	//    BONUS[03] = [ 1 ]
	//     BONUS[03,01] = > 1282 <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   INCOME = [ 0 ]

}
