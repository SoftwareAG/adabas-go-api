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
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type Employees struct {
	ID        string
	Birth     int64
	Name      string `adabas:"Name"`
	FirstName string `adabas:"FirstName"`
}

func initEmployees(t *testing.T) error {
	ada := NewAdabas(adabasModDBID)
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
	f, lErr := initLogWithFile("structure.log")
	if !assert.NoError(t, lErr) {
		return
	}
	defer f.Close()

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}

	log.Infof("TEST: %s", t.Name())
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
		&Employees{ID: "GOSTORE", Birth: 123478, Name: "ABC"},
	}
	err = ReflectStore(e, connection, "Employees")
	if assert.NoError(t, err) {
		connection.EndTransaction()
	}
}

func TestStructSimple(t *testing.T) {
	f, lErr := initLogWithFile("structure.log")
	if !assert.NoError(t, lErr) {
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	//	st := reflect.StructOf(e)
	fmt.Println(e, ts)
	employeesType := reflect.TypeOf((*Employees)(nil)).Elem()
	fmt.Println(reflect.TypeOf((*Employees)(nil)).Elem())
	list, err := ReflectSearch("Employees", employeesType, connection, "ID=GOSTORE")
	if !assert.NoError(t, err) {
		return
	}
	for c, l := range list {
		e := l.(*Employees)
		fmt.Printf("%d:%#v %s\n", c, l, e.Name)
	}
}
