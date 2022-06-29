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
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type structureElement struct {
	Values   []IAdaValue
	valueMap map[string]IAdaValue
}

func newStructureElement() *structureElement {
	return &structureElement{valueMap: make(map[string]IAdaValue)}
}

// StructureValueTraverser structure value traverser
type StructureValueTraverser interface {
	Traverse(t TraverserValuesMethods, x interface{}) (ret TraverseResult, err error)
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
	Central.Log.Debugf("Init sub values for %s[%d,%d] -> |%d,%d| - %d init MU fields=%v", value.adatype.Name(), value.PeriodIndex(),
		value.MultipleIndex(), peIndex, muIndex, index, initMuFields)

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
			/*if st.Type() == FieldTypeMultiplefield {
				stv.setMultipleIndex(muIndex)
			}*/
			Central.Log.Debugf("Add to %s[%d,%d] element %s[%d,%d] --> PEindex=%d MUindex=%d index=%d", value.Type().Name(), value.PeriodIndex(),
				value.MultipleIndex(), stv.Type().Name(),
				stv.PeriodIndex(), stv.MultipleIndex(), peIndex, muIndex, index)
			err = value.addValue(stv, peIndex, muIndex)
			if err != nil {
				Central.Log.Debugf("Error (addValue) %v", err)
				return
			}
			if stv.Type().IsStructure() {
				stv.(*StructureValue).initMultipleSubValues(index, peIndex, muIndex, initMuFields)
			}
		}
		Central.Log.Debugf("Finished Init sub values for %s len=%d", value.Type().Name(), len(value.Elements))
	} else {
		Central.Log.Debugf("Skip Init sub values for %s", value.Type().Name())
		// debug.PrintStack()
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
	second     uint32
	needSecond SecondCall
}

