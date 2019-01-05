package adatypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructureValue(t *testing.T) {
	multipleLayout := []IAdaType{
		NewType(FieldTypePacked, "PM"),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
	}
	groupLayout := []IAdaType{
		NewType(FieldTypeCharacter, "GC"),
		NewStructureList(FieldTypeMultiplefield, "PM", OccNone, multipleLayout),
		NewType(FieldTypeString, "GS"),
		NewType(FieldTypePacked, "GP"),
	}
	sl := NewStructureList(FieldTypeGroup, "GR", OccNone, groupLayout)
	assert.Equal(t, "GR", sl.Name())
	assert.Equal(t, " 1, GR  ; GR  PE=false MU=false REMOVE=true", sl.String())
	v, err := sl.Value()
	vsl := v.(*StructureValue)
	assert.NoError(t, err)
	assert.Equal(t, "", vsl.String())
	vpm := vsl.search("PM")
	assert.NotNil(t, vpm)
	assert.Equal(t, 1, vsl.NrElements())
	assert.NotNil(t, vsl.Value())
	eui32, errui32 := vsl.UInt32()
	assert.Equal(t, uint32(0), eui32)
	assert.Error(t, errui32)
	eui64, errui64 := vsl.UInt64()
	assert.Equal(t, uint64(0), eui64)
	assert.Error(t, errui64)
}
