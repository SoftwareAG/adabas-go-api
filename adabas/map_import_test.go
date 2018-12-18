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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapImportPrepare(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	deleteRequest := NewDeleteRequest("23", 250)
	request := NewRequest("23", 250)
	request.Limit = 0
	defer request.Close()
	err := request.QueryFields("")
	if !assert.NoError(t, err) {
		return
	}
	err = request.ReadPhysicalSequenceWithParser(testCallback, deleteRequest)
	if !assert.NoError(t, err) {
		return
	}
	err = deleteRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

}

func TestMapImport(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()
	adabas := NewAdabas(23)
	mr := NewMapRepository(adabas, 250)
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + "/" + "EmployeeX.systrans"

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 11}
	maps, err := mr.ImportMapRepository(adabas, "*", name, dataRepository)
	if !assert.NoError(t, err) {
		fmt.Println(err)
	}
	fmt.Println("Number of maps", len(maps))
	for _, m := range maps {
		fmt.Println("MAP", m.Name)
		err = m.Store()
		if !assert.NoError(t, err) {
			return
		}
	}

}

func TestMapImportMassLoad(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	adabas := NewAdabas(23)
	mr := NewMapRepository(adabas, 250)
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + "/" + "Empl-MassLoad.systrans"

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 11}
	maps, err := mr.ImportMapRepository(adabas, "*", name, dataRepository)
	if !assert.NoError(t, err) {
		fmt.Println(err)
	}
	fmt.Println("Number of maps", len(maps))
	for _, m := range maps {
		m.Name = "EMPLDDM-MASSLOAD"
		fmt.Println("MAP", m.Name)
		err = m.Store()
		if !assert.NoError(t, err) {
			return
		}
	}

}
