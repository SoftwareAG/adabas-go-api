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

type packedValue struct {
	adaValue
	value []byte
}

func newPackedValue(initType IAdaType) *packedValue {
	value := packedValue{adaValue: adaValue{adatype: initType}}
	vlen := initType.Length()
	value.value = make([]byte, vlen)
	if vlen > 0 {
		value.value[vlen-1] = positivePackedIndicator()
	}
	return &value
}

func (value *packedValue) ByteValue() byte {
	return value.value[0]
}

func (value *packedValue) String() string {
	packedInt := value.packedToLong()
	sv := strconv.FormatInt(packedInt, 10)
	if value.Type().Fractional() > 0 {
		sv = sv[:value.Type().Fractional()-1] + "." + sv[value.Type().Fractional()-1:]
	}
	return sv
}

func (value *packedValue) Value() interface{} {
	return value.value
}

func (value *packedValue) Bytes() []byte {
	return value.value
}

// SetStringValue set the string value of the value
func (value *packedValue) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 10, 64)
	if err == nil {
		value.LongToPacked(iv, value.adaValue.adatype.Length())
	}
}

func (value *packedValue) SetValue(v interface{}) error {
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

		value.LongToPacked(v, iLen)
		Central.Log.Debugf("Packed value %s", value.String())
	}
	return nil
}

func (value *packedValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := value.commonFormatBuffer(buffer, option)
	if len == 0 {
		len = 15
	}
	return len
}

func (value *packedValue) StoreBuffer(helper *BufferHelper) error {
	if value.Type().Length() == 0 {
		vlen := len(value.value)
		if vlen == 0 {
			return helper.putBytes([]byte{2, positivePackedIndicator()})
		}
		Central.Log.Debugf("Create variable len=%d", vlen)
		err := helper.putByte(byte(vlen + 1))
		if err != nil {
			return err
		}
	}
	return helper.putBytes(value.value)
}

func (value *packedValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall && !value.Type().HasFlagSet(FlagOptionSecondCall) {
		Central.Log.Debugf("Skip parsing %s offset=%d", value.Type().Name(), helper.offset)
		return
	}
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
	if err != nil {
		return
	}
	Central.Log.Debugf("Buffer get packed %s -> offset=%d/%X(%d)", value.adatype.Name(), helper.offset, helper.offset, len(value.value))
	return
}

func (value *packedValue) Int32() (int32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return int32(v), nil
}

func (value *packedValue) UInt32() (uint32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint32(v), nil
}
func (value *packedValue) Int64() (int64, error) {
	v := value.packedToLong()
	if value.Type().Fractional() > 0 {
		m := int64(fractional(value.Type().Fractional()))
		if v-(v%m) == v {
			return v / m, nil
		}
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return int64(v), nil
}
func (value *packedValue) UInt64() (uint64, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint64(v), nil
}

func fractional(f uint32) uint32 {
	x := uint32(1)
	for i := uint32(0); i < f; i++ {
		x *= 10
	}
	return x
}

func (value *packedValue) Float() (float64, error) {
	v := float64(value.packedToLong())
	if value.Type().Fractional() > 0 {
		v = v / float64(fractional(value.Type().Fractional()))
	}
	return v, nil
}

func (value *packedValue) packedToLong() int64 {
	Central.Log.Debugf("Packed %v", value.value)

	if value == nil {
		return 0
	}
	base := int64(1)
	longValue := int64(0)
	sign := int64(1)
	for i := len(value.value); i > 0; i-- {
		h := value.value[i-1] & 0x0f
		if h < 0xa {
			longValue += int64(h) * base
			base *= 10
		} else {
			if h == 0xb || h == 0xd {
				sign = -1
			}
			base = 1
		}
		// System.out.print(h + " ");
		h = (value.value[i-1] & 0xf0) >> 4
		longValue += int64(h) * base
		base *= 10
	}
	longValue *= sign
	Central.Log.Debugf("packed to long conversion to %d", longValue)

	return longValue
}

func (value *packedValue) checkValidValue(intValue int64, len uint32) error {
	maxValue := uint64(1)
	for i := uint64(0); i < (uint64(len)*2)-1; i++ {
		maxValue *= 10
	}
	absValue := uint64(math.Abs(float64(intValue)))
	Central.Log.Debugf("Check valid value absolute value %d < max %d", absValue, maxValue)
	if absValue < maxValue {
		return nil
	}
	return NewGenericError(57, value.Type().Name(), intValue, len)
}

func (value *packedValue) createLength(v int64) uint32 {
	maxValue := int64(1)
	cipher := uint32(0)
	for maxValue < v {
		maxValue *= 10
		cipher++
	}
	cipher = (cipher + 1) / 2
	value.value = make([]byte, cipher)
	Central.Log.Debugf("Create size of %d for %d", cipher, v)
	return cipher
}

// LongToPacked convert long values (int64) to packed values
func (value *packedValue) LongToPacked(intValue int64, len uint32) {
	Central.Log.Debugf("Convert int64=%d of len=%d to packed", intValue, len)
	b := make([]byte, len)
	var v int64
	if intValue < 0 {
		v = -intValue
	} else {
		v = intValue
	}
	var x int64
	start := int64(len) - 2
	if start < -1 {
		Central.Log.Debugf("Start negative %d", start)
		return
	}
	if intValue > 0 {
		b[len-1] = positivePackedIndicator()
	} else {
		b[len-1] = negativePackedIndicator()
	}
	x = int64(v % 10)
	v = (v - x) / 10
	b[len-1] |= uint8(x << 4)
	Central.Log.Debugf("len=%d start=%d", len, start)
	for i := start; i >= 0; i-- {
		x = int64(v % 10)
		v = (v - x) / 10
		Central.Log.Debugf("index=%d start=%d", i, start)
		b[i] = uint8(x)
		x = int64(v % 10)
		v = (v - x) / 10
		b[i] |= uint8(x << 4)
	}
	Central.Log.Debugf("Final value converted %v 0x%X", b, b)

	value.value = b
}

func positivePackedIndicator() byte {
	/*	AdaPlatformInformation platformInformation =
			getType().getPlatformInformation();
		if (platformInformation == null) {
			if (AdabasDirectCall.isDefaultEbcdic()) {
				return 0xf;
			} else {
				return 0xc;
			}
		}
		if (platformInformation.isMainframe()) {
			return 0xf;
		} else { */
	return 0xc
	//}

}

func negativePackedIndicator() byte {
	/*	AdaPlatformInformation platformInformation =
			getType().getPlatformInformation();
		if (platformInformation == null) {
			if (AdabasDirectCall.isDefaultEbcdic()) {
				return 0xd;
			} else {
				return 0xb;
			}
		}
		if (platformInformation.isMainframe()) {
			return 0xd;
		} else {*/
	return 0xb
	//}

}
