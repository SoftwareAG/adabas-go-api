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
	"testing"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestMapRepository(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(24)
	defer ada.Close()
	AddGlobalMapRepositoryReference("24,4")
	defer DelGlobalMapRepository(ada, 4)
	id := NewAdabasID()
	defer id.Close()
	adabasMap, rep, err := SearchMapRepository(id, "EMPLOYEES-NAT-DDM")
	assert.NoError(t, err)
	assert.NotNil(t, adabasMap)
	assert.NotNil(t, rep)
	assert.Equal(t, Dbid(24), rep.DatabaseURL.URL.Dbid)
	//assert.Equal(t, "", rep.DatabaseURL.URL.Host)

}

func TestGlobalMapRepository(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(23)
	defer ada.Close()
	AddGlobalMapRepository(ada.URL, 4)
	defer DelGlobalMapRepository(ada.URL, 4)
	ada.SetDbid(24)
	AddGlobalMapRepository(ada.URL, 4)
	defer DelGlobalMapRepository(ada.URL, 4)

	ada2, _ := NewAdabas(1)
	defer ada2.Close()
	adabasMaps, err := AllGlobalMaps(ada2)
	assert.NoError(t, err)
	assert.NotNil(t, adabasMaps)
	for _, m := range adabasMaps {
		fmt.Printf("%s -> %d\n", m.Name, m.Isn)
	}
	listMaps, lerr := AllGlobalMapNames(ada2)
	assert.NoError(t, lerr)
	assert.NotNil(t, listMaps)
	for _, m := range listMaps {
		fmt.Printf("%s\n", m)
	}

}

func TestGlobalMapConnectionString(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(24)
	defer ada.Close()
	AddGlobalMapRepository(ada.URL, 4)
	defer DelGlobalMapRepository(ada.URL, 4)

	connection, cerr := NewConnection("acj;map=EMPLOYEES")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateReadRequest()
	if !assert.NoError(t, rerr) {
		return
	}
	request.QueryFields("name,personnel-id")
	result, err := request.ReadLogicalWith("personnel-id=[11100301:11100303]")
	if !assert.NoError(t, err) {
		return
	}
	_ = result.DumpValues()
}

func TestGlobalMapConnectionDirect(t *testing.T) {
	initTestLogWithFile(t, "map_repositories.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(24)
	defer ada.Close()
	AddGlobalMapRepository(ada.URL, 4)
	defer DelGlobalMapRepository(ada, 4)

	connection, cerr := NewConnection("acj;map")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest("EMPLOYEES")
	if !assert.NoError(t, rerr) {
		return
	}
	request.QueryFields("name,personnel-id")
	result, err := request.ReadLogicalWith("personnel-id=[11100301:11100303]")
	if !assert.NoError(t, err) {
		return
	}
	_ = result.DumpValues()
}

func TestThreadMapCache(t *testing.T) {
	initTestLogWithFile(t, "global_map_repositories.log")

	DumpGlobalMapRepositories()
	CleanGlobalMapRepository()

	StartAsynchronousMapCache(10)
	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ada, _ := NewAdabas(23)
	defer ada.Close()
	m, _, err := SearchMapRepository(ada.ID, "VEHICLESGo")
	assert.Nil(t, m)
	if !assert.Error(t, err) {
		fmt.Println("Map got:", m)
		return
	}
	fmt.Println("Search failed: ", err)
	AddGlobalMapRepository(ada.URL, 250)
	defer DelGlobalMapRepository(ada, 250)
	time.Sleep(60 * time.Second)
	m, _, err = SearchMapRepository(ada.ID, "VEHICLESGo")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Map names found: ", m.Name)
	assert.Equal(t, "VEHICLESGo", m.Name)

}
