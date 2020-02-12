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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const BlockSize = 4096

func TestConnectionStorePartial(t *testing.T) {
	initTestLogWithFile(t, "connection_lobstore.log")

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("AA,RA")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	err = record.SetValue("AA", "SLOB")
	assert.NoError(t, err)
	err = record.SetValue("RA", "ABCCCDADSDSDSDD")
	assert.NoError(t, err)
	//record.DumpValues()
	err = storeRequest.Store(record)
	assert.NoError(t, err)
	err = record.SetValue("AA", "PARTLOB")
	assert.NoError(t, err)
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + "img" + string(os.PathSeparator) + "106-0687_IMG.JPG"
	x, ferr := ioutil.ReadFile(name)
	if assert.NoError(t, ferr) {
		err = record.SetValue("RA", x)
		assert.NoError(t, err)
		//record.DumpValues()
		err = storeRequest.Store(record)
		assert.NoError(t, err)
	}
}

func TestConnectionStorePartialStream(t *testing.T) {
	initTestLogWithFile(t, "connection_lobstore.log")

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("AA")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	err = record.SetValue("AA", "STLOB")
	assert.NoError(t, err)
	err = storeRequest.Store(record)
	assert.NoError(t, err)
	isn := record.Isn
	storeRequest, serr = connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRequest.StoreFields("RA")
	if !assert.NoError(t, err) {
		return
	}
	record, rerr = storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return
	}
	record.Isn = isn
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + "img" + string(os.PathSeparator) + "106-0687_IMG.JPG"
	x, ferr := ioutil.ReadFile(name)
	if assert.NoError(t, ferr) {
		blockBegin := uint32(0)
		err = record.SetPartialValue("RA", blockBegin, x[blockBegin:blockBegin+BlockSize])
		assert.NoError(t, err)
		//record.DumpValues()
		err = storeRequest.Update(record)
		assert.NoError(t, err)
	}
}
