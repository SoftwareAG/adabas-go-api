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

import "bytes"

// RedefinitionValue redefinition value struct
type RedefinitionValue struct {
	adaValue
	mainValue IAdaValue
	subValues []IAdaValue
}

func newRedefinition(redType IAdaType) *RedefinitionValue {
	redefinitionType := redType.(*RedefinitionType)

	r, err := redefinitionType.MainType.Value()
	if err != nil {
		return nil
	}
	subValues := make([]IAdaValue, 0)
	for _, s := range redefinitionType.SubTypes {
		sub, serr := s.Value()
		if serr != nil {
			return nil
		}
		subValues = append(subValues, sub)
	}
	return &RedefinitionValue{adaValue: adaValue{adatype: redefinitionType},
		mainValue: r, subValues: subValues}
}

// SetValue set value for structure
func (value *RedefinitionValue) SetValue(v interface{}) error {
	return nil
}

// SetStringValue set string value of redefinition
func (value *RedefinitionValue) SetStringValue(stValue string) {
}

// String string representation of redefinition
func (value *RedefinitionValue) String() string {
	return ""
}

// FormatBuffer provide the format buffer of this structure
func (value *RedefinitionValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	return value.mainValue.FormatBuffer(buffer, option)
}

// Value return the values of an structure value
func (value *RedefinitionValue) Value() interface{} {
	return nil
}

// Bytes byte array representation of the value
func (value *RedefinitionValue) Bytes() []byte {
	var empty []byte
	return empty
}

// StoreBuffer store buffer format generator
func (value *RedefinitionValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	// Skip normal fields in second call
	if option != nil && option.SecondCall > 0 {
		return nil
	}
	Central.Log.Debugf("Store buffer redefinition")
	subHelper := NewDynamicHelper(helper.order)
	for _, s := range value.subValues {
		err := s.StoreBuffer(subHelper, nil)
		if err != nil {
			Central.Log.Debugf("Error store buffer redefinition values: %s", s.Type().Name())
			return err
		}
	}
	return helper.putBytes(subHelper.Buffer())
}

// parseBuffer parse buffer
func (value *RedefinitionValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	_, err = value.mainValue.parseBuffer(helper, option)
	if err != nil {
		return EndTraverser, err
	}
	vb := value.mainValue.Bytes()
	subHelper := NewHelper(vb, len(vb), helper.order)
	for _, s := range value.subValues {
		_, err = s.parseBuffer(subHelper, option)
		if err != nil {
			Central.Log.Debugf("Error parsing redefinition values: %s", s.Type().Name())
			return Continue, err
		}
	}
	return SkipTree, err
}

// Int8 integer representation
func (value *RedefinitionValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}

// UInt8 unsigned integer representation
func (value *RedefinitionValue) UInt8() (uint8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}

// Int16 integer representation
func (value *RedefinitionValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}

// UInt16 unsigned integer representation
func (value *RedefinitionValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
}

// Int32 integer representation
func (value *RedefinitionValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

// UInt32 unsigned integer representation
func (value *RedefinitionValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}

// Int64 integer 64Bit representation
func (value *RedefinitionValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}

// UInt64 unsigned integer 64Bit represenation
func (value *RedefinitionValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}

// Float floating representation
func (value *RedefinitionValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}

// Traverse traverse redefinition fields
func (value *RedefinitionValue) Traverse(t TraverserValuesMethods, x interface{}) (ret TraverseResult, err error) {
	for e, v := range value.subValues {
		Central.Log.Debugf("Traverse node %d.element at %s[%d,%d] (%s) for %s[%d,%d] (%s)", e, v.Type().Name(),
			v.PeriodIndex(), v.MultipleIndex(), v.Type().Type().name(), value.Type().Name(), value.PeriodIndex(),
			value.MultipleIndex(), value.Type().Type().name())
		if value.PeriodIndex() != v.PeriodIndex() {
			if value.Type().Type() != FieldTypePeriodGroup {
				Central.Log.Debugf("!!!!----> Error index parent not correct for %s of %s", v.Type().Name(), value.Type().Name())
			}
		}
		if t.EnterFunction != nil {
			ret, err = t.EnterFunction(v, x)
			if err != nil || ret == EndTraverser {
				return
			}
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("%s-%s: Got structure return directive : %d", value.Type().Name(), v.Type().Name(),
				ret)
			LogMultiLineString(true, FormatByteBuffer("DATA: ", v.Bytes()))
		}
		if ret == SkipStructure {
			Central.Log.Debugf("Skip structure tree ... ")
			return Continue, nil
		}
		if v.Type().IsStructure() && ret != SkipTree {
			Central.Log.Debugf("Traverse tree %s", v.Type().Name())
			ret, err = v.(*StructureValue).Traverse(t, x)
			if err != nil || ret == EndTraverser {
				return
			}
		}
		if t.LeaveFunction != nil {
			ret, err = t.LeaveFunction(v, x)
			if err != nil || ret == EndTraverser {
				return
			}
		}
	}
	return Continue, nil
}
