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
	"strings"
)

// Isn Adabas Internal ISN
type Isn uint64

// Definition struct defines main entry point for parser structure
type Definition struct {
	fileFieldTree   *StructureType
	activeFieldTree *StructureType
	Values          []IAdaValue
}

type parserBufferTr struct {
	helper *BufferHelper
	option *BufferOption
}

func parseBufferValues(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	parameter := x.(*parserBufferTr)

	Central.Log.Debugf("Start parsing value .... %s offset=%d/%X", adaValue.Type().Name(),
		parameter.helper.offset, parameter.helper.offset)
	Central.Log.Debugf("Parse value .... second=%v need second=%v",
		parameter.option.SecondCall, parameter.option.NeedSecondCall)
	// On second call, to collect MU fields in an PE group, skip all other parser tasks
	if parameter.option.SecondCall && !adaValue.Type().HasFlagSet(FlagOptionMUGhost) && adaValue.Type().Type() != FieldTypeLBString {
		return Continue, nil
	}
	result, err = adaValue.parseBuffer(parameter.helper, parameter.option)
	Central.Log.Debugf("End Parseing value .... %s pos=%d", adaValue.Type().Name(), parameter.helper.offset)
	return
}

// ParseBuffer method start parsing the definition
func (def *Definition) ParseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	if def.Values == nil {
		def.Values, err = parseBufferTypes(helper, option, def.activeFieldTree, 0)
	} else {
		x := parserBufferTr{helper: helper, option: option}
		t := TraverserValuesMethods{EnterFunction: parseBufferValues}
		_, err = def.TraverseValues(t, &x)
		if err != nil {
			return
		}
		// for _, v := range def.Values {
		// 	// if v.Type().IsStructure() {
		// 	// 	_, err = parseBufferTypes(helper, option, v, 0)
		// 	// } else {
		// 	v.parseBuffer(helper, option)
		// 	// }
		// }
	}

	return
}

