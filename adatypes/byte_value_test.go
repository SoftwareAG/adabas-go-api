package adatypes

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIntByte(t *testing.T) {
	f, err := initLogWithFile("byte.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeByte, "XX")
	int2 := newByteValue(adaType)
	assert.Equal(t, int8(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, int8(2), int2.value)
	bint2 := int2.Bytes()
	fmt.Println(bint2)
	assert.Equal(t, 1, len(bint2))
	int2.SetStringValue("100")
	assert.Equal(t, int8(100), int2.value)

	int2.SetValue(100)
	assert.Equal(t, int8(100), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(100), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(100), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(100), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 100.0, fl)

}

func TestUIntByte(t *testing.T) {
	f, err := initLogWithFile("byte.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUByte, "XX")
	int2 := newUByteValue(adaType)
	assert.Equal(t, uint8(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, uint8(2), int2.value)
	bint2 := int2.Bytes()
	fmt.Println(bint2)
	assert.Equal(t, 1, len(bint2))
	int2.SetStringValue("50")
	assert.Equal(t, uint8(50), int2.value)

	int2.SetValue(100)
	assert.Equal(t, uint8(100), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(100), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(100), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(100), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 100.0, fl)

}
