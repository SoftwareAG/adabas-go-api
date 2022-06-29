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

package adatypes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructureValueGroup(t *testing.T) {
	err := initLogWithFile("structure_value.log")
	if !assert.NoError(t, err) {
		return
	}

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
	sl := NewStructureList(FieldTypeGroup, "GR", OccByte, groupLayout)
	assert.Equal(t, "GR", sl.Name())
	assert.Equal(t, " 1, GR  ; GR", sl.String())
	v, err := sl.Value()
	vsl := v.(*StructureValue)
	vsl.initSubValues(0, 0, true)
	assert.NoError(t, err)
	assert.Equal(t, "", vsl.String())
	vpm := vsl.search("PM")
	assert.NotNil(t, vpm)
	assert.Equal(t, 1, vsl.NrElements())
	assert.NotNil(t, vsl.Value())
	eui32, errui32 := vsl.UInt32()
	assert.Equal(t, uint32(0), eui32)
	assert.Error(t, errui32)
	eui64, errui64 := vsl.UInt64()
	assert.Equal(t, uint64(0), eui64)
	assert.Error(t, errui64)

	option := &BufferOption{}
	var buffer bytes.Buffer
	vsl.FormatBuffer(&buffer, option)
	assert.Equal(t, "", buffer.String())

	b := make([]byte, 100)
	b[0] = 2
	helper := NewHelper(b, 100, endian())
	_, err = vsl.parseBuffer(helper, option)
	assert.NoError(t, err)
}

func TestStructureValuePeriod(t *testing.T) {
	err := initLogWithFile("structure_value.log")
	if !assert.NoError(t, err) {
		return
	}

	groupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeCharacter, "GC", 1),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	sl := NewStructureList(FieldTypePeriodGroup, "PE", OccByte, groupLayout)
	assert.Equal(t, "PE", sl.Name())
	assert.Equal(t, " 1, PE ,PE ; PE", sl.String())
	v, err := sl.Value()
	assert.NoError(t, err)
	vsl := v.(*StructureValue)
	b := make([]byte, 100)
	if bigEndian() {
		b[3] = 2
	} else {
		b[0] = 2
	}
	b[4] = 'X'
	b[5] = 0x1c
	helper := NewHelper(b, 100, endian())

	option := &BufferOption{}
	_, err = vsl.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, "", vsl.String())
	vpm := vsl.search("GS")
	assert.NotNil(t, vpm)
	assert.Equal(t, 2, vsl.NrElements())
	assert.NotNil(t, vsl.Value())
	eui32, errui32 := vsl.UInt32()
	assert.Equal(t, uint32(0), eui32)
	assert.Error(t, errui32)
	eui64, errui64 := vsl.UInt64()
	assert.Equal(t, uint64(0), eui64)
	assert.Error(t, errui64)

	var buffer bytes.Buffer
	vsl.FormatBuffer(&buffer, option)
	assert.Equal(t, "PEC,4,B,PE1-N", buffer.String())

	gc1a := vsl.Get("GC", 1)
	assert.NotNil(t, gc1a)
	assert.Equal(t, uint8(0x0), gc1a.Value())
	gc1b := vsl.Get("GC", 1)
	assert.NotNil(t, gc1b)
	assert.Equal(t, uint8(0x0), gc1b.Value())
	assert.Equal(t, gc1a, gc1b)
	gc2 := vsl.Get("GC", 2)
	assert.NotNil(t, gc2)
	assert.Equal(t, uint8(0x0), gc2.Value())
	assert.NotEqual(t, gc1a, gc2)

}

