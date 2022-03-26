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
	"encoding/binary"
	"math"
	"strconv"
)

type uint64Value struct {
	adaValue
	value uint64
}

func newUInt8Value(initType IAdaType) *uint64Value {
	if initType == nil {
		return nil
	}
	value := uint64Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *uint64Value) ByteValue() byte {
	return byte(value.value)
}

func (value *uint64Value) String() string {
	return strconv.FormatUint(value.value, 10)
}

func (value *uint64Value) Value() interface{} {
	return value.value
}

func (value *uint64Value) Bytes() []byte {
	v := make([]byte, 8)
	value.adatype.Endian().PutUint64(v, value.value)
	return v
}

func (value *uint64Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseUint(stValue, 10, 64)
	if err == nil {
		value.value = iv
	}
}

func (value *uint64Value) SetValue(v interface{}) error {
	x, err := value.commonUInt64Convert(v)
	if err != nil {
		return err
	}
	Central.Log.Debugf("Set UInt8 value to >%d<", x)
	value.value = x
	return nil
}

func (value *uint64Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *uint64Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutUInt64(value.value)
}

func (value *uint64Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		rbLen, lerr := helper.ReceiveUInt8()
		if lerr != nil {
			return EndTraverser, lerr
		}
		rbLen--
		Central.Log.Debugf("Buffer get variable length=%d", rbLen)
		switch rbLen {
		case 1:
			vba, verr := helper.ReceiveUInt8()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint64(vba)
		case 2:
			vba, verr := helper.ReceiveUInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint64(vba)
		case 4:
			vba, verr := helper.ReceiveUInt32()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint64(vba)
		case 8:
			vba, verr := helper.ReceiveUInt64()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = vba
		default:
			vba, verr := helper.ReceiveBytes(uint32(rbLen))
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = 0
			for i := range vba {
				ei := i
				if bigEndian() {
					ei = int(rbLen) - 1 - i
				}
				value.value = value.value + uint64(uint32(vba[ei])<<(uint32(i)*8))
			}
		}
	} else {
		value.value, err = helper.ReceiveUInt64()
	}
	Central.Log.Debugf("Buffer get uint8 %d", helper.offset)
	return
}

func (value *uint64Value) Int8() (int8, error) {
	if value.value > uint64(math.MaxInt8) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
	}
	return int8(value.value), nil
}

func (value *uint64Value) UInt8() (uint8, error) {
	if value.value > uint64(math.MaxUint8) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint8(value.value), nil
}
func (value *uint64Value) Int16() (int16, error) {
	if value.value > uint64(math.MaxInt16) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int16(value.value), nil
}

func (value *uint64Value) UInt16() (uint16, error) {
	if value.value > uint64(math.MaxUint16) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return uint16(value.value), nil
}
func (value *uint64Value) Int32() (int32, error) {
	if value.value > uint64(math.MaxInt32) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int32(value.value), nil
}

func (value *uint64Value) UInt32() (uint32, error) {
	if value.value > uint64(math.MaxUint32) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint32(value.value), nil
}
func (value *uint64Value) Int64() (int64, error) {
	if value.value > uint64(math.MaxInt64) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
	}
	return int64(value.value), nil
}
func (value *uint64Value) UInt64() (uint64, error) {
	return value.value, nil
}
func (value *uint64Value) Float() (float64, error) {
	return float64(value.value), nil
}

type int64Value struct {
	adaValue
	value int64
}

func newInt8Value(initType IAdaType) *int64Value {
	if initType == nil {
		return nil
	}
	value := int64Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *int64Value) ByteValue() byte {
	return byte(value.value)
}

func (value *int64Value) String() string {
	return strconv.FormatInt(value.value, 10)
}

func (value *int64Value) Value() interface{} {
	return value.value
}

func (value *int64Value) Bytes() []byte {
	v := make([]byte, 8)
	binary.PutVarint(v, int64(value.value))
	return v
}

func (value *int64Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 10, 64)
	if err == nil {
		value.value = iv
	}
}

func (value *int64Value) SetValue(v interface{}) error {
	x, err := value.commonInt64Convert(v)
	if err != nil {
		return err
	}
	value.value = x
	return nil
}

func (value *int64Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *int64Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutInt64(value.value)
}

func (value *int64Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		rbLen, lerr := helper.ReceiveUInt8()
		if lerr != nil {
			return EndTraverser, lerr
		}
		rbLen--
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Buffer get variable length=%d", rbLen)
		}
		switch rbLen {
		case 1:
			vba, verr := helper.ReceiveInt8()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int64(vba)
		case 2:
			vba, verr := helper.ReceiveInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int64(vba)
		case 4:
			vba, verr := helper.ReceiveInt32()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int64(vba)
		case 8:
			vba, verr := helper.ReceiveInt64()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = vba
		default:
			vba, verr := helper.ReceiveBytes(uint32(rbLen))
			if verr != nil {
				return EndTraverser, verr
			}
			v8 := make([]byte, 8)
			if bigEndian() {
				copy(v8[8-rbLen:], vba[:])
			} else {
				copy(v8[:rbLen], vba[:])
			}
			buf := bytes.NewBuffer(v8)
			verr = binary.Read(buf, helper.order, &value.value)
			if verr != nil {
				return EndTraverser, verr
			}
		}
	} else {
		value.value, err = helper.ReceiveInt64()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get int8 %d", helper.offset)
	}
	return
}

func (value *int64Value) Int8() (int8, error) {
	if int64(math.MaxInt8) < value.value {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int8(value.value), nil
}

func (value *int64Value) UInt8() (uint8, error) {
	if value.value < 0 || value.value > int64(math.MaxUint32) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint8(value.value), nil
}
func (value *int64Value) Int16() (int16, error) {
	return int16(value.value), nil
}
func (value *int64Value) UInt16() (uint16, error) {
	if value.value < 0 || value.value > int64(math.MaxUint32) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return uint16(value.value), nil
}
func (value *int64Value) Int32() (int32, error) {
	if value.value < math.MinInt32 || value.value > math.MaxInt32 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return int32(value.value), nil
}
func (value *int64Value) UInt32() (uint32, error) {
	if value.value < 0 || value.value > int64(math.MaxUint32) {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint32(value.value), nil
}
func (value *int64Value) Int64() (int64, error) {
	return value.value, nil
}
func (value *int64Value) UInt64() (uint64, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
	}
	return uint64(value.value), nil
}
func (value *int64Value) Float() (float64, error) {
	return float64(value.value), nil
}
