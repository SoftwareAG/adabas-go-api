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
	"regexp"
	"strconv"
)

const (
	ratadie = 1721426
	daysecs = 86400
)

// packedValue defines the packed value of Adabas packed format
type packedValue struct {
	adaValue
	value []byte
}

// newPackedValue new packed value reference generated
func newPackedValue(initType IAdaType) *packedValue {
	value := packedValue{adaValue: adaValue{adatype: initType}}
	vlen := initType.Length()
	value.value = make([]byte, vlen)
	if vlen > 0 {
		value.value[vlen-1] = positivePackedIndicator()
	}
	return &value
}

// ByteValue byte value
func (value *packedValue) ByteValue() byte {
	return value.value[0]
}

// String return the string representation
func (value *packedValue) String() string {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Generate packed string for %s, use format type %c", value.Type().Name(), value.Type().FormatType())
	}
	packedInt := value.packedToLong()
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Got int packed value %d", packedInt)
	}
	var sv string
	switch value.Type().FormatType() {
	case 'D':
		sv = natDateToString(packedInt)
	case 'T':
		sv = natTimeToString(packedInt)
	default:
		sv = strconv.FormatInt(packedInt, 10)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("In-between packed value %s fractional=%d", sv, value.Type().Fractional())
		}
		if value.Type().Fractional() > 0 {
			l := uint32(len(sv))
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Fractional packed value %s", sv)
			}
			if l <= value.Type().Fractional() {
				var buffer bytes.Buffer
				buffer.WriteString("0.")
				for i := l; i < value.Type().Fractional(); i++ {
					buffer.WriteRune('0')
				}
				buffer.WriteString(sv)
				sv = buffer.String()
			} else {
				sv = sv[:l-value.Type().Fractional()] + "." + sv[l-value.Type().Fractional():]
			}
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Return packed value %s", sv)
	}
	return sv
}

// natTimeToString Natural time transfered to string representation
func natTimeToString(nattime int64) string {
	microsecs := (nattime - 365*daysecs*10) * 100000
	microsec := microsecs % 1000000
	i := microsecs / 1000000
	second := i % 60
	i /= 60
	minute := i % 60
	i /= 60
	hour := i % 24
	i /= 24
	jdn := int(i + ratadie)
	j := jdn + 32044

	b := (4*j + 3) / 146097
	c := j - b*146097/4

	d := (4*c + 3) / 1461
	e := c - 1461*d/4
	m := (5*e + 2) / 153

	day := e - (153*m+2)/5 + 1
	month := m + 3 - 12*(m/10)
	year := b*100 + d - 4800 + m/10
	microsec /= 100000
	return fmt.Sprintf("%4d/%02d/%02d %02d:%02d:%02d.%d", year, month,
		day, hour, minute, second, microsec)
}

func natDateToString(jdn int64) string {
	j := jdn + 1721426 - 365 + 32044

	b := (4*j + 3) / 146097
	c := j - b*146097/4

	d := (4*c + 3) / 1461
	e := c - 1461*d/4
	m := (5*e + 2) / 153

	day := e - (153*m+2)/5 + 1
	month := m + 3 - 12*(m/10)
	year := b*100 + d - 4800 + m/10

	return fmt.Sprintf("%4d/%02d/%02d", year, month, day)
}

func (value *packedValue) Value() interface{} {
	return value.value
}

func (value *packedValue) Bytes() []byte {
	return value.value
}

// SetStringValue set the string value of the value
func (value *packedValue) SetStringValue(stValue string) {
	iv, err := strconv.ParseInt(stValue, 10, 64)
	if err == nil {
		value.LongToPacked(iv, value.adaValue.adatype.Length())
	}
}

// SetValue set the packed value by the given interface value
func (value *packedValue) SetValue(v interface{}) (err error) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set packed value to %v", v)
	}
	iLen := value.Type().Length()
	var v64 int64
	switch tv := v.(type) {
	case []byte:
		switch {
		case iLen != 0 && uint32(len(tv)) > iLen:
			return NewGenericError(59)
		case uint32(len(tv)) < iLen:
			copy(value.value, tv)
		default:
			value.value = tv
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Use byte array")
		}
		return nil
	case string:
		if Central.IsDebugLevel() {
			Central.Log.Debugf("String parse format type %v", value.Type().FormatType())
		}
		switch value.Type().FormatType() {
		case 'T':
			v64, err = parseDateTime(v)
			if err != nil {
				return err
			}
		case 'D':
			v64, err = parseDate(v)
			if err != nil {
				return err
			}
		default:
			v64, err = value.commonInt64Convert(v)
			if err != nil {
				return err
			}
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("String parse got %v", v64)
		}
	default:
		v64, err = value.commonInt64Convert(v)
		if err != nil {
			return err
		}
	}
	if iLen != 0 {
		err = value.checkValidValue(v64, value.Type().Length())
		if err != nil {
			return err
		}
	} else {
		iLen = value.createLength(v64)
	}

	value.LongToPacked(v64, iLen)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Packed value %s", value.String())
	}
	return nil
}

