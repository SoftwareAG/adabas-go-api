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
	"math"
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

// AlphaConverter Alpha converter
// type AlphaConverter func(string, []byte) ([]byte, error)

// ConvertUnicode unicode converter interface
// - Decode to decode byte array from unicode
// - Encode to encode byte array to unicode
type ConvertUnicode interface {
	Decode([]byte) ([]byte, error)
	Encode([]byte) ([]byte, error)
}

// stringValue string structure
type stringValue struct {
	adaValue
	value          []byte
	lobSize        uint32
	PartialLobSize uint32
	PartialLobRead bool
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
	stringValue.PartialLobSize = PartialLobSize
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Init Partial LOB size set to %d", stringValue.PartialLobSize)
	}
	stringValue.PartialLobRead = false
	return &stringValue
}

func (value *stringValue) ByteValue() byte {
	return value.value[0]
}

func (value *stringValue) String() string {
	convertedvalue := value.Value().([]byte)
	b := bytes.Trim(convertedvalue, "\x00")
	return string(b)
}

func (value *stringValue) Value() interface{} {
	converter := value.Type().Convert()
	if converter != nil {
		convertedvalue, _ := converter.Decode(value.value)
		return convertedvalue
	}
	return value.value
}

func (value *stringValue) setStringWithSize(sv string) {
	addSpaces := int(value.adatype.Length()) - len(sv)
	y := ""
	if Central.IsDebugLevel() {
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
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set value to >%s<", value.value)
	}
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
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Work on value=%p, got value of %d\n", value, len(value.value))
	}
	return value.value
}

func (value *stringValue) SetStringValue(stValue string) {
	value.setStringWithSize(stValue)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("set string value to %s (%d)", stValue, len(value.value))
	}
}

func (value *stringValue) SetValue(v interface{}) error {
	switch tv := v.(type) {
	case string:
		if value.Type().Length() > 0 && uint32(len(tv)) > value.Type().Length() {
			return NewGenericError(77, len(tv), value.Type().Length())
		}
		value.setStringWithSize(tv)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Set value to >%s<", value.value)
		}
	case []byte:
		if value.Type().Length() == 0 {
			value.value = tv
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Set dynamic content with len=%d", len(value.value))
			}
		} else {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Set static at value of len=%d\n", value.Type().Length())
			}

			value.value = []byte(strings.Repeat(" ", int(value.Type().Length())))
			length := len(tv)
			if length > int(value.Type().Length()) {
				length = int(value.Type().Length())
			}
			copy(value.value, tv[:length])
		}
	case byte:
		if value.Type().Length() == 0 {
			value.value = []byte{tv}
		} else {
			value.value = []byte(strings.Repeat(" ", int(value.Type().Length())))
			value.value[0] = tv
		}
	case reflect.Value:
		value.setStringWithSize(tv.String())
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.String:
			sv := reflect.ValueOf(v).String()
			if value.Type().Length() > 0 && uint32(len(sv)) > value.Type().Length() {
				return NewGenericError(77, len(sv), value.Type().Length())
			}
			value.setStringWithSize(sv)
			Central.Log.Debugf("Set value to >%s<", value.value)
		default:
			return NewGenericError(103, fmt.Sprintf("%T", v), value.Type().Name())
		}
	}
	return nil
}

