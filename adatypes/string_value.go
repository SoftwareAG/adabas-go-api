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
	"errors"
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// PartialLobSize partial lob read size of first read
const PartialLobSize = 4096

// stringValue string structure
type stringValue struct {
	adaValue
	value   []byte
	lobSize uint32
}

func newStringValue(initType IAdaType) *stringValue {
	value := adaValue{adatype: initType}
	var stringValue stringValue
	stringValue.adaValue = value
	return &stringValue
}

func (value *stringValue) ByteValue() byte {
	return value.value[0]
}

func (value *stringValue) String() string {
	return string(value.value)
}

func (value *stringValue) Value() interface{} {
	return value.value
}

func (value *stringValue) setStringWithSize(sv string) {
	addSpaces := int(value.adatype.Length()) - len(sv)
	y := ""
	Central.Log.Debugf("Set value and add %d spaces", addSpaces)
	if addSpaces < 0 {
		value.value = []byte(sv[:-addSpaces])
	} else {
		y = strings.Repeat(" ", addSpaces)
		value.value = []byte(sv + y)
	}
	Central.Log.Debugf("Set value to >%s<", value.value)
}

func (value *stringValue) Bytes() []byte {
	// if value.Type().Length() == 0 {
	// 	varString := make([]byte, len(value.value)+1)
	// 	varString[0] = uint8(len(value.value) + 1)
	// 	copy(varString[1:], value.value)
	// 	Central.Log.Debugf("Variable length buffer %d -> %d", len(varString), value.Type().Length())
	// 	Central.Log.Debugf("Variable byte array:", varString)
	// 	return varString
	// }
	Central.Log.Debugf("Work on value=%p, got value of %d\n", value, len(value.value))
	return value.value
}

func (value *stringValue) SetStringValue(stValue string) {
	value.setStringWithSize(stValue)
	Central.Log.Debugf("set string value to %s (%d)", stValue, len(value.value))
}

func (value *stringValue) SetValue(v interface{}) error {
	switch v.(type) {
	case string:
		sv := v.(string)
		value.setStringWithSize(sv)
		Central.Log.Debugf("Set value to >%s<", value.value)
	case []byte:
		if value.Type().Length() == 0 {
			value.value = v.([]byte)
			Central.Log.Debugf("Set dynamic content of $p with len=$d", value, len(value.value))
		} else {
			Central.Log.Debugf("Set static value=%p, at value of %d\n", value, value.Type().Length())

			val := v.([]byte)
			value.value = make([]byte, value.Type().Length())
			length := len(val)
			if length > int(value.Type().Length()) {
				length = int(value.Type().Length())
			}
			copy(value.value, val[:length])
		}
	case reflect.Value:
		vv := v.(reflect.Value)
		value.setStringWithSize(vv.String())
	default:
		return fmt.Errorf("Input value %T not valid", v)
	}
	return nil
}

func (value *stringValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := uint32(0)
	Central.Log.Debugf("Generate FormatBuffer %s of length=%d and storeCall=%v", value.adatype.Type().name(), value.adatype.Length(), option.StoreCall)
	if value.adatype.Type() == FieldTypeLBString && value.adatype.Length() == 0 && !option.StoreCall {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		// If LOB field is read, use part
		if option.SecondCall {
			buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), PartialLobSize+1, value.lobSize))
			len = value.lobSize // - PartialLobSize
		} else {
			buffer.WriteString(fmt.Sprintf("%sL,4,%s(0,%d)", value.Type().ShortName(), value.Type().ShortName(), PartialLobSize))
			len = 4 + PartialLobSize
		}
	} else {
		len := value.commonFormatBuffer(buffer, option)
		if len == 0 {
			switch value.adatype.Type() {
			case FieldTypeLAString:
				len = 1114
			case FieldTypeLBString:
				len = 16381
			default:
				len = 253
			}
		}
	}
	return len
}

