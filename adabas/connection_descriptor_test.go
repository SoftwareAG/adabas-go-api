/*
* Copyright © 2019-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	connection.Open()
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
	readRequest.QueryFields("AU,AV")

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

	connection.Open()
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
		request.QueryFields("*")
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
			validateResult(t, "descriptorinmap", result)
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
			validateResult(t, "descriptorinmapwithquery", result)
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
	result, rerr := readRequest.HistogramBy("S3")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("S3")
	fmt.Println(result)
}