func parseDate(v interface{}) (int64, error) {

	var re = regexp.MustCompile(`(?m)(\d+)/(\d\d?)/(\d\d?)`)
	match := re.FindAllStringSubmatch(v.(string), -1)
	if len(match) == 0 || len(match[0]) < 4 {
		return 0, nil
	}
	m, merr := strconv.Atoi(match[0][2])
	if merr != nil {
		return 0, merr
	}
	y, yerr := strconv.Atoi(match[0][1])
	if yerr != nil {
		return 0, yerr
	}
	d, derr := strconv.Atoi(match[0][3])
	if derr != nil {
		return 0, derr
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("day=%d,month=%d,year=%d", d, m, y)
	}
	return convertDate2NatDate(d, m, y), nil
}

func convertDate2NatDate(d, m, y int) int64 {
	if m > 2 {
		m = m - 3
	} else {
		m = m + 9
		y = y - 1
	}

	c := int64(y / 100)
	ya := int64(y) - 100*c
	j := 146097*c/4 + 1461*ya/4 + (153*int64(m)+2)/5 + int64(d) + 1721119
	result := j - ratadie + 365
	if Central.IsDebugLevel() {
		Central.Log.Debugf("result->%d", result)
	}
	return result
}

func (value *packedValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := value.commonFormatBuffer(buffer, option, value.Type().Length())
	if len == 0 {
		len = 15
	}
	return len
}

func parseDateTime(v interface{}) (res int64, err error) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Parse data time %v", v)
	}
	var re = regexp.MustCompile(`(?m)(\d+)/(\d\d?)/(\d\d?) (\d\d?):(\d\d?):?(\d\d?)?\.?(\d*)?`)
	match := re.FindAllStringSubmatch(v.(string), -1)
	//fmt.Println(match)
	m, merr := strconv.Atoi(match[0][2])
	if merr != nil {
		Central.Log.Debugf("Error parsing minute")
		return 0, merr
	}
	y, yerr := strconv.Atoi(match[0][1])
	if yerr != nil {
		Central.Log.Debugf("Error parsing year")
		return 0, yerr
	}
	d, derr := strconv.Atoi(match[0][3])
	if derr != nil {
		Central.Log.Debugf("Error parsing day")
		return 0, derr
	}
	hour, derr := strconv.Atoi(match[0][4])
	if derr != nil {
		Central.Log.Debugf("Error parsing hour")
		return 0, derr
	}
	minute, derr := strconv.Atoi(match[0][5])
	if derr != nil {
		Central.Log.Debugf("Error parsing minute")
		return 0, derr
	}
	seconds := 0
	if match[0][6] != "" {
		seconds, err = strconv.Atoi(match[0][6])
		if err != nil {
			Central.Log.Debugf("Error parsing seconds")
			return 0, err
		}
	}
	microsec := 0
	if match[0][7] != "" {
		microsec, err = strconv.Atoi(match[0][7])
		if err != nil {
			Central.Log.Debugf("Error parsing microsec %v", match[0][7])
			return 0, err
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Evaluated %d.%d.%d %d:%d:%d.%d", d, m, y, hour, minute, seconds, microsec)
	}
	days := convertDate2NatDate(d, m, y)
	nattime := int64(microsec) + int64(math.Pow(10, 6))*(int64(seconds)+60*int64(minute+60*hour)+int64(days)*daysecs)
	res = (nattime / int64(math.Pow(10, 5))) // + 365 * DAYSECS * 10);
	return
}

func (value *packedValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	if value.Type().Length() == 0 {
		vlen := len(value.value)
		if vlen == 0 {
			return helper.putBytes([]byte{2, positivePackedIndicator()})
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Create variable len=%d", vlen)
		}
		err := helper.putByte(byte(vlen + 1))
		if err != nil {
			return err
		}
	}
	return helper.putBytes(value.value)
}

func (value *packedValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall > 0 && !value.Type().HasFlagSet(FlagOptionSecondCall) {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Skip parsing packed %s offset=%d, not needed at second call", value.Type().Name(), helper.offset)
		}
		return
	}
	if value.Type().Length() == 0 {
		length, errh := helper.ReceiveUInt8()
		if errh != nil {
			return EndTraverser, errh
		}
		if uint8(len(value.value)) != length-1 {
			value.value = make([]byte, length-1)
		}
	}

	value.value, err = helper.ReceiveBytes(uint32(len(value.value)))
	if err != nil {
		return
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Buffer get packed %s -> offset=%d/%X(%d)", value.adatype.Name(), helper.offset, helper.offset, len(value.value))
	}
	return
}

func (value *packedValue) Int8() (int8, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return int8(v), nil
}

