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
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelperBuffer(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{1, 1, 2, 3, 4, 5, 6, 'a', 'b', 'A', 'Z'}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	assert.Equal(t, uint32(0), helper.offset)
	b, err := helper.ReceiveUInt8()
	assert.Equal(t, nil, err)
	assert.Equal(t, byte(1), b)
	assert.Equal(t, uint32(1), helper.offset)

	recBytes, err2 := helper.ReceiveBytes(uint32(6))
	assert.Equal(t, uint32(7), helper.offset)
	assert.Equal(t, nil, err2)
	assert.Equal(t, 6, len(recBytes))
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, recBytes)

	recString, err2 := helper.ReceiveString(uint32(4))
	assert.Equal(t, uint32(11), helper.offset)
	assert.Equal(t, nil, err2)
	assert.Equal(t, 6, len(recBytes))
	assert.Equal(t, "abAZ", recString)

	_, err = helper.ReceiveUInt8()
	assert.Equal(t, uint32(11), helper.offset)
	assert.Error(t, err)

	pos, posErr := helper.position(2)
	assert.NoError(t, posErr)
	assert.Equal(t, 2, pos)
	assert.Equal(t, uint32(2), helper.offset)

	recInt8, _ := helper.ReceiveInt8()
	assert.Equal(t, int8(2), recInt8)
}

func TestHelperInteger16(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{1, 0, 0, 1, 1, 2, 0x3, 0x4}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	assert.Equal(t, uint32(0), helper.offset)
	b, err := helper.ReceiveUInt16()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint16(0x1), b)
	assert.Equal(t, uint32(2), helper.offset)

	i32, err2 := helper.ReceiveInt16()
	assert.Equal(t, nil, err2)
	assert.Equal(t, int16(256), i32)
	assert.Equal(t, uint32(4), helper.offset)

	i322, err3 := helper.ReceiveInt16()
	assert.Equal(t, nil, err3)
	assert.Equal(t, int16(513), i322)
	assert.Equal(t, uint32(6), helper.offset)

	i422, err4 := helper.ReceiveUInt16()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint16(0x403), i422)
	assert.Equal(t, uint32(8), helper.offset)

}

func TestHelperInteger32(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{1, 0, 0, 0, 1, 0, 0, 0, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	assert.Equal(t, uint32(0), helper.offset)
	b, err := helper.ReceiveUInt32()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(1), b)
	assert.Equal(t, uint32(4), helper.offset)

	i32, err2 := helper.ReceiveInt32()
	assert.Equal(t, nil, err2)
	assert.Equal(t, int32(1), i32)
	assert.Equal(t, uint32(8), helper.offset)

	i322, err3 := helper.ReceiveInt32()
	assert.Equal(t, nil, err3)
	assert.Equal(t, int32(67305985), i322)
	assert.Equal(t, uint32(12), helper.offset)

	i422, err4 := helper.ReceiveUInt32()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint32(0x4030201), i422)
	assert.Equal(t, uint32(16), helper.offset)

}

func TestHelperIntegerLittleEndian64(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	assert.Equal(t, uint32(0), helper.offset)
	b, err := helper.ReceiveUInt64()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(0x1), b)
	assert.Equal(t, uint32(8), helper.offset)

	i642, err2 := helper.ReceiveInt64()
	assert.Equal(t, nil, err2)
	assert.Equal(t, int64(72057594037927936), i642)
	assert.Equal(t, uint32(16), helper.offset)

	i322, err3 := helper.ReceiveInt64()
	assert.Equal(t, nil, err3)
	assert.Equal(t, int64(578437695752307201), i322)
	assert.Equal(t, uint32(24), helper.offset)

	i422, err4 := helper.ReceiveUInt64()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint64(0x102030405060708), i422)
	assert.Equal(t, uint32(32), helper.offset)

	helper.position(16)
	assert.Equal(t, uint8(1), helper.buffer[helper.offset])
	i32, err4 := helper.ReceiveUInt32()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint32(0x04030201), i32)
	assert.Equal(t, uint32(20), helper.offset)

}

func TestHelperIntegerBigEndian64(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 8, 7, 6, 5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 6, 7, 8}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.BigEndian)
	assert.Equal(t, uint32(0), helper.offset)
	b, err := helper.ReceiveUInt64()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(0x1), b)
	assert.Equal(t, uint32(8), helper.offset)

	i642, err2 := helper.ReceiveInt64()
	assert.Equal(t, nil, err2)
	assert.Equal(t, int64(72057594037927936), i642)
	assert.Equal(t, uint32(16), helper.offset)

	i322, err3 := helper.ReceiveInt64()
	assert.Equal(t, nil, err3)
	assert.Equal(t, int64(578437695752307201), i322)
	assert.Equal(t, uint32(24), helper.offset)

	i422, err4 := helper.ReceiveUInt64()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint64(0x102030405060708), i422)
	assert.Equal(t, uint32(32), helper.offset)

	helper.position(16)
	assert.Equal(t, uint8(8), helper.buffer[helper.offset])
	i32, err4 := helper.ReceiveUInt32()
	assert.Equal(t, nil, err4)
	assert.Equal(t, uint32(0x08070605), i32)
	assert.Equal(t, uint32(20), helper.offset)

	assert.Equal(t, uint8(4), helper.buffer[helper.offset])
	i32s, err4 := helper.ReceiveInt32()
	assert.Equal(t, nil, err4)
	assert.Equal(t, int32(0x04030201), i32s)
	assert.Equal(t, uint32(24), helper.offset)

}

func TestHelperWriteDataUInt(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}
	helper := NewDynamicHelper(binary.BigEndian)
	helper.PutUInt8(1)
	assert.Equal(t, 1, len(helper.buffer))
	helper.PutUInt16(2)
	assert.Equal(t, 3, len(helper.buffer))
	helper.PutUInt32(4)
	assert.Equal(t, 7, len(helper.buffer))
	helper.PutUInt64(4)
	assert.Equal(t, 15, len(helper.buffer))
}

func TestHelperBufferBufferOverflow(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte{1, 1, 2, 3, 4, 5, 6, 'a', 'b', 'A', 'Z'}
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	res, err := helper.ReceiveBytes(100)
	if !assert.Error(t, err) {
		return
	}
	assert.Nil(t, res)
	res, err = helper.ReceiveBytes(12)
	if !assert.Error(t, err) {
		return
	}
	assert.Nil(t, res)
	res, err = helper.ReceiveBytes(11)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, byteBuffer, res)

}

func TestHelperStringBufferOverflow(t *testing.T) {
	err := initLogWithFile("buffer_helper.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	byteBuffer := []byte("ABCDEFG")
	helper := NewHelper(byteBuffer, len(byteBuffer), binary.LittleEndian)
	res, err := helper.ReceiveString(100)
	if !assert.Error(t, err) {
		return
	}
	assert.Empty(t, res)
	res, err = helper.ReceiveString(8)
	if !assert.Error(t, err) {
		return
	}
	assert.Empty(t, res)
	res, err = helper.ReceiveString(7)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "ABCDEFG", res)
	helper.offset = 0
	res, err = helper.ReceiveString(3)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "ABC", res)
	res, err = helper.ReceiveString(4)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "DEFG", res)

}
