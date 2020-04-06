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
	"runtime"
	"strconv"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"golang.org/x/net/html/charset"

	"github.com/stretchr/testify/assert"
)

func TestMapBasic(t *testing.T) {
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	e, n := charset.Lookup("ISO8859-1")
	d, nus := charset.Lookup("US-ASCII")
	fmt.Println("Lookup:", n, nus)
	germanTests := []string{"���", "���", "���", "���"}
	g2 := []int8{-28, -10, -4}
	{
		gb := make([]byte, 3)
		for i, b := range g2 {
			gb[i] = byte(b)
		}
		dst := make([]byte, 20)
		enc := d.NewDecoder()
		nd, ns, err := enc.Transform(dst, gb, false)
		fmt.Println(nus, "->", g2, gb)
		fmt.Println("G1 error ->", err)
		fmt.Println("G1->", nd, ns, dst, string(dst))

	}
	for _, g := range g2 {
		fmt.Printf("G2 %d\n", g)
	}
	{
		gb := make([]byte, 3)
		for i, b := range g2 {
			gb[i] = byte(b)
		}
		dst := make([]byte, 10)
		enc := e.NewDecoder()
		nd, ns, err := enc.Transform(dst, gb, false)
		fmt.Println("G2->", g2)
		fmt.Println("G2->", gb)
		fmt.Println("G2 error ->", err)
		fmt.Println("G2->", nd, ns, dst, string(dst))

	}
	for _, g := range []byte(germanTests[0]) {
		fmt.Printf("gx2 %d\n", g)
	}

	for _, g := range germanTests {
		fmt.Println("Origin:", []byte(g), g)
		fmt.Println("Lookup", n)
		enc := e.NewEncoder()
		dst := make([]byte, 10)
		nd, ns, err := enc.Transform(dst, []byte(g), false)
		fmt.Println(nd, ns, err, dst, string(dst))
	}

	m := NewAdabasMap("testmap")
	m.setDefaultOptions("charset=ISO8859-1")
	assert.Equal(t, "testmap", m.Name)
	assert.Equal(t, "ISO8859-1", m.DefaultCharset)
	m.setDefaultOptions("charset=US-ASCII")
	assert.Equal(t, "testmap", m.Name)
	assert.Equal(t, "US-ASCII", m.DefaultCharset)
	e, cname := charset.Lookup(m.DefaultCharset)
	assert.Equal(t, "", e)
	assert.Equal(t, "", cname)

}

func TestMapFieldFieldName(t *testing.T) {
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	tests := []struct {
		name string
		cc   mapField
		want string
	}{
		{"MapFieldDataFnr", mapFieldDataFnr, "RF"},
		{"MapFieldHost", mapFieldHost, "AB"},
		{"MapFieldModifyTime", mapFieldModifyTime, "ZB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cc.fieldName(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapField.fieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapFields(t *testing.T) {
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 4)
	adatypes.Central.Log.Debugf("Repository %#v\n", mr)
	m, err := mr.readAdabasMap(adabas, "EMPLOYEES-NAT-DDM")
	if !assert.NoError(t, err) {
		fmt.Println("Error found", err)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(m)

	groupLayout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeCharacter, "AE"),
		adatypes.NewType(adatypes.FieldTypeString, "AD"),
		adatypes.NewType(adatypes.FieldTypePacked, "AC"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt8, "AA"),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "AB", adatypes.OccNone, groupLayout),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)

	m.adaptFieldType(testDefinition, nil)
	testDefinition.DumpTypes(false, true)
}

func TestMaps(t *testing.T) {
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 250)
	err := mr.LoadMapRepository(adabas)
	assert.NoError(t, err)
	if err != nil {
		fmt.Println(err)
	} else {
		nr := 1
		for name, f := range mr.mapNames {
			if assert.NotZero(t, f.isn) {
				fmt.Printf("%s: ISN: %d\n", name, f.isn)
			} else {
				fmt.Printf("%s: Empty\n", name)
			}
			nr++
		}
	}
}

func TestMapCreate(t *testing.T) {
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	defer adabas.Close()

	repository := NewMapRepository(adabas, 250)
	dataRepository := NewMapRepository(adabas, 11)
	newMap := NewAdabasMap("GOTESTMAP", &repository.DatabaseURL)
	newMap.Data = &dataRepository.DatabaseURL
	newMap.addFields("AA", "PERSONNEL-ID")
	newMap.addFields("AB", "FULL-NAME")
	newMap.addFields("AC", "FIRST-NAME")
	newMap.addFields("AD", "MIDDLE-NAME")
	newMap.addFields("AE", "NAME")

	err := newMap.Store()
	assert.NoError(t, err)
}

func TestMapFieldsMainframe(t *testing.T) {
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "map.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	network := os.Getenv("ADAMFDBID")
	if network == "" {
		fmt.Println("Mainframe database not defined")
		return
	}
	dbid, derr := strconv.Atoi(network)
	if !assert.NoError(t, derr) {
		return
	}

	adabas, _ := NewAdabas(Dbid(dbid))
	defer adabas.Close()

	mr := NewMapRepository(adabas, 4)
	adatypes.Central.Log.Debugf("Repository %#v\n", mr)
	m, err := mr.readAdabasMap(adabas, "EMPLOYEES-NAT-MF")
	if !assert.NoError(t, err) {
		fmt.Println("Error found", err)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	assert.Equal(t, "EMPLOYEES-NAT-MF", m.Name)
	fmt.Println(m)
}
