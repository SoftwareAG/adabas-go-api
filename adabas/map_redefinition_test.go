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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestReadPartRedefinition(t *testing.T) {
	initTestLogWithFile(t, "redefinition.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("MapperRedefTest")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	fmt.Println("Map loaded", request.adabasMap.String())
	err = request.QueryFields("Personel-Id,LONGSTRING")
	if !assert.NoError(t, err) {
		fmt.Println("Error query request", err)
		return
	}
	request.Limit = 4
	request.definition.DumpTypes(false, false)
	request.definition.DumpTypes(false, true)
	result, qerr := request.ReadLogicalWith("Personel-Id=[REDEF0:REDEFA]")
	if !assert.NoError(t, qerr) {
		fmt.Println("Error read sequence", qerr)
		return
	}
	result.DumpValues()
	if assert.Equal(t, 4, len(result.Values)) {
		record := result.Values[0]
		f := record.HashFields["LONGSTRING"]
		if assert.NotNil(t, f) {
			assert.Equal(t, "", f.String())
		}
	}
}

func TestRedefinition(t *testing.T) {
	initTestLogWithFile(t, "redefinition.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("MapperRedefTest")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	fmt.Println("Map loaded", request.adabasMap.String())
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		fmt.Println("Error query request", err)
		return
	}
	request.Limit = 4
	request.definition.DumpTypes(false, false)
	request.definition.DumpTypes(false, true)
	result, qerr := request.ReadLogicalWith("Personel-Id=[REDEF0:REDEFA]")
	if !assert.NoError(t, qerr) {
		fmt.Println("Error read sequence", qerr)
		return
	}
	result.DumpValues()
	if assert.Equal(t, 4, len(result.Values)) {
		record := result.Values[0]
		f := record.HashFields["LONGSTRING"]
		if assert.NotNil(t, f) {
			assert.Equal(t, "", f.String())
		}
	}
}
