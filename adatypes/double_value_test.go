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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDouble(t *testing.T) {
	err := initLogWithFile("double.log")
	if !assert.NoError(t, err) {
		return
	}

	adaType := NewType(FieldTypeDouble, "DL")
	adaType.length = 8
	fl := newDoubleValue(adaType)
	assert.Equal(t, float64(0), fl.Value())
	fl.SetStringValue("0.1")
	assert.Equal(t, float64(0.1), fl.Value())
	fl.SetStringValue("10.1")
	assert.Equal(t, float64(10.1), fl.Value())
	fl.SetValue(0.5)
	assert.Equal(t, float64(0.5), fl.Value())

}

func TestDoubleCheck(t *testing.T) {
	err := initLogWithFile("double.log")
	if !assert.NoError(t, err) {
		return
	}

	adaType := NewType(FieldTypeDouble, "FL")
	adaType.length = 8
	fl := newDoubleValue(adaType)
	assert.Equal(t, float64(0), fl.Value())
	fl.SetStringValue("0.1")
	assert.Equal(t, float64(0.1), fl.Value())
	fl.SetStringValue("10.1")
	if bigEndian() {
		assert.Equal(t, []byte{0x40, 0x24, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}, fl.Bytes())
	} else {
		assert.Equal(t, []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x24, 0x40}, fl.Bytes())
	}
	assert.Equal(t, float64(10.1), fl.Value())
	fl.SetStringValue("-10.1")
	if bigEndian() {
		assert.Equal(t, []byte{0xc0, 0x24, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}, fl.Bytes())
	} else {
		assert.Equal(t, []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x24, 0xc0}, fl.Bytes())
	}
	assert.Equal(t, float64(-10.1), fl.Value())
	fl.SetValue(0.5)
	assert.Equal(t, float64(0.5), fl.Value())
	_, serr := fl.Int32()
	assert.Error(t, serr)
	if bigEndian() {
		assert.Equal(t, []byte{0x3f, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, fl.Bytes())
	} else {
		assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe0, 0x3f}, fl.Bytes())
	}
	fl.SetValue("10.0")
	if bigEndian() {
		assert.Equal(t, []byte{0x40, 0x24, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, fl.Bytes())
	} else {
		assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x24, 0x40}, fl.Bytes())
	}
	assert.Equal(t, float64(10.0), fl.Value())
	u32int, e32err := fl.UInt32()
	assert.NoError(t, e32err)
	assert.Equal(t, uint32(10), u32int)
	u64int, e64err := fl.UInt64()
	assert.NoError(t, e64err)
	assert.Equal(t, uint64(10), u64int)
	i32int, i32err := fl.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(10), i32int)
	i64int, i64err := fl.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(10), i64int)
	fl.SetValue(float32(20.1))
	assert.True(t, math.Abs(fl.Value().(float64)-20.1) < 0.001)
	fl.SetValue(float64(21.1))
	assert.Equal(t, float64(21.1), fl.Value())
	fl.SetValue(uint32(22))
	assert.Equal(t, float64(22.0), fl.Value())
	fl.SetValue(uint64(23))
	assert.Equal(t, float64(23.0), fl.Value())
	if bigEndian() {
		fl.SetValue([]byte{0xc0, 0x24, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	} else {
		fl.SetValue([]byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x24, 0xc0})
	}
	assert.Equal(t, float64(-10.1), fl.Value())
}
