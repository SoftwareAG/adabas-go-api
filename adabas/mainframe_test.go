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
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func ExampleConnection_readMainframe() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
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
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0251
	// Record Quantity: 0003
	//   AA = > 11100301 <
	//   AB = [ 1 ]
	//    AC = > HANS                 <
	//    AE = > BERGMANN             <
	//    AD = > WILHELM              <
	// Record Isn: 0383
	// Record Quantity: 0003
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <

}

func ExampleConnection_readBorderMainframe() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
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
	result, err = request.ReadLogicalWith("AA=(11100301:11100303)")
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0383
	// Record Quantity: 0001
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <

}

func ExampleConnection_readNoMinimumMainframe() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
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
	result, err = request.ReadLogicalWith("AA=(11100301:11100303]")
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0383
	// Record Quantity: 0002
	//   AA = > 11100302 <
	//   AB = [ 1 ]
	//    AC = > ROSWITHA             <
	//    AE = > HAIBACH              <
	//    AD = > ELLEN                <
	// Record Isn: 0252
	// Record Quantity: 0002
	//   AA = > 11100303 <
	//   AB = [ 1 ]
	//    AC = > KRISTINA             <
	//    AE = > FALTER               <
	//    AD = > MARIA                <

}

func ExampleConnection_readNoMaximumMainframe() {
	initLogWithFile("mainframe.log")
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;target=" + network)
	if cerr != nil {
		fmt.Println("Connection to database error:", cerr)
		return
	}
	defer connection.Close()
	request, err := connection.CreateReadRequest(1)
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
	result, err = request.ReadLogicalWith("AA=[1100301:11100303)")
	if err != nil {
		fmt.Println("Error reading", err)
		return
	}
	fmt.Println("Result data:")
	result.DumpValues()
	// Output: Connection :  Adabas url=54712 fnr=0
	// Limit query data:
	// Read logical data:
	// Result data:
	// Dump all result values
	// Record Isn: 0204
	// Record Quantity: 0017
	//   AA = > 11100102 <
	//   AB = [ 1 ]
	//    AC = > EDGAR                <
	//    AE = > SCHINDLER            <
	//    AD = > PETER                <
	// Record Isn: 0205
	// Record Quantity: 0017
	//   AA = > 11100105 <
	//   AB = [ 1 ]
	//    AC = > CHRISTIAN            <
	//    AE = > SCHIRM               <
	//    AD = >                      <

}

func ExampleConnection_periodGroupMfPart() {
	f, _ := initLogWithFile("connection.log")
	defer f.Close()

	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	connection, cerr := NewConnection("acj;map;config=[" + network + ",4]")
	if cerr != nil {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error open connection", cerr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	request.QueryFields("personnnel-id,income")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("personnnel-id=[11100303:11100304]")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = result.DumpValues()
	if err != nil {
		fmt.Println("Error dump values", err)
	}

	// Output: Dump all result values
	// Record Isn: 0252
	// Record Quantity: 0002
	//   personnnel-id = > 11100303 <
	//   income = [ 3 ]
	//    curr-code[01] = > DM  <
	//    salary[01] = > 42600 <
	//    bonus[01] = [ 2 ]
	//     bonus[01,01] = > 3350 <
	//     bonus[01,02] = > 3000 <
	//    curr-code[02] = > DM  <
	//    salary[02] = > 41000 <
	//    bonus[02] = [ 1 ]
	//     bonus[02,01] = > 3000 <
	//    curr-code[03] = > DM  <
	//    salary[03] = > 39600 <
	//    bonus[03] = [ 1 ]
	//     bonus[03,01] = > 2500 <
	// Record Isn: 0253
	// Record Quantity: 0002
	//   personnnel-id = > 11100304 <
	//   income = [ 2 ]
	//    curr-code[01] = > DM  <
	//    salary[01] = > 49200 <
	//    bonus[01] = [ 2 ]
	//     bonus[01,01] = > 4400 <
	//     bonus[01,02] = > 2000 <
	//    curr-code[02] = > DM  <
	//    salary[02] = > 47000 <
	//    bonus[02] = [ 1 ]
	//     bonus[02,01] = > 3800 <

}

func TestConnectionPEMUMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	log.Debug("Created connection : ", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("*")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("personnnel-id=11100301")
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 1092, len(result.Values)) {
			ae := result.Values[1].HashFields["name"]
			fmt.Println("Check HAIBACH ...")
			assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}

}

func TestConnectionPEShiftMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	log.Debug("Created connection : ", connection)
	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-MF")
	if assert.NoError(t, err) {
		fmt.Println("Limit query data:")
		request.QueryFields("personnnel-id,language")
		request.Limit = 0
		fmt.Println("Read logigcal data:")
		result, err := request.ReadLogicalWith("personnnel-id=[11400327:11500303]")
		assert.NoError(t, err)
		fmt.Println("Result data:")
		result.DumpValues()
		fmt.Println("Check size ...", len(result.Values))
		if assert.Equal(t, 1092, len(result.Values)) {
			ae := result.Values[1].HashFields["name"]
			fmt.Println("Check HAIBACH ...")
			assert.Equal(t, "HAIBACH", strings.TrimSpace(ae.String()))
			ei64, xErr := ae.Int64()
			assert.Error(t, xErr, "Error should be send if value is string")
			assert.Equal(t, int64(0), ei64)
		}
	}

}

func TestConnectionAllMfMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	log.Debug("Created connection : ", connection)
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
