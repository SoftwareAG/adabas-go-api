/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInt2Byte(t *testing.T) {
	f, err := initLogWithFile("int2.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeInt2, "XX")
	int2 := newInt2Value(adaType)
	assert.Equal(t, int16(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, int16(2), int2.value)
	bint2 := int2.Bytes()
	fmt.Println(bint2)
	assert.Equal(t, 2, len(bint2))
	int2.SetStringValue("2000")
	assert.Equal(t, int16(2000), int2.value)

	int2.SetValue(1024)
	assert.Equal(t, int16(1024), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)

}

func TestUInt2Byte(t *testing.T) {
	f, err := initLogWithFile("int2.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUInt2, "XX")
	int2 := newUInt2Value(adaType)
	assert.Equal(t, uint16(0), int2.value)
	int2.SetStringValue("2")
	assert.Equal(t, uint16(2), int2.value)
	bint2 := int2.Bytes()
	fmt.Println(bint2)
	assert.Equal(t, 2, len(bint2))
	int2.SetStringValue("2000")
	assert.Equal(t, uint16(2000), int2.value)

	int2.SetValue(1024)
	assert.Equal(t, uint16(1024), int2.Value())
	i32, i32err := int2.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1024), i32)
	i64, i64err := int2.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(1024), i64)
	ui64, ui64err := int2.UInt64()
	assert.NoError(t, ui64err)
	assert.Equal(t, uint64(1024), ui64)
	fl, flerr := int2.Float()
	assert.NoError(t, flerr)
	assert.Equal(t, 1024.0, fl)

}
