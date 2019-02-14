/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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
	f := initTestLogWithFile(t, "range.log")
	defer f.Close()
	r := NewRange(1, 2)
	assert.Equal(t, "1-2", r.FormatBuffer())
	r = NewRange(1, lastEntry)
	assert.Equal(t, "1-N", r.FormatBuffer())
	r = NewRange(2, 3)
	assert.Equal(t, "2-3", r.FormatBuffer())
	r = NewRange(3, 3)
	assert.Equal(t, "3", r.FormatBuffer())
	r = NewRange(3, 2)
	assert.Nil(t, r)
	r = NewSingleRange(2)
	assert.Equal(t, "2", r.FormatBuffer())
	r = NewSingleRange(lastEntry)
	assert.Equal(t, "N", r.FormatBuffer())
	r = NewLastRange()
	assert.Equal(t, "N", r.FormatBuffer())
	r = NewEmptyRange()
	assert.Equal(t, "", r.FormatBuffer())
}

func TestRangeParser(t *testing.T) {
	f := initTestLogWithFile(t, "range.log")
	defer f.Close()
	r := NewRangeParser("1-N")
	assert.Equal(t, "1-N", r.FormatBuffer())
	r = NewRangeParser("N")
	if !assert.NotNil(t, r) {
		return
	}
	assert.Equal(t, lastEntry, r.from)
	assert.Equal(t, lastEntry, r.to)
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
