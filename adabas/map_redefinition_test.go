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
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func importMaps(ada *Adabas, mr *Repository, fileName string) error {
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + fileName
	fmt.Println("Loading ....", fileName)
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	maps, perr := ParseJSONFileForFields(file)
	if perr != nil {
		fmt.Println("Error parsing file", perr)
		return perr
	}
	fmt.Println("Number of maps", len(maps))
	for _, m := range maps {
		fmt.Println("MAP", m.Name)
		fmt.Printf("  %s %d\n", m.Data.URL.String(), m.Data.Fnr)
		for _, f := range m.Fields {
			fmt.Printf("   ln=%s sn=%s len=%d format=%s content=%s\n", f.LongName, f.ShortName, f.Length, f.FormatType, f.ContentType)
		}
		m.Repository = &mr.DatabaseURL
		err = m.Store()
		if err != nil {
			return err
		}
	}
	return nil
}

func checkMapAvailable(mapName, fileName string) error {
	adabas, _ := NewAdabas(23)
	defer adabas.Close()
	mr := NewMapRepository(adabas, 250)
	_, err := mr.SearchMap(adabas, mapName)
	if err != nil {
		return importMaps(adabas, mr, fileName)
	}
	fmt.Println("Map loaded: ", mapName)
	return nil
}

func loadTestData() (err error) {

	ada, _ := NewAdabas(23)
	defer ada.Close()
	repository := NewMapRepository(ada, 250)
	storeRequest, _ := NewStoreRequest("MapperRedefTest", ada, repository)
	defer storeRequest.Close()
	err = storeRequest.StoreFields("*")
	if err != nil {
		fmt.Println("Store Fields", err)
		//return err
		panic("Store fields " + err.Error())
	}

	for i := 10; i < 200; i += 10 {
		storeRecord, serr := storeRequest.CreateRecord()
		if serr != nil {
			fmt.Println("Create record", serr)
			return serr
		}
		storeRecord.SetValue("Personel-Id", fmt.Sprintf("REDEF%d", i/10))
		storeRecord.SetValue("REFFIELD1", 124+i)
		storeRecord.SetValue("REFFIELD2", 12+i)
		storeRecord.SetValue("REFFIELD3", i)
		storeRecord.SetValue("LONGPART", "ABCDEFGHIJ")
		serr = storeRequest.Store(storeRecord)
		if serr != nil {
			fmt.Println("Store request", serr)
			return serr
			//panic("Store request: " + serr.Error())
		}
		serr = storeRequest.EndTransaction()
		if serr != nil {
			fmt.Println("End transaction", serr)
			return serr
		}
	}
	return nil
}

func TestReadPartRedefinition(t *testing.T) {
	initTestLogWithFile(t, "redefinition.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	loadTestData()

	merr := checkMapAvailable("MapperRedefTest", "Redefinition.json")
	if !assert.NoError(t, merr) {
		return
	}

	connection, err := NewConnection("ada;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("MapperRedefTest")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	fmt.Println("Map loaded", request.adabasMap.String())
	err = request.QueryFields("Personel-Id,LONGSTRING")
	if !assert.NoError(t, err) {
		fmt.Println("Error query request", err)
		return
	}
	request.Limit = 4
	result, qerr := request.ReadLogicalWith("Personel-Id=[REDEF0:REDEFA]")
	if !assert.NoError(t, qerr) {
		fmt.Println("Error read sequence", qerr)
		return
	}
	result.DumpValues()
	if assert.Equal(t, 4, len(result.Values)) {
		record := result.Values[0]
		f := record.HashFields["LONGSTRING"]
		if assert.NotNil(t, f) {
			assert.Equal(t, "", f.String())
		}
	}
}

func TestRedefinition(t *testing.T) {
	initTestLogWithFile(t, "redefinition.log")
	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	merr := checkMapAvailable("MapperRedefTest", "Redefinition.json")
	if !assert.NoError(t, merr) {
		return
	}

	connection, err := NewConnection("ada;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("MapperRedefTest")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	fmt.Println("Map loaded", request.adabasMap.String())
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		fmt.Println("Error query request", err)
		return
	}
	request.Limit = 4

	result, qerr := request.ReadLogicalWith("Personel-Id=[REDEF0:REDEFA]")
	if !assert.NoError(t, qerr) {
		fmt.Println("Error read sequence", qerr)
		return
	}
	result.DumpValues()
	if assert.Equal(t, 4, len(result.Values)) {
		record := result.Values[0]
		f := record.HashFields["LONGSTRING"]
		if assert.NotNil(t, f) {
			assert.Equal(t, "", f.String())
		}
	}
}
