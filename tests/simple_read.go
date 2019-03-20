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

package main

import (
	"fmt"

	"github.com/SoftwareAG/adabas-go-api/adabas"
)

func main() {

	connection, err := adabas.NewConnection("acj;target=12")
	if err != nil {
		fmt.Printf("NewConnection() error=%v\n", err)
		return
	}
	defer connection.Close()
	err = connection.Open()
	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}
	readRequest, cerr := connection.CreateFileReadRequest(11)
	if cerr != nil {
		fmt.Printf("CreateFileReadRequest() error=%v\n", err)
		return
	}
	err = readRequest.QueryFields("AA,AB")
	if err != nil {
		fmt.Printf("QueryFields() error=%v\n", err)
		return
	}
	readRequest.Limit = 0
	result, err := readRequest.ReadLogicalWith("AA=60010001")
	if err != nil {
		fmt.Printf("ReadLogicalWith() error=%v\n", err)
		return
	}
	fmt.Printf("ReadLogicalWith() result=%v\n", result)
	var aa, ac, ad, ae string
	// Read given AA(alpha) and all entries of group AB to string variables
	result.Scan(&aa, &ac, &ad, &ae)

}
