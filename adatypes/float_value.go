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
	"fmt"
	"math"
	"strconv"
)

type floatValue struct {
	adaValue
	value []byte
}

func newFloatValue(initType IAdaType) *floatValue {
	value := floatValue{adaValue: adaValue{adatype: initType}}
	value.value = make([]byte, initType.Length())
	return &value
}

func float32ToByte(f interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, endian(), f)
	if err != nil {
		Central.Log.Debugf("binary.Write failed: %v", err)
		return nil
	}
	return buf.Bytes()
}

func byteToFLoat32(b []byte) float32 {
	buf := bytes.NewBuffer(b)
	var f float32
	err := binary.Read(buf, endian(), &f)
	if err != nil {
		Central.Log.Debugf("binary.Read failed: %v", err)
		return 0
	}
	return f

}

func (value *floatValue) ByteValue() byte {
	return value.value[0]
}

func (value *floatValue) String() string {
	return fmt.Sprintf("%f", byteToFLoat32(value.value))
}

func (value *floatValue) Value() interface{} {
	return byteToFLoat32(value.value)
}

func (value *floatValue) Bytes() []byte {
	return value.value
}

func (value *floatValue) SetStringValue(stValue string) {
	f, err := strconv.ParseFloat(stValue, 32)
	if err == nil {
		value.value = float32ToByte(float32(f))
	} else {
		fmt.Println(err)
	}
}

func (value *floatValue) SetValue(v interface{}) error {
	switch tv := v.(type) {
	case float32:
		value.value = float32ToByte(tv)
	case float64:
		f := float32(tv)
		value.value = float32ToByte(f)
	case string:
		value.SetStringValue(tv)
	case []byte:
		if uint32(len(tv)) > value.Type().Length() {
			return NewGenericError(104, len(tv), value.Type().Name())
		}
		copy(value.value[:len(tv)], tv[:])
		// value.value = bv
	default:
		i, err := value.commonInt64Convert(v)
		if err != nil {
			return err
		}
		value.value = float32ToByte(float32(i))
		return nil
	}
	return nil
}

func (value *floatValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option, value.Type().Length())
}

func (value *floatValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	return helper.putBytes(value.value)
}

func (value *floatValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	value.value, err = helper.ReceiveBytes(value.Type().Length())
	return
}

func (value *floatValue) Int8() (int8, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int8(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}

func (value *floatValue) UInt8() (uint8, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint8(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}
func (value *floatValue) Int16() (int16, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int16(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}

func (value *floatValue) UInt16() (uint16, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint16(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
}

func (value *floatValue) Int32() (int32, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int32(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

func (value *floatValue) UInt32() (uint32, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint32(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}

func (value *floatValue) Int64() (int64, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int64(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}

func (value *floatValue) UInt64() (uint64, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint64(fl), nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}

func (value *floatValue) Float() (float64, error) {
	return float64(byteToFLoat32(value.value)), nil
}