// FormatBuffer generate format buffer for the string value
func (value *stringValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	recLength := uint32(0)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Generate FormatBuffer name=%s (%s) of length=%d/%d and storeCall=%v",
			value.adatype.Name(), value.adatype.Type().name(), value.adatype.Length(), len(value.value), option.StoreCall)
	}
	// If store is request and lobsize is bigger then chunk size, do partial lob store calls
	if option.StoreCall && len(value.value) > PartialStoreLobSizeChunks {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Generate FormatBuffer second call %d", option.SecondCall)
		}
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
			if Central.IsDebugLevel() {
				Central.Log.Debugf("%d.Partial %s -> %d/%d of %d", option.SecondCall, value.Type().ShortName(), start, end, len(value.value))
			}
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), start, end))
		} else {
			partialRange := value.Type().PartialRange()
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Partial Range %#v\n", partialRange)
			}
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			if partialRange != nil {
				if partialRange.from == 0 {
					buffer.WriteString(fmt.Sprintf("%s(*,%d)", value.Type().ShortName(), partialRange.to))
					recLength = uint32(partialRange.to)
				} else {
					buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), partialRange.from, partialRange.to))
					recLength = uint32(partialRange.to)
				}
			} else {
				buffer.WriteString(fmt.Sprintf("%s(1,%d)", value.Type().ShortName(), PartialStoreLobSizeChunks))
				recLength = 4 + PartialStoreLobSizeChunks
				option.NeedSecondCall = StoreSecond
			}
		}
		return recLength
	}
	if value.adatype.Type() == FieldTypeLBString && value.adatype.Length() == 0 && !option.StoreCall {
		// If LOB field is read, use part
		if option.SecondCall > 0 {
			if value.lobSize > value.PartialLobSize {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), value.PartialLobSize+1, value.lobSize-value.PartialLobSize))
				recLength = value.lobSize - value.PartialLobSize
			}
		} else {
			partialRange := value.Type().PartialRange()
			if Central.IsDebugLevel() {
				Central.Log.Debugf("String value Partial Range %#v", partialRange)
			}
			indexRange := getValueIndexRange(value)
			// indexRange := ""
			// if value.Type().PeriodicRange().to != LastEntry {
			// 	indexRange = fmt.Sprintf("%d", value.Type().PeriodicRange().from)
			// }
			// if value.Type().MultipleRange().to != LastEntry {
			// 	if indexRange != "" {
			// 		indexRange += ","
			// 	}
			// 	indexRange += fmt.Sprintf("%d", value.Type().MultipleRange().from)
			// }
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			if partialRange != nil {
				if partialRange.to == 0 {
					buffer.WriteString(fmt.Sprintf("%s%s(*,%d)", value.Type().ShortName(), indexRange, partialRange.to))
					recLength = uint32(partialRange.to)
				} else {
					buffer.WriteString(fmt.Sprintf("%s%s(%d,%d)", value.Type().ShortName(), indexRange, partialRange.from, partialRange.to))
					recLength = uint32(partialRange.to)
				}
			} else {
				if !option.PartialRead {
					blockSize := option.BlockSize
					if blockSize == 0 {
						blockSize = PartialLobSize
					}
					buffer.WriteString(fmt.Sprintf("%sL%s,4,%s%s(1,%d)",
						value.Type().ShortName(), indexRange,
						value.Type().ShortName(), indexRange, blockSize))
				} else {
					buffer.WriteString(fmt.Sprintf("%sL%s,4,%s%s(*,%d)",
						value.Type().ShortName(), indexRange,
						value.Type().ShortName(), indexRange, value.PartialLobSize))
				}
				recLength = 4 + value.PartialLobSize
			}
		}
	} else {
		partial := value.Type().PartialRange()
		if partial != nil {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Generate partial format buffer %d,%d", partial.from, partial.to)
			}
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s(%d,%d)", value.Type().ShortName(), partial.from, partial.to))
			recLength = uint32(partial.to)
		} else {
			l := uint32(len(value.value))
			if l == 0 {
				l = 1
			}
			recLength = value.commonFormatBuffer(buffer, option, l)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("String value format buffer length for %s -> %d",
					value.Type().ShortName(), recLength)
			}
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
	}
	return recLength
}

