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

/* Template only: Collation descriptors are not supported at the moment */

package adatypes

import (
	"bytes"
	"fmt"
)

type referentialValue struct {
	adaValue
}

func newReferentialValue(initType IAdaType) *referentialValue {
	value := referentialValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (value *referentialValue) ByteValue() byte {
	return ' '
}

func (value *referentialValue) String() string {
	return fmt.Sprintf("%s=REFINT(%s,%d,%s)", value.Type().Name(), value.Type().Name(), 1, value.Type().Name())
}

func (value *referentialValue) Value() interface{} {
	return ""
}

// Bytes byte array representation of the value
func (value *referentialValue) Bytes() []byte {
	var empty []byte
	return empty
}

func (value *referentialValue) SetStringValue(stValue string) {
}

func (value *referentialValue) SetValue(v interface{}) error {
	return NewGenericError(37)
}

func (value *referentialValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return 0
}

func (value *referentialValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	return NewGenericError(37)
}

func (value *referentialValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	Central.Log.Debugf("Skip Buffer get collation descriptor %p value for %d", value, helper.offset)
	return
}

func (value *referentialValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}
func (value *referentialValue) UInt8() (uint8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}
func (value *referentialValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}
func (value *referentialValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
}
func (value *referentialValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}
func (value *referentialValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}
func (value *referentialValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}
func (value *referentialValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}
func (value *referentialValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}