func evaluateFieldNames(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	efns := x.(*evaluateFieldNameStructure)
	Central.Log.Debugf("Evaluate field %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		if adaValue.Type().Type() == FieldTypeMultiplefield {
			if efns.second == 0 && adaValue.Type().HasFlagSet(FlagOptionPE) && !adaValue.Type().PeriodicRange().IsSingleIndex() {
				Central.Log.Debugf("Skip PE/multiple field %s in first call (%s)", adaValue.Type().Name(), adaValue.Type().PeriodicRange().FormatBuffer())
				efns.needSecond = ReadSecond
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
	Central.Log.Debugf("EFNS need second call %d", efns.needSecond)
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
	Central.Log.Debugf("Check descriptor read %v", option.DescriptorRead)
	if option.DescriptorRead {
		occNumber = 1
	} else {
		// In the second call the occurrence is available
		if option.SecondCall > 0 && value.Type().Type() == FieldTypePeriodGroup {
			occNumber = value.NrElements()
			Central.Log.Debugf("Second call use available occurrence %d Type %s", occNumber, value.Type().Type().name())
		} else {
			occNumber, err = value.evaluateOccurrence(helper)
			if err != nil {
				return
			}
			Central.Log.Debugf("Call got occurrence %d available Type %s", occNumber, value.Type().Type().name())
		}
	}
	Central.Log.Debugf("PE occurrence %s has %d entries pos=%d", value.Type().Name(), occNumber, helper.offset)
	if occNumber > 0 {
		lastNumber := uint32(occNumber)
		if adaType.peRange.multiplier() != allEntries {
			occNumber = adaType.peRange.multiplier()
		}
		Central.Log.Debugf("%s read %d entries", value.Type().Name(), occNumber)
		if occNumber > 10000 {
			Central.Log.Debugf("Too many occurrences")
			return SkipTree, NewGenericError(181)
		}
		if len(value.Elements) != occNumber {
			peIndex := value.peIndex
			muIndex := uint32(0)
			for i := uint32(0); i < uint32(occNumber); i++ {
				if value.Type().Type() == FieldTypePeriodGroup {
					peIndex = adaType.peRange.index(i+1, lastNumber)
				} else {
					muIndex = i + 1
				}
				Central.Log.Debugf("Work on %s PE=%d MU=%d last=%d PEv=%d PErange=%d MUrange=%d",
					adaType.Name(), peIndex, muIndex, lastNumber, value.peIndex,
					adaType.PeriodicRange().from, adaType.MultipleRange().from)
				value.initMultipleSubValues(i+1, peIndex, muIndex, true)
			}
			if option.SecondCall > 0 &&
				(value.Type().HasFlagSet(FlagOptionPE) && value.Type().Type() == FieldTypeMultiplefield) {
				return value.parsePeriodMultiple(helper, option)
			}
			return value.parsePeriodGroup(helper, option, occNumber)
		}
		for _, e := range value.Elements {
			for _, v := range e.Values {
				v.parseBuffer(helper, option)
			}
		}

	}
	Central.Log.Debugf("No occurrence, check shift of PE empty part, sn=%s mainframe=%v need second=%v pos=%d", value.Type().Name(), option.Mainframe,
		option.NeedSecondCall, helper.offset)
	if option.Mainframe {
		Central.Log.Debugf("Are on mainframe, shift PE empty part pos=%d/%X", helper.offset, helper.offset)
		err = value.shiftEmptyMfBuffer(helper)
		if err != nil {
			return EndTraverser, err
		}
		Central.Log.Debugf("After shift PE empty part pos=%d/%X", helper.offset, helper.offset)
	}
	res = SkipTree
	return
}

func (value *StructureValue) parsePeriodGroup(helper *BufferHelper, option *BufferOption, occNumber int) (res TraverseResult, err error) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Parse period group/structure %s offset=%d/%X need second=%v", value.Type().Name(),
			helper.offset, helper.offset, option.NeedSecondCall)
	}
	/* Evaluate the fields which need to be parsed in the period group */
	tm := TraverserValuesMethods{EnterFunction: evaluateFieldNames}
	efns := &evaluateFieldNameStructure{namesMap: make(map[string]bool), second: option.SecondCall}
	res, err = value.Traverse(tm, efns)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Got %d names got need second=%v was need second=%v", len(efns.names), efns.needSecond, option.NeedSecondCall)
	}
	if option.NeedSecondCall == NoneSecond {
		option.NeedSecondCall = efns.needSecond
	}
	for _, n := range efns.names {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Parse start of name : %s offset=%d/%X need second=%v", n, helper.offset,
				helper.offset, option.NeedSecondCall)
		}
		for i := 0; i < occNumber; i++ {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Get occurrence : %d -> %d", (i + 1), value.NrElements())
			}
			v := value.Get(n, i+1)
			if v == nil {
				return EndTraverser, NewGenericError(171)
			}
			//v.setPeriodIndex(uint32(i + 1))
			if v.Type().IsStructure() {
				st := v.Type().(*StructureType)
				if st.Type() == FieldTypeMultiplefield && st.HasFlagSet(FlagOptionPE) && !st.PeriodicRange().IsSingleIndex() {
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Skip %s PE=%d", st.Name(), v.PeriodIndex())
					}
					if option.NeedSecondCall = ReadSecond; option.StoreCall {
						option.NeedSecondCall = StoreSecond
					}
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Parse PG: need second call %d", option.NeedSecondCall)
					}
				} else {
					nrMu := uint32(1)
					if !st.MultipleRange().IsSingleIndex() {
						nrMu, err = helper.ReceiveUInt32()
						if err != nil {
							return
						}
					}
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Got Nr of Multiple Fields = %d creating them ... for %d (%s/%s)",
							nrMu, v.PeriodIndex(), st.PeriodicRange().FormatBuffer(), st.MultipleRange().FormatBuffer())
					}
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
						muStructure.addValue(sv, v.PeriodIndex(), muIndex)
						if st.PeriodicRange().IsSingleIndex() {
							_, err = sv.parseBuffer(helper, option)
							if err != nil {
								return
							}
						} else {
							if Central.IsDebugLevel() {
								Central.Log.Debugf("MU index %d,%d -> %d", sv.PeriodIndex(), sv.MultipleIndex(), i)
								Central.Log.Debugf("Due to Period and MU field, need second call call (PE/MU) for %s", value.Type().Name())
							}
							if option.NeedSecondCall = ReadSecond; option.StoreCall {
								option.NeedSecondCall = StoreSecond
							}
							if Central.IsDebugLevel() {
								Central.Log.Debugf("Parse PG2: need second call %d", option.NeedSecondCall)
							}
						}
					}
				}
			} else {
				/* Parse field value for each non-structure field */
				res, err = v.parseBuffer(helper, option)
				if err != nil {
					return
				}
				Central.Log.Debugf("Parsed to %s[%d,%d] index is %d", v.Type().Name(), v.PeriodIndex(), v.MultipleIndex(), i+1)
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
			Central.Log.Debugf("Parsed Value %s -> len=%s type=%T", v.Type().Name(), len(v.Bytes()), v)
		}
	}
	Central.Log.Debugf("End parsing MU in PE")
	res = SkipTree
	return
}

