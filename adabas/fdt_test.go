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
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

var employeeFdt = []byte{68, 2, 0, 0, 0, 0, 34, 0, 104, 91, 138, 78, 61, 254, 4, 0, 70, 16, 65, 65, 65, 129, 0, 1, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 66, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 67, 65, 16, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 69, 65, 134, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0, 70, 16, 65, 68, 65, 16, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 70, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 70, 16, 65, 71, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0,
	70, 16, 65, 72, 80, 128, 1, 1, 0, 0, 0, 0, 4, 0, 0, 0, 70, 16, 65, 49, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 73, 65, 48, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0, 70, 16, 65, 74, 65, 144, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 75, 65, 16, 0, 2, 0, 0, 0, 0, 10, 0, 0, 0, 70, 16, 65, 76, 65, 16, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0,
	70, 16, 65, 50, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 78, 65, 16, 0, 2, 0, 0, 0, 0, 6, 0, 0, 0,
	70, 16, 65, 77, 65, 16, 0, 2, 0, 0, 0, 0, 15, 0, 0, 0, 70, 16, 65, 79, 65, 130, 0, 1, 0, 0, 0, 0, 6, 0, 0, 0,
	70, 16, 65, 80, 65, 144, 0, 1, 0, 0, 0, 0, 25, 0, 0, 0, 70, 16, 65, 81, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 82, 65, 26, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 65, 83, 80, 26, 0, 2, 0, 0, 0, 0, 5, 0, 0, 0,
	70, 16, 65, 84, 80, 56, 0, 2, 0, 0, 0, 0, 5, 0, 0, 0, 70, 16, 65, 51, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 85, 85, 2, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0, 70, 16, 65, 86, 85, 18, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0,
	70, 16, 65, 87, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 88, 85, 24, 0, 2, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 89, 85, 24, 0, 2, 0, 0, 0, 0, 8, 0, 0, 0, 70, 16, 65, 90, 65, 176, 0, 1, 0, 0, 0, 0, 3, 0, 0, 0,
	80, 12, 80, 72, 65, 0, 20, 0, 0, 0, 65, 69, 84, 24, 72, 49, 66, 144, 4, 0, 0, 2, 65, 85, 1, 0, 2, 0, 65,
	86, 1, 0, 2, 0, 0, 0, 83, 16, 83, 49, 65, 128, 4, 0, 0, 1, 65, 79, 1, 0, 4, 0,
	84, 24, 83, 50, 65, 128, 26, 0, 0, 2, 65, 79, 1, 0, 6, 0, 65, 69, 1, 0, 20, 0, 0, 0,
	84, 24, 83, 51, 65, 152, 12, 0, 0, 2, 65, 82, 1, 0, 3, 0, 65, 83, 1, 0, 9, 0, 0, 0}
