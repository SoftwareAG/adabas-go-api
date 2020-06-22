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
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

var lastIsn adatypes.Isn

func BenchmarkConnection_cached(b *testing.B) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: BenchmarkConnection_cached")

	adatypes.InitDefinitionCache()
	defer adatypes.FinitDefinitionCache()

	for i := 0; i < 1000; i++ {
		// fmt.Print(".")
		// if (i+1)%100 == 0 {
		// 	fmt.Printf("%d/1000\n", i)
		// }
		err = readAll(b)
		if err != nil {
			return
		}
	}
}

func readAll(b *testing.B) error {
	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(b, cerr) {
		return cerr
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(b, rerr) {
		fmt.Println("Error create request", rerr)
		return rerr
	}
	err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
	if !assert.NoError(b, err) {
		return err
	}
	request.Limit = 0
	result, rErr := request.ReadLogicalBy("NAME")
	if !assert.NoError(b, rErr) {
		return rErr
	}
	if !assert.Equal(b, 1107, len(result.Values)) {
		return fmt.Errorf("Error length mismatch")
	}
	return nil
}

func BenchmarkConnection_noreconnect(b *testing.B) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: BenchmarkConnection_noreconnect")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(b, cerr) {
		return
	}
	defer connection.Close()

	for i := 0; i < 1000; i++ {
		// fmt.Print(".")
		// if (i+1)%100 == 0 {
		// 	fmt.Printf("%d/1000\n", i)
		// }
		request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
		if !assert.NoError(b, rerr) {
			fmt.Println("Error create request", rerr)
			return
		}
		err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
		if !assert.NoError(b, err) {
			return
		}
		request.Limit = 0
		var result *Response
		result, err = request.ReadLogicalBy("NAME")
		if !assert.NoError(b, err) {
			return
		}
		if !assert.Equal(b, 1107, len(result.Values)) {
			return
		}
	}
}

func TestAuth(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	//connection, cerr := NewConnection("acj;map;config=[177(adatcp://pinas:60177),4]")
	connection, cerr := NewConnection("acj;target=" + adabasStatDBIDs + ";auth=NONE,user=TestAuth,id=4,host=xx")
	if !assert.NoError(t, cerr) {
		return
	}
	assert.Contains(t, connection.ID.String(), "xx      :TestAuth [4] ")
	connection.Close()

	connection, cerr = NewConnection("acj;target=" + adabasStatDBIDs + ";auth=NONE,user=ABCDEFGHIJ,id=65535,host=KLMNOPQRSTUVWXYZ")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	assert.Contains(t, connection.ID.String(), "KLMNOPQR:ABCDEFGH [65535] ")
}

func TestConnectionRemoteMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[177(adatcp://" + adabasTCPLocation() + "),4];auth=NONE,user=TCRemMap")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	for i := 0; i < 5; i++ {
		request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
		if !assert.NoError(t, rerr) {
			fmt.Println("Error create request", rerr)
			return
		}
		err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 0
		var result *Response
		result, err = request.ReadLogicalBy("NAME")
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, 1107, len(result.Values)) {
			return
		}
	}
}

func TestConnectionMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map=EMPLOYEES-NAT-DDM;config=[" + adabasStatDBIDs + ",4];auth=NONE,user=XMAP")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	for i := 0; i < 5; i++ {
		fmt.Printf("%d. instance called\n", i+1)
		request, rerr := connection.CreateReadRequest()
		if !assert.NoError(t, rerr) {
			fmt.Println("Error create request", rerr)
			return
		}
		err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 0
		var result *Response
		result, err = request.ReadLogicalBy("NAME")
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, 1107, len(result.Values)) {
			return
		}
	}
}

func BenchmarkConnection_noreconnectremote(b *testing.B) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: BenchmarkConnection_noreconnectremote")

	connection, cerr := NewConnection("acj;map;config=[177(adatcp://" + adabasTCPLocation() + "),4]")
	if !assert.NoError(b, cerr) {
		return
	}
	defer connection.Close()

	for i := 0; i < 1000; i++ {
		// fmt.Print(".")
		// if (i+1)%100 == 0 {
		// 	fmt.Printf("%d/1000\n", i)
		// }
		request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
		if !assert.NoError(b, rerr) {
			fmt.Println("Error create request", rerr)
			return
		}
		err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
		if !assert.NoError(b, err) {
			return
		}
		request.Limit = 0
		var result *Response
		result, err = request.ReadLogicalBy("NAME")
		if !assert.NoError(b, err) {
			return
		}
		if !assert.Equal(b, 1107, len(result.Values)) {
			return
		}
	}
}

func checkVehicleMap(mapName string, jsonImport string) error {
	databaseURL := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 250}
	mr := NewMapRepositoryWithURL(*databaseURL)
	a, _ := NewAdabas(adabasStatDBID)
	defer a.Close()
	_, err := mr.SearchMap(a, mapName)
	if err != nil {
		fmt.Println("Search map, try loading map ...", err)
		maps, err := LoadJSONMap(jsonImport)
		if err != nil {
			return err
		}
		fmt.Println("Number of maps", len(maps))
		for _, m := range maps {
			m.Repository = databaseURL
			fmt.Println("Load map ...", m.Name)
			err = m.Store()
			if err != nil {
				fmt.Println("Error loading map ...", err)
				return err
			}
		}
	}
	return nil
}

func TestConnectionWithMultipleMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_map.log")

	cerr := checkVehicleMap("VehicleMap", "VehicleMap.json")
	if !assert.NoError(t, cerr) {
		return
	}
	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250|24,4]")
	if !assert.NoError(t, cerr) {
		return
	}

	defer connection.Close()
	fmt.Println("Connection : ", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		if assert.Equal(t, 3, len(result.Values)) {
			ae := result.Values[1].HashFields["NAME"]
			assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}
	request, err = connection.CreateMapReadRequest("VehicleMap")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("Vendor,Model")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("Vendor=RENAULT")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		if assert.Equal(t, 57, len(result.Values)) {
			ae := result.Values[1].HashFields["Vendor"]
			assert.Equal(t, "RENAULT", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}

}

func TestConnectionMapPointingToRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_map.log")

	cerr := checkVehicleMap("REMPL11", "rempl11.json")
	if !assert.NoError(t, cerr) {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250];auth=NONE,user=TCMapPoin,id=4,host=REMOTE")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection : ", connection)
	request, err := connection.CreateMapReadRequest("REMPL11")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		if assert.NotNil(t, result) {
			if assert.Equal(t, 3, len(result.Values)) {
				ae := result.Values[1].HashFields["NAME"]
				assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
				ei64, xErr := ae.Int64()
				assert.Error(t, xErr, "Error should be send if value is string")
				assert.Equal(t, int64(0), ei64)
			}
		}
	}
}

func copyRecordData(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	record := x.(*Record)
	fmt.Println(adaValue.Type().Name(), "=", adaValue.String())
	err := record.SetValueWithIndex(adaValue.Type().Name(), nil, adaValue.Value())
	if err != nil {
		fmt.Println("Error add Value: ", err)
		return adatypes.EndTraverser, err
	}
	val, _ := record.SearchValue(adaValue.Type().Name())
	fmt.Println("Search Value", val.String())
	return adatypes.Continue, nil
}

func copyData(adabasRequest *adatypes.Request, x interface{}) (err error) {
	store := x.(*StoreRequest)
	var record *Record
	record, err = store.CreateRecord()
	if err != nil {
		fmt.Printf("Error creating record %v\n", err)
		return
	}
	tm := adatypes.TraverserValuesMethods{EnterFunction: copyRecordData}
	adabasRequest.Definition.TraverseValues(tm, record)
	fmt.Println("Record=", record.String())

	adatypes.Central.Log.Debugf("Store init ..........")
	err = store.Store(record)
	if err != nil {
		return err
	}
	adatypes.Central.Log.Debugf("Store done ..........")

	return
}

func TestConnectionCopyMapTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	databaseURL := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4}
	mr := NewMapRepositoryWithURL(*databaseURL)
	ada, _ := NewAdabas(adabasModDBID)
	_, err := mr.SearchMap(ada, "COPYEMPL")
	if err != nil {
		maps, merr := LoadJSONMap("COPYEMPL.json")
		if !assert.NoError(t, merr) {
			return
		}
		fmt.Println("Number of maps", len(maps))
		assert.Equal(t, 1, len(maps))
		maps[0].Repository = databaseURL
		err = maps[0].Store()
		if !assert.NoError(t, err) {
			return
		}
	}
	defer ada.Close()
	DumpGlobalMapRepositories()
	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	rconnection, rerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, rerr) {
		return
	}
	defer rconnection.Close()
	fmt.Println("Connection : ", connection)
	store, err := connection.CreateMapStoreRequest("COPYEMPL")
	if !assert.NoError(t, err) {
		return
	}
	err = store.StoreFields("NAME,PERSONNEL-ID")
	if !assert.NoError(t, err) {
		return
	}
	request, rerr := rconnection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, rerr) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		result := &Response{}
		fmt.Println("Read logigcal data:")
		err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100000:11101000]", copyData, store)
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		if !assert.Equal(t, 0, len(result.Values)) {
			return
		}
	}
	err = store.EndTransaction()
	assert.NoError(t, err)

	connection.Close()
}

func ExampleConnection_readWithMap() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: ExampleAdabas_readFileDefinitionMap")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
	if err != nil {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100314:11100317]")
	if err != nil {
		fmt.Println("Error read logical data", err)
		return
	}

	result.DumpValues()
	// Output:Read logigcal data:
	// Dump all result values
	// Record Isn: 0393
	//   PERSONNEL-ID = > 11100314 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > WOLFGANG             <
	//    NAME = > SCHMIDT              <
	// Record Isn: 0261
	//   PERSONNEL-ID = > 11100315 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > GLORIA               <
	//    NAME = > MERTEN               <
	// Record Isn: 0262
	//   PERSONNEL-ID = > 11100316 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > HEINZ                <
	//    NAME = > RAMSER               <
	// Record Isn: 0263
	//   PERSONNEL-ID = > 11100317 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > ALFONS               <
	//    NAME = > DORSCH               <
}