// Parse the structure
func (value *StructureValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if option.SecondCall > 0 {
		if !(value.Type().HasFlagSet(FlagOptionPE) && value.Type().Type() == FieldTypeMultiplefield) {
			Central.Log.Debugf("Skip parsing structure value %s offset=%X", value.Type().Name(), helper.offset)
			return
		}
	}
	Central.Log.Debugf("Parse structure buffer %s/%s secondCall=%v offset=%d/%X pe=%v mu=%v", value.Type().Name(), value.Type().ShortName(),
		option.SecondCall, helper.offset, helper.offset, value.adatype.HasFlagSet(FlagOptionPE), value.adatype.HasFlagSet(FlagOptionAtomicFB))
	if value.adatype.HasFlagSet(FlagOptionPE) && value.adatype.HasFlagSet(FlagOptionAtomicFB) {
		return value.parseBufferWithMUPE(helper, option)
	}
	return value.parseBufferWithoutMUPE(helper, option)
}

// Evaluate the occurrence of the structure
func (value *StructureValue) evaluateOccurrence(helper *BufferHelper) (occNumber int, err error) {
	subStructureType := value.adatype.(*StructureType)
	switch {
	case subStructureType.Type() == FieldTypePeriodGroup && subStructureType.peRange.IsSingleIndex():
		Central.Log.Debugf("Single PE index occurence only 1")
		return 1, nil
	case subStructureType.Type() == FieldTypeMultiplefield && subStructureType.muRange.IsSingleIndex():
		Central.Log.Debugf("Single MU index occurence only 1")
		return 1, nil
	case subStructureType.Type() == FieldTypePeriodGroup && subStructureType.HasFlagSet(FlagOptionSingleIndex):
		if len(value.Elements) > 0 {
			return len(value.Elements), nil
		}
		subStructureType.occ = 1
	default:
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Single index flag: %v (%s)", subStructureType.HasFlagSet(FlagOptionSingleIndex), subStructureType.Type().name())
			Central.Log.Debugf("PE range: %s", subStructureType.peRange.FormatBuffer())
			Central.Log.Debugf("MU range: %s", subStructureType.muRange.FormatBuffer())
		}
	}
	// if subStructureType.HasFlagSet(FlagOptionSingleIndex) {
	// 	Central.Log.Debugf("Single index occurence only 1")
	// 	return 1, nil
	// }
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
		case OccNone:
		}
	}
	Central.Log.Debugf("Evaluate occurrence for %s of type %d to %d offset after=%d", value.Type().Name(),
		subStructureType.occ, occNumber, helper.offset)
	return
}

