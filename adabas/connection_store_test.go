/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
	"testing"
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
