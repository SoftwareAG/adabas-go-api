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
	"reflect"
	"strings"
)

// PartialLobSize partial lob read size of first read
const PartialLobSize = 4096

// PartialStoreLobSizeChunks chunk size storing lobs
const PartialStoreLobSizeChunks = 4096 * 10

// PartialValue partial value definition
type PartialValue interface {
	SetPartial(x, y uint32)
}

// stringValue string structure
type stringValue struct {
	adaValue
	value   []byte
	lobSize uint32
	partial []uint32
}

func newStringValue(initType IAdaType) *stringValue {
	if initType == nil {
		return nil
	}
	value := adaValue{adatype: initType}
	var stringValue stringValue
	stringValue.adaValue = value
	if initType.Length() > 0 {
		stringValue.value = []byte(strings.Repeat(" ", int(initType.Length())))
	} else {
		stringValue.value = make([]byte, 0)
	}
	return &stringValue
}

func (value *stringValue) ByteValue() byte {
	return value.value[0]
}

func (value *stringValue) String() string {
	b := bytes.Trim(value.value, "\x00")
	return string(b)
}

func (value *stringValue) Value() interface{} {
	return value.value
}

func (value *stringValue) setStringWithSize(sv string) {
	addSpaces := int(value.adatype.Length()) - len(sv)
	y := ""
	Central.Log.Debugf("Set value and add %d spaces", addSpaces)
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
		if value.Type().Length() > 0 && uint32(len(sv)) > value.Type().Length() {
			return NewGenericError(77, len(sv), value.Type().Length())
		}
		value.setStringWithSize(sv)
		Central.Log.Debugf("Set value to >%s<", value.value)
	case []byte:
		if value.Type().Length() == 0 {
			value.value = v.([]byte)
			Central.Log.Debugf("Set dynamic content with len=%d", len(value.value))
		} else {
			Central.Log.Debugf("Set static at value of len=%d\n", value.Type().Length())

			val := v.([]byte)
			value.value = []byte(strings.Repeat(" ", int(value.Type().Length())))
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
		return NewGenericError(103, fmt.Sprintf("%T", v), value.Type().Name())
	}
	return nil
}

// FormatBuffer generate format buffer for the string value
func (value *stringValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	recLength := uint32(0)
	Central.Log.Debugf("Generate FormatBuffer %s of length=%d/%d and storeCall=%v", value.adatype.Type().name(), value.adatype.Length(), len(value.value), option.StoreCall)
	// If store is request and lobsize is bigger then chunk size, do partial lob store calls
	if option.StoreCall && len(value.value) > PartialStoreLobSizeChunks {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		Central.Log.Debugf("Generate FormatBuffer second call %d", option.SecondCall)
		if option.SecondCall > 0 {
			start := uint32(option.SecondCall)*uint32(PartialStoreLobSizeChunks) + 1
			end := uint32(PartialStoreLobSizeChunks)
			recLength = PartialStoreLobSizeChunks
			if start+end > uint32(len(value.value)) {
				end = uint32(len(value.value)) - start + 1
				recLength = end
				option.NeedSecondCall = NoneSecond
			} else {
				option.NeedSecondCall = StoreSecond
			}
			buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), start, end))
		} else {
			partialRange := value.Type().PartialRange()
			Central.Log.Debugf("Partial Range %#v\n", partialRange)
			if partialRange != nil {
				buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), partialRange.from, partialRange.to))
				recLength = uint32(partialRange.to - partialRange.from)
			} else {
				buffer.WriteString(fmt.Sprintf("%s(1,%d)", value.Type().ShortName(), PartialStoreLobSizeChunks))
				recLength = 4 + PartialStoreLobSizeChunks
				option.NeedSecondCall = StoreSecond
			}
		}
		return recLength
	}
	if value.adatype.Type() == FieldTypeLBString && value.adatype.Length() == 0 && !option.StoreCall {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		// If LOB field is read, use part
		if option.SecondCall > 0 {
			if value.lobSize > PartialLobSize {
				buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), PartialLobSize+1, value.lobSize-PartialLobSize))
				recLength = value.lobSize - PartialLobSize
			}
		} else {
			partialRange := value.Type().PartialRange()
			Central.Log.Debugf("Partial Range %#v\n------\n", partialRange)
			if partialRange != nil {
				buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), partialRange.from, partialRange.to))
				recLength = uint32(partialRange.to - partialRange.from)
			} else {
				buffer.WriteString(fmt.Sprintf("%sL,4,%s(0,%d)", value.Type().ShortName(), value.Type().ShortName(), PartialLobSize))
				recLength = 4 + PartialLobSize
			}
		}
	} else {
		recLength = value.commonFormatBuffer(buffer, option)
		Central.Log.Debugf("String value format buffer length for %s -> %d", value.Type().ShortName(), recLength)
		if recLength == 0 {
			switch value.adatype.Type() {
			case FieldTypeLAString:
				recLength = 1114
			case FieldTypeLBString:
				recLength = 16381
			default:
				recLength = 253
			}
		}
	}
	return recLength
}