// Parse the buffer containing no PE and MU fields
func (value *StructureValue) parseBufferWithoutMUPE(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Parse Buffer structure without MUPE name=%s offset=%d remaining=%d length=%d value length=%d type=%d", value.Type().Name(),
			helper.offset, helper.Remaining(), len(helper.buffer), len(value.Elements), value.Type().Type())
	}

	var occNumber int
	Central.Log.Debugf("Check descriptor read %v", option.DescriptorRead)
	if option.DescriptorRead {
		occNumber = 1
	} else {
		if option.SecondCall > 0 /*&& value.Type().Type() == FieldTypePeriodGroup */ {
			occNumber = value.NrElements()
			Central.Log.Debugf("Second call use available occurrence %d", occNumber)
		} else {
			occNumber, err = value.evaluateOccurrence(helper)
			if err != nil {
				return
			}
		}
		// TODO Remove because it it only a limit and assert statement
		if occNumber > 4000 && !strings.HasPrefix(value.Type().Name(), "fdt") {
			return SkipTree, NewGenericError(182, value.Type().Name(), occNumber)
		}
	}
	Central.Log.Debugf("Occurrence %d period parent index=%d", occNumber, value.peIndex)
	switch value.Type().Type() {
	case FieldTypePeriodGroup:
		Central.Log.Debugf("Init period group values occurrence=%d mainframe=%v", occNumber, option.Mainframe)
		if occNumber == 0 {
			if option.Mainframe {
				err = value.shiftEmptyMfBuffer(helper)
				if err != nil {
					return EndTraverser, err
				}
			}
			Central.Log.Debugf("Skip PE shifted to offset=%d/%X", helper.offset, helper.offset)
			return
		}
		for i := uint32(0); i < uint32(occNumber); i++ {
			value.initSubValues(i+1, i+1, true)
		}
		Central.Log.Debugf("Init period group sub values finished, elements=%d ", value.NrElements())
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
	// Go through all occurrences and check remaining buffer size
	index := 0
	for index < occNumber && helper.Remaining() > 0 {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("index=%d remaining Buffer structure remaining=%d pos=%d",
				index, helper.Remaining(), helper.offset)
		}
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
			value.elementMap[uint32(index+1)] = element
		}
		if values != nil && value.adatype.Type() != FieldTypeGroup {
			value.Elements[index].Values = values
		}
		index++
		if Central.IsDebugLevel() {
			Central.Log.Debugf("------------------ Ending Parse index of structure index=%d len elements=%d",
				index, len(value.Elements))
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Sructure parse ready for %s index=%d occ=%d value length=%d pos=%d",
			value.Type().Name(), index, occNumber, len(value.Elements), helper.offset)
	}
	return
}

func (value *StructureValue) shiftEmptyMfBuffer(helper *BufferHelper) (err error) {
	if value.Type().Type() == FieldTypeMultiplefield {
		st := value.Type().(*StructureType)
		subType := st.SubTypes[0]
		_, err = helper.ReceiveBytes(subType.Length())
		return

	}
	size := uint32(0)
	t := TraverserMethods{EnterFunction: countPEsize}
	adaType := value.Type().(*StructureType)
	adaType.Traverse(t, 1, &size)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Skip parsing %s/%s type=%s, shift PE empty part %d bytes remaining=%d",
			value.Type().Name(), value.Type().ShortName(), value.Type().Type().name(), size, helper.Remaining())
	}
	_, err = helper.ReceiveBytes(size)
	return
}

