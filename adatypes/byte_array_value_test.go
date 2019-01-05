package adatypes

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestByteArray(t *testing.T) {
	f, err := initLogWithFile("byte_array.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeByteArray, "XX")
	barray := newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0}, barray.value)

	adaType = NewTypeWithLength(FieldTypeByteArray, "XX", 2)
	barray = newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0, 0x0}, barray.value)
	assert.Equal(t, "0", barray.String())
	var buffer bytes.Buffer
	len := barray.FormatBuffer(&buffer, NewBufferOption(false, false))
	assert.Equal(t, uint32(2), len)
	assert.Equal(t, "XX,2,B", buffer.String())

}