func getValueIndexRange(value IAdaValue) string {
	indexRange := ""
	if Central.IsDebugLevel() {
		Central.Log.Debugf("PE range %d", value.PeriodIndex())
		Central.Log.Debugf("MU range %d", value.MultipleIndex())
	}
	if value.PeriodIndex() > 0 && value.PeriodIndex() < math.MaxUint32-2 {
		indexRange = fmt.Sprintf("%d", value.PeriodIndex())
	}
	if value.MultipleIndex() > 0 && value.MultipleIndex() < math.MaxUint32-2 {
		if indexRange != "" {
			indexRange += fmt.Sprintf("(%d)", value.MultipleIndex())
		} else {
			indexRange += fmt.Sprintf("%d", value.MultipleIndex())
		}
	}
	Central.Log.Debugf("Value index Range %s", indexRange)
	return indexRange
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
	partial := value.Type().PartialRange()
	if partial != nil {
		if partial.to != len(value.value) {
			return NewGenericError(135, len(value.value), partial.to)
		}
		err := helper.putBytes(value.value)
		if err != nil {
			return err
		}
		return nil
	}

	stringLen := len(value.value)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Store string %s at %d len=%d", value.Type().Name(), len(helper.buffer), value.Type().Length())
		Central.Log.Debugf("Current buffer size = %d/%d offset = %d/%x", len(helper.buffer), stringLen, helper.offset, helper.offset)
	}
	if stringLen > PartialStoreLobSizeChunks {
		start := uint32(option.SecondCall) * PartialStoreLobSizeChunks
		end := start + PartialStoreLobSizeChunks
		if end > uint32(len(value.value)) {
			end = uint32(len(value.value))
		}
		err := helper.putBytes(value.value[start:end])
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Partial buffer added to size = %d (chunk=%d) offset = %d (%v)",
				len(helper.buffer), PartialStoreLobSizeChunks, helper.offset, err)
		}
		return err
	}
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	wrBytes := []byte(value.value)
	if stringLen == 0 {
		//stringLen = 1
		wrBytes = []byte{' '}
	}
	// if value.Type().Length() == 0 {
	// 	Central.Log.Debugf("Add length to buffer ...%d", stringLen)
	// 	switch value.adatype.Type() {
	// 	case FieldTypeLAString:
	// 		helper.PutUInt16(uint16(stringLen) + 2)
	// 	case FieldTypeLBString:
	// 		helper.PutUInt32(uint32(stringLen) + 4)
	// 	default:
	// 		helper.PutUInt8(uint8(stringLen) + 1)
	// 	}
	// }
	if uint32(len(wrBytes)) < value.Type().Length() {
		for i := uint32(len(wrBytes)); i < value.Type().Length(); i++ {
			wrBytes = append(wrBytes, ' ')
		}
	}
	err := helper.putBytes(wrBytes)
	if err != nil {
		return err
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("All data buffer size = %d written %d", len(helper.buffer), len(wrBytes))
		Central.Log.Debugf("Done store string %s at %d", value.Type().Name(), len(helper.buffer))
	}
	return nil
}

