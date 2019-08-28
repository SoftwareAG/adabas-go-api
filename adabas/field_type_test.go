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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

const recordNamePrefix = "FIELD-TYPE-TEST"

func TestFieldTypeStore(t *testing.T) {
	initTestLogWithFile(t, "field_type.log")

	cErr := clearFile(270)
	if !assert.NoError(t, cErr) {
		return
	}

	storeRequest, _ := NewStoreRequest(adabasModDBIDs, 270)
	defer storeRequest.Close()
	//err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	//	err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP")
	err := storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr := storeRequest.CreateRecord()
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRecord.SetValue("AF", recordNamePrefix)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr = storeRequest.CreateRecord()
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRecord.SetValue("S1", "-1")
	if !assert.NoError(t, err) {
		return
	}
	x1, _ := storeRecord.searchValue("S1")
	if !assert.Equal(t, "-1", x1.String()) {
		return
	}
	err = storeRecord.SetValue("U1", "1")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S2", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	x2, _ := storeRecord.searchValue("S2")
	if !assert.Equal(t, "-1000", x2.String()) {
		return
	}
	err = storeRecord.SetValue("U2", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S4", "-100000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U4", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S8", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U8", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("BR", []byte{0x0, 0x10, 0x20})
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("B1", []byte{0xff, 0x10, 0x5, 0x0, 0x10, 0x20})
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("F4", 21.1)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("F8", 123456.1)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("A1", "X")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AS", "NORMALSTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("A2", "LARGESTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AB", "LOBSTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AF", recordNamePrefix)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("WU", "Санкт-Петербург")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("WL", "அ-8பவனி கொம்பிலேக்")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("W4", "ಸೆನಿಓರ್ ಪ್ರೋಗ್ೃಾಮ್ಮೇರ್")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("WF", "директор")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("PA", "123")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("PF", "1234")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("UP", "51234")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("UF", "542")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("UE", "1234")
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

func TestFieldTypeRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "field_type.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(270)
	if !assert.NoError(t, err) {
		return
	}
	//err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1")
	//err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	err = request.QueryFields("*")
	if !assert.NoError(t, cerr) {
		return
	}
	request.Limit = 0
	request.RecordBufferShift = 64000
	result, rerr := request.ReadLogicalWith("AF=" + recordNamePrefix)
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 2, len(result.Values))
		assert.Equal(t, 2, result.NrRecords())
		//err = result.DumpValues()
		//assert.NoError(t, err)
		kaVal := result.Values[1].HashFields["S1"]
		assert.Equal(t, "-1", kaVal.String())
		kaVal = result.Values[1].HashFields["U1"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "1", kaVal.String())
		}
		kaVal = result.Values[1].HashFields["S2"]
		assert.Equal(t, "-1000", kaVal.String())
		kaVal = result.Values[1].HashFields["S4"]
		assert.Equal(t, "-100000", kaVal.String())
		kaVal = result.Values[1].HashFields["BR"]
		if bigEndian() {
			assert.Equal(t, []byte{0x10, 0x20}, kaVal.Value())
		} else {
			assert.Equal(t, []byte{0x0, 0x10, 0x20}, kaVal.Value())
		}
		db := []byte{0xff, 0x10, 0x5, 0x0, 0x10, 0x20}
		b := make([]byte, 122)
		copy(b[:len(db)], db)
		kaVal = result.Values[1].HashFields["B1"]
		assert.Equal(t, b, kaVal.Value())
		kaVal = result.Values[1].HashFields["A1"]
		assert.Equal(t, "X", kaVal.String())
		kaVal = result.Values[1].HashFields["F4"]
		assert.Equal(t, "21.100000", kaVal.String())
		kaVal = result.Values[1].HashFields["F8"]
		assert.Equal(t, float64(123456.100000), kaVal.Value())
		err = jsonOutput(result.Values[0])
		if !assert.NoError(t, err) {
			return
		}
		jsonOutput(result.Values[1])
		if !assert.NoError(t, err) {
			return
		}
	}
}

