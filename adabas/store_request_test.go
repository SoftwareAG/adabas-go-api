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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

const (
	storeEmployeesMap     = "EMPLDDM-GOSTORE"
	massLoadEmployees     = "EMPLDDM-GOLOAD"
	massLoadSystrans      = "Empl-MassLoad.systrans"
	massLoadSystransStore = "EMPLDDM-GOLOAD-STORE"
	mapVehicles           = "VEHICLES"
	vehicleSystransStore  = "Vehi.systrans"
	lengthPicture         = 1386643
)

func TestStoreErrorCase(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	s, err := NewStoreRequest()
	assert.Nil(t, s)
	assert.Error(t, err)
	s, err = NewStoreRequest("aa", "a")
	assert.Nil(t, s)
	assert.Error(t, err)
	fmt.Println(err.Error())
	s, err = NewStoreRequest("aa", 1)
	assert.Nil(t, s)
	assert.Error(t, err)
	fmt.Println(err.Error())
	s, err = NewStoreRequest("99999", 1)
	assert.Nil(t, s)
	assert.Error(t, err)
	fmt.Println(err.Error())
	s, err = NewStoreRequest(256, 1)
	assert.Nil(t, s)
	assert.Error(t, err)
	fmt.Println(err.Error())
}

func TestStoreAdabasFields(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}

	storeRequest, _ := NewStoreRequest("23", 16)
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
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

}

func prepareCreateTestMap(mapName string, fileName string, dataRepository *DatabaseURL) error {
	adabas, _ := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 250)
	sm, err := mr.SearchMap(adabas, mapName)
	if err == nil {
		if sm != nil {
			return nil
		}
		return errors.New("Empty search result of map")
	}

	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + fileName
	m, merr := mr.ImportMapRepository(adabas, "*", name, dataRepository)
	if merr != nil {
		fmt.Println("Error importing map", merr)
		return merr
	}
	//	fmt.Printf("Successfull importing map: %s\n", mapName)
	if len(m) == 1 {
		m[0].Name = mapName
		err = m[0].Store()
		if err != nil {
			fmt.Printf("Error storing map %s %v\n", mapName, err)
			return err
		}
	}
	return nil
}