// Search for structure field entries by name
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
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Traverse level %d structure: %s", value.Type().Level(), value.Type().Name())
		Central.Log.Debugf("Nr sub elements=%d", value.NrElements())
	}
	if value.Elements != nil { // && len(value.Elements[0].Values) > 0 {
		nr := len(value.Elements)
		for e, val := range value.Elements {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("%d: Nr sub values=%d", e, len(val.Values))
			}
			if t.ElementFunction != nil {
				ret, err = t.ElementFunction(value, e, nr, x)
				if err != nil || ret == EndTraverser {
					return
				}
			}
			for i, v := range val.Values {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Traverse node %d.element  and %d.value at %s[%d,%d] (%s) for %s[%d,%d] (%s)", e, i, v.Type().Name(),
						v.PeriodIndex(), v.MultipleIndex(), v.Type().Type().name(), value.Type().Name(), value.PeriodIndex(),
						value.MultipleIndex(), value.Type().Type().name())
					if value.PeriodIndex() != v.PeriodIndex() {
						if value.Type().Type() != FieldTypePeriodGroup {
							Central.Log.Debugf("!!!!----> Error index parent not correct for %s of %s", v.Type().Name(), value.Type().Name())
						}
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
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Skip structure tree ... ")
					}
					return Continue, nil
				}
				if v.Type().IsStructure() && ret != SkipTree {
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Traverse tree %s", v.Type().Name())
					}
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
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Traverse index=%d/%d pfield=%s-field=%s", i, nr, value.Type().Name(), v.Type().Name())
				}
			}
		}
	}
	return Continue, nil
}

// Get get the value of an named tree node with an specific index
func (value *StructureValue) Get(fieldName string, index int) IAdaValue {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Get field %s index %d -> %d", fieldName, index, len(value.Elements))
	}
	if len(value.Elements) < index {
		Central.Log.Debugf("Not got index")
		return nil
	}
	structElement := value.Elements[index-1]
	if vr, ok := structElement.valueMap[fieldName]; ok {
		Central.Log.Debugf("Got value map entry %#v", structElement.valueMap)
		return vr
	}
	Central.Log.Debugf("Nr values %d", len(structElement.Values))
	for _, vr := range structElement.Values {
		Central.Log.Debugf("Check %s -> %s", vr.Type().Name(), fieldName)
		if vr.Type().Name() == fieldName {
			Central.Log.Debugf("Found index %d to %s[%d,%d]", index, vr.Type().Name(), vr.PeriodIndex(), vr.MultipleIndex())
			return vr
		}
		if vr.Type().IsStructure() {
			svr := vr.(*StructureValue).Get(fieldName, index)
			if svr != nil {
				return svr
			}
		}
	}
	Central.Log.Debugf("No %s entry found with index=%d", fieldName, index)
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
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		switch value.Type().Type() {
		case FieldTypeMultiplefield:
			vi := reflect.ValueOf(v)
			for i := 0; i < vi.Len(); i++ {
				muStructureType := value.Type().(*StructureType)
				sv, typeErr := muStructureType.SubTypes[0].Value()
				if typeErr != nil {
					return typeErr
				}
				sv.setMultipleIndex(uint32(i + 1))
				sv.setPeriodIndex(value.PeriodIndex())
				sv.SetValue(vi.Index(i).Interface())
				value.addValue(sv, value.PeriodIndex(), uint32(i+1))
			}
		case FieldTypePeriodGroup:
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Check preiod group slice possible")
			}
			vi := reflect.ValueOf(v)
			ti := reflect.TypeOf(v)
			if ti.Kind() == reflect.Ptr {
				ti = ti.Elem()
			}
			jsonV, _ := json.Marshal(v)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Work on group entry %s -> %s", ti.Name(), string(jsonV))
			}
			for i := 0; i < vi.Len(); i++ {
				value.initMultipleSubValues(uint32(i+1), uint32(i+1), 0, false)
				if Central.IsDebugLevel() {
					Central.Log.Debugf("%d. Element len is %d", i, len(value.Elements))
				}
				iv := vi.Index(i)
				if iv.Kind() == reflect.Ptr {
					iv = iv.Elem()
				}
				ti = reflect.TypeOf(iv.Interface())
				for j, x := range value.Elements[i].Values {
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Try setting element %d/%d -> %s", i, j, x.Type().Name())
					}

					s := iv.FieldByName(x.Type().Name())
					if s.IsValid() {
						err := x.SetValue(s.Interface())
						if err != nil {
							Central.Log.Debugf("Error seting value for %s", x.Type().Name())
							return err
						}
					} else {
						if Central.IsDebugLevel() {
							Central.Log.Debugf("Try search tag of number of fields %d", ti.NumField())
						}
						sn := extractAdabasTagShortName(ti, x.Type().Name())
						s := iv.FieldByName(sn)
						if s.IsValid() {
							err := x.SetValue(s.Interface())
							if err != nil {
								Central.Log.Debugf("Error setting value for %s", x.Type().Name())
								return err
							}
							// return nil
						} else {
							if Central.IsDebugLevel() {
								Central.Log.Errorf("Invalid or missing field for %s", x.Type().Name())
							}
						}
					}
				}
			}
			if Central.IsDebugLevel() {
				Central.Log.Debugf("PE entries %d", value.NrElements())
			}
		default:
		}
	case reflect.Ptr, reflect.Struct:
		if value.Type().Type() != FieldTypeMultiplefield && value.Type().Type() != FieldTypePeriodGroup {
			Central.Log.Debugf("Check struct possible")
		}
	default:
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Structure set interface, not implement yet %s -> %v", value.Type().Name(), v)
		}
	}
	return nil
}