func ExampleConnection_readWithMapFormatted() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: ExampleConnection_readWithMapFormatted")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID,BIRTH")
	if err != nil {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100314:11100317]")
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	result.DumpValues()
	// Output:Read logigcal data:
	// Dump all result values
	// Record Isn: 0393
	//   PERSONNEL-ID = > 11100314 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > WOLFGANG             <
	//    NAME = > SCHMIDT              <
	//   BIRTH = > 1953/08/18 <
	// Record Isn: 0261
	//   PERSONNEL-ID = > 11100315 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > GLORIA               <
	//    NAME = > MERTEN               <
	//   BIRTH = > 1949/11/02 <
	// Record Isn: 0262
	//   PERSONNEL-ID = > 11100316 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > HEINZ                <
	//    NAME = > RAMSER               <
	//   BIRTH = > 1978/12/23 <
	// Record Isn: 0263
	//   PERSONNEL-ID = > 11100317 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > ALFONS               <
	//    NAME = > DORSCH               <
	//   BIRTH = > 1948/02/29 <
}

func ExampleConnection_readFileDefinitionMapGroup() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: ExampleConnection_readFileDefinitionMapGroup")
	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("FULL-NAME,PERSONNEL-ID,SALARY")
	if err != nil {
		fmt.Println("Error query fields for request", err)
		return
	}
	request.Limit = 0
	fmt.Println("Read logigcal data:")
	var result *Response
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100315:11100316]")
	if err != nil {
		fmt.Println("Error read logical data", err)
		return
	}
	result.DumpValues()
	// Output: Read logigcal data:
	// Dump all result values
	// Record Isn: 0261
	//   PERSONNEL-ID = > 11100315 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > GLORIA               <
	//    NAME = > MERTEN               <
	//    MIDDLE-I = > E <
	//   INCOME = [ 2 ]
	//    SALARY[01] = > 19076 <
	//    SALARY[02] = > 18000 <
	// Record Isn: 0262
	//   PERSONNEL-ID = > 11100316 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > HEINZ                <
	//    NAME = > RAMSER               <
	//    MIDDLE-I = > E <
	//   INCOME = [ 1 ]
	//    SALARY[01] = > 28307 <
}

func BenchmarkConnection_simple(b *testing.B) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: BenchmarkConnection_simple")

	for i := 0; i < 100; i++ {
		// fmt.Print(".")
		// if (i+1)%100 == 0 {
		// 	fmt.Printf("%d/1000\n", i)
		// }
		err = readAll(b)
		if err != nil {
			return
		}
	}
}

func addEmployeeRecord(t *testing.T, storeRequest *StoreRequest, val string) error {
	storeRecord16, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return rErr
	}
	err := storeRecord16.SetValue("PERSONNEL-ID", val)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRecord16.SetValue("FIRST-NAME", "THORSTEN "+val)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRecord16.SetValue("MIDDLE-I", "TKN")
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRecord16.SetValue("NAME", "STORAGE_MAP")
	if !assert.NoError(t, err) {
		return err
	}
	// storeRecord16.DumpValues()
	// fmt.Println("Stored Employees request")
	adatypes.Central.Log.Debugf("Vehicles store started")
	err = storeRequest.Store(storeRecord16)
	if !assert.NoError(t, err) {
		return err
	}

	return nil
}

func addVehiclesRecord(t *testing.T, storeRequest *StoreRequest, val string) error {
	storeRecord, rErr := storeRequest.CreateRecord()
	if !assert.NoError(t, rErr) {
		return rErr
	}
	err := storeRecord.SetValue("REG-NUM", val)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRecord.SetValue("MAKE", "Concept "+val)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRecord.SetValue("MODEL", "Tesla")
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return err
	}
	lastIsn = storeRecord.Isn

	return nil
}

const multipleMapRefName = "M16555"
const multipleMapRefName2 = "M19555"

func TestConnectionSimpleMultipleMapStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}
	nr, cerr := checkStoreByFile(t, adabasModDBIDs, 16, multipleMapRefName)
	assert.NoError(t, cerr)
	assert.Equal(t, 0, nr)
	nr, cerr = checkStoreByFile(t, adabasModDBIDs, 19, multipleMapRefName2)
	assert.NoError(t, cerr)
	assert.Equal(t, 0, nr)

	adatypes.Central.Log.Infof("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	dataRepository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 19}
	vehicleMapName := mapVehicles + "Go"
	perr = prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}

	adatypes.Central.Log.Infof("Create connection...")
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest16, err := connection.CreateMapStoreRequest(massLoadSystransStore)
	if !assert.NoError(t, err) {
		return
	}
	recErr := storeRequest16.StoreFields("PERSONNEL-ID,FULL-NAME")
	if !assert.NoError(t, recErr) {
		return
	}
	err = addEmployeeRecord(t, storeRequest16, multipleMapRefName+"_0")
	if err != nil {
		return
	}
	storeRequest19, cErr := connection.CreateMapStoreRequest(vehicleMapName)
	if !assert.NoError(t, cErr) {
		return
	}
	recErr = storeRequest19.StoreFields("REG-NUM,CAR-DETAILS")
	if !assert.NoError(t, recErr) {
		return
	}
	err = addVehiclesRecord(t, storeRequest19, multipleMapRefName2+"_0")
	if !assert.NoError(t, err) {
		return
	}
	for i := 1; i < 10; i++ {
		x := strconv.Itoa(i)
		err = addEmployeeRecord(t, storeRequest16, multipleMapRefName+"_"+x)
		if !assert.NoError(t, err) {
			return
		}

	}
	err = addVehiclesRecord(t, storeRequest19, multipleMapRefName2+"_1")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("End transaction")
	connection.EndTransaction()
	fmt.Println("Check stored data")

	adatypes.Central.Log.Infof("Check stored data")
	nr, cerr = checkStoreByFile(t, adabasModDBIDs, 16, multipleMapRefName)
	assert.NoError(t, cerr)
	assert.Equal(t, 10, nr)
	nr, cerr = checkStoreByFile(t, adabasModDBIDs, 19, multipleMapRefName2)
	assert.NoError(t, cerr)
	assert.Equal(t, 2, nr)

	connection.Close()

	AddGlobalMapRepositoryReference(adabasModDBIDs + ",250")
	a, _ := NewAdabas(1)
	defer DelGlobalMapRepository(a, 250)

	connection, err = NewConnection("acj;map")
	if !assert.NoError(t, err) {
		return
	}
	deleteRequest, derr := connection.CreateMapDeleteRequest(vehicleMapName)
	if !assert.NoError(t, derr) {
		return
	}

	err = deleteRequest.Delete(adatypes.Isn(lastIsn))
	assert.NoError(t, derr)
	connection.Close()
}

func ExampleConnection_mapStore() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: ExampleConnection_mapStore")

	if cErr := clearFile(16); cErr != nil {
		return
	}
	if cErr := clearFile(19); cErr != nil {
		fmt.Println("Error clearing 19", cErr)
		return
	}

	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	if perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository); perr != nil {
		fmt.Println("Error creating map", massLoadSystransStore, perr)
		return
	}
	dataRepository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 19}
	vehicleMapName := mapVehicles + "Go"
	if perr := prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository); perr != nil {
		return
	}

	adatypes.Central.Log.Infof("Create connection...")
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if err != nil {
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest16, rErr := connection.CreateMapStoreRequest(massLoadSystransStore)
	if rErr != nil {
		return
	}
	storeRequest16.StoreFields("PERSONNEL-ID,NAME")
	record, err := storeRequest16.CreateRecord()
	if err != nil {
		fmt.Println("Error create record", err)
		return
	}
	_ = record.SetValueWithIndex("PERSONNEL-ID", nil, "26555_0")
	_ = record.SetValueWithIndex("NAME", nil, "WABER")
	_ = record.SetValueWithIndex("FIRST-NAME", nil, "EMIL")
	_ = record.SetValueWithIndex("MIDDLE-I", nil, "MERK")
	err = storeRequest16.Store(record)
	if err != nil {
		fmt.Println("Error store record", err)
		return
	}
	storeRequest19, rErr := connection.CreateMapStoreRequest(vehicleMapName)
	if rErr != nil {
		fmt.Println("Error create store request vehicle", rErr)
		return
	}
	err = storeRequest19.StoreFields("REG-NUM,PERSONNEL-ID,CAR-DETAILS")
	if err != nil {
		fmt.Println("Error store fields", err)
		return
	}

	record, err = storeRequest19.CreateRecord()
	if err != nil {
		fmt.Println("Create record", err)
		return
	}
	err = record.SetValueWithIndex("REG-NUM", nil, "29555_0")
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	err = record.SetValueWithIndex("PERSONNEL-ID", nil, "WABER")
	if err != nil {
		fmt.Println("Error search in "+vehicleMapName, err)
		return
	}
	err = record.SetValueWithIndex("MAKE", nil, "EMIL")
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	err = record.SetValueWithIndex("MODEL", nil, "MERK")
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	err = storeRequest19.Store(record)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	err = connection.EndTransaction()
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	fmt.Println("Read file ..." + massLoadSystransStore)
	err = dumpMapStoredData(adabasModDBIDs, massLoadSystransStore, "26555")
	if err != nil {
		fmt.Println("Error reading "+massLoadSystransStore, err)
		return
	}
	fmt.Println("Read file ..." + vehicleMapName)
	err = dumpMapStoredData(adabasModDBIDs, vehicleMapName, "29555")
	if err != nil {
		fmt.Println("Error reading "+vehicleMapName, err)
		return
	}

	// Output: Clear file  16
	// Success clearing file  16
	// Clear file  19
	// Success clearing file  19
	// Read file ...EMPLDDM-GOLOAD-STORE
	// Dump all result values
	// Record Isn: 0001
	//   PERSONNEL-ID = > 26555_0  <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = >                      <
	//    NAME = > WABER                <
	//    MIDDLE-I = >            <
	// Read file ...VEHICLESGo
	// Dump all result values
	// Record Isn: 0001
	//   REG-NUM = > 29555_0         <
	//   PERSONNEL-ID = > WABER    <
	//   CAR-DETAILS = [ 1 ]
	//    MAKE = > EMIL                 <
	//    MODEL = > MERK                 <
	//    COLOR = >            <

}