var newEmployeeFdt = []byte{72, 4, 0, 0, 0, 0, 64, 0, 72, 246, 120, 78, 61, 254, 4, 0, 70, 16, 65, 48, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 65, 65, 129, 3, 2, 0, 0, 0, 0, 8, 0, 0, 0, 70, 16, 65, 66, 32, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 67, 70, 128, 0, 3, 0, 0, 0, 0, 4, 0, 0, 0, 70, 16, 65, 68, 66, 16, 32, 3, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 69, 65, 16, 200, 3, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 66, 48, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 66, 65, 87, 16, 0, 2, 0, 0, 0, 0, 40, 0, 0, 0, 70, 16, 66, 66, 87, 16, 0, 2, 0, 0, 0, 0, 40, 0, 0, 0,
	70, 16, 66, 67, 87, 146, 0, 2, 0, 0, 0, 0, 50, 0, 0, 0, 70, 16, 67, 65, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0,
	70, 16, 68, 65, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 70, 16, 69, 65, 80, 128, 1, 1, 0, 0, 0, 0, 4, 0, 0, 0,
	70, 16, 70, 48, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 70, 65, 87, 56, 0, 2, 0, 0, 0, 0, 60, 0, 0, 0,
	70, 16, 70, 66, 87, 152, 0, 2, 0, 0, 0, 0, 40, 0, 0, 0, 70, 16, 70, 67, 65, 24, 0, 2, 0, 0, 0, 0, 10, 0, 0, 0,
	70, 16, 70, 68, 65, 24, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 70, 49, 32, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 70, 69, 65, 24, 0, 3, 0, 0, 0, 0, 6, 0, 0, 0, 70, 16, 70, 70, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0,
	70, 16, 70, 71, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0, 70, 16, 70, 72, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0,
	70, 16, 70, 73, 65, 184, 0, 3, 0, 0, 0, 0, 80, 0, 0, 0, 70, 16, 73, 48, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 73, 65, 87, 56, 0, 2, 0, 0, 0, 0, 40, 0, 0, 0, 70, 16, 73, 66, 87, 152, 0, 2, 0, 0, 0, 0, 40, 0, 0, 0,
	70, 16, 73, 67, 65, 24, 0, 2, 0, 0, 0, 0, 10, 0, 0, 0, 70, 16, 73, 68, 65, 24, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0,
	70, 16, 73, 69, 65, 24, 0, 2, 0, 0, 0, 0, 5, 0, 0, 0, 70, 16, 73, 49, 32, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 73, 70, 65, 24, 0, 3, 0, 0, 0, 0, 6, 0, 0, 0, 70, 16, 73, 71, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0,
	70, 16, 73, 72, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0, 70, 16, 73, 73, 65, 24, 0, 3, 0, 0, 0, 0, 15, 0, 0, 0,
	70, 16, 73, 74, 65, 184, 0, 3, 0, 0, 0, 0, 80, 0, 0, 0, 70, 16, 74, 65, 65, 130, 0, 1, 0, 0, 0, 0, 6, 0, 0, 0,
	70, 16, 75, 65, 87, 144, 0, 1, 0, 0, 0, 0, 66, 0, 0, 0, 70, 16, 76, 48, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 76, 65, 65, 26, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 76, 66, 80, 26, 0, 2, 0, 0, 0, 0, 6, 0, 0, 0,
	70, 16, 76, 67, 80, 184, 0, 2, 0, 0, 0, 0, 6, 0, 0, 0, 70, 16, 77, 65, 71, 16, 0, 1, 0, 0, 0, 0, 4, 0, 0, 0,
	70, 16, 78, 48, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 78, 65, 85, 2, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0,
	70, 16, 78, 66, 85, 18, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 79, 48, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 79, 65, 85, 24, 0, 2, 1, 0, 0, 0, 8, 0, 0, 0, 70, 16, 79, 66, 85, 24, 0, 2, 1, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 80, 65, 65, 176, 0, 1, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 81, 65, 80, 0, 0, 1, 0, 0, 0, 0, 7, 0, 0, 0,
	70, 16, 82, 65, 65, 16, 196, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 83, 48, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 83, 65, 87, 24, 0, 2, 0, 0, 0, 0, 80, 0, 0, 0, 70, 16, 83, 66, 65, 24, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0,
	70, 16, 83, 67, 65, 56, 196, 2, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 84, 67, 85, 0, 0, 1, 4, 64, 1, 0, 20, 0, 0, 0,
	70, 16, 84, 85, 85, 32, 0, 1, 4, 0, 1, 0, 20, 0, 0, 0, 67, 48, 67, 78, 87, 144, 120, 4, 66, 67, 120, 4, 0, 32, 39, 100, 101,
	64, 99, 111, 108, 108, 97, 116, 105, 111, 110, 61, 112, 104, 111, 110, 101, 98, 111, 111, 107, 39, 44,
	80, 82, 73, 77, 65, 82, 89, 0, 0,
	84, 24, 72, 49, 66, 144, 5, 0, 0, 2, 78, 65, 1, 0, 2, 0, 78, 66, 1, 0, 3, 0, 0, 0, 83, 16, 83, 49, 65, 128, 2, 0, 0, 1, 74, 65, 1, 0, 2, 0,
	84, 24, 83, 50, 65, 144, 46, 0, 0, 2, 74, 65, 1, 0, 6, 0, 66, 67, 1, 0, 40, 0, 0, 0,
	84, 24, 83, 51, 65, 152, 9, 0, 0, 2, 76, 65, 1, 0, 3, 0, 76, 66, 1, 0, 6, 0, 0, 0, 82, 16, 72, 79, 12, 0, 0, 0, 65, 65, 65, 67, 1, 0, 0, 0}
