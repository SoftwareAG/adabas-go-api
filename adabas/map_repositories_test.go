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

package adabas

import (
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestMapRepositoryReadAll(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	defer adabas.Close()
	mr := NewMapRepository(adabas, 4)
	adabasMaps, err := mr.LoadAllMaps(adabas)
	assert.NoError(t, err)
	assert.NotNil(t, adabasMaps)
	assert.NotEqual(t, 0, len(adabasMaps))
	for _, m := range adabasMaps {
		fmt.Println(m.Name)
	}
}

func TestMapRepositoryRead(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	defer adabas.Close()
	mr := NewMapRepository(adabas, 4)
	employeeMap, serr := mr.SearchMap(adabas, "EMPLOYEES-NAT-DDM")
	assert.NotNil(t, employeeMap)
	assert.NoError(t, serr)
	// fmt.Println(">", employeeMap.String())
	// adabasMaps, err := mr.LoadAllMaps(adabas)
	// assert.NoError(t, err)
	// assert.NotNil(t, adabasMaps)
	// assert.NotEqual(t, 0, len(adabasMaps))
	// for _, m := range adabasMaps {
	// 	if m.Name == "EMPLOYEES-NAT-DDM" {
	// 		employeeMap = m
	// 	}
	// }
	// fmt.Println(">", employeeMap.String())
	x := employeeMap.fieldMap["AA"]
	assert.NotNil(t, x)
	// fmt.Printf("%#v", x)
}
