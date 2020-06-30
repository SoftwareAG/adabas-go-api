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
	fileFields      map[string]IAdaType
	fileFieldTree   *StructureType
	activeFields    map[string]IAdaType
	activeFieldTree *StructureType
	Values          []IAdaValue
}

type parserBufferTr struct {
	// contains the helper buffer pointer reading the data from the record buffer
	helper *BufferHelper
	// buffer options used to define second call and others
	option *BufferOption
	// Prefix defines the reference prefix used for @references
	prefix     string
	definition *Definition
}

func parseBufferValues(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	parameter := x.(*parserBufferTr)

	if adaValue.Type().HasFlagSet(FlagOptionReference) {
		name := adaValue.Type().Name()
		if name[0] != '@' {
			adaType := adaValue.Type().(*AdaType)
			delete(parameter.definition.activeFields, adaType.name)
			adaType.name = "@" + adaType.name
			parameter.definition.activeFields[adaType.name] = adaType
		} else {
			name = name[1:]
		}
		adaValue.SetValue(parameter.prefix + name)
		return Continue, nil
	}

	Central.Log.Debugf("Start parsing value .... %s offset=%d/%X type=%s", adaValue.Type().Name(),
		parameter.helper.offset, parameter.helper.offset, adaValue.Type().Type().name())
	Central.Log.Debugf("Parse value %s/%s .... second=%v need second=%v pe=%v", adaValue.Type().ShortName(), adaValue.Type().Name(),
		parameter.option.SecondCall, parameter.option.NeedSecondCall, adaValue.Type().HasFlagSet(FlagOptionPE))
	// On second call, to collect MU fields in an PE group, skip all other parser tasks
	if !(adaValue.Type().HasFlagSet(FlagOptionPE) && adaValue.Type().Type() == FieldTypeMultiplefield) {
		if parameter.option.SecondCall > 0 && !adaValue.Type().HasFlagSet(FlagOptionMUGhost) && adaValue.Type().Type() != FieldTypeLBString {
			Central.Log.Debugf("Second call skip parsing %s", adaValue.Type().Name())
			return Continue, nil
		}
	}
	result, err = adaValue.parseBuffer(parameter.helper, parameter.option)
	Central.Log.Debugf("End Parseing value .... %s pos=%d need second=%v",
		adaValue.Type().Name(), parameter.helper.offset, parameter.option.NeedSecondCall)
	return
}

// Register Register field types
func (def *Definition) Register(t IAdaType) {
	def.fileFields[t.Name()] = t
}

