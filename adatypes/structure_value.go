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
	"math"

	log "github.com/sirupsen/logrus"
)

type structureElement struct {
	Values   []IAdaValue
	peIndex  uint32
	valueMap map[string]IAdaValue
}

func newStructureElement() *structureElement {
	return &structureElement{valueMap: make(map[string]IAdaValue)}
}

// StructureValue structure value struct
type StructureValue struct {
	adaValue
	Elements   []*structureElement
	elementMap map[uint32]*structureElement
}

func newStructure(initType IAdaType) *StructureValue {
	Central.Log.Debugf("Create new structure value %s", initType.Name())
	value := StructureValue{adaValue: adaValue{adatype: initType}}
	value.elementMap = make(map[uint32]*structureElement)
	switch initType.Type() {
	case FieldTypeGroup:
	//	value.initSubValues(0, 0, false)
	default:
	}
	return &value
}

/*
 * Init sub structures with empty value fields
 */
func (value *StructureValue) initSubValues(index uint32, peIndex uint32, initMuFields bool) {
	subType := value.adatype.(*StructureType)
	Central.Log.Debugf("Init sub values for %s[%d,%d] -> %d,%d", value.adatype.Name(), value.PeriodIndex(),
		value.MultipleIndex(), peIndex, index)

	if value.Type().Type() != FieldTypeMultiplefield || initMuFields {
		for _, st := range subType.SubTypes {
			if st.HasFlagSet(FlagOptionMUGhost) && !initMuFields {
				continue
			}
			Central.Log.Debugf("Init sub structure %s(%s) for structure %s period index=%d", st.Name(), st.Type().name(),
				value.Type().Name(), peIndex)
			stv, err := st.Value()
			if err != nil {
				Central.Log.Debugf("Error %v", err)
				return
			}
			stv.setPeriodIndex(peIndex)
			Central.Log.Debugf("Add to %s[%d,%d] element %s[%d,%d] --> index=%d", value.Type().Name(), value.PeriodIndex(),
				value.MultipleIndex(), stv.Type().Name(),
				stv.PeriodIndex(), stv.MultipleIndex(), peIndex)
			value.addValue(stv, index)
			if stv.Type().IsStructure() {
				stv.(*StructureValue).initSubValues(index, peIndex, false)
			}
		}
		Central.Log.Debugf("Finished Init sub values for %s len=%d", value.Type().Name(), len(value.Elements))
	} else {
		Central.Log.Debugf("Skip Init sub values for %s", value.Type().Name())
	}
}

func (value *StructureValue) String() string {
	return ""
}

// PeriodIndex returns the period index of the structured value
func (value *StructureValue) PeriodIndex() uint32 {
	return value.peIndex
}

type evaluateFieldNameStructure struct {
	names    []string
	namesMap map[string]bool
}

func evaluateFieldNames(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	efns := x.(*evaluateFieldNameStructure)
	Central.Log.Debugf("Evaluate field %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		if adaValue.Type().Type() == FieldTypeMultiplefield {
			if _, ok := efns.namesMap[adaValue.Type().Name()]; !ok {
				Central.Log.Debugf("Add multiple field")
				efns.names = append(efns.names, adaValue.Type().Name())
				efns.namesMap[adaValue.Type().Name()] = true
			}
		}
	} else {
		if !adaValue.Type().HasFlagSet(FlagOptionMUGhost) {
			if _, ok := efns.namesMap[adaValue.Type().Name()]; !ok {
				Central.Log.Debugf("Add field")
				efns.names = append(efns.names, adaValue.Type().Name())
				efns.namesMap[adaValue.Type().Name()] = true
			}
		}
	}
	return Continue, nil
}

