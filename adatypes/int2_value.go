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

type uint16Value struct {
	adaValue
	value uint16
}

func newUInt2Value(initType IAdaType) *uint16Value {
	value := uint16Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *uint16Value) ByteValue() byte {
	return byte(value.value)
}

func (value *uint16Value) String() string {
	return strconv.Itoa(int(value.value))
}

func (value *uint16Value) Value() interface{} {
	return value.value
}

func (value *uint16Value) Bytes() []byte {
	v := make([]byte, 2)
	value.adatype.Endian().PutUint16(v, value.value)
	return v
}

func (value *uint16Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 0, 64)
	if err == nil && iv >= 0 && iv <= math.MaxUint16 {
		value.value = uint16(iv)
	}
}

func (value *uint16Value) SetValue(v interface{}) error {
	val, err := value.commonUInt64Convert(v)
	if err != nil {
		return err
	}
	if val > math.MaxUint16 {
		return NewGenericError(117, val)
	}
	value.value = uint16(val)
	return nil
}

func (value *uint16Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *uint16Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutUInt16(value.value)
}

func (value *uint16Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		len, lerr := helper.ReceiveUInt8()
		if lerr != nil {
			return EndTraverser, lerr
		}
		len--
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Buffer get variable length=%d", len)
		}
		if len == 1 {
			vba, verr := helper.ReceiveUInt8()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint16(vba)
		} else {
			vba, verr := helper.ReceiveUInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = uint16(vba)
		}
	} else {
		value.value, err = helper.ReceiveUInt16()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get uint2 offset=%d %s", helper.offset, value.Type().String())
	}
	return
}

func (value *uint16Value) Int8() (int8, error) {
	if value.value > uint16(math.MaxInt8) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
	}
	return int8(value.value), nil
}

func (value *uint16Value) UInt8() (uint8, error) {
	if value.value > uint16(math.MaxUint8) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
	}
	return uint8(value.value), nil
}
func (value *uint16Value) Int16() (int16, error) {
	if value.value > uint16(math.MaxInt16) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
	}
	return int16(value.value), nil
}

func (value *uint16Value) UInt16() (uint16, error) {
	return uint16(value.value), nil
}
func (value *uint16Value) Int32() (int32, error) {
	if value.value > uint16(math.MaxInt16) {
		return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
	}
	return int32(value.value), nil
}

func (value *uint16Value) UInt32() (uint32, error) {
	return uint32(value.value), nil
}
func (value *uint16Value) Int64() (int64, error) {
	return int64(value.value), nil
}
func (value *uint16Value) UInt64() (uint64, error) {
	return uint64(value.value), nil
}
func (value *uint16Value) Float() (float64, error) {
	return float64(value.value), nil
}

type int16Value struct {
	adaValue
	value int16
}

func newInt2Value(initType IAdaType) *int16Value {
	value := int16Value{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *int16Value) ByteValue() byte {
	return byte(value.value)
}

func (value *int16Value) String() string {
	return strconv.Itoa(int(value.value))
}

func (value *int16Value) Value() interface{} {
	return value.value
}

func (value *int16Value) Bytes() []byte {
	v := make([]byte, 2)
	binary.PutVarint(v, int64(value.value))
	return v
}

func (value *int16Value) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 0, 64)
	if err == nil && iv >= math.MinInt16 && iv <= math.MaxInt16 {
		value.value = int16(iv)
	}
}

func (value *int16Value) SetValue(v interface{}) error {
	val, err := value.commonInt64Convert(v)
	if err != nil {
		return err
	}
	value.value = int16(val)
	return nil
}

func (value *int16Value) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *int16Value) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.PutInt16(value.value)
}

func (value *int16Value) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if value.Type().Length() == 0 {
		len, lerr := helper.ReceiveInt8()
		if lerr != nil {
			return EndTraverser, lerr
		}
		len--
		Central.Log.Debugf("Buffer get variable length=%d", len)
		if len == 1 {
			vba, verr := helper.ReceiveInt8()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int16(vba)
		} else {
			vba, verr := helper.ReceiveInt16()
			if verr != nil {
				return EndTraverser, verr
			}
			value.value = int16(vba)
		}
	} else {
		value.value, err = helper.ReceiveInt16()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get int2 offset=%d %s", helper.offset, value.Type().String())
	}
	return
}

func (value *int16Value) Int8() (int8, error) {
	return int8(value.value), nil
}

func (value *int16Value) UInt8() (uint8, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
	}
	return uint8(value.value), nil
}
func (value *int16Value) Int16() (int16, error) {
	return int16(value.value), nil
}

func (value *int16Value) UInt16() (uint16, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return uint16(value.value), nil
}
func (value *int16Value) Int32() (int32, error) {
	return int32(value.value), nil
}

func (value *int16Value) UInt32() (uint32, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return uint32(value.value), nil
}

func (value *int16Value) Int64() (int64, error) {
	return int64(value.value), nil
}
func (value *int16Value) UInt64() (uint64, error) {
	if value.value < 0 {
		return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
	}
	return uint64(value.value), nil
}
func (value *int16Value) Float() (float64, error) {
	return float64(value.value), nil
}
