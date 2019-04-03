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
	"math"
	"strconv"
)

type unpackedValue struct {
	adaValue
	value []byte
}

func newUnpackedValue(initType IAdaType) *unpackedValue {
	if initType == nil {
		return nil
	}
	value := unpackedValue{adaValue: adaValue{adatype: initType}}
	value.LongToUnpacked(0, int(initType.Length()), false)
	return &value
}

func (value *unpackedValue) ByteValue() byte {
	return value.value[0]
}

func (value *unpackedValue) String() string {
	unpackedInt := value.unpackedToLong(false)
	return strconv.FormatInt(unpackedInt, 10)
}

func (value *unpackedValue) Value() interface{} {
	return value.value
}

func (value *unpackedValue) Bytes() []byte {
	return value.value
}

// SetStringValue set the string value of the value
func (value *unpackedValue) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 10, 64)
	if err == nil {
		value.LongToUnpacked(iv, int(value.adaValue.adatype.Length()), false)
	}
}

func (value *unpackedValue) SetValue(v interface{}) error {
	Central.Log.Debugf("Set packed value to %v", v)
	iLen := value.Type().Length()
	switch v.(type) {
	case []byte:
		bv := v.([]byte)
		switch {
		case iLen != 0 && uint32(len(bv)) > iLen:
			return NewGenericError(59)
		case uint32(len(bv)) < iLen:
			copy(value.value, bv)
		default:
			value.value = bv
		}
		Central.Log.Debugf("Use byte array")
	default:
		v, err := value.commonInt64Convert(v)
		if err != nil {
			return err
		}
		Central.Log.Debugf("Got ... %v", v)
		if iLen != 0 {
			err = value.checkValidValue(v, value.Type().Length())
			if err != nil {
				return err
			}
		} else {
			iLen = value.createLength(v)
		}

		value.LongToUnpacked(v, int(iLen), false)
		Central.Log.Debugf("Packed value %s", value.String())
	}
	return nil
}

func (value *unpackedValue) checkValidValue(intValue int64, len uint32) error {
	maxValue := int64(1)
	for i := uint32(0); i < len; i++ {
		maxValue *= 10
	}
	absValue := int64(math.Abs(float64(intValue)))
	Central.Log.Debugf("Check valid value absolute value %d < max %d", absValue, maxValue)
	if absValue < maxValue {
		return nil
	}
	return NewGenericError(57, value.Type().Name(), intValue, len)
}

func (value *unpackedValue) createLength(v int64) uint32 {
	maxValue := int64(1)
	cipher := uint32(0)
	for maxValue < v {
		maxValue *= 10
		cipher++
	}
	cipher = cipher + 1
	value.value = make([]byte, cipher)
	Central.Log.Debugf("Create size of %d for %d", cipher, v)
	return cipher
}

func (value *unpackedValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := value.commonFormatBuffer(buffer, option)
	if len == 0 {
		len = 29
	}
	return len
}

func (value *unpackedValue) StoreBuffer(helper *BufferHelper) error {
	if value.Type().Length() == 0 {
		if len(value.value) > 0 {
			err := helper.putByte(byte(len(value.value) + 1))
			if err != nil {
				return err
			}
			return helper.putBytes(value.value)
		}
		return helper.putBytes([]byte{2, 0x30})

	}
	return helper.putBytes(value.value)
}

func (value *unpackedValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		length, errh := helper.ReceiveUInt8()
		if errh != nil {
			return EndTraverser, errh
		}
		if uint8(len(value.value)) != length-1 {
			value.value = make([]byte, length-1)
		}
	}
	value.value, err = helper.ReceiveBytes(uint32(len(value.value)))
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get unpacked offset=%d", helper.offset)
		Central.Log.Debugf("GOT: %s", FormatBytes("UNPACK -> ", value.value, 0, -1))
	}
	return
}

func (value *unpackedValue) Int32() (int32, error) {
	return int32(value.unpackedToLong(false)), nil
}

func (value *unpackedValue) UInt32() (uint32, error) {
	return uint32(value.unpackedToLong(false)), nil
}
func (value *unpackedValue) Int64() (int64, error) {
	return int64(value.unpackedToLong(false)), nil
}
func (value *unpackedValue) UInt64() (uint64, error) {
	return uint64(value.unpackedToLong(false)), nil
}
func (value *unpackedValue) Float() (float64, error) {
	return float64(value.unpackedToLong(false)), nil
}

func (value *unpackedValue) unpackedToLong(ebcdic bool) int64 {
	end := len(value.value) - 1
	// In case it is variable length
	if end < 0 {
		return 0
	}
	for (end > 0) && (value.value[end] == 0) {
		end--
	}

	v := make([]byte, end+1)
	copy(v[:], value.value[:end+1])
	longValue := int64(0)
	base := int64(1)
	for i := end; i >= 0; i-- {
		longValue += (int64(v[i]) & 0xf) * base
		base *= 10
	}
	Central.Log.Debugf("Index %d of %d", end, len(v))
	if ebcdic {
		if (v[end+1-1] & 0xf0) < 0xf0 {
			longValue = -longValue
		}
	} else {
		if (v[end+1-1] & 0xf0) > 0x30 {
			longValue = -longValue
		}
	}
	Central.Log.Debugf("unpacked to long %v -> %d", ebcdic, longValue)

	return longValue
}

func (value *unpackedValue) LongToUnpacked(intValue int64, len int, ebcdic bool) {
	Central.Log.Debugf("Convert integer %d", intValue)
	b := make([]byte, len)
	upperByte := uint8(0x30)
	negativByte := uint8(0x70)
	if ebcdic {
		upperByte = uint8(0xf0)
		negativByte = uint8(0xd0)
	}
	var v int64
	if intValue < 0 {
		v = -intValue
	} else {
		v = intValue
	}
	for i := len - 1; i >= 0; i-- {
		x := int64(v % 10)
		v = (v - x) / 10
		b[i] = uint8(int64(upperByte) | x)
	}
	if intValue < 0 {
		b[len-1] = uint8(negativByte | (b[len-1] & 0xf))
	}
	Central.Log.Debugf("Unpacked byte array %X", b)
	value.value = b
}