func dumpMapStoredData(target string, mapName string, search string) error {
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if err != nil {
		return err
	}
	defer connection.Close()
	readRequest, rrerr := connection.CreateMapReadRequest(mapName)
	if rrerr != nil {
		return rrerr
	}
	fields := "PERSONNEL-ID,FULL-NAME"
	searchField := "PERSONNEL-ID"

	switch mapName {
	case mapVehicles:
		fields = "AA,CD"
		searchField = "AA"
	case mapVehicles + "Go":
		fields = "REG-NUM,PERSONNEL-ID,CAR-DETAILS"
		searchField = "REG-NUM"
	}
	err = readRequest.QueryFields(fields)
	if err != nil {
		return err
	}
	result, rerr := readRequest.ReadLogicalWith(searchField + "=[" + search + "_:" + search + "_Z]")
	if rerr != nil {
		return rerr
	}
	for i, record := range result.Values {
		record.Isn = adatypes.Isn(i + 1)
	}
	result.DumpValues()
	return nil
}

func ExampleConnection_readShortMap() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: func ExampleConnection_readShortMap()")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLSHORT")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("*")
	if err != nil {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadLogicalWith("ID=[11100314:11100317]")
	if err != nil {
		fmt.Println("Error read logical", err)
		return
	}
	result.DumpValues()
	// Output:Read logigcal data:
	// Dump all result values
	// Record Isn: 0393
	//   ID = > 11100314 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > WOLFGANG             <
	//    NAME = > SCHMIDT              <
	//    MIDDLE-NAME = > MARIA                <
	// Record Isn: 0261
	//   ID = > 11100315 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > GLORIA               <
	//    NAME = > MERTEN               <
	//    MIDDLE-NAME = > ELISABETH            <
	// Record Isn: 0262
	//   ID = > 11100316 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > HEINZ                <
	//    NAME = > RAMSER               <
	//    MIDDLE-NAME = > EWALD                <
	// Record Isn: 0263
	//   ID = > 11100317 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > ALFONS               <
	//    NAME = > DORSCH               <
	//    MIDDLE-NAME = > FRITZ                <
}

func ExampleConnection_readLongMapIsn() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: ExampleAdabas_readFileDefinitionMap")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("*")
	if err != nil {
		fmt.Println("Error query fields", err)
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadISN(1)
	if err != nil {
		fmt.Println("Error search value", err)
		return
	}
	for _, v := range result.Values {
		f, e := v.SearchValue("creation_time")
		if e != nil {
			fmt.Println("Error search value", e)
			return
		}
		f.SetValue(0)
		f, e = v.SearchValue("Last_Updates[01]")
		if e != nil || f == nil {
			fmt.Println(e)
			return
		}
		f.SetValue(0)
	}

	fmt.Println(result.String())
	// Output:Read logigcal data:
	// Record Isn: 0001
	//   personnel-data = [ 1 ]
	//    personnel-id = > 50005800 <
	//    id-data = [ 1 ]
	//     personnel-no_-UQ_taken- = > 0 <
	//     id-card = > 0 <
	//     signature = >  <
	//   full-name = [ 1 ]
	//    first-name = > Simone <
	//    middle-name = >   <
	//    name = > Adam <
	//   mar-stat = > M <
	//   sex = > F <
	//   birth = > 718460 <
	//   private-address = [ 1 ]
	//    address-line[01] = [ 1 ]
	//     address-line[01,01] = > 26 Avenue Rhin Et Da <
	//    city[01] = > Joigny <
	//    post-code[01] = > 89300 <
	//    country[01] = > F <
	//    phone-email[01] = [ 1 ]
	//     area-code[01] = > 1033 <
	//     private-phone[01] = > 44864858 <
	//     private-fax[01] = >   <
	//     private-mobile[01] = >   <
	//     private-email[01] = [ 0 ]
	//   business-address = [ 0 ]
	//   dept = > VENT59 <
	//   job-title = > Chef de Service <
	//   income = [ 1 ]
	//    curr-code[01] = > EUR <
	//    salary_P9.2[01] = > 963 <
	//    bonus_P9.2[01] = [ 1 ]
	//     bonus_P9.2[01,01] = > 138 <
	//   total_income_-EUR- = > 0.000000 <
	//   leave-date = [ 1 ]
	//    leave-due = > 19 <
	//    leave-taken_N2.1 = > 5 <
	//   leave-booked = [ 1 ]
	//    leave-start[01] = > 20070801 <
	//    leave-end[01] = > 20070831 <
	//   language = [ 2 ]
	//    language[01] = > FRE <
	//    language[02] = > ENG <
	//   last_update_--TIMX- = > 0 <
	//   picture = >  <
	//   documents = [ 0 ]
	//   creation_time = > 0 <
	//   Last_Updates = [ 1 ]
	//    Last_Updates[01] = > 0 <
}

