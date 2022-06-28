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
	"fmt"
	"strings"
)

// unicodeValue unicode/wide character structure
type unicodeValue struct {
	adaValue
	value   []byte
	lobSize uint32
}

func newUnicodeValue(initType IAdaType) *unicodeValue {
	if initType == nil {
		return nil
	}
	value := adaValue{adatype: initType}
	var unicodeValue unicodeValue
	unicodeValue.adaValue = value
	if initType.Length() > 0 {
		unicodeValue.value = []byte(strings.Repeat(" ", int(initType.Length())))
	} else {
		unicodeValue.value = make([]byte, 0)
	}
	return &unicodeValue
}

func (value *unicodeValue) ByteValue() byte {
	return value.value[0]
}

func (value *unicodeValue) String() string {
	return string(value.value)
}

func (value *unicodeValue) Value() interface{} {
	return value.value
}

func (value *unicodeValue) setStringWithSize(sv string) {
	debug := Central.IsDebugLevel()

	if debug {
		Central.Log.Debugf("Set spaces for len %d and sv len %d", int(value.adatype.Length()), len(sv))
	}
	addSpaces := int(value.adatype.Length()) - len(sv)
	y := ""
	if debug {
		Central.Log.Debugf("Set value and add %d spaces", addSpaces)
	}
	if addSpaces < 0 {
		if value.adatype.Length() > 0 {
			value.value = []byte(sv[:value.adatype.Length()])
		} else {
			value.value = []byte(sv)
		}
	} else {
		y = strings.Repeat(" ", addSpaces)
		value.value = []byte(sv + y)
	}
	if debug {
		Central.Log.Debugf("Set value to >%s<", value.value)
	}
}

func (value *unicodeValue) Bytes() []byte {
	return value.value
}

func (value *unicodeValue) SetStringValue(stValue string) {
	value.setStringWithSize(stValue)
	Central.Log.Debugf("set string value to %s (%d)", stValue, len(value.value))
}

func (value *unicodeValue) SetValue(v interface{}) error {
	switch tv := v.(type) {
	case string:
		value.setStringWithSize(tv)
		Central.Log.Debugf("Set value to >%s<", value.value)
	case []byte:
		if value.Type().Length() == 0 {
			value.value = tv
		} else {
			value.value = []byte(strings.Repeat(" ", int(value.Type().Length())))
			//make([]byte, value.Type().Length())
			length := len(tv)
			if length > int(value.Type().Length()) {
				length = int(value.Type().Length())
			}
			copy(value.value, tv[:length])
		}
	default:
		return NewGenericError(103, fmt.Sprintf("%T", v), value.Type().Name())
	}
	return nil
}

// FormatBuffer generates format buffer part of this value
func (value *unicodeValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := uint32(0)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Generate FormatBuffer %s of length=%d and storeCall=%v",
			value.adatype.Type().name(), value.adatype.Length(), option.StoreCall)
	}
	if value.adatype.Type() == FieldTypeLBUnicode && value.adatype.Length() == 0 && !option.StoreCall {
		indexRange := getValueIndexRange(value)
		// If LOB field is read, use part
		if option.SecondCall > 0 {
			buffer.WriteString(fmt.Sprintf("%s%s(%d,%d)", value.Type().ShortName(), indexRange, PartialLobSize+1, value.lobSize))
			len = value.lobSize // - PartialLobSize
		} else {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s%sL,4,%s%s(0,%d)", value.Type().ShortName(), indexRange, value.Type().ShortName(), indexRange, PartialLobSize))
			len = 4 + PartialLobSize
		}
	} else {
		len = value.commonFormatBuffer(buffer, option, value.Type().Length())
		if len == 0 {
			switch value.adatype.Type() {
			case FieldTypeLAUnicode:
				len = 1114
			case FieldTypeLBUnicode:
				len = 16381
			default:
				len = 253
			}
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Record buffer size %d", len)
	}
	return len
}

func (value *unicodeValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	debug := Central.IsDebugLevel()
	if debug {
		Central.Log.Debugf("Store unicode %s at %d len=%d", value.Type().Name(), len(helper.buffer), value.Type().Length())
	}
	if value.Type().Length() == 0 {
		if debug {
			Central.Log.Debugf("Add length to buffer ...%d", len(value.value))
		}
		switch value.adatype.Type() {
		case FieldTypeLAUnicode:
			helper.PutUInt16(uint16(len(value.value)) + 2)
		case FieldTypeLBUnicode:
			helper.PutUInt32(uint32(len(value.value)) + 4)
		default:
			helper.PutUInt8(uint8(len(value.value)) + 1)
		}
	}
	if debug {
		Central.Log.Debugf("Current buffer size = %d", len(helper.buffer))
	}
	helper.putBytes([]byte(value.value))
	if debug {
		Central.Log.Debugf("All data buffer size = %d", len(helper.buffer))
		Central.Log.Debugf("Done store unicode %s at %d", value.Type().Name(), len(helper.buffer))
	}
	return nil
}

