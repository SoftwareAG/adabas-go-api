package adatypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhonetic(t *testing.T) {
	adaType := NewType(FieldTypePhonetic, "PH")
	ph := newPhoneticValue(adaType)
	assert.Equal(t, "", ph.String())
	err := ph.SetValue(1234)
	assert.Error(t, err)
	i64, i64err := ph.Int64()
	assert.Error(t, i64err)
	assert.Equal(t, int64(0), i64)

	ui64, ui64err := ph.UInt64()
	assert.Error(t, ui64err)
	assert.Equal(t, uint64(0), ui64)

}