func ExampleConnection_readLongMapRange() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: ExampleConnection_readLongMapRange")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("*")
	if err != nil {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadLogicalWith("personnel-id=[50005800:50005801]")
	if err != nil {
		fmt.Println("Read error", err)
		return
	}
	for _, v := range result.Values {
		f, e := v.SearchValue("creation_time")
		if e != nil {
			fmt.Println("Search error", e)
			return
		}
		f.SetValue(0)
		f, e = v.SearchValue("Last_Updates[01]")
		if e != nil || f == nil {
			fmt.Println(e)
			return
		}
		f.SetValue(0)
	}
	result.DumpValues()
	// Output:Read logigcal data:
	// Dump all result values
	// Record Isn: 0001
	//   personnel-data = [ 1 ]
	//    personnel-id = > 50005800 <
	//    id-data = [ 1 ]
	//     personnel-no_-UQ_taken- = > 0 <
	//     id-card = > 0 <
	//     signature = >  <
	//   full-name = [ 1 ]
	//    first-name = > Simone <
	//    middle-name = >   <
	//    name = > Adam <
	//   mar-stat = > M <
	//   sex = > F <
	//   birth = > 718460 <
	//   private-address = [ 1 ]
	//    address-line[01] = [ 1 ]
	//     address-line[01,01] = > 26 Avenue Rhin Et Da <
	//    city[01] = > Joigny <
	//    post-code[01] = > 89300 <
	//    country[01] = > F <
	//    phone-email[01] = [ 1 ]
	//     area-code[01] = > 1033 <
	//     private-phone[01] = > 44864858 <
	//     private-fax[01] = >   <
	//     private-mobile[01] = >   <
	//     private-email[01] = [ 0 ]
	//   business-address = [ 0 ]
	//   dept = > VENT59 <
	//   job-title = > Chef de Service <
	//   income = [ 1 ]
	//    curr-code[01] = > EUR <
	//    salary_P9.2[01] = > 963 <
	//    bonus_P9.2[01] = [ 1 ]
	//     bonus_P9.2[01,01] = > 138 <
	//   total_income_-EUR- = > 0.000000 <
	//   leave-date = [ 1 ]
	//    leave-due = > 19 <
	//    leave-taken_N2.1 = > 5 <
	//   leave-booked = [ 1 ]
	//    leave-start[01] = > 20070801 <
	//    leave-end[01] = > 20070831 <
	//   language = [ 2 ]
	//    language[01] = > FRE <
	//    language[02] = > ENG <
	//   last_update_--TIMX- = > 0 <
	//   picture = >  <
	//   documents = [ 0 ]
	//   creation_time = > 0 <
	//   Last_Updates = [ 1 ]
	//    Last_Updates[01] = > 0 <
	// Record Isn: 1251
	//   personnel-data = [ 1 ]
	//    personnel-id = > 50005801 <
	//    id-data = [ 1 ]
	//     personnel-no_-UQ_taken- = > 0 <
	//     id-card = > 0 <
	//     signature = >  <
	//   full-name = [ 1 ]
	//    first-name = > वासुदेव <
	//    middle-name = > मूर्ती <
	//    name = > कुमार <
	//   mar-stat = > M <
	//   sex = > M <
	//   birth = > 721484 <
	//   private-address = [ 1 ]
	//    address-line[01] = [ 1 ]
	//     address-line[01,01] = > ह-1,दिशा स्क्यलैइन म <
	//    city[01] = > नोयडा <
	//    post-code[01] = > 201301 <
	//    country[01] = > IND <
	//    phone-email[01] = [ 1 ]
	//     area-code[01] = > 01189 <
	//     private-phone[01] = > 233449 <
	//     private-fax[01] = >   <
	//     private-mobile[01] = >   <
	//     private-email[01] = [ 0 ]
	//   business-address = [ 0 ]
	//   dept = > COMP02 <
	//   job-title = > सीनियर प्रोग्रामर <
	//   income = [ 1 ]
	//    curr-code[01] = > INR <
	//    salary_P9.2[01] = > 45000 <
	//    bonus_P9.2[01] = [ 5 ]
	//     bonus_P9.2[01,01] = > 5000 <
	//     bonus_P9.2[01,02] = > 5000 <
	//     bonus_P9.2[01,03] = > 5000 <
	//     bonus_P9.2[01,04] = > 5000 <
	//     bonus_P9.2[01,05] = > 5000 <
	//   total_income_-EUR- = > 0.000000 <
	//   leave-date = [ 1 ]
	//    leave-due = > 8 <
	//    leave-taken_N2.1 = > 7 <
	//   leave-booked = [ 1 ]
	//    leave-start[01] = > 20060915 <
	//    leave-end[01] = > 20060922 <
	//   language = [ 2 ]
	//    language[01] = > HIN <
	//    language[02] = > ENG <
	//   last_update_--TIMX- = > 0 <
	//   picture = >  <
	//   documents = [ 0 ]
	//   creation_time = > 0 <
	//   Last_Updates = [ 1 ]
	//    Last_Updates[01] = > 0 <
}

func TestConnection_readAllMap(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadISN(1)
	if !assert.NoError(t, err) {
		return
	}
	//result.DumpValues()
	if assert.Equal(t, 1, len(result.Values)) {
		v, verr := result.Values[0].SearchValue("last_update_--TIMX-")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "0", v.String())
	}

}