// Parse buffer IAdaTypes, go through all structures and generate corresponding IAdaTypes
func parseBufferTypes(helper *BufferHelper, option *BufferOption, str interface{}, peIndex uint32) (adaValues []IAdaValue, err error) {
	var parent *StructureType
	var parentStructure *StructureValue
	switch str.(type) {
	case *StructureType:
		Central.Log.Debugf("Parent structure value not available")
		parent = str.(*StructureType)
	default:
		Central.Log.Debugf("Parent structure value available")
		parentStructure = str.(*StructureValue)
		parent = parentStructure.adatype.(*StructureType)
	}
	Central.Log.Debugf("================== Parse Buffer for IAdaTypes of %s -> value avail.=%v index=%d",
		parent.Name(), (parentStructure != nil), peIndex)
	var types []IAdaType
	types = parent.SubTypes
	var conditionMatrix []byte

	// First get reference field index if index is needed for conditional parsing
	Central.Log.Debugf("Parent refField=%d length=%d", parent.condition.refField, len(types))
	refField := func() int {
		if parent.condition.refField != NoReferenceField {
			return parent.condition.refField
		}
		return len(types) - 1
	}()

	// Check if length field of structure is defined
	lengthFieldIndex := parent.condition.lengthFieldIndex
	endOfBuffer := helper.offset

	// Create IAdaTypes until reference index or the end of the types
	// if no reference index available
	Central.Log.Debugf("Reference field index=%d length field index=%d", refField, lengthFieldIndex)
	for i := 0; i < refField+1; i++ {
		Central.Log.Debugf("Parse type -> %s offset=%d", types[i].Name(), helper.offset)
		var value IAdaValue
		if parentStructure != nil && len(parentStructure.Elements) > int(peIndex) {
			value = parentStructure.Elements[peIndex].valueMap[types[i].Name()]
			Central.Log.Debugf("Got out of map ->  ", value, " for index ", peIndex)
		}
		if value == nil {
			Central.Log.Debugf("Value nil, not in parent structure")
			value, err = types[i].Value()
			if err != nil {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Error create value for type ", types[i].String())
				}
				return
			}
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Append value to values : %v", parentStructure)
			}
			adaValues = append(adaValues, value)

		}
		// Parse value of the type
		if value.Type().HasFlagSet(FlagOptionPE) {
			value.setPeriodIndex(peIndex + 1)
		}
		if value.Type().HasFlagSet(FlagOptionMUGhost) {
			Central.Log.Debugf("Set MU index to %d", (peIndex + 1))
			value.setMultipleIndex(peIndex + 1)
		}
		Central.Log.Debugf("Call parse buffer of field %s", types[i].Name())
		_, err = value.parseBuffer(helper, option)
		if err != nil {
			return
		}
		var at IAdaType
		at = parent
		// TODO Check why parent not used
		types[i].SetParent(at)

		// Found length field index, calculate end of buffer
		if i == lengthFieldIndex {
			lengthFieldValue := value.(*ubyteValue)
			endOfBuffer += uint32(lengthFieldValue.ByteValue())
			Central.Log.Debugf("Found end of buffer at %d", endOfBuffer)
		}

		// If reference field found, get condition matrix
		if parent != nil && i == parent.condition.refField {
			refValue := value.(*ubyteValue)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Value : %v offset=%d", refValue.ByteValue(), helper.offset)
				Central.Log.Debugf("Found reference field %d %v %s %v", i, refValue.ByteValue(),
					parent.Name(), parent.condition)
			}
			conditionMatrix = parent.condition.conditionMatrix[refValue.ByteValue()]
			if conditionMatrix == nil {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Allthough refernce value given, condition matrix missing offset=%d refField=%v",
						helper.offset, parent.condition.refField, parent)
				}
				err = NewGenericError(120)
				return
			}
		}
	}

	// If condition matrix is found, generate corresponding IAdaTypes for the structure
	if conditionMatrix != nil {
		Central.Log.Debugf("Condition matrix %v", conditionMatrix)
		for _, ref := range conditionMatrix {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Get reference field %s %v %d offset=%d", types[ref].String(), ref, len(types), helper.offset)
			}
			value, subErr := types[ref].Value()
			if subErr != nil {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Error creating field value for %s", types[ref].String())
				}
				err = subErr
				return
			}
			value.setPeriodIndex(peIndex)
			value.parseBuffer(helper, option)
			if int(ref) == lengthFieldIndex {
				lengthFieldValue := value.(*ubyteValue)
				endOfBuffer += uint32(lengthFieldValue.ByteValue())
				Central.Log.Debugf("Found end of buffer at %d", endOfBuffer)
			}

			if Central.IsDebugLevel() {
				Central.Log.Debugf("Got value for type %s: %p", types[ref].String(), value)
			}
			adaValues = append(adaValues, value)
		}
	}
	if lengthFieldIndex > 0 {
		pos, posErr := helper.position(endOfBuffer)
		if posErr != nil {
			err = posErr
			return
		}
		if pos == -1 {
			Central.Log.Debugf("Position error")
		}
	}

	if Central.IsDebugLevel() {
		Central.Log.Debugf("================== Ending Parse buffer for IAdaTypes of %v", parent)
	}

	return
}

// NewDefinition create new Definition instance
func NewDefinition() *Definition {
	def := &Definition{activeFieldTree: NewStructure()}
	def.fileFieldTree = def.activeFieldTree
	return def
}

// NewDefinitionWithTypes create new Definition instance
func NewDefinitionWithTypes(types []IAdaType) *Definition {
	def := NewDefinition()
	def.activeFieldTree.SubTypes = types
	def.activeFieldTree.condition = FieldCondition{
		lengthFieldIndex: -1,
		refField:         NoReferenceField}
	def.fileFieldTree = def.activeFieldTree
	def.InitReferences()
	for _, v := range types {
		v.SetParent(def.activeFieldTree)
	}
	Central.Log.Debugf("Ready creation of definition with types")
	return def
}

// NewDefinitionWithCondition create new definition with condition
// func NewDefinitionWithCondition(types []IAdaType, condition FieldCondition) *Definition {
// 	def := NewDefinition()
// 	def.fileFieldTree = def.activeFieldTree
// 	def.activeFieldTree.SubTypes = types
// 	t := TraverserMethods{EnterFunction: adaptFlags}
// 	def.TraverseTypes(t, false, nil)
// 	def.activeFieldTree.condition = condition
// 	Central.Log.Debugf("Create new defintion with condition %d", condition.lengthFieldIndex)
// 	for _, v := range types {
// 		v.SetParent(def.activeFieldTree)
// 	}
// 	return def
// }

