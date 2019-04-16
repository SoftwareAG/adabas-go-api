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
	"strconv"
)

type byteValue struct {
	adaValue
	value int8
}

func newByteValue(initType IAdaType) *byteValue {
	value := byteValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *byteValue) ByteValue() byte {
	return byte(value.value)
}

func (value *byteValue) String() string {
	return fmt.Sprintf("%d", value.value)
}

func (value *byteValue) Value() interface{} {
	return value.value
}

func (value *byteValue) Bytes() []byte {
	return []byte{byte(value.value)}
}

func (value *byteValue) SetStringValue(stValue string) {
	iv, err := strconv.Atoi(stValue)
	if err == nil {
		value.value = int8(iv)
	}
}

func (value *byteValue) SetValue(v interface{}) error {
	switch v.(type) {
	case byte, int8:
		value.value = v.(int8)
		return nil
	case int, int32:
		val := v.(int)
		value.value = int8(val)
		return nil
	case int64:
		val := v.(int64)
		value.value = int8(val)
		return nil
	case string:
		sv := v.(string)
		ba, err := strconv.Atoi(sv)
		if err != nil {
			return err
		}
		value.value = int8(ba)
		return nil
	case []byte:
		value.value = convertByteToInt8(v.([]byte)[0])
		return nil
	}
	return NewGenericError(103, fmt.Sprintf("%T", v), value.Type().Name())
}

func (value *byteValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option)
}

func (value *byteValue) StoreBuffer(helper *BufferHelper) error {
	return helper.putByte(byte(value.value))
}

func (value *byteValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	value.value, err = helper.ReceiveInt8()
	Central.Log.Debugf("Buffer get byte offset=%d %s", helper.offset, value.Type().Name())
	return
}

func (value *byteValue) Int32() (int32, error) {
	return int32(value.value), nil
}

func (value *byteValue) UInt32() (uint32, error) {
	return uint32(value.value), nil
}
func (value *byteValue) Int64() (int64, error) {
	return int64(value.value), nil
}
func (value *byteValue) UInt64() (uint64, error) {
	return uint64(value.value), nil
}
func (value *byteValue) Float() (float64, error) {
	return float64(value.value), nil
}

type ubyteValue struct {
	adaValue
	value uint8
}

func newUByteValue(initType IAdaType) *ubyteValue {
	value := ubyteValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *ubyteValue) ByteValue() byte {
	return value.value
}

func (value *ubyteValue) String() string {
	if value.Type().FormatType() == 'L' {
		if value.value == 0 {
			return "false"
		}
		return "true"
	}
	return fmt.Sprintf("%d", value.value)
}

func (value *ubyteValue) Value() interface{} {
	return value.value
}

func (value *ubyteValue) Bytes() []byte {
	return []byte{value.value}
}

func (value *ubyteValue) SetStringValue(stValue string) {
	iv, err := strconv.Atoi(stValue)
	if err == nil {
		value.value = uint8(iv)
	}
}

func (value *ubyteValue) SetValue(v interface{}) error {
	if value.Type().Type() == FieldTypeCharacter {
		switch v.(type) {
		case string:
			s := v.(string)
			sb := []byte(s)
			if len(sb) > 1 {
				return NewGenericError(108)
			}
			value.value = sb[0]
			return nil
		default:
		}
	}
	val, err := value.commonUInt64Convert(v)
	if err != nil {
		return err
	}
	value.value = uint8(val)
	return nil
}

func (value *ubyteValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option)
}

func (value *ubyteValue) StoreBuffer(helper *BufferHelper) error {
	return helper.putByte(value.value)
}

func (value *ubyteValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	value.value, err = helper.ReceiveUInt8()
	Central.Log.Debugf("Buffer get ubyte offset=%d %s", helper.offset, value.Type().Name())
	return
}

func (value *ubyteValue) Int32() (int32, error) {
	return int32(value.value), nil
}

func (value *ubyteValue) UInt32() (uint32, error) {
	return uint32(value.value), nil
}

func (value *ubyteValue) Int64() (int64, error) {
	return int64(value.value), nil
}

func (value *ubyteValue) UInt64() (uint64, error) {
	return uint64(value.value), nil
}

func (value *ubyteValue) Float() (float64, error) {
	return float64(value.value), nil
}
