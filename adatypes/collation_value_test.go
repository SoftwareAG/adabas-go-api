package adatypes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollationValue(t *testing.T) {
	collationType := NewCollationType("AA", 4, "PA", "de@collation=phonebook")

	v, err := collationType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "", v.String())
	sv := v.(*collationValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	sv.FormatBuffer(&buffer, option)
	assert.Equal(t, "", buffer.String())
	helper := NewHelper([]byte{0x1, 0x2, 0x3, 0x4, 0xff}, 100, endian())
	sv.parseBuffer(helper, option)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4, 0xff}, helper.Buffer())
	assert.Nil(t, sv.Bytes())
	assert.Error(t, sv.SetValue("123"))
	assert.Equal(t, byte(' '), sv.ByteValue())
	assert.Equal(t, uint32(0), helper.Offset())
	sv.StoreBuffer(helper)
	assert.Equal(t, uint32(0), helper.Offset())
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
