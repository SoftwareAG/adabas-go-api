/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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

package adatypes

import (
	"fmt"
)

func ExampleDefinition_ShouldRestrictToFields() {
	err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Init log error:", err)
		return
	}

	InitDefinitionCache()
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.PutCache("AA")
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpTypes(false, true)
	err = testDefinition.ShouldRestrictToFields("GC,I8")
	if err != nil {
		fmt.Println("Restrict original entry", err)
		return
	}
	definition := CreateDefinitionByCache("AA")
	if definition == nil {
		fmt.Println("Error create cache definition nil")
		return
	}
	err = definition.ShouldRestrictToFields("GC,I8")
	if err != nil {
		fmt.Println("Restrict cached entry error", err)
		return
	}
	definition.DumpTypes(false, false)
	definition.DumpTypes(false, true)

	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump all active field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump all active field types:
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//   1, I8, 8, B  ; I8
}
