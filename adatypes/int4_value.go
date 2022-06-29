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

type uint32Value struct {
	adaValue
	value uint32
}

func newUInt4Value(initType IAdaType) *uint32Value {
	value := uint32Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *uint32Value) ByteValue() byte {
	return byte(value.value)
}

func (value *uint32Value) String() string {
	return strconv.Itoa(int(value.value))
}

func (value *uint32Value) Value() interface{} {
	return value.value
}

func (value *uint32Value) Bytes() []byte {
	v := make([]byte, 4)
	value.adatype.Endian().PutUint32(v, value.value)
	return v
}

func (value *uint32Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 0, 64)
	if err == nil {
		if iv < 0 || iv > math.MaxUint32 {
			return
		}
		value.value = uint32(iv)
	}
}

func (value *uint32Value) SetValue(v interface{}) error {
	x, err := value.commonUInt64Convert(v)
	if err != nil {
		return err
	}
	if x <= math.MaxUint32 {
		value.value = uint32(x)
		return nil
	}
	return NewGenericError(117, x)
}

func (value *uint32Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *uint32Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutUInt32(value.value)
}

func (value *uint32Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		rbLen, lerr := helper.ReceiveInt8()
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
			value.value = uint32(vba)
		case 2:
			vba, verr := helper.ReceiveUInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint32(vba)
		case 3:
			vba, verr := helper.ReceiveBytes(uint32(rbLen))
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = 0
			for i := range vba {
				ei := i
				if bigEndian() {
					ei = 3 - 1 - i
				}
				value.value = value.value + uint32(vba[ei])<<(uint32(i)*8)
			}
		case 4:
			vba, verr := helper.ReceiveUInt32()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint32(vba)
		default:
			return EndTraverser, NewGenericError(89)
		}
	} else {
		value.value, err = helper.ReceiveUInt32()
	}
	Central.Log.Debugf("Buffer get uint4 offset=%d %s %d", helper.offset, value.Type().Name(), value.value)
	return
}

func (value *uint32Value) Int8() (int8, error) {
	if value.value > uint32(math.MaxInt32) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int8(value.value), nil
}

func (value *uint32Value) UInt8() (uint8, error) {
	return uint8(value.value), nil
}
func (value *uint32Value) Int16() (int16, error) {
	if value.value > uint32(math.MaxInt32) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int16(value.value), nil
}

func (value *uint32Value) UInt16() (uint16, error) {
	return uint16(value.value), nil
}
func (value *uint32Value) Int32() (int32, error) {
	if value.value > uint32(math.MaxInt32) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
	}
	return int32(value.value), nil
}

func (value *uint32Value) UInt32() (uint32, error) {
	return value.value, nil
}
func (value *uint32Value) Int64() (int64, error) {
	return int64(value.value), nil
}
func (value *uint32Value) UInt64() (uint64, error) {
	return uint64(value.value), nil
}
func (value *uint32Value) Float() (float64, error) {
	return float64(value.value), nil
}

type int32Value struct {
	adaValue
	value int32
}

func newInt4Value(initType IAdaType) *int32Value {
	value := int32Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *int32Value) ByteValue() byte {
	return byte(value.value)
}

func (value *int32Value) String() string {
	return strconv.Itoa(int(value.value))
}

func (value *int32Value) Value() interface{} {
	return value.value
}

func (value *int32Value) Bytes() []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, endian(), value.value)
	if err != nil {
		return make([]byte, 0)
	}
	return buffer.Bytes()
}

func (value *int32Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 10, 32)
	if err == nil {
		value.value = int32(iv)
	}
}

func (value *int32Value) SetValue(v interface{}) error {
	x, err := value.commonInt64Convert(v)
	if err != nil {
		return err
	}
	value.value = int32(x)
	return nil
}

func (value *int32Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *int32Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutInt32(value.value)
}

func (value *int32Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		rbLen, lerr := helper.ReceiveInt8()
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
			value.value = int32(vba)
		case 2:
			vba, verr := helper.ReceiveInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int32(vba)
		case 4:
			vba, verr := helper.ReceiveInt32()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int32(vba)
		default:
			vba, verr := helper.ReceiveBytes(uint32(rbLen))
			if verr != nil {
				return EndTraverser, verr
			}
			v4 := make([]byte, 4)
			if bigEndian() {
				copy(v4[4-rbLen:], vba[:])
			} else {
				copy(v4[:rbLen], vba[:])
			}
			buf := bytes.NewBuffer(v4)
			verr = binary.Read(buf, helper.order, &value.value)
			if verr != nil {
				return EndTraverser, verr
			}
		}
	} else {
		value.value, err = helper.ReceiveInt32()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get int4 offset=%d %s", helper.offset, value.Type().Name())
	}
	return
}

func (value *int32Value) Int8() (int8, error) {
	return int8(value.value), nil
}

func (value *int32Value) UInt8() (uint8, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint8(value.value), nil
}
func (value *int32Value) Int16() (int16, error) {
	return int16(value.value), nil
}

func (value *int32Value) UInt16() (uint16, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint16(value.value), nil
}
func (value *int32Value) Int32() (int32, error) {
	return value.value, nil
}

func (value *int32Value) UInt32() (uint32, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint32(value.value), nil
}

func (value *int32Value) Int64() (int64, error) {
	return int64(value.value), nil
}

func (value *int32Value) UInt64() (uint64, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
	}
	return uint64(value.value), nil
}

func (value *int32Value) Float() (float64, error) {
	return float64(value.value), nil
}
