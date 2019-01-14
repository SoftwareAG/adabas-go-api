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

type doubleValue struct {
	adaValue
	value []byte
}

func newDoubleValue(initType IAdaType) *doubleValue {
	value := doubleValue{adaValue: adaValue{adatype: initType}}
	value.value = make([]byte, 8)
	return &value
}

func float64ToByte(f float64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, endian(), f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func byteToFLoat64(b []byte) float64 {
	buf := bytes.NewBuffer(b)
	var f float64
	err := binary.Read(buf, endian(), &f)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return f

}

func (value *doubleValue) ByteValue() byte {
	return value.value[0]
}

func (value *doubleValue) String() string {
	return fmt.Sprintf("%f", byteToFLoat64(value.value))
}

func (value *doubleValue) Value() interface{} {
	return byteToFLoat64(value.value)
}

func (value *doubleValue) Bytes() []byte {
	return value.value
}

func (value *doubleValue) SetStringValue(stValue string) {
	f, err := strconv.ParseFloat(stValue, 64)
	if err == nil {
		value.value = float64ToByte(f)
	}
}

func (value *doubleValue) SetValue(v interface{}) error {
	switch v.(type) {
	case float32:
		f := float64(v.(float32))
		value.value = float64ToByte(f)
	case float64:
		value.value = float64ToByte(v.(float64))
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
		value.value = float64ToByte(float64(i))
		return nil
	}
	return nil
}

func (value *doubleValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.commonFormatBuffer(buffer, option)
}

func (value *doubleValue) StoreBuffer(helper *BufferHelper) error {
	helper.putBytes(value.value)
	return nil
}

func (value *doubleValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	value.value, err = helper.ReceiveBytes(value.Type().Length())
	return
}

func (value *doubleValue) Int32() (int32, error) {
	fl := byteToFLoat64(value.value)
	if fl != math.Trunc(fl) {
		return 0, errors.New("Cannot convert value to signed 32-bit integer")
	}
	return int32(fl), nil
}

func (value *doubleValue) UInt32() (uint32, error) {
	return uint32(byteToFLoat64(value.value)), nil
}
func (value *doubleValue) Int64() (int64, error) {
	return int64(byteToFLoat64(value.value)), nil
}
func (value *doubleValue) UInt64() (uint64, error) {
	return uint64(byteToFLoat64(value.value)), nil
}
func (value *doubleValue) Float() (float64, error) {
	return byteToFLoat64(value.value), nil
}
