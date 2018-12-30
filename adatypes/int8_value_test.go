package adatypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInt8Nil(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	up := newInt8Value(nil)
	assert.Nil(t, up)
}

func TestInt8(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
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
	v = []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	up.SetValue(v)
	assert.Equal(t, int64(1), up.Value())
	up.SetValue(1024)
	assert.Equal(t, int64(1024), up.Value())
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
}

func TestInt8FormatBuffer(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
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
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
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
	helper = &BufferHelper{order: binary.LittleEndian, buffer: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5}}
	res, err = up.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, TraverseResult(0), res)
	assert.Equal(t, int64(360287970189639680), up.Value())
}

func TestUInt8Nil(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	up := newUInt8Value(nil)
	assert.Nil(t, up)
}

func TestUInt8(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
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
}

func TestUInt8FormatBuffer(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUInt8, "U8")
	up := newUInt8Value(adaType)
	fmt.Println("Integer 8 value ", up.value)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := up.FormatBuffer(&buffer, option)
	assert.Equal(t, "U8,8,B", buffer.String())
	assert.Equal(t, uint32(8), len)
}
