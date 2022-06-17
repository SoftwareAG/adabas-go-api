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

package adatypes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefinitionGroup(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "G1"),
		NewType(FieldTypeString, "GX"),
		NewType(FieldTypePacked, "PA"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,G1,1,A,GX,1,A,PA,1,P,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionPeriodic(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "PG", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	// testDefinition.DumpTypes(false)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	//assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,PGC,4,B,GC1-N,1,A,GS1-N,1,A,GP1-N,1,P,I8,8,B.",
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,PGC,4,B,PG1-N,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionMultiple(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	multipleLayout[0].AddFlag(FlagOptionAtomicFB)
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "P1"),
		NewStructureList(FieldTypeMultiplefield, "PM", OccNone, multipleLayout),
		NewType(FieldTypeString, "PA"),
		NewType(FieldTypePacked, "PX"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "PG", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,P1,1,A,PMC,4,B,PM1-N,1,P,PA,1,A,PX,1,P,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionQuerySimple(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "P1"),
		NewStructureList(FieldTypeMultiplefield, "PM", OccNone, multipleLayout),
		NewType(FieldTypeString, "PA"),
		NewType(FieldTypePacked, "PX"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "PG", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	err = testDefinition.ShouldRestrictToFields("U4,I2")
	assert.Equal(t, nil, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,I2,2,B.",
		request.FormatBuffer.String())

}

func TestDefinitionQueryGroupField(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "PM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	err = testDefinition.ShouldRestrictToFields("U4,GS")
	assert.Equal(t, nil, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,GS,1,A.",
		request.FormatBuffer.String())
}

func TestDefinitionQueryGroupFieldTwice(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
	}
	groupLayout3 := []IAdaType{
		NewType(FieldTypeCharacter, "YC"),
		NewType(FieldTypeString, "YS"),
		NewType(FieldTypePacked, "YP"),
	}
	for _, l := range groupLayout3 {
		l.SetLevel(3)
	}
	groupLayout2 := []IAdaType{
		NewType(FieldTypeCharacter, "XC"),
		NewStructureList(FieldTypeGroup, "YY", OccNone, groupLayout3),
		NewType(FieldTypeString, "XS"),
		NewType(FieldTypePacked, "XP"),
	}
	for _, l := range groupLayout2 {
		l.SetLevel(2)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewStructureList(FieldTypeGroup, "XX", OccNone, groupLayout2),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	err = testDefinition.ShouldRestrictToFields("U4,GS,YS")
	assert.Equal(t, nil, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)
	assert.Equal(t, "U4,4,B,YS,1,A,GS,1,A.", request.FormatBuffer.String())
	adaType, err := testDefinition.SearchType("B1")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, "B1", adaType.Name())
	adaType, err = testDefinition.SearchType("XP")
	assert.NoError(t, err)
	assert.Equal(t, "XP", adaType.Name())
	assert.Equal(t, FieldTypePacked, adaType.Type())
}

func TestDefinitionQueryWithLongname(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewLongNameType(FieldTypePacked, "Packed", "PA"),
	}
	groupLayout := []IAdaType{
		NewLongNameType(FieldTypeCharacter, "GroupCharacter", "GC"),
		NewLongNameStructureList(FieldTypeMultiplefield, "Packed", "PA", OccNone, multipleLayout),
		NewLongNameType(FieldTypeString, "GroupString", "GS"),
		NewLongNameType(FieldTypePacked, "GroupPacked", "GP"),
	}
	layout := []IAdaType{
		NewLongNameType(FieldTypeUInt4, "UInt4", "U4"),
		NewLongNameType(FieldTypeByte, "Byte", "B1"),
		NewLongNameType(FieldTypeUByte, "UnsignedByte", "UB"),
		NewLongNameType(FieldTypeUInt2, "Int2", "I2"),
		NewLongNameType(FieldTypeUInt8, "UInt8", "U8"),
		NewLongNameStructureList(FieldTypeGroup, "Group", "GR", OccNone, groupLayout),
		NewLongNameType(FieldTypeUInt8, "Int8", "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	err = testDefinition.ShouldRestrictToFields("UInt4,Int2")
	assert.Equal(t, nil, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,I2,2,B.",
		request.FormatBuffer.String())

}

func TestDefinitionCreateValues(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
	}
	groupLayout3 := []IAdaType{
		NewType(FieldTypeCharacter, "YC"),
		NewType(FieldTypeString, "YS"),
		NewType(FieldTypePacked, "YP"),
	}
	for _, l := range groupLayout3 {
		l.SetLevel(3)
	}
	groupLayout2 := []IAdaType{
		NewType(FieldTypeCharacter, "XC"),
		NewStructureList(FieldTypeGroup, "YY", OccNone, groupLayout3),
		NewType(FieldTypeString, "XS"),
		NewType(FieldTypePacked, "XP"),
	}
	for _, l := range groupLayout2 {
		l.SetLevel(2)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewStructureList(FieldTypeGroup, "XX", OccNone, groupLayout2),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	assert.Nil(t, testDefinition.Values)
	testDefinition.CreateValues(false)
	assert.NotNil(t, testDefinition.Values)
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpValues(false)
}

func TestDefinitionQueryMultipleField(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "GM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)

	err = testDefinition.ShouldRestrictToFields("U4,GS,GM")
	assert.Equal(t, nil, err)
	// testDefinition.DumpTypes(false, false)
	// testDefinition.DumpTypes(false, true)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)
	Central.Log.Debugf(" ------------------------ after create adabas request 0 0")
	assert.Equal(t, "U4,4,B,GMC,4,B,GM1-N,1,P,GS,1,A.",
		request.FormatBuffer.String())

	adabasParameter := &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err = testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)

	Central.Log.Debugf(" ------------------------ after create adabas request 1 0")
	testDefinition.DumpValues(false)

	assert.Equal(t, "U4,4,B,GS,1,A.",
		request.FormatBuffer.String())

	v := testDefinition.Search("GM")
	sv := v.(*StructureValue)
	st := v.Type().(*StructureType)
	muV, err := st.SubTypes[0].Value()
	muV.setMultipleIndex(1)
	assert.NoError(t, err)
	sv.addValue(muV, 0)

	testDefinition.DumpValues(false)
	Central.Log.Debugf(" ------------------------ before create adabas request 0 0")

	parameter = &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 1, Mainframe: false}
	request, err = testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)
	Central.Log.Debugf(" ------------------------ after create adabas request 0 1")

	assert.Equal(t, ".",
		request.FormatBuffer.String())

}