func adaptParentReference(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	adaType.SetParent(parentType)
	adaType.SetLevel(uint8(level))
	if adaType.Type() == FieldTypeMultiplefield {
		p := adaType.GetParent()
		for p != nil {
			p.AddFlag(FlagOptionMU)
			p = p.GetParent()
		}
	}
	adaptFlags(adaType, parentType, level, x)
	return nil
}

// InitReferences Temporary flag inherit on all tree nodes
func (def *Definition) InitReferences() {
	t := TraverserMethods{EnterFunction: adaptParentReference}
	def.TraverseTypes(t, false, nil)
}

// Traverse traverse through the tree of definition calling a callback method
func adaptFlags(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	if parentType != nil {
		if parentType.HasFlagSet(FlagOptionPE) {
			Central.Log.Debugf("%s: Set PE flag", adaType.Name())
			adaType.AddFlag(FlagOptionPE)
		}
		if adaType.Type() == FieldTypeMultiplefield {
			currentType := parentType
			for currentType != nil {
				Central.Log.Debugf("%s: Set MU flag", currentType.Name())
				currentType.AddFlag(FlagOptionMU)
				// TODO Adapt current type to adapt parent information
				currentType = currentType.GetParent()
			}
		}
		if adaType.HasFlagSet(FlagOptionMU) && adaType.IsStructure() {
			structureType := adaType.(*StructureType)
			for _, t := range structureType.SubTypes {
				t.AddFlag(FlagOptionMU)
			}

		}
	}
	return nil
}

// String return the content of the definition
func (def *Definition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Definition types:\n")
	for _, value := range def.activeFieldTree.SubTypes {
		output := fmt.Sprintf("%s\n", value.String())
		buffer.WriteString(output)
	}
	if len(def.Values) > 0 {
		buffer.WriteString("Definition IAdaTypes:\n")
		for index, value := range def.Values {
			output := fmt.Sprintf("%03d %s=%s\n", (index + 1), value.Type().Name(), value.String())
			buffer.WriteString(output)
		}
	}
	return buffer.String()
}

type searchByName struct {
	name    string
	peIndex uint32
	muIndex uint32
	found   IAdaValue
	grFound IAdaValue
}

func searchValueByName(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	search := x.(*searchByName)
	Central.Log.Debugf("Search value by name %s and index %d:%d, found %s %d/%d", search.name, search.peIndex,
		search.muIndex, adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex())
	if adaValue.Type().Name() == search.name {
		if search.peIndex == adaValue.PeriodIndex() &&
			search.muIndex == adaValue.MultipleIndex() {
			search.found = adaValue
			return EndTraverser, nil
		}
		if adaValue.Type().IsStructure() {
			search.grFound = adaValue
		}
	}
	return Continue, nil
}

func searchValueByNameEnd(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	search := x.(*searchByName)
	Central.Log.Debugf("Search end value by name %s and index %d:%d, found %s %d/%d", search.name, search.peIndex,
		search.muIndex, adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex())
	if adaValue.Type().Name() == search.name {
		if search.peIndex == adaValue.PeriodIndex() && adaValue.Type().IsStructure() {
			search.grFound = adaValue
			return EndTraverser, nil
		}
	}
	return Continue, nil
}

// Search search for a specific field structure in the tree
func (def *Definition) Search(fieldName string) IAdaValue {
	x := searchByName{name: fieldName}
	t := TraverserValuesMethods{EnterFunction: searchValueByName}
	_, err := def.TraverseValues(t, &x)
	if err == nil {
		return x.found
	}

	return nil
}