var hyperExitEmployeeFdt = []byte{88, 2, 0, 0, 0, 0, 35, 0, 139, 14, 54, 235, 169, 73, 5, 0, 70, 16, 65, 65, 65, 129, 0, 1, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 66, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 67, 65, 16, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 69, 65, 134, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0, 70, 16, 65, 68, 65, 16, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 70, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 70, 16, 65, 71, 65, 64, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0,
	70, 16, 65, 72, 80, 128, 1, 1, 0, 0, 0, 0, 4, 0, 0, 0, 70, 16, 65, 49, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 73, 65, 48, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0,
	70, 16, 65, 74, 65, 144, 0, 2, 0, 0, 0, 0, 20, 0, 0, 0, 70, 16, 65, 75, 65, 16, 0, 2, 0, 0, 0, 0, 10, 0, 0, 0,
	70, 16, 65, 76, 65, 16, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0,
	70, 16, 65, 50, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 78, 65, 16, 0, 2, 0, 0, 0, 0, 6, 0, 0, 0,
	70, 16, 65, 77, 65, 16, 0, 2, 0, 0, 0, 0, 15, 0, 0, 0,
	70, 16, 65, 79, 65, 130, 0, 1, 0, 0, 0, 0, 6, 0, 0, 0, 70, 16, 65, 80, 65, 144, 0, 1, 0, 0, 0, 0, 25, 0, 0, 0,
	70, 16, 65, 81, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	70, 16, 65, 82, 65, 26, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 70, 16, 65, 83, 80, 26, 0, 2, 0, 0, 0, 0, 5, 0, 0, 0,
	70, 16, 65, 84, 80, 56, 0, 2, 0, 0, 0, 0, 5, 0, 0, 0,
	70, 16, 65, 51, 32, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 85, 85, 2, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0,
	70, 16, 65, 86, 85, 18, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0,
	70, 16, 65, 87, 32, 8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 70, 16, 65, 88, 85, 24, 0, 2, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 89, 85, 24, 0, 2, 0, 0, 0, 0, 8, 0, 0, 0,
	70, 16, 65, 90, 65, 176, 0, 1, 0, 0, 0, 0, 3, 0, 0, 0, 80, 12, 80, 72, 65, 0, 20, 0, 0, 0, 65, 69, 84, 24,
	72, 49, 66, 144, 4, 0, 0, 2, 65, 85, 1, 0, 2, 0, 65,
	86, 1, 0, 2, 0, 0, 0, 83, 16, 83, 49, 65, 128, 2, 0, 0, 1, 65, 79, 1, 0, 2, 0, 84, 24, 83, 50, 65, 128, 26, 0, 0, 2, 65,
	79, 1, 0, 6, 0, 65, 69, 1, 0, 20, 0, 0, 0,
	84, 24, 83, 51, 65, 152, 8, 0, 0, 2, 65, 82, 1, 0, 3, 0, 65, 83, 1, 0, 5, 0, 0, 0, 72, 20, 72, 89, 65, 48, 20, 0, 1, 0, 0, 4,
	65, 65, 65, 67, 65, 73, 65, 70}

func TestFdtDefinition(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")
	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	assert.Equal(t, byte('F'), fieldIdentifierField.code())
	assert.Equal(t, byte('S'), fieldIdentifierSub.code())
	assert.Equal(t, byte('T'), fieldIdentifierSuper.code())
	assert.Equal(t, byte('P'), fieldIdentifierPhonetic.code())
	assert.Equal(t, []byte{1, 2, 3, 4, 12, 17, 18}, fdtCondition[fieldIdentifierPhonetic.code()])
	assert.Equal(t, []byte{1, 2, 3, 4, 12, 18, 23, 24, 25}, fdtCondition[fieldIdentifierCollation.code()])
}

func traverseOutput(IAdaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	y := strings.Repeat(" ", level)
	fmt.Println(y, level, ". ", IAdaType.String())
	return nil
}

func TestFdtParse(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	fmt.Println("Parse FDT structure")
	helper := adatypes.NewHelper(employeeFdt, len(employeeFdt), binary.LittleEndian)
	option := adatypes.NewBufferOption(false, 0)
	fdtDefinition := createFdtDefintion()
	_, err := fdtDefinition.ParseBuffer(helper, option, "")
	assert.NoError(t, err)
	fdt := fdtDefinition.Search("fdt")
	fmt.Println("FDT ", fdt.PeriodIndex())
	tm := adatypes.NewTraverserMethods(traverseOutput)
	err = fdtDefinition.TraverseTypes(tm, true, nil)
	assert.NoError(t, err)
}

func TestFdtStructure(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	fmt.Println("Test FDT structure")
	helper := adatypes.NewHelper(employeeFdt, len(employeeFdt), binary.LittleEndian)
	option := adatypes.NewBufferOption(false, 0)
	fdtDefinition := createFdtDefintion()
	res, pErr := fdtDefinition.ParseBuffer(helper, option, "")
	assert.NoError(t, pErr)
	assert.Equal(t, adatypes.Continue, res)
	fdtTable, err := createFieldDefinitionTable(fdtDefinition)
	assert.Nil(t, err, "Error creating fdt table")
	fmt.Println("FDT TABLE: ", fdtTable)
}

func TestFdtStructureNewEmployee(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	fmt.Println("Test FDT structure")
	helper := adatypes.NewHelper(newEmployeeFdt, len(newEmployeeFdt), binary.LittleEndian)
	option := adatypes.NewBufferOption(false, 0)
	fdtDefinition := createFdtDefintion()
	_, err := fdtDefinition.ParseBuffer(helper, option, "")
	assert.NoError(t, err, "Error parsing fdt table")
	fdtTable, err := createFieldDefinitionTable(fdtDefinition)
	assert.NoError(t, err, "Error creating fdt table")
	fmt.Println("FDT TABLE: ", fdtTable)
}

