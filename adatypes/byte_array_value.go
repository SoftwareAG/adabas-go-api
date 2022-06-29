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
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type byteArrayValue struct {
	adaValue
	value []byte
}

func newByteArrayValue(initType IAdaType) *byteArrayValue {
	value := byteArrayValue{adaValue: adaValue{adatype: initType}}
	value.value = make([]byte, initType.Length())
	return &value
}

func (value *byteArrayValue) ByteValue() byte {
	if len(value.value) > 0 {
		return byte(value.value[0])
	}
	return 0
}

func (value *byteArrayValue) String() string {
	var buffer bytes.Buffer
	for _, b := range value.value {
		if buffer.Len() > 0 {
			buffer.WriteRune(' ')
		} else {
			buffer.WriteRune('[')
		}
		buffer.WriteString(fmt.Sprintf("%d", b))
	}
	buffer.WriteRune(']')
	return buffer.String()
}

func (value *byteArrayValue) Value() interface{} {
	return value.value
}

func (value *byteArrayValue) Bytes() []byte {
	return value.value
}

func (value *byteArrayValue) SetStringValue(stValue string) {
	if strings.HasPrefix(stValue, "0x") {
		decoded, err := hex.DecodeString(stValue[2:])
		if err == nil {
			if value.Type().Length() == 0 {
				value.value = decoded
			} else {
				x := len(decoded)
				value.value = append(decoded[:], value.value[x:]...)
			}
		}
	} else {
		iv, err := strconv.ParseUint(stValue, 10, 64)
		if err == nil {
			if value.Type().Length() == 0 {
				value.value = make([]byte, 8)
			}
			switch {
			case len(value.value) >= 8:
				endian().PutUint64(value.value, iv)
			case len(value.value) >= 4:
				if iv > math.MaxUint32 {
					return
				}

				endian().PutUint32(value.value, uint32(iv))
			case len(value.value) >= 2:
				if iv > math.MaxUint16 {
					return
				}
				endian().PutUint16(value.value, uint16(iv))
			case len(value.value) >= 1:
				x, aerr := strconv.ParseInt(stValue, 0, 64)
				if aerr != nil {
					return
				}
				if x < math.MinInt8 || x > math.MaxInt8 {
					return
				}
				value.value[0] = byte(x)
			default:
			}
		}
	}
}

func (value *byteArrayValue) SetValue(v interface{}) error {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set value for %s using ... %T", value.Type().Name(), v)
	}
	switch tv := v.(type) {
	case []byte:
		if value.Type().Length() == 0 {
			value.value = tv
		} else {
			copy(value.value, tv)
		}
		return nil
	case string:
		value.value = []byte(tv)
		return nil
	}
	return NewGenericError(100, fmt.Sprintf("%T", v), value.Type().Name(), fmt.Sprintf("%T", value))
}

func (value *byteArrayValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := value.commonFormatBuffer(buffer, option, value.Type().Length())
	if len == 0 {
		len = 126
	}
	return len
}

func (value *byteArrayValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	if value.Type().Length() == 0 {
		if len(value.value) > 0 {
			Central.Log.Debugf("Add byte array ...")
			err := helper.putByte(byte(len(value.value) + 1))
			if err != nil {
				return err
			}
			return helper.putBytes(value.value)
		}
		Central.Log.Debugf("Add empty byte array")
		err := helper.putByte(2)
		if err != nil {
			return err
		}
		return helper.putByte(0)

	}
	Central.Log.Debugf("Fix byte array len ...")
	return helper.putBytes(value.value)
}

func (value *byteArrayValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	len := uint8(value.Type().Length())
	if len == 0 {
		len, err = helper.ReceiveUInt8()
		if err != nil {
			return
		}
		if len == 0 {
			return EndTraverser, NewGenericError(88)
		}
		len--
	}
	value.value, err = helper.ReceiveBytes(uint32(len))
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Byte array parse bytes offset=%X len=%d value=%#v", helper.offset, len, value.value)
	}
	return
}

func (value *byteArrayValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}

func (value *byteArrayValue) UInt8() (uint8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}
func (value *byteArrayValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}

func (value *byteArrayValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
}
func (value *byteArrayValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

func (value *byteArrayValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}
func (value *byteArrayValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}
func (value *byteArrayValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}
func (value *byteArrayValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}