// SearchByIndex search for a specific field structure in the tree of an period group or multiple field
func (def *Definition) SearchByIndex(fieldName string, index []uint32, create bool) (value IAdaValue, err error) {
	var t IAdaType
	t, err = def.SearchType(fieldName)
	if err != nil {
		return
	}

	// Receive main parent
	c := t
	for c.GetParent() != nil && c.GetParent().Name() != "" {
		c = c.GetParent()
	}

	// Main group name if period group use other
	Central.Log.Debugf("Main group parent name : %s", c.Name())
	if c.Type() == FieldTypePeriodGroup {
		var v IAdaValue
		for _, v = range def.Values {
			if v.Type().Name() == c.Name() {
				break
			}
		}
		strv := v.(*StructureValue)
		if index == nil || len(index) == 0 {
			err = NewGenericError(121)
			return
		}
		Central.Log.Debugf("Use index for field %v", index[0])
		element := strv.elementMap[index[0]-1]
		if element == nil {
			if create {
				Central.Log.Debugf("Create new Element %d", index[0])
				strv.initSubValues(index[0]-1, index[0], true)
				element = strv.elementMap[index[0]-1]
			} else {
				err = NewGenericError(122)
				return
			}
		}
		Central.Log.Debugf("Element : %#v", element)
		for _, v = range element.Values {
			x := searchByName{name: fieldName}
			switch {
			case index == nil:
			case len(index) > 1:
				x.peIndex = index[0]
				x.muIndex = index[1]
			case len(index) > 0:
				x.peIndex = index[0]
			default:
			}
			tvm := TraverserValuesMethods{EnterFunction: searchValueByName, LeaveFunction: searchValueByNameEnd}
			_, err = strv.Traverse(tvm, &x)
			if err == nil {
				if x.found != nil {
					Central.Log.Debugf("Found value searching %s under %s", x.found.Type().Name(), strv.Type().Name())
					if x.found.Type().Type() == FieldTypeMultiplefield {
						if len(index) < 2 {
							return nil, NewGenericError(61)
						}
						strv := x.found.(*StructureValue)
						element := strv.elementMap[index[1]]
						if element == nil {
							err = NewGenericError(123)
							return
						}
					}
					value = x.found
					Central.Log.Debugf("Found element: %v", value.Type().Name())
					return
				}
				if x.grFound != nil {
					Central.Log.Debugf("Element not found, but group found: %v[%d:%d]", x.grFound.Type().Name(),
						x.grFound.PeriodIndex(), x.grFound.MultipleIndex())
					if create {
						strv := x.grFound.(*StructureValue)
						st := x.grFound.Type().(*StructureType)
						value, err = st.SubTypes[0].Value()
						if err != nil {
							Central.Log.Debugf("Error creating sub types %v", err)
							value = nil
							return
						}
						strv.addValue(value, index[0])
						value.setPeriodIndex(index[0])
						value.setMultipleIndex(index[1])
						Central.Log.Debugf("New MU value index %d:%d", value.PeriodIndex(), value.MultipleIndex())
						return

					}

				}
			}
			Central.Log.Debugf("Not found or error searching: %v", err)
		}
	} else {
		// No period group
		Central.Log.Debugf("Search index for structure %s %v", fieldName, len(index))
		v := def.Search(fieldName)
		if v == nil {
			err = NewGenericError(124)
			return
		}
		if v.Type().IsStructure() {
			strv := v.(*StructureValue)
			element := strv.elementMap[index[0]]
			if element == nil {
				Central.Log.Debugf("Index on %s no element on index %v", v.Type().Name(), index[0])
				err = NewGenericError(123)
				return
			}
			Central.Log.Debugf("Index on %s found element on index %v", v.Type().Name(), index[0])
			value = element.Values[0]
		} else {
			value = v
			Central.Log.Debugf("Found value %s", value.Type().Name())
		}
		return
	}

	err = NewGenericError(125)
	return
}

// AppendType append the given type to the type list
func (def *Definition) AppendType(adaType IAdaType) {
	def.activeFieldTree.SubTypes = append(def.activeFieldTree.SubTypes, adaType)
	adaType.SetParent(def.activeFieldTree)
}

// Fieldnames list of fields part of the query
func (def *Definition) Fieldnames() []string {
	typeList := make([]string, 0)
	t := TraverserMethods{EnterFunction: func(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
		typeList = append(typeList, adaType.Name())
		return nil
	}}

	def.TraverseTypes(t, true, typeList)
	return typeList
}

type stackParameter struct {
	definition     *Definition
	forStoring     bool
	stack          *Stack
	structureValue *StructureValue
}