func createPeriodGroupMultiplerField() *Definition {
	multipleLayout := []IAdaType{
		NewTypeWithLength(FieldTypePacked, "GM", 5),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(FlagOptionSecondCall)
		l.AddFlag(FlagOptionMUGhost)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func createPeriodGroupSuperDescriptor() *Definition {
	multipleLayout := []IAdaType{
		NewTypeWithLength(FieldTypePacked, "GM", 5),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(FlagOptionSecondCall)
		l.AddFlag(FlagOptionMUGhost)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
		NewSuperType("S1", 0),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func createPeriodGroupMultiplerLobField() *Definition {
	multipleLayout := []IAdaType{
		NewTypeWithLength(FieldTypeLBString, "GM", 0),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(FlagOptionSecondCall)
		l.AddFlag(FlagOptionMUGhost)
	}
	multipleLayout[0].AddOption(FieldOptionMU)
	multipleLayout[0].AddOption(FieldOptionNU)
	multipleLayout[0].AddOption(FieldOptionNB)
	multipleLayout[0].AddOption(FieldOptionNV)
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "GM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	// groupLayout[1].AddOption(FieldOptionMU)
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "GR", OccNone, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, l := range layout {
		l.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func ExampleDefinition_dumpValues() {
	err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpValues(false)
	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump values :   U4 = >0<
	//  B1 = >0<
	//  UB = >0<
	//  I2 = >0<
	//  U8 = >0<
	//  GR = [0]
	//  I8 = >0<

}

func ExampleDefinition_search() {
	err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)

	err = testDefinition.SetValueWithIndex("UB", nil, 1)
	if err != nil {
		fmt.Println("Add value to UB:", err)
		return
	}
	err = testDefinition.SetValueWithIndex("GC", []uint32{1}, "A")
	if err != nil {
		fmt.Println("Add Value of GC:", err)
		return
	}
	err = testDefinition.SetValueWithIndex("GM", []uint32{1, 1}, 123)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	err = testDefinition.SetValueWithIndex("GM", []uint32{2, 1}, 555)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	Central.Log.Debugf("Add GM [2,2]")
	err = testDefinition.SetValueWithIndex("GM", []uint32{2, 2}, 111)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	Central.Log.Debugf("Add GM [2,3]")
	err = testDefinition.SetValueWithIndex("GM", []uint32{2, 3}, 777)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	Central.Log.Debugf("Done GM [2,3]")
	Central.Log.Debugf("Add GM [2,5]")
	err = testDefinition.SetValueWithIndex("GM", []uint32{2, 5}, 8888)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	Central.Log.Debugf("Done GM [2,5]")
	err = testDefinition.SetValueWithIndex("GM", []uint32{2, 15}, 10000)
	if err != nil {
		fmt.Println("Add Value of GM:", err)
		return
	}
	testDefinition.DumpValues(false)
	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump values :   U4 = >0<
	//  B1 = >0<
	//  UB = >1<
	//  I2 = >0<
	//  U8 = >0<
	//  GR = [2]
	//   GC[1] = >65<
	//   GM[1] = [1]
	//    GM[1,1] = >123<
	//   GS[1] = > <
	//   GP[1] = >0<
	//   GC[2] = >0<
	//   GM[2] = [1]
	//    GM[2,1] = >555<
	//    GM[2,2] = >111<
	//    GM[2,3] = >777<
	//    GM[2,5] = >8888<
	//    GM[2,15] = >10000<
	//   GS[2] = > <
	//   GP[2] = >0<
	//  I8 = >0<
}

func ExampleDefinition_addValue() {
	err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)

	err = testDefinition.SetValueWithIndex("UB", nil, 1)
	if err != nil {
		fmt.Println("Add value to UB:", err)
		return
	}
	err = testDefinition.SetValueWithIndex("GC", []uint32{1}, "A")
	if err != nil {
		fmt.Println("Add Value of GC:", err)
		return
	}
	testDefinition.DumpValues(false)
	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//
	// Dump values :   U4 = >0<
	//  B1 = >0<
	//  UB = >1<
	//  I2 = >0<
	//  U8 = >0<
	//  GR = [1]
	//   GC[1] = >65<
	//   GM[1] = [0]
	//   GS[1] = > <
	//   GP[1] = >0<
	//  I8 = >0<
}

func TestDefinitionQueryPeriodGroupMultipleField(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)
	// Generate format buffer for first read call
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)
	testDefinition.DumpValues(false)
	Central.Log.Debugf(" ------------------------ after create adabas request 0 0")
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,GRC,4,B,GC1-N,1,A,GS1-N,1,A,GP1-N,1,P,I8,8,B.",
		request.FormatBuffer.String())

	// Generate format buffer for first store call
	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err = testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)

	Central.Log.Debugf(" ------------------------ after create adabas request 1 0")
	testDefinition.DumpValues(false)

	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,I8,8,B.",
		request.FormatBuffer.String(), "Wrong store format buffer")

	err = testDefinition.SetValueWithIndex("GM", []uint32{1, 1}, 1)
	assert.NoError(t, err)
	err = testDefinition.SetValueWithIndex("GM", nil, 1)
	assert.Error(t, err)
	// v := testDefinition.Search("GM")
	// sv := v.(*StructureValue)
	// st := v.Type().(*StructureType)
	// muV, err := st.SubTypes[0].Value()
	// muV.setMultipleIndex(1)
	// muV.setPeriodIndex(1)
	// assert.NoError(t, err)
	// sv.addValue(muV, 1)

	// Generate format buffer for first store call with PE/MU field data
	Central.Log.Debugf(" ------------------------ before create adabas request with data 1 0")
	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err = testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)
	Central.Log.Debugf(" ------------------------ after create adabas request with data 1 0")
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,GC1,1,A,GM1(1),5,P,GS1,1,A,GP1,1,P,I8,8,B.",
		request.FormatBuffer.String(), "Wrong store format buffer with PE/MU data")

	testDefinition.DumpValues(false)
	Central.Log.Debugf(" ------------------------ before create adabas request 0 0")

	// Generate format buffer for second read call with missing PE/MU field data
	adabasParameter = &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 1, Mainframe: false}
	request, err = testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)
	Central.Log.Debugf(" ------------------------ after create adabas request 0 1")

	assert.Equal(t, "GM1C,4,B,GM1(1-N),5.",
		request.FormatBuffer.String())

}

