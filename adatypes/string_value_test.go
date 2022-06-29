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
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringNil(t *testing.T) {
	adaValue := newStringValue(nil)
	assert.Nil(t, adaValue)
}

func TestStringValue(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "", adaValue.String())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC", adaValue.String())
}

func TestStringTruncate(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 2
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "  ", adaValue.String())
	adaValue.SetValue("AB")
	v = []byte{0x41, 0x42}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
	err = adaValue.SetValue("ABCD")
	if !assert.Error(t, err) {
		return
	}
	err = adaValue.SetValue("AB")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
	v = []byte{0x41, 0x42, 0x43}
	err = adaValue.SetValue(v)
	if !assert.NoError(t, err) {
		return
	}

	v = []byte{0x41, 0x42}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
}

func TestStringSpaces(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "          ", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	assert.NoError(t, adaValue.SetValue("ABC"))
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC       ", adaValue.String())
	assert.NoError(t, adaValue.SetValue("äöüß"))
	v = []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "äöüß  ", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	assert.Equal(t, byte(0xc3), adaValue.ByteValue())
	v = []byte{0x41, 0x42, 0x43}
	assert.NoError(t, adaValue.SetValue(v))
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC       ", adaValue.String())
	adaValue.SetStringValue("ANCDX")
	v = []byte{0x41, 0x4e, 0x43, 0x44, 0x58, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ANCDX     ", adaValue.String())
}

func TestStringInvalid(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
	i32, err := adaValue.Int32()
	assert.Equal(t, int32(0), i32)
	assert.Error(t, err)
	ui32, uierr := adaValue.UInt32()
	assert.Equal(t, uint32(0), ui32)
	assert.Error(t, uierr)
	i64, i64err := adaValue.Int64()
	assert.Equal(t, int64(0), i64)
	assert.Error(t, i64err)
	ui64, ui64err := adaValue.UInt64()
	assert.Equal(t, uint64(0), ui64)
	assert.Error(t, ui64err)
	fl, flerr := adaValue.Float()
	assert.Equal(t, 0.0, fl)
	assert.Error(t, flerr)
}

func TestStringFormatBuffer(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XX,10,A", buffer.String())
	assert.Equal(t, uint32(10), len)
}

func TestStringFormatBufferVariable(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XX,0,A", buffer.String())
	assert.Equal(t, uint32(253), len)
}

func TestStringLBFormatBufferVariable(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeLBString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XXL,4,XX(1,4096)", buffer.String())
	assert.Equal(t, uint32(4100), len)
}

func TestStringStoreBuffer(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	option := &BufferOption{}
	err = adaValue.StoreBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
}

func TestStringStoreBufferVariable(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	option := &BufferOption{}
	err = adaValue.StoreBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, helper.Buffer())
}

func TestStringParseBufferVariable(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	helper := &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x9, 0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}}
	var res TraverseResult
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, adaValue.Value())
	helper = &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x4, 0x41, 0x42, 0x43}}
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, adaValue.Value())
}

func TestStringLBParseBufferVariable(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeLBString, "LB")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	checkInfo := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x0, 0x0, 0x0, 0x0}
	buffer.Write([]byte{0xc, 0x0, 0x0, 0x0})
	buffer.Write(checkInfo)
	gs := 4100 - buffer.Len()
	buffer.Write(make([]byte, gs))
	assert.Equal(t, 4100, len(buffer.Bytes()))
	//	helper := &BufferHelper{order: binary.LittleEndian, buffer: []byte{0xc, 0x0, 0x0, 0x0, 0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}}
	helper := &BufferHelper{order: binary.LittleEndian, buffer: buffer.Bytes()}
	var res TraverseResult
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, checkInfo, adaValue.Value())
	buffer.Reset()
	v := []byte{0x41, 0x42, 0x43, 0x0, 0x0, 0x0, 0x0}
	buffer.Write([]byte{0x7, 0x0, 0x0, 0x0})
	buffer.Write(v)
	gs = 4100 - buffer.Len()
	buffer.Write(make([]byte, gs))
	assert.Equal(t, 4100, len(buffer.Bytes()))
	helper = &BufferHelper{order: binary.LittleEndian, buffer: buffer.Bytes()}
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, v, adaValue.Value())
}

func TestStringConverter(t *testing.T) {
	err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeString, "XX")
	typ.SetCharset("ISO-8859-15")
	typ.length = 13
	adaValue := newStringValue(typ)
	assert.Equal(t, "             ", adaValue.String())
	err = adaValue.SetValue([]byte{97, 98, 99, 36, 164, 252, 228, 246, 40, 41, 33, 43, 35})
	assert.NoError(t, err)
	assert.Equal(t, "abc$€üäö()!+#", adaValue.String())
	typ.SetCharset("windows-1251")
	err = adaValue.SetValue([]byte{207, 238, 234, 243, 239, 224, 242, 229, 235, 232})
	assert.NoError(t, err)
	assert.Equal(t, "Покупатели   ", adaValue.String())
}
