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

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStorePEMU(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(16)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	err = record.SetValue("AA", "PEMU")
	assert.NoError(t, err)
	err = record.SetValue("AE", "XNAME")
	assert.NoError(t, err)
	err = record.SetValue("AS[1]", 123)
	assert.NoError(t, err)
	err = record.SetValue("AR[1]", "223")
	assert.NoError(t, err)
	err = record.SetValue("AT[1,1]", 323)
	assert.NoError(t, err)
	err = record.SetValue("AT[1][2]", 456)
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("Set AZ[1]")
	err = record.SetValue("AZ[1]", "123")
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("Set AZ[2]")
	err = record.SetValue("AZ[2]", "999")
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("AZ set")
	//record.DumpValues()
	err = storeRequest.Store(record)
	assert.NoError(t, err)
}

type testCopy struct {
	i          uint32
	indexField string
	store      *StoreRequest
}

func copyStream(record *Record, x interface{}) error {
	tc := x.(*testCopy)
	fmt.Printf("Store %d -> %d\n", record.Isn, record.Quantity)
	record.DumpValues()
	err := tc.store.Store(record)
	tc.i++
	return err
}

func checkContent(t *testing.T, name, target string, file Fnr) error {
	connection, err := NewConnection("acj;target=" + target)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()
	readRequest, rerr := connection.CreateFileReadRequest(file)
	if !assert.NoError(t, rerr) {
		return rerr
	}
	err = readRequest.QueryFields("*")
	if !assert.NoError(t, err) {
		return err
	}

	result, rErr := readRequest.ReadPhysicalSequence()
	if !assert.NoError(t, rErr) {
		return rErr
	}
	return validateResult(t, name, result)
}

func copyAdabasFile(t *testing.T, fields, target string, org, dest Fnr) error {
	connection, err := NewConnection("acj;target=" + target)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateFileReadRequest(org)
	if !assert.NoError(t, rerr) {
		return rerr
	}
	err = readRequest.QueryFields(fields)
	if !assert.NoError(t, err) {
		return err
	}

	storeRequest, serr := connection.CreateStoreRequest(dest)
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

func TestConnectionStoreCopy(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	clearAdabasFile(t, adabasModDBIDs, 16)
	err := copyAdabasFile(t, "*", adabasModDBIDs, 11, 16)
	if !assert.NoError(t, err) {
		return
	}

	checkContent(t, "checkTestCopy", adabasModDBIDs, 16)
}

func TestConnectionStoreRestrictedCopy(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	clearAdabasFile(t, adabasModDBIDs, 16)

	err := copyAdabasFile(t, "AA,AB,AQ", adabasModDBIDs, 11, 16)
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "checkTestRestrictedCopy", adabasModDBIDs, 16)
}

func TestConnectionStoreSQL(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	err = record.SetValue("AA", "PEMU")
	assert.NoError(t, err)
	err = record.SetValue("AE", "XNAME")
	assert.NoError(t, err)
	err = record.SetValue("LB[1]", 123)
	assert.NoError(t, err)
	err = record.SetValue("LB[1]", "223")
	assert.NoError(t, err)
	err = record.SetValue("LC[1,1]", 323)
	assert.NoError(t, err)
	err = record.SetValue("LC[1][2]", 456)
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("Set SA[1]")
	err = record.SetValue("SA[1]", "123")
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("Set SA[2]")
	err = record.SetValue("SA[2]", "999")
	assert.NoError(t, err)
	adatypes.Central.Log.Debugf("AZ set")
	//record.DumpValues()
	err = storeRequest.Store(record)
	assert.NoError(t, err)
}

func TestConnectionStoreAnalytics(t *testing.T) {
	initTestLogWithFile(t, "connection_store.log")

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(23)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	err = record.SetValue("AA", "PEMU")
	assert.NoError(t, err)
	record.DumpValues()
	err = storeRequest.Store(record)
	assert.NoError(t, err)
}