func extractAdabasTagShortName(ti reflect.Type, searchName string) string {
	for fi := 0; fi < ti.NumField(); fi++ {
		s := ti.FieldByIndex([]int{fi})
		if Central.IsDebugLevel() {
			Central.Log.Debugf("%d Tag =  %s -> %s", fi, s.Tag, s.Name)
		}
		if x, ok := s.Tag.Lookup("adabas"); ok {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Adabas tag: %s", x)
			}
			p := strings.Split(x, ":")
			if len(p) > 2 && p[2] == searchName {
				return s.Name
			}
		}
	}
	return ""
}

func (value *StructureValue) formatBufferSecondCall(buffer *bytes.Buffer, option *BufferOption) uint32 {
	structureType := value.Type().(*StructureType)
	if structureType.Type() == FieldTypeMultiplefield && structureType.HasFlagSet(FlagOptionPE) {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Generate FB for second call [%d,%d]", value.peIndex, value.muIndex)
		}
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}

		x := value.peIndex
		r := structureType.muRange.FormatBuffer()
		buffer.WriteString(value.Type().ShortName())
		buffer.WriteString(strconv.FormatInt(int64(x), 10))
		buffer.WriteString("C,4,B," + value.Type().ShortName())
		buffer.WriteString(strconv.FormatInt(int64(x), 10))
		buffer.WriteString("(" + r + "),")
		buffer.WriteString(strconv.FormatInt(int64(structureType.SubTypes[0].Length()), 10))
		// buffer.WriteString(fmt.Sprintf("%s%dC,4,B,%s%d(%s),%d",
		// 	value.Type().ShortName(), x, value.Type().ShortName(), x, r, structureType.SubTypes[0].Length()))

		if Central.IsDebugLevel() {
			Central.Log.Debugf("FB of second call %s", buffer.String())
		}
		return 4 + structureType.SubTypes[0].Length()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Skip because second call")
	}
	return 0

}

