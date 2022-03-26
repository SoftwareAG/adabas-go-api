/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"reflect"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type Employees struct {
	Index     uint64 `adabas:"#ISN" json:"-"`
	ID        string
	Birth     int64
	Name      string `adabas:"Name"`
	FirstName string `adabas:"FirstName"`
}

func initEmployees(t *testing.T) error {
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	mr := NewMapRepository(ada, 249)

	_, mErr := mr.SearchMap(ada, "Employees")
	if mErr == nil {
		fmt.Println("Employees map already available")
		return mErr
	}
	fmt.Println("Error reading map:", mErr)

	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + "Employees.json"
	fmt.Println("Loading ...." + name)
	file, err := os.Open(name)
	if !assert.NoError(t, err) {
		return err
	}
	defer file.Close()

	mapRepository := &mr.DatabaseURL

	maps, err := ParseJSONFileForFields(file)
	if !assert.NoError(t, err) {
		return err
	}
	assert.Equal(t, 1, len(maps))
	for _, m := range maps {
		m.Repository = mapRepository
		err = m.Store()
		if err != nil {
			return err
		}
	}

	return nil
}

func TestStructStore(t *testing.T) {
	lErr := initLogWithFile("structure.log")
	if !assert.NoError(t, lErr) {
		return
	}

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	initEmployees(t)
	connection, err := NewConnection("acj;map;config=[23,249]")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, connection) {
		return
	}
	defer connection.Close()

	e := []*Employees{
		{ID: "GOSTORE", Birth: 123478, Name: "ABC"},
	}
	err = connection.ReflectStore(e, "Employees")
	if assert.NoError(t, err) {
		fmt.Println("ISN:", e[0].Index)
		assert.NotEqual(t, 0, e[0].Index)
		err = connection.EndTransaction()
		assert.NoError(t, err)
	}
}

func TestStructSimple(t *testing.T) {
	lErr := initLogWithFile("structure.log")
	if !assert.NoError(t, lErr) {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	ierr := initEmployees(t)
	if !assert.NoError(t, ierr) {
		return
	}
	connection, err := NewConnection("acj;map;config=[23,249]")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, connection) {
		return
	}
	defer connection.Close()

	e := Employees{Name: "ABC"}
	ts := reflect.TypeOf(e)
	fmt.Println(e, ts)
	employeesType := reflect.TypeOf((*Employees)(nil)).Elem()
	fmt.Println(reflect.TypeOf((*Employees)(nil)).Elem())
	list, err := connection.ReflectSearch("Employees", employeesType, "ID=GOSTORE")
	if !assert.NoError(t, err) {
		return
	}
	for c, l := range list {
		e := l.(*Employees)
		fmt.Printf("%d.record:%#v %s -> Index=%d\n", c, l, e.Name, e.Index)
		assert.NotEqual(t, uint64(0), e.Index)
	}
}