func addValueToStructure(parameter *stackParameter, value IAdaValue) {
	Central.Log.Debugf("Add value for %s = %v -> %s", value.Type().Name(), value.String(), value.Type().Type().FormatCharacter())
	if parameter.structureValue == nil {
		Central.Log.Debugf("Add to main")
		parameter.definition.Values = append(parameter.definition.Values, value)
	} else {
		if parameter.structureValue.Type().Type() == FieldTypePeriodGroup {
			parameter.structureValue.addValue(value, 1)
		} else {
			parameter.structureValue.addValue(value, 0)
		}
	}
}

// create value function used in traverser to create a tree per type element
func createValue(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	parameter := x.(*stackParameter)
	if parameter.structureValue != nil {
		Central.Log.Debugf("parent is %s %d for %d", parameter.structureValue.Type().Name(), parameter.structureValue.Type().Level(), adaType.Level())
		for parameter.structureValue != nil && parameter.structureValue.Type().Level() != (adaType.Level()-1) {
			element, _ := parameter.stack.Pop()
			parameter.structureValue = element.(*StructureValue)
			if parameter.structureValue == nil {
				Central.Log.Debugf("Top received")
			} else {
				Central.Log.Debugf("Check parent is %s %d", parameter.structureValue.Type().Name(), parameter.structureValue.Type().Level())
			}
		}
	}
	if parameter.forStoring && adaType.IsSpecialDescriptor() {
		Central.Log.Debugf("For storing or is descriptor, skip creating value")
		return nil
	}
	Central.Log.Debugf("Create value for level=%d %s -> %d", level, adaType.Name(), adaType.Level())
	if adaType.IsStructure() {
		if adaType.Type() != FieldTypePeriodGroup && adaType.HasFlagSet(FlagOptionPE) {
			return nil
		}
		parameter.stack.Push(parameter.structureValue)
		Central.Log.Debugf("Create structure value for %s", adaType.Name())
		value, subErr := adaType.Value()
		if subErr != nil {
			Central.Log.Debugf("Error %v", subErr)
			return subErr
		}
		addValueToStructure(parameter, value)
		parameter.structureValue = value.(*StructureValue)
	} else {
		// Don't create Period group field elements
		if adaType.HasFlagSet(FlagOptionPE) {
			return nil
		}
		// Don't create ghost nodes for MU fields
		if adaType.HasFlagSet(FlagOptionMUGhost) {
			return nil
		}
		if parameter.structureValue == nil {
			Central.Log.Debugf("Add node value %s to main", adaType.Name())
			value, subErr := adaType.Value()
			if subErr != nil {
				Central.Log.Debugf("Error %v", subErr)
				return subErr
			}
			parameter.definition.Values = append(parameter.definition.Values, value)
		} else {
			// Check if value already exists
			ok := false
			if len(parameter.structureValue.Elements) > 0 {
				_, ok = parameter.structureValue.Elements[0].valueMap[adaType.Name()]
			}
			if !ok {
				Central.Log.Debugf("Add node value %s to structure", adaType.Name())
				value, subErr := adaType.Value()
				if subErr != nil {
					Central.Log.Debugf("Error %v", subErr)
					return subErr
				}
				addValueToStructure(parameter, value)
			} else {
				Central.Log.Debugf("Skip because already added")
			}
		}
	}
	Central.Log.Debugf("Finished creating value level=%d name=%s", adaType.Level(), adaType.Name())
	return nil
}

// CreateValues Create new value tree
func (def *Definition) CreateValues(forStoring bool) (err error) {
	def.Values = nil
	Central.Log.Debugf("Create values from types for storing=%v", forStoring)
	parameter := &stackParameter{definition: def, forStoring: forStoring, stack: NewStack()}
	t := TraverserMethods{EnterFunction: createValue}
	err = def.TraverseTypes(t, true, parameter)
	Central.Log.Debugf("Done creating values ... %v", err)
	return
}

type fieldQuery struct {
	name       string
	fieldRange *AdaRange
}

// field map containing structure and definition
type fieldMap struct {
	set             map[string]*fieldQuery
	strCount        map[string]*StructureType
	definition      *Definition
	parentStructure *StructureType
	lastStructure   *StructureType
	stackStructure  *Stack
}

