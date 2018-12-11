/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRequestPhysicalSimpleTypes(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result, err := request.ReadPhysicalSequence()
	if assert.NoError(t, err) {
		result.DumpValues()
	}
}

func TestRequestPhysicalMultipleField(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	result := &RequestResult{}
	err := request.ReadLogicalWithWithParser("AA=60010001", nil, result)
	request.Close()
	if err != nil {
		fmt.Println(err)
		assert.NoError(t, err)
	} else {
		result.DumpValues()
	}
}

func TestReadRequestLogicalBy(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result := &RequestResult{}
	err := request.ReadLogicalByWithParser("AA", nil, result)
	fmt.Println("Dump result received ...")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		result.DumpValues()
	}
}

func TestReadRequestLogicalByAll(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.Limit = 2
	result := &RequestResult{}
	err := request.ReadLogicalByWithParser("AA", nil, result)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	ID := NewAdabasID()
	adabas, aerr := NewAdabasWithID(url, ID)
	if !assert.NoError(t, aerr) {
		return
	}
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.Limit = 2
	result := &RequestResult{}
	err := request.ReadLogicalByWithParser("AA", nil, result)
	fmt.Println("Dump result received ...")
	assert.Error(t, err)
	assert.Equal(t, "Entire Network client not supported, use port 0 and Entire Network native access", err.Error())
	// if !assert.NoError(t, err) {
	// 	return
	// }
	// assert.NotNil(t, result)
	// if result != nil {
	// 	result.DumpValues()
	// }
	// if assert.Equal(t, 2, len(result.Values)) {
	// 	v := result.Values[0].HashFields["AJ"]
	// 	assert.Equal(t, "HEPPENHEIM          ", v.String())
	// 	v = result.Values[1].HashFields["AJ"]
	// 	assert.Equal(t, "DARMSTADT           ", v.String())
	// 	v = result.Values[1].HashFields["AZ"]
	// 	assert.Equal(t, adatypes.FieldTypeMultiplefield, v.Type().Type())
	// }
}

func ExampleReadRequest_ReadLogicalBy() {
	f, err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.Limit = 2
	request.QueryFields("AA,AC,AD")
	result := &RequestResult{}
	err = request.ReadLogicalByWithParser("AA", nil, result)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.QueryFields("AA,AC,AD")
	result := &RequestResult{}
	err := request.ReadLogicalByWithParser("S1", nil, result)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
	}
}

func TestReadRequestHistogramDescriptorField(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.Limit = 10
	result, err := request.HistogramBy("S1")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
	}
	assert.Equal(t, 10, len(result.Values))
	adatypes.Central.Log.Debugf("Index  1 %p", result.Values[0].Value[0])
	adatypes.Central.Log.Debugf("Index  2 %p", result.Values[1].Value[0])
	adatypes.Central.Log.Debugf("Index 10 %p", result.Values[9].Value[0])
	assert.Equal(t, "ADMA", result.Values[0].Value[0].String())
	assert.Equal(t, "COMP", result.Values[1].Value[0].String())
	assert.Equal(t, "SYSA", result.Values[9].Value[0].String())
}

func ExampleReadRequest_histogramWith() {
	f, err := initLogWithFile("request.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(24)
	request := NewRequestAdabas(adabas, 4)
	defer request.Close()
	result := &RequestResult{}
	err := request.ReadLogicalWithWithParser("RN='EMPLOYEES-NAT-DDM'", nil, result)
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, result)
	if result != nil {
		fmt.Println("Dump result received ...")
		result.DumpValues()
		assert.Equal(t, 1, len(result.Values))
	} else {
		fmt.Println("Error result nil ...")
	}
}

func TestReadRequestMissingFile(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(24)
	request := NewRequestAdabas(adabas, 123)
	defer request.Close()
	result := &RequestResult{}
	err := request.ReadLogicalWithWithParser("RN='EMPLOYEES-NAT-DDM'", nil, result)
	fmt.Println("Read done ...")
	assert.Error(t, err)
}

func BenchmarkReadRequest_Small(b *testing.B) {
	f, err := initLogLevelWithFile("request-bench.log", log.ErrorLevel)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	defer request.Close()
	request.Limit = 0
	request.QueryFields("AA,AC,AD")
	result := &RequestResult{}
	err = request.ReadLogicalByWithParser("AA", nil, result)
	fmt.Println("Dump result received ...")
	if result != nil {
		assert.Equal(b, 1107, result.NrRecords())
		// result.DumpValues()
	}
}
func BenchmarkReadRequest(b *testing.B) {
	f, err := initLogLevelWithFile("request-bench.log", log.ErrorLevel)
	defer f.Close()

	assert.NoError(b, err)

	adabas := NewAdabas(23)
	request := NewRequestAdabas(adabas, 11)
	request.Limit = 0
	defer request.Close()
	result := &RequestResult{}
	err = request.ReadLogicalByWithParser("AA", nil, result)
	assert.NoError(b, err)
	assert.NotNil(b, result)
	if result != nil {
		fmt.Println("Dump result received ...")
		assert.Equal(b, 1107, result.NrRecords())
		// result.DumpValues()
	}

}

func TestRequestWithMapLogicalBy(t *testing.T) {
	//	f, err := initLogLevelWithFile("request.log", log.DebugLevel)
	f, err := initLogWithFile("request.log")
	if err != nil {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(24)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewMapNameRequestRepo("EMPLOYEES-NAT-DDM", mapRepository)
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
		fmt.Println("After query fields")
		result := &RequestResult{}
		err = request.ReadLogicalByWithParser("PERSONNEL-ID", nil, result)
		if assert.NoError(t, err) {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	}
}

func TestRequestWithMapRepositoryLogicalBy(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	ada := NewAdabas(24)
	AddMapRepository(ada, 4)
	defer DelMapRepository(ada, 4)

	request, err := NewMapNameRequest(ada, "EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	fmt.Println("Open database ...", openErr)
	if assert.NoError(t, openErr) {
		err = request.QueryFields("PERSONNEL-ID,FIRST-NAME,NAME")
		if err != nil {
			return
		}
		fmt.Println("After query fields")
		result := &RequestResult{}
		err := request.ReadLogicalByWithParser("PERSONNEL-ID", nil, result)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	}
}

func TestRequestWithMapDirectRepositoryLogicalBy(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	request, err := NewMapNameRequestRepo("EMPLOYEES-NAT-DDM",
		NewMapRepository(NewAdabas(24), 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	openErr := request.Open()
	fmt.Println("Open database ...", openErr)
	if assert.NoError(t, openErr) {
		err = request.QueryFields("PERSONNEL-ID,FIRST-NAME,NAME")
		if err != nil {
			return
		}
		fmt.Println("After query fields")
		result := &RequestResult{}
		err := request.ReadLogicalByWithParser("PERSONNEL-ID", nil, result)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			fmt.Println("Dump result received ...")
			result.DumpValues()
		}
	}
}
