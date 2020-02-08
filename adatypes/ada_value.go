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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

const defaultMultipleSize = 2048

// SecondCall second call enum
type SecondCall uint8

const (
	// NoneSecond No second call
	NoneSecond SecondCall = iota
	// ReadSecond Read call for second
	ReadSecond
	// StoreSecond Write call for second
	StoreSecond
)

// BufferOption option for buffer parsing
type BufferOption struct {
	MultifetchCall bool
	StoreCall      bool
	NeedSecondCall SecondCall
	ExchangeRecord bool
	PartialLobSize bool
	Ascending      bool
	Mainframe      bool
	SecondCall     uint8
	multipleSize   uint32
}

// NewBufferOption create option to parse the buffer
func NewBufferOption(store bool, secondCall uint8) *BufferOption {
	return NewBufferOption3(store, secondCall, false)
}

// NewBufferOption3 create option to parse the buffer
func NewBufferOption3(store bool, secondCall uint8, mainframe bool) *BufferOption {
	return &BufferOption{MultifetchCall: false, StoreCall: store,
		ExchangeRecord: false, SecondCall: secondCall, NeedSecondCall: NoneSecond,
		multipleSize: defaultMultipleSize, Ascending: true, Mainframe: mainframe}
}

// IAdaValue defines standard interface for all values
type IAdaValue interface {
	// ByteValue() byte
	parseBuffer(helper *BufferHelper, option *BufferOption) (TraverseResult, error)
	Type() IAdaType
	String() string
	Bytes() []byte
	PeriodIndex() uint32
	setPeriodIndex(peIndex uint32)
	setMultipleIndex(muIndex uint32)
	MultipleIndex() uint32
	FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32
	Value() interface{}
	SetParent(parentAdaValue IAdaValue)
	Parent() IAdaValue
	SetStringValue(string)
	SetValue(interface{}) error
	StoreBuffer(*BufferHelper, *BufferOption) error
	// Int32 convert current Adabas value into integer value if possible, if not fill error
	Int32() (int32, error)
	UInt32() (uint32, error)
	Int64() (int64, error)
	UInt64() (uint64, error)
	Float() (float64, error)
}

type adaValue struct {
	adatype IAdaType
	parent  IAdaValue
	peIndex uint32
	muIndex uint32
	//	searchLength uint32
}

func (adavalue adaValue) Type() IAdaType {
	return adavalue.adatype
}

func bigEndian() (ret bool) {
	i := 0x1
	bs := (*[4]byte)(unsafe.Pointer(&i))
	return bs[0] == 0
}

