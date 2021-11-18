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
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestConnectionMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("*")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 20
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalBy("personnnel-id")
		if !assert.NoError(t, err) {
			return
		}
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 20, len(result.Values)) {
			ae := result.Values[1].HashFields["name"]
			fmt.Println("Check SCHIRM ...")
			assert.Equal(t, "SCHIRM", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
			ae = result.Values[19].HashFields["name"]
			fmt.Println("Check BLAU ...")
			assert.Equal(t, "BLAU", strings.TrimSpace(ae.String()))
			validateResult(t, "mfread", result)
		}
	}

}

func TestConnectionSearchMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("full-name")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("name=SMITH")
		assert.NoError(t, err)
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 19, len(result.Values)) {
			ae := result.Values[1].HashFields["middle-name"]
			fmt.Println("Check C. ...")
			assert.Equal(t, "C.", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
			ae = result.Values[17].HashFields["middle-name"]
			fmt.Println("Check RODNEY ...")
			assert.Equal(t, "RODNEY", strings.TrimSpace(ae.String()))
			validateResult(t, "mfsearch", result)
		}
	}

}

func TestConnectionAndSearchMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		err = request.QueryFields("full-name")
		if !assert.NoError(t, err) {
			return
		}
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("name>'ADAM' AND name<'AECKERLE'")
		assert.NoError(t, err)
		validateResult(t, "mfandsearch", result)
	}

}

func TestConnectionRangeSearchMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("full-name")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("name=(HELL:HERZOG)")
		assert.NoError(t, err)
		validateResult(t, "mfrangesearch", result)
	}

}

func TestConnectionHistogramMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("full-name")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.HistogramWith("name=['U':'W']")
		assert.NoError(t, err)
		validateResult(t, "mfhistogram", result)
	}

}

func TestConnectionMfTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	fmt.Println("Connect to ", network)
	connection, cerr := NewConnection("acj;target=" + network)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateFileReadRequest(1)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("AA,AB,AS[N]")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AE=SMITH")
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 19, len(result.Values))
		assert.Equal(t, 19, result.NrRecords())
		// err = result.DumpValues()
		// assert.NoError(t, err)
		kaVal := result.Values[0].HashFields["AE"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "SMITH               ", kaVal.String())
		}
		kaVal = result.Values[9].HashFields["AA"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "20001000", kaVal.String())
		}

		record := result.Isn(1106)
		assert.NotNil(t, record)
		validateResult(t, t.Name(), result)
	}
}

func TestConnectionMfTestSuiteSalary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	fmt.Println("Connect to ", network)
	connection, cerr := NewConnection("acj;target=" + network)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if !assert.NoError(t, openErr) {
		return
	}
	fmt.Println("Adabas connection opened to ", connection.GetAdabasInformation())
	assert.Contains(t, connection.GetAdabasInformation(), "Mainframe,High Order")

	request, err := connection.CreateFileReadRequest(1)
	if !assert.NoError(t, err) {
		return
	}
	request.QueryFields("AA,AB,AS,AT")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("AA=11100301 OR AA=11222222")
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 2, len(result.Values))
		assert.Equal(t, 2, result.NrRecords())
		kaVal := result.Values[0].HashFields["AE"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "BERGMANN            ", kaVal.String())
		}
		kaVal = result.Values[1].HashFields["AA"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "11222222", kaVal.String())
		}

		record := result.Isn(251)
		assert.NotNil(t, record)

		validateResult(t, t.Name(), result)
	}
}
