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
	"runtime"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestConnectionSecure_fail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}

	initTestLogWithFile(t, "connection_secure.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=25;auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}

	request, rerr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("AA")
	if assert.Error(t, err) {
		assert.Equal(t, "ADAGEC801F: Security violation: Authentication error (rsp=200,subrsp=31,dbid=25,file=0)", err.Error())
	}
}

func TestConnectionSecure_pwd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "connection_secure.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=25;auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	connection.AddCredential("hkaf", "dummy1")

	request, rerr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("AA,AE")

	if !assert.NoError(t, err) {
		fmt.Println("Error query fields for request", err)
		return
	}
	request.Limit = 0
	fmt.Println("Read logigcal data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[11100315:11100316]")
	if err != nil {
		fmt.Println("Error read logical data", err)
		return
	}
	result.DumpValues()
	// Output: XX
}
