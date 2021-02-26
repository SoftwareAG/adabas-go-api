/*
* Copyright © 2020-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"strings"
	"testing"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestHoldResponse(t *testing.T) {
	initTestLogWithFile(t, "connection_hold.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	wait := make(chan bool)
	end := make(chan bool)
	go parallelAccessHoldResponse(t, wait, end, false)

	connection, err := NewConnection("ada;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.SetHoldRecords(adatypes.HoldResponse)
	err = readRequest.QueryFields("AA")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Waiting for hold thread ....")
	w := <-wait
	fmt.Println("Read hold ....")
	_, rerr := readRequest.ReadISN(1)
	if !assert.Error(t, rerr) {
		return
	}
	fmt.Println("Got error", rerr)
	assert.True(t, strings.HasPrefix(rerr.Error(), "ADAGE91000:"))
	end <- true
	for w {
		select {
		case <-time.After(20 * time.Second):
			fmt.Println("timeout received")
			assert.Fail(t, "timeout received")
			w = false
		case w = <-wait:
			fmt.Println("wait received")
		case e := <-end:
			assert.True(t, e)
			w = false
		}
	}
}

func TestHoldRead(t *testing.T) {
	initTestLogWithFile(t, "connection_hold.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	wait := make(chan bool)
	end := make(chan bool)
	go parallelAccessHoldResponse(t, wait, end, true)

	connection, err := NewConnection("ada;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.SetHoldRecords(adatypes.HoldAccess)
	err = readRequest.QueryFields("AA")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Waiting for hold thread ....")
	<-wait
	fmt.Println("Read hold ....")
	_, rerr := readRequest.ReadLogicalWith("AA=50005800")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Wait timeout ended ....")
	select {
	case e := <-end:
		assert.True(t, e, "Wrong end")
	case <-time.After(5 * time.Second):
		end <- true
	}
	fmt.Println("Wait hold thread ended ....")
	<-end
}

func parallelAccessHoldResponse(t *testing.T, wait chan bool, end chan bool, useTimeout bool) {
	fmt.Println("Start hold access ....")
	connection, err := NewConnection("ada;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.SetHoldRecords(adatypes.HoldResponse)
	err = readRequest.QueryFields("AA")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Read in hold ISN 1 ....")
	for i := adatypes.Isn(1); i < 10; i++ {
		_, rerr := readRequest.ReadISN(i)
		if !assert.NoError(t, rerr) {
			fmt.Println("Error parallel access.", rerr)
			end <- false
			return
		}
	}
	fmt.Println("In hold ISN 1 ....")
	wait <- true
	if useTimeout {
		fmt.Println("Sleep 10 seconds ....")
		time.Sleep(10 * time.Second)
		connection.Release()
		fmt.Println("Release ....")
	}
	<-end

	fmt.Println("End parallel access.")
	end <- true
}