func endian() binary.ByteOrder {
	if bigEndian() {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// common format buffer generation
func (adavalue *adaValue) commonFormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	if option.SecondCall > 0 {
		Central.Log.Debugf("Work on %s -> second=%v\n", adavalue.Type().Name(), adavalue.Type().HasFlagSet(FlagOptionSecondCall))
		if adavalue.Type().HasFlagSet(FlagOptionSecondCall) {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			var index string
			switch {
			case adavalue.PeriodIndex() > 0 && adavalue.MultipleIndex() > 0:
				index = fmt.Sprintf("%d(%d)", adavalue.PeriodIndex(), adavalue.MultipleIndex())
			case adavalue.PeriodIndex() > 0:
				index = fmt.Sprintf("%d", adavalue.PeriodIndex())
			case adavalue.MultipleIndex() > 0:
				index = fmt.Sprintf("%d", adavalue.PeriodIndex())
			default:
				index = ""
			}
			buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", adavalue.Type().ShortName(),
				index, adavalue.Type().Length(), adavalue.Type().Type().FormatCharacter()))
			return adavalue.Type().Length()
		}
		return 0
	}
	if option.StoreCall {
		Central.Log.Debugf("Common store call for %s len=%d", adavalue.Type().Name(), adavalue.Type().Length())
		// debug.PrintStack()
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		fieldIndex := ""
		switch {
		case adavalue.PeriodIndex() > 0 && adavalue.MultipleIndex() > 0:
			fieldIndex = fmt.Sprintf("%d(%d)", adavalue.PeriodIndex(), adavalue.MultipleIndex())
		case adavalue.PeriodIndex() > 0:
			fieldIndex = fmt.Sprintf("%d", adavalue.PeriodIndex())
		case adavalue.MultipleIndex() > 0:
			fieldIndex = fmt.Sprintf("%d", adavalue.MultipleIndex())
		default:
		}
		buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", adavalue.Type().ShortName(),
			fieldIndex, adavalue.Type().Length(), adavalue.Type().Type().FormatCharacter()))
		return 0
	}
	Central.Log.Debugf("Common FormatBuffer for %s", adavalue.Type().Name())
	Central.Log.Debugf("PE flag set=%v Type is MU %v[%v] MU ghost=%v[%v]", adavalue.adatype.HasFlagSet(FlagOptionPE),
		(adavalue.adatype.Type() == FieldTypeMultiplefield), adavalue.adatype.Type(),
		adavalue.adatype.HasFlagSet(FlagOptionMUGhost), adavalue.adatype.HasFlagSet(FlagOptionAtomicFB))
	// Skip PE fields with MU fields
	if adavalue.adatype.HasFlagSet(FlagOptionPE) && adavalue.Type().HasFlagSet(FlagOptionMUGhost) {
		Central.Log.Debugf("Skip ... because PE and MU ghost")
		return 0
	}
	if adavalue.Type().HasFlagSet(FlagOptionMUGhost) && option.StoreCall {
		buffer.WriteString(fmt.Sprintf(",%s%d,%d,%s", adavalue.Type().ShortName(),
			adavalue.muIndex, adavalue.Type().Length(), adavalue.Type().Type().FormatCharacter()))
		return adavalue.Type().Length()
	}
	if adavalue.adatype.HasFlagSet(FlagOptionPE) {
		if !adavalue.Type().HasFlagSet(FlagOptionAtomicFB) && !adavalue.Type().HasFlagSet(FlagOptionPart) {
			Central.Log.Debugf("Skip ... because PE and not MU")
			return 0
		}
	}
	if buffer.Len() > 0 {
		buffer.WriteString(",")
	}
	buffer.WriteString(adavalue.Type().ShortName())
	Central.Log.Debugf("FormatBuffer generation: %s has flag PE %v", adavalue.Type().Name(), adavalue.Type().HasFlagSet(FlagOptionPE))
	Central.Log.Debugf("%s Type %p", adavalue.Type().Name(), adavalue.Type())
	both := false
	if adavalue.Type().HasFlagSet(FlagOptionPE) {
		buffer.WriteString("1-N")
		both = true
	}
	Central.Log.Debugf("%s has flag MU %v MU ghost %v period %v", adavalue.Type().Name(), adavalue.Type().HasFlagSet(FlagOptionAtomicFB),
		adavalue.Type().HasFlagSet(FlagOptionMUGhost), adavalue.Type().HasFlagSet(FlagOptionPE))
	if adavalue.adatype.Type() == FieldTypeMultiplefield {
		if both {
			buffer.WriteString("(")
		}
		buffer.WriteString("1-N")
		if both {
			buffer.WriteString(")")
		}
	}
	buffer.WriteString(",")
	buffer.WriteString(strconv.Itoa(int(adavalue.Type().Length())))
	buffer.WriteString(",")
	buffer.WriteString(adavalue.Type().Type().FormatCharacter())
	Central.Log.Debugf("Final element Formatbuffer: %s", buffer.String())
	return adavalue.Type().Length()
}

// common format buffer generation
func (adavalue *adaValue) commonUInt64Convert(x interface{}) (uint64, error) {
	var val uint64
	switch v := x.(type) {
	case string:
		sval, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		val = uint64(sval)
	case uint64:
		val = v
	case int64:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case int:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case uint32:
		val = uint64(x.(uint32))
	case int32:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case uint16:
		val = uint64(v)
	case int16:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case uint8:
		val = uint64(v)
	case int8:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case float64:
		if v < 0 {
			return 0, NewGenericError(101, fmt.Sprintf("%T", x))
		}
		val = uint64(v)
	case json.Number:
		i64, err := v.Int64()
		if err != nil {
			return 0, err
		}
		val = uint64(i64)
	case []byte:
		switch len(v) {
		case 1:
			buf := bytes.NewBuffer(v)
			var res uint8
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return uint64(res), nil
		case 2:
			buf := bytes.NewBuffer(v)
			var res uint16
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return uint64(res), nil
		case 4:
			buf := bytes.NewBuffer(v)
			var res uint32
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return uint64(res), nil
		case 8:
			buf := bytes.NewBuffer(v)
			var res uint64
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return res, nil
		default:
		}
		Central.Log.Debugf("Error converting to byte slice: %v", x)
		return 0, NewGenericError(104, len(v), adavalue.Type().Name())
	default:
		return 0, NewGenericError(103, fmt.Sprintf("%T", x), adavalue.Type().Name())
	}
	return val, nil
}

