/*
* Copyright © 2020-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"sync"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestInterfaceMap(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;map;config=[23,4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeeMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.SearchAndOrder("Id=[1:2]", "LastName")
	if !assert.NoError(t, rerr) {
		return
	}
	for _, v := range response.Data {
		e := v.(*EmployeeMap)
		fmt.Printf("%s %s %T\n", e.Name, e.ID, v)
	}
}

func TestParallelStructIpc(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	var waitGroup sync.WaitGroup
	waitGroup.Add(10)
	for i := 0; i < 5; i++ {
		go callFullName(t, &waitGroup, "23")
	}
	for i := 0; i < 5; i++ {
		go callEmployees(t, &waitGroup, "23")
	}
	waitGroup.Wait()
}

func TestParallelStructTcp(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	var waitGroup sync.WaitGroup
	waitGroup.Add(10)
	for i := 0; i < 5; i++ {
		go callFullName(t, &waitGroup, "23(adatcp://localhost:60023)")
	}
	for i := 0; i < 5; i++ {
		go callEmployees(t, &waitGroup, "23")
	}
	waitGroup.Wait()
}

func callFullName(t *testing.T, waitGroup *sync.WaitGroup, dest string) {
	connection, cerr := NewConnection("acj;inmap=" + dest + ",11")
	if !assert.NoError(t, cerr) {
		waitGroup.Done()
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&FullNameInMap{})
	if !assert.NoError(t, err) {
		waitGroup.Done()
		return
	}
	request.Limit = 0
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		waitGroup.Done()
		return
	}
	response, rerr := request.SearchAndOrder("AE=SMITH", "AE")
	if !assert.NoError(t, rerr) {
		waitGroup.Done()
		return
	}
	assert.Len(t, response.Data, 19)
	for _, v := range response.Data {
		if !assert.IsType(t, &FullNameInMap{}, v) {
			return
		}
	}
	waitGroup.Done()
}

func callEmployees(t *testing.T, waitGroup *sync.WaitGroup, dest string) {
	connection, cerr := NewConnection("acj;inmap=" + dest + ",11")
	if !assert.NoError(t, cerr) {
		waitGroup.Done()
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		waitGroup.Done()
		return
	}
	request.Limit = 0
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		waitGroup.Done()
		return
	}
	a := []int{200, 507, 757, 807, 1007}
	for i := 0; i < 5; i++ {
		s := fmt.Sprintf("AA=[1:%d]", i+2)
		response, rerr := request.SearchAndOrder(s, "AE")
		if !assert.NoError(t, rerr) {
			waitGroup.Done()
			return
		}
		assert.Len(t, response.Data, a[i], s)
		for _, v := range response.Data {
			if !assert.IsType(t, &EmployeesInMap{}, v) {
				return
			}
			// e := v.(*EmployeesInMap)
			// fmt.Printf("%s %s %T\n", e.FullName.FirstName, e.ID, v)
		}

	}
	waitGroup.Done()
}

func TestInlineMapSearchAndOrderDifferentFiles(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=24")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{}, 11)
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.SearchAndOrder("AA=50005600", "AE")
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	if assert.Len(t, response.Data, 1) {
		assert.Equal(t, "50005600", response.Data[0].(*EmployeesInMap).ID)
		assert.Equal(t, "HUMBERTO            ", response.Data[0].(*EmployeesInMap).FullName.FirstName)
		assert.Equal(t, "MORENO              ", response.Data[0].(*EmployeesInMap).FullName.Name)
		assert.Equal(t, "VENT07", response.Data[0].(*EmployeesInMap).Department)
	}
	_ = response.DumpValues()
	assert.Len(t, response.Values, 0)
	request2, err2 := connection.CreateMapReadRequest(&VehicleInMap{}, 12)
	if !assert.NoError(t, err2) {
		return
	}
	err = request2.QueryFields("AA,AC,AE,AF,AG")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr = request2.SearchAndOrder("AC=50005600", "AC")
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	if assert.Len(t, response.Data, 1) {
		assert.Equal(t, "50005600", response.Data[0].(*VehicleInMap).ID)
		assert.Equal(t, "301                 ", response.Data[0].(*VehicleInMap).Model)
		assert.Equal(t, "GRISE     ", response.Data[0].(*VehicleInMap).Color)
		assert.Equal(t, uint64(1980), response.Data[0].(*VehicleInMap).Year)
	}

}