func (value *packedValue) UInt8() (uint8, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint8(v), nil
}
func (value *packedValue) Int16() (int16, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return int16(v), nil
}

func (value *packedValue) UInt16() (uint16, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint16(v), nil
}
func (value *packedValue) Int32() (int32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return int32(v), nil
}

func (value *packedValue) UInt32() (uint32, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint32(v), nil
}
func (value *packedValue) Int64() (int64, error) {
	v := value.packedToLong()
	if value.Type().Fractional() > 0 {
		m := int64(fractional(value.Type().Fractional()))
		if v-(v%m) == v {
			return v / m, nil
		}
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	return int64(v), nil
}
func (value *packedValue) UInt64() (uint64, error) {
	if value.Type().Fractional() > 0 {
		return 0, NewGenericError(112, value.Type().Name(), value.Type().Fractional())
	}
	v := value.packedToLong()
	return uint64(v), nil
}

func fractional(f uint32) uint32 {
	x := uint32(1)
	for i := uint32(0); i < f; i++ {
		x *= 10
	}
	return x
}

func (value *packedValue) Float() (float64, error) {
	v := float64(value.packedToLong())
	if value.Type().Fractional() > 0 {
		v = v / float64(fractional(value.Type().Fractional()))
	}
	return v, nil
}

func (value *packedValue) packedToLong() int64 {
	if value == nil {
		return 0
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Packed %v", value.value)
	}

	base := int64(1)
	longValue := int64(0)
	sign := int64(1)
	for i := len(value.value); i > 0; i-- {
		h := value.value[i-1] & 0x0f
		if h < 0xa {
			longValue += int64(h) * base
			base *= 10
		} else {
			if h == 0xb || h == 0xd {
				sign = -1
			}
			base = 1
		}
		// System.out.print(h + " ");
		h = (value.value[i-1] & 0xf0) >> 4
		longValue += int64(h) * base
		base *= 10
	}
	longValue *= sign
	if Central.IsDebugLevel() {
		Central.Log.Debugf("packed to long conversion to %d", longValue)
	}

	return longValue
}

func (value *packedValue) checkValidValue(intValue int64, len uint32) error {
	maxValue := uint64(1)
	for i := uint64(0); i < (uint64(len)*2)-1; i++ {
		maxValue *= 10
	}
	absValue := uint64(math.Abs(float64(intValue)))
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Check valid value absolute value %d < max %d", absValue, maxValue)
	}
	if absValue < maxValue {
		return nil
	}
	return NewGenericError(57, value.Type().Name(), intValue, len)
}

func (value *packedValue) createLength(v int64) uint32 {
	maxValue := int64(1)
	cipher := uint32(0)
	for maxValue < v {
		maxValue *= 10
		cipher++
	}
	cipher = (cipher + 1) / 2
	value.value = make([]byte, cipher)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Create size of %d for %d", cipher, v)
	}
	return cipher
}

// LongToPacked convert long values (int64) to packed values
func (value *packedValue) LongToPacked(intValue int64, len uint32) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Convert int64=%d of len=%d to packed", intValue, len)
	}
	b := make([]byte, len)
	var v int64
	if intValue < 0 {
		v = -intValue
	} else {
		v = intValue
	}
	var x int64
	start := int64(len) - 2
	if start < -1 {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Start negative %d", start)
		}
		return
	}
	if intValue > 0 {
		b[len-1] = positivePackedIndicator()
	} else {
		b[len-1] = negativePackedIndicator()
	}
	x = int64(v % 10)
	v = (v - x) / 10
	b[len-1] |= uint8(x << 4)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("len=%d start=%d", len, start)
	}
	for i := start; i >= 0; i-- {
		x = int64(v % 10)
		v = (v - x) / 10
		if Central.IsDebugLevel() {
			Central.Log.Debugf("index=%d start=%d", i, start)
		}
		b[i] = uint8(x)
		x = int64(v % 10)
		v = (v - x) / 10
		b[i] |= uint8(x << 4)
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Final value converted %v 0x%X", b, b)
	}

	value.value = b
}

func positivePackedIndicator() byte {
	/*	AdaPlatformInformation platformInformation =
			getType().getPlatformInformation();
		if (platformInformation == null) {
			if (AdabasDirectCall.isDefaultEbcdic()) {
				return 0xf;
			} else {
				return 0xc;
			}
		}
		if (platformInformation.isMainframe()) {
			return 0xf;
		} else { */
	return 0xc
	//}

}

func negativePackedIndicator() byte {
	/*	AdaPlatformInformation platformInformation =
			getType().getPlatformInformation();
		if (platformInformation == null) {
			if (AdabasDirectCall.isDefaultEbcdic()) {
				return 0xd;
			} else {
				return 0xb;
			}
		}
		if (platformInformation.isMainframe()) {
			return 0xd;
		} else {*/
	return 0xb
	//}

}
