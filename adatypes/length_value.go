/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"math"
	"strconv"
)

// length value
type lengthValue struct {
	adaValue
	value uint32
}

func newLengthValue(initType IAdaType) *lengthValue {
	value := lengthValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *lengthValue) ByteValue() uint64 {
	return uint64(value.value)
}

func (value *lengthValue) String() string {
	return fmt.Sprintf("%d", value.value)
}

func (value *lengthValue) Value() interface{} {
	return value.value
}

func (value *lengthValue) Bytes() []byte {
	v := make([]byte, 4)
	value.adatype.Endian().PutUint32(v, value.value)
	return v
}

func (value *lengthValue) SetStringValue(stValue string) {
	iv, err := strconv.ParseUint(stValue, 10, 32)
	if err == nil {
		value.value = uint32(iv)
	}
}

func (value *lengthValue) SetValue(v interface{}) error {
	val, err := value.commonUInt64Convert(v)
	if err != nil {
		return err
	}
	if val > math.MaxUint32 {
		return NewGenericError(117, val)
	}
	value.value = uint32(val)
	return nil
}

func (value *lengthValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	if option.SecondCall > 0 {
		return 0
	}
	if buffer.Len() > 0 {
		buffer.WriteRune(',')
	}
	fn := value.Type().Name()
	if fn[0] == '#' {
		fn = fn[1:]
	}
	if value.Type().HasFlagSet(FlagOptionLengthPE) {
		buffer.WriteString(fn + "C,4,B")
	} else {
		buffer.WriteString(fn + "L,4,B")
	}
	return 4
}

func (value *lengthValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	return nil
}

func (value *lengthValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall > 0 {
		return Continue, nil
	}
	Central.Log.Debugf("Parse length value")
	value.value, err = helper.ReceiveUInt32()
	if err != nil {
		return SkipTree, err
	}
	return Continue, nil
}

func (value *lengthValue) Int8() (int8, error) {
	return int8(value.value), nil
}

func (value *lengthValue) UInt8() (uint8, error) {
	return uint8(value.value), nil
}
func (value *lengthValue) Int16() (int16, error) {
	return int16(value.value), nil
}

func (value *lengthValue) UInt16() (uint16, error) {
	return uint16(value.value), nil
}
func (value *lengthValue) Int32() (int32, error) {
	return int32(value.value), nil
}

func (value *lengthValue) UInt32() (uint32, error) {
	return uint32(value.value), nil
}

func (value *lengthValue) Int64() (int64, error) {
	return int64(value.value), nil
}

func (value *lengthValue) UInt64() (uint64, error) {
	return uint64(value.value), nil
}

func (value *lengthValue) Float() (float64, error) {
	return float64(value.value), nil
}
