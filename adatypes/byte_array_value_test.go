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

func TestByteArray(t *testing.T) {
	err := initLogWithFile("byte_array.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeByteArray, "XX")
	barray := newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0}, barray.value)

	adaType = NewTypeWithLength(FieldTypeByteArray, "XX", 2)
	barray = newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0, 0x0}, barray.value)
	assert.Equal(t, "[0 0]", barray.String())
	var buffer bytes.Buffer
	len := barray.FormatBuffer(&buffer, NewBufferOption(false, 0))
	assert.Equal(t, uint32(2), len)
	assert.Equal(t, "XX,2,B", buffer.String())

}

func TestByteArraySet(t *testing.T) {
	err := initLogWithFile("byte_array.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewTypeWithLength(FieldTypeByteArray, "XX", 0)
	barray := newByteArrayValue(adaType)
	assert.Equal(t, uint32(0), adaType.Length())
	assert.Equal(t, []byte{}, barray.value)
	barray.SetStringValue("0x1010")
	assert.Equal(t, []byte{0x10, 0x10}, barray.value)
	barray.SetStringValue("1010")
	if bigEndian() {
		assert.Equal(t, "[0 0 0 0 0 0 3 242]", barray.String())
	} else {
		assert.Equal(t, "[242 3 0 0 0 0 0 0]", barray.String())
	}

	adaType = NewTypeWithLength(FieldTypeByteArray, "XX", 2)
	barray = newByteArrayValue(adaType)
	assert.Equal(t, uint32(2), adaType.Length())
	assert.Equal(t, []byte{0x0, 0x0}, barray.value)
	assert.Equal(t, "[0 0]", barray.String())
	barray.SetStringValue("0x1010")
	assert.Equal(t, []byte{0x10, 0x10}, barray.value)
	assert.Equal(t, "[16 16]", barray.String())
	barray.SetStringValue("1010")
	if bigEndian() {
		assert.Equal(t, []byte{0x03, 0xf2}, barray.value)
		assert.Equal(t, "[3 242]", barray.String())
	} else {
		assert.Equal(t, []byte{0xf2, 0x03}, barray.value)
		assert.Equal(t, "[242 3]", barray.String())
	}

	adaType = NewTypeWithLength(FieldTypeByteArray, "XX", 8)
	barray = newByteArrayValue(adaType)
	assert.Equal(t, uint32(8), adaType.Length())
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, barray.value)
	assert.Equal(t, "[0 0 0 0 0 0 0 0]", barray.String())
	barray.SetStringValue("0x1010")
	assert.Equal(t, []byte{0x10, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, barray.value)
	assert.Equal(t, "[16 16 0 0 0 0 0 0]", barray.String())
	barray.SetStringValue("1010")
	if bigEndian() {
		assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0xf2}, barray.value)
		assert.Equal(t, "[0 0 0 0 0 0 3 242]", barray.String())
	} else {
		assert.Equal(t, []byte{0xf2, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, barray.value)
		assert.Equal(t, "[242 3 0 0 0 0 0 0]", barray.String())
	}

}