func TestDefinitionRestrictPeriodic(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "PG", OccCapacity, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	err = testDefinition.ShouldRestrictToFields("U4,PG")
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpTypes(false, true)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,PGC,4,B,PG1-N.",
		request.FormatBuffer.String())

}

func createLayout() *Definition {
	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "MA"),
	}
	multipleLayout[0].AddFlag(FlagOptionSecondCall)
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "PG", OccSingle, groupLayout),
		NewStructureList(FieldTypeGroup, "GM", OccSingle, multipleLayout),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func createLayoutWithPEandMU() *Definition {
	multiplePeriodLayout := []IAdaType{
		NewTypeWithLength(FieldTypeString, "PM", 10),
	}
	multiplePeriodLayout[0].AddFlag(FlagOptionMUGhost)
	multiplePeriodLayout[0].AddFlag(FlagOptionAtomicFB)
	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "GM"),
	}
	multipleLayout[0].AddFlag(FlagOptionMUGhost)
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
		NewStructureList(FieldTypeMultiplefield, "PM", OccCapacity, multiplePeriodLayout),
		NewType(FieldTypePacked, "GP"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "PG", OccCapacity, groupLayout),
		NewStructureList(FieldTypeMultiplefield, "GM", OccSingle, multipleLayout),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func TestDefinitionRestrictPeriodicWithMU(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	testDefinition.DumpValues(false)
	err = testDefinition.ShouldRestrictToFields("U4,PG")
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(adabasParameter)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,PGC,4,B,GC1-N,1,A,GS1-N,1,A,GP1-N,1,P.",
		request.FormatBuffer.String())
	testDefinition.DumpTypes(false, true)
}

