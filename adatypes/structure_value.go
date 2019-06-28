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
	value.initMultipleSubValues(index, peIndex, 0, initMuFields)
}

/*
 * Init sub structures with empty value fields
 */
func (value *StructureValue) initMultipleSubValues(index uint32, peIndex uint32, muIndex uint32, initMuFields bool) {
	subType := value.adatype.(*StructureType)
	Central.Log.Debugf("Init sub values for %s[%d,%d] -> |%d,%d| - %d", value.adatype.Name(), value.PeriodIndex(),
		value.MultipleIndex(), peIndex, muIndex, index)

	if value.Type().Type() != FieldTypeMultiplefield || initMuFields {
		for _, st := range subType.SubTypes {
			if st.HasFlagSet(FlagOptionMUGhost) && !initMuFields {
				continue
			}
			Central.Log.Debugf("Init sub structure %s(%s) for structure %s period index=%d multiple index=%d",
				st.Name(), st.Type().name(), value.Type().Name(), peIndex, muIndex)
			stv, err := st.Value()
			if err != nil {
				Central.Log.Debugf("Error %v", err)
				return
			}
			stv.setPeriodIndex(peIndex)
			stv.setMultipleIndex(muIndex)
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
	names      []string
	namesMap   map[string]bool
	second     bool
	needSecond bool
}

func evaluateFieldNames(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	efns := x.(*evaluateFieldNameStructure)
	Central.Log.Debugf("Evaluate field %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		if adaValue.Type().Type() == FieldTypeMultiplefield {
			if !efns.second && adaValue.Type().HasFlagSet(FlagOptionPE) {
				Central.Log.Debugf("Skip PE/multiple field %s in first call", adaValue.Type().Name())
				efns.needSecond = true
				return SkipTree, nil
			} else if _, ok := efns.namesMap[adaValue.Type().Name()]; !ok {
				Central.Log.Debugf("Add multiple field %s", adaValue.Type().Name())
				efns.names = append(efns.names, adaValue.Type().Name())
				efns.namesMap[adaValue.Type().Name()] = true
			}
		}
	} else {
		if !adaValue.Type().HasFlagSet(FlagOptionMUGhost) {
			if _, ok := efns.namesMap[adaValue.Type().Name()]; !ok {
				Central.Log.Debugf("Add field %s", adaValue.Type().Name())
				efns.names = append(efns.names, adaValue.Type().Name())
				efns.namesMap[adaValue.Type().Name()] = true
			}
		}
	}
	return Continue, nil
}

func countPEsize(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	size := x.(*uint32)

	if adaType.Type() != FieldTypeMultiplefield && !adaType.HasFlagSet(FlagOptionMUGhost) {
		*size = *size + adaType.Length()
		Central.Log.Debugf("Add to PE size: %s -> %d", adaType.Name(), *size)
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
	if value.Type().Type() != FieldTypePeriodGroup &&
		!(value.Type().HasFlagSet(FlagOptionPE) && value.Type().Type() == FieldTypeMultiplefield) {
		Central.Log.Debugf("Skip not group -> %s", value.Type().Name())
		return
	}
	Central.Log.Debugf("%s/%s parse buffer for MU in PE/first call", value.Type().Name(), value.Type().ShortName())
	var occNumber int
	// TODO
	//	if option.SecondCall {
	occNumber, err = value.evaluateOccurrence(helper)
	//	} else {
	//		occNumber = value.NrElements()
	//	}
	Central.Log.Debugf("PE occurence %s has %d entries pos=%d", value.Type().Name(), occNumber, helper.offset)
	if occNumber > 0 {
		lastNumber := uint32(occNumber)
		if adaType.peRange.multiplier() != allEntries {
			occNumber = adaType.peRange.multiplier()
		}
		Central.Log.Debugf("%s read %d entries", value.Type().Name(), occNumber)
		if occNumber > 10000 {
			Central.Log.Debugf("Too many occurences")
			panic("Too many occurence entries")
		}
		peIndex := value.peIndex
		muIndex := uint32(0)
		for i := uint32(0); i < uint32(occNumber); i++ {
			if value.Type().Type() == FieldTypePeriodGroup {
				peIndex = adaType.peRange.index(i+1, lastNumber)
			} else {
				muIndex = i + 1
			}
			Central.Log.Debugf("Work on %d/%d", peIndex, lastNumber)
			value.initMultipleSubValues(i, peIndex, muIndex, true)
		}
		if option.SecondCall &&
			(value.Type().HasFlagSet(FlagOptionPE) && value.Type().Type() == FieldTypeMultiplefield) {
			return value.parsePeriodMultiple(helper, option)
		}
		return value.parsePeriodGroup(helper, option, occNumber)
	}
	Central.Log.Debugf("No occurence, check shift of PE empty part %v need second=%v pos=%d", option.Mainframe,
		option.NeedSecondCall, helper.offset)
	if option.Mainframe {
		Central.Log.Debugf("Are on mainframe, shift PE empty part")
		value.shiftPeriod(helper)
	}

	res = SkipStructure
	return
}

func (value *StructureValue) parsePeriodGroup(helper *BufferHelper, option *BufferOption, occNumber int) (res TraverseResult, err error) {
	Central.Log.Debugf("Parse period group/structure %s offset=%d/%X need second=%v", value.Type().Name(),
		helper.offset, helper.offset, option.NeedSecondCall)
	/* Evaluate the fields which need to be parsed in the period group */
	tm := TraverserValuesMethods{EnterFunction: evaluateFieldNames}
	efns := &evaluateFieldNameStructure{namesMap: make(map[string]bool), second: option.SecondCall}
	res, err = value.Traverse(tm, efns)
	Central.Log.Debugf("Got %d names got need second=%v was need second=%v", len(efns.names), efns.needSecond, option.NeedSecondCall)
	if !option.NeedSecondCall {
		option.NeedSecondCall = efns.needSecond
	}
	for _, n := range efns.names {
		Central.Log.Debugf("Parse start of name : %s offset=%d/%X need second=%v", n, helper.offset,
			helper.offset, option.NeedSecondCall)
		for i := 0; i < occNumber; i++ {
			Central.Log.Debugf("Get occurence : %d -> %d", (i + 1), value.NrElements())
			v := value.Get(n, i+1)
			//v.setPeriodIndex(uint32(i + 1))
			if v.Type().IsStructure() {
				st := v.Type().(*StructureType)
				if st.Type() == FieldTypeMultiplefield && st.HasFlagSet(FlagOptionPE) {
					Central.Log.Debugf("Skip %s PE=%d", st.Name(), v.PeriodIndex())
					option.NeedSecondCall = true
				} else {
					nrMu, nrMerr := helper.ReceiveUInt32()
					if nrMerr != nil {
						err = nrMerr
						return
					}
					Central.Log.Debugf("Got Nr of Multiple Fields = %d creating them ... for %d", nrMu, v.PeriodIndex())
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
						//sv.setPeriodIndex(uint32(i + 1))
						sv.setPeriodIndex(v.PeriodIndex())
						muStructure.addValue(sv, muIndex)
						Central.Log.Debugf("MU index %d,%d -> %d", sv.PeriodIndex(), sv.MultipleIndex(), i)
						Central.Log.Debugf("Due to Period and MU field, need second call call (PE/MU) for %s", value.Type().Name())
						option.NeedSecondCall = true
					}
				}
			} else {
				/* Parse field value for each non-structure field */
				res, err = v.parseBuffer(helper, option)
				if err != nil {
					return
				}
				Central.Log.Debugf("%s parsed to %d,%d", v.Type().Name(), v.PeriodIndex(), v.MultipleIndex())
				// if value.Type().Type() == FieldTypeMultiplefield {
				// 	v.setMultipleIndex(uint32(i + 1))
				// 	Central.Log.Debugf("MU index %d,%d -> %d", v.PeriodIndex(), v.MultipleIndex(), i)
				// }
			}
		}
		Central.Log.Debugf("Parse end of name : %s offset=%d/%X", n, helper.offset, helper.offset)
	}
	res = SkipStructure
	return
}

func (value *StructureValue) parsePeriodMultiple(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	Central.Log.Debugf("Parse MU in PE added nodes")
	for _, e := range value.Elements {
		for _, v := range e.Values {
			v.Type().AddFlag(FlagOptionSecondCall)
			res, err = v.parseBuffer(helper, option)
			if err != nil {
				return
			}
			Central.Log.Debugf("New Value %s -> %s", v.Type().Name(), v.String())
		}
	}
	Central.Log.Debugf("End parsing MU in PE")
	res = SkipTree
	return
}

// Parse the structure
func (value *StructureValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall {
		if !(value.Type().HasFlagSet(FlagOptionPE) && value.Type().Type() == FieldTypeMultiplefield) {
			Central.Log.Debugf("Skip parsing %s offset=%X", value.Type().Name(), helper.offset)
			return
		}
	}
	Central.Log.Debugf("Parse structure buffer %s/%s secondCall=%v offset=%d/%X", value.Type().Name(), value.Type().ShortName(),
		option.SecondCall, helper.offset, helper.offset)
	if value.adatype.HasFlagSet(FlagOptionPE) && value.adatype.HasFlagSet(FlagOptionMU) {
		return value.parseBufferWithMUPE(helper, option)
	}
	return value.parseBufferWithoutMUPE(helper, option)
}

// Evaluate the occurence of the structure
func (value *StructureValue) evaluateOccurrence(helper *BufferHelper) (occNumber int, err error) {
	subStructureType := value.adatype.(*StructureType)
	occNumber = math.MaxInt32
	Central.Log.Debugf("Current structure occurrence %d", subStructureType.occ)
	if subStructureType.occ > 0 {
		occNumber = subStructureType.occ
	} else {
		switch subStructureType.occ {
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
	Central.Log.Debugf("Evaluate occurrence for %s of type %d to %d", value.Type().Name(), subStructureType.occ, occNumber)
	return
}

// Parse the buffer containing no PE and MU fields
func (value *StructureValue) parseBufferWithoutMUPE(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	Central.Log.Debugf("Parse Buffer structure without MUPE name=%s offset=%d remaining=%d length=%d value length=%d type=%d", value.Type().Name(),
		helper.offset, helper.Remaining(), len(helper.buffer), len(value.Elements), value.Type().Type())
	var occNumber int
	occNumber, err = value.evaluateOccurrence(helper)
	if err != nil {
		return
	}
	Central.Log.Debugf("Occurrence %d period index=%d", occNumber, value.peIndex)
	switch value.Type().Type() {
	case FieldTypePeriodGroup:
		Central.Log.Debugf("Init period group values occurence=%d mainframe=%v", occNumber, option.Mainframe)
		if occNumber == 0 {
			if option.Mainframe {
				value.shiftPeriod(helper)
			}
			Central.Log.Debugf("Skip PE shifted to offset=%d/%X", helper.offset, helper.offset)
			return
		}
		for i := uint32(0); i < uint32(occNumber); i++ {
			value.initSubValues(i, i+1, true)
		}
		Central.Log.Debugf("Init period group sub values finished")
		return
	case FieldTypeMultiplefield:
		if occNumber == 0 {
			if option.Mainframe {
				adaType := value.Type().(*StructureType)
				helper.ReceiveBytes(adaType.SubTypes[0].Length())
			}
			Central.Log.Debugf("Skip MU shifted to offset=%d/%X", helper.offset, helper.offset)
			return
		}
		Central.Log.Debugf("Init multiple field sub values")
		lastNumber := uint32(occNumber)
		adaType := value.Type().(*StructureType)
		if adaType.muRange.multiplier() != allEntries {
			occNumber = adaType.muRange.multiplier()
		}
		Central.Log.Debugf("Defined range for values: %s", adaType.muRange.FormatBuffer())
		for i := uint32(0); i < uint32(occNumber); i++ {
			muIndex := adaType.muRange.index(i+1, lastNumber)
			Central.Log.Debugf("%d. Work on MU index = %d/%d", i, muIndex, lastNumber)
			value.initMultipleSubValues(i, value.peIndex, muIndex, true)
		}
		Central.Log.Debugf("Init multiple fields sub values finished")
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

func (value *StructureValue) shiftPeriod(helper *BufferHelper) {
	size := uint32(0)
	t := TraverserMethods{EnterFunction: countPEsize}
	adaType := value.Type().(*StructureType)
	adaType.Traverse(t, 1, &size)
	Central.Log.Debugf("Skip parsing, shift PE empty part of size=%d", size)
	helper.ReceiveBytes(size)
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
				Central.Log.Debugf("Traverse index=%d/%d pfield=%s-field=%s", i, nr, value.Type().Name(), v.Type().Name())
			}
		}
	}
	return Continue, nil
}

// Get get the value of an named tree node with an specific index
func (value *StructureValue) Get(fieldName string, index int) IAdaValue {
	Central.Log.Debugf("Get field %s index %d -> %d", fieldName, index, len(value.Elements))
	if len(value.Elements) < index {
		return nil
	}
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
	Central.Log.Fatal("Structure set string, not implement yet")
}

// SetValue set value for structure
func (value *StructureValue) SetValue(v interface{}) error {
	Central.Log.Infof("Structure set string, not implement yet %s -> %v", value.Type().Name(), v)
	return nil
}

// FormatBuffer provide the format buffer of this structure
func (value *StructureValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	Central.Log.Debugf("Write FormatBuffer for structure of %s store=%v", value.Type().Name(), option.StoreCall)
	structureType := value.Type().(*StructureType)
	if option.SecondCall {
		if structureType.Type() == FieldTypeMultiplefield && structureType.HasFlagSet(FlagOptionPE) {
			Central.Log.Debugf("Generate FB for second call [%d,%d]", value.peIndex, value.muIndex)
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}

			x := value.peIndex
			r := structureType.muRange.FormatBuffer()
			buffer.WriteString(fmt.Sprintf("%s%dC,4,B,%s%d(%s),%d",
				value.Type().ShortName(), x, value.Type().ShortName(), x, r, structureType.SubTypes[0].Length()))

			Central.Log.Debugf("FB of second call %s", buffer.String())
			return 4 + structureType.SubTypes[0].Length()
		}
		Central.Log.Debugf("Skip because second call")
		return 0
	}
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
					buffer.WriteString(value.Type().ShortName() + "C,4,B")
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
				buffer.WriteString(value.Type().ShortName() + "C,4,B")
				if !value.Type().HasFlagSet(FlagOptionMU) {
					r := structureType.peRange.FormatBuffer()
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
	subValue.SetParent(value)
	var element *structureElement
	var ok bool
	if value.Elements == nil {
		Central.Log.Debugf("Elements empty")
	} else {
		Central.Log.Debugf("Elements in list %d", len(value.Elements))
	}
	curIndex := index
	Central.Log.Debugf("Current add value index = %d", curIndex)
	if element, ok = value.elementMap[curIndex]; !ok {
		element = newStructureElement()
		value.Elements = append(value.Elements, element)
		// fmt.Println(value.Type().Name(), " add index ", curIndex)
		value.elementMap[curIndex] = element
		Central.Log.Debugf("Create new Elements on index %d", curIndex)
	} else {
		Central.Log.Debugf("Elements already part of map %d", curIndex)
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
		// if value.Type().Type() == FieldTypePeriodGroup {
		// 	Central.Log.Debugf("%s: Set given Period index %d", value.Type().Name(), (curIndex + 1))
		// 	//subValue.setPeriodIndex(curIndex + 1)
		// } else {
		// 	Central.Log.Debugf("%s: Set upper Period index %d", value.Type().Name(), value.PeriodIndex())
		// 	//subValue.setPeriodIndex(value.PeriodIndex())
		// }
		// if value.Type().Type() == FieldTypeMultiplefield {
		// 	subValue.setMultipleIndex(curIndex + 1)
		// }
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