func (value *stringValue) StoreBuffer(helper *BufferHelper) error {
	Central.Log.Debugf("Store string %s at %d len=%d", value.Type().Name(), len(helper.buffer), value.Type().Length())
	Central.Log.Debugf("Current buffer size = %d", len(helper.buffer))
	stringLen := len(value.value)
	wrBytes := []byte(value.value)
	if stringLen == 0 {
		stringLen = 1
		wrBytes = []byte{' '}
	}
	if value.Type().Length() == 0 {
		Central.Log.Debugf("Add length to buffer ...%d", stringLen)
		switch value.adatype.Type() {
		case FieldTypeLAString:
			helper.PutUInt16(uint16(stringLen) + 2)
		case FieldTypeLBString:
			helper.PutUInt32(uint32(stringLen) + 4)
		default:
			helper.PutUInt8(uint8(stringLen) + 1)
		}
	}
	if uint32(len(wrBytes)) < value.Type().Length() {
		for i := uint32(len(wrBytes)); i < value.Type().Length(); i++ {
			wrBytes = append(wrBytes, ' ')
		}
	}
	helper.putBytes(wrBytes)
	Central.Log.Debugf("All data buffer size = %d written %d", len(helper.buffer), len(wrBytes))
	Central.Log.Debugf("Done store string %s at %d", value.Type().Name(), len(helper.buffer))
	return nil
}

func (value *stringValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {

	if option.SecondCall {
		Central.Log.Debugf("Old size of lob data %d of %d", len(value.value), value.lobSize)
		if value.Type().Type() == FieldTypeLBString && uint32(len(value.value)) < value.lobSize {
			data, rErr := helper.ReceiveBytes(value.lobSize - uint32(len(value.value)))
			if rErr != nil {
				err = rErr
				return EndTraverser, err
			}
			value.value = append(value.value, data...)
			if log.GetLevel() == log.DebugLevel {
				LogMultiLineString(FormatByteBuffer("(2)LOB Buffer: ", value.value))
			}

			Central.Log.Debugf("New size of lob data %d", len(value.value))
		}
		if !value.Type().HasFlagSet(FlagOptionSecondCall) {
			Central.Log.Debugf("Skip parsing %s offset=%d", value.Type().Name(), helper.offset)
			return
		}
	}

	fieldLength := value.adatype.Length()
	if fieldLength == 0 {
		switch value.adatype.Type() {
		case FieldTypeLAString:
			length, errh := helper.ReceiveUInt16()
			if errh != nil {
				return EndTraverser, errh
			}
			fieldLength = uint32(length - 2)
		case FieldTypeLBString:
			value.lobSize, err = helper.ReceiveUInt32()
			if err != nil {
				return EndTraverser, err
			}
			fieldLength = PartialLobSize // uint32(length - 4)
			Central.Log.Debugf("Take partial buffer .... of size=%d current lob size is %d", PartialLobSize, value.lobSize)
		default:
			length, errh := helper.ReceiveUInt8()
			if errh != nil {
				return EndTraverser, errh
			}
			fieldLength = uint32(length - 1)
		}
	}
	Central.Log.Debugf("%s length set to %d", value.Type().Name(), fieldLength)

	value.value, err = helper.ReceiveBytes(fieldLength)
	if value.adatype.Type() == FieldTypeLBString {
		switch {
		case value.lobSize < PartialLobSize:
			value.value = value.value[:value.lobSize]
		case value.lobSize > PartialLobSize:
			Central.Log.Debugf("Due to lobSize is bigger then partial size, need secand call (lob) for %s", value.Type().Name())
			option.NeedSecondCall = true
		default:
		}
		if log.GetLevel() == log.DebugLevel {
			Central.Log.Debugf("Buffer get lob string offset=%d %s size=%d/%d", helper.offset, value.Type().Name(), len(value.value), value.lobSize)
			LogMultiLineString(FormatByteBuffer("LOB Buffer: ", value.value))
		}

	} else {
		Central.Log.Debugf("Buffer get string offset=%d %s:%s size=%d", helper.offset, value.Type().Name(), value.value, len(value.value))
	}
	return
}

func (value *stringValue) Int32() (int32, error) {
	return 0, errors.New("Cannot convert value to signed 32-bit integer")
}

func (value *stringValue) UInt32() (uint32, error) {
	return 0, errors.New("Cannot convert value to unsigned 32-bit integer")
}
func (value *stringValue) Int64() (int64, error) {
	return 0, errors.New("Cannot convert value to signed 64-bit integer")
}
func (value *stringValue) UInt64() (uint64, error) {
	return 0, errors.New("Cannot convert value to unsigned 64-bit integer")
}
func (value *stringValue) Float() (float64, error) {
	return 0, errors.New("Cannot convert value to 64-bit float")
}