func (value *unicodeValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {

	debug := Central.IsDebugLevel()
	if option.SecondCall > 0 {
		if debug {
			Central.Log.Debugf("Old size of lob data %d of %d", len(value.value), value.lobSize)
		}
		if value.Type().Type() == FieldTypeLBUnicode && uint32(len(value.value)) < value.lobSize {
			data, rErr := helper.ReceiveBytes(value.lobSize - uint32(len(value.value)))
			if rErr != nil {
				err = rErr
				return EndTraverser, err
			}
			value.value = append(value.value, data...)
			if debug {
				LogMultiLineString(true, FormatByteBuffer("(2)LOB Buffer: ", value.value))
				Central.Log.Debugf("New size of lob data %d", len(value.value))
			}
		}
		if !value.Type().HasFlagSet(FlagOptionSecondCall) {
			Central.Log.Debugf("Skip parsing %s offset=%d, no second call flag", value.Type().Name(), helper.offset)
			return
		}
	}
	if debug {
		Central.Log.Debugf("Start parsing value unicode .... %s offset=%d/%X type=%s", value.Type().Name(),
			helper.offset, helper.offset, value.Type().Type().name())
	}

	fieldLength := value.adatype.Length()
	if fieldLength == 0 {
		if debug {
			Central.Log.Debugf("Field length dynamic")
		}

		switch value.adatype.Type() {
		case FieldTypeLAUnicode:
			length, errh := helper.ReceiveUInt16()
			if errh != nil {
				return EndTraverser, errh
			}
			fieldLength = uint32(length - 2)
			if debug {
				Central.Log.Debugf("Take field length 16 =%d", fieldLength)
			}
		case FieldTypeLBUnicode:
			value.lobSize, err = helper.ReceiveUInt32()
			if err != nil {
				return EndTraverser, err
			}
			if debug {
				Central.Log.Debugf("Got lobSize=%d", value.lobSize)
			}
			fieldLength = uint32(value.lobSize - 4)
			if debug {
				Central.Log.Debugf("Take partial buffer .... of size=%d offset=%d", PartialLobSize, helper.offset)
			}
		default:
			length, errh := helper.ReceiveUInt8()
			if errh != nil {
				return EndTraverser, errh
			}
			fieldLength = uint32(length - 1)
			if debug {
				Central.Log.Debugf("Take field length 8 =%d", fieldLength)
			}
		}
	}
	if debug {
		Central.Log.Debugf("%s length set to %d", value.Type().Name(), fieldLength)
	}

	value.value, err = helper.ReceiveBytes(fieldLength)
	if value.adatype.Type() == FieldTypeLBUnicode && option.PartialRead {
		switch {
		case value.lobSize < PartialLobSize:
			if debug {
				Central.Log.Debugf("Due to Unicode LOB lobSize is smaller then partial size for %s", value.Type().Name())
			}
			value.value = value.value[:value.lobSize]
		case value.lobSize > PartialLobSize:
			if debug {
				Central.Log.Debugf("Due to Unicode LOB lobSize is bigger then partial size, need second call (lob) for %s", value.Type().Name())
			}
			if option.NeedSecondCall = ReadSecond; option.StoreCall {
				option.NeedSecondCall = StoreSecond
			}
			if debug {
				Central.Log.Debugf("Unicode LOB: need second call %d", option.NeedSecondCall)
			}
		default:
		}
		if debug {
			Central.Log.Debugf("Buffer get lob string offset=%d %s size=%d/%d", helper.offset, value.Type().Name(), len(value.value), value.lobSize)
			LogMultiLineString(true, FormatByteBuffer("LOB Buffer: ", value.value))
		}

	} else {
		if debug {
			Central.Log.Debugf("Buffer get string offset=%d %s:%s size=%d", helper.offset, value.Type().Name(), value.value, len(value.value))
		}
	}
	return
}

func (value *unicodeValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}

func (value *unicodeValue) UInt8() (uint8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}
func (value *unicodeValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}

func (value *unicodeValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
}
func (value *unicodeValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

func (value *unicodeValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}
func (value *unicodeValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}
func (value *unicodeValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "sunsigned 64-bit integer")
}
func (value *unicodeValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}