// FormatBuffer provide the format buffer of this structure
func (value *StructureValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Write FormatBuffer for structure of %s store=%v", value.Type().Name(), option.StoreCall)
	}
	if option.SecondCall > 0 {
		return value.formatBufferSecondCall(buffer, option)
	}
	if value.Type().Type() == FieldTypeMultiplefield && value.Type().HasFlagSet(FlagOptionSingleIndex) {
		Central.Log.Debugf("Single index FB?")
		return 0
	}
	structureType := value.Type().(*StructureType)
	recordBufferLength := uint32(0)
	if structureType.NrFields() > 0 {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Structure FormatBuffer %s type=%d nrFields=%d", value.Type().Name(), value.Type().Type(), structureType.NrFields())
		}
		switch value.Type().Type() {
		case FieldTypeMultiplefield:
			// if structureType.HasFlagSet(FlagOptionSingleIndex) {
			// 	fmt.Println("FB:", structureType.peRange.FormatBuffer())
			// }
			if !option.StoreCall {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				p := "1-N"
				//r := structureType.Range.FormatBuffer()
				if value.Type().HasFlagSet(FlagOptionPE) {
					buffer.WriteString(value.Type().ShortName() + p + "(C),4")
				} else {
					buffer.WriteString(value.Type().ShortName() + "C,4,B,")
					muType := structureType.SubTypes[0]
					buffer.WriteString(value.Type().ShortName())
					buffer.WriteString(p + ",")
					buffer.WriteString(strconv.FormatInt(int64(muType.Length()), 10))
					buffer.WriteString("," + muType.Type().FormatCharacter())
					// buffer.WriteString(fmt.Sprintf("%s%s,%d,%s",
					// 	value.Type().ShortName(), p, muType.Length(), muType.Type().FormatCharacter()))
				}
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Current MU field %s, search in %d nodes", value.Type().Name(), len(value.Elements))
				}
				recordBufferLength += option.multipleSize
			}
		case FieldTypePeriodGroup:
			if option.StoreCall {

			} else {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				if !value.Type().HasFlagSet(FlagOptionSingleIndex) {
					buffer.WriteString(value.Type().ShortName() + "C,4,B")
				}
				if Central.IsDebugLevel() {
					Central.Log.Debugf("%s Flag option %d %v %d", structureType.Name(), structureType.flags, structureType.HasFlagSet(FlagOptionPart), FlagOptionPart)
				}
				if !value.Type().HasFlagSet(FlagOptionAtomicFB) && !value.Type().HasFlagSet(FlagOptionPart) {
					r := structureType.peRange.FormatBuffer()
					if Central.IsDebugLevel() {
						Central.Log.Debugf("Add generic format buffer field with range %s", r)
					}
					buffer.WriteString("," + value.Type().ShortName() + r)
				}
				recordBufferLength += option.multipleSize
			}
		default:
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Final structure RB FormatBuffer for %s: %s", value.Type().Name(), buffer.String())
	}
	return recordBufferLength
}

// StoreBuffer generate store buffer
func (value *StructureValue) StoreBuffer(helper *BufferHelper, option *BufferOption) error {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Skip store structured record buffer for %s at %d", value.Type().Name(), len(helper.buffer))
	}
	return nil
}

