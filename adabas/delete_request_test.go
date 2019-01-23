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

func testCallback(adabasRequest *adatypes.AdabasRequest, x interface{}) (err error) {
	deleteRequest := x.(*DeleteRequest)
	isn := adabasRequest.Isn
	fmt.Printf("Delete ISN: %d on %s/%d\n", adabasRequest.Isn, deleteRequest.repository.URL.String(), deleteRequest.repository.Fnr)
	err = deleteRequest.Delete(isn)
	return
}

func prepareCall(t *testing.T, mapName string) {
	adabas := NewAdabas(adabasModDBID)
	mr := NewMapRepository(adabas, 250)
	readRequest, rErr := NewMapNameRequestRepo(mapName, adabas, mr)
	if !assert.NoError(t, rErr) {
		return
	}
	defer readRequest.Close()
	readRequest.Limit = 0
	readRequest.QueryFields("")
	result, rerr := readRequest.ReadPhysicalSequence()
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Nr entries in database", result.NrRecords())

	storeRequest, err := NewAdabasMapNameStoreRequest(adabas, readRequest.adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()

	recErr := storeRequest.StoreFields("PERSONNEL-ID,FULL-NAME")
	if !assert.NoError(t, recErr) {
		return
	}

	for i := 0; i < 3-result.NrRecords(); i++ {
		fmt.Println("Add record", i)
		storeRecord, rErr := storeRequest.CreateRecord()
		if !assert.NoError(t, rErr) {
			return
		}
		if !assert.NotNil(t, storeRecord) {
			return
		}
		err = storeRecord.SetValue("PERSONNEL-ID", fmt.Sprintf("K%07d", i+1))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("NAME", fmt.Sprintf("NAME XXX %07d", i+1))
		if !assert.NoError(t, err) {
			return
		}
		storeRequest.Store(storeRecord)
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

}

func TestDeleteRequestByMapNameCommonRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "delete.log")
	defer f.Close()

	mapName := storeEmployeesMap
	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, storeEmployeesMap, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}

	prepareCall(t, mapName)

	adabas := NewAdabas(adabasModDBID)
	fmt.Println("Delete record with map name:", mapName)
	AddMapRepository(adabas, 250)
	defer DelMapRepository(adabas, 250)

	deleteRequest, err := NewMapNameDeleteRequest(adabas, mapName)
	if !assert.NoError(t, err) {
		fmt.Println("Delete Request error", err)
		return
	}
	fmt.Println("Check request in map", mapName, "and delete in", deleteRequest.adabas.String(), deleteRequest.repository.Fnr)
	if !assert.NotNil(t, deleteRequest) {
		fmt.Println("Delete Request nil", deleteRequest)
		return
	}
	defer deleteRequest.Close()
	fmt.Println("Query entries in map", mapName)
	adatypes.Central.Log.Debugf("New map request after clear map")
	readRequest, rErr := NewMapNameRequest(adabas, mapName)
	if !assert.NoError(t, rErr) {
		return
	}
	defer readRequest.Close()
	fmt.Println("Clear all entries in map", mapName)
	// Need to call all and don't need to read the data for deleting all records
	readRequest.Limit = 1
	readRequest.QueryFields("")
	fmt.Println("Read request in map", mapName, "and delete in", readRequest.adabas.String(), readRequest.repository.Fnr)
	result, rerr := readRequest.ReadPhysicalSequence()
	if !assert.NoError(t, rerr) {
		return
	}

	if !assert.Equal(t, 1, len(result.Values)) {
		return
	}
	fmt.Println("Values: ", len(result.Values), result.NrRecords())
	err = deleteRequest.Delete(result.Values[0].Isn)
	if !assert.NoError(t, rerr) {
		return
	}
	err = deleteRequest.EndTransaction()
	if !assert.NoError(t, rerr) {
		return
	}
}

func TestDeleteRequestByMapNameRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "delete.log")
	defer f.Close()

	mapName := storeEmployeesMap
	adabas := NewAdabas(adabasModDBID)
	fmt.Println("Delete record with map name:", mapName)
	mr := NewMapRepository(adabas, 250)

	deleteRequest, err := NewMapNameDeleteRequestRepo(mapName, adabas, mr)
	if !assert.NoError(t, err) {
		fmt.Println("Delete Request error", err)
		return
	}
	fmt.Println("Check request in map", mapName, "and delete in", deleteRequest.adabas.String(), deleteRequest.repository.Fnr)
	if !assert.NotNil(t, deleteRequest) {
		fmt.Println("Delete Request nil", deleteRequest)
		return
	}
	defer deleteRequest.Close()
	fmt.Println("Query entries in map", mapName)
	adatypes.Central.Log.Debugf("New map request after clear map")
	readRequest, rErr := NewMapNameRequestRepo(mapName, adabas, mr)
	if !assert.NoError(t, rErr) {
		return
	}
	defer readRequest.Close()
	fmt.Println("Clear all entries in map", mapName)
	// Need to call all and don't need to read the data for deleting all records
	readRequest.Limit = 1
	readRequest.QueryFields("")
	fmt.Println("Read request in map", mapName, "and delete in", readRequest.adabas.String(), readRequest.repository.Fnr)
	result, rerr := readRequest.ReadPhysicalSequence()
	if !assert.NoError(t, rerr) {
		return
	}

	if !assert.Equal(t, 1, len(result.Values)) {
		return
	}
	fmt.Println("Values: ", len(result.Values), result.NrRecords())
	err = deleteRequest.Delete(result.Values[0].Isn)
	if !assert.NoError(t, rerr) {
		return
	}
	err = deleteRequest.EndTransaction()
	if !assert.NoError(t, rerr) {
		return
	}
}

func clearFile(file uint32) error {
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if err != nil {
		return err
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(file)
	if err != nil {
		return rErr
	}
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(file)
	if dErr != nil {
		return dErr
	}
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	if err != nil {
		return err
	}
	err = deleteRequest.EndTransaction()
	if err != nil {
		return err
	}

	return nil
}

func TestDeleteRequestRefreshFile16(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "delete.log")
	defer f.Close()

	cErr := clearFile(16)
	assert.NoError(t, cErr)

}
