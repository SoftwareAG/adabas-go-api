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
	"bytes"
	"encoding/binary"
	"errors"

	log "github.com/sirupsen/logrus"
)

// BufferHelper buffer helper structure used to parse the buffer
type BufferHelper struct {
	offset uint32
	buffer []byte
	max    int
	order  binary.ByteOrder
}

// BufferOverflow error indicates the read after the buffer maximal position
const BufferOverflow = -1

// NewDynamicHelper create a new buffer helper instance
func NewDynamicHelper(order binary.ByteOrder) *BufferHelper {
	return &BufferHelper{offset: 0, buffer: []byte{}, max: 0, order: order}
}

// NewHelper create a new buffer helper instance
func NewHelper(buffer []byte, max int, order binary.ByteOrder) *BufferHelper {
	return &BufferHelper{offset: 0, buffer: buffer, max: max, order: order}
}

// Shrink shrink the buffer to the given length
func (helper *BufferHelper) Shrink(length uint32) (err error) {
	helper.max = int(length)
	return
}

func (helper *BufferHelper) position(pos uint32) (newPos int, err error) {
	Central.Log.Debugf("Position to be set %v offset=%v max=%v", pos, helper.offset, helper.max)
	if helper.max >= int(pos) {
		helper.offset = uint32(pos)
	} else {
		log.Fatal("Position error", helper.offset, helper.max)
		err = NewGenericError(38, pos)
		return
	}
	newPos = int(pos)
	return
}

// Buffer buffer array
func (helper *BufferHelper) Buffer() []byte {
	return helper.buffer
}

