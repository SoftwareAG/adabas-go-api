package adabas

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkConnection_cached(b *testing.B) {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

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
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if cerr != nil {
		return cerr
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
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
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[24,4]")
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
		result := &Response{}
		err = request.ReadLogicalByWithParser("NAME", nil, result)
		if !assert.NoError(b, err) {
			return
		}
		if !assert.Equal(b, 1107, len(result.Values)) {
			return
		}
	}
}

func TestAuth(t *testing.T) {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//connection, cerr := NewConnection("acj;map;config=[177(adatcp://pinas:60177),4]")
	connection, cerr := NewConnection("acj;target=24;auth=NONE,user=TestAuth,id=4,host=xx")
	if !assert.NoError(t, cerr) {
		return
	}
	assert.Contains(t, connection.ID.String(), "xx      :TestAuth [4] ")
	connection.Close()

	connection, cerr = NewConnection("acj;target=24;auth=NONE,user=ABCDEFGHIJ,id=65535,host=KLMNOPQRSTUVWXYZ")
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
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//connection, cerr := NewConnection("acj;map;config=[177(adatcp://pinas:60177),4]")
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
		result := &Response{}
		err = request.ReadLogicalByWithParser("NAME", nil, result)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, 1107, len(result.Values)) {
			return
		}
	}
}

func BenchmarkConnection_noreconnectremote(b *testing.B) {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

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
		result := &Response{}
		err = request.ReadLogicalByWithParser("NAME", nil, result)
		if !assert.NoError(b, err) {
			return
		}
		if !assert.Equal(b, 1107, len(result.Values)) {
			return
		}
	}
}

func TestConnectionWithMultipleMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4]")
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
		result := &Response{}
		fmt.Println("Read logigcal data:")
		err := request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100301:11100303]", nil, result)
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
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
		result := &Response{}
		fmt.Println("Read logigcal data:")
		err := request.ReadLogicalWithWithParser("Vendor=RENAULT", nil, result)
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
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
	f := initTestLogWithFile(t, "connection_map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4];auth=NONE,user=TCMapPoin,id=4,host=REMOTE")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection : ", connection)
	request, err := connection.CreateMapReadRequest("REMOTEEMPL")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		result := &Response{}
		fmt.Println("Read logigcal data:")
		err := request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100301:11100303]", nil, result)
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
		if assert.Equal(t, 3, len(result.Values)) {
			ae := result.Values[1].HashFields["NAME"]
			assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
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
	f := initTestLogWithFile(t, "connection_map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection : ", connection)
	store, err := connection.CreateMapStoreRequest("COPYEMPL")
	if !assert.NoError(t, err) {
		return
	}
	store.StoreFields("NAME,PERSONNEL-ID")
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, rerr) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		result := &Response{}
		fmt.Println("Read logigcal data:")
		err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100000:11101000]", copyData, store)
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
		if !assert.Equal(t, 0, len(result.Values)) {
			return
		}
	}
	err = store.EndTransaction()
	assert.NoError(t, err)
}

func ExampleConnection_readWithMap() {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	log.Infof("TEST: ExampleAdabas_readFileDefinitionMap")

	connection, cerr := NewConnection("acj;map;config=[24,4]")
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
	result := &Response{}
	fmt.Println("Read logigcal data:")
	err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100314:11100317]", nil, result)
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

func ExampleConnection_readFileDefinitionMapGroup() {
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[24,4]")
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
	result := &Response{}
	fmt.Println("Read logigcal data:")
	err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100315:11100316]", nil, result)
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
	f, err := initLogWithFile("connection_map.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

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
	storeRecord16.DumpValues()
	fmt.Println("Stored Employees request")
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
	storeRecord.DumpValues()
	fmt.Println("Store Vehicle request")
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return err
	}

	return nil
}

const multipleTransactionRefName = "M16555"
const multipleTransactionRefName2 = "M19555"

func TestConnectionSimpleMultipleMapStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}

	log.Infof("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	dataRepository = &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 19}
	vehicleMapName := mapVehicles + "Go"
	perr = prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}

	log.Infof("Create connection...")
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
	err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_0")
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
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_0")
	if !assert.NoError(t, err) {
		return
	}
	for i := 1; i < 10; i++ {
		x := strconv.Itoa(i)
		err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_"+x)
		if !assert.NoError(t, err) {
			return
		}

	}
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_1")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("End transaction")
	connection.EndTransaction()
	fmt.Println("Check stored data")

	log.Infof("Check stored data")
	checkStoreByFile(t, adabasModDBIDs, 16, multipleTransactionRefName)
	checkStoreByFile(t, adabasModDBIDs, 19, multipleTransactionRefName2)

}

func ExampleConnection_mapStore() {
	f, err := initLogWithFile("connection.log")
	if err != nil {
		return
	}
	defer f.Close()

	if cErr := clearFile(16); cErr != nil {
		return
	}
	if cErr := clearFile(19); cErr != nil {
		return
	}

	dataRepository := &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 16}
	if perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository); perr != nil {
		fmt.Println("Error creating map", massLoadSystransStore, perr)
		return
	}
	dataRepository = &DatabaseURL{URL: *newURLWithDbid(adabasModDBID), Fnr: 19}
	vehicleMapName := mapVehicles + "Go"
	if perr := prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository); perr != nil {
		return
	}

	log.Infof("Create connection...")
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if err != nil {
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest16, rErr := connection.CreateMapStoreRequest(massLoadSystransStore)
	if err != nil {
		return
	}
	storeRequest16.StoreFields("PERSONNEL-ID,NAME")
	record, err := storeRequest16.CreateRecord()
	err = record.SetValueWithIndex("PERSONNEL-ID", nil, "26555_0")
	err = record.SetValueWithIndex("NAME", nil, "WABER")
	err = record.SetValueWithIndex("FIRST-NAME", nil, "EMIL")
	err = record.SetValueWithIndex("MIDDLE-I", nil, "MERK")
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

	// Output: Read file ...EMPLDDM-GOLOAD-STORE
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