func ExampleDefinition_dumpValuesAll() {
	err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Error init log ", err)
		return
	}

	testDefinition := createLayout()

	testDefinition.DumpTypes(false, true)
	testDefinition.DumpValues(false)

	// Output: Dump all active field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, PG  ; PG
	//     2, GC, 1, A  ; GC
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, GM  ; GM
	//     2, MA, 1, P  ; MA
	//
	// Dump values :   U4 = >0<
	//  B1 = >0<
	//  UB = >0<
	//  I2 = >0<
	//  U8 = >0<
	//  PG = [1]
	//   GC = >0<
	//   GS = > <
	//   GP = >0<
	//  GM = [1]
	//   MA = >0<

}

func ExampleDefinition_dumpValuesRestrict() {
	err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Error init log ", err)
		return
	}

	testDefinition := createLayout()

	err = testDefinition.ShouldRestrictToFields("U4,PG")
	if err != nil {
		fmt.Println("Error restrict fields ", err)
		return
	}
	testDefinition.DumpTypes(false, true)
	testDefinition.DumpValues(false)

	// Output: Dump all active field types:
	//   1, U4, 4, B  ; U4
	//   1, PG  ; PG
	//     2, GC, 1, A  ; GC
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//
	// Dump values :   U4 = >0<
	//  PG = [1]
	//   GC = >0<
	//   GS = > <
	//   GP = >0<

}

func TestDefinition_restrict(t *testing.T) {
	err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Error init log ", err)
		return
	}

	groupLayout2 := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
	}

	groupLayout := []IAdaType{
		NewType(FieldTypePacked, "GP"),
		NewStructureList(FieldTypeGroup, "G2", 1, groupLayout2),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "P1", OccCapacity, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()

	err = testDefinition.ShouldRestrictToFields("G2")
	if !assert.NoError(t, err) {
		fmt.Println("Error restrict fields ", err)
		return
	}
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpTypes(false, true)
	testDefinition.DumpValues(false)
	err = testDefinition.CreateValues(false)
	if !assert.NoError(t, err) {
		fmt.Println("Error create values", err)
		return
	}
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "P1C,4,B,GC1-N,1,A,GS1-N,1,A.", req.FormatBuffer.String())
	rerr = testDefinition.RestrictFieldSlice([]string{"GC[N]"})
	if !assert.NoError(t, rerr) {
		fmt.Println("Restrict request", rerr)
		return
	}
	adabasParameter = &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "GCN,1,A.", req.FormatBuffer.String())

}

func lobDefinition() *Definition {
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "CH"),
		NewType(FieldTypeLBString, "LB"),
		NewType(FieldTypeString, "ST"),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
	}
	groupLayout[1].SetLength(0)
	groupLayout[2].SetLength(0)
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewStructureList(FieldTypeGroup, "G1", OccNone, groupLayout),
		NewType(FieldTypeUByte, "UB"),
	}
	layout[0].AddOption(FieldOptionUQ)
	for _, l := range layout {
		l.SetLevel(1)
	}

	return NewDefinitionWithTypes(layout)

}

