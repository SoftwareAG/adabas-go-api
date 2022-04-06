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
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnicodeNil(t *testing.T) {
	adaValue := newUnicodeValue(nil)
	assert.Nil(t, adaValue)
}

func TestUnicodeValue(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "", adaValue.String())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC", adaValue.String())
	adaValue.SetValue("äöüß")
	v = []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "äöüß", adaValue.String())
	v = []byte{0x41, 0x42, 0x43}
	adaValue.SetValue(v)
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC", adaValue.String())
}

func TestUnicodeTruncate(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 2
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "  ", adaValue.String())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
	err = adaValue.SetValue("äöüß")
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0xc3, 0xa4}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ä", adaValue.String())
	v = []byte{0x41, 0x42, 0x43}
	err = adaValue.SetValue(v)
	if !assert.NoError(t, err) {
		return
	}

	v = []byte{0x41, 0x42}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
}

func TestUnicodeSpaces(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 10
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "          ", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC       ", adaValue.String())
	adaValue.SetValue("äöüß")
	v = []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "äöüß  ", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	assert.Equal(t, byte(0xc3), adaValue.ByteValue())
	v = []byte{0x41, 0x42, 0x43}
	adaValue.SetValue(v)
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC       ", adaValue.String())
}

func TestUnicodeInvalid(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 10
	adaValue := newUnicodeValue(typ)
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

func TestUnicodeFormatBuffer(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 10
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XX,10,W", buffer.String())
	assert.Equal(t, uint32(10), len)
}

func TestUnicodeFormatBufferVariable(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XX,0,W", buffer.String())
	assert.Equal(t, uint32(253), len)
}

func TestUnicodeStoreBuffer(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 10
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	err = adaValue.StoreBuffer(helper, nil)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper, nil)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetStringValue("ANCDX")
	v = []byte{0x41, 0x4e, 0x43, 0x44, 0x58, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ANCDX     ", adaValue.String())
}

func TestUnicodeStoreBufferVariable(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	err = adaValue.StoreBuffer(helper, nil)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0x9, 0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper, nil)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x4, 0x41, 0x42, 0x43}
	assert.Equal(t, v, helper.Buffer())
}

func TestUnicodeParseBufferVariable(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeUnicode, "XX")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
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

func TestUnicodeLBParseBufferVariable(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}
	typ := NewType(FieldTypeLBUnicode, "LB")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	option := &BufferOption{}
	helper := &BufferHelper{order: binary.LittleEndian, buffer: []byte{0xc, 0x0, 0x0, 0x0, 0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}}
	var res TraverseResult
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, adaValue.Value())
	helper = &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x7, 0x0, 0x0, 0x0, 0x41, 0x42, 0x43}}
	res, err = adaValue.parseBuffer(helper, option)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, TraverseResult(0), res)
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, adaValue.Value())
}

func TestUnicodeLBVariable(t *testing.T) {
	err := initLogWithFile("unicode_value.log")
	if !assert.NoError(t, err) {
		return
	}

	typ := NewType(FieldTypeLBUnicode, "XX")
	typ.length = 0
	adaValue := newUnicodeValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42, 0x43}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC", adaValue.String())
	adaValue.SetValue("äöüß")
	v = []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "äöüß", adaValue.String())
	assert.Equal(t, v, adaValue.Bytes())
	assert.Equal(t, byte(0xc3), adaValue.ByteValue())
	v = []byte{0x41, 0x42, 0x43}
	adaValue.SetValue(v)
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ABC", adaValue.String())
	option := &BufferOption{}
	var buffer bytes.Buffer
	l := adaValue.FormatBuffer(&buffer, option)
	assert.Equal(t, "XXL,4,XX(0,4096)", buffer.String())
	assert.Equal(t, uint32(0x1004), l)
}
