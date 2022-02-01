/*
* Copyright © 2021-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"crypto/md5"
	"fmt"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type lobTest struct {
	Index uint64 `adabas:":isn"`
	Name  string `adabas:"::BB"`
}

func TestStreamStoreNoET(t *testing.T) {
	initTestLogWithFile(t, "stream_store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	//	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	connection, err := NewConnection("acj;target=adatcp://localhost:600" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(202)
	if !assert.NoError(t, serr) {
		return
	}
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	p = p + "/../files/img/106-0687_IMG.JPG"
	f, err := os.Open(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if !assert.NoError(t, err) {
		return
	}
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Printf("Number of bytes read: %d/%d\n", n, len(data))
	md5sum := fmt.Sprintf("%X", md5.Sum(data))
	assert.Equal(t, "343EEB0AB6E46058428490679A15A02B", md5sum)

	lobEntry := &lobTest{Name: "1234Test.JPG"}
	err = storeRequest.StoreData(lobEntry)
	assert.NoError(t, err)
	fmt.Println("ISN:", lobEntry.Index)
	err = storeRequest.EndTransaction()
	assert.NoError(t, err)

	from := uint64(0)
	blocksize := uint64(64000)
	counter := 0
	for {
		// fmt.Println(from, from+blocksize, "Store", len(data))
		err = storeRequest.UpdateLOBRecord(adatypes.Isn(lobEntry.Index), "DC", from, data[from:int(from+blocksize)])
		if !assert.NoError(t, err, fmt.Sprintf("Error on offset %d", from)) {
			return
		}
		from += blocksize
		if int(from) >= len(data) {
			break
		}
		if len(data) < int(from+blocksize) {
			blocksize = uint64(len(data)) % blocksize
		}
		counter++
	}
	assert.Equal(t, 21, counter)
	request, rerr := connection.CreateFileReadRequest(202)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	// Read all data at once as reference
	rerr = request.QueryFields("DC")
	if !assert.NoError(t, rerr) {
		return
	}
	record, err := request.ReadISN(adatypes.Isn(lobEntry.Index))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, record.NrRecords())
	refValue, err := record.Values[0].SearchValue("DC")
	assert.NoError(t, err)
	refData := refValue.Bytes()
	assert.Equal(t, 1386643, len(refData))
	md5sum = fmt.Sprintf("%X", md5.Sum(refData))
	assert.Equal(t, "343EEB0AB6E46058428490679A15A02B", md5sum)

}

func TestStreamStoreWithET(t *testing.T) {
	initTestLogWithFile(t, "stream_store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	//	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	connection, err := NewConnection("acj;target=adatcp://localhost:600" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(202)
	if !assert.NoError(t, serr) {
		return
	}
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	p = p + "/../files/img/106-0687_IMG.JPG"
	f, err := os.Open(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if !assert.NoError(t, err) {
		return
	}
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Printf("Number of bytes read: %d/%d\n", n, len(data))
	md5sum := fmt.Sprintf("%X", md5.Sum(data))
	assert.Equal(t, "343EEB0AB6E46058428490679A15A02B", md5sum)

	lobEntry := &lobTest{Name: "1234Test.JPG"}
	err = storeRequest.StoreData(lobEntry)
	assert.NoError(t, err)
	fmt.Println("ISN:", lobEntry.Index)
	err = storeRequest.EndTransaction()
	assert.NoError(t, err)

	from := uint64(0)
	blocksize := uint64(64000)
	counter := 0
	for {
		// fmt.Println(from, from+blocksize, "Store", len(data))
		err = storeRequest.UpdateLOBRecord(adatypes.Isn(lobEntry.Index), "DC", from, data[from:int(from+blocksize)])
		if !assert.NoError(t, err, fmt.Sprintf("Error on offset %d", from)) {
			return
		}
		from += blocksize
		if int(from) >= len(data) {
			break
		}
		if len(data) < int(from+blocksize) {
			blocksize = uint64(len(data)) % blocksize
		}
		counter++
		if counter%3 == 0 {
			err = storeRequest.EndTransaction()
			if !assert.NoError(t, err) {
				fmt.Println("Error end the transaction", err)
				return
			}
		}
	}
	assert.Equal(t, 21, counter)

	request, rerr := connection.CreateFileReadRequest(202)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error creating map read request", rerr)
		return
	}
	// Read all data at once as reference
	rerr = request.QueryFields("DC")
	if !assert.NoError(t, rerr) {
		return
	}
	record, err := request.ReadISN(adatypes.Isn(lobEntry.Index))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, record.NrRecords())
	refValue, err := record.Values[0].SearchValue("DC")
	assert.NoError(t, err)
	refData := refValue.Bytes()
	assert.Equal(t, 1386643, len(refData))
	md5sum = fmt.Sprintf("%X", md5.Sum(refData))
	assert.Equal(t, "343EEB0AB6E46058428490679A15A02B", md5sum)

}