func TestStoreRequestConstructor(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	ada, _ := NewAdabas(adabasStatDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest("EMPLOYEES-NAT-DDM", ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, storeRequest)

}

func TestStoreFailMapFieldsCheck(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 11}
	perr := prepareCreateTestMap(massLoadEmployees, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	adatypes.Central.Log.Debugf("Prepare of test finished")
	ada, _ := NewAdabas(adabasModDBID)
	AddGlobalMapRepository(ada.URL, 250)
	defer DelGlobalMapRepository(ada.URL, 250)
	adatypes.Central.Log.Debugf("Search map in repository")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadEmployees)
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
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	fmt.Println("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}

	ada, _ := NewAdabas(adabasModDBID)
	AddGlobalMapRepository(ada.URL, 250)
	defer DelGlobalMapRepository(ada.URL, 250)

	fmt.Println("Prepare clear test map")
	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	fmt.Println("Store map request")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
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
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
}

func clearAdabasFile(t *testing.T, target string, fnr Fnr) error {
	fmt.Println("Clear Adabas file", target, "/", fnr)
	id := NewAdabasID()
	adabas, err := NewAdabas(target, id)
	if err != nil {
		return err
	}
	deleteRequest := NewDeleteRequestAdabas(adabas, fnr)
	defer deleteRequest.Close()
	readRequest, _ := NewReadRequest(adabas, fnr)
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
	adabasMap, _, err := SearchMapRepository(adabas.ID, mapName)
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
	readRequest, rErr := NewReadRequest(adabas, mapName)
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
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	AddGlobalMapRepository(ada.URL, 250)
	defer DelGlobalMapRepository(ada.URL, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
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
	err = storeRequest.EndTransaction()
	assert.NoError(t, err)
}

func TestStoreUpdateMapField(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()

	AddGlobalMapRepository(ada, 250)
	defer DelGlobalMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	adatypes.Central.Log.Debugf("Create new map store request")
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
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

	adatypes.Central.Log.Infof("First validate data in database ....")
	checkUpdateCorrectRead(t, "1111111", storeRecord.Isn)

	err = storeRecord.SetValue("PERSONNEL-ID", "9999999")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.Update(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

	adatypes.Central.Log.Infof("Second validate data in database ....")
	checkUpdateCorrectRead(t, "9999999", storeRecord.Isn)

}

func checkUpdateCorrectRead(t *testing.T, value string, isn adatypes.Isn) {
	checkUpdateCorrectReadNumber(t, value, []adatypes.Isn{isn}, 1)
}

func checkUpdateCorrectReadNumber(t *testing.T, value string, isns []adatypes.Isn, number int) {
	id := NewAdabasID()
	copy(id.AdaID.User[:], []byte("CHECK   "))
	adabas, err := NewAdabas(adabasModDBIDs, id)
	if !assert.NoError(t, err) {
		return
	}
	request, _ := NewReadRequest(adabas, 16)
	defer request.Close()
	request.QueryFields("AA")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[" + value + ":" + value + "a]")
	if !assert.NoError(t, err) {
		return
	}
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		_ = result.DumpValues()
	}
	assert.Equal(t, number, len(result.Values))

}

func TestStoreWithMapLobFile(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	p = p + "/../files/img/106-0687_IMG.JPG"
	f, err := os.Open(p)
	if !assert.NoError(t, err) {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if !assert.NoError(t, err) {
		fmt.Println(err)
		return
	}
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	if !assert.NoError(t, err) {
		fmt.Println(err)
		return
	}
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

	ada, _ := NewAdabas(adabasModDBID)
	AddGlobalMapRepository(ada, 4)
	defer DelGlobalMapRepository(ada, 4)
	DumpGlobalMapRepositories()

	adabasMap, _, serr := SearchMapRepository(ada.ID, "LOBEXAMPLE")
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
	_ = storeRecord.SetValue("Picture", data)
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
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	validateUsingAdabas(t, storeRecord.Isn)
	validateUsingMap(t, storeRecord.Isn)
}

func validateUsingAdabas(t *testing.T, isn adatypes.Isn) {
	fmt.Println("Validate using Adabas and ISN=", isn)
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 202)
	defer request.Close()
	_, openErr := request.Open()
	if assert.NoError(t, openErr) {
		err := request.QueryFields("DC")
		if !assert.NoError(t, err) {
			return
		}
		fmt.Println("Query fields defined, send read ...")
		var result *Response
		result, err = request.ReadISN(isn)
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
			fmt.Println("Data validated with classic methods")
		} else {
			fmt.Println("Error validating data with classic methods")
		}
	}
}

func validateUsingMap(t *testing.T, isn adatypes.Isn) {
	fmt.Println("Validate using Map and ISN=", isn)
	adabas, _ := NewAdabas(adabasModDBID)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest("LOBEXAMPLE", adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, request) {
		return
	}
	defer request.Close()
	_, openErr := request.Open()
	if assert.NoError(t, openErr) {
		err := request.QueryFields("Picture")
		if !assert.NoError(t, err) {
			return
		}
		fmt.Println("Query defined, read record ...")
		var result *Response
		result, err = request.ReadISN(isn)
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
	fmt.Println("Data validated with map methods")
}

func TestStoreMapMissing(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	fmt.Println("Validate using Map invalid")
	adabas, _ := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest("NONMAP", adabas, mapRepository)
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

func TestStorePeriod(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	mr := NewMapRepository(adabas, 250)
	mapName := massLoadSystransStore

	readRequest, rErr := NewReadRequest(mapName, adabas, mr)
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

	recErr := storeRequest.StoreFields("PERSONNEL-ID,FULL-NAME,SALARY,BONUS")
	if !assert.NoError(t, recErr) {
		return
	}

	for i := 0; i < 1; i++ {
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
		err = storeRecord.SetValueWithIndex("SALARY", []uint32{3}, 100000)
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("SALARY[2]", 50000)
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("BONUS[1][1]", 1000)
		if !assert.NoError(t, err) {
			return
		}

		err = storeRequest.Store(storeRecord)
		if !assert.NoError(t, err) {
			return
		}
	}
	fmt.Println("End of transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Done")

}

func TestStoreEndTransaction(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()

	AddGlobalMapRepository(ada, 250)
	defer DelGlobalMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	adatypes.Central.Log.Debugf("Create new map store request")
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()

	recErr := storeRequest.StoreFields("PERSONNEL-ID,NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	var isns []adatypes.Isn
	for i := 0; i < 10; i++ {
		storeRecord, rErr := storeRequest.CreateRecord()
		if !assert.NoError(t, rErr) {
			return
		}
		if !assert.NotNil(t, storeRecord) {
			return
		}
		err = storeRecord.SetValue("PERSONNEL-ID", fmt.Sprintf("CLTEST%02d", (i+1)))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("NAME", fmt.Sprintf("CLTEST%d", i))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRequest.Store(storeRecord)
		if !assert.NoError(t, err) {
			return
		}
		isns = append(isns, storeRecord.Isn)
	}
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 10)

	err = storeRequest.EndTransaction()
	assert.NoError(t, err)
	adatypes.Central.Log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 10)
}

func TestStoreCloseWithBackout(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()

	AddGlobalMapRepository(ada, 250)
	defer DelGlobalMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	adatypes.Central.Log.Debugf("Create new map store request")
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()

	recErr := storeRequest.StoreFields("PERSONNEL-ID,NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	var isns []adatypes.Isn
	for i := 0; i < 10; i++ {
		storeRecord, rErr := storeRequest.CreateRecord()
		if !assert.NoError(t, rErr) {
			return
		}
		if !assert.NotNil(t, storeRecord) {
			return
		}
		err = storeRecord.SetValue("PERSONNEL-ID", fmt.Sprintf("CLTEST%02d", (i+1)))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("NAME", fmt.Sprintf("CLTEST%d", i))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRequest.Store(storeRecord)
		if !assert.NoError(t, err) {
			return
		}
		isns = append(isns, storeRecord.Isn)
	}
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 10)

	storeRequest.Close()

	adatypes.Central.Log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 0)
}

func TestStoreBackout(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()

	AddGlobalMapRepository(ada, 250)
	defer DelGlobalMapRepository(ada, 250)

	clearErr := clearMap(t, ada, massLoadSystransStore)
	if !assert.NoError(t, clearErr) {
		return
	}
	adatypes.Central.Log.Debugf("Search map after clear map")
	adabasMap, _, serr := SearchMapRepository(ada.ID, massLoadSystransStore)
	if !assert.NoError(t, serr) {
		return
	}
	adatypes.Central.Log.Debugf("Create new map store request")
	storeRequest, err := NewAdabasMapNameStoreRequest(ada, adabasMap)
	if !assert.NoError(t, err) {
		return
	}
	defer storeRequest.Close()

	recErr := storeRequest.StoreFields("PERSONNEL-ID,NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	var isns []adatypes.Isn
	for i := 0; i < 10; i++ {
		storeRecord, rErr := storeRequest.CreateRecord()
		if !assert.NoError(t, rErr) {
			return
		}
		if !assert.NotNil(t, storeRecord) {
			return
		}
		err = storeRecord.SetValue("PERSONNEL-ID", fmt.Sprintf("BTTEST%02d", (i+1)))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRecord.SetValue("NAME", fmt.Sprintf("BTTEST%d", i))
		if !assert.NoError(t, err) {
			return
		}
		err = storeRequest.Store(storeRecord)
		if !assert.NoError(t, err) {
			return
		}
		isns = append(isns, storeRecord.Isn)
	}
	checkUpdateCorrectReadNumber(t, "BTTEST", isns, 10)

	err = storeRequest.BackoutTransaction()
	assert.NoError(t, err)
	adatypes.Central.Log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "BTTEST", isns, 0)
}

func TestUpdateWithMapLob(t *testing.T) {
	initTestLogWithFile(t, "store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

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
	if !assert.Equal(t, 1386643, len(data)) {
		return
	}

	h := sha256.New()
	h.Write(data)
	fmt.Printf("SHA FILE: %x\n", h.Sum(nil))
	chkPic := fmt.Sprintf("%X", h.Sum(nil))

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	storeRequest, err := connection.CreateMapStoreRequest("LOBEXAMPLE")
	if !assert.NoError(t, err) {
		return
	}

	adatypes.Central.Log.Debugf("Store fields prepare Picture")
	fmt.Println("Store fields, insert record")
	recErr := storeRequest.StoreFields("Filename,PictureSHAchecksum,ThumbnailSHAchecksum")
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
	_ = storeRecord.SetValue("Filename", "lobtest")
	_ = storeRecord.SetValue("PictureSHAchecksum", chkPic)
	_ = storeRecord.SetValue("ThumbnailSHAchecksum", "x")

	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Store record into ISN=", storeRecord.Isn)
	if !assert.True(t, storeRecord.Isn > 0) {
		return
	}
	isn := storeRecord.Isn

	recErr = storeRequest.StoreFields("Picture")
	if !assert.NoError(t, recErr) {
		return
	}
	storeRecord, rErr = storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return
	}
	if !assert.NotNil(t, storeRecord) {
		return
	}
	storeRecord.Isn = isn
	storeRecord.LobEndTransaction = true
	_ = storeRecord.SetValue("Picture", data)
	fmt.Println("Update record into ISN=", storeRecord.Isn)
	adatypes.Central.Log.Debugf("Update data in ISN=%d field Picture", isn)
	err = storeRequest.Update(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	adatypes.Central.Log.Debugf("Dpne update data in ISN=%d field Picture", isn)

	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

	readRequest, rerr := connection.CreateMapReadRequest("LOBEXAMPLE")
	if !assert.NoError(t, rerr) {
		return
	}

	rerr = readRequest.QueryFields("Picture")
	if !assert.NoError(t, rerr) {
		return
	}
	result, rrErr := readRequest.ReadISN(isn)
	if !assert.NoError(t, rrErr) {
		return
	}
	assert.Equal(t, 1, len(result.Values))
	v, _ := result.Values[0].SearchValue("Picture")
	vb := v.Bytes()
	if !assert.Equal(t, 1386643, len(vb)) {
		return
	}
	h = sha256.New()
	h.Write(vb)
	fmt.Printf("SHA SAVED: %x\n", h.Sum(nil))
	savedPic := fmt.Sprintf("%X", h.Sum(nil))
	assert.Equal(t, chkPic, savedPic)
}
