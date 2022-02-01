/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"strings"
	"sync"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestConnectionComplexSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=" + adabasModDBIDs + ";auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	err = connection.Open()
	if !assert.NoError(t, err) {
		return
	}
	readRequest, rErr := connection.CreateFileReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("AA,AB")

	adatypes.Central.Log.Debugf("Test Search complex with ...")
	result, rerr := readRequest.ReadLogicalWith("AA=[11100301:11100305] AND AE='SMITH'")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Complex search done")
	fmt.Println(result)
}

func TestConnectionSuperDescriptor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("AU,AV")
	assert.NoError(t, err)

	adatypes.Central.Log.Debugf("Test Search complex with ...")
	result, rerr := readRequest.ReadLogicalBy("S1")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Super Descriptor read done")
	fmt.Println(result.String())
}

func TestConnectionSuperDescSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs + ";auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	err = connection.Open()
	if !assert.NoError(t, err) {
		return
	}
	readRequest, rErr := connection.CreateFileReadRequest(16)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("AA,AB")
	assert.NoError(t, err)

	adatypes.Central.Log.Debugf("Test Search complex with ...")
	result, rerr := readRequest.ReadLogicalWith("S2=['BADABAS__'0:'BADABAS__'255]")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Complex search done")
	fmt.Println(result)
}

func TestConnectionDescriptorinMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_descriptor.log")

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
		err = request.QueryFields("*")
		assert.NoError(t, err)
		request.Limit = 5
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalBy("DEPARTMENT")
		if !assert.NoError(t, err) {
			return
		}
		// result.DumpValues()
		ae := result.Values[0].HashFields["DEPARTMENT"]
		fmt.Println("Check DEPARTMENT ...")
		if assert.NotNil(t, ae) {
			assert.Equal(t, "ADMA", strings.TrimSpace(ae.String()))
			_ = validateResult(t, "descriptorinmap", result)
		}
	}

}

func TestConnectionDescriptorinMapWithQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_descriptor.log")

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
		err = request.QueryFields("*")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 5
		fmt.Println("Read logigcal data:")
		result, err := request.HistogramBy("DEPARTMENT")
		if !assert.NoError(t, err) {
			return
		}
		ae := result.Values[0].HashFields["DEPARTMENT"]
		fmt.Println("Check DEPARTMENT ...")
		if assert.NotNil(t, ae) {
			assert.Equal(t, "ADMA", strings.TrimSpace(ae.String()))
			_ = validateResult(t, "descriptorinmapwithquery", result)
		}
	}

}

func TestConnectionSuperDescriptorinMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_descriptor.log")

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
		request.Limit = 5
		fmt.Println("Read logigcal data:")
		result, err := request.HistogramBy("DEPARTMENT")
		assert.NoError(t, err)
		ae := result.Values[0].HashFields["DEPARTMENT"]
		fmt.Println("Check DEPARTMENT ...")
		if assert.NotNil(t, ae) {
			assert.Equal(t, "ADMA", strings.TrimSpace(ae.String()))
			assert.Equal(t, uint64(8), result.Values[0].Quantity)
			validateResult(t, "superdescriptorinmap", result)
		}
	}

}

var wg sync.WaitGroup

func TestConnectionDescriptorinMapSuper(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	var testcases = [][]string{{"searchandorder", "NAME='SMITH'", "DEPARTMENT"},
		{"searchnoorder", "NAME='SMITH'", ""},
		{"onlyorder", "", "DEPARTMENT"}}
	for _, s := range testcases {
		wg.Add(1)
		go testSearchAndOrder(t, s[0], s[1], s[2])
	}
	wg.Wait()
}

func testSearchAndOrder(t *testing.T, name, search, sortedby string) {
	defer wg.Done()
	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, err, name) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("FULL-NAME,DEPARTMENT")
		if !assert.NoError(t, err, name) {
			return
		}
		request.Limit = 5
		fmt.Println("Read logigcal data:")
		result, err := request.SearchAndOrder(search, sortedby)
		if !assert.NoError(t, err, name) {
			return
		}
		validateResult(t, name, result)
	}

}

func TestConnectionSuperDescriptors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=" + adabasStatDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.Limit = 4
	result, rerr := readRequest.HistogramBy("S3")
	if !assert.NoError(t, rerr) {
		return
	}
	// fmt.Println("S3")
	// fmt.Println(result.String())
	if assert.Len(t, result.Values, 4) {
		assert.Equal(t, "ISN=2 quantity=1\n S3=\"'DKK' 100000\"\n", result.Values[0].String())
		assert.Equal(t, "ISN=5 quantity=1\n S3=\"'DKK' 140000\"\n", result.Values[3].String())
	}
}

func TestDescriptor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection_descriptor.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("ada;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	readRequest, rErr := connection.CreateFileReadRequest(270)
	if !assert.NoError(t, rErr) {
		return
	}
	readRequest.QueryFields("IT,S1,U1,S2,U2,S4,U4,S8,U8,BB,BR,B1,TY,F4,F8,AA,A1,AS,A2,AB,AF,WC,WU,WL,W4,WF,PI,PA,PF,UI,UP,UF,UE,ZB,S3,SU")
	readRequest.Limit = 4
	result, rerr := readRequest.ReadPhysicalSequence()
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("LEN:", len(result.Values))
	if assert.Len(t, result.Values, 4) {
		assert.Equal(t, "ISN=4 quantity=0\n IT=\"\"\n S1=\"0\"\n U1=\"0\"\n S2=\"0\"\n U2=\"0\"\n S4=\"0\"\n U4=\"0\"\n S8=\"0\"\n U8=\"0\"\n BB=\"\"\n BR=\"[0]\"\n B1=\"[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]\"\n TY=\"\"\n F4=\"0.000000\"\n F8=\"0.000000\"\n AA=\"\"\n A1=\" \"\n AS=\" \"\n A2=\" \"\n AB=\" \"\n AF=\"HUMBERTO            \"\n WC=\"\"\n WU=\" \"\n WL=\" \"\n W4=\" \"\n WF=\"MORENO                                            \"\n PI=\"\"\n PA=\"0\"\n PF=\"2\"\n UI=\"\"\n UP=\"0\"\n UF=\"0\"\n UE=\"0\"\n ZB=\"\"\n ZB=\"20190207140701\"\n S3=\"'HUM' 2\"\n SU=\"'HUM' 2 MOR\"\n", result.Values[3].String())
	}
}
