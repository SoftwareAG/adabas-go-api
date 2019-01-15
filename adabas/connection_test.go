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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func initTestLogWithFile(t *testing.T, fileName string) *os.File {
	file, err := initLogWithFile(fileName)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
		return nil
	}
	return file
}

func entireNetworkLocation() string {
	network := os.Getenv("WCPHOST")
	if network == "" {
		return "localhost:50001"
	}
	return network
}

func adabasTCPLocation() string {
	network := os.Getenv("ADATCPHOST")
	if network == "" {
		return "localhost:60001"
	}
	return network
}

func initLogWithFile(fileName string) (file *os.File, err error) {
	level := log.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	switch ed {
	case "1":
		level = log.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	case "2":
		level = log.InfoLevel
	default:
		level = log.ErrorLevel
	}
	return initLogLevelWithFile(fileName, level)
}

func initLogLevelWithFile(fileName string, level log.Level) (file *os.File, err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	name := p + "/" + fileName
	file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	log.SetLevel(level)

	log.SetOutput(file)
	myLog := log.New()
	myLog.SetLevel(level)
	myLog.Out = file

	myLog.Infof("Set debug level to %s", level)

	// log.SetOutput(file)
	adatypes.Central.Log = myLog

	return
}

type parseTestStructure struct {
	storeRequest *StoreRequest
	fields       string
	t            *testing.T
}

func parseTestConnection(adabasRequest *adatypes.AdabasRequest, x interface{}) (err error) {
	fmt.Println("Parse Test connection")
	parseTestStructure := x.(parseTestStructure)
	if parseTestStructure.t == nil {
		panic("Parse test structure empty test instance")
	}
	if !assert.NotNil(parseTestStructure.t, adabasRequest.Definition.Values) {
		log.Debugf("Parse Buffer .... values avail.=%v", (adabasRequest.Definition.Values == nil))
		return fmt.Errorf("Data value empty")
	}
	storeRequest := parseTestStructure.storeRequest
	dErr := storeRequest.StoreFields(parseTestStructure.fields)
	if !assert.NoError(parseTestStructure.t, dErr) {
		return
	}

	storeRecord, sErr := storeRequest.CreateRecord()
	assert.NoError(parseTestStructure.t, sErr)
	if sErr != nil {
		err = sErr
		fmt.Println("Store record error ...", err)
		return
	}
	fmt.Println("Found ISN: ", adabasRequest.Isn, " len=", len(adabasRequest.Definition.Values))
	if !assert.NotNil(parseTestStructure.t, adabasRequest.Definition.Values) {
		return
	}
	storeRecord.Value = adabasRequest.Definition.Values
	for _, f := range strings.Split(parseTestStructure.fields, ",") {
		if _, ok := storeRecord.HashFields[f]; !ok {
			err = adatypes.NewGenericError(47, f)
			return
		}
	}
	fmt.Println("Store record:")
	storeRecord.DumpValues()
	//log.Println("Store record =====================================")
	err = storeRequest.Store(storeRecord)
	fmt.Println("ISN: ", storeRecord.Isn, " -> ", err)
	return
}

func deleteRecords(adabasRequest *adatypes.AdabasRequest, x interface{}) (err error) {
	deleteRequest := x.(*DeleteRequest)
	// fmt.Printf("Delete ISN: %d on %s/%d\n", adabasRequest.Isn, deleteRequest.repository.URL.String(), deleteRequest.repository.Fnr)
	err = deleteRequest.Delete(adabasRequest.Isn)
	return
}

func TestConnectionSimpleTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	assert.NoError(t, dErr)
	deleteRequest.EndTransaction()

	request, rErr2 := connection.CreateReadRequest(11)
	if !assert.NoError(t, rErr2) {
		return
	}
	err = request.loadDefinition()
	if !assert.NoError(t, err) {
		return
	}

	log.Debug("Loaded Definition in Tests")
	request.definition.DumpTypes(false, false)

	storeRequest, sErr := connection.CreateStoreRequest(16)
	if !assert.NoError(t, sErr) {
		return
	}

	parseTestStructure := parseTestStructure{storeRequest: storeRequest, t: t, fields: "AA,AC,AD,AE"}
	request.QueryFields(parseTestStructure.fields)
	assert.NotNil(t, request.definition)
	request.Limit = 3
	fmt.Println("Result data:")
	err = request.ReadPhysicalSequenceWithParser(parseTestConnection, parseTestStructure)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()
}

func TestConnectionOpenOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	err = connection.Open()
	if !assert.NoError(t, err) {
		return
	}
	err = connection.Open()
	if !assert.NoError(t, err) {
		return
	}
	err = connection.Release()
	if !assert.NoError(t, err) {
		return
	}

}

func TestConnectionOpenFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=222")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	err = connection.Open()
	assert.Error(t, err)
}

func TestConnectionMultipleFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(16)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	deleteRequest.EndTransaction()

	request, rErr2 := connection.CreateReadRequest(11)
	assert.NoError(t, rErr2)
	storeRequest, sErr := connection.CreateStoreRequest(16)
	assert.NoError(t, sErr)
	parseTestStructure := parseTestStructure{storeRequest: storeRequest, t: t, fields: "AA,AC,AD,AE,AZ"}
	request.QueryFields(parseTestStructure.fields)
	request.Limit = 3
	fmt.Println("Read physical")
	parseTestStructure.t = t
	err = request.ReadPhysicalSequenceWithParser(parseTestConnection, parseTestStructure)
	assert.NoError(t, err)
	fmt.Println("End transaction")
	storeRequest.EndTransaction()
}

func TestConnectionStorePeriodFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	fmt.Println("Delete done, call end of transaction")
	log.Debug("Delete done, call end of transaction")
	deleteRequest.EndTransaction()

	fmt.Println("Call Read to 11")
	request, rErr2 := connection.CreateReadRequest(11)
	assert.NoError(t, rErr2)
	fmt.Println("Call Store to 16")
	storeRequest, sErr := connection.CreateStoreRequest(16)
	assert.NoError(t, sErr)
	fmt.Println("Parse test structure")
	parseTestStructure := parseTestStructure{storeRequest: storeRequest, t: t, fields: "AA,AC,AD,AE,AW"}
	request.QueryFields(parseTestStructure.fields)
	fmt.Println("Result data:")
	parseTestStructure.t = t
	adatypes.Central.Log.Debugf("Test Read logical with ...")
	err = request.ReadLogicalWithWithParser("AA=[11100301:11100305]", parseTestConnection, parseTestStructure)
	fmt.Println("Read logical done")
	assert.NoError(t, err)
	storeRequest.EndTransaction()
}

func TestConnectionMultifetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, connection) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 10

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	result := &RequestResult{}
	err = readRequest.ReadPhysicalSequenceWithParser(nil, result)
	assert.NoError(t, err)
	// result.DumpValues()
	assert.Equal(t, 1107, len(result.Values))
}

func TestConnectionPeriodAndMultipleField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(11)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.Limit = 0

	qErr := readRequest.QueryFields("AA,AQ,AZ")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	result, rErr := readRequest.ReadISN(499)
	if !assert.NoError(t, rErr) {
		return
	}
	result.DumpValues()
}

func TestConnectionRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.Error(t, openErr)
	assert.Equal(t, "Entire Network client not supported, use port 0 and Entire Network native access", openErr.Error())
	// if assert.NoError(t, openErr) {
	// 	request, err := connection.CreateReadRequest(11)
	// 	assert.NoError(t, err)

	// 	request.QueryFields("AA,AC,AD,AE,AH,AV")
	// 	request.Limit = 0
	// 	result := &RequestResult{}
	// 	err = request.ReadLogicalWithWithParser("AA=[11100301:11100303]", nil, result)
	// 	assert.NoError(t, err)
	// 	fmt.Println("Result data:")
	// 	//result.DumpValues()
	// 	assert.Equal(t, 3, len(result.Values))
	// 	ae := result.Values[1].HashFields["AE"]
	// 	assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
	// 	ei64, xErr := ae.Int64()
	// 	assert.Error(t, xErr, "Error should be send if value is string")
	// 	assert.Equal(t, int64(0), ei64)
	// 	ah := result.Values[1].HashFields["AH"]
	// 	assert.Equal(t, "713196", strings.TrimSpace(ah.String()))
	// 	var i64 int64
	// 	var ui64 uint64
	// 	var i32 int32
	// 	var ui32 uint32
	// 	i64, err = ah.Int64()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, int64(713196), i64)
	// 	ui64, err = ah.UInt64()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, uint64(713196), ui64)
	// 	i32, err = ah.Int32()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, int32(713196), i32)
	// 	ui32, err = ah.UInt32()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, uint32(713196), ui32)
	// 	raw := ah.Bytes()
	// 	assert.Equal(t, []byte{0x7, 0x13, 0x19, 0x6c}, raw)

	// 	av := result.Values[2].HashFields["AV"]
	// 	assert.Equal(t, "3", strings.TrimSpace(av.String()))
	// 	i64, err = av.Int64()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, int64(3), i64)
	// 	ui64, err = av.UInt64()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, uint64(3), ui64)
	// 	i32, err = av.Int32()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, int32(3), i32)
	// 	ui32, err = av.UInt32()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, uint32(3), ui32)
	// 	raw = av.Bytes()
	// 	assert.Equal(t, []byte{0x30, 0x33}, raw)
	// }

}

func TestConnectionWithMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;map;config=[24,4]")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection : ", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err) {
		fmt.Println("Connection : ", connection)
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		result := &RequestResult{}
		fmt.Println("Read logigcal data:")
		err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100301:11100303]", nil, result)
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

func TestConnectionAllMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	log.Debug("Created connection : ", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		result := &RequestResult{}
		fmt.Println("Read logigcal data:")
		err := request.ReadPhysicalSequenceWithParser(nil, result)
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 1107, len(result.Values)) {
			ae := result.Values[1].HashFields["NAME"]
			fmt.Println("Check MORENO ...")
			assert.Equal(t, "MORENO", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}

}

func ExampleReadRequest_file() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=23")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("AA=[11100301:11100303]", nil, result)
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0251
	//   AA = > 11100301 <
	//   AB = [ 1 ]
	//    AC = > HANS                 <
	//    AE = > BERGMANN             <
	//    AD = > WILHELM              <
	// Record Isn: 0383
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <
}

func ExampleReadRequest_wide_character() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=23")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(9)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("B0,F0,KA")
	request.Limit = 2
	fmt.Println("Read logical data:")
	result, rErr := request.ReadISN(1200)
	if rErr != nil {
		fmt.Println("Error reading", rErr)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	result, rErr = request.ReadISN(1250)
	if rErr != nil {
		fmt.Println("Error reading", rErr)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	result, rErr = request.ReadISN(1270)
	if rErr != nil {
		fmt.Println("Error reading", rErr)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 1200
	//   B0 = [ 1 ]
	//    BA = > Karin                                    <
	//    BB = >                                          <
	//    BC = > Norlin                                             <
	//   F0 = [ 1 ]
	//    FA[01] = [ 1 ]
	//     FA[01,01] = >  Trångsund 4                                                <
	//    FB[01] = > STOCKHOLM                                <
	//    FC[01] = > 111 29     <
	//    FD[01] = > S   <
	//    F1[01] = [ 1 ]
	//     FE[01] = >  08    <
	//     FF[01] = > 659803          <
	//     FG[01] = >                 <
	//     FH[01] = >                 <
	//     FI[01] = [ 0 ]
	//   KA = > försäljningsrepresentant                                         <
	// Result data:
	// Dump all result values
	// Record Isn: 1250
	//   B0 = [ 1 ]
	//    BA = > Игорь                               <
	//    BB = > Петрович                         <
	//    BC = > Михайлов                                   <
	//   F0 = [ 1 ]
	//    FA[01] = [ 1 ]
	//     FA[01,01] = > Ивановская 26-5                                    <
	//    FB[01] = > Санкт-Петербург            <
	//    FC[01] = > 190202     <
	//    FD[01] = > RUS <
	//    F1[01] = [ 1 ]
	//     FE[01] = > 812    <
	//     FF[01] = > 8781132         <
	//     FG[01] = >                 <
	//     FH[01] = >                 <
	//     FI[01] = [ 0 ]
	//   KA = > директор                                                   <
	// Result data:
	// Dump all result values
	// Record Isn: 1270
	//   B0 = [ 1 ]
	//    BA = > महेश                             <
	//    BB = > जाधव                             <
	//    BC = > कुलदीप                                 <
	//   F0 = [ 1 ]
	//    FA[01] = [ 1 ]
	//     FA[01,01] = > 18-क/12 रानीगंज कैला                 <
	//    FB[01] = > जयपुर                          <
	//    FC[01] = > 302001     <
	//    FD[01] = > IND <
	//    F1[01] = [ 1 ]
	//     FE[01] = > 06726  <
	//     FF[01] = > 672309          <
	//     FG[01] = >                 <
	//     FH[01] = >                 <
	//     FI[01] = [ 0 ]
	//   KA = > रीसेपसणिस्त                                  <
}

type testedValue struct {
	longName  string
	shortName string
	length    uint32
	index     uint32
}

type testedValueChecker struct {
	tvcMap map[string]*testedValue
	t      *testing.T
}

func registerTestedValuesAvailable(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	tvc := x.(*testedValueChecker)

	if adaValue.Type().Name() == "MA" {
		structureValue := adaValue.(*adatypes.StructureValue)
		for currentIndex := 1; currentIndex < structureValue.NrElements()+1; currentIndex++ {
			v := structureValue.Get("MB", currentIndex)
			// fmt.Printf("Got v >%s<\n", v)
			vt := strings.TrimSpace(v.String())
			if tv, ok := tvc.tvcMap[vt]; ok {
				vln := structureValue.Get("MD", currentIndex)
				assert.Equal(tvc.t, tv.longName, strings.TrimSpace(vln.String()))
				vln = structureValue.Get("ML", currentIndex)
				assert.Equal(tvc.t, tv.index, uint32(currentIndex))
			} else {
				// fmt.Println("No Found tv element ", ok)

			}
		}
	}
	return adatypes.Continue, nil
}

func TestConnectionReadMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, cerr := NewConnection("acj;target=24")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(4)
	if !assert.NoError(t, err) {
		fmt.Println("Error read map : ", err)
		return
	}
	if !assert.NotNil(t, request) {
		return
	}
	fmt.Println("Connection : ", connection)

	request.QueryFields("RN,MA")
	request.Limit = 2
	result := &RequestResult{}
	// Read only 'EMPLOYEES-NAT-DDM' map
	err = request.ReadLogicalWithWithParser("RN=EMPLOYEES-NAT-DDM", nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.True(t, len(result.Values) > 0) {
		fmt.Println("Result data:")
		record := result.Values[0]
		tm := adatypes.TraverserValuesMethods{EnterFunction: registerTestedValuesAvailable}
		tvc := &testedValueChecker{t: t}
		tvc.tvcMap = map[string]*testedValue{
			"AA": &testedValue{shortName: "AA", longName: "PERSONNEL-ID", length: 8, index: 1},
			"AB": &testedValue{shortName: "AB", longName: "FULL-NAME", length: 0, index: 2},
			"AD": &testedValue{shortName: "AD", longName: "MIDDLE-I", length: 10, index: 4},
			"AG": &testedValue{shortName: "AG", longName: "SEX", length: 1, index: 7},
			"AI": &testedValue{shortName: "AI", longName: "ADDRESS-LINE", length: 20, index: 10},
			"AP": &testedValue{shortName: "AP", longName: "JOB-TITLE", length: 25, index: 18},
			"AZ": &testedValue{shortName: "AZ", longName: "LANG", length: 3, index: 29},
			"S3": &testedValue{shortName: "S3", longName: "CURRENCY-SALARY", length: 0, index: 33},
		}
		record.traverse(tm, tvc)
		// result.DumpValues()
	}

}

func ExampleReadRequest_blendMap() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if cerr != nil {
		return
	}
	defer connection.Close()
	fmt.Println("Connection :", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection :", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("NAME,PERSONNEL-ID")
	request.Limit = 2
	result := &RequestResult{}
	fmt.Println("Read logical data:")
	err = request.ReadLogicalWithWithParser("PERSONNEL-ID=[11100301:11100303]", nil, result)
	if err != nil {
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection : Target not defined
	// Connection : Map=EMPLOYEES-NAT-DDM Adabas url=24 fnr=0 connection file=11
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0251
	//   PERSONNEL-ID = > 11100301 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > BERGMANN             <
	// Record Isn: 0383
	//   PERSONNEL-ID = > 11100302 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > HAIBACH              <
}

func ExampleReadRequest_isn() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=23")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	result := &RequestResult{}
	fmt.Println("Read ISN 250:")
	err = request.ReadISNWithParser(250, nil, result)
	if err != nil {
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=23 fnr=0
	// Read ISN 250:
	// Result data:
	// Dump all result values
	// Record Isn: 0250
	//   AA = > 11222222 <
	//   AB = [ 1 ]
	//    AC = > ANTONIA              <
	//    AE = > MARTENS              <
	//    AD = > MARIA                <
	//   AF = > S <
	//   AG = > F <
	//   AH = > 713104 <
	//   A1 = [ 1 ]
	//    AI = [ 3 ]
	//     AI[01] = > C/O H.KOERBER        <
	//     AI[02] = > AM DORNKAMP 20       <
	//     AI[03] = > 4590 CLOPPENBURG     <
	//    AJ = > CLOPPENBURG          <
	//    AK = > 4590       <
	//    AL = > D   <
	//   A2 = [ 1 ]
	//    AN = > 04471  <
	//    AM = > 3082            <
	//   AO = > MGMT00 <
	//   AP = > DATENSCHUTZBEAUFTRAGTE    <
	//   AQ = [ 3 ]
	//    AR[01] = > EUR <
	//    AS[01] = > 29743 <
	//    AT[01] = [ 2 ]
	//     AT[01,01] = > 4615 <
	//     AT[01,02] = > 8000 <
	//    AR[02] = > EUR <
	//    AS[02] = > 22153 <
	//    AT[02] = [ 2 ]
	//     AT[02,01] = > 3589 <
	//     AT[02,02] = > 6000 <
	//    AR[03] = > EUR <
	//    AS[03] = > 20769 <
	//    AT[03] = [ 1 ]
	//     AT[03,01] = > 1538 <
	//   A3 = [ 1 ]
	//    AU = > 33 <
	//    AV = > 4 <
	//   AW = [ 2 ]
	//    AX[01] = > 19980701 <
	//    AY[01] = > 19980702 <
	//    AX[02] = > 19980811 <
	//    AY[02] = > 19980812 <
	//   AZ = [ 2 ]
	//    AZ[01] = > GER <
	//    AZ[02] = > TUR <
	//   PH = >  <
	//   H1 = > 3304 <
	//   S1 = > MGMT <
	//   S2 = > MGMT00MARTENS              <
	//   S3 = >  <
}

func TestConnectionADATCPSimpleRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "111(adatcp://pctkn10:60001)"
	fmt.Println("Connect to ", url)
	connection, err := NewConnection("acj;target=" + url + ")")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
}

func TestConnectionReadOneLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadISNWithParser(1, nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		fmt.Printf("Result: %p\n", result)
		err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionReadAllLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadPhysicalSequenceWithParser(nil, result)
	// err = request.ReadISNWithParser(202, nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		// fmt.Printf("Result: %p\n", result)
		// err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionReadSpecialLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	result := &RequestResult{}
	// err = request.ReadPhysicalSequenceWithParser(nil, result)
	err = request.ReadISNWithParser(380, nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		fmt.Printf("Result: %p\n", result)
		err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionADATCPReadRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "111(adatcp://pctkn10:60001)"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadPhysicalSequenceWithParser(nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		// fmt.Printf("Result: %p\n", result)
		//err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionReadUnicode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("B0,JA,KA")
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadLogicalWithWithParser("AA=[40003001:40005001]", nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 10, len(result.Values))
		assert.Equal(t, 10, result.NrRecords())
		// err = result.DumpValues()
		// assert.NoError(t, err)
		kaVal := result.Values[0].HashFields["KA"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "रीसेपसणिस्त                                 ", kaVal.String())
		}
		kaVal = result.Values[9].HashFields["KA"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "ಸೆನಿಓರ್ ಪ್ರೋಗ್ೃಾಮ್ಮೇರ್  ", kaVal.String())
		}

		record := result.Isn(1265)
		assert.NotNil(t, record)
	}
}

func TestConnectionReadDeepPEFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("AA,F0")
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadLogicalWithWithParser("AA=[40003001:40005001]", nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		err = result.DumpValues()
		assert.NoError(t, err)
		assert.Equal(t, 10, result.NrRecords())
		kaVal, err := result.Values[0].SearchValueIndex("FB", []uint32{1})
		assert.NoError(t, err)
		assert.NotNil(t, kaVal)
		assert.Equal(t, "जयपुर                         ", kaVal.String())
		kaVal, err = result.Values[0].SearchValueIndex("FG", []uint32{1})
		assert.NoError(t, err)
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "               ", kaVal.String())
		}
	}
}

