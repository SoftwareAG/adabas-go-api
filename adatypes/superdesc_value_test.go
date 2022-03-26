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

func TestSuperDesc(t *testing.T) {
	opt := byte(0x1)
	superType := NewSuperType("AA", opt)
	superType.AddSubEntry("AX", 1, 3)
	superType.AddSubEntry("AZ", 1, 2)
	superType.FdtFormat = 'A'

	v, err := superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "", v.String())
	sv := v.(*superDescValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	sv.FormatBuffer(&buffer, option)
	assert.Equal(t, "AA,5,A", buffer.String())
	helper := NewHelper([]byte{0x1, 0x2, 0x3, 0x4, 0xff}, 100, endian())
	sv.parseBuffer(helper, option)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4, 0xff}, helper.Buffer())
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4, 0xff}, sv.Bytes())
	assert.Nil(t, sv.SetValue("123"))
	assert.Equal(t, byte(' '), sv.ByteValue())
	assert.Equal(t, uint32(5), helper.Offset())
	sv.StoreBuffer(helper, nil)
	assert.Equal(t, uint32(5), helper.Offset())
	_, err = sv.Int32()
	assert.Error(t, err)
	_, err = sv.Int64()
	assert.Error(t, err)
	_, err = sv.UInt32()
	assert.Error(t, err)
	_, err = sv.UInt64()
	assert.Error(t, err)
	_, err = sv.Float()
	assert.Error(t, err)

}

func createExampleDefinition() *Definition {
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewStructureList(FieldTypeGroup, "PG", OccSingle, groupLayout),
		NewTypeWithLength(FieldTypeString, "A1", 10),
		NewType(FieldTypeInt4, "I4"),
		NewTypeWithLength(FieldTypeString, "B1", 10),
		NewTypeWithLength(FieldTypeString, "C1", 10),
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func TestSuperDescString(t *testing.T) {
	opt := byte(0x1)
	superType := NewSuperType("AA", opt)
	superType.AddSubEntry("A1", 1, 3)
	superType.AddSubEntry("B1", 1, 2)
	superType.FdtFormat = 'A'

	superType.InitSubTypes(createExampleDefinition())
	v, err := superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "'   ' '  '", v.String())
	err = v.SetValue([]byte("ABCDEFG"))
	assert.NoError(t, err)
	assert.Equal(t, "'ABC' 'DE'", v.String())

	err = v.SetValue([]byte("ABCDEFG"))
	assert.NoError(t, err)
	assert.Equal(t, "'ABC' 'DE'", v.String())

	err = v.SetValue([]byte("ABC"))
	assert.NoError(t, err)
	assert.Equal(t, "'ABC' '  '", v.String())

	superType = NewSuperType("II", opt)
	superType.AddSubEntry("A1", 1, 3)
	superType.AddSubEntry("I4", 1, 4)
	superType.FdtFormat = 'A'
	superType.InitSubTypes(createExampleDefinition())
	v, err = superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "'   ' 0", v.String())
	err = v.SetValue([]byte("ABC"))
	assert.NoError(t, err)
	assert.Equal(t, "'ABC' 0", v.String())
	err = v.SetValue([]byte("ABCDE"))
	assert.NoError(t, err)
	assert.Equal(t, "'ABC' 17732", v.String())
	err = v.SetValue([]byte("A"))
	assert.NoError(t, err)
	assert.Equal(t, "'A  ' 0", v.String())

}

func TestSubDescString(t *testing.T) {
	opt := byte(0x1)
	superType := NewSuperType("AA", opt)
	superType.AddSubEntry("A1", 1, 3)
	superType.FdtFormat = 'A'

	superType.InitSubTypes(createExampleDefinition())
	v, err := superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "   ", v.String())
	err = v.SetValue([]byte("ABCDEFG"))
	assert.NoError(t, err)
	assert.Equal(t, "ABC", v.String())

	err = v.SetValue([]byte("ABCDEFG"))
	assert.NoError(t, err)
	assert.Equal(t, "ABC", v.String())

	err = v.SetValue([]byte("ABC"))
	assert.NoError(t, err)
	assert.Equal(t, "ABC", v.String())

	superType = NewSuperType("II", opt)
	superType.AddSubEntry("I4", 1, 4)
	superType.FdtFormat = 'A'
	superType.InitSubTypes(createExampleDefinition())
	v, err = superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "0", v.String())
	err = v.SetValue([]byte("ABC"))
	assert.NoError(t, err)
	assert.Equal(t, "0", v.String())
	err = v.SetValue([]byte("ABCDE"))
	assert.NoError(t, err)
	assert.Equal(t, "1145258561", v.String())
	err = v.SetValue([]byte("A"))
	assert.NoError(t, err)
	assert.Equal(t, "65", v.String())

}