func (fieldMap *fieldMap) evaluateTopLevelStructure(level uint8) {
	Central.Log.Debugf("%d check level %d", fieldMap.lastStructure.Level(), level)
	for fieldMap.lastStructure.Level() >= level {
		popElement, _ := fieldMap.stackStructure.Pop()
		if popElement == nil {
			Central.Log.Debugf("No element in stack")
			fieldMap.lastStructure = fieldMap.parentStructure
			Central.Log.Debugf("Set main structure parent to %v", fieldMap.lastStructure)
			break
		}
		fieldMap.lastStructure = popElement.(*StructureType)
		Central.Log.Debugf("Set new structure parent to %v", fieldMap.lastStructure)
		Central.Log.Debugf("%d check level %d", fieldMap.lastStructure.Level(), level)
	}

}

func removeFieldTraverser(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	fieldMap := x.(*fieldMap)
	Central.Log.Debugf("Check remove field on type %s with parent %s(parent remove=%v)", adaType.Name(), parentType.Name(),
		parentType.HasFlagSet(FlagOptionToBeRemoved))
	// Check if field is in request
	fq, ok := fieldMap.set[adaType.Name()]
	if ok {
		delete(fieldMap.set, adaType.Name())
	}
	// Structure need to be copied each time because of tree to nodes of fields
	if adaType.IsStructure() {
		if adaType.Type() == FieldTypeMultiplefield && !ok && fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved) {
			Central.Log.Debugf("Skip MU field %s", adaType.Name())
			return nil
		}
		oldStructure := adaType.(*StructureType)
		newStructure := NewStructure()
		*newStructure = *oldStructure
		if fq != nil && fq.fieldRange != nil {
			newStructure.Range = *fq.fieldRange
			Central.Log.Debugf("-------<<<< Range %s=%s", adaType.Name(), fq.fieldRange.FormatBuffer())
		}
		Central.Log.Debugf("%s current structure parent is %s (%v)", adaType.Name(),
			fieldMap.lastStructure.Name(), fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved))
		Central.Log.Debugf("Structure=%p -> %s", newStructure, newStructure.Name())
		newStructure.SubTypes = []IAdaType{}
		fieldMap.evaluateTopLevelStructure(newStructure.Level())
		fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newStructure)
		newStructure.SetParent(fieldMap.lastStructure)
		if fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved) {
			if !ok {
				newStructure.AddFlag(FlagOptionToBeRemoved)
			} else {
				newStructure.RemoveFlag(FlagOptionToBeRemoved)
			}
		} else {
			newStructure.RemoveFlag(FlagOptionToBeRemoved)
		}
		Central.Log.Debugf("Add structure for active tree %d >%s< %d >%s<", newStructure.Level(),
			adaType.Name(), fieldMap.lastStructure.Level(), fieldMap.lastStructure.Name())
		fieldMap.lastStructure = newStructure
		fieldMap.stackStructure.Push(fieldMap.lastStructure)
		fieldMap.strCount[adaType.Name()] = newStructure
		Central.Log.Debugf("Create structure %s value=%p to %p parent=%p remove=%v", newStructure.Name(), newStructure,
			fieldMap.lastStructure, newStructure.parentType, newStructure.HasFlagSet(FlagOptionToBeRemoved))
	} else {
		Central.Log.Debugf("In map=%v Level=%d < %d", ok, fieldMap.lastStructure.Level(),
			adaType.Level())
		fieldMap.evaluateTopLevelStructure(adaType.Level())

		// Skip MU field type if parent is not available
		if parentType.Type() == FieldTypeMultiplefield && fieldMap.lastStructure.Name() != parentType.Name() {
			Central.Log.Debugf("Skip MU field %s", adaType.Name())
			return nil
		}

		// Needed to check if not group is selected in query
		remove := fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved)
		if !ok && remove {
			Central.Log.Debugf("Skip copy to active field, because field %s is not part of map map=%v remove=%v",
				adaType.Name(), ok, remove)
		} else {
			Central.Log.Debugf("Current parent %d %s -> %d %s map=%v remove=%v", fieldMap.lastStructure.Level(), fieldMap.lastStructure.Name(),
				adaType.Level(), adaType.Name(), ok, remove)

			// Dependent on type create copy of field
			switch adaType.Type() {
			case FieldTypeSuperDesc:
				newType := &AdaSuperType{}
				oldType := adaType.(*AdaSuperType)
				*newType = *oldType
				newType.SetParent(fieldMap.lastStructure)
				fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newType)
				newType.RemoveFlag(FlagOptionToBeRemoved)
			case FieldTypeHyperDesc:
			case FieldTypePhonetic:
			default:
				newType := &AdaType{}
				oldType := adaType.(*AdaType)
				*newType = *oldType
				newType.SetParent(fieldMap.lastStructure)
				fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newType)
				Central.Log.Debugf("Add type to %s value=%p count=%d %p", fieldMap.lastStructure.Name(), fieldMap.lastStructure, fieldMap.lastStructure.NrFields(), fieldMap.lastStructure.parentType)
				Central.Log.Debugf("Add type entry in structure %s", newType.Name())
				newType.RemoveFlag(FlagOptionToBeRemoved)
			}
		}
	}
	return nil
}

