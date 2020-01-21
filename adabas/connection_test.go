/*
* Copyright © 2018-2020 Software AG, Darmstadt, Germany and/or its licensors
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
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

func initTestLogWithFile(t *testing.T, fileName string) {
	err := initLogWithFile(fileName)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
		return
	}
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

func initLogWithFile(fileName string) (err error) {
	level := "error"
	ed := os.Getenv("ENABLE_DEBUG")
	switch ed {
	case "1":
		level = "debug"
		adatypes.Central.SetDebugLevel(true)
	case "2":
		level = "info"
	default:
	}
	return initLogLevelWithFile(fileName, level)
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Remove leading slash left by url.Parse()
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func initLogLevelWithFile(fileName string, level string) (err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	var name string
	if runtime.GOOS == "windows" {
		zap.RegisterSink("winfile", newWinFileSink)
		//		OutputPaths: []string{"stdout", "winfile:///" + filepath.Join(GlobalConfigDir.Path, "info.log.json")},
		name = "winfile:///" + p + string(os.PathSeparator) + fileName
	} else {
		name = "file://" + filepath.ToSlash(p+string(os.PathSeparator)+fileName)
	}

	rawJSON := []byte(`{
	"level": "error",
	"encoding": "console",
	"outputPaths": [ "/tmp/logs"],
	"errorOutputPaths": ["stderr"],
	"encoderConfig": {
	  "messageKey": "message",
	  "levelKey": "level",
	  "levelEncoder": "lowercase"
	}
  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	l := zapcore.ErrorLevel
	switch level {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	default:
	}
	cfg.Level.SetLevel(l)
	cfg.OutputPaths = []string{name}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	adatypes.Central.Log = sugar

	sugar.Infof("AdabasGoApi logger initialization succeeded")

	return nil
}

type parseTestStructure struct {
	storeRequest *StoreRequest
	fields       string
	t            *testing.T
}

func parseTestConnection(adabasRequest *adatypes.Request, x interface{}) (err error) {
	fmt.Println("Parse Test connection")
	parseTestStructure := x.(parseTestStructure)
	if parseTestStructure.t == nil {
		panic("Parse test structure empty test instance")
	}
	if !assert.NotNil(parseTestStructure.t, adabasRequest.Definition.Values) {
		adatypes.Central.Log.Debugf("Parse Buffer .... values avail.=%v", (adabasRequest.Definition.Values == nil))
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
	// storeRecord.DumpValues()
	//adatypes.Central.Log.Println("Store record =====================================")
	err = storeRequest.Store(storeRecord)
	fmt.Println("ISN: ", storeRecord.Isn, " -> ", err)
	return
}

func deleteRecords(adabasRequest *adatypes.Request, x interface{}) (err error) {
	deleteRequest := x.(*DeleteRequest)
	// fmt.Printf("Delete ISN: %d on %s/%d\n", adabasRequest.Isn, deleteRequest.repository.URL.String(), deleteRequest.repository.Fnr)
	err = deleteRequest.Delete(adabasRequest.Isn)
	return
}

func TestConnectionSimpleTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	assert.NoError(t, err)
	deleteRequest.EndTransaction()

	request, rErr2 := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rErr2) {
		return
	}
	err = request.loadDefinition()
	if !assert.NoError(t, err) {
		return
	}

	adatypes.Central.Log.Debugf("Loaded Definition in Tests")
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
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()
}

func TestConnectionOpenOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
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
	// time.Sleep(10 * time.Second)
}

func TestConnectionOpenFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(16)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	assert.NoError(t, err)
	deleteRequest.EndTransaction()

	request, rErr2 := connection.CreateFileReadRequest(11)
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(16)
	assert.NoError(t, dErr)
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	assert.NoError(t, err)
	fmt.Println("Delete done, call end of transaction")
	adatypes.Central.Log.Debugf("Delete done, call end of transaction")
	deleteRequest.EndTransaction()

	fmt.Println("Call Read to 11")
	request, rErr2 := connection.CreateFileReadRequest(11)
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, connection) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 10

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	var result *Response
	result, err = readRequest.ReadPhysicalSequence()
	assert.NoError(t, err)
	// result.DumpValues()
	assert.Equal(t, 1107, len(result.Values))
}

func TestConnectionNoMultifetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, connection) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 1

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	var result *Response
	result, err = readRequest.ReadPhysicalSequence()
	assert.NoError(t, err)
	assert.Equal(t, 1107, len(result.Values))
}

func TestConnectionPeriodAndMultipleField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.Limit = 0

	qErr := readRequest.QueryFields("AA,AQ,AZ")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	result, readErr := readRequest.ReadISN(499)
	if !assert.NoError(t, readErr) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AA")) {
		return
	}
	if !assert.Equal(t, int32(2), result.Values[0].ValueQuantity("AQ")) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AS[1]")) {
		return
	}
	if !assert.Equal(t, int32(0), result.Values[0].ValueQuantity("AT[1]")) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AS[2]")) {
		return
	}
	if !assert.Equal(t, int32(0), result.Values[0].ValueQuantity("AT[2]")) {
		return
	}
	if !assert.Equal(t, int32(2), result.Values[0].ValueQuantity("AZ")) {
		return
	}
	// result.DumpValues()
}

func TestConnectionPeriodAndMultipleQuantity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.Limit = 0

	qErr := readRequest.QueryFields("AA,AQ,AW")
	assert.NoError(t, qErr)
	fmt.Println("Result data:")
	result, readErr := readRequest.ReadISN(250)
	if !assert.NoError(t, readErr) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AA")) {
		return
	}
	if !assert.Equal(t, int32(3), result.Values[0].ValueQuantity("AQ")) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AS[1]")) {
		return
	}
	if !assert.Equal(t, int32(2), result.Values[0].ValueQuantity("AT[1]")) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AS[2]")) {
		return
	}
	if !assert.Equal(t, int32(2), result.Values[0].ValueQuantity("AT[2]")) {
		return
	}
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("AT[3]")) {
		return
	}
	if !assert.Equal(t, int32(2), result.Values[0].ValueQuantity("AW")) {
		return
	}
	// result.DumpValues()
}

func TestConnectionRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	_, cerr := NewConnection("acj;target=" + url + ")")
	assert.Error(t, cerr)
	assert.Equal(t, "ADG0000115: Entire Network target drivers cannot be connect directly, configure Adabas client.", cerr.Error())

}

func TestConnectionWithMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
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
		fmt.Println("Read logigcal data:")
		var result *Response
		result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
		assert.NoError(t, err)
		fmt.Println("Result data:")
		// result.DumpValues()
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
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadPhysicalSequence()
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
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

func ExampleConnection_readLogicalWith() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	var result *Response
	fmt.Println("Read logical data:")
	result, err = request.ReadLogicalWith("AA=[11100301:11100303]")
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

func ExampleConnection_periodGroup() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB,AQ,AZ")
	fmt.Println("Read logical data:")
	result, rerr := request.ReadISN(250)
	if rerr != nil {
		fmt.Println("Error reading", rerr)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0250
	//   AA = > 11222222 <
	//   AB = [ 1 ]
	//    AC = > ANTONIA              <
	//    AE = > MARTENS              <
	//    AD = > MARIA                <
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
	//   AZ = [ 2 ]
	//    AZ[01] = > GER <
	//    AZ[02] = > TUR <
}

func ExampleConnection_wideCharacter() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(9)
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

func ExampleConnection_marhsalJSONComplete() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("*")
	fmt.Println("Read logical data:")
	result, rErr := request.ReadISN(1)
	if rErr != nil {
		fmt.Println("Error reading", rErr)
		return
	}
	x, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		fmt.Println("Error", jsonErr)
		return
	}
	fmt.Println(string(x))

	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// {"Records":[{"A1":{"AI":["26 AVENUE RHIN ET DA"],"AJ":"JOIGNY","AK":"89300","AL":"F"},"A2":{"AM":"44864858","AN":"1033"},"A3":{"AU":19,"AV":5},"AA":"50005800","AB":{"AC":"SIMONE","AD":"","AE":"ADAM"},"AF":"M","AG":"F","AH":712981,"AO":"VENT59","AP":"CHEF DE SERVICE","AQ":[{"AR":"EUR","AS":963,"AT":[138]}],"AW":[{"AX":19990801,"AY":19990831}],"AZ":["FRE","ENG"],"ISN":1}]}
}

func ExampleConnection_marhsalJSON() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("*")
	fmt.Println("Read logical data:")
	result, rErr := request.ReadISN(250)
	if rErr != nil {
		fmt.Println("Error reading", rErr)
		return
	}
	x, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		fmt.Println("Error", jsonErr)
		return
	}
	fmt.Println(string(x))

	// Output: Connection :  Adabas url=23 fnr=0
	// Limit query data:
	// Read logical data:
	// {"Records":[{"A1":{"AI":["C/O H.KOERBER","AM DORNKAMP 20","4590 CLOPPENBURG"],"AJ":"CLOPPENBURG","AK":"4590","AL":"D"},"A2":{"AM":"3082","AN":"04471"},"A3":{"AU":33,"AV":4},"AA":"11222222","AB":{"AC":"ANTONIA","AD":"MARIA","AE":"MARTENS"},"AF":"S","AG":"F","AH":713104,"AO":"MGMT00","AP":"DATENSCHUTZBEAUFTRAGTE","AQ":[{"AR":"EUR","AS":29743,"AT":[4615,8000]},{"AR":"EUR","AS":22153,"AT":[3589,6000]},{"AR":"EUR","AS":20769,"AT":[1538]}],"AW":[{"AX":19980701,"AY":19980702},{"AX":19980811,"AY":19980812}],"AZ":["GER","TUR"],"ISN":250}]}
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
				//	vln = structureValue.Get("ML", currentIndex)
				assert.Equal(tvc.t, tv.index, uint32(currentIndex))
				// } else {
				// 	// fmt.Println("No Found tv element ", ok)

			}
		}
	}
	return adatypes.Continue, nil
}

func TestConnectionReadMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;target=24")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(4)
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
	// Read only 'EMPLOYEES-NAT-DDM' map
	var result *Response
	result, err = request.ReadLogicalWith("RN=EMPLOYEES-NAT-DDM")
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
		record.Traverse(tm, tvc)
		// result.DumpValues()
	}

}

func ExampleConnection_map() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;map;config=[24,4]")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error read map : ", err)
		return
	}
	fmt.Println("Connection :", connection)
	request.QueryFields("NAME,PERSONNEL-ID")
	request.Limit = 2
	fmt.Println("Read logical data, two records:")
	result, rerr := request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
	if rerr != nil {
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection : Map=EMPLOYEES-NAT-DDM Adabas url=24 fnr=0 connection file=11
	// Read logical data, two records:
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

func ExampleConnection_readIsn() {
	initLogWithFile("connection.log")
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Error create request: ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Read ISN 250:")
	var result *Response
	result, err = request.ReadISN(250)
	if err != nil {
		fmt.Println("Error reading ISN: ", err)
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	url := adabasTCPLocation()
	fmt.Println("Connect to ", url)
	connection, err := NewConnection("acj;target=177(adatcp://" + url + ")")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
}

func ExampleReadRequest_readISN() {
	err := initLogWithFile("connection.log")
	if err != nil {
		fmt.Println("Log init error", err)
		return
	}

	url := adabasModDBIDs
	connection, cerr := NewConnection("acj;target=" + url)
	if cerr != nil {
		fmt.Println("Connection error", err)
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Open error", openErr)
	}
	request, err := connection.CreateFileReadRequest(11)
	if err != nil {
		fmt.Println("Create read request error", openErr)
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadISN(1)
	if err != nil {
		fmt.Println("Read  error", openErr)
		return
	}
	if result != nil {
		err = result.DumpValues()
		if err != nil {
			fmt.Println("Dump values  error", openErr)
			return
		}
	}

	// Output: Adabas url=23 fnr=0
	// Dump all result values
	// Record Isn: 0001
	//   AA = > 50005800 <
	//   AB = [ 1 ]
	//    AC = > SIMONE               <
	//    AE = > ADAM                 <
	//    AD = >                      <
	//   AF = > M <
	//   AG = > F <
	//   AH = > 712981 <
	//   A1 = [ 1 ]
	//    AI = [ 1 ]
	//     AI[01] = > 26 AVENUE RHIN ET DA <
	//    AJ = > JOIGNY               <
	//    AK = > 89300      <
	//    AL = > F   <
	//   A2 = [ 1 ]
	//    AN = > 1033   <
	//    AM = > 44864858        <
	//   AO = > VENT59 <
	//   AP = > CHEF DE SERVICE           <
	//   AQ = [ 1 ]
	//    AR[01] = > EUR <
	//    AS[01] = > 963 <
	//    AT[01] = [ 1 ]
	//     AT[01,01] = > 138 <
	//   A3 = [ 1 ]
	//    AU = > 19 <
	//    AV = > 5 <
	//   AW = [ 1 ]
	//    AX[01] = > 19990801 <
	//    AY[01] = > 19990831 <
	//   AZ = [ 2 ]
	//    AZ[01] = > FRE <
	//    AZ[02] = > ENG <
	//   PH = >  <
	//   H1 = > 1905 <
	//   S1 = > VENT <
	//   S2 = > VENT59ADAM                 <
	//   S3 = >  <
}

func TestConnectionReadAllLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadPhysicalSequence()
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadISN(380)
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasTCPLocation()
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=177(adatcp://" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection:", connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Multifetch entries:", request.Multifetch)
	request.Limit = 0
	var result *Response
	result, err = request.ReadPhysicalSequence()
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		// fmt.Printf("Result: %p\n", result)
		//err = result.DumpValues()
		assert.NoError(t, err)
		assert.Equal(t, 1107, len(result.Values))
	}
}

func TestConnectionReadUnicode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("B0,JA,KA")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=[40003001:40005001]")
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("AA,F0")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=[40003001:40005001]")
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=[40003001:40005001]")
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		err = result.DumpValues()
		assert.NoError(t, err)
	}
}

func TestConnectionRead9FieldPicture(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasStatDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("RA")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=11300323")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	assert.NoError(t, err)
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("RA")) {
		return
	}
	v, _ := result.Values[0].SearchValue("RA")
	raw := v.Bytes()
	assert.Equal(t, 183049, len(raw))
	md5sum := fmt.Sprintf("%X", md5.Sum(raw))
	assert.Equal(t, "8B124C139790221469EF6308D6554660", md5sum)

}

func TestConnectionRead9FieldDocument(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasStatDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("SC")
	request.Limit = 0
	request.Multifetch = 1
	request.RecordBufferShift = 10000000
	var result *Response
	adatypes.Central.Log.Infof("TEST: Start Read call")
	result, err = request.ReadLogicalWith("AA=11300323")
	adatypes.Central.Log.Infof("TEST: Read call done")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	assert.NoError(t, err)
	if !assert.Equal(t, int32(1), result.Values[0].ValueQuantity("SC")) {
		return
	}
	if !assert.Equal(t, int32(3), result.Values[0].ValueQuantity("SC[1]")) {
		return
	}
	checkChecksum(t, result.Values[0], "SC[1,1]", "7B64C5D56AED33B749B0653DADC02F2D", 26477)
	checkChecksum(t, result.Values[0], "SC[1,2]", "532A1D58A92EE7E206A250B6DD5FC08B", 87529)
	checkChecksum(t, result.Values[0], "SC[1,3]", "297E8428DCA7CF22062D93CDA0CC359A", 23118)
}

func TestConnectionRead9FieldDocumentTo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasStatDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("AA,SC")
	request.Limit = 0
	request.Multifetch = 1
	request.RecordBufferShift = 10000000
	var result *Response
	adatypes.Central.Log.Infof("TEST: Start Read call")
	result, err = request.ReadLogicalWith("AA<=11300323")
	adatypes.Central.Log.Infof("TEST: Read call done")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 84, len(result.Values))
	valueIndex := -1
	for i, v := range result.Values {
		tv, _ := v.SearchValue("AA")
		if tv.String() == "11300323" {
			fmt.Println("Found value at ", i)
			valueIndex = i
			break
		}
	}
	assert.NoError(t, err)
	if !assert.Equal(t, int32(1), result.Values[valueIndex].ValueQuantity("SC")) {
		return
	}
	if !assert.Equal(t, int32(3), result.Values[valueIndex].ValueQuantity("SC[1]")) {
		return
	}
	checkChecksum(t, result.Values[valueIndex], "SC[1,1]", "7B64C5D56AED33B749B0653DADC02F2D", 26477)
	checkChecksum(t, result.Values[valueIndex], "SC[1,2]", "532A1D58A92EE7E206A250B6DD5FC08B", 87529)
	checkChecksum(t, result.Values[valueIndex], "SC[1,3]", "297E8428DCA7CF22062D93CDA0CC359A", 23118)
}

func checkChecksum(t *testing.T, record *Record, field, expectMd5 string, expectLen int) {
	v, verr := record.SearchValue(field)
	if !assert.NoError(t, verr) {
		return
	}
	if !assert.NotNil(t, v) {
		return
	}

	raw := v.Bytes()
	assert.Equal(t, expectLen, len(raw))
	md5sum := fmt.Sprintf("%X", md5.Sum(raw))
	assert.Equal(t, expectMd5, md5sum)
	return
}

func TestConnectionReadOnlyPEFields9(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("F0,L0")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=40003001")
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
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasTCPLocation()
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=177(adatcp://" + url + ")")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
}

func TestConnectionNotConnected(t *testing.T) {
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
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

func ExampleConnection_endTransaction() {
	lerr := initLogWithFile("connection.log")
	if lerr != nil {
		return
	}

	fmt.Println("Example for EndTransaction()")
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if err != nil {
		fmt.Println("Error creating connection", err)
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest, rErr := connection.CreateStoreRequest(16)
	if rErr != nil {
		return
	}
	// Define fields to be included in the request
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
	if err != nil {
		fmt.Println("Error set value", err)
		return
	}
	err = record.SetValueWithIndex("AC", nil, "WABER")
	if err != nil {
		fmt.Println("Error set value", err)
		return
	}
	err = record.SetValueWithIndex("AD", nil, "EMIL")
	if err != nil {
		fmt.Println("Error set value", err)
		return
	}
	err = record.SetValueWithIndex("AE", nil, "MERK")
	if err != nil {
		fmt.Println("Error set value", err)
		return
	}

	// Store the record in the database
	err = storeRequest.Store(record)
	if err != nil {
		fmt.Println("Store record error", err)
		return
	}

	// ET end of transaction final commit the transaction
	err = storeRequest.EndTransaction()
	if err != nil {
		fmt.Println("End transaction error", err)
		return
	}
	fmt.Println("Record stored, check content ...")
	readRequest, rrerr := connection.CreateFileReadRequest(16)
	if rrerr != nil {
		fmt.Println("Read request error", rrerr)
		return
	}
	err = readRequest.QueryFields("AA,AB")
	if err != nil {
		fmt.Println("Query fields error", err)
		return
	}
	result, rerr := readRequest.ReadLogicalWith("AA=[777777_:777777_Z]")
	if rerr != nil {
		fmt.Println("Read record error", rerr)
		return
	}
	if len(result.Values) != 1 {
		fmt.Println("Records received not correct", len(result.Values))
		return
	}
	// To adapt output for example
	result.Values[0].Isn = 0
	result.DumpValues()

	// Output: Example for EndTransaction()
	// Record stored, check content ...
	// Dump all result values
	//   AA = > 777777_0 <
	//   AB = [ 1 ]
	//    AC = > WABER                <
	//    AE = > MERK                 <
	//    AD = > EMIL                 <

}

func checkStoreByFile(t *testing.T, target string, file Fnr, search string) error {
	connection, err := NewConnection("acj;target=" + target)
	if !assert.NoError(t, err) {
		return err
	}
	defer connection.Close()
	readRequest, rrerr := connection.CreateFileReadRequest(file)
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

func validateResult(t *testing.T, search string, result *Response) error {
	if !assert.NotNil(t, result) {
		return fmt.Errorf("Result empty")
	}
	fmt.Printf("Validate result %d values\n", len(result.Values))
	if !assert.True(t, len(result.Values) > 0) {
		return fmt.Errorf("Result zero")
	}
	resultJSON, err := json.Marshal(result.Values)
	if !assert.NoError(t, err) {
		return err
	}
	var re = regexp.MustCompile(`(?m)[,]?"ISN[^,]*[},]`)
	resultJSON = re.ReplaceAll(resultJSON, []byte(""))
	// fmt.Println(string(resultJSON))
	rw := os.Getenv("REFERENCES")
	doWrite := os.Getenv("REFERENCE_WRITE")
	destinationFile := rw + string(os.PathSeparator) + search + ".json"
	if _, err := os.Stat(destinationFile); os.IsNotExist(err) {
		doWrite = "1"
	}
	if doWrite == "" {
		fmt.Println("Check reference to", destinationFile)
		referenceJSON, err := ioutil.ReadFile(destinationFile)
		if !assert.NoError(t, err) {
			return err
		}
		fmt.Println("Compare reference with result")
		assert.Equal(t, referenceJSON, resultJSON, "Reference not equal result")
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
	initTestLogWithFile(t, "connection.log")

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
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
	assert.NoError(t, err)
	_ = record.SetValueWithIndex("AA", nil, "16555_0")
	_ = record.SetValueWithIndex("AC", nil, "WABER")
	_ = record.SetValueWithIndex("AD", nil, "EMIL")
	_ = record.SetValueWithIndex("AE", nil, "MERK")
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
	if !assert.NoError(t, err) {
		return
	}
	_ = record.SetValueWithIndex("AA", nil, "19555_0")
	_ = record.SetValueWithIndex("AC", nil, "WABER")
	_ = record.SetValueWithIndex("AD", nil, "EMIL")
	_ = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest19.Store(record)
	if !assert.NoError(t, err) {
		return
	}

	err = connection.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkStoreByFile(t, adabasModDBIDs, 16, "16555")
	checkStoreByFile(t, adabasModDBIDs, 19, "19555")
}

func ExampleConnection_store() {
	err := initLogWithFile("connection.log")
	if err != nil {
		return
	}

	if cErr := clearFile(16); cErr != nil {
		return
	}
	if cErr := clearFile(19); cErr != nil {
		return
	}

	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if err != nil {
		return
	}
	defer connection.Close()
	connection.Open()
	storeRequest16, rErr := connection.CreateStoreRequest(16)
	if rErr != nil {
		return
	}
	serr := storeRequest16.StoreFields("AA,AB")
	if serr != nil {
		return
	}
	record, err := storeRequest16.CreateRecord()
	if err != nil {
		return
	}
	_ = record.SetValueWithIndex("AA", nil, "16555_0")
	_ = record.SetValueWithIndex("AC", nil, "WABER")
	_ = record.SetValueWithIndex("AD", nil, "EMIL")
	_ = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest16.Store(record)
	if err != nil {
		return
	}
	storeRequest19, rErr := connection.CreateStoreRequest(19)
	if rErr != nil {
		return
	}
	err = storeRequest19.StoreFields("AA,CD")
	if err != nil {
		return
	}
	record, err = storeRequest19.CreateRecord()
	if err != nil {
		return
	}
	_ = record.SetValueWithIndex("AA", nil, "19555_0")
	_ = record.SetValueWithIndex("AC", nil, "WABER")
	_ = record.SetValueWithIndex("AD", nil, "EMIL")
	_ = record.SetValueWithIndex("AE", nil, "MERK")
	err = storeRequest19.Store(record)
	if err != nil {
		return
	}

	err = connection.EndTransaction()
	if err != nil {
		return
	}
	fmt.Println("Read file 16 ...")
	err = dumpStoredData(adabasModDBIDs, 16, "16555")
	if err != nil {
		return
	}
	fmt.Println("Read file 19 ...")
	err = dumpStoredData(adabasModDBIDs, 19, "19555")
	if err != nil {
		return
	}

	// Output: Read file 16 ...
	// Dump all result values
	// Record Isn: 0001
	//   AA = > 16555_0  <
	//   AB = [ 1 ]
	//    AC = > WABER                <
	//    AE = > MERK                 <
	//    AD = > EMIL                 <
	// Read file 19 ...
	// Dump all result values
	// Record Isn: 0001
	//   AA = > 19555_0         <
	//   CD = [ 1 ]
	//    AD = > EMIL                 <
	//    AE = > MERK                 <
	//    AF = >            <

}

func dumpStoredData(target string, file Fnr, search string) error {
	connection, err := NewConnection("acj;target=" + target)
	if err != nil {
		return err
	}
	defer connection.Close()
	readRequest, rrerr := connection.CreateFileReadRequest(file)
	if rrerr != nil {
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

func TestConnection_NewConnectionError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := adabasModDBIDs
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("xxx;target=" + url)
	assert.Error(t, cerr)
	assert.Nil(t, connection)

	connection, cerr = NewConnection("ada;target=xxxx")
	assert.Error(t, cerr)
	assert.Nil(t, connection)

	connection, cerr = NewConnection("ada;target=1;abr=1")
	assert.Error(t, cerr)
	assert.Nil(t, connection)

}

func TestConnectionLobADATCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping LOB call in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=24(adatcp://localhost:60024)")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(202)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("DC,EC")
	if !assert.NoError(t, err) {
		return
	}
	result, rerr := readRequest.ReadISN(1)
	if !assert.NoError(t, rerr) {
		return
	}

	dc, serr := result.Values[0].SearchValue("DC")
	if !assert.NoError(t, serr) {
		return
	}
	assert.NotNil(t, dc)
	h := sha1.New()
	_, err = h.Write(dc.Bytes())
	assert.NoError(t, err)
	fmt.Printf("SHA ALL: %x\n", h.Sum(nil))
	assert.Equal(t, "a147a6bff1d2dc47e2e63404c3548d939764e6d2", fmt.Sprintf("%x", h.Sum(nil)))
	ea, eerr := result.Values[0].SearchValue("EC")
	assert.NoError(t, eerr)
	assert.NotNil(t, ea)
	assert.Equal(t, ea.String(), fmt.Sprintf("%x", h.Sum(nil)))
}

func TestConnectionLobCheckAllIPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=" + adabasStatDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(202)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("DC,EC")
	assert.NoError(t, err)
	result, rerr := readRequest.ReadLogicalBy("BB")
	if !assert.NoError(t, rerr) {
		return
	}

	for _, v := range result.Values {
		dc, serr := v.SearchValue("DC")
		assert.NoError(t, serr)
		assert.NotNil(t, dc)
		h := sha1.New()
		_, err = h.Write(dc.Bytes())
		assert.NoError(t, err)
		ea, eerr := v.SearchValue("EC")
		assert.NoError(t, eerr)
		assert.NotNil(t, ea)
		fmt.Printf("SHA ALL: %x\n", h.Sum(nil))
		assert.Equal(t, ea.String(), fmt.Sprintf("%x", h.Sum(nil)))
	}
}

func TestConnectionLobCheckAllADATCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=24(adatcp://localhost:60024)")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(202)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("DC,EC")
	assert.NoError(t, err)
	result, rerr := readRequest.ReadLogicalBy("BB")
	if !assert.NoError(t, rerr) {
		return
	}

	for _, v := range result.Values {
		dc, serr := v.SearchValue("DC")
		assert.NoError(t, serr)
		assert.NotNil(t, dc)
		h := sha1.New()
		_, err = h.Write(dc.Bytes())
		assert.NoError(t, err)
		ea, eerr := v.SearchValue("EC")
		assert.NoError(t, eerr)
		assert.NotNil(t, ea)
		fmt.Printf("SHA ALL: %x\n", h.Sum(nil))
		assert.Equal(t, ea.String(), fmt.Sprintf("%x", h.Sum(nil)))
	}
}

func TestConnectionFile9Isn242(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, rErr) {
		return
	}
	err = readRequest.QueryFields("A0,B0,DA,EA,F0,I0,L0,N0")
	if !assert.NoError(t, err) {
		return
	}
	result, rerr := readRequest.ReadISN(242)
	assert.NoError(t, rerr)
	assert.NotNil(t, result)
	//fmt.Println(result.String())
}

func TestConnectionFile9Isn297(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(9)
	if !assert.NoError(t, rErr) {
		return
	}
	err = readRequest.QueryFields("A0,B0,DA,EA,F0,I0,L0,N0")
	if !assert.NoError(t, err) {
		return
	}
	result, rerr := readRequest.ReadISN(297)
	assert.NoError(t, rerr)
	assert.NotNil(t, result)

	//fmt.Println(result.String())
}