func TestDefinitionStoreBigLob(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := lobDefinition()
	assert.Nil(t, testDefinition.Values)
	testDefinition.CreateValues(false)
	assert.NotNil(t, testDefinition.Values)
	// testDefinition.DumpTypes(false, false)
	// testDefinition.DumpValues(false)
	Central.Log.Debugf("Test: no second call, read")
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false, BlockSize: 2345}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "U4,4,B,CH,1,A,LBL,4,LB(1,2345),ST,0,A,UB,1,B.", req.FormatBuffer.String())
	s := testDefinition.Search("LB").(*stringValue)
	s.lobSize = 1000000

	Central.Log.Debugf("Test: second call, read")
	adabasParameter = &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 1, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// TODO Preparation for chunked store lobs
	assert.Equal(t, "LB(4097,995904).", req.FormatBuffer.String())
	//groupLayout[1].SetLength(160000)
	s.value = make([]byte, 160000)
	Central.Log.Debugf("Test: no second call, store")
	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "U4,4,B,CH,1,A,LB(1,40960),ST,1,A,UB,1,B.", req.FormatBuffer.String())
	Central.Log.Debugf("Test: second call, store")
	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 1, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "LB(40961,40960).", req.FormatBuffer.String())
	Central.Log.Debugf("Test: second call, store")
	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 3, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "LB(122881,37120).", req.FormatBuffer.String())

}

func TestDefinitionLob(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := lobDefinition()
	assert.Nil(t, testDefinition.Values)
	testDefinition.CreateValues(false)
	assert.NotNil(t, testDefinition.Values)
	err = testDefinition.RestrictFieldSlice([]string{"LB", "U4"})
	if !assert.NoError(t, err) {
		fmt.Println("restrict request", err)
		return
	}
	// testDefinition.DumpTypes(false, false)
	// testDefinition.DumpValues(false)
	Central.Log.Debugf("Test: no second call, read")
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false, BlockSize: 12345}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "U4,4,B,LBL,4,LB(1,12345).", req.FormatBuffer.String())

	err = testDefinition.RestrictFieldSlice([]string{"LB(1,100)", "U4"})
	if !assert.NoError(t, err) {
		fmt.Println("Restrict request with partial lob", err)
		return
	}
	f := testDefinition.activeFields["LB"]
	assert.Equal(t, "  2, LB, 0, A ,LB ; LB", f.String())
	// testDefinition.DumpTypes(false, false)
	// testDefinition.DumpValues(false)
	Central.Log.Debugf("Test: no second call, read")
	adabasParameter = &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// TODO Implement range for partial lob
	assert.Equal(t, "U4,4,B,LB(1,100).", req.FormatBuffer.String())
	s := testDefinition.Search("LB").(*stringValue)
	s.lobSize = 1000000
	s.value = make([]byte, 160000)

	adabasParameter = &AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr = testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// TODO Implement range for partial lob
	assert.Equal(t, "U4,4,B,LB(1,100).", req.FormatBuffer.String())

}

func TestDefinitionMultipleField(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	err = testDefinition.ShouldRestrictToFieldSlice([]string{"U4", "GM"})
	assert.NoError(t, err)
	testDefinition.DumpTypes(false, true, "GM")
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// Reset tree
	err = testDefinition.ShouldRestrictToFields("*")
	// TODO Implement range for partial lob
	assert.Equal(t, "U4,4,B,GMC,4,B,GM1-N,1,P.", req.FormatBuffer.String())
	err = testDefinition.ShouldRestrictToFieldSlice([]string{"U4", "PM[1]"})
	assert.NoError(t, err)
}

func TestDefinitionSingleIndex(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	err = testDefinition.ShouldRestrictToFieldSlice([]string{"U4", "PM[1]"})
	assert.NoError(t, err)
	testDefinition.DumpTypes(false, true, "PM[1]")
	// testDefinition.DumpValues(false)
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// TODO Implement range for partial lob
	assert.Equal(t, "U4,4,B,PM1C,4,B,PM1(1-N),10,A.", req.FormatBuffer.String())
}

func TestDefinitionBothIndexes(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	err = testDefinition.ShouldRestrictToFieldSlice([]string{"U4", "PM[1,2]"})
	assert.NoError(t, err)
	testDefinition.DumpTypes(false, true, "PM[1,2]")
	testDefinition.DumpValues(false)
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	// TODO Implement range for partial lob
	assert.Equal(t, "U4,4,B,PM1(2),10,A.", req.FormatBuffer.String())
}