func TestFieldTypeReadBR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "field_type.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(270)
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("BR")
	if !assert.NoError(t, cerr) {
		return
	}
	request.Limit = 0
	result, rerr := request.ReadLogicalWith("AF=" + recordNamePrefix)
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 2, len(result.Values))
		assert.Equal(t, 2, result.NrRecords())
		err = result.DumpValues()
		assert.NoError(t, err)
		kaVal := result.Values[1].HashFields["BR"]
		if bigEndian() {
			assert.Equal(t, []byte{0x10, 0x20}, kaVal.Value())
		} else {
			assert.Equal(t, []byte{0x0, 0x10, 0x20}, kaVal.Value())
		}
		err = jsonOutput(result.Values[0])
		if !assert.NoError(t, err) {
			return
		}
		jsonOutput(result.Values[1])
		if !assert.NoError(t, err) {
			return
		}
	}
}

func jsonOutput(r *Record) error {
	x, jsonErr := json.Marshal(r)
	if jsonErr != nil {
		fmt.Println("Error", jsonErr)
		// r.DumpValues()
		return jsonErr
	}
	fmt.Println(string(x))
	return nil
}

func dumpFieldTypeTestPrepare(x interface{}, b interface{}) (adatypes.TraverseResult, error) {
	record := x.(*Record)
	if record == nil {
		return adatypes.EndTraverser, adatypes.NewGenericError(25)
	}
	fmt.Printf("Record found:\n")
	return adatypes.Continue, nil
}

func dumpFieldTypeValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	y := strings.Repeat(" ", int(adaValue.Type().Level()))

	if x == nil {
		brackets := ""
		switch {
		case adaValue.PeriodIndex() > 0 && adaValue.MultipleIndex() > 0:
			brackets = fmt.Sprintf("[%02d,%02d]", adaValue.PeriodIndex(), adaValue.MultipleIndex())
		case adaValue.PeriodIndex() > 0:
			brackets = fmt.Sprintf("[%02d]", adaValue.PeriodIndex())
		case adaValue.MultipleIndex() > 0:
			brackets = fmt.Sprintf("[%02d]", adaValue.MultipleIndex())
		default:
		}

		if adaValue.Type().IsStructure() {
			structureValue := adaValue.(*adatypes.StructureValue)
			fmt.Println(y+" "+adaValue.Type().Name()+brackets+" = [", structureValue.NrElements(), "]")
		} else {
			fmt.Printf("%s %s%s = > %s <\n", y, adaValue.Type().Name(), brackets, adaValue.String())
		}
	} else {
		buffer := x.(*bytes.Buffer)
		buffer.WriteString(fmt.Sprintln(y, adaValue.Type().Name(), "= >", adaValue.String(), "<"))
	}

	return adatypes.Continue, nil
}

func TestConnection_fieldType(t *testing.T) {
	err := initLogWithFile("field_type.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		fmt.Println("Error creating database connection", cerr)
		return
	}
	defer connection.Close()

	fmt.Println(connection)
	openErr := connection.Open()
	if !assert.NoError(t, openErr) {
		fmt.Println("Error opening database", openErr)
		return
	}
	request, err := connection.CreateFileReadRequest(270)
	if !assert.NoError(t, err) {
		fmt.Println("Error creating read request", err)
		return
	}
	err = request.QueryFields("IT,BB,TY,AA,WC,PI,UI")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}
	request.Limit = 0
	request.RecordBufferShift = 64000
	result, rerr := request.ReadLogicalWith("AF=" + recordNamePrefix)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error reading records", rerr)
		return
	}
	// tv := adatypes.TraverserValuesMethods{PrepareFunction: dumpFieldTypeTestPrepare, EnterFunction: dumpFieldTypeValues}
	// _, err = result.TraverseValues(tv, nil)
	// if !assert.NoError(t, err) {
	// 	fmt.Println("Error traversing records", err)
	// 	return
	// }
	fmt.Println("Endian:", Endian())
	err = validateResult(t, "field_Types_"+Endian().String(), result)
	assert.NoError(t, err)
}