// ShouldRestrictToFields Restrict the tree to contain only the given nodes, remove the value tree
func (def *Definition) ShouldRestrictToFields(fields string) (err error) {
	if fields == "*" {
		def.activeFieldTree = def.fileFieldTree
		return
	}
	var field []string
	if fields != "" {
		field = strings.Split(fields, ",")
	}
	return def.ShouldRestrictToFieldSlice(field)
}

func (def *Definition) newFieldMap(field []string) (*fieldMap, error) {
	// BUG(tkn) Check if fields are valid!!!!
	fieldMap := &fieldMap{definition: def}
	fieldMap.set = make(map[string]*fieldQuery)
	fieldMap.strCount = make(map[string]*StructureType)
	fieldMap.stackStructure = NewStack()
	if field != nil {
		for _, f := range field {
			b := strings.Index(f, "[")
			fl := f
			var r *AdaRange
			if b > 0 {
				fl = f[:b]
				e := strings.Index(f, "]")
				r = NewRangeParser(f[b+1 : e])
				if r == nil {
					return nil, NewGenericError(129, f)
				}
				Central.Log.Debugf("Add to map: %s -> %s", fl, r.FormatBuffer())
			}
			fieldMap.set[fl] = &fieldQuery{name: fl, fieldRange: r}
		}
	}
	fieldMap.parentStructure = NewStructure()
	fieldMap.lastStructure = fieldMap.parentStructure
	return fieldMap, nil
}

// ShouldRestrictToFieldSlice Restrict the tree to contain only the given nodes
func (def *Definition) ShouldRestrictToFieldSlice(field []string) (err error) {
	Central.Log.Debugf("Restrict fields to %#v", field)
	def.Values = nil
	fieldMap, ferr := def.newFieldMap(field)
	if ferr != nil {
		err = ferr
		return
	}
	t := TraverserMethods{EnterFunction: removeFieldTraverser}
	err = def.TraverseTypes(t, true, fieldMap)
	if err != nil {
		return
	}

	if len(fieldMap.set) > 0 {
		for f := range fieldMap.set {
			err = NewGenericError(50, f)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Error restict fieldMap ... %v", err)
				def.DumpTypes(true, false)
			}
			return
		}
	}

	Central.Log.Debugf("Remove/Cleanup empty structures ...")
	for _, strType := range fieldMap.strCount {
		removeFromTree(strType)
	}
	def.activeFieldTree = fieldMap.parentStructure
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Final restricted type tree .........")
		def.DumpTypes(true, true)
	}
	return
}

func removeFromTree(value *StructureType) {
	if !value.HasFlagSet(FlagOptionToBeRemoved) {
		return
	}
	Central.Log.Debugf("Remove empty nodes from value: %s len=%d", value.Name(), value.NrFields())
	if value.NrFields() == 0 {
		Central.Log.Debugf("No sub fields, remove value %s value=%p count=%d", value.Name(), value, value.NrFields())
		Central.Log.Debugf("Remove value: %s fields=%d", value.Name(), value.NrFields())
		if value.parentType != nil {
			parent := value.parentType.(*StructureType)
			parent.RemoveField(&value.CommonType)
			value.SetParent(nil)
			if parent.NrFields() == 0 {
				Central.Log.Debugf("Remove parent: %s cause %d", parent.Name(), parent.NrFields())
				removeFromTree(parent)
			}
		}
	} else {
		Central.Log.Debugf("Value %s value=%p count=%d contains >0 entries:", value.Name(), value, value.NrFields())
		for _, t := range value.SubTypes {
			Central.Log.Debugf("Contains ", t.Name())
		}
	}
}

