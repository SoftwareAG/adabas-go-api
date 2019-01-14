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
	"errors"
	"fmt"
	"strconv"
)

type superDescValue struct {
	adaValue
	value []byte
}

func newSuperDescriptorValue(initType IAdaType) *superDescValue {
	value := superDescValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *superDescValue) ByteValue() byte {
	return ' '
}

func (value *superDescValue) String() string {
	return string(value.value)
}

func (value *superDescValue) Value() interface{} {
	return value.value
}

// Bytes byte array representation of the value
func (value *superDescValue) Bytes() []byte {
	return value.value
}

func (value *superDescValue) SetStringValue(stValue string) {
}

func (value *superDescValue) SetValue(v interface{}) error {
	return nil
}

func (value *superDescValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	if option.SecondCall {
		return 0
	}
	if value.adatype.IsOption(FieldOptionPE) || value.adatype.IsOption(FieldOptionPE) {
		return 0
	}
	Central.Log.Debugf("Write super descriptor FB part of %s", value.adatype.Name())
	if buffer.Len() > 0 {
		buffer.WriteString(",")
	}
	adaType := value.Type().(*AdaSuperType)
	buffer.WriteString(adaType.shortName)
	buffer.WriteString(",")
	buffer.WriteString(strconv.Itoa(int(value.Type().Length())))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprintf("%c", adaType.FdtFormat))
	Central.Log.Debugf("Got FB %s", buffer.String())
	return adaType.length
}

func (value *superDescValue) StoreBuffer(helper *BufferHelper) error {
	return nil
}

func (value *superDescValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall {
		return
	}
	if value.adatype.IsOption(FieldOptionPE) || value.adatype.IsOption(FieldOptionPE) {
		return
	}
	value.value, err = helper.ReceiveBytes(value.adatype.Length())
	Central.Log.Debugf("Buffer get super descriptor %p value for %d -> %s", value, helper.offset, string(value.value))
	return
}

func (value *superDescValue) Int32() (int32, error) {
	return 0, errors.New("Cannot convert value to signed 32-bit integer")
}

func (value *superDescValue) UInt32() (uint32, error) {
	return 0, errors.New("Cannot convert value to unsigned 32-bit integer")
}
func (value *superDescValue) Int64() (int64, error) {
	return 0, errors.New("Cannot convert value to signed 64-bit integer")
}
func (value *superDescValue) UInt64() (uint64, error) {
	return 0, errors.New("Cannot convert value to unsigned 64-bit integer")
}
func (value *superDescValue) Float() (float64, error) {
	return 0, errors.New("Cannot convert value to 64-bit float")
}
