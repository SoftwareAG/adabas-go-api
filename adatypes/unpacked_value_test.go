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
	"bytes"
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUnpackedNil(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	up := newUnpackedValue(nil)
	assert.Nil(t, up)
}

func TestUnpacked(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUnpacked, "UP")
	adaType.length = 4
	up := newUnpackedValue(adaType)
	fmt.Println("Unpacked value ", up.value)
	up.LongToUnpacked(0, 4, false)
	fmt.Println("Unpacked value 0 ", up.value)
	up.SetValue(1)
	i32, i32err := up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(1), i32)
	up.SetValue(uint32(2))
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	i32, i32err = up.Int32()
	assert.NoError(t, i32err)
	assert.Equal(t, int32(2), i32)
	up.SetValue("123")
	u32, u32err := up.UInt32()
	assert.NoError(t, u32err)
	assert.Equal(t, uint32(123), u32)
	assert.Equal(t, []byte{0x30, 0x31, 0x32, 0x33}, up.Value())
	up.SetValue(uint64(1))
	u64, u64err := up.UInt64()
	assert.NoError(t, u64err)
	assert.Equal(t, uint64(1), u64)
	err = up.SetValue(12345)
	assert.Error(t, err)
	err = up.SetValue(1234)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x31, 0x32, 0x33, 0x34}, up.Value())
	err = up.SetValue(-1234)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x31, 0x32, 0x33, 0x74}, up.Value())
	i64, i64err := up.Int64()
	assert.NoError(t, i64err)
	assert.Equal(t, int64(-1234), i64)

}

func TestUnpackedFormatBuffer(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUnpacked, "UP")
	adaType.length = 4
	up := newUnpackedValue(adaType)
	fmt.Println("Unpacked value ", up.value)
	option := &BufferOption{}
	var buffer bytes.Buffer
	len := up.FormatBuffer(&buffer, option)
	assert.Equal(t, "UP,4,U", buffer.String())
	assert.Equal(t, uint32(4), len)
}
