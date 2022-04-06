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

package adabas

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func updateStream(record *Record, x interface{}) error {
	tc := x.(*testCopy)
	// fmt.Printf("Update %d -> %d\n", record.Isn, record.Quantity)
	err := record.SetValue(tc.indexField, fmt.Sprintf("%05d", tc.i))
	if err != nil {
		return err
	}
	err = tc.store.Update(record)
	if err != nil {
		return err
	}
	tc.i++
	return err
}

func TestConnectionStoreCopyUpdate(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	clearAdabasFile(t, adabasModDBIDs, 16)
	err := copyAdabasFile(t, "*", adabasModDBIDs, 11, 16)
	if !assert.NoError(t, err) {
		return
	}

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateFileReadRequest(16)
	if !assert.NoError(t, rerr) {
		return
	}
	err = readRequest.QueryFields("AA")
	if !assert.NoError(t, err) {
		return
	}

	storeRequest, serr := connection.CreateStoreRequest(16)
	if !assert.NoError(t, serr) {
		return
	}
	fmt.Println("Read physical read...", adabasModDBIDs, 16)
	tc := &testCopy{store: storeRequest, indexField: "AA"}
	_, err = readRequest.ReadPhysicalSequenceStream(updateStream, tc)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestUpdate", adabasModDBIDs, 16)
}

func copyAdabasMap(t *testing.T, fields, target string, file Fnr, orig, dest string) error {
	connection, err := NewConnection(fmt.Sprintf("acj;map;config=[%s,%d]", target, file))
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateMapReadRequest(orig)
	if !assert.NoError(t, rerr) {
		return rerr
	}
	err = readRequest.QueryFields(fields)
	if !assert.NoError(t, err) {
		return err
	}

	storeRequest, serr := connection.CreateMapStoreRequest(dest)
	if !assert.NoError(t, serr) {
		return serr
	}
	fmt.Println("Read physical read...", target, dest)
	tc := &testCopy{store: storeRequest, indexField: "AA"}
	_, err = readRequest.ReadPhysicalSequenceStream(copyStream, tc)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return err
	}
	return nil
}

const (
	origMap = "EmployeesOrig"
	destMap = "EmployeesDest"
)

func prepareMaps(t *testing.T, mapName string) error {
	maps, err := LoadJSONMap(mapName)
	if !assert.NoError(t, err) {
		return err
	}
	if !assert.Len(t, maps, 1) {
		return fmt.Errorf("Array error")
	}
	u, _ := NewURL(adabasModDBIDs)
	maps[0].Repository = &DatabaseURL{URL: *u, Fnr: 249}
	maps[0].Name = origMap
	err = maps[0].Store()
	if !assert.NoError(t, err) {
		return err
	}
	maps[0].Name = destMap
	maps[0].Data.Fnr = 16
	err = maps[0].Store()
	if !assert.NoError(t, err) {
		return err
	}
	return nil
}

func TestConnectionMapStoreCopyUpdate(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	clearAdabasFile(t, adabasModDBIDs, 16)
	clearAdabasFile(t, adabasModDBIDs, 249)

	err := prepareMaps(t, "EMPLOYEES-NAT-DDM.json")
	if err != nil {
		return
	}
	err = copyAdabasMap(t, "*", adabasModDBIDs, 249, origMap, destMap)
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestMapCopy", adabasModDBIDs, 16)

	connection, err := NewConnection(fmt.Sprintf("acj;map;config=[%s,%d]", adabasModDBIDs, 249))
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateMapReadRequest(destMap)
	if !assert.NoError(t, rerr) {
		return
	}
	err = readRequest.QueryFields("PERSONNEL-ID")
	if !assert.NoError(t, err) {
		return
	}

	storeRequest, serr := connection.CreateMapStoreRequest(destMap)
	if !assert.NoError(t, serr) {
		return
	}
	fmt.Println("Read physical read...", destMap)
	tc := &testCopy{store: storeRequest, indexField: "PERSONNEL-ID"}
	_, err = readRequest.ReadPhysicalSequenceStream(updateStream, tc)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestMapUpdate", adabasModDBIDs, 16)
}

type EmployeesUpdate struct {
	Index     uint64 `adabas:"#ISN" json:"-"`
	ID        string
	Birth     int64
	Name      string `adabas:"Name"`
	FirstName string `adabas:"FirstName"`
}

func TestConnectionInterfaceStoreCopyUpdate(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	clearAdabasFile(t, adabasModDBIDs, 16)
	clearAdabasFile(t, adabasModDBIDs, 249)

	err := prepareMaps(t, "EMPLOYEES-NAT-DDM.json")
	if err != nil {
		return
	}
	maps, err := LoadJSONMap("Employees.json")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, maps, 1) {
		return
	}
	u, _ := NewURL(adabasModDBIDs)
	maps[0].Repository = &DatabaseURL{URL: *u, Fnr: 249}
	maps[0].Name = "EmployeesUpdate"
	maps[0].Data.Fnr = 16
	err = maps[0].Store()
	if !assert.NoError(t, err) {
		return
	}

	err = copyAdabasMap(t, "*", adabasModDBIDs, 249, origMap, destMap)
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestInterfaceCopy", adabasModDBIDs, 16)

	connection, err := NewConnection(fmt.Sprintf("acj;map;config=[%s,%d]", adabasModDBIDs, 249))
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateMapReadRequest(&EmployeesUpdate{})
	if !assert.NoError(t, rerr) {
		return
	}
	err = readRequest.QueryFields("ID")
	if !assert.NoError(t, err) {
		return
	}

	storeRequest, serr := connection.CreateMapStoreRequest(&EmployeesUpdate{})
	if !assert.NoError(t, serr) {
		return
	}
	fmt.Println("Read physical read...", destMap)
	tc := &testCopy{store: storeRequest, indexField: "ID"}
	_, err = readRequest.ReadPhysicalSequenceStream(updateStream, tc)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestInterfaceUpdate", adabasModDBIDs, 16)
}