// common format buffer generation
func (adavalue *adaValue) commonInt64Convert(x interface{}) (int64, error) {
	Central.Log.Debugf("Convert common int64 value %s %v %s %T", adavalue.Type().Name(), x, reflect.TypeOf(x).Name(), x)
	var val int64
	multiplier := math.Pow10(int(adavalue.Type().Fractional()))
	switch v := x.(type) {
	case string:
		sval, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		val = int64(sval) * int64(multiplier)
	case int8:
		val = int64(v) * int64(multiplier)
	case int16:
		val = int64(v) * int64(multiplier)
	case int32:
		val = int64(v) * int64(multiplier)
	case int64:
		val = v * int64(multiplier)
	case uint8:
		val = int64(v) * int64(multiplier)
	case uint16:
		val = int64(v) * int64(multiplier)
	case uint32:
		val = int64(v) * int64(multiplier)
	case uint64:
		val = int64(v) * int64(multiplier)
	case int:
		val = int64(v) * int64(multiplier)
	case json.Number:
		i64, err := v.Int64()
		if err != nil {
			return 0, err
		}
		val = i64 * int64(multiplier)
	case []byte:
		switch len(v) {
		case 1:
			buf := bytes.NewBuffer(v)
			var res int8
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return int64(res), nil
		case 2:
			buf := bytes.NewBuffer(v)
			var res int16
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return int64(res), nil
		case 4:
			buf := bytes.NewBuffer(v)
			var res int32
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return int64(res), nil
		case 8:
			buf := bytes.NewBuffer(v)
			var res int64
			err := binary.Read(buf, endian(), &res)
			if err != nil {
				return 0, err
			}
			return res, nil
		default:
		}
		Central.Log.Debugf("Error converting to byte slice: %v", x)
		return 0, NewGenericError(104, len(v), adavalue.Type().Name())
	case float64:
		if adavalue.Type().Fractional() == 0 {
			if v != float64(int64(v)) {
				Central.Log.Debugf("Error converting %v", x)
				return 0, NewGenericError(103, fmt.Sprintf("%T", x), adavalue.Type().Name())
			}
		}
		val := int64(v * multiplier)
		return val, nil
	default:
		Central.Log.Debugf("Error converting %v", x)
		return 0, NewGenericError(103, fmt.Sprintf("%T", x), adavalue.Type().Name())
	}
	Central.Log.Debugf("Converted value %v from %T", val, x)
	return val, nil
}

type fillerValue struct {
	adaValue
}

func newFillerValue(initType IAdaType) *fillerValue {
	value := fillerValue{adaValue: adaValue{adatype: initType}}
	return &value
}

func (adavalue *adaValue) PeriodIndex() uint32 {
	return adavalue.peIndex
}

func (adavalue *adaValue) setPeriodIndex(index uint32) {
	adavalue.peIndex = index
}

func (adavalue adaValue) MultipleIndex() uint32 {
	return adavalue.muIndex
}

func (adavalue *adaValue) setMultipleIndex(index uint32) {
	adavalue.muIndex = index
}

func (adavalue *adaValue) SetParent(parentAdaValue IAdaValue) {
	adavalue.parent = parentAdaValue
}

func (adavalue *adaValue) Parent() IAdaValue {
	return adavalue.parent
}

func (value *fillerValue) ByteValue() byte {
	return ' '
}

func (value *fillerValue) String() string {
	return "FILLER"
}

func (value *fillerValue) Value() interface{} {
	return nil
}

// Bytes byte array representation of the value
func (value *fillerValue) Bytes() []byte {
	var empty []byte
	return empty
}

func (value *fillerValue) SetStringValue(stValue string) {
}

func (value *fillerValue) SetValue(v interface{}) error {
	return nil
}

func (value *fillerValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return 0
}

func (value *fillerValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	return nil
}

func (value *fillerValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	_, err = helper.ReceiveBytes(value.adatype.Length())
	if err != nil {
		res = EndTraverser
		return
	}
	res = Continue
	Central.Log.Debugf("Buffer get filler for offset=%v %s", helper.offset, value.Type().String())
	return
}

func (value *fillerValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

func (value *fillerValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}
func (value *fillerValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}
func (value *fillerValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}
func (value *fillerValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}