func TestStructureValuePeriodMU(t *testing.T) {
	err := initLogWithFile("structure_value.log")
	if !assert.NoError(t, err) {
		return
	}

	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(FlagOptionMUGhost)
	}
	groupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeCharacter, "GC", 1),
		NewStructureList(FieldTypeMultiplefield, "PM", OccByte, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}

	for _, l := range groupLayout {
		l.SetLevel(1)
		l.AddFlag(FlagOptionMUGhost)
		if l.Type() == FieldTypeMultiplefield {
			s := l.(*StructureType)
			l.AddFlag(FlagOptionAtomicFB)
			s.occ = OccCapacity
		}
	}
	sl := NewStructureList(FieldTypePeriodGroup, "PE", OccByte, groupLayout)
	sl.AddFlag(FlagOptionMUGhost)
	sl.AddFlag(FlagOptionAtomicFB)
	assert.Equal(t, "PE", sl.Name())
	assert.Equal(t, " 1, PE ,PE ; PE", sl.String())
	v, err := sl.Value()
	assert.NoError(t, err)
	vsl := v.(*StructureValue)
	b := make([]byte, 100)
	if bigEndian() {
		b[3] = 2
	} else {
		b[0] = 2
	}
	b[4] = 'X'
	b[5] = 0x1c
	helper := NewHelper(b, 100, endian())

	option := &BufferOption{}
	assert.Equal(t, NoneSecond, option.NeedSecondCall)
	var buffer bytes.Buffer
	vsl.FormatBuffer(&buffer, option)
	assert.Equal(t, "PEC,4,B", buffer.String())
	_, err = vsl.parseBuffer(helper, option)
	assert.Equal(t, ReadSecond, option.NeedSecondCall)
	assert.NoError(t, err)
	assert.Equal(t, "", vsl.String())
	vpm := vsl.search("PM")
	assert.NotNil(t, vpm)
	assert.Equal(t, 2, vsl.NrElements())
	assert.NotNil(t, vsl.Value())
	eui32, errui32 := vsl.UInt32()
	assert.Equal(t, uint32(0), eui32)
	assert.Error(t, errui32)
	eui64, errui64 := vsl.UInt64()
	assert.Equal(t, uint64(0), eui64)
	assert.Error(t, errui64)

	buffer.Reset()
	option.SecondCall = 1
	vsl.FormatBuffer(&buffer, option)
	assert.Equal(t, "", buffer.String())

	gc1a := vsl.Get("GC", 1)
	assert.NotNil(t, gc1a)
	assert.Equal(t, uint8(0x0), gc1a.Value())
	gc1b := vsl.Get("GC", 1)
	assert.NotNil(t, gc1b)
	assert.Equal(t, uint8(0x0), gc1b.Value())
	assert.Equal(t, gc1a, gc1b)
	gc2 := vsl.Get("GC", 2)
	if !assert.NotNil(t, gc2) {
		return
	}
	assert.Equal(t, uint8(0x0), gc2.Value())
	assert.NotEqual(t, gc1a, gc2)

	_, err = vsl.UInt32()
	assert.Error(t, err)
	_, err = vsl.UInt64()
	assert.Error(t, err)
	_, err = vsl.Int32()
	assert.Error(t, err)
	_, err = vsl.Int64()
	assert.Error(t, err)
	_, err = vsl.Float()
	assert.Error(t, err)
}

func TestStructureValuePeriodLast(t *testing.T) {
	err := initLogWithFile("structure_value.log")
	if !assert.NoError(t, err) {
		return
	}

	groupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeCharacter, "GC", 1),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	sl := NewStructureList(FieldTypePeriodGroup, "PE", OccByte, groupLayout)
	sl.SetRange(NewRange(1, 2))
	assert.Equal(t, "PE", sl.Name())
	assert.Equal(t, " 1, PE ,PE ; PE", sl.String())
	v, err := sl.Value()
	assert.NoError(t, err)
	vsl := v.(*StructureValue)
	b := make([]byte, 100)
	if bigEndian() {
		b[3] = 2
	} else {
		b[0] = 2
	}
	b[4] = 'X'
	b[5] = 0x1c
	helper := NewHelper(b, 100, endian())

	option := &BufferOption{}
	_, err = vsl.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, "", vsl.String())
	vpm := vsl.search("GS")
	assert.NotNil(t, vpm)
	assert.Equal(t, 2, vsl.NrElements())
	assert.NotNil(t, vsl.Value())
	eui32, errui32 := vsl.UInt32()
	assert.Equal(t, uint32(0), eui32)
	assert.Error(t, errui32)
	eui64, errui64 := vsl.UInt64()
	assert.Equal(t, uint64(0), eui64)
	assert.Error(t, errui64)

	var buffer bytes.Buffer
	vsl.FormatBuffer(&buffer, option)
	assert.Equal(t, "PEC,4,B,PE1-2", buffer.String())

	gc1a := vsl.Get("GC", 1)
	assert.NotNil(t, gc1a)
	assert.Equal(t, uint8(0x0), gc1a.Value())
	gc1b := vsl.Get("GC", 1)
	assert.NotNil(t, gc1b)
	assert.Equal(t, uint8(0x0), gc1b.Value())
	assert.Equal(t, gc1a, gc1b)
	gc2 := vsl.Get("GC", 2)
	assert.NotNil(t, gc2)
	assert.Equal(t, uint8(0x0), gc2.Value())
	assert.NotEqual(t, gc1a, gc2)

}
