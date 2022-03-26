/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntByte(t *testing.T) {
	err := initLogWithFile("byte.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeByte, "XX")
	int2 := newByteValue(adaType)
	assert.Equal(t, int8(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, int8(2), int2.value)
	bint2 := int2.Bytes()
	// fmt.Println(bint2)
	assert.Equal(t, 1, len(bint2))
	int2.SetStringValue("100")
	assert.Equal(t, int8(100), int2.value)

	int2.SetValue(100)
	assert.Equal(t, int8(100), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(100), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(100), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(100), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 100.0, fl)

	err = int2.SetValue(-1)
	assert.NoError(t, err)

	assert.Equal(t, int8(-1), int2.value)
	assert.Equal(t, []byte{0xff}, int2.Bytes())
	int2.SetValue(-2)
	assert.Equal(t, int8(-2), int2.value)
	assert.Equal(t, []byte{0xfe}, int2.Bytes())
	// fmt.Println(int2.String())

	b := []byte{0xff, 0xfe, 0, 1, 2, 126, 127, 128, 129}
	iv := []int8{-1, -2, 0, 1, 2, 126, int8(math.MaxInt8), int8(math.MinInt8), -127}
	// fmt.Printf("Range %d to %d\n", int8(math.MaxInt8), int8(math.MinInt8))
	assert.Equal(t, int8(b[0]), int8(-1))
	assert.Equal(t, int8(b[1]), int8(-2))
	for i, bv := range b {
		assert.Equal(t, iv[i], convert(bv))
	}
	//	assert.Equal(t, 0x45, byte(0xff))
	assert.Equal(t, int8(-1), convert(b[0]))
	assert.Equal(t, int8(-2), convert(b[1]))
}

func convert(b byte) int8 {
	var i8 int8
	if b > 126 {
		i8 = -int8(256 - int(b))
	} else {
		i8 = int8(b)
	}
	return i8
}

func TestUIntByte(t *testing.T) {
	err := initLogWithFile("byte.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeUByte, "XX")
	int2 := newUByteValue(adaType)
	assert.Equal(t, uint8(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, uint8(2), int2.value)
	bint2 := int2.Bytes()
	fmt.Println(bint2)
	assert.Equal(t, 1, len(bint2))
	int2.SetStringValue("50")
	assert.Equal(t, uint8(50), int2.value)

	int2.SetValue(100)
	assert.Equal(t, uint8(100), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(100), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(100), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(100), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 100.0, fl)

}