// Remaining remaining bytes in the buffer
func (helper *BufferHelper) Remaining() int {
	remaining := helper.max - int(helper.offset)
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Offset offset of used bytes in the buffer
func (helper *BufferHelper) Offset() uint32 {
	return helper.offset
}

// ReceiveString receive string of length
func (helper *BufferHelper) ReceiveString(length uint32) (res string, err error) {
	if (helper.offset + 1) <= uint32(len(helper.buffer)) {
		b := string(helper.buffer[helper.offset : helper.offset+length])
		helper.offset += length
		res = b
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// ReceiveBytes receive bytes length
func (helper *BufferHelper) ReceiveBytes(length uint32) (res []byte, err error) {
	if (helper.offset + 1) <= uint32(len(helper.buffer)) {
		Central.Log.Debugf("Length=%d on offset %d", length, helper.offset)
		res = make([]byte, length)
		copy(res, helper.buffer[helper.offset:helper.offset+length])
		helper.offset += length
		return
	}
	err = errors.New("Buffer overflow")
	return
}

func (helper *BufferHelper) putBytes(data []byte) (err error) {
	Central.Log.Debugf("Put bytes of len %d offset=%d len=%d", len(data), helper.offset, len(helper.buffer))
	helper.buffer = append(helper.buffer, data...)
	helper.offset = uint32(len(helper.buffer))
	Central.Log.Debugf("After put bytes of len %d offset=%d len=%d", len(data), helper.offset, len(helper.buffer))
	return
}

func (helper *BufferHelper) putByte(data byte) (err error) {
	helper.buffer = append(helper.buffer, data)
	helper.offset = uint32(len(helper.buffer))
	return
}

// ReceiveBytesOcc receive bytes using a specific occurence
func (helper *BufferHelper) ReceiveBytesOcc(occ int) (res []byte, err error) {
	var length uint32
	switch occ {
	case OccByte:
		occ, subErr := helper.ReceiveUInt8()
		if subErr != nil {
			err = subErr
			return
		}
		length = uint32(occ)
		break
	default:
		err = errors.New("Occurence not defined")
		return
	}
	return helper.ReceiveBytes(length)
}

// ReceiveInt8 receive 1-byte integer
func (helper *BufferHelper) ReceiveInt8() (res int8, err error) {
	if (helper.offset + 1) <= uint32(len(helper.buffer)) {
		b := helper.buffer[helper.offset]
		helper.offset++
		res = int8(b)
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// ReceiveUInt8 receive 1-byte  unsigned integer
func (helper *BufferHelper) ReceiveUInt8() (res uint8, err error) {
	if (helper.offset + 1) <= uint32(len(helper.buffer)) {
		b := helper.buffer[helper.offset]
		helper.offset++
		res = b
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutUInt8 put 1-byte unsigned integer
func (helper *BufferHelper) PutUInt8(data uint8) (err error) {
	tmp := make([]byte, 0)
	tmp = append(tmp, data)
	Central.Log.Debugf("Temporary len %d offset=%d len=%d", len(tmp), helper.offset, len(helper.buffer))
	helper.putBytes(tmp)
	return
}

// ReceiveUInt16 receive 2-byte  unsigned integer
func (helper *BufferHelper) ReceiveUInt16() (res uint16, err error) {
	if (helper.offset + 2) <= uint32(len(helper.buffer)) {
		res = helper.order.Uint16(helper.buffer[helper.offset : helper.offset+2])
		helper.offset += 2
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutUInt16 put 2-byte unsigned integer
func (helper *BufferHelper) PutUInt16(data uint16) (err error) {
	tmp := make([]byte, 2)
	helper.order.PutUint16(tmp, data)
	helper.putBytes(tmp)
	return
}

// ReceiveInt16 receive 2-byte integer
func (helper *BufferHelper) ReceiveInt16() (res int16, err error) {
	if (helper.offset + 2) <= uint32(len(helper.buffer)) {
		buf := bytes.NewBuffer(helper.buffer[helper.offset : helper.offset+2])
		binary.Read(buf, helper.order, &res)
		helper.offset += 2
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutInt16 put 2-byte  integer
func (helper *BufferHelper) PutInt16(data int16) (err error) {
	tmp := make([]byte, 2)
	binary.PutVarint(tmp, int64(data))
	helper.putBytes(tmp)
	return
}

// ReceiveUInt32 receive 4-byte unsigned integer
func (helper *BufferHelper) ReceiveUInt32() (res uint32, err error) {
	if (helper.offset + 4) <= uint32(len(helper.buffer)) {
		res = helper.order.Uint32(helper.buffer[helper.offset : helper.offset+4])
		helper.offset += 4
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutUInt32 put 4-byte unsigned integer
func (helper *BufferHelper) PutUInt32(data uint32) (err error) {
	tmp := make([]byte, 4)
	helper.order.PutUint32(tmp, data)
	helper.putBytes(tmp)
	return
}

// ReceiveInt32 reveive 4-byte integer
func (helper *BufferHelper) ReceiveInt32() (res int32, err error) {
	if (helper.offset + 4) <= uint32(len(helper.buffer)) {
		buf := bytes.NewBuffer(helper.buffer[helper.offset : helper.offset+4])
		binary.Read(buf, helper.order, &res)
		helper.offset += 4
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutInt32 put 4-byte  integer
func (helper *BufferHelper) PutInt32(data int32) (err error) {
	tmp := make([]byte, 4)
	binary.PutVarint(tmp, int64(data))
	// helper.order.PutInt64(tmp, data)
	helper.putBytes(tmp)
	return
}

// ReceiveUInt64 reveive 8-byte unsigned integer
func (helper *BufferHelper) ReceiveUInt64() (res uint64, err error) {
	if (helper.offset + 8) <= uint32(len(helper.buffer)) {
		res = helper.order.Uint64(helper.buffer[helper.offset : helper.offset+8])
		helper.offset += 8
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutUInt64 put 8-byte unsigned integer
func (helper *BufferHelper) PutUInt64(data uint64) (err error) {
	tmp := make([]byte, 8)
	helper.order.PutUint64(tmp, data)
	helper.putBytes(tmp)
	return
}

// ReceiveInt64 reveive 8-byte integer
func (helper *BufferHelper) ReceiveInt64() (res int64, err error) {
	if (helper.offset + 8) <= uint32(len(helper.buffer)) {
		buf := bytes.NewBuffer(helper.buffer[helper.offset : helper.offset+8])
		binary.Read(buf, helper.order, &res)
		helper.offset += 8
		return
	}
	err = errors.New("Buffer overflow")
	return
}

// PutInt64 put 8-byte  integer
func (helper *BufferHelper) PutInt64(data int64) (err error) {
	tmp := make([]byte, 8)
	binary.PutVarint(tmp, data)
	// helper.order.PutInt64(tmp, data)
	helper.putBytes(tmp)
	return
}
