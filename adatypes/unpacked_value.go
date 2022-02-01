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
	"bytes"
	"math"
	"strconv"
)

// unpackedValue handle Adabas fields with the Packed format
// type. The unpacked value is defined with corresponding
// values in an byte.
type unpackedValue struct {
	adaValue
	value []byte
}

// newUnpackedValue creates new unpacked value
func newUnpackedValue(initType IAdaType) *unpackedValue {
	if initType == nil {
		return nil
	}
	value := unpackedValue{adaValue: adaValue{adatype: initType}}
	value.LongToUnpacked(0, int(initType.Length()), false)
	return &value
}

// ByteValue byte value of the unpacked value
func (value *unpackedValue) ByteValue() byte {
	return value.value[0]
}

// String string value representation of the unpacked value
func (value *unpackedValue) String() string {
	unpackedInt := value.unpackedToLong(false)
	sv := strconv.FormatInt(unpackedInt, 10)
	if value.Type().Fractional() > 0 {
		l := uint32(len(sv))
		if l <= value.Type().Fractional() {
			var buffer bytes.Buffer
			buffer.WriteString("0.")
			for i := l; i < value.Type().Fractional(); i++ {
				buffer.WriteRune('0')
			}
			buffer.WriteString(sv)
			sv = buffer.String()
		} else {
			sv = sv[:l-value.Type().Fractional()] + "." + sv[l-value.Type().Fractional():]

		}
	}
	return sv

}

// Value return the raw value of the unpacked value
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

// SetValue set the unpacked value
func (value *unpackedValue) SetValue(v interface{}) error {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set packed value to %v", v)
	}
	iLen := value.Type().Length()
	switch bv := v.(type) {
	case []byte:
		switch {
		case iLen != 0 && uint32(len(bv)) > iLen:
			return NewGenericError(59)
		case uint32(len(bv)) < iLen:
			copy(value.value, bv)
		default:
			value.value = bv
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Use byte array")
		}
	default:
		v, err := value.commonInt64Convert(v)
		if err != nil {
			return err
		}
		if iLen != 0 {
			err = value.checkValidValue(v, value.Type().Length())
			if err != nil {
				return err
			}
		} else {
			iLen = value.createLength(v)
		}

		value.LongToUnpacked(v, int(iLen), false)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Packed value %s", value.String())
		}
	}
	return nil
}

func (value *unpackedValue) checkValidValue(intValue int64, len uint32) error {
	maxValue := int64(1)
	for i := uint32(0); i < len; i++ {
		maxValue *= 10
	}
	absValue := int64(math.Abs(float64(intValue)))
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Check valid value absolute value %d < max %d", absValue, maxValue)
	}
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
	len := value.commonFormatBuffer(buffer, option, value.Type().Length())
	if len == 0 {
		len = 29
	}
	return len
}

// StoreBuffer store buffer generating the record buffer used in the Adabas
// call
func (value *unpackedValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
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

// parseBuffer parse the record buffer defined by the corresponding definition
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
		Central.Log.Debugf("GOT: %s", FormatBytes("UNPACK -> ", value.value, len(value.value), 0, -1, false))
	}
	return
}

// Int8 unpacked value returns the 8-byte integer
func (value *unpackedValue) Int8() (int8, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return int8(value.unpackedToLong(false)), nil
}

// Uint8 unpacked value returns the 8-byte unsigned integer
func (value *unpackedValue) UInt8() (uint8, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return uint8(value.unpackedToLong(false)), nil
}

// Int16 unpacked value returns the 16-byte signed integer
func (value *unpackedValue) Int16() (int16, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return int16(value.unpackedToLong(false)), nil
}

// UInt16 unpacked value returns the 16-byte unsigned integer
func (value *unpackedValue) UInt16() (uint16, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return uint16(value.unpackedToLong(false)), nil
}

// Int32 unpacked value returns the 32-byte signed integer
func (value *unpackedValue) Int32() (int32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return int32(value.unpackedToLong(false)), nil
}

// UInt32 unpacked value returns the 32-byte unsigned integer
func (value *unpackedValue) UInt32() (uint32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return uint32(value.unpackedToLong(false)), nil
}

// Int64 unpacked value returns the 64-byte signed integer
func (value *unpackedValue) Int64() (int64, error) {
	v := value.unpackedToLong(false)
	if value.Type().Fractional() > 0 {
		m := int64(fractional(value.Type().Fractional()))
		if v-(v%m) == v {
			return v / m, nil
		}
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return v, nil
}

// UInt64 unpacked value returns the 64-byte unsigned integer
func (value *unpackedValue) UInt64() (uint64, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return uint64(value.unpackedToLong(false)), nil
}

// Float unpacked value returns the floating point representation
func (value *unpackedValue) Float() (float64, error) {
	v := float64(value.unpackedToLong(false))
	if value.Type().Fractional() > 0 {
		v = v / float64(fractional(value.Type().Fractional()))
	}
	return v, nil
}

// unpackedToLong convert unpacked to long value
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

// LongToUnpacked convert long to unpacked  value
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
