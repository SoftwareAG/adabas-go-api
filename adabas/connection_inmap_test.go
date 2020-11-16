/*
* Copyright © 2020 Software AG, Darmstadt, Germany and/or its licensors
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
	"runtime"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type IncomeInMap struct {
	Salary   uint64   `adabas:"::AS"`
	Bonus    []uint64 `adabas:"::AT"`
	Currency string   `adabas:"::AR"`
	Summary  uint64   `adabas:":ignore"`
}

type EmployeesInMap struct {
	Index      uint64         `adabas:":isn"`
	ID         string         `adabas:":key:AA"`
	FullName   *FullNameInMap `adabas:"::AB"`
	Birth      uint64         `adabas:"::AH"`
	Department string         `adabas:"::AO"`
	Income     []*IncomeInMap `adabas:"::AQ"`
	Language   []string       `adabas:"::AZ"`
}

type FullNameInMap struct {
	FirstName  string `adabas:"::AC"`
	MiddleName string `adabas:"::AD"`
	Name       string `adabas:"::AE"`
}

func TestInlineMap(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23,11")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.ReadISN(1024)
	if !assert.NoError(t, rerr) {
		return
	}
	response.DumpData()
	response.DumpValues()
}

func TestInlineStoreMap(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	clearAdabasFile(t, adabasModDBIDs, 16)

	connection, cerr := NewConnection("acj;inmap=23,16")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapStoreRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	e := &EmployeesInMap{FullName: &FullNameInMap{FirstName: "Anton", Name: "Skeleton", MiddleName: "Otto"}, Birth: 1234}
	rerr := request.StoreData(e)
	if !assert.NoError(t, rerr) {
		return
	}
	err = request.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "inmapstore", "23", 16)
}
