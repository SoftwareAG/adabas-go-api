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
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt4Byte(t *testing.T) {
	err := initLogWithFile("int4.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt4, "XX")
	int4 := newInt4Value(adaType)
	assert.Equal(t, int32(0), int4.value)
	int4.SetStringValue("2")
	assert.Equal(t, int32(2), int4.value)
	bint4 := int4.Bytes()
	fmt.Println(bint4)
	assert.Equal(t, 4, len(bint4))
	int4.SetValue(math.MaxInt32)
	assert.Equal(t, int32(math.MaxInt32), int4.value)
	var maxBuffer []byte
	if bigEndian() {
		maxBuffer = []byte{0x7f, 0xff, 0xff, 0xff}
	} else {
		maxBuffer = []byte{0xff, 0xff, 0xff, 0x7f}
	}
	assert.Equal(t, maxBuffer, int4.Bytes())
	int4.SetStringValue("2000")
	assert.Equal(t, int32(2000), int4.value)

	helper := NewHelper(maxBuffer, 4, endian())
	int4.parseBuffer(helper, NewBufferOption(false, 0))
	assert.Equal(t, int32(math.MaxInt32), int4.value)
	assert.Equal(t, maxBuffer, int4.Bytes())

	int4.SetValue(1024)
	assert.Equal(t, int32(1024), int4.Value())
	i32, i32err := int4.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := int4.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := int4.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := int4.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)

	assert.Equal(t, "1024", int4.String())

}

func TestInt4Variable(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt4, "I4")
	adaType.SetLength(0)
	up := newInt4Value(adaType)
	checkValueInt64(t, up, []byte{2, 1}, 1)
	checkValueInt64(t, up, []byte{2, 255}, -1)
	checkValueInt64(t, up, []byte{3, 1, 1}, 0x101)
	checkValueInt64(t, up, []byte{4, 1, 1, 1}, 65793)
	if bigEndian() {
		checkValueInt64(t, up, []byte{4, 1, 1, 255}, 66047)
		checkValueInt64(t, up, []byte{4, 255, 1, 1}, 16711937)
		checkValueInt64(t, up, []byte{5, 255, 255, 1, 1}, -65279)
		checkValueInt64(t, up, []byte{4, 255, 0, 0}, 16711680)
	} else {
		checkValueInt64(t, up, []byte{4, 255, 1, 1}, 66047)
		checkValueInt64(t, up, []byte{4, 1, 1, 255}, 16711937)
		checkValueInt64(t, up, []byte{5, 1, 1, 255, 255}, -65279)
		checkValueInt64(t, up, []byte{4, 0, 0, 255}, 16711680)
	}
}

func TestUInt4Variable(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUInt4, "I4")
	adaType.SetLength(0)
	up := newUInt4Value(adaType)
	checkValueUInt64(t, up, []byte{2, 1}, 1)
	checkValueUInt64(t, up, []byte{3, 1, 1}, 0x101)
	checkValueUInt64(t, up, []byte{4, 1, 1, 1}, 0x10101)
}

func TestInt4Max(t *testing.T) {
	v := make([]byte, 5)
	binary.PutUvarint(v, uint64(4294967295))
	fmt.Printf("%x\n", v)
	v = make([]byte, 4)
	endian().PutUint32(v, uint32(4294967295))
	fmt.Printf("%x\n", v)

	endian().PutUint32(v, uint32(4294967295))
	fmt.Printf("%x\n", v)

}

func TestUInt4Byte(t *testing.T) {
	err := initLogWithFile("int4.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUInt4, "XX")
	int4 := newUInt4Value(adaType)
	assert.Equal(t, uint32(0), int4.value)
	int4.SetStringValue("2")
	assert.Equal(t, uint32(2), int4.value)
	bint4 := int4.Bytes()
	fmt.Println(bint4)
	assert.Equal(t, 4, len(bint4))
	int4.SetValue(uint32(4294967295))
	assert.Equal(t, uint32(4294967295), int4.value)
	maxBuffer := []byte{0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, maxBuffer, int4.Bytes())
	int4.SetStringValue("2000")
	assert.Equal(t, uint32(2000), int4.value)

	helper := NewHelper(maxBuffer, 4, binary.LittleEndian)
	int4.parseBuffer(helper, NewBufferOption(false, 0))
	assert.Equal(t, uint32(4294967295), int4.value)
	assert.Equal(t, maxBuffer, int4.Bytes())

	int4.SetValue(1024)
	assert.Equal(t, uint32(1024), int4.Value())
	ui32, ui32err := int4.UInt32()
	assert.NoError(t, ui32err)
	assert.Equal(t, uint32(1024), ui32)
	i32, i32err := int4.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := int4.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := int4.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := int4.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)
	assert.Equal(t, "1024", int4.String())

}
