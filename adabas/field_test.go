package adabas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldQuery(t *testing.T) {
	fq, err := NewFieldQuery("AA")
	assert.NoError(t, err)
	assert.Equal(t, ' ', fq.Prefix)
	assert.Equal(t, "AA", fq.Name)
	assert.Equal(t, uint32(0), fq.PeriodicIndex)
	assert.Equal(t, uint32(0), fq.MultipleIndex)
	fq, err = NewFieldQuery("AB[1]")
	assert.NoError(t, err)
	assert.Equal(t, ' ', fq.Prefix)
	assert.Equal(t, "AB", fq.Name)
	assert.Equal(t, uint32(1), fq.PeriodicIndex)
	assert.Equal(t, uint32(0), fq.MultipleIndex)
	fq, err = NewFieldQuery("AC[1][2]")
	assert.NoError(t, err)
	assert.Equal(t, ' ', fq.Prefix)
	assert.Equal(t, "AC", fq.Name)
	assert.Equal(t, uint32(1), fq.PeriodicIndex)
	assert.Equal(t, uint32(2), fq.MultipleIndex)
	fq, err = NewFieldQuery("AD[3,4]")
	assert.NoError(t, err)
	assert.Equal(t, ' ', fq.Prefix)
	assert.Equal(t, "AD", fq.Name)
	assert.Equal(t, uint32(3), fq.PeriodicIndex)
	assert.Equal(t, uint32(4), fq.MultipleIndex)
	fq, err = NewFieldQuery("#AA")
	assert.NoError(t, err)
	assert.Equal(t, '#', fq.Prefix)
	assert.Equal(t, "AA", fq.Name)
	assert.Equal(t, uint32(0), fq.PeriodicIndex)
	assert.Equal(t, uint32(0), fq.MultipleIndex)
}
