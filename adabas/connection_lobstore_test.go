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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

const BlockSize = adatypes.PartialStoreLobSizeChunks

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

func storePerStream(t *testing.T, x []byte) (adatypes.Isn, error) {
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return 0, err
	}
	defer connection.Close()

	fmt.Println("Store LOB using partial lob")
	storeRequest, serr := connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return 0, serr
	}
	err = storeRequest.StoreFields("AA")
	if !assert.NoError(t, err) {
		return 0, err
	}
	record, rerr := storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return 0, rerr
	}
	err = record.SetValue("AA", "STLOB")
	assert.NoError(t, err)
	err = storeRequest.Store(record)
	assert.NoError(t, err)
	isn := record.Isn
	storeRequest, serr = connection.CreateStoreRequest(17)
	if !assert.NoError(t, serr) {
		return 0, serr
	}
	err = storeRequest.StoreFields("RA")
	if !assert.NoError(t, err) {
		return 0, err
	}
	record, rerr = storeRequest.CreateRecord()
	if !assert.NoError(t, rerr) {
		return 0, rerr
	}
	record.Isn = isn

	blockBegin := uint32(0)
	for i := blockBegin; i < uint32(len(x)); i += BlockSize {
		e := i + BlockSize
		if e > uint32(len(x)) {
			e = uint32(len(x))
		}
		fmt.Println("Write block", i, e, len(x))
		err = record.SetPartialValue("RA", i+1, x[i:e])
		if !assert.NoError(t, err) {
			return 0, err
		}
		err = storeRequest.Update(record)
		if !assert.NoError(t, err) {
			return 0, err
		}

	}
	connection.EndTransaction()
	return isn, nil
}

func verifyStorePerStream(t *testing.T, isn adatypes.Isn, x []byte) error {
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()

	fmt.Println("Verify stored LOB")

	request, rErr := connection.CreateFileReadRequest(17)
	if !assert.NoError(t, rErr) {
		return rErr
	}
	err = request.QueryFields("AA,RA")
	if !assert.NoError(t, err) {
		return err
	}
	result, verr := request.ReadISN(isn)
	if !assert.NoError(t, verr) {
		return err
	}
	if !assert.Equal(t, 1, len(result.Values)) {
		return nil
	}
	aaValue, _ := result.Values[0].SearchValue("AA")
	assert.Equal(t, "STLOB   ", aaValue.String())
	assert.Equal(t, "STLOB   ", result.Values[0].AlphaValue("AA"))
	raValue, rerr := result.Values[0].SearchValue("RA")
	if !assert.NoError(t, rerr) {
		return rerr
	}
	raw := raValue.Bytes()
	assert.Equal(t, 1386643, len(x))
	assert.Equal(t, 1386643, len(raw))
	if !assert.Equal(t, x, raw) {
		return fmt.Errorf("Data read not equal")
	}
	return nil
}

func verifyReadWithStream(t *testing.T, isn adatypes.Isn, x []byte) error {
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()
	fmt.Println("Verify stored LOB reading partial")

	request, rErr := connection.CreateFileReadRequest(17)
	if !assert.NoError(t, rErr) {
		return rErr
	}
	err = request.QueryFields("RA(1,1000)")
	if !assert.NoError(t, err) {
		return err
	}
	result, verr := request.ReadISN(isn)
	if !assert.NoError(t, verr) {
		return err
	}
	if !assert.Equal(t, 1, len(result.Values)) {
		return nil
	}
	raValue, rerr := result.Values[0].SearchValue("RA")
	if !assert.NoError(t, rerr) {
		return rerr
	}
	raw := raValue.Bytes()
	assert.Equal(t, 1000, len(raw))
	assert.Equal(t, x[0:1000], raw)
	return nil
}

func TestConnectionStorePartialStream(t *testing.T) {
	initTestLogWithFile(t, "connection_lobstore.log")
	cErr := clearFile(17)
	if !assert.NoError(t, cErr) {
		return
	}

	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + "img" + string(os.PathSeparator) + "106-0687_IMG.JPG"
	x, ferr := ioutil.ReadFile(name)
	if !assert.NoError(t, ferr) {
		return
	}
	isn, err := storePerStream(t, x)
	if err != nil {
		return
	}
	err = verifyStorePerStream(t, isn, x)
	if err != nil {
		return
	}
	verifyReadWithStream(t, isn, x)
}

func TestReadLogicalWithCursoring_LOB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_cursoring.log")

	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateFileReadRequest(17)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	fmt.Println("Init stream ...")
	col, cerr := request.ReadLobStream("", "RA")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error reading stream with using cursoring", cerr)
		return
	}
	fmt.Println("Read next cursor record...")
	counter := 0
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if record == nil {
			fmt.Println("Record nil received")
			return
		}
		counter++
		if !assert.NoError(t, rerr) {
			fmt.Println("Error reading logical with using cursoring", rerr)
			return
		}
		adatypes.Central.Log.Debugf("Read next cursor record...%d", counter)
	}
	assert.Equal(t, 1107, counter)
	fmt.Println("Last cursor record read")

}