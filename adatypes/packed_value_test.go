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
	"encoding/binary"
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPackedData(t *testing.T) {
	f, err := initLogWithFile("packed.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypePacked, "PA")
	adaType.length = 4
	pa := newPackedValue(adaType)
	fmt.Println("Packed value", pa.value)
	pa.LongToPacked(0, 4)
	fmt.Printf("Unpacked value  0 = %X\n", pa.value)
	assert.EqualValues(t, []byte{0x0, 0x0, 0x0, 0xb}, pa.value)

	pa.LongToPacked(10, 4)
	fmt.Printf("Unpacked value 10 = %X\n", pa.value)
	assert.EqualValues(t, []byte{0x0, 0x0, 0x1, 0xc}, pa.value)
	assert.Equal(t, "10", pa.String())

	pa.LongToPacked(9, 4)
	fmt.Printf("Unpacked value 9 = %X\n", pa.value)
	assert.EqualValues(t, []byte{0x0, 0x0, 0x0, 0x9c}, pa.value)
	assert.Equal(t, "9", pa.String())

	pa.LongToPacked(-10, 4)
	fmt.Printf("Unpacked value 10 = %X\n", pa.value)
	assert.EqualValues(t, []byte{0x0, 0x0, 0x1, 0xb}, pa.value)
	assert.Equal(t, "-10", pa.String())

	pa.SetStringValue("234560")
	assert.EqualValues(t, []byte{0x2, 0x34, 0x56, 0xc}, pa.value)
	assert.Equal(t, "234560", pa.String())

	pa.value = []byte{0x00, 0x00, 0x24, 0x61, 0x5c}
	assert.Equal(t, int64(24615), pa.packedToLong())
	if Central.IsDebugLevel() {
		fmt.Println(FormatByteBuffer("Packed format", pa.value))
	}

	err = pa.SetValue(123)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), pa.packedToLong())
	assert.Equal(t, "123", pa.String())

	assert.Equal(t, binary.LittleEndian, pa.Type().Endian())
	pa.Type().SetEndian(binary.LittleEndian)
	assert.Equal(t, binary.LittleEndian, pa.Type().Endian())
	pa.Type().SetEndian(binary.BigEndian)
	assert.Equal(t, binary.BigEndian, pa.Type().Endian())
	assert.False(t, pa.Type().IsStructure())
	assert.False(t, pa.Type().IsSpecialDescriptor())
}

func TestPackedCheckValid(t *testing.T) {
	f, err := initLogWithFile("packed.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypePacked, "PA")
	adaType.length = 1
	pa := newPackedValue(adaType)
	err = pa.SetValue(123)
	if !assert.NotNil(t, err) {
		return
	}
	assert.Error(t, err)
	assert.Equal(t, "ADG0000057: Packed value of PA validation error, value 123 does not fit into 1-packed", err.Error())
	err = pa.SetValue(9)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "9", pa.String())
}

func TestPackedCheckFractional(t *testing.T) {
	f, err := initLogWithFile("packed.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypePacked, "PA")
	adaType.length = 10
	adaType.SetFractional(2)
	pa := newPackedValue(adaType)
	err = pa.SetValue(1.23)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x3c}, pa.value)
	f64, ferr := pa.Float()
	if !assert.NoError(t, ferr) {
		return
	}
	assert.Equal(t, 1.23, f64)
	assert.Equal(t, "1.23", pa.String())
	_, err = pa.Int32()
	assert.Error(t, err)
	assert.Equal(t, "ADG0000112: Integer representation of value PA is not possible because of fractional value 2", err.Error())
	_, err = pa.Int64()
	assert.Error(t, err)
	assert.Equal(t, "ADG0000112: Integer representation of value PA is not possible because of fractional value 2", err.Error())
	err = pa.SetValue(1)
	if !assert.NoError(t, err) {
		return
	}
	i64, ierr := pa.Int64()
	assert.NoError(t, ierr)
	assert.Equal(t, int64(1), i64)

}
