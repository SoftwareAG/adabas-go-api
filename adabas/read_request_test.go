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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

func TestRequestPhysicalSimpleTypes(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result, err := request.ReadPhysicalSequence()
	if assert.NoError(t, err) {
		result.DumpValues()
	}
}

func TestRequestPhysicalMultipleField(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	err := request.QueryFields("AA,AC,AD,AZ")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, request.definition) {
		return
	}
	request.definition.DumpTypes(false, true)
	result, rErr := request.ReadPhysicalSequence()
	if assert.NoError(t, rErr) {
		result.DumpValues()
	}
}

func TestRequestLogicalWithQueryFields(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result, err := request.ReadLogicalWith("AA=60010001")
	request.Close()
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		result.DumpValues()
	}
}

func TestRequestLogicalWithFields(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	result, err := request.ReadLogicalWith("AA=60010001")
	request.Close()
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		result.DumpValues()
	}
}

func TestReadRequestLogicalBy(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result, err := request.ReadLogicalBy("AA")
	fmt.Println("Dump result received ...")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		result.DumpValues()
	}
}

func TestReadRequestLogicalByAll(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.Limit = 2
	result, err := request.ReadLogicalBy("AA")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Dump result received ...")
	if result != nil {
		result.DumpValues()
	}
	if !assert.Equal(t, 2, len(result.Values)) {
		t.Fatalf("Occurens of result does not fit %d!=2", len(result.Values))
		return
	}
	v := result.Values[0].HashFields["AJ"]
	if !assert.NotNil(t, v) {
		return
	}
	assert.Equal(t, "HEPPENHEIM          ", v.String())
	v = result.Values[1].HashFields["AJ"]
	assert.Equal(t, "DARMSTADT           ", v.String())
	v = result.Values[1].HashFields["AZ"]
	assert.Equal(t, adatypes.FieldTypeMultiplefield, v.Type().Type())
}

func TestRequestRemoteLogicalByAll(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	ID := NewAdabasID()
	adabas, aerr := NewAdabasWithID(url, ID)
	if !assert.NoError(t, aerr) {
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.Limit = 2
	_, err := request.ReadLogicalBy("AA")
	fmt.Println("Dump result received ...")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000068: Entire Network client not supported, use port 0 and Entire Network native access", err.Error())
}

func ExampleReadRequest_ReadLogicalBy() {
	err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.Limit = 2
	request.QueryFields("AA,AC,AD")
	var result *Response
	result, err = request.ReadLogicalBy("AA")
	fmt.Println("Dump result received ...")
	if result != nil {
		result.DumpValues()
	}

	// Output:
	// Dump result received ...
	// Dump all result values
	// Record Isn: 0204
	//   AA = > 11100102 <
	//   AB = [ 1 ]
	//    AC = > EDGAR                <
	//    AD = > PETER                <
	// Record Isn: 0205
	//   AA = > 11100105 <
	//   AB = [ 1 ]
	//    AC = > CHRISTIAN            <
	//    AD = >                      <
}

func TestReadRequestLogicalBySuperDescriptor(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result, err := request.ReadLogicalBy("S1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
	}
}

func TestReadRequestAllJson(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 16)
	defer request.Close()
	request.QueryFields("*")
	request.Limit = 1
	result, err := request.ReadLogicalBy("AA")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// result.DumpValues()
	res, jerr := json.Marshal(result)
	if jerr != nil {
		fmt.Println("Error generating JSON", jerr)
		return
	}
	fmt.Println(string(res))

}

func TestReadRequestHistogramDescriptorField(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	request.Limit = 10
	defer request.Close()
	result, err := request.HistogramBy("AA")
	assert.NoError(t, err)
	if !assert.NotNil(t, result) {
		return
	}
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
	}
	assert.Equal(t, "11100102", result.Values[0].Value[0].String())
	assert.Equal(t, "11100105", result.Values[1].Value[0].String())
	assert.Equal(t, "11100113", result.Values[9].Value[0].String())
}