func TestConnection_readReference(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("LOBPICTURE")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("Filename,@Thumbnail")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadISN(1)
	if !assert.NoError(t, err) {
		return
	}
	// result.DumpValues()
	if assert.Equal(t, 1, len(result.Values)) {
		v, verr := result.Values[0].SearchValue("Filename")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "106-0670_IMG.JPG", v.String())
		v, verr = result.Values[0].SearchValue("@Thumbnail")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "/image/map/LOBPICTURE/1/Thumbnail", v.String())
	}

	result, err = request.ReadISN(2)
	if !assert.NoError(t, err) {
		return
	}
	// result.DumpValues()
	if assert.Equal(t, 1, len(result.Values)) {
		v, verr := result.Values[0].SearchValue("Filename")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "p1.jpg", v.String())
		v, verr = result.Values[0].SearchValue("@Thumbnail")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "/image/map/LOBPICTURE/2/Thumbnail", v.String())
	}

}

func TestConnection_readReferenceList(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("LOBPICTURE")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("Filename,@Thumbnail")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 3
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadLogicalBy("Filename")
	if !assert.NoError(t, err) {
		return
	}
	// result.DumpValues()
	if assert.Equal(t, 3, len(result.Values)) {
		v, verr := result.Values[0].SearchValue("Filename")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "106-0670_IMG.JPG", v.String())
		v, verr = result.Values[0].SearchValue("@Thumbnail")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "/image/map/LOBPICTURE/1/Thumbnail", v.String())
		v, verr = result.Values[1].SearchValue("Filename")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "DSCF3544_2.JPG", v.String())
		v, verr = result.Values[1].SearchValue("@Thumbnail")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "/image/map/LOBPICTURE/27/Thumbnail", v.String())
		v, verr = result.Values[2].SearchValue("Filename")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "DSCN0529.JPG", v.String())
		v, verr = result.Values[2].SearchValue("@Thumbnail")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "/image/map/LOBPICTURE/26/Thumbnail", v.String())
	}

}

func ExampleConnection_mapReadUnicode() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: ExampleConnection_mapReadUnicode")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("PERSONNEL-ID,FULL-NAME")
	if err != nil {
		return
	}
	request.Start = 1025
	request.Limit = 3
	var result *Response
	fmt.Println("Read using ISN order:")
	result, err = request.ReadByISN()
	if err != nil {
		fmt.Println("Error reading ISN order", err)
		return
	}
	result.DumpValues()

	// Output: Read using ISN order:
	// Dump all result values
	// Record Isn: 1025
	//   PERSONNEL-ID = > 30021215 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > SYLVIA               <
	//    NAME = > BURTON               <
	//    MIDDLE-I = > J <
	// Record Isn: 1026
	//   PERSONNEL-ID = > 30021311 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > GERARD               <
	//    NAME = > JOHNSTONE            <
	//    MIDDLE-I = > E <
	// Record Isn: 1027
	//   PERSONNEL-ID = > 30021312 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > NORMA                <
	//    NAME = > FRANCIS              <
	//    MIDDLE-I = >   <
}

func ExampleConnection_mapReadUnicodeNew() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: ExampleConnection_mapReadUnicodeNew")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("personnel-id,full-name")
	if err != nil {
		return
	}
	request.Start = 1250
	request.Limit = 3
	var result *Response
	fmt.Println("Read using ISN order:")
	result, err = request.ReadByISN()
	if err != nil {
		fmt.Println("Error reading ISN order", err)
		return
	}
	result.DumpValues()

	// Output: Read using ISN order:
	// Dump all result values
	// Record Isn: 1250
	//   personnel-data = [ 1 ]
	//    personnel-id = > 73002200 <
	//   full-name = [ 1 ]
	//    first-name = > Игорь <
	//    middle-name = > Петрович <
	//    name = > Михайлов <
	// Record Isn: 1251
	//   personnel-data = [ 1 ]
	//    personnel-id = > 50005801 <
	//   full-name = [ 1 ]
	//    first-name = > वासुदेव <
	//    middle-name = > मूर्ती <
	//    name = > कुमार <
	// Record Isn: 1252
	//   personnel-data = [ 1 ]
	//    personnel-id = > 50005501 <
	//   full-name = [ 1 ]
	//    first-name = > विनोद <
	//    middle-name = > अभगे <
	//    name = > अरविद <
}

func TestConnection_readGroup(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("FULL-ADDRESS")
	if !assert.NoError(t, err) {
		return
	}
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadISN(1)
	if !assert.NoError(t, err) {
		return
	}
	// result.DumpValues()
	v, verr := result.Values[0].SearchValue("CITY")
	if !assert.NoError(t, verr) {
		return
	}
	assert.Equal(t, "JOIGNY              ", v.String())
	v, verr = result.Values[0].SearchValue("ZIP")
	if !assert.NoError(t, verr) {
		return
	}
	assert.Equal(t, "89300     ", v.String())
	v, verr = result.Values[0].SearchValue("COUNTRY")
	if !assert.NoError(t, verr) {
		return
	}
	assert.Equal(t, "F  ", v.String())

}