func TestConnectionReadAllFields9(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	result := &RequestResult{}
	err = request.ReadLogicalWithWithParser("AA=[40003001:40005001]", nil, result)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionADIS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "111(adatcp://pctkn10:60001)"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
}

func TestConnectionNotConnected(t *testing.T) {
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "111(adatcp://xxx:60001)"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	assert.Error(t, openErr, "Error expected because host not exist")
	assert.Equal(t, "ADAGE95000: System communication error (rsp=149,subrsp=0,dbid=111(adatcp://xxx:60001),file=0)", openErr.Error())
}

func TestConnectionSimpleStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest, rErr := connection.CreateStoreRequest(16)
	if rErr != nil {
		return
	}
	sferr := storeRequest.StoreFields("AA,AB")
	if sferr != nil {
		fmt.Println("Error setting fields", sferr)
		return
	}
	record, err := storeRequest.CreateRecord()
	if err != nil {
		fmt.Println("Error creating record", err)
		return
	}
	err = record.SetValueWithIndex("AA", nil, "777777_0")
	err = record.SetValueWithIndex("AC", nil, "WABER")
	err = record.SetValueWithIndex("AD", nil, "EMIL")
	err = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest.Store(record)
	if !assert.NoError(t, err) {
		return
	}

	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkStoreByFile(t, "23", 16, "777777")
}

