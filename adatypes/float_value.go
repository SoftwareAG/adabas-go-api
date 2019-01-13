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
	switch v.(type) {
	case float32:
		value.value = float32ToByte(v.(float32))
	case float64:
		f := float32(v.(float64))
		value.value = float32ToByte(f)
	case string:
		vs := v.(string)
		value.SetStringValue(vs)
	case []byte:
		bv := v.([]byte)
		if uint32(len(bv)) > value.Type().Length() {
			return errors.New("Cannot set byte array, length to small")
		}
		copy(value.value[:len(bv)], bv[:])
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
	return value.commonFormatBuffer(buffer, option)
}

func (value *floatValue) StoreBuffer(helper *BufferHelper) error {
	helper.putBytes(value.value)
	return nil
}

func (value *floatValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	value.value, err = helper.ReceiveBytes(value.Type().Length())
	return
}

func (value *floatValue) Int32() (int32, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int32(fl), nil
	}
	return 0, errors.New("Cannot convert value to signed 32-bit integer")
	//	return int32(byteToFLoat32(value.value)), nil
}

func (value *floatValue) UInt32() (uint32, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint32(fl), nil
	}
	return 0, errors.New("Cannot convert value to unsigned 32-bit integer")
	//	return uint32(byteToFLoat32(value.value)), nil
}

func (value *floatValue) Int64() (int64, error) {
	fl := byteToFLoat32(value.value)
	if fl == float32(math.Floor(float64(fl))) {
		return int64(fl), nil
	}
	return 0, errors.New("Cannot convert value to signed 64-bit integer")
	//	return int64(byteToFLoat32(value.value)), nil
}

func (value *floatValue) UInt64() (uint64, error) {
	fl := byteToFLoat32(value.value)
	if fl >= 0 && fl == float32(math.Floor(float64(fl))) {
		return uint64(fl), nil
	}
	return 0, errors.New("Cannot convert value to unsigned 64-bit integer")
	//	return uint64(byteToFLoat32(value.value)), nil
}

func (value *floatValue) Float() (float64, error) {
	return float64(byteToFLoat32(value.value)), nil
}
