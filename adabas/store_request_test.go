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
	log "github.com/sirupsen/logrus"

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

func TestStoreAdabasFields(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

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
	adabas := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 250)
	sm, err := mr.SearchMap(adabas, mapName)
	if err == nil {
		assert.NotNil(t, sm)
		return nil
	}

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

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 11}
	perr := prepareCreateTestMap(t, massLoadEmployees, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
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

	log.Infof("TEST: %s", t.Name())

	fmt.Println("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}

	ada := NewAdabas(adabasModDBID)
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

	log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
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

	log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
	defer ada.Close()

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
	storeRequest.EndTransaction()

	log.Infof("First validate data in database ....")
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

	log.Infof("Second validate data in database ....")
	checkUpdateCorrectRead(t, "9999999", storeRecord.Isn)

}

func checkUpdateCorrectRead(t *testing.T, value string, isn adatypes.Isn) {
	checkUpdateCorrectReadNumber(t, value, []adatypes.Isn{isn}, 1)
}

func checkUpdateCorrectReadNumber(t *testing.T, value string, isns []adatypes.Isn, number int) {
	id := NewAdabasID()
	copy(id.AdaID.User[:], []byte("CHECK   "))
	adabas, err := NewAdabasWithID(adabasModDBIDs, id)
	if !assert.NoError(t, err) {
		return
	}
	request := NewRequestAdabas(adabas, 16)
	defer request.Close()
	request.QueryFields("AA")
	result := &Response{}
	err = request.ReadLogicalWithWithParser("AA=["+value+":"+value+"a]", nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		result.DumpValues()
	}
	assert.Equal(t, number, len(result.Values))

}

func TestStoreWithMapLobFile(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	// dataRepository := &DatabaseURL{Adabas: NewAdabas(adabasModDBID), Fnr: 11}
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

	ada := NewAdabas(adabasModDBID)
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
	adabas := NewAdabas(adabasModDBID)
	request := NewRequestAdabas(adabas, 160)
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		err := request.QueryFields("DC")
		if !assert.NoError(t, openErr) {
			return
		}
		fmt.Println("After query fields")
		result := &Response{}
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
	adabas := NewAdabas(adabasModDBID)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewMapNameRequestRepo("LOBEXAMPLE", adabas, mapRepository)
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
		result := &Response{}
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

	log.Infof("TEST: %s", t.Name())

	fmt.Println("Validate using Map invalid")
	adabas := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewMapNameRequestRepo("NONMAP", adabas, mapRepository)
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
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	adabas := NewAdabas(adabasModDBID)
	mr := NewMapRepository(adabas, 250)
	mapName := massLoadSystransStore

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

		storeRequest.Store(storeRecord)
	}
	fmt.Println("End of transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Done")

}

func TestStoreEndTransaction(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
	defer ada.Close()

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

	storeRequest.EndTransaction()

	log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 10)
}

func TestStoreCloseWithBackout(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
	defer ada.Close()

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

	log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "CLTEST", isns, 0)
}

func TestStoreBackout(t *testing.T) {
	f := initTestLogWithFile(t, "store.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(t, massLoadSystransStore, massLoadSystrans, dataRepository)
	if perr != nil {
		return
	}
	ada := NewAdabas(adabasModDBID)
	defer ada.Close()

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

	storeRequest.BackoutTransaction()

	log.Infof("First validate data in database ....")
	checkUpdateCorrectReadNumber(t, "BTTEST", isns, 0)
}