func checkStoreByFile(t *testing.T, target string, file uint32, search string) error {
	connection, err := NewConnection("acj;target=" + target)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()
	readRequest, rrerr := connection.CreateReadRequest(file)
	if !assert.NoError(t, rrerr) {
		return rrerr
	}
	fields := "AA,AB"
	searchField := "AA"

	switch file {
	case 18:
		fields = "CA,EB"
		searchField = "CA"
	case 19:
		fields = "AA,CD"
		searchField = "AA"
	}
	err = readRequest.QueryFields(fields)
	if !assert.NoError(t, err) {
		return err
	}
	result, rerr := readRequest.ReadLogicalWith(searchField + "=[" + search + "_:" + search + "_Z]")
	if !assert.NoError(t, rerr) {
		return rerr
	}
	return validateResult(t, search, result)
}

func validateResult(t *testing.T, search string, result *RequestResult) error {
	if !assert.NotNil(t, result) {
		return fmt.Errorf("Result empty")
	}
	fmt.Printf("Validate result %+v %d\n", result, len(result.Values))
	if !assert.True(t, len(result.Values) > 0) {
		return fmt.Errorf("Result zero")
	}
	resultJSON, err := json.Marshal(result.Values)
	if !assert.NoError(t, err) {
		return err
	}
	var re = regexp.MustCompile(`(?m)"ISN[^,]*,`)
	resultJSON = re.ReplaceAll(resultJSON, []byte(""))
	rw := os.Getenv("REFERENCES")
	doWrite := os.Getenv("REFERENCE_WRITE")
	destinationFile := rw + "/" + search + ".json"
	if _, err := os.Stat(destinationFile); os.IsNotExist(err) {
		doWrite = "1"
	}
	if doWrite == "" {
		fmt.Println("Check reference to", destinationFile)
		referenceJSON, err := ioutil.ReadFile(destinationFile)
		if !assert.NoError(t, err) {
			return err
		}
		fmt.Println("Reference check inactive")
		assert.Equal(t, referenceJSON, resultJSON)
	} else {
		fmt.Println("Write reference check to", destinationFile)
		os.Remove(destinationFile)
		err = ioutil.WriteFile(destinationFile, resultJSON, 0644)
		if !assert.NoError(t, err) {
			return err
		}
	}
	return nil
}

func TestConnectionSimpleMultipleStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}

	log.Debug("TEST: ", t.Name())
	connection, err := NewConnection("acj;target=23")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest16, rErr := connection.CreateStoreRequest(16)
	if !assert.NoError(t, rErr) {
		return
	}
	storeRequest16.StoreFields("AA,AB")
	record, err := storeRequest16.CreateRecord()
	err = record.SetValueWithIndex("AA", nil, "16555_0")
	err = record.SetValueWithIndex("AC", nil, "WABER")
	err = record.SetValueWithIndex("AD", nil, "EMIL")
	err = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest16.Store(record)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest19, rErr := connection.CreateStoreRequest(19)
	if !assert.NoError(t, rErr) {
		return
	}
	storeRequest19.StoreFields("AA,CD")
	record, err = storeRequest19.CreateRecord()
	err = record.SetValueWithIndex("AA", nil, "19555_0")
	err = record.SetValueWithIndex("AC", nil, "WABER")
	err = record.SetValueWithIndex("AD", nil, "EMIL")
	err = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest19.Store(record)
	if !assert.NoError(t, err) {
		return
	}

	err = connection.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkStoreByFile(t, "23", 16, "16555")
	checkStoreByFile(t, "23", 19, "19555")
}
