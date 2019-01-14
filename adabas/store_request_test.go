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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

const (
	massLoadEmployees     = "EMPLDDM-MASSLOAD"
	massLoadSystrans      = "Empl-MassLoad.systrans"
	massLoadSystransStore = "EMPLDDM-MASSLOAD-STORE"
	mapVehicles           = "VEHICLES"
	vehicleSystransStore  = "Vehi.systrans"
	lengthPicture         = 1386643
)

func TestStoreAdabasFields(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}

	storeRequest := NewStoreRequest("23", 16)
	defer storeRequest.Close()
	err := storeRequest.StoreFields("AA,AC,AD,AE")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr := storeRequest.CreateRecord()
	if !assert.NoError(t, serr) {
		return
	}
	storeRecord.DumpValues()
	err = storeRecord.SetValue("AA", "123")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AC", "MANUEL")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AD", "TIMM")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AE", "KNOPFS")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord.DumpValues()
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()
}

func prepareCreateTestMap(t *testing.T, mapName string, fileName string, dataRepository *DatabaseURL) error {
	// fmt.Println("Check existing of map", mapName)
	adabas := NewAdabas(23)
	mr := NewMapRepository(adabas, 250)
	sm, err := mr.SearchMap(adabas, mapName)
	if err == nil {
		assert.NotNil(t, sm)
		//		fmt.Println(mapName, "Map found in repository", sm.Name)
		return nil
	}
	//	fmt.Println("Map", mapName, "not found in repository:", err)
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + "/" + fileName
	m, merr := mr.ImportMapRepository(adabas, "*", name, dataRepository)
	if !assert.NoError(t, merr) {
		return merr
	}
	//	fmt.Printf("Successfull importing map: %s\n", mapName)
	if assert.Equal(t, 1, len(m)) {
		m[0].Name = mapName
		err = m[0].Store()
		//		fmt.Printf("Storing map returned: %v\n", err)
		if assert.NoError(t, err) {
			//			fmt.Println("Map imported in repository")
		} else {
			return err
		}
	}
	return nil
}

func TestStoreFailMapFieldsCheck(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	fmt.Println("Start : TestStoreFailMapFieldsCheck")

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 11}
	perr := prepareCreateTestMap(t, massLoadEmployees, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(23)
	AddMapRepository(ada, 250)
	defer DelMapRepository(ada, 250)
	adabasMap, serr := SearchMapRepository(ada, massLoadEmployees)
	if !assert.NoError(t, serr) {
		return
	}
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()
	recErr := storeRequest.StoreFields("NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	if !assert.NotNil(t, storeRecord) {
		return
	}
	// storeRecord.DumpValues()
	err = storeRecord.SetValue("PERSONNEL-ID", "123")
	assert.Error(t, err)
	fmt.Println("Got expected error", err)
}