func TestDefinitionLength(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayout()
	err = testDefinition.ShouldRestrictToFields("#ISN,U4,#GS,@GS")
	assert.NoError(t, err)
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	helper := NewDynamicHelper(endian())
	req.RecordBuffer = helper
	req.Parser = testParser
	req.Limit = 1
	req.Multifetch = 1
	req.Isn = 10
	req.Definition = testDefinition
	if !assert.NotNil(t, req.Definition) {
		return
	}
	req.RecordBuffer.PutInt32(2)
	req.RecordBuffer.PutInt32(10)
	// TODO do not parse references!!!
	//req.RecordBuffer.PutInt64(0)
	req.RecordBuffer.offset = 0
	assert.Equal(t, "GSL,4,B,U4,4,B.", req.FormatBuffer.String())
	count := uint64(0)
	result, err := req.ParseBuffer(&count, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), result)
	testDefinition.DumpValues(false)
	assert.Equal(t, FieldTypeFieldLength, testDefinition.Values[0].Type().Type())
	assert.Equal(t, "2", testDefinition.Values[0].String())
}

func TestDefinitionLastPeriodEntry(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	Central.Log.Infof("_______ Restrict fields")
	err = testDefinition.ShouldRestrictToFields("GS[N]")
	assert.NoError(t, err)
	Central.Log.Infof("_______ Create request")
	adabasParameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, rerr := testDefinition.CreateAdabasRequest(adabasParameter)
	if !assert.NoError(t, rerr) {
		fmt.Println("Create request", rerr)
		return
	}
	assert.Equal(t, "GSN,1,A.", req.FormatBuffer.String())
	helper := NewDynamicHelper(endian())
	req.RecordBuffer = helper
	req.Parser = testParser
	req.Limit = 1
	req.Multifetch = 1
	req.Isn = 10
	req.Definition = testDefinition
	if !assert.NotNil(t, req.Definition) {
		return
	}
	req.RecordBuffer.putByte('a')
	req.RecordBuffer.offset = 0
	ty, terr := testDefinition.SearchType("GM")
	assert.NoError(t, terr)
	aty := ty.(*StructureType)
	assert.Equal(t, FieldTypeMultiplefield, aty.Type())
	count := uint64(0)
	Central.Log.Infof("_______ Parse buffer")
	result, err := req.ParseBuffer(&count, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), result)
	testDefinition.DumpValues(false)
	assert.Equal(t, FieldTypePeriodGroup, testDefinition.Values[0].Type().Type())
	v := testDefinition.Values[0].(*StructureValue)
	assert.Equal(t, 1, v.NrElements())
}

func ExampleDefinition_treecopy() {
	err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	testDefinition := createPeriodGroupSuperDescriptor()
	testDefinition.DumpTypes(false, false)
	if testDefinition.activeFieldTree != testDefinition.fileFieldTree {
		fmt.Println("ERROR equal")
	}
	err = testDefinition.copyActiveTree()
	if err != nil {
		fmt.Println("ERROR copyActiveTree", err)
		return
	}
	if testDefinition.activeFieldTree == testDefinition.fileFieldTree {
		fmt.Println("ERROR equal")
	}
	testDefinition.RemoveSpecialDescriptors()
	testDefinition.DumpTypes(false, false)

	// Output:
	// 	Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//  S1= ; S1
	//
	// Dump all file field types:
	//   1, U4, 4, B  ; U4
	//   1, B1, 1, F  ; B1
	//   1, UB, 1, B  ; UB
	//   1, I2, 2, B  ; I2
	//   1, U8, 8, B  ; U8
	//   1, GR ,PE ; GR
	//     2, GC, 1, A  ; GC
	//     2, GM, 5, P ,MU; GM
	//       3, GM, 5, P  ; GM
	//     2, GS, 1, A  ; GS
	//     2, GP, 1, P  ; GP
	//   1, I8, 8, B  ; I8
	//  S1= ; S1

}

