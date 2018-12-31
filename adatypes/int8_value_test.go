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
	i64, i64err := up.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := up.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := up.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)
}

func ExampleInt8_SetValue() {
	f, err := initLogWithFile("unpacked.log")
	if err != nil {
		fmt.Println("Error enable log")
		return
	}
	defer f.Close()

	adaType := NewType(FieldTypeInt8, "I8")
	up := newInt8Value(adaType)
	fmt.Println("Integer default value :", up.value)
	up.SetValue(1000)
	fmt.Printf("Integer 1000 value : %d %T\n", up.value, up.value)
	up.SetValue(int64(math.MinInt64))
	fmt.Printf("Integer minimal value : %d %T\n", up.value, up.value)
	up.SetValue(int64(math.MaxInt64))
	fmt.Printf("Integer maximal value : %d %T\n", up.value, up.value)
	up.SetValue(int8(10))
	fmt.Printf("Integer 10 (8bit) value : %d %T\n", up.value, up.value)
	up.SetValue(int16(100))
	fmt.Printf("Integer 100 (16bit) value : %d %T\n", up.value, up.value)
	up.SetValue(int32(1000))
	fmt.Printf("Integer 1000 (32bit) value : %d %T\n", up.value, up.value)

	// Output: 	Integer default value : 0
	// Integer 1000 value : 1000 int64
	// Integer minimal value : -9223372036854775808 int64
	// Integer maximal value : 9223372036854775807 int64
	// Integer 10 (8bit) value : 9223372036854775807 int64
	// Integer 100 (16bit) value : 9223372036854775807 int64
	// Integer 1000 (32bit) value : 9223372036854775807 int64
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