// ParseBuffer method start parsing the definition
func (def *Definition) ParseBuffer(helper *BufferHelper, option *BufferOption, prefix string) (res TraverseResult, err error) {
	if def.Values == nil {
		Central.Log.Debugf("Parse buffer types...")
		def.Values, err = parseBufferTypes(helper, option, def.activeFieldTree, 0)
	} else {
		Central.Log.Debugf("Parse buffer values... avail.=%v", (def.Values != nil))
		x := parserBufferTr{helper: helper, option: option, prefix: prefix, definition: def}
		t := TraverserValuesMethods{EnterFunction: parseBufferValues}
		res, err = def.TraverseValues(t, &x)
		if err != nil {
			Central.Log.Debugf("Error parsing buffer values... %v", err)
			return
		}
		Central.Log.Debugf("End parse buffer values... %p avail.=%v", def, (def.Values != nil))
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
		Central.Log.Debugf("Parent structure value available %T", str)
		parentStructure = str.(*StructureValue)
		parent = parentStructure.adatype.(*StructureType)
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("================== Parse Buffer for IAdaTypes of %s -> value avail.=%v index=%d need second=%v",
			parent.Name(), (parentStructure != nil), peIndex, option.NeedSecondCall)
	}

	types := parent.SubTypes
	var conditionMatrix []byte

	// First get reference field index if index is needed for conditional parsing
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Parent refField=%d length=%d", parent.condition.refField, len(types))
	}
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
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Reference field index=%d length field index=%d need second=%v", refField,
			lengthFieldIndex, option.NeedSecondCall)
	}
	for i := 0; i < refField+1; i++ {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Parse type -> %s offset=%d", types[i].Name(), helper.offset)
		}
		var value IAdaValue
		if parentStructure != nil && len(parentStructure.Elements) > int(peIndex) {
			value = parentStructure.Elements[peIndex].valueMap[types[i].Name()]
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Got out of map ->  ", value, " for index ", peIndex)
			}
		} else {
			if parentStructure != nil {
				Central.Log.Debugf("Len parent structure %d", len(parentStructure.Elements))
			}
		}
		if value == nil {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Value nil, not in parent structure")
			}
			value, err = types[i].Value()
			if err != nil {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Error create value for type ", types[i].String())
				}
				return
			}
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Append value to values : %v %p <- %p", parentStructure, parentStructure, value)
			}
			adaValues = append(adaValues, value)
			if parentStructure != nil && peIndex == 0 {
				if len(parentStructure.Elements) > 0 {
					parentStructure.Elements[0].Values = append(parentStructure.Elements[0].Values, value)
					parentStructure.Elements[0].valueMap[types[i].Name()] = value
				} else {
					x := &structureElement{valueMap: make(map[string]IAdaValue)}
					x.Values = append(x.Values, value)
					x.valueMap[types[i].Name()] = value
					parentStructure.Elements = append(parentStructure.Elements, x)
				}
			}
		}
		// If part of multiple field or period group set index value
		if value.Type().HasFlagSet(FlagOptionPE) {
			value.setPeriodIndex(peIndex + 1)
		}
		if value.Type().HasFlagSet(FlagOptionMUGhost) {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Set MU index to %d", (peIndex + 1))
			}
			value.setMultipleIndex(peIndex + 1)
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Call parse buffer of field %s", types[i].Name())
		}
		_, err = value.parseBuffer(helper, option)
		if err != nil {
			Central.Log.Debugf("Error parse buffer %v", err)
			return
		}
		//var at IAdaType
		at := parent
		// TODO Check why parent not used
		types[i].SetParent(at)

		// Found length field index, calculate end of buffer
		if i == lengthFieldIndex {
			lengthFieldValue := value.(*ubyteValue)
			endOfBuffer += uint32(lengthFieldValue.ByteValue())
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Found end of buffer at %d", endOfBuffer)
			}
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
				Central.Log.Debugf("Get reference field %s %v %d offset=%d(%X)", types[ref].String(), ref, len(types), helper.offset, helper.offset)
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
	} else {
		Central.Log.Debugf("No condition matrix step")
	}
	if lengthFieldIndex > 0 {
		pos, posErr := helper.position(endOfBuffer)
		if posErr != nil {
			err = posErr
			Central.Log.Debugf("Position error %v", posErr)
			return
		}
		if pos == -1 {
			Central.Log.Debugf("Position error")
		}
	}

	if Central.IsDebugLevel() {
		Central.Log.Debugf("================== Ending Parse buffer for IAdaTypes of %v need second=%v", parent, option.NeedSecondCall)
	}

	return
}

// NewDefinition create new Definition instance
func NewDefinition() *Definition {
	def := &Definition{fileFields: make(map[string]IAdaType), activeFieldTree: NewStructure()}
	def.fileFieldTree = def.activeFieldTree
	return def
}

// NewDefinitionWithTypes create new Definition instance adding the given types into the tree
func NewDefinitionWithTypes(types []IAdaType) *Definition {
	def := NewDefinition()
	def.activeFieldTree.SubTypes = types
	def.activeFieldTree.condition = NewFieldCondition()
	def.fileFieldTree = def.activeFieldTree
	def.InitReferences()
	def.activeFields = make(map[string]IAdaType)
	for _, v := range types {
		v.SetParent(def.activeFieldTree)
	}
	initFieldHash(def, types)
	Central.Log.Debugf("Ready creation of definition with types")
	return def
}