func TestDefinitionRestrictCheck(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	for _, x := range [][]string{{"U4,I2,U8", "[U4 I2 U8]"}, {"U4,B1,U8", "[U4 B1 U8]"},
		{"GP", "[PG GP]"}, {"PG,U8", "[U8 PG GC GS PM PM GP]"}} {
		testDefinition := createLayoutWithPEandMU()
		Central.Log.Infof("_______ Restrict fields")
		err = testDefinition.ShouldRestrictToFields(x[0])
		fmt.Println("TEST -> ", x[0], testDefinition.Fieldnames())
		if !assert.Equal(t, x[1], fmt.Sprintf("%v", testDefinition.Fieldnames())) && !assert.NoError(t, err) {
			return
		}
		// testDefinition.DumpTypes(false, true, x[0])
	}

	testDefinition := createLayoutWithPEandMU()
	Central.Log.Infof("_______ Restrict fields")
	err = testDefinition.ShouldRestrictToFields("PM[1]")
	testDefinition.DumpTypes(false, true, "AA")
	testDefinition.DumpValues(false)
	fmt.Println("TEST -> ", testDefinition.Fieldnames())
	if !assert.NoError(t, err) {
		return
	}
	pmField := testDefinition.activeFields["PM"]
	fmt.Printf("%d", pmField.Type())
	assert.NotNil(t, pmField)
	assert.Equal(t, "PM", pmField.Name())
	assert.Equal(t, 1, pmField.MultipleRange().from)
	assert.Equal(t, -2, pmField.MultipleRange().to)
	assert.Nil(t, pmField.PartialRange())

	testDefinition = createLayoutWithPEandMU()
	Central.Log.Infof("_______ Restrict fields")
	err = testDefinition.ShouldRestrictToFields("PM[1][1]")
	fmt.Println("TEST -> ", testDefinition.Fieldnames())
	if !assert.NoError(t, err) {
		return
	}
	pmField = testDefinition.activeFields["PM"]
	assert.NotNil(t, pmField)
	assert.Equal(t, "PM", pmField.Name())
	assert.Equal(t, 1, pmField.MultipleRange().from)
	assert.Equal(t, -2, pmField.MultipleRange().to)
	assert.Nil(t, pmField.PartialRange())
}

func TestDefinitionPEMUSingle(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerLobField()
	err = testDefinition.ShouldRestrictToFields("GM[1,2]")
	assert.Nil(t, err)
	testDefinition.CreateValues(false)
	testDefinition.DumpTypes(false, true, "XX")
	testDefinition.DumpValues(false)
	v, err := testDefinition.SearchByIndex("GM", []uint32{1, 2}, true)
	testDefinition.DumpTypes(false, true, "XX")
	testDefinition.DumpValues(false)

	// Assert Nil
	assert.Nil(t, err)
	assert.Equal(t, "   3, GM, 0, A ,NU,NV,NB,MU,LB ; GM", v.Type().String())
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err)

	assert.Equal(t, "GM1(2),0,A.",
		request.FormatBuffer.String())

}

func TestDefinitionPEMUFieldSingle(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerLobField()
	err = testDefinition.ShouldRestrictToFields("GM[1,2]")
	if !assert.Nil(t, err) {
		return
	}
	testDefinition.DumpTypes(true, true, "Before values")
	testDefinition.DumpValues(true)
	Central.Log.Debugf("Create values ... for testing")
	testDefinition.CreateValues(false)
	testDefinition.DumpValues(true)
	v := testDefinition.Search("GM[1][2]")
	if !assert.NotNil(t, v) {
		fmt.Printf("v=%#v", v)
		return
	}
	testDefinition.DumpTypes(false, true, "XX")
	testDefinition.DumpValues(false)

	// Assert Nil
	assert.Nil(t, err, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err, err)

	sc, scerr := testDefinition.SearchType("GM")
	assert.Nil(t, scerr)
	assert.Equal(t, "  2, GM, 0, A ,NU,NV,NB,MU,LB; GM",
		sc.String())
	// fmt.Printf("%T %s -> %v - [%s][%s]", sc, sc, scerr, sc.PartialRange().FormatBuffer(), sc.PeriodicRange().FormatBuffer())

	assert.Equal(t, "GM1(2),0,A.",
		request.FormatBuffer.String())

}

func TestDefinitionPEMUFieldTwo(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerLobField()
	err = testDefinition.ShouldRestrictToFields("GM[1,2],GM[1,11]")
	if !assert.Nil(t, err) {
		return
	}
	testDefinition.DumpTypes(true, true, "Before values")
	testDefinition.DumpValues(true)
	Central.Log.Debugf("Create values ... for testing")
	testDefinition.CreateValues(false)
	testDefinition.DumpValues(true)
	v := testDefinition.Search("GM[1][11]")
	if !assert.NotNil(t, v) {
		fmt.Printf("v=%#v", v)
		return
	}
	testDefinition.DumpTypes(false, true, "XX")
	testDefinition.DumpValues(false)

	// Assert Nil
	assert.Nil(t, err, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err, err)

	sc, scerr := testDefinition.SearchType("GM")
	assert.Nil(t, scerr)
	assert.Equal(t, "  2, GM, 0, A ,NU,NV,NB,MU,LB; GM",
		sc.String())
	// fmt.Printf("%T %s -> %v - [%s][%s]", sc, sc, scerr, sc.PartialRange().FormatBuffer(), sc.PeriodicRange().FormatBuffer())

	assert.Equal(t, "GM1(2),0,A,GM1(11),0,A.",
		request.FormatBuffer.String())

}

