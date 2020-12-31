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
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestConnection_readMainframe(t *testing.T) {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(1)
	if err != nil {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[11100301:11100303]")
	if !assert.NoError(t, err) {
		fmt.Println("Error reading", err)
		return
	}
	assert.NotNil(t, result)
	fmt.Println("Result data:")
	result.DumpValues()
	validateResult(t, "readMainframe", result)
}
func TestConnection_readBorderMainframe(t *testing.T) {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if !assert.NoError(t, cerr) {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(1)
	if !assert.NoError(t, err) {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	request.QueryFields("AA,AB")
	request.Limit = 2
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=(11100301:11100303)")
	if !assert.NoError(t, err) {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	validateResult(t, "readBorderMainframe", result)
}

func TestConnection_readNoMinimumMainframe(t *testing.T) {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if assert.NoError(t, cerr) {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(1)
	if assert.NoError(t, err) {
		fmt.Println("Error creating read request : ", err)
		return
	}
	assert.NotNil(t, connection)
	assert.NotNil(t, request)
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	err = request.QueryFields("AA,AB")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 2
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=(11100301:11100303]")
	if !assert.NoError(t, err) {
		fmt.Println("Error reading", err)
		return
	}
	validateResult(t, "readNoMinimumMainframe", result)
}

func TestConnection_readNoMaximumMainframe(t *testing.T) {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if !assert.NoError(t, cerr) {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateFileReadRequest(1)
	if !assert.NoError(t, err) {
		fmt.Println("Error creating read request : ", err)
		return
	}
	fmt.Println("Connection : ", connection)

	fmt.Println("Limit query data:")
	err = request.QueryFields("AA,AB")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields : ", err)
		return
	}
	request.Limit = 2
	fmt.Println("Read logical data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[1100301:11100303)")
	if !assert.NoError(t, err) {
		fmt.Println("Error reading", err)
		return
	}
	assert.NotNil(t, result)
	if result == nil {
		fmt.Println("Result empty")
		return
	}
	if result.Values == nil {
		fmt.Println("Values empty")
		return
	}
	validateResult(t, "readNoMaximumMainframe", result)

}

func TestConnection_periodGroupMfPart(t *testing.T) {
	initLogWithFile("connection.log")

	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if !assert.NoError(t, cerr) {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if !assert.NoError(t, openErr) {
		fmt.Println("Error open connection", openErr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if !assert.NoError(t, err) {
		fmt.Println("Error create request", err)
		return
	}
	err = request.QueryFields("personnnel-id,income")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("personnnel-id=[11100303:11100304]")
	if !assert.NoError(t, err) {
		fmt.Println("Error create request", err)
		return
	}
	validateResult(t, "periodGroupMfPart", result)
}

func TestConnectionPEMUMfMap(t *testing.T) {
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
		t.Skip("Mainframe database not defined, skipping ...")
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
		request.QueryFields("*")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("personnnel-id=11100301")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 1, len(result.Values)) {
			ae := result.Values[0].HashFields["name"]
			fmt.Println("Check BERGMANN ...")
			assert.Equal(t, "BERGMANN", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}

}

func TestConnectionPEShiftMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mainframe tests in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
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
		request.QueryFields("*")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("personnnel-id=[11400327:11500303]")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 5, len(result.Values)) {
			ae := result.Values[1].HashFields["name"]
			fmt.Println("Check SCHILLING ...")
			assert.Equal(t, "SCHILLING", strings.TrimSpace(ae.String()))
			ae = result.Values[2].HashFields["name"]
			val := result.Values[2]
			fmt.Println("Check FREI ...")
			assert.Equal(t, "FREI", strings.TrimSpace(ae.String()))
			nv, _ := val.searchValue("name")
			assert.Equal(t, "FREI", strings.TrimSpace(nv.String()))
			assert.Equal(t, int32(3), val.ValueQuantity("income"))
			assert.Equal(t, int32(3), val.ValueQuantity("bonus"))
			assert.Equal(t, int32(0), val.ValueQuantity("bonus", 2))
			assert.Equal(t, int32(0), val.ValueQuantity("bonus", 3))
			_, aerr := val.SearchValueIndex("bonus", []uint32{3, 12})
			assert.Error(t, aerr)
		}
	}

}

func TestConnectionPEShiftMfMapShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mainframe tests in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		t.Skip("Mainframe database not defined, skipping ...")
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
		request.QueryFields("personnnel-id,income,leave-date")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("personnnel-id=[11400327:11500303]")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 5, len(result.Values)) {
			ae := result.Values[1].HashFields["personnnel-id"]
			fmt.Println("Check 11400328 ...")
			assert.Equal(t, "11400328", strings.TrimSpace(ae.String()))
			ae = result.Values[2].HashFields["personnnel-id"]
			val := result.Values[2]
			fmt.Println("Check 11500301 ...")
			assert.Equal(t, "11500301", strings.TrimSpace(ae.String()))
			nv, _ := val.searchValue("personnnel-id")
			assert.Equal(t, "11500301", strings.TrimSpace(nv.String()))
			assert.Equal(t, int32(3), val.ValueQuantity("income"))
			assert.Equal(t, int32(3), val.ValueQuantity("bonus"))
			assert.Equal(t, int32(0), val.ValueQuantity("bonus", 2))
			assert.Equal(t, int32(0), val.ValueQuantity("bonus", 3))
			_, aerr := val.SearchValueIndex("bonus", []uint32{3, 12})
			assert.Error(t, aerr)
		}
	}

}

func TestConnectionAllMfMap(t *testing.T) {
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
		t.Skip("Mainframe database not defined, skipping ...")
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
		request.QueryFields("*")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalBy("personnnel-id")
		assert.NoError(t, err)
		// fmt.Println("Result data:")
		// result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 1107, len(result.Values)) {
			ae := result.Values[1].HashFields["name"]
			fmt.Println("Check SCHIRM ...")
			assert.Equal(t, "SCHIRM", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
			ae = result.Values[1106].HashFields["name"]
			fmt.Println("Check OSEA ...")
			assert.Equal(t, "OSEA", strings.TrimSpace(ae.String()))
		}
	}

}
