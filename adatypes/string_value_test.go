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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	typ := NewType(FieldTypeString, "XX")
	typ.length = 2
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	v := []byte{0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "  ", adaValue.String())
	adaValue.SetValue("ABC")
	v = []byte{0x41, 0x42}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "AB", adaValue.String())
	err = adaValue.SetValue("ABCD")
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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
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
	adaValue.SetStringValue("ANCDX")
	v = []byte{0x41, 0x4e, 0x43, 0x44, 0x58, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, adaValue.Value())
	assert.Equal(t, "ANCDX     ", adaValue.String())
}

func TestStringInvalid(t *testing.T) {
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
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

func TestStringStoreBuffer(t *testing.T) {
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	typ := NewType(FieldTypeString, "XX")
	typ.length = 10
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	err = adaValue.StoreBuffer(helper)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x41, 0x42, 0x43, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	assert.Equal(t, v, helper.Buffer())
}

func TestStringStoreBufferVariable(t *testing.T) {
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	typ := NewType(FieldTypeString, "XX")
	typ.length = 0
	adaValue := newStringValue(typ)
	assert.NotNil(t, adaValue)
	adaValue.SetValue("äöüß")
	helper := &BufferHelper{}
	err = adaValue.StoreBuffer(helper)
	if !assert.NoError(t, err) {
		return
	}
	v := []byte{0x9, 0xc3, 0xa4, 0xc3, 0xb6, 0xc3, 0xbc, 0xc3, 0x9f}
	assert.Equal(t, v, helper.Buffer())
	adaValue.SetValue("ABC")
	helper = &BufferHelper{}
	err = adaValue.StoreBuffer(helper)
	if !assert.NoError(t, err) {
		return
	}
	v = []byte{0x4, 0x41, 0x42, 0x43}
	assert.Equal(t, v, helper.Buffer())
}

func TestStringParseBufferVariable(t *testing.T) {
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
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
	f, err := initLogWithFile("string_value.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	typ := NewType(FieldTypeLBString, "LB")
	typ.length = 0
	adaValue := newStringValue(typ)
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
