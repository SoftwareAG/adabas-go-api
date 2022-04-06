/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	ferr := initLogWithFile("messages.log")
	if ferr != nil {
		fmt.Println(ferr)
		return
	}

	err := NewGenericError(02, "XX")
	assert.Equal(t, "ADG0000002: Invalid Adabas command send: XX", err.Error())
	assert.Equal(t, "ADG0000002: Invalid Adabas command send: XX", err.Error())
	assert.Equal(t, "Fehlerhaftes Adabas Kommando gesendet: XX", err.Translate("de"))

}

func ExampleNewGenericError_print() {
	ferr := initLogWithFile("messages.log")
	if ferr != nil {
		fmt.Println(ferr)
		return
	}

	err := NewGenericError(02, "XX")
	fmt.Println(err)
	fmt.Println("Code", err.Code)
	fmt.Println("Message", err.Message)

	err = NewGenericError(05)
	fmt.Println(err)
	fmt.Println("Code", err.Code)
	fmt.Println("Message", err.Message)
	// Output: ADG0000002: Invalid Adabas command send: XX
	// Code ADG0000002
	// Message Invalid Adabas command send: XX
	// ADG0000005: Repository not defined
	// Code ADG0000005
	// Message Repository not defined
}