// NewDefinitionClone clone new Definition instance using old definition and clone the
// active tree to the new one
func NewDefinitionClone(old *Definition) *Definition {
	newDefinition := NewDefinition()
	newDefinition.fileFieldTree = old.fileFieldTree
	newDefinition.fileFields = old.fileFields
	newDefinition.activeFieldTree = old.fileFieldTree
	return newDefinition
}

func initFieldHash(def *Definition, types []IAdaType) {
	for _, v := range types {
		def.fileFields[v.Name()] = v
		def.activeFields[v.Name()] = v
		if v.IsStructure() && v.Type() != FieldTypeMultiplefield {
			sv := v.(*StructureType)
			initFieldHash(def, sv.SubTypes)
		}
	}
}

// Adapt parent reference to inherit flags
func adaptParentReference(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	adaType.SetParent(parentType)
	adaType.SetLevel(uint8(level))
	if adaType.Type() == FieldTypeMultiplefield {
		p := adaType.GetParent()
		for p != nil {
			p.AddFlag(FlagOptionAtomicFB)
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
				Central.Log.Debugf("Adapt parent field flags for %s, need atomic FB", currentType.ShortName())
				currentType.AddFlag(FlagOptionAtomicFB)
				// TODO Adapt current type to adapt parent information
				currentType = currentType.GetParent()
			}
		}
		if adaType.HasFlagSet(FlagOptionAtomicFB) && adaType.IsStructure() {
			structureType := adaType.(*StructureType)
			Central.Log.Debugf("Adapt sub field flags for %s, need atomic FB", adaType.ShortName())
			for _, t := range structureType.SubTypes {
				t.AddFlag(FlagOptionAtomicFB)
			}

		}
	}
	return nil
}

// String return the content of the definition
func (def *Definition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Definition types:\n")
	t := TraverserMethods{EnterFunction: func(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
		output := fmt.Sprintf("%s\n", adaType.String())
		buffer.WriteString(output)
		return nil
	}}

	def.TraverseTypes(t, true, nil)
	// for _, value := range def.activeFieldTree.SubTypes {
	// 	output := fmt.Sprintf("%s\n", value.String())
	// 	buffer.WriteString(output)
	// }
	// if len(def.Values) > 0 {
	// 	buffer.WriteString("Definition IAdaTypes:\n")
	// 	for index, value := range def.Values {
	// 		output := fmt.Sprintf("%03d %s=%s\n", (index + 1), value.Type().Name(), value.String())
	// 		buffer.WriteString(output)
	// 	}
	// }
	return buffer.String()
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
func traverserCreateValue(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
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
	if adaType.IsStructure() && adaType.Type() != FieldTypeRedefinition {
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
	t := TraverserMethods{EnterFunction: traverserCreateValue}
	err = def.TraverseTypes(t, true, parameter)
	Central.Log.Debugf("Done creating values ... %v", err)
	Central.Log.Debugf("Created %d values", len(def.Values))
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
	Central.Log.Debugf("Found value to add to %s[%d,%d] type=%v %T %T", val.Type().Name(),
		val.PeriodIndex(), val.MultipleIndex(), val.Type().Type().name(), val, val.Type())
	switch val.Type().Type() {
	case FieldTypeMultiplefield:
		sv := val.(*StructureValue)
		st := sv.Type().(*StructureType)
		//sv.Type().
		//	sv.Elements = append(sv.Elements, subValue)
		// if len(sv.Elements) == 0 {
		// 	e := &structureElement{}
		// 	Central.Log.Debugf("Add empty element to %s",sv.Type().Name())
		// 	sv.Elements = append(sv.Elements, e)
		// }
		// if len(sv.Elements[0].Values) >= int(index[0]) {
		// 	Central.Log.Debugf("Adapt %#v", st.SubTypes)
		// 	subValue := sv.Elements[0].Values[int(index[0]-1)]
		// 	err = subValue.SetValue(x)
		// } else {
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
		// }
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
