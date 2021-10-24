/*
* Copyright © 2018-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt8Nil(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	up := newInt8Value(nil)
	assert.Nil(t, up)
}

func TestInt8(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt8, "I8")
	up := newInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	assert.Equal(t, int64(0), up.Value())
	up.SetValue(1000)
	assert.Equal(t, int64(1000), up.Value())
	up.SetValue(int64(math.MinInt64))
	assert.Equal(t, int64(math.MinInt64), up.Value())
	up.SetValue(int64(math.MaxInt64))
	assert.Equal(t, int64(math.MaxInt64), up.Value())
	i32, i32err := up.Int32()
	assert.Error(t, i32err)
	assert.Equal(t, int32(0), i32)
	v := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	up.SetValue(v)
	assert.Equal(t, int64(-1), up.Value())
	up.SetValue(0)
	assert.Equal(t, int64(0), up.Value())
	if bigEndian() {
		v = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}
	} else {
		v = []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	}
	up.SetValue(v)
	assert.Equal(t, int64(1), up.Value())
	up.SetValue(1024)
	assert.Equal(t, int64(1024), up.Value())
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := up.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := up.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := up.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)
	v = []byte{0x80, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	assert.Equal(t, v, up.Bytes())
	assert.Equal(t, "1024", up.String())

	up.SetValue(1234)
	ui32, ui32err := up.UInt32()
	assert.NoError(t, ui32err)
	assert.Equal(t, uint32(1234), ui32)
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1234), i32)

	up.SetValue(-1234)
	ui32, ui32err = up.UInt32()
	assert.Error(t, ui32err)
	assert.Equal(t, uint32(0), ui32)
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(-1234), i32)

}

func TestInt8Variable(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt8, "I8")
	adaType.SetLength(0)
	up := newInt8Value(adaType)
	checkValueInt64(t, up, []byte{2, 1}, 1)
	checkValueInt64(t, up, []byte{2, 255}, -1)
	checkValueInt64(t, up, []byte{3, 1, 1}, 0x101)
	checkValueInt64(t, up, []byte{4, 1, 1, 1}, 65793)
	if bigEndian() {
		checkValueInt64(t, up, []byte{4, 255, 0, 0}, 16711680)
	} else {
		checkValueInt64(t, up, []byte{4, 0, 0, 255}, 16711680)
	}
	checkValueInt64(t, up, []byte{5, 1, 1, 1, 1}, 0x1010101)
}

func checkValueInt64(t *testing.T, up IAdaValue, input []byte, expect int64) {
	helper := NewDynamicHelper(endian())
	helper.putBytes(input)
	helper.offset = 0
	option := &BufferOption{}
	_, perr := up.parseBuffer(helper, option)
	assert.NoError(t, perr)
	v, err := up.Int64()
	assert.NoError(t, err)
	assert.Equal(t, expect, v)

}

func ExampleIAdaType_setValue() {
	err := initLogWithFile("unpacked.log")
	if err != nil {
		fmt.Println("Error enable log")
		return
	}

	adaType := NewType(FieldTypeInt8, "I8")
	up := newInt8Value(adaType)
	fmt.Println("Integer default value :", up.value)
	up.SetValue(1000)
	fmt.Printf("Integer 1000 value : %d %T\n", up.value, up.value)
	up.SetValue(int64(math.MinInt64))
	fmt.Printf("Integer minimal value : %d %T\n", up.value, up.value)
	up.SetValue(int64(math.MaxInt64))
	fmt.Printf("Integer maximal value : %d %T\n", up.value, up.value)
	err = up.SetValue(int8(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 10 (8bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(int16(100))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 100 (16bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(int32(1000))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 1000 (32bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue("87654")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 87654 (string) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint8(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 10 (8bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint16(100))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 100 (16bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint32(1000))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 1000 (32bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue([]byte{0x50})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 80 (1-byte array) value : %d %T\n", up.value, up.value)
	err = up.SetValue([]byte{0xfe})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer -2 (1-byte array) value : %d %T\n", up.value, up.value)
	if bigEndian() {
		err = up.SetValue([]byte{0x2, 0x50})
	} else {
		err = up.SetValue([]byte{0x50, 0x2})
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 592 (2-byte array) value : %d %T\n", up.value, up.value)
	if bigEndian() {
		err = up.SetValue([]byte{0x4, 0x3, 0x2, 0x50})
	} else {
		err = up.SetValue([]byte{0x50, 0x2, 0x3, 0x4})
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 67306064 (4-byte array) value : %d %T\n", up.value, up.value)

	// Output: 	Integer default value : 0
	// Integer 1000 value : 1000 int64
	// Integer minimal value : -9223372036854775808 int64
	// Integer maximal value : 9223372036854775807 int64
	// Integer 10 (8bit) value : 10 int64
	// Integer 100 (16bit) value : 100 int64
	// Integer 1000 (32bit) value : 1000 int64
	// Integer 87654 (string) value : 87654 int64
	// Integer 10 (8bit) value : 10 int64
	// Integer 100 (16bit) value : 100 int64
	// Integer 1000 (32bit) value : 1000 int64
	// Integer 80 (1-byte array) value : 80 int64
	// Integer -2 (1-byte array) value : -2 int64
	// Integer 592 (2-byte array) value : 592 int64
	// Integer 67306064 (4-byte array) value : 67306064 int64
}

func TestInt8FormatBuffer(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt8, "I8")
	up := newInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := up.FormatBuffer(&buffer, option)
	assert.Equal(t, "I8,8,F", buffer.String())
	assert.Equal(t, uint32(8), len)
}

func TestInt8ParseBuffer(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeInt8, "I8")
	up := newInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	option := &BufferOption{}
	helper := &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}}
	var res TraverseResult
	res, err = up.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, int64(5), up.Value())
	helper = &BufferHelper{order: binary.BigEndian, buffer: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5}}
	res, err = up.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, int64(5), up.Value())
	helper = &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5}}
	res, err = up.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, int64(360287970189639680), up.Value())
}

func TestUInt8Nil(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	up := newUInt8Value(nil)
	assert.Nil(t, up)
}

func TestUInt8(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUInt8, "I8")
	up := newUInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	assert.Equal(t, uint64(0), up.Value())
	up.SetValue(1000)
	assert.Equal(t, uint64(1000), up.Value())
	up.SetValue(uint64(math.MaxUint64))
	assert.Equal(t, uint64(math.MaxUint64), up.Value())
	up.SetValue(0)
	assert.Equal(t, uint64(0), up.Value())
	up.SetValue(1024)
	assert.Equal(t, uint64(1024), up.Value())
	i32, i32err := up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := up.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := up.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := up.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)

	var v []byte
	if bigEndian() {
		v = []byte{0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4, 0x0}
	} else {
		v = []byte{0x00, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	}
	assert.Equal(t, v, up.Bytes())
	assert.Equal(t, "1024", up.String())

	up.SetValue(1234)
	ui32, ui32err := up.UInt32()
	assert.NoError(t, ui32err)
	assert.Equal(t, uint32(1234), ui32)
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1234), i32)

	err = up.SetValue(-1234)
	assert.Error(t, err)
	ui32, ui32err = up.UInt32()
	assert.NoError(t, ui32err)
	assert.Equal(t, uint32(1234), ui32)
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1234), i32)
	ui64, ui64err = up.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1234), ui64)

}

func TestUInt8Variable(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUInt8, "I8")
	adaType.SetLength(0)
	up := newUInt8Value(adaType)
	checkValueUInt64(t, up, []byte{2, 1}, 1)
	checkValueUInt64(t, up, []byte{3, 1, 1}, 0x101)
	checkValueUInt64(t, up, []byte{4, 1, 1, 1}, 0x10101)
	checkValueUInt64(t, up, []byte{5, 1, 1, 1, 1}, 0x1010101)
}

func checkValueUInt64(t *testing.T, up IAdaValue, input []byte, expect uint64) {
	helper := NewDynamicHelper(binary.LittleEndian)
	perr := helper.putBytes(input)
	assert.NoError(t, perr)
	helper.offset = 0
	option := &BufferOption{}
	_, perr = up.parseBuffer(helper, option)
	assert.NoError(t, perr)
	v, err := up.UInt64()
	assert.NoError(t, err)
	assert.Equal(t, expect, v)

}

func ExampleIAdaType_uint8SetValue() {
	err := initLogWithFile("unpacked.log")
	if err != nil {
		fmt.Println("Error enable log")
		return
	}

	adaType := NewType(FieldTypeUInt8, "U8")
	up := newUInt8Value(adaType)
	fmt.Println("Unsigned Integer default value :", up.value)
	up.SetValue(1000)
	fmt.Printf("Integer 1000 value : %d %T\n", up.value, up.value)
	err = up.SetValue(int64(math.MinInt64))
	if err == nil {
		fmt.Println("ERROR: negative value should be cause error")
		return
	}
	fmt.Println(err)
	up.SetValue(int64(math.MaxInt64))
	fmt.Printf("Integer maximal value : %d %T\n", up.value, up.value)
	err = up.SetValue(int8(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 10 (8bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(int16(100))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 100 (16bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(int32(1000))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 1000 (32bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue("87654")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 87654 (string) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint8(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 10 (8bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint16(100))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 100 (16bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint32(1000))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 1000 (32bit) value : %d %T\n", up.value, up.value)
	err = up.SetValue(uint8(80))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 80 (1-byte array) value : %d %T\n", up.value, up.value)
	err = up.SetValue([]byte{0xfe})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 254 (1-byte array) value : %d %T\n", up.value, up.value)
	if bigEndian() {
		err = up.SetValue([]byte{0x2, 0x50})
	} else {
		err = up.SetValue([]byte{0x50, 0x2})
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 592 (2-byte array) value : %d %T\n", up.value, up.value)
	if bigEndian() {
		err = up.SetValue([]byte{0x4, 0x3, 0x2, 0x50})
	} else {
		err = up.SetValue([]byte{0x50, 0x2, 0x3, 0x4})
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Integer 67306064 (4-byte array) value : %d %T\n", up.value, up.value)

	// Output: 	Unsigned Integer default value : 0
	// Integer 1000 value : 1000 uint64
	// ADG0000101: Cannot convert negative value to unsigned of type unsigned int64 (int64,-9223372036854775808)
	// Integer maximal value : 9223372036854775807 uint64
	// Integer 10 (8bit) value : 10 uint64
	// Integer 100 (16bit) value : 100 uint64
	// Integer 1000 (32bit) value : 1000 uint64
	// Integer 87654 (string) value : 87654 uint64
	// Integer 10 (8bit) value : 10 uint64
	// Integer 100 (16bit) value : 100 uint64
	// Integer 1000 (32bit) value : 1000 uint64
	// Integer 80 (1-byte array) value : 80 uint64
	// Integer 254 (1-byte array) value : 254 uint64
	// Integer 592 (2-byte array) value : 592 uint64
	// Integer 67306064 (4-byte array) value : 67306064 uint64
}

func TestUInt8FormatBuffer(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUInt8, "U8")
	up := newUInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := up.FormatBuffer(&buffer, option)
	assert.Equal(t, "U8,8,B", buffer.String())
	assert.Equal(t, uint32(8), len)
}