func (value *stringValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	if value.adatype.Type() == FieldTypeLBString && value.adatype.PartialRange() != nil {
		partialRange := value.adatype.PartialRange()
		// Write only part of partial lob, partial lob fragment need to be started from the beginning
		err := helper.putBytes(value.value[0:partialRange.to])
		if err != nil {
			return err
		}
		return nil
	}
	Central.Log.Debugf("Store string %s at %d len=%d", value.Type().Name(), len(helper.buffer), value.Type().Length())
	stringLen := len(value.value)
	Central.Log.Debugf("Current buffer size = %d/%d offset = %d", len(helper.buffer), stringLen, helper.offset)
	if stringLen > PartialStoreLobSizeChunks {
		start := uint32(option.SecondCall) * PartialStoreLobSizeChunks
		end := start + PartialStoreLobSizeChunks
		if end > uint32(len(value.value)) {
			end = uint32(len(value.value))
		}
		helper.putBytes(value.value[start:end])
		Central.Log.Debugf("Partial buffer added to size = %d (chunk=%d) offset = %d", len(helper.buffer), PartialStoreLobSizeChunks, helper.offset)
		return nil
	}
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
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

	if option.SecondCall > 0 {
		if value.Type().HasFlagSet(FlagOptionMUGhost) && value.Type().HasFlagSet(FlagOptionPE) {
			Central.Log.Debugf("MU flag evaluate length at offset %d", helper.Offset())
			value.lobSize, err = helper.ReceiveUInt32()
			if err != nil {
				return EndTraverser, err
			}
			value.lobSize -= 4
			Central.Log.Debugf("Byte to query: %d", value.lobSize)
			value.value, err = helper.ReceiveBytes(value.lobSize)
			if err != nil {
				return EndTraverser, err
			}
			return Continue, nil
		}
		if value.lobSize < PartialLobSize {
			Central.Log.Debugf("value lob size %d lower then partial lob size %d", value.lobSize, PartialLobSize)
			return Continue, nil
		}
		Central.Log.Debugf("Old size of lob data %d of %d offset=%d/%X", len(value.value), value.lobSize, helper.offset, helper.offset)
		if value.Type().Type() == FieldTypeLBString && uint32(len(value.value)) < value.lobSize {
			Central.Log.Debugf("Read bytes : %d", value.lobSize-uint32(len(value.value)))
			data, rErr := helper.ReceiveBytes(value.lobSize - uint32(len(value.value)))
			if rErr != nil {
				err = rErr
				return EndTraverser, err
			}
			value.value = append(value.value, data...)
			if Central.IsDebugLevel() {
				LogMultiLineString(FormatByteBuffer("Data: ", data))
				LogMultiLineString(FormatByteBuffer("(2)LOB Buffer: ", value.value))
				Central.Log.Debugf("New size of lob data %d offset=%d/%X", len(value.value), helper.Offset, helper.Offset)
			}
		}
		if !value.Type().HasFlagSet(FlagOptionSecondCall) {
			Central.Log.Debugf("Skip parsing %s offset=%d, no second call flag", value.Type().Name(), helper.offset)
			return Continue, nil
		}
	}

	if value.adatype.Type() == FieldTypeLBString && value.adatype.PartialRange() != nil {
		partialRange := value.adatype.PartialRange()
		// Read only the partial lob info
		value.value, err = helper.ReceiveBytes(uint32(partialRange.to))
		if err != nil {
			return EndTraverser, err
		}
		return Continue, nil
	}

	fieldLength := value.adatype.Length()
	switch fieldLength {
	case 0:
		if value.adatype.HasFlagSet(FlagOptionLengthNotIncluded) {
			length, errh := helper.ReceiveUInt8()
			if errh != nil {
				return EndTraverser, errh
			}
			fieldLength = uint32(length)
		} else {
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
				//value.lobSize = uint32(value.lobSize - 4)
				// if value.lobSize > PartialLobSize {
				fieldLength = PartialLobSize
				//} else {
				// fieldLength = value.lobSize
				//}
				Central.Log.Debugf("Take partial buffer .... of size=%d current lob size is %d", PartialLobSize, value.lobSize)
			default:
				length, errh := helper.ReceiveUInt8()
				if errh != nil {
					return EndTraverser, errh
				}
				fieldLength = uint32(length - 1)
			}
		}
	default:
	}
	Central.Log.Debugf("Alpha %s length set to %d", value.Type().Name(), fieldLength)

	value.value, err = helper.ReceiveBytes(fieldLength)
	if err != nil {
		return EndTraverser, err
	}
	if value.adatype.Type() == FieldTypeLBString {
		switch {
		case value.lobSize <= PartialLobSize:
			if len(value.value) < int(value.lobSize) {
				err = NewGenericError(56, len(value.value), value.lobSize)
				Central.Log.Debugf("Error parsing lob: %s", err.Error())
				return
			}
			Central.Log.Debugf("Use subset of lob partial: %d of %d", value.lobSize, len(value.value))
			if value.lobSize > 4 {
				value.value = value.value[:value.lobSize-4]
			} else {
				value.value = make([]byte, 0)
			}
		case value.lobSize > PartialLobSize:
			Central.Log.Debugf("Due to lobSize is bigger then partial size, need second call (lob) for %s", value.Type().Name())

			if option.NeedSecondCall = ReadSecond; option.StoreCall {
				option.NeedSecondCall = StoreSecond
			}
			Central.Log.Debugf("LOB: need second call %d", option.NeedSecondCall)
		default:
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Buffer get lob string offset=%d %s size=%d/%d", helper.offset, value.Type().Name(), len(value.value), value.lobSize)
			LogMultiLineString(FormatByteBuffer("LOB Buffer: ", value.value))
		}

	} else {
		Central.Log.Debugf("Buffer get string offset=%d %s:%s size=%d", helper.offset, value.Type().Name(), value.value, len(value.value))
	}
	Central.Log.Debugf("Rest of buffer %d", helper.Remaining())
	return
}

func (value *stringValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

func (value *stringValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}
func (value *stringValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}
func (value *stringValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}
func (value *stringValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}

func (value *stringValue) SetPartial(x, y uint32) {
	value.partial = []uint32{x, y}
}