func ExampleConnection_mapReadDisjunctSearch() {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	adatypes.Central.Log.Infof("TEST: ExampleConnection_mapReadDisjunctSearch")

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("personnel-id")
	if err != nil {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read using ISN order:")
	result, err = request.ReadLogicalWith("name=SMITH")
	if err != nil {
		fmt.Println("Error reading ISN order", err)
		return
	}
	result.DumpValues()

	// Output: Read using ISN order:
	// Dump all result values
	// Record Isn: 0579
	//   personnel-data = [ 1 ]
	//    personnel-id = > 20009300 <
	// Record Isn: 0634
	//   personnel-data = [ 1 ]
	//    personnel-id = > 20015400 <
	// Record Isn: 0670
	//   personnel-data = [ 1 ]
	//    personnel-id = > 20018800 <
	// Record Isn: 0727
	//   personnel-data = [ 1 ]
	//    personnel-id = > 20025200 <
	// Record Isn: 0787
	//   personnel-data = [ 1 ]
	//    personnel-id = > 20000400 <
}

func TestConnectionAndSearchMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("FULL-NAME")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("NAME>'ADAM' AND NAME<'AECKERLE'")
		assert.NoError(t, err)
		validateResult(t, "osandsearch", result)
	}

}

func TestConnectionOsMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("*")
		assert.NoError(t, err)
		request.Limit = 20
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalBy("PERSONNEL-ID")
		if !assert.NoError(t, err) {
			return
		}
		result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 20, len(result.Values)) {
			ae := result.Values[1].HashFields["NAME"]
			fmt.Println("Check SCHIRM ...")
			assert.Equal(t, "SCHIRM", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
			ae = result.Values[19].HashFields["NAME"]
			fmt.Println("Check BLAU ...")
			assert.Equal(t, "BLAU", strings.TrimSpace(ae.String()))
			validateResult(t, "osallread", result)
		}
	}

}

func TestConnection_cyrillicMap(t *testing.T) {
	err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("Cyrilic2")
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	fmt.Println("Read logigcal data:")
	result, err = request.ReadISN(1)
	if !assert.NoError(t, err) {
		return
	}
	result.DumpValues()
	if assert.Equal(t, 1, len(result.Values)) {
		v, verr := result.Values[0].SearchValue("cyrilic")
		if !assert.NoError(t, verr) {
			return
		}
		assert.Equal(t, "Покупатели", v.String())
	}

}

func TestConnectionMapCharset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("Cyrilic2")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("*")
		assert.NoError(t, err)
		request.Limit = 20
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalBy("ascii")
		if !assert.NoError(t, err) {
			return
		}
		result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 20, len(result.Values)) {
			ae := result.Values[1].HashFields["cyrilic"]
			assert.Equal(t, "АлтайГАЗавтосервис", strings.TrimSpace(ae.String()))
			aa := result.Values[1].HashFields["ascii"]
			assert.Equal(t, "10", strings.TrimSpace(aa.String()))
			ae = result.Values[19].HashFields["cyrilic"]
			assert.Equal(t, "XXI Век-Авто", strings.TrimSpace(ae.String()))
			aa = result.Values[19].HashFields["ascii"]
			assert.Equal(t, "10H", strings.TrimSpace(aa.String()))
			validateResult(t, "cyrilic2", result)
		}
	}

}

func TestConnectionMapFractional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("Fractional")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("*")
		assert.NoError(t, err)
		request.Limit = 7
		fmt.Println("Read logigcal data:")
		result, err := request.ReadByISN()
		if !assert.NoError(t, err) {
			return
		}
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 7, len(result.Values)) {
			ae := result.Values[1].HashFields["ALPHA10"]
			assert.Equal(t, "ABC10", strings.TrimSpace(ae.String()))
			aa := result.Values[1].HashFields["ALPHA3"]
			assert.Equal(t, "ABC", strings.TrimSpace(aa.String()))
			aa = result.Values[1].HashFields["FRACT1"]
			assert.Equal(t, "10.44", strings.TrimSpace(aa.String()))
			aa = result.Values[1].HashFields["FLOAT8"]
			assert.Equal(t, "10.200000", strings.TrimSpace(aa.String()))
			ae = result.Values[0].HashFields["ALPHA10"]
			assert.Equal(t, "ABC1", strings.TrimSpace(ae.String()))
			aa = result.Values[0].HashFields["FLOAT8"]
			assert.Equal(t, "1.200000", strings.TrimSpace(aa.String()))
			ae = result.Values[6].HashFields["ALPHA10"]
			assert.Equal(t, "ABC1000000", strings.TrimSpace(ae.String()))
			aa = result.Values[6].HashFields["ALPHA3"]
			assert.Equal(t, "ABC", strings.TrimSpace(aa.String()))
			aa = result.Values[6].HashFields["FRACT1"]
			assert.Equal(t, "1000000.44", strings.TrimSpace(aa.String()))
			aa = result.Values[6].HashFields["FLOAT8"]
			assert.Equal(t, "1000000.200000", strings.TrimSpace(aa.String()))
			validateResult(t, "fractional", result)
		}
	}

}

func TestConnectionMUsystemField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[23,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("ADABAS_MAP")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("MAPPING")
		assert.NoError(t, err)
		request.Limit = 10
		request.Multifetch = 1
		fmt.Println("Read logigcal data:")
		result, err := request.ReadByISN()
		if !assert.NoError(t, err) {
			return
		}
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 7, len(result.Values)) {
			validateResult(t, "ADABAS_MAP", result)
		}
	}

}
