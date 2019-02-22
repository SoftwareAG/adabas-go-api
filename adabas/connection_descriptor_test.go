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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConnectionComplexSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_descriptor.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	f := initTestLogWithFile(t, "connection_descriptor.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	f := initTestLogWithFile(t, "connection_descriptor.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