func countMU(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	helper := x.(*BufferHelper)
	if adaType.Type() == FieldTypeMultiplefield {
		Central.Log.Debugf("Skip MU counter %s", adaType.Name())
		_, err := helper.ReceiveUInt32()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 Parse buffer if a period group contains multiple fields. In that case the buffer parser need to parse
 field by field and not the group alltogether
*/
func (value *StructureValue) parseBufferWithMUPE(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	Central.Log.Debugf("Parse Buffer structure with (MUPE) name=%s offset=%d remaining=%d length=%d value length=%d", value.Type().Name(),
		helper.offset, helper.Remaining(), len(helper.buffer), len(value.Elements))

	adaType := value.Type().(*StructureType)
	if value.Type().Type() != FieldTypePeriodGroup {
		Central.Log.Debugf("Skip not group -> %s", value.Type().Name())
		return
	}
	Central.Log.Debugf("%s parse buffer for MU in PE/first call", value.Type().Name())
	var occNumber int
	occNumber, err = value.evaluateOccurence(helper)
	Central.Log.Debugf("%s has %d entries", value.Type().Name(), occNumber)
	if occNumber > 10000 {
		Central.Log.Debugf("Too many occurences")
		panic("Too many occurence entries")
	}
	for i := uint32(0); i < uint32(occNumber); i++ {
		value.initSubValues(i, i+1, true)
	}
	if occNumber == 0 {
		Central.Log.Debugf("Skip parsing, evaluate MU for empty counter")
		t := TraverserMethods{EnterFunction: countMU}
		adaType.Traverse(t, 1, helper)

	} else {
		/* Evaluate the fields which need ot be parsed in the period group */
		tm := TraverserValuesMethods{EnterFunction: evaluateFieldNames}
		efns := &evaluateFieldNameStructure{namesMap: make(map[string]bool)}
		res, err = value.Traverse(tm, efns)
		Central.Log.Debugf("Got %d names", len(efns.names))
		for _, n := range efns.names {
			Central.Log.Debugf("Found name : %s", n)
			for i := 0; i < occNumber; i++ {
				Central.Log.Debugf("Get occurence : %d", (i + 1))
				v := value.Get(n, i+1)
				v.setPeriodIndex(uint32(i + 1))
				if v.Type().IsStructure() {
					nrMu, nrMerr := helper.ReceiveUInt32()
					if nrMerr != nil {
						err = nrMerr
						return
					}
					Central.Log.Debugf("Got Nr of Multiple Fields = %d creating them ...", nrMu)
					/* Initialize MU elements dependent on the counter result */
					for muIndex := uint32(0); muIndex < nrMu; muIndex++ {
						muStructureType := v.Type().(*StructureType)
						Central.Log.Debugf("Create index MU %d", (muIndex + 1))
						sv, typeErr := muStructureType.SubTypes[0].Value()
						if typeErr != nil {
							err = typeErr
							return
						}
						muStructure := v.(*StructureValue)
						sv.Type().AddFlag(FlagOptionSecondCall)
						sv.setMultipleIndex(muIndex + 1)
						sv.setPeriodIndex(uint32(i + 1))
						muStructure.addValue(sv, muIndex)
						Central.Log.Debugf("MU index %d,%d -> %d", sv.PeriodIndex(), sv.MultipleIndex(), i)
						Central.Log.Debugf("Due to Period and MU field, need second call call (PE/MU) for %s", value.Type().Name())
						option.NeedSecondCall = true
					}
				} else {
					/* Parse field value for each non-structure field */
					res, err = v.parseBuffer(helper, option)
					if err != nil {
						return
					}
				}
			}
		}
	}

	res = SkipStructure
	return
}

// Parse the structure
func (value *StructureValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall {
		Central.Log.Debugf("Skip parsing %s offset=%X", value.Type().Name(), helper.offset)
		return
	}
	Central.Log.Debugf("Parse structure buffer %s secondCall=%v offset=%d/%X", value.Type().Name(), option.SecondCall, helper.offset, helper.offset)
	if value.adatype.HasFlagSet(FlagOptionPE) && value.adatype.HasFlagSet(FlagOptionMU) {
		return value.parseBufferWithMUPE(helper, option)
	}
	return value.parseBufferWithoutMUPE(helper, option)
}

// Evaluate the occurence of the structure
func (value *StructureValue) evaluateOccurence(helper *BufferHelper) (occNumber int, err error) {
	subStructure := value.adatype.(*StructureType)
	occNumber = math.MaxInt32
	if subStructure.occ > 0 {
		occNumber = subStructure.occ
	} else {
		switch subStructure.occ {
		case OccCapacity:
			res, subErr := helper.ReceiveUInt32()
			if subErr != nil {
				err = subErr
				return
			}
			occNumber = int(res)
		case OccSingle:
			occNumber = 1
		case OccByte:
			res, subErr := helper.ReceiveUInt8()
			if subErr != nil {
				err = subErr
				return
			}
			occNumber = int(res)
			break
		case OccNone:
			break
		}
	}
	Central.Log.Debugf("Evaluate occurrence for %s of type %d to %d", value.Type().Name(), subStructure.occ, occNumber)
	return
}

// Parse the buffer containing no PE and MU fields
func (value *StructureValue) parseBufferWithoutMUPE(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	Central.Log.Debugf("Parse Buffer structure without MUPE name=%s offset=%d remaining=%d length=%d value length=%d type=%d", value.Type().Name(),
		helper.offset, helper.Remaining(), len(helper.buffer), len(value.Elements), value.Type().Type())
	var occNumber int
	occNumber, err = value.evaluateOccurence(helper)
	if err != nil {
		return
	}
	Central.Log.Debugf("Occurence %d period index=%d", occNumber, value.peIndex)
	switch value.Type().Type() {
	case FieldTypePeriodGroup, FieldTypeMultiplefield:
		Central.Log.Debugf("Init values")
		for i := uint32(0); i < uint32(occNumber); i++ {
			value.initSubValues(i, i+1, true)
		}
		Central.Log.Debugf("Init values finished")
		return
	case FieldTypeStructure:
	default:
		Central.Log.Debugf("Unused type=%d", value.Type().Type())
		return
	}
	Central.Log.Debugf("Start going through elements=%d", value.NrElements())
	index := 0
	for index < occNumber && helper.Remaining() > 0 {
		Central.Log.Debugf("------------------ Parse index of structure index=%d name=%s", index, value.Type().Name())
		Central.Log.Debugf("index=%d remaining Buffer structure remaining=%d pos=%d",
			index, helper.Remaining(), helper.offset)
		//peIndex := value.PeriodIndex() + 1 + uint32(index)
		values, pErr := parseBufferTypes(helper, option, value, uint32(index))
		if pErr != nil {
			res = EndTraverser
			err = pErr
			Central.Log.Debugf("Parse buffer error in structure %s:%v", value.adatype.Name(), err)
			return
		}
		if len(value.Elements) <= index {
			element := newStructureElement()
			value.Elements = append(value.Elements, element)
			value.elementMap[uint32(index)] = element
		}
		if values != nil && value.adatype.Type() != FieldTypeGroup {
			value.Elements[index].Values = values
		}
		index++
		Central.Log.Debugf("------------------ Ending Parse index of structure index=%d len elements=%d", index, len(value.Elements))
	}
	Central.Log.Debugf("Sructure parse ready for %s index=%d occ=%d value length=%d pos=%d",
		value.Type().Name(), index, occNumber, len(value.Elements), helper.offset)
	return
}

// Search for structures by name
func (value *StructureValue) search(fieldName string) IAdaValue {
	Central.Log.Debugf("Search field %s elements=%d", fieldName, len(value.Elements))
	for _, val := range value.Elements {
		for _, v := range val.Values {
			Central.Log.Debugf("Searched in value %s", v.Type().Name())
			if v.Type().Name() == fieldName {
				return v
			}
			if v.Type().IsStructure() {
				Central.Log.Debugf("Structure search")
				subValue := v.(*StructureValue).search(fieldName)
				if subValue != nil {
					return subValue
				}
			} else {
				Central.Log.Debugf("No structure search")
			}
		}
	}
	Central.Log.Debugf("Searched field %s not found", fieldName)
	return nil
}

// Traverse Traverse through the definition tree calling a callback method for each node
func (value *StructureValue) Traverse(t TraverserValuesMethods, x interface{}) (ret TraverseResult, err error) {
	Central.Log.Debugf("Traverse level %d structure: %s", value.Type().Level(), value.Type().Name())
	Central.Log.Debugf("Nr sub elements=%d", value.NrElements())
	if value.Elements != nil { // && len(value.Elements[0].Values) > 0 {
		nr := len(value.Elements)
		for e, val := range value.Elements {
			Central.Log.Debugf("%d: Nr sub values=%d", e, len(val.Values))
			if t.ElementFunction != nil {
				ret, err = t.ElementFunction(value, e, nr, x)
				if err != nil || ret == EndTraverser {
					return
				}
			}
			for i, v := range val.Values {
				Central.Log.Debugf("Traverse node %d.element and %d.value at %s[%d,%d] for %s[%d,%d]", e, i, v.Type().Name(),
					v.PeriodIndex(), v.MultipleIndex(), value.Type().Name(), value.PeriodIndex(), value.MultipleIndex())
				if value.PeriodIndex() != v.PeriodIndex() {
					if value.Type().Type() != FieldTypePeriodGroup {
						//panic("Error index parent not correct")
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
					LogMultiLineString(FormatByteBuffer("DATA: ", v.Bytes()))
				}
				if ret == SkipStructure {
					Central.Log.Debugf("Skip structure tree ... ")
					return Continue, nil
				}
				if v.Type().IsStructure() {
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
				Central.Log.Debugf("Traverse index=%d/%d pfield=%s-field=%s", i, nr, value.Type().Name(), v.Type().Name())
			}
		}
	}
	return Continue, nil
}

// Get get the value of an named tree node with an specific index
func (value *StructureValue) Get(fieldName string, index int) IAdaValue {
	v := value.Elements[index-1]
	for _, vr := range v.Values {
		if vr.Type().Name() == fieldName {
			return vr
		}
		if vr.Type().IsStructure() {
			svr := vr.(*StructureValue).Get(fieldName, index)
			if svr != nil {
				return svr
			}
		}
	}
	return nil
}

// NrElements number of structure values
func (value *StructureValue) NrElements() int {
	if value.Elements == nil {
		return 0
	}
	return len(value.Elements)
}

// NrValues number of structure values
func (value *StructureValue) NrValues(index uint32) int {
	if value.Elements == nil {
		return -1
	}
	return len(value.Elements[index-1].Values)
}

// Value return the values of an structure value
func (value *StructureValue) Value() interface{} {
	return value.Elements
}

// Bytes byte array representation of the value
func (value *StructureValue) Bytes() []byte {
	var empty []byte
	return empty
}

// SetStringValue set the string value of the value
func (value *StructureValue) SetStringValue(stValue string) {
	log.Fatal("Structure set string, not implement yet")
}

// SetValue set value for structure
func (value *StructureValue) SetValue(v interface{}) error {
	Central.Log.Infof("Structure set string, not implement yet %s -> %v", value.Type().Name(), v)
	return nil
}

// FormatBuffer provide the format buffer of this structure
func (value *StructureValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	Central.Log.Debugf("Write FormatBuffer for structure of %s store=%v ", value.Type().Name(), option.StoreCall)
	if option.SecondCall {
		return 0
	}
	structureType := value.Type().(*StructureType)
	recordBufferLength := uint32(0)
	if structureType.NrFields() > 0 {
		Central.Log.Debugf("Structure FormatBuffer %s type=%d nrFields=%d", value.Type().Name(), value.Type().Type(), structureType.NrFields())
		switch value.Type().Type() {
		case FieldTypeMultiplefield:
			if option.StoreCall {
				if value.NrElements() > 0 {
					// if buffer.Len() > 0 {
					// 	buffer.WriteString(",")
					// }
					// for idxElement, element := range value.Elements {
					// 	for idxValue := range element.Values {
					// 		Central.Log.Debugf("StoreCall: %d -> %d", idxElement, idxValue)
					// 		buffer.WriteString(fmt.Sprintf("%s%d", value.Type().Name(), idxElement))
					// 	}
					// }
				}
			} else {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				p := "1-N"
				//r := structureType.Range.FormatBuffer()
				if value.Type().HasFlagSet(FlagOptionPE) {
					buffer.WriteString(value.Type().ShortName() + p + "(C),4")
				} else {
					buffer.WriteString(value.Type().ShortName() + "C,4")
					muType := structureType.SubTypes[0]
					buffer.WriteString(fmt.Sprintf(",%s%s,%d,%s",
						value.Type().ShortName(), p, muType.Length(), muType.Type().FormatCharacter()))
				}

				Central.Log.Debugf("Current MU field %s, search in %d nodes", value.Type().Name(), len(value.Elements))
				recordBufferLength += option.multipleSize
			}
		case FieldTypePeriodGroup:
			if option.StoreCall {

			} else {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				buffer.WriteString(value.Type().ShortName() + "C,4")
				if !value.Type().HasFlagSet(FlagOptionMU) {
					r := structureType.Range.FormatBuffer()
					buffer.WriteString("," + value.Type().ShortName() + r)
				}
				recordBufferLength += option.multipleSize
			}
		default:
		}
	}
	Central.Log.Debugf("Final structure RB FormatBuffer for %s: %s", value.Type().Name(), buffer.String())
	return recordBufferLength
}

// StoreBuffer generate store buffer
func (value *StructureValue) StoreBuffer(helper *BufferHelper) error {
	Central.Log.Debugf("Skip store structured record buffer for %s at %d", value.Type().Name(), len(helper.buffer))
	return nil
}

// addValue Add sub value with given index
func (value *StructureValue) addValue(subValue IAdaValue, index uint32) error {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Add value to list for %s[%d,%d], appending %s[%d,%d]", value.Type().Name(), value.PeriodIndex(), value.MultipleIndex(),
			subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())
	}
	var element *structureElement
	var ok bool
	if value.Elements == nil {
		Central.Log.Debugf("Elements empty")
	} else {
		Central.Log.Debugf("Elements =%d", len(value.Elements))
	}
	curIndex := index
	Central.Log.Debugf("Current add value index = %d", curIndex)
	if element, ok = value.elementMap[curIndex]; !ok {
		element = newStructureElement()
		value.Elements = append(value.Elements, element)
		// fmt.Println(value.Type().Name(), " add index ", curIndex)
		value.elementMap[curIndex] = element
		Central.Log.Debugf("Create new Elements on index %d", curIndex)
	}
	s := fmt.Sprintf("%s-%d-%d", subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Search for %s", s)
	}
	var v IAdaValue
	if v, ok = element.valueMap[s]; ok {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Add sub value found %s[%d:%d] %T",
				v.Type().Name(), v.PeriodIndex(), v.MultipleIndex(), v)
		}
	} else {
		if value.Elements == nil {
			Central.Log.Debugf("Create new list for %s and append", value.Type().Name())
			var values []IAdaValue
			values = append(values, subValue)

			element.Values = values
		} else {
			Central.Log.Debugf("Append list to %s", value.Type().Name())
			element.Values = append(element.Values, subValue)
		}
		Central.Log.Debugf("Add sub value new %s[%d:%d] %T",
			subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex(), subValue)
		if value.Type().Type() == FieldTypePeriodGroup {
			Central.Log.Debugf("%s: Set given Period index %d", value.Type().Name(), (curIndex + 1))
			subValue.setPeriodIndex(curIndex + 1)
		} else {
			Central.Log.Debugf("%s: Set upper Period index %d", value.Type().Name(), value.PeriodIndex())
			subValue.setPeriodIndex(value.PeriodIndex())
		}
		if value.Type().Type() == FieldTypeMultiplefield {
			subValue.setMultipleIndex(curIndex + 1)
		}
		element.valueMap[fmt.Sprintf("%s-%d-%d", subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())] = subValue
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Final list for %s[%d,%d] = %d elements for %s[%d,%d]", value.Type().Name(), value.PeriodIndex(),
			value.MultipleIndex(), len(value.Elements), subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())
	}
	return nil
}

// Int32 not used
func (value *StructureValue) Int32() (int32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 32-bit integer")
}

// UInt32 not used
func (value *StructureValue) UInt32() (uint32, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 32-bit integer")
}

// Int64 not used
func (value *StructureValue) Int64() (int64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 64-bit integer")
}

// UInt64 not used
func (value *StructureValue) UInt64() (uint64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 64-bit integer")
}

// Float not used
func (value *StructureValue) Float() (float64, error) {
	return 0, NewGenericError(105, value.Type().Name(), "64-bit float")
}

func (value *StructureValue) setPeriodIndex(index uint32) {
	Central.Log.Debugf("Set %s structure period index = %d -> %d", value.Type().Name(), value.PeriodIndex(), index)
	value.peIndex = index
	for _, val := range value.Elements {
		for _, v := range val.Values {
			Central.Log.Debugf("Set %s period index in structure %d -> %d", v.Type().Name(), v.PeriodIndex(), index)
			v.setPeriodIndex(1)
		}
	}
}