func (value *stringValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	debug := Central.IsDebugLevel()
	if option.SecondCall > 0 {
		if debug {
			Central.Log.Debugf("Second call size of lob data %d of %d offset=%d/%X %v secondCall=%d",
				len(value.value), value.lobSize, helper.offset, helper.offset, option.PartialRead, option.SecondCall)
		}
		switch {
		case value.Type().HasFlagSet(FlagOptionMUGhost) && value.Type().HasFlagSet(FlagOptionPE):
			if debug {
				Central.Log.Debugf("MU flag evaluate length at offset %d", helper.Offset())
			}
			value.lobSize, err = helper.ReceiveUInt32()
			if err != nil {
				return EndTraverser, err
			}
			value.lobSize -= 4
			if debug {
				Central.Log.Debugf("Byte to query: %d", value.lobSize)
			}
			value.value, err = helper.ReceiveBytes(value.lobSize)
			if err != nil {
				return EndTraverser, err
			}
			return Continue, nil
		case value.lobSize < value.PartialLobSize:
			if debug {
				Central.Log.Debugf("value LOB size %d lower then partial lob size %d", value.lobSize, PartialLobSize)
			}
			return Continue, nil
		case option.PartialRead:
			if debug {
				Central.Log.Debugf("LOB size=%d offset=%d", value.lobSize, helper.offset)
			}
			value.value, err = helper.ReceiveBytes(value.PartialLobSize)
			if err != nil {
				return EndTraverser, err
			}
			return Continue, nil
		case value.Type().Type() == FieldTypeLBString && uint32(len(value.value)) < value.lobSize:
			if debug {
				Central.Log.Debugf("Read LOB bytes : %d", value.lobSize-uint32(len(value.value)))
			}
			data, rErr := helper.ReceiveBytes(value.lobSize - uint32(len(value.value)))
			if rErr != nil {
				err = rErr
				return EndTraverser, err
			}
			value.value = append(value.value, data...)
			if debug {
				LogMultiLineString(true, FormatByteBuffer("Data: ", data))
				LogMultiLineString(true, FormatByteBuffer("(2)LOB Buffer: ", value.value))
				Central.Log.Debugf("New size of lob data %d offset=%d/%X", len(value.value), helper.Offset, helper.Offset)
			}
			return Continue, nil
		case !value.Type().HasFlagSet(FlagOptionSecondCall):
			if debug {
				Central.Log.Debugf("Skip parsing %s offset=%d, no second call flag", value.Type().Name(), helper.offset)
			}
			return Continue, nil

		default:
		}
	}
	if debug {
		Central.Log.Debugf("Go into parse call size of lob data %d of %d offset=%d/%X %v secondCall=%d",
			len(value.value), value.lobSize, helper.offset, helper.offset, option.PartialRead, option.SecondCall)
	}
	if value.adatype.Type() == FieldTypeLBString && value.adatype.PartialRange() != nil {
		partialRange := value.adatype.PartialRange()
		if debug {
			Central.Log.Debugf("Read only to range %d bytes", partialRange.to)
		}
		// Read only the partial lob info
		value.value, err = helper.ReceiveBytes(uint32(partialRange.to))
		if err != nil {
			return EndTraverser, err
		}
		return Continue, nil
	}

	fieldLength := value.adatype.Length()
	if debug {
		Central.Log.Debugf("Normal parse size of lob data %d of %d offset=%d length=%d",
			len(value.value), value.lobSize, helper.offset, fieldLength)
	}
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
				fieldLength = value.PartialLobSize
				if debug {
					Central.Log.Debugf("Take partial buffer .... of size=%d current lob size is %d quantity=%d",
						value.PartialLobSize, value.lobSize, option.LowerLimit)
				}
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
	if debug {
		Central.Log.Debugf("Alpha %s length set to %d partial=%d", value.Type().Name(), fieldLength, value.PartialLobSize)
	}

	value.value, err = helper.ReceiveBytes(fieldLength)
	if err != nil {
		if debug {
			Central.Log.Debugf("Field length error: %v", err)
		}
		return EndTraverser, err
	}
	if value.adatype.Type() == FieldTypeLBString {
		switch {
		case value.lobSize <= value.PartialLobSize:
			if len(value.value) < int(value.lobSize) {
				err = NewGenericError(56, len(value.value), value.lobSize)
				Central.Log.Debugf("Error parsing lob: %s", err.Error())
				return
			}
			if debug {
				Central.Log.Debugf("Use subset of lob partial: %d of %d", value.lobSize, len(value.value))
			}
			//if value.lobSize > 4 {
			value.value = value.value[:value.lobSize]
			//} else {
			//	value.value = make([]byte, 0)
			//}
		case option.PartialRead:
			partSize := uint32(0)
			if option.LowerLimit < uint64(value.lobSize) {
				partSize = uint32(option.LowerLimit) % value.PartialLobSize
			}
			if option.LowerLimit > uint64(value.lobSize) {
				partSize = value.lobSize % value.PartialLobSize
			}
			if partSize == 0 {
				partSize = value.PartialLobSize
			}
			if debug {
				Central.Log.Debugf("Partial read %s of value size %d, reduce slice, partSize=%d",
					value.Type().Name(), len(value.value), partSize)
			}
			value.value = append([]byte(nil), value.value[:partSize]...)
		case value.lobSize > value.PartialLobSize:
			if debug {
				Central.Log.Debugf("Due to lobSize is bigger then partial size, need second call (lob) for %s", value.Type().Name())
			}

			if option.NeedSecondCall = ReadSecond; option.StoreCall {
				option.NeedSecondCall = StoreSecond
			}
			// value.Type().AddFlag(FlagOptionSecondCall)

			if debug {
				Central.Log.Debugf("LOB: need second call %d", option.NeedSecondCall)
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
	if debug {
		Central.Log.Debugf("Rest of buffer %d", helper.Remaining())
	}
	return
}

func (value *stringValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}
func (value *stringValue) UInt8() (uint8, error) {
	if value.adatype.Length() == 1 {
		return value.value[0], nil
	}
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}
func (value *stringValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}
func (value *stringValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
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
	value.Type().SetPartialRange(NewPartialRange(int(x), int(y)))
	// if value.Type().PartialRange() == nil {
	// 	panic(fmt.Sprintf("Partial range errror: %d,%d", x, y))
	// }
	//value.partial = []uint32{x, y}
}

func (value *stringValue) LobBlockSize() uint64 {
	return uint64(value.PartialLobSize)
}

func (value *stringValue) SetLobBlockSize(partialBlockSize uint64) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Partial LOB size set to %d", partialBlockSize)
	}
	value.PartialLobSize = uint32(partialBlockSize)
}

func (value *stringValue) SetLobPartRead(partialPartRead bool) {
	value.PartialLobRead = partialPartRead
}