// addValue Add sub value with given index
func (value *StructureValue) addValue(subValue IAdaValue, peindex uint32, muindex uint32) error {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Add value index PE=%d MU=%d to list for %s[%d,%d], appending %s[%d,%d] %p", peindex, muindex, value.Type().Name(), value.PeriodIndex(), value.MultipleIndex(),
			subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex(), value)
	}
	if value.Type().Type() == FieldTypeMultiplefield && muindex == 0 {
		Central.Log.Debugf("Skip MU index")
		// debug.PrintStack()
		return nil
	}
	//Central.Log.Debugf("Stack trace:\n%s", string(debug.Stack()))
	subValue.SetParent(value)
	var element *structureElement
	var ok bool
	lenElements := 0
	if value.Elements != nil {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Before Elements in list %d", len(value.Elements))
		}
		lenElements = len(value.Elements)
	}
	curIndex := peindex
	if value.Type().Type() == FieldTypeMultiplefield {
		curIndex = muindex
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("curIndex=%d PE index=%d MU index=%d ghost=%v", curIndex, peindex, muindex, subValue.Type().HasFlagSet(FlagOptionMUGhost))
		Central.Log.Debugf("Current add check current index = %d lenElements=%d", curIndex, lenElements)
	}
	if element, ok = value.elementMap[curIndex]; !ok {
		element = newStructureElement()
		value.Elements = append(value.Elements, element)
		value.elementMap[curIndex] = element
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Create new Elements on index %d", curIndex)
		}
	} else {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Elements already part of map %d", curIndex)
		}
	}
	if subValue.PeriodIndex() == 0 {
		subValue.setPeriodIndex(peindex)
	}
	if value.Type().Type() == FieldTypeMultiplefield && subValue.MultipleIndex() == 0 {
		subValue.setMultipleIndex(muindex)
	}
	Central.Log.Debugf("Current period index for %s[%d:%d]", subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())
	s := convertMapIndex(subValue)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Search for %s", s)
	}
	var v IAdaValue
	if v, ok = element.valueMap[s]; ok {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Found sub value found %s[%d:%d] %T",
				v.Type().Name(), v.PeriodIndex(), v.MultipleIndex(), v)
		}
	} else {
		// Check elements list already available
		if value.Elements == nil {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Create new list for %s and append", value.Type().Name())
			}
			// If MU field and index not already initialized, define index
			if value.Type().Type() == FieldTypeMultiplefield {
				/*if subValue.MultipleIndex() == 0 {
					subValue.setMultipleIndex(1)
				} else {*/
				if value.MultipleIndex() != 0 {
					subValue.setMultipleIndex(value.MultipleIndex())
				}
				//}
			}
			var values []IAdaValue
			values = append(values, subValue)
			element.Values = values
		} else {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Append list to %s len=%d", value.Type().Name(), len(element.Values))
			}

			// If MU field and index not already initialized, define index
			if value.Type().Type() == FieldTypeMultiplefield && subValue.MultipleIndex() == 0 {
				subValue.setMultipleIndex(uint32(lenElements + 1))
				// subValue.setMultipleIndex(uint32(len(element.Values) + 1))
			}
			element.Values = append(element.Values, subValue)
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Add sub value new %s[%d:%d] %T previous %s mapIndex %s",
				subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex(), subValue, s, convertMapIndex(subValue))
		}
		element.valueMap[convertMapIndex(subValue)] = subValue
		//element.valueMap[fmt.Sprintf("%s-%d-%d", subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())] = subValue
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Final list for %s[%d,%d] = %d elements for %s[%d,%d]", value.Type().Name(), value.PeriodIndex(),
			value.MultipleIndex(), len(value.Elements), subValue.Type().Name(), subValue.PeriodIndex(), subValue.MultipleIndex())
	}
	return nil
}

func convertMapIndex(subValue IAdaValue) string {
	buf := make([]byte, 0, 30)
	buf = append(buf, subValue.Type().Name()...)
	buf = append(buf, '-')
	buf = strconv.AppendUint(buf, uint64(subValue.PeriodIndex()), 10)
	buf = append(buf, '-')
	buf = strconv.AppendUint(buf, uint64(subValue.MultipleIndex()), 10)
	return string(buf)
}

// Int8 not used
func (value *StructureValue) Int8() (int8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 8-bit integer")
}

// UInt8 not used
func (value *StructureValue) UInt8() (uint8, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 8-bit integer")
}

// Int16 not used
func (value *StructureValue) Int16() (int16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "signed 16-bit integer")
}

// UInt16 not used
func (value *StructureValue) UInt16() (uint16, error) {
	return 0, NewGenericError(105, value.Type().Name(), "unsigned 16-bit integer")
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
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set %s structure period index = %d -> %d", value.Type().Name(), value.PeriodIndex(), index)
	}
	value.peIndex = index
	for _, val := range value.Elements {
		for _, v := range val.Values {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Set %s period index in structure %d -> %d", v.Type().Name(), v.PeriodIndex(), index)
			}
			v.setPeriodIndex(1)
		}
	}
}
