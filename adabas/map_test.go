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
	"reflect"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMapFieldFieldName(t *testing.T) {
	f := initTestLogWithFile(t, "map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
	f := initTestLogWithFile(t, "map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(24)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 4)
	log.Debugf("Repository %#v\n", *mr)
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

	m.adaptFieldType(testDefinition)
	testDefinition.DumpTypes(false, true)
}

func TestMaps(t *testing.T) {
	f := initTestLogWithFile(t, "map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	defer adabas.Close()

	mr := NewMapRepository(adabas, 250)
	err := mr.LoadMapRepository(adabas)
	assert.NoError(t, err)
	if err != nil {
		fmt.Println(err)
	} else {
		nr := 1
		for name, isn := range mr.MapNames {
			if assert.NotZero(t, isn) {
				fmt.Printf("%s: ISN: %d\n", name, isn)
			} else {
				fmt.Printf("%s: Empty\n", name)
			}
			nr++
		}
	}
}

func TestMapCreate(t *testing.T) {
	f := initTestLogWithFile(t, "map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
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
