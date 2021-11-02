package adabas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCode(t *testing.T) {
	cc := s4
	assert.Equal(t, "S4", cc.command())
	cc = v3
	assert.Equal(t, "V3", cc.command())
	assert.Equal(t, 'V', cc.code()[0])
	assert.Equal(t, '3', cc.code()[1])

	assert.True(t, validAcbxCommand(cc.code()))
	assert.False(t, validAcbxCommand([2]byte{'X', '1'}))
}
