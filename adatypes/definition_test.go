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

package adatypes

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefinitionGroup(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,G1,1,A,GX,1,A,PA,1,P,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionPeriodic(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,PGC,4,PG1-N,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionMultiple(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	multipleLayout[0].AddFlag(FlagOptionMU)
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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,P1,1,A,PMC,4,PM1-N,1,P,PA,1,A,PX,1,P,I8,8,B.",
		request.FormatBuffer.String())

}

func TestDefinitionQuerySimple(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,I2,2,B.",
		request.FormatBuffer.String())

}

func TestDefinitionQueryGroupField(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,GS,1,A.",
		request.FormatBuffer.String())
}

func TestDefinitionQueryGroupFieldTwice(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
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
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,I2,2,B.",
		request.FormatBuffer.String())

}

func TestDefinitionCreateValues(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)
	log.Debug(" ------------------------ after create adabas request 0 0")
	assert.Equal(t, "U4,4,B,GMC,4,GM1-N,1,P,GS,1,A.",
		request.FormatBuffer.String())

	request, err = testDefinition.CreateAdabasRequest(true, false)
	assert.Nil(t, err)

	log.Debug(" ------------------------ after create adabas request 1 0")
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
	log.Debug(" ------------------------ before create adabas request 0 0")

	request, err = testDefinition.CreateAdabasRequest(false, true)
	assert.Nil(t, err)
	log.Debug(" ------------------------ after create adabas request 0 1")

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

func ExampleDefinition_dumpValues() {
	f, err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	defer f.Close()
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpValues(false)
	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
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
	f, err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	defer f.Close()
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
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
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
	f, err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	defer f.Close()
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
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
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
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Infof("TEST: %s", t.Name())

	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.DumpTypes(false, false)
	// Generate format buffer for first read call
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)
	testDefinition.DumpValues(false)
	log.Debug(" ------------------------ after create adabas request 0 0")
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,GRC,4,GC1-N,1,A,GM1-NC,4,GS1-N,1,A,GP1-N,1,P,I8,8,B.",
		request.FormatBuffer.String())

	// Generate format buffer for first store call
	request, err = testDefinition.CreateAdabasRequest(true, false)
	assert.Nil(t, err)

	log.Debug(" ------------------------ after create adabas request 1 0")
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
	log.Debug(" ------------------------ before create adabas request with data 1 0")
	request, err = testDefinition.CreateAdabasRequest(true, false)
	assert.Nil(t, err)
	log.Debug(" ------------------------ after create adabas request with data 1 0")
	assert.Equal(t, "U4,4,B,B1,1,F,UB,1,B,I2,2,B,U8,8,B,GC1,1,A,GM1(1),5,P,GS1,1,A,GP1,1,P,I8,8,B.",
		request.FormatBuffer.String(), "Wrong store format buffer with PE/MU data")

	testDefinition.DumpValues(false)
	log.Debug(" ------------------------ before create adabas request 0 0")

	// Generate format buffer for second read call with missing PE/MU field data
	request, err = testDefinition.CreateAdabasRequest(false, true)
	assert.Nil(t, err)
	log.Debug(" ------------------------ after create adabas request 0 1")

	assert.Equal(t, "GM1(1),5,P.",
		request.FormatBuffer.String())

}

func TestDefinitionRestrictPeriodic(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	defer f.Close()
	log.Infof("TEST: %s", t.Name())

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
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,PGC,4,PG1-N.",
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
	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "GM"),
	}
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
		NewStructureList(FieldTypeMultiplefield, "GM", OccSingle, multipleLayout),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func TestDefinitionRestrictPeriodicWithMU(t *testing.T) {
	f, err := initLogWithFile("definition.log")
	if !assert.NoError(t, err) {
		return
	}

	defer f.Close()
	log.Infof("TEST: %s", t.Name())

	testDefinition := createLayoutWithPEandMU()
	testDefinition.DumpValues(false)
	err = testDefinition.ShouldRestrictToFields("U4,PG")
	request, err := testDefinition.CreateAdabasRequest(false, false)
	assert.Nil(t, err)

	assert.Equal(t, "U4,4,B,PGC,4,PG1-N.",
		request.FormatBuffer.String())
	testDefinition.DumpTypes(false, true)
}

func ExampleDefinition_dumpValuesAll() {
	f, err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Error init log ", err)
		return
	}
	defer f.Close()

	testDefinition := createLayout()

	testDefinition.DumpTypes(false, true)
	testDefinition.DumpValues(false)

	// Output: Dump all active field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, PG  ; PG  PE=false MU=false REMOVE=true
	//     2, GC, 1, A  ; GC  PE=false MU=false REMOVE=true
	//     2, GS, 1, A  ; GS  PE=false MU=false REMOVE=true
	//     2, GP, 1, P  ; GP  PE=false MU=false REMOVE=true
	//   1, GM  ; GM  PE=false MU=false REMOVE=true
	//     2, MA, 1, P  ; MA  PE=false MU=false REMOVE=true
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
	f, err := initLogWithFile("definition.log")
	if err != nil {
		fmt.Println("Error init log ", err)
		return
	}
	defer f.Close()

	testDefinition := createLayout()

	err = testDefinition.ShouldRestrictToFields("U4,PG")
	if err != nil {
		fmt.Println("Error restrict fields ", err)
		return
	}
	testDefinition.DumpTypes(false, true)
	testDefinition.DumpValues(false)

	// Output: Dump all active field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=false
	//   1, PG  ; PG  PE=false MU=false REMOVE=false
	//     2, GC, 1, A  ; GC  PE=false MU=false REMOVE=false
	//     2, GS, 1, A  ; GS  PE=false MU=false REMOVE=false
	//     2, GP, 1, P  ; GP  PE=false MU=false REMOVE=false
	//
	// Dump values :   U4 = >0<
	//  PG = [1]
	//   GC = >0<
	//   GS = > <
	//   GP = >0<

}