func TestStoreMapFields(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	fmt.Println("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}

	ada := NewAdabas(23)
	AddMapRepository(ada, 250)
	defer DelMapRepository(ada, 250)

	fmt.Println("Prepare clear test map")
	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	fmt.Println("Store map request")
	adabasMap, serr := SearchMapRepository(ada, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()
	recErr := storeRequest.StoreFields("PERSONNEL-ID,FULL-NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	storeRecord.DumpValues()
	err = storeRecord.SetValue("PERSONNEL-ID", "55555")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("FIRST-NAME", "THORSTEN")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("MIDDLE-I", "TIMM")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("NAME", "STORAGE_MAP")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord.DumpValues()
	fmt.Println("Store  request")
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("End transaction")
	storeRequest.EndTransaction()
}

func clearAdabasFile(t *testing.T, target string, fnr uint32) error {
	fmt.Println("Clear Adabas file", target, "/", fnr)
	adabas, err := NewAdabass(target)
	if err != nil {
		return err
	}
	deleteRequest := NewDeleteRequestAdabas(adabas, fnr)
	defer deleteRequest.Close()
	readRequest := NewRequestAdabas(adabas, fnr)
	defer readRequest.Close()
	// Need to call all and don't need to read the data for deleting all records
	readRequest.Limit = 0
	readRequest.QueryFields("")
	err = readRequest.ReadPhysicalSequenceWithParser(testCallback, deleteRequest)
	if err != nil {
		return err
	}
	err = deleteRequest.EndTransaction()
	fmt.Println("Cleared file")
	return err
}

func clearMap(t *testing.T, adabas *Adabas, mapName string) error {
	fmt.Println("Clear map", mapName)
	adabasMap, err := SearchMapRepository(adabas, mapName)
	if !assert.NoError(t, err) {
		return err
	}
	if !assert.Equal(t, adabas.Acbx.Acbxdbid, adabasMap.URL().Dbid) {
		return fmt.Errorf("Error dbid mismatch")
	}
	if !assert.Equal(t, adabas.URL.String(), adabasMap.URL().String()) {
		return fmt.Errorf("Error URL mismatch")
	}
	fmt.Println("Map found", adabasMap.Name, adabasMap.Repository.URL.String(), adabasMap.Repository.Fnr,
		adabasMap.Data.URL.String(), adabasMap.Data.Fnr)
	deleteRequest, err := NewMapDeleteRequest(adabas, adabasMap)
	if !assert.NoError(t, err) {
		fmt.Println("Delete Request error", err)
		return err
	}
	fmt.Println("Check request in map", mapName, "and delete in", deleteRequest.adabas.String(), deleteRequest.repository.Fnr)
	if !assert.NotNil(t, deleteRequest) {
		fmt.Println("Delete Request nil", deleteRequest)
		return fmt.Errorf("Delete request nil in clearMap")
	}
	defer deleteRequest.Close()
	fmt.Println("Query entries in map", mapName)
	adatypes.Central.Log.Debugf("New map request after clear map")
	readRequest, rErr := NewMapNameRequest(adabas, mapName)
	if !assert.NoError(t, rErr) {
		return rErr
	}
	defer readRequest.Close()
	fmt.Println("Clear all entries in map", mapName)
	// Need to call all and don't need to read the data for deleting all records
	readRequest.Limit = 0
	readRequest.QueryFields("")
	fmt.Println("Read request in map", mapName, "and delete in", readRequest.adabas.String(), readRequest.repository.Fnr)
	err = readRequest.ReadPhysicalSequenceWithParser(testCallback, deleteRequest)
	if !assert.NoError(t, err) {
		return err
	}
	err = deleteRequest.EndTransaction()
	fmt.Println("Cleared file")
	return err
}

func TestStoreMapFieldsPeriods(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(23)
	AddMapRepository(ada, 250)
	defer DelMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, serr := SearchMapRepository(ada, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()
	recErr := storeRequest.StoreFields("PERSONNEL-ID,SALARY,BONUS")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	if !assert.NotNil(t, storeRecord) {
		return
	}
	storeRecord.DumpValues()
	err = storeRecord.SetValue("PERSONNEL-ID", "555551")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValueWithIndex("SALARY", []uint32{1}, 123)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValueWithIndex("BONUS", []uint32{1, 1}, 22)
	if !assert.NoError(t, err) {
		return
	}
	storeRecord.DumpValues()
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()
}

func TestStoreUpdateMapField(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(23), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(23)
	AddMapRepository(ada, 250)
	defer DelMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, serr := SearchMapRepository(ada, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()
	recErr := storeRequest.StoreFields("PERSONNEL-ID")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	if !assert.NotNil(t, storeRecord) {
		return
	}
	err = storeRecord.SetValue("PERSONNEL-ID", "1111111")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()

	checkUpdateCorrectRead(t, "1111111", storeRecord.Isn)

	err = storeRecord.SetValue("PERSONNEL-ID", "9999999")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.Update(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()

	checkUpdateCorrectRead(t, "9999999", storeRecord.Isn)

}

func checkUpdateCorrectRead(t *testing.T, value string, isn adatypes.Isn) {
	id := NewAdabasID()
	copy(id.AdaID.User[:], []byte("CHECK   "))
	adabas, err := NewAdabasWithID("23", id)
	if !assert.NoError(t, err) {
		return
	}
	request := NewRequestAdabas(adabas, 16)
	defer request.Close()
	request.QueryFields("AA")
	result := &RequestResult{}
	err = request.ReadLogicalWithWithParser("AA="+value, nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		result.DumpValues()
	}
	assert.Equal(t, 1, len(result.Values))

}

func TestStoreWithMapLobFile(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	// dataRepository := &DatabaseURL{Adabas: NewAdabas(23), Fnr: 11}
	// prepareCreateTestMap(t, "LOBPUCTURE", "Lobpicture.systrans", dataRepository)

	//	clearAdabasFile(t, "24", 4)
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
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	fmt.Printf("Number of bytes read: %d/%d\n", n, len(data))
	if !assert.Equal(t, lengthPicture, len(data)) {
		return
	}

	h := sha256.New()
	h.Write(data[:20])
	fmt.Printf("SHA 20: %x\n", h.Sum(nil))

	h = sha256.New()
	h.Write(data[:1024])
	fmt.Printf("SHA 1024: %x\n", h.Sum(nil))
	h = sha256.New()
	h.Write(data[:10000])
	fmt.Printf("SHA 10000: %x\n", h.Sum(nil))
	h = sha256.New()
	h.Write(data)
	fmt.Printf("SHA ALL: %x\n", h.Sum(nil))

	ada := NewAdabas(23)
	AddMapRepository(ada, 4)
	defer DelMapRepository(ada, 4)

	adabasMap, serr := SearchMapRepository(ada, "LOBEXAMPLE")
	if !assert.NoError(t, serr) {
		return
	}
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()

	adatypes.Central.Log.Debugf("Store fields prepare Picture")
	recErr := storeRequest.StoreFields("Picture")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	if !assert.NotNil(t, storeRecord) {
		return
	}
	adatypes.Central.Log.Debugf("Seearch type Picture")
	adaV, sTerr := storeRequest.definition.SearchType("Picture")
	if !assert.NoError(t, sTerr) {
		return
	}
	if !assert.Equal(t, uint32(0), adaV.Length()) {
		return
	}
	adatypes.Central.Log.Debugf("Set value to Picture")
	storeRecord.SetValue("Picture", data)
	adatypes.Central.Log.Debugf("Done set value to Picture, searching ...")
	s, verr := storeRecord.SearchValue("Picture")
	if !assert.NoError(t, verr) {
		return
	}
	if !assert.Equal(t, len(data), len(s.Bytes())) {
		return
	}
	if !assert.Equal(t, lengthPicture, len(s.Bytes())) {
		return
	}

	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Store record into ISN=", storeRecord.Isn)
	storeRequest.EndTransaction()
	validateUsingAdabas(t, storeRecord.Isn)
	validateUsingMap(t, storeRecord.Isn)
}

func validateUsingAdabas(t *testing.T, isn adatypes.Isn) {
	fmt.Println("Validate using Adabas")
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 160)
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		err := request.QueryFields("DC")
		if !assert.NoError(t, openErr) {
			return
		}
		fmt.Println("After query fields")
		result := &RequestResult{}
		err = request.ReadISNWithParser(isn, nil, result)
		if assert.NoError(t, err) {
			picValue := result.Values[0].HashFields["DC"]
			//			fmt.Println("Dump result received ...")
			// fmt.Println(adatypes.FormatByteBuffer("RESULT:", picValue.Bytes()[:20]))
			// assert.Equal(t, lengthPicture, len(picValue.Bytes()))
			h := sha256.New()
			h.Write(picValue.Bytes()[:20])
			fmt.Printf("SHA 20: %x\n", h.Sum(nil))
			h = sha256.New()
			h.Write(picValue.Bytes()[:1024])
			fmt.Printf("SHA 1024: %x\n", h.Sum(nil))
			h = sha256.New()
			h.Write(picValue.Bytes()[:10000])
			fmt.Printf("SHA 10000: %x\n", h.Sum(nil))
			h = sha256.New()
			h.Write(picValue.Bytes())
			fmt.Printf("SHA ALL: %x\n", h.Sum(nil))
			assert.Equal(t, "b79169e9c696cff3005ca49fce8e91c7f8b8ecc61fecb0a58b644bc4b68d7689",
				hex.EncodeToString(h.Sum(nil)))
		}
	}
}

func validateUsingMap(t *testing.T, isn adatypes.Isn) {
	fmt.Println("Validate using Map")
	adabas := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewMapNameRequestRepo("LOBEXAMPLE", mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, request) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		err := request.QueryFields("Picture")
		if !assert.NoError(t, err) {
			return
		}
		fmt.Println("After query fields")
		result := &RequestResult{}
		err = request.ReadISNWithParser(isn, nil, result)
		if assert.NoError(t, err) {
			if !assert.NotNil(t, result) && !assert.NotNil(t, result.Values) {
				return
			}
			if !assert.True(t, len(result.Values) > 0) {
				return
			}
			if !assert.NotNil(t, result.Values[0].HashFields) {
				return
			}
			picValue := result.Values[0].HashFields["Picture"]
			if !assert.NotNil(t, picValue) {
				return
			}
			assert.Equal(t, lengthPicture, len(picValue.Bytes()))
			h := sha256.New()
			h.Write(picValue.Bytes())
			fmt.Printf("SHA: %x\n", h.Sum(nil))
			assert.Equal(t, "b79169e9c696cff3005ca49fce8e91c7f8b8ecc61fecb0a58b644bc4b68d7689",
				hex.EncodeToString(h.Sum(nil)))
		}
	}
}

func TestStoreMapMissing(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	fmt.Println("Validate using Map invalid")
	adabas := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewMapNameRequestRepo("NONMAP", mapRepository)
	if assert.Error(t, err) {
		if assert.Nil(t, request) {
			assert.Equal(t, "ADG0000014: Map NONMAP not found in repository", err.Error())
			return
		}
		return
	}
	if request != nil {
		request.Close()
	}
}
