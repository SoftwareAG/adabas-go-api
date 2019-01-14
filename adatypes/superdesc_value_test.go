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
	sv.StoreBuffer(helper)
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