func TestReadRequestHistogramSuperDescriptor(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.Limit = 10
	result, err := request.HistogramBy("S1")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	// if result != nil {
	// 	fmt.Println("Dump result received ...")
	// 	result.DumpValues()
	// }
	assert.Equal(t, 10, len(result.Values))
	adatypes.Central.Log.Debugf("Index  1 %p", result.Values[0].Value[0])
	adatypes.Central.Log.Debugf("Index  2 %p", result.Values[1].Value[0])
	adatypes.Central.Log.Debugf("Index 10 %p", result.Values[9].Value[0])
	assert.Equal(t, "ADMA", result.Values[0].Value[0].String())
	assert.Equal(t, "COMP", result.Values[1].Value[0].String())
	assert.Equal(t, "SYSA", result.Values[9].Value[0].String())
}

func ExampleReadRequest_histogramWith() {
	err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	result, err := request.HistogramWith("AA=20010100")
	if err == nil {
		if result != nil {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	} else {
		fmt.Println(err)
	}
	// Output:
	// Dump result received ...
	// Dump all result values
	// Record Quantity: 0001
	//   AA = > 20010100 <
}

func TestReadRequestReadMap(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	request, _ := NewReadRequest(adabas, 4)
	defer request.Close()
	result, err := request.ReadLogicalWith("RN='EMPLOYEES-NAT-DDM'")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, result)
	if result != nil {
		// fmt.Println("Dump result received ...")
		// result.DumpValues()
		assert.Equal(t, 1, len(result.Values))
	} else {
		fmt.Println("Error result nil ...")
	}
}

func TestReadRequestMissingFile(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	request, _ := NewReadRequest(adabas, 123)
	defer request.Close()
	_, err := request.ReadLogicalWith("RN='EMPLOYEES-NAT-DDM'")
	fmt.Println("Read done ...")
	assert.Error(t, err)
}

func dumpStream(record *Record, x interface{}) error {
	i := x.(*uint32)
	a, _ := record.SearchValue("AE")
	fmt.Printf("Read %d -> %s = %d\n", record.Isn, a, record.Quantity)
	(*i)++
	return nil
}

func TestReadRequestWithStream(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	i := uint32(0)
	result, err := request.ReadLogicalWithStream("AE='SMITH'", dumpStream, &i)
	fmt.Println("Read done ...")
	assert.NoError(t, err)
	assert.Equal(t, uint32(19), i)
	if assert.NotNil(t, result) {
		result.DumpValues()
	}
}

func ExampleReadRequest_histogramWithStream() {
	err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println("Error init log", err)
		return
	}

	adabas, _ := NewAdabas(24)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	i := uint32(0)
	result, err := request.HistogramWithStream("AE='SMITH'", dumpStream, &i)
	fmt.Println("Read done ...")
	if err != nil {
		fmt.Println("Error reading histogram", err)
		return
	}
	if i != 1 {
		fmt.Println("Index error", i)
	}
	if result != nil {
		result.DumpValues()
		fmt.Println("Result set should be empty")
	}

	// Output: Read 0 -> SMITH                = 19
	// Read done ...
	// Dump all result values
	// Result set should be empty
}

func TestReadRequestPhysicalStream(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	//request.QueryFields("AE")
	i := uint32(0)
	result, err := request.ReadPhysicalSequenceStream(dumpStream, &i)
	fmt.Println("Read done ...")
	assert.NoError(t, err)
	assert.Equal(t, uint32(20), i)
	if assert.NotNil(t, result) {
		result.DumpValues()
	}
}

func BenchmarkReadRequest_Small(b *testing.B) {
	err := initLogLevelWithFile("request-bench.log", "error")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	defer request.Close()
	request.Limit = 0
	request.QueryFields("AA,AC,AD")

	var result *Response
	result, err = request.ReadLogicalBy("AA")
	//fmt.Println("Dump result received ...")
	if result != nil {
		assert.Equal(b, 1107, result.NrRecords())
		// result.DumpValues()
	}
}
func BenchmarkReadRequest(b *testing.B) {
	err := initLogLevelWithFile("request-bench.log", "error")

	assert.NoError(b, err)

	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 11)
	request.Limit = 0
	defer request.Close()
	var result *Response
	result, err = request.ReadLogicalBy("AA")
	assert.NoError(b, err)
	assert.NotNil(b, result)
	if result != nil {
		//fmt.Println("Dump result received ...")
		assert.Equal(b, 1107, result.NrRecords())
		// result.DumpValues()
	}

}