func TestDefinitionPEMUFieldTwoPeriods(t *testing.T) {
	err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerLobField()
	err = testDefinition.ShouldRestrictToFields("GM[1,2],GM[2,11]")
	if !assert.Nil(t, err) {
		return
	}
	testDefinition.DumpTypes(true, true, "Before values")
	testDefinition.DumpValues(true)
	Central.Log.Debugf("Create values ... for testing")
	testDefinition.CreateValues(false)
	testDefinition.DumpValues(true)
	v := testDefinition.Search("GM[2][11]")
	if !assert.NotNil(t, v) {
		fmt.Printf("v=%#v", v)
		return
	}
	testDefinition.DumpTypes(false, true, "XX")
	testDefinition.DumpValues(false)

	// Assert Nil
	assert.Nil(t, err, err)
	parameter := &AdabasRequestParameter{Store: false, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	request, err := testDefinition.CreateAdabasRequest(parameter)
	assert.Nil(t, err, err)

	sc, scerr := testDefinition.SearchType("GM")
	assert.Nil(t, scerr)
	assert.Equal(t, "  2, GM, 0, A ,NU,NV,NB,MU,LB; GM",
		sc.String())
	// fmt.Printf("%T %s -> %v - [%s][%s]", sc, sc, scerr, sc.PartialRange().FormatBuffer(), sc.PeriodicRange().FormatBuffer())

	assert.Equal(t, "GM1(2),0,A,GM2(11),0,A.",
		request.FormatBuffer.String())

}

func employeeDefinition() *Definition {
	multipleLayout := []IAdaType{
		NewTypeWithLength(FieldTypePacked, "AT", 5),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(FlagOptionMUGhost)
	}
	multipleLayout[0].AddOption(FieldOptionMU)
	multipleLayout[0].AddOption(FieldOptionNU)
	multipleLayout[0].AddOption(FieldOptionNB)
	multipleLayout[0].AddOption(FieldOptionNV)
	peGroupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeString, "AR", 3),
		NewTypeWithLength(FieldTypePacked, "AS", 3),
		NewStructureList(FieldTypeMultiplefield, "AT", OccNone, multipleLayout),
	}
	groupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeString, "AC", 20),
		NewTypeWithLength(FieldTypeString, "AD", 20),
		NewTypeWithLength(FieldTypeString, "AE", 20),
	}
	groupLayout2 := []IAdaType{
		NewTypeWithLength(FieldTypeString, "AN", 6),
		NewTypeWithLength(FieldTypeString, "AM", 15),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
		l.AddOption(FieldOptionNU)
	}
	for _, l := range groupLayout2 {
		l.SetLevel(2)
		l.AddOption(FieldOptionNU)
	}
	// groupLayout[1].AddOption(FieldOptionMU)
	layout := []IAdaType{
		NewTypeWithLength(FieldTypeString, "AA", 8),
		NewStructureList(FieldTypeGroup, "AB", OccNone, groupLayout),
		NewStructureList(FieldTypeGroup, "A2", OccNone, groupLayout2),
		NewStructureList(FieldTypePeriodGroup, "AQ", OccNone, peGroupLayout),
	}
	for i, l := range layout {
		l.SetLevel(1)
		if i != 0 {
			l.AddOption(FieldOptionNU)
		}
	}
	layout[0].AddOption(FieldOptionDE)
	layout[0].AddOption(FieldOptionUQ)
	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func TestEmployeeDefinition(t *testing.T) {
	definitionEmployees := employeeDefinition()
	fmt.Println(definitionEmployees.String())
	req, err := CreateTestRequest(true, definitionEmployees)
	assert.NoError(t, err)
	err = definitionEmployees.CreateValues(true)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "AA,8,A,AC,20,A,AD,20,A,AE,20,A,AN,6,A,AM,15,A.", req.FormatBuffer.String())
}

func CreateTestRequest(store bool, testDefinition *Definition) (*Request, error) {
	if testDefinition == nil {
		return nil, fmt.Errorf("test definition not defined")
	}
	adabasParameter := &AdabasRequestParameter{Store: store, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, err := testDefinition.CreateAdabasRequest(adabasParameter)
	if err != nil {
		fmt.Println("Create request", err)
		return nil, err
	}
	helper := NewDynamicHelper(endian())
	req.RecordBuffer = helper
	req.Parser = testParser
	req.Limit = 1
	req.Multifetch = 1
	req.Isn = 10
	req.Definition = testDefinition
	req.RecordBuffer.PutInt32(2)
	req.RecordBuffer.PutInt32(10)
	req.RecordBuffer.offset = 0
	return req, nil
}