func TestFdtStructureHyperExitEmployee(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	fmt.Println("Test FDT structure")
	helper := adatypes.NewHelper(hyperExitEmployeeFdt, len(hyperExitEmployeeFdt), binary.LittleEndian)
	fdtDefinition := createFdtDefintion()
	_, err := fdtDefinition.ParseBuffer(helper, adatypes.NewBufferOption(false, 0), "")
	assert.NoError(t, err)
	fdtTable, err := createFieldDefinitionTable(fdtDefinition)
	assert.Nil(t, err, "Error creating fdt table")
	assert.NoError(t, err)
	fmt.Println("FDT TABLE: ", fdtTable)
}

func TestReadFDT(t *testing.T) {
	initTestLogWithFile(t, "fdt.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasStatDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	def, err := connection.adabasToData.ReadFileDefinition(11)
	if !assert.NoError(t, err) {
		return
	}
	if connection.adabasToData.Acbx.Acbxrsp != 0 {
		t.Fatal(connection.adabasToData.getAdabasMessage(), connection.adabasToData.Acbx.Acbxrsp)
	}

	validateFile(t, "fdtTest11", []byte(def.String()), txtType)

	ph, pherr := def.SearchType("PH")
	assert.Equal(t, "PH", ph.Name())
	assert.NoError(t, pherr)
	assert.Equal(t, uint32(0), ph.Length())
	phType := ph.(*adatypes.AdaPhoneticType)
	assert.Equal(t, "AE", string(phType.ParentName[:]))
	assert.Equal(t, uint16(0x14), phType.DescriptorLength)

	def, err = connection.adabasToData.ReadFileDefinition(9)
	if !assert.NoError(t, err) {
		return
	}
	if connection.adabasToData.Acbx.Acbxrsp != 0 {
		t.Fatal(connection.adabasToData.getAdabasMessage(), connection.adabasToData.Acbx.Acbxrsp)
	}

	validateFile(t, "fdtTest9", []byte(def.String()), txtType)

	col, serr := def.SearchType("CN")
	assert.Equal(t, "CN", col.Name())
	assert.NoError(t, serr)
	assert.Equal(t, uint32(1144), col.Length())
	colType := col.(*adatypes.AdaCollationType)
	assert.Equal(t, "'de@collation=phonebook',PRIMARY", colType.CollAttribute)
	assert.Equal(t, "BC", string(colType.ParentName[:]))
	assert.Equal(t, " ", col.Type().FormatCharacter())
	assert.Equal(t, "NU HE", col.Option())

	hy, herr := def.SearchType("H1")
	assert.NoError(t, herr)
	assert.Equal(t, uint32(5), hy.Length())
	hyp := hy.(*adatypes.AdaSuperType)
	assert.Equal(t, "H1", hy.Name())
	assert.Equal(t, "B", string(hyp.FdtFormat))
	assert.Equal(t, "NU", hy.Option())

	s1, s1err := def.SearchType("S1")
	assert.NoError(t, s1err)
	assert.Equal(t, "S1", s1.Name())
	assert.Equal(t, uint32(2), s1.Length())
	s1t := s1.(*adatypes.AdaSuperType)
	assert.Equal(t, "A", string(s1t.FdtFormat))
	assert.Equal(t, "", s1.Option())

	s2, s2err := def.SearchType("S2")
	assert.NoError(t, s2err)
	assert.Equal(t, "S2", s2.Name())
	assert.Equal(t, uint32(46), s2.Length())
	s2t := s2.(*adatypes.AdaSuperType)
	assert.Equal(t, "A", string(s2t.FdtFormat))
	assert.Equal(t, "NU", s2.Option())

	s3, s3err := def.SearchType("S3")
	assert.NoError(t, s3err)
	assert.Equal(t, "S3", s3.Name())
	assert.Equal(t, uint32(9), s3.Length())
	s3t := s3.(*adatypes.AdaSuperType)
	assert.Equal(t, "A", string(s3t.FdtFormat))
	assert.Equal(t, "NU,PE", s3.Option())

	ho, hoerr := def.SearchType("HO")
	assert.NoError(t, hoerr)
	assert.Equal(t, "HO=REFINT(AC,12,AA/DX,UX) ; HO", ho.String())
	assert.Equal(t, "HO", ho.Name())
	assert.Equal(t, uint32(0), ho.Length())
	hot := ho.(*adatypes.AdaReferentialType)
	assert.Equal(t, "PRIMARY", hot.ReferentialType())
	assert.Equal(t, "DELETE_NOACTION", hot.DeleteAction())
	assert.Equal(t, "UPDATE_NOACTION", hot.UpdateAction())
	assert.Equal(t, "AA", hot.PrimaryKeyName())
	assert.Equal(t, "AC", hot.ForeignKeyName())
	assert.Equal(t, uint32(12), hot.ReferentialFile())

}