type search struct {
	name    string
	adaType IAdaType
}

func findType(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	search := x.(*search)
	Central.Log.Debugf("Check search %s:%s search=%s", adaType.Name(), adaType.ShortName(), search.name)
	if adaType.Name() == search.name {
		search.adaType = adaType
		Central.Log.Debugf("Found type ...")
		return NewGenericError(126, search.name)
	}
	return nil
}

// SearchType search for a type definition in the tree
func (def *Definition) SearchType(fieldName string) (adaType IAdaType, err error) {
	search := &search{name: fieldName}
	level := 1
	t := TraverserMethods{EnterFunction: findType}
	if def.fileFieldTree == nil {
		err = def.activeFieldTree.Traverse(t, level+1, search)
	} else {
		err = def.fileFieldTree.Traverse(t, level+1, search)
	}
	if err == nil {
		err = NewGenericError(41, fieldName)
		return
	}
	err = nil
	if search.adaType == nil {
		Central.Log.Debugf("AdaType not found ", fieldName)
		err = NewGenericError(42, fieldName)
		return
	}
	Central.Log.Debugf("Found adaType for search field %s -> %s", fieldName, search.adaType)
	adaType = search.adaType
	return
}

// SetValueWithIndex Add value to an node element
func (def *Definition) SetValueWithIndex(name string, index []uint32, x interface{}) error {
	typ, err := def.SearchType(name)
	if err != nil {
		Central.Log.Debugf("Search type error: %v", err)
		return err
	}
	if len(index) == 2 && typ.Type() != FieldTypeMultiplefield && index[1] > 0 {
		return NewGenericError(62)
	}
	Central.Log.Debugf("Set value %s with index=%#v value=%v", name, index, x)
	var val IAdaValue
	if !typ.HasFlagSet(FlagOptionPE) {
		Central.Log.Debugf("Search name ....%s", name)
		val = def.Search(name)
		if val == nil {
			return NewGenericError(63, name)
		}
	} else {
		Central.Log.Debugf("Search indexed period group ....")
		val, err = def.SearchByIndex(name, index, true)
		if err != nil {
			return err
		}
		if val == nil {
			return NewGenericError(127, name)
		}
	}
	Central.Log.Debugf("Found value to add to %s type=%v [%d,%d] %T %T", val.Type().Name(), val.Type().Type(),
		val.PeriodIndex(), val.MultipleIndex(), val, val.Type())
	switch val.Type().Type() {
	case FieldTypeMultiplefield:
		sv := val.(*StructureValue)
		st := sv.Type().(*StructureType)
		//sv.Type().
		//	sv.Elements = append(sv.Elements, subValue)
		if len(sv.Elements) == 0 {
			e := &structureElement{}
			sv.Elements = append(sv.Elements, e)
		}
		if len(sv.Elements[0].Values) >= int(index[0]) {
			Central.Log.Debugf("Adapt %#v", st.SubTypes)
			subValue := sv.Elements[0].Values[int(index[0]-1)]
			err = subValue.SetValue(x)
		} else {
			subValue, serr := st.SubTypes[0].Value()
			if serr != nil {
				return serr
			}
			err = subValue.SetValue(x)
			if err != nil {
				return err
			}
			err = sv.addValue(subValue, index[0])
			// subValue.setMultipleIndex(index[0])
			// sv.Elements[0].Values = append(sv.Elements[0].Values, subValue)
		}
		Central.Log.Debugf("Add Multiple field, elements=%d", len(sv.Elements))
	default:
		err = val.SetValue(x)
	}
	return err
}

// Descriptors Return slice of descriptor field names given
func (def *Definition) Descriptors(descriptors string) (desc []string, err error) {
	descFields := strings.Split(descriptors, ",")
	for _, d := range descFields {
		adaType, searchErr := def.SearchType(d)
		if searchErr != nil {
			err = searchErr
			return
		}
		if adaType == nil {
			err = NewGenericError(43, d)
			return nil, err
		}
		desc = append(desc, adaType.ShortName())
	}
	Central.Log.Debugf("Descriptors: %v", desc)
	return
}
