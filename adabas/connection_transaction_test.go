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
	"strconv"
	"testing"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

var vehicleMapName = mapVehicles + "Go"

const multipleTransactionRefName = "T16555"
const multipleTransactionRefName2 = "T19555"

func prepareStoreAndHold(t *testing.T, c chan bool) {
	adatypes.Central.Log.Infof("Create connection...")
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	defer connection.Close()

	storeRequest16, err := connection.CreateMapStoreRequest(massLoadSystransStore)
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	recErr := storeRequest16.StoreFields("PERSONNEL-ID,FULL-NAME")
	if !assert.NoError(t, recErr) {
		c <- false
		return
	}
	err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_0")
	if err != nil {
		c <- false
		return
	}
	storeRequest19, cErr := connection.CreateMapStoreRequest(vehicleMapName)
	if !assert.NoError(t, cErr) {
		c <- false
		return
	}
	recErr = storeRequest19.StoreFields("REG-NUM,CAR-DETAILS")
	if !assert.NoError(t, recErr) {
		c <- false
		return
	}
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_0")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	for i := 1; i < 10; i++ {
		x := strconv.Itoa(i)
		err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_"+x)
		if !assert.NoError(t, err) {
			c <- false
			return
		}

	}
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_1")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	fmt.Println("Records set in hold")
	c <- true
	time.Sleep(10 * time.Second)
	fmt.Println("End transaction")
	err = connection.EndTransaction()
	assert.NoError(t, err)
	fmt.Println("Notify main function")
	c <- true

}

func TestConnectionTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_transaction.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}

	adatypes.Central.Log.Infof("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	dataRepository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 19}
	perr = prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}

	c := make(chan bool)
	go prepareStoreAndHold(t, c)
	x := <-c
	if !x && t.Failed() {
		return
	}
	fmt.Println("Read records set in hold")

	connection, err := NewConnection("acj;map=" + massLoadSystransStore + ";config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateReadRequest()
	if !assert.NoError(t, rerr) {
		return
	}
	//request.SetHoldRecords(adatypes.HoldWait)
	request.QueryFields("AA")
	_, rrerr := request.ReadPhysicalSequence()
	if !assert.NoError(t, rrerr) {
		return
	}
	x = <-c

	fmt.Println("Check stored data", x)
	adatypes.Central.Log.Infof("Check stored data")
	checkStoreByFile(t, adabasModDBIDs, 16, multipleTransactionRefName)
	checkStoreByFile(t, adabasModDBIDs, 19, multipleTransactionRefName2)

}