func TestRequestWithMapLogicalBy(t *testing.T) {
	//	f, err := initLogLevelWithFile("request.log", adatypes.Central.Log.DebugLevel)
	err := initLogWithFile("request.log")
	if err != nil {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest("EMPLOYEES-NAT-DDM", adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, request) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		err = request.QueryFields("PERSONNEL-ID,FIRST-NAME,NAME")
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, request.IsOpen())

		fmt.Println("After query fields")
		var result *Response
		result, err = request.ReadLogicalBy("PERSONNEL-ID")
		if assert.NoError(t, err) {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	}
}

func traverseFieldCounter(IAdaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	fi := x.(*int)
	*fi++
	fmt.Println("A")
	return nil
}

func TestRequestWithMapRepositoryLogicalBy(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(24)
	AddGlobalMapRepository(ada.URL, 4)
	defer DelGlobalMapRepository(ada.URL, 4)

	request, err := NewReadRequest(ada, "EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	fmt.Println("Open database ...", openErr)
	fmt.Printf("Status ...%#v", request.adabas.status)
	assert.NotNil(t, request.adabas.status.platform)
	if assert.NoError(t, openErr) {
		err = request.QueryFields("PERSONNEL-ID,FIRST-NAME,NAME")
		if err != nil {
			return
		}
		fsize := 0
		tm := adatypes.NewTraverserMethods(traverseFieldCounter)
		request.TraverseFields(tm, &fsize)
		assert.Equal(t, 4, fsize)

		fmt.Println("After query fields")
		fmt.Printf("Status ...%#v", request.adabas.status)
		assert.NotNil(t, request.adabas.status.platform)
		result, err := request.ReadLogicalBy("PERSONNEL-ID")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	}
}

func TestRequestWithMapDirectRepositoryLogicalBy(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(24)
	request, err := NewReadRequest("EMPLOYEES-NAT-DDM", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		err = request.QueryFields("PERSONNEL-ID,FIRST-NAME,NAME")
		if err != nil {
			return
		}
		result, err := request.ReadLogicalBy("PERSONNEL-ID")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, 20, len(result.Values))
		}
	}
}

func TestReadMaps(t *testing.T) {
	initTestLogWithFile(t, "request.log")
	ada, _ := NewAdabas(24)
	request, _ := NewReadRequest(ada, 4)
	request.Limit = 0
	defer ada.Close()
	request.QueryFields(mapFieldName.fieldName())
	result, err := request.ReadLogicalBy(mapFieldName.fieldName())
	if err != nil {
		return
	}
	result.DumpValues()
}

func TestMapRequestWithHistogramBy(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(24)
	request, err := NewReadRequest("EMPLOYEES-NAT-DDM", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		result, err := request.HistogramBy("DEPARTMENT")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, 13, len(result.Values))
			assert.Equal(t, uint64(8), result.Values[0].Quantity)
			assert.Equal(t, uint64(95), result.Values[12].Quantity)
		}
	}
}

func TestMapRequestWithHistogramWith(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(24)
	request, err := NewReadRequest("EMPLOYEES-NAT-DDM", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		result, err := request.HistogramWith("DEPARTMENT=ADMA")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, 1, len(result.Values))
			assert.Equal(t, uint64(8), result.Values[0].Quantity)
		}
	}
}

func TestMapRequestFractional(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(24)
	request, err := NewReadRequest("Fractional", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	if assert.NoError(t, openErr) {
		result, err := request.ReadLogicalBy("FRACT1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, 7, len(result.Values))
			result.DumpValues()
			x, serr := result.Values[0].SearchValue("FRACT1")
			assert.NoError(t, serr)
			assert.Equal(t, "1.44", x.String())
		}
	}
}

func ExampleReadRequest_ReadPhysical() {
	err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 225)
	defer request.Close()
	request.Limit = 2
	request.QueryFields("*")
	var result *Response
	result, err = request.ReadPhysicalSequence()
	fmt.Println("Dump result received ...")
	if result != nil {
		result.DumpValues()
	}

	// Output:
	// Dump result received ...
	// Dump all result values
	// Record Isn: 0204
	//   AA = > 11100102 <
	//   AB = [ 1 ]
	//    AC = > EDGAR                <
	//    AD = > PETER                <
	// Record Isn: 0205
	//   AA = > 11100105 <
	//   AB = [ 1 ]
	//    AC = > CHRISTIAN            <
	//    AD = >                      <
}

func TestReadPElevel2Group(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	request, _ := NewReadRequest(adabas, 225)
	defer request.Close()
	request.QueryFields("AD")
	result, err := request.ReadPhysicalSequence()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
	}
}
