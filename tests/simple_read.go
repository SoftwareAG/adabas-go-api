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
	connection.Open()
	readRequest, err := connection.CreateReadRequest(11)
	if err != nil {
		fmt.Printf("CreateReadRequest() error=%v\n", err)
		return
	}
	readRequest.QueryFields("AA,AB")
	readRequest.Limit = 0
	result, err := readRequest.ReadLogicalWith("AA=60010001")
	if err != nil {
		fmt.Printf("ReadLogicalWith() error=%v\n", err)
		return
	}
	fmt.Printf("ReadLogicalWith() result=%v\n", result)

}
