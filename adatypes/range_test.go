/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */

package adatypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeInt(t *testing.T) {
	initTestLogWithFile(t, "range.log")
	r := NewRange(1, 2)
	assert.Equal(t, "1-2", r.FormatBuffer())
	assert.Equal(t, 2, r.multiplier())
	r = NewRange(1, LastEntry)
	assert.Equal(t, "1-N", r.FormatBuffer())
	assert.Equal(t, allEntries, r.multiplier())
	assert.Equal(t, uint32(1), r.index(1, 10))
	assert.Equal(t, uint32(2), r.index(2, 10))
	assert.Equal(t, uint32(5), r.index(5, 10))
	assert.Equal(t, uint32(10), r.index(10, 10))
	r = NewRange(2, 3)
	assert.Equal(t, "2-3", r.FormatBuffer())
	assert.Equal(t, 2, r.multiplier())
	assert.Equal(t, uint32(2), r.index(1, 10))
	assert.Equal(t, uint32(3), r.index(2, 10))
	r = NewRange(3, 3)
	assert.Equal(t, "3", r.FormatBuffer())
	assert.Equal(t, 1, r.multiplier())
	assert.Equal(t, uint32(3), r.index(1, 10))
	r = NewRange(3, 2)
	assert.Nil(t, r)
	r = NewSingleRange(2)
	assert.Equal(t, "2", r.FormatBuffer())
	assert.Equal(t, 1, r.multiplier())
	r = NewSingleRange(LastEntry)
	assert.Equal(t, "N", r.FormatBuffer())
	assert.Equal(t, uint32(10), r.index(1, 10))
	assert.Equal(t, 1, r.multiplier())
	r = NewLastRange()
	assert.Equal(t, "N", r.FormatBuffer())
	assert.Equal(t, 1, r.multiplier())
	assert.Equal(t, uint32(10), r.index(1, 10))
	r = NewEmptyRange()
	assert.Equal(t, "", r.FormatBuffer())
	assert.Equal(t, 1, r.multiplier())
	assert.Equal(t, uint32(1), r.index(1, 10))
}

func TestRangeParser(t *testing.T) {
	initTestLogWithFile(t, "range.log")
	r := NewRangeParser("1-N")
	assert.Equal(t, "1-N", r.FormatBuffer())
	assert.Equal(t, allEntries, r.multiplier())
	r = NewRangeParser("N")
	if !assert.NotNil(t, r) {
		return
	}
	assert.Equal(t, LastEntry, r.from)
	assert.Equal(t, LastEntry, r.to)
	assert.Equal(t, "N", r.FormatBuffer())
	r = NewRangeParser("1")
	assert.Equal(t, "1", r.FormatBuffer())
	r = NewRangeParser("X")
	assert.Nil(t, r)
	r = NewRangeParser("1-1N")
	assert.Nil(t, r)

	r = NewRangeParser("3-2")
	assert.Nil(t, r)

	r = NewRangeParser("N-2")
	assert.Nil(t, r)

}
