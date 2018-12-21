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
		fmt.Println("Connection to database error:",cerr)
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
	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0251
	//   AA = > 11100301 <
	//   AB = [ 1 ]
	//    AC = > HANS                 <
	//    AE = > BERGMANN             <
	//    AD = > WILHELM              <
	// Record Isn: 0383
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <

}
