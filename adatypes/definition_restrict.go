/*
* Copyright © 2019-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type fieldQuery struct {
	name         string
	reference    bool
	fieldRange   []*AdaRange
	partialRange *AdaRange
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

// evaluateTopLevelStructure evaluate the structure node which is responsible
// for the given level.
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

// removeStructure set the `remove` flag to define, that the structure is not
// part of the query.
func removeStructure(adaType IAdaType, fieldMap *fieldMap, fq *fieldQuery, ok bool, parentLast bool) {
	oldStructure := adaType.(*StructureType)
	newStructure := NewStructure()
	*newStructure = *oldStructure
	if fq != nil && fq.fieldRange != nil {
		Central.Log.Debugf("-------<<<< No field Range ")
		switch adaType.Type() {
		case FieldTypeMultiplefield:
			if adaType.HasFlagSet(FlagOptionMUGhost) {
				newStructure.peRange = *fq.fieldRange[0]
				newStructure.muRange = *fq.fieldRange[1]
				Central.Log.Debugf("-------<<<< PE/MU Range %s=[%s,%s]", adaType.Name(),
					fq.fieldRange[0].FormatBuffer(), fq.fieldRange[1].FormatBuffer())
			} else {
				if adaType.HasFlagSet(FlagOptionPE) {
					newStructure.peRange = *fq.fieldRange[0]
					st := newStructure.SubTypes[0].(*AdaType)
					st.peRange = *fq.fieldRange[0]
				} else {
					newStructure.muRange = *fq.fieldRange[0]
					st := newStructure.SubTypes[0].(*AdaType)
					st.muRange = *fq.fieldRange[0]
				}
				Central.Log.Debugf("-------<<<< PE Range %s=%s -> %v", adaType.Name(), fq.fieldRange[0].FormatBuffer(),
					adaType.HasFlagSet(FlagOptionPE))
				if len(fq.fieldRange) > 1 {
					newStructure.peRange = *fq.fieldRange[0]
					st := newStructure.SubTypes[0].(*AdaType)
					st.peRange = *fq.fieldRange[0]
					newStructure.muRange = *fq.fieldRange[1]
					st.muRange = *fq.fieldRange[1]
					Central.Log.Debugf("-------<<<< MU Range %s=%s", adaType.Name(), fq.fieldRange[1].FormatBuffer())
				}
			}
		case FieldTypePeriodGroup:
			newStructure.peRange = *fq.fieldRange[0]
			Central.Log.Debugf("-------<<<< PE Range %s=%s", adaType.Name(), fq.fieldRange[0].FormatBuffer())
		default:
		}
	} else {
		Central.Log.Debugf("-------<<<< Last Range %s=[%s->%s] last=%s pl=%v -> MU=%s %s", adaType.Name(),
			fieldMap.lastStructure.peRange.FormatBuffer(),
			newStructure.peRange.FormatBuffer(), fieldMap.lastStructure.Name(), parentLast,
			newStructure.muRange.FormatBuffer(), fieldMap.lastStructure.muRange.FormatBuffer())
		if parentLast {
			newStructure.peRange = fieldMap.lastStructure.peRange
			//newStructure.muRange = fieldMap.lastStructure.muRange
		}
		Central.Log.Debugf("-------<<<< Org. Range %s=%s %s", adaType.Name(), newStructure.peRange.FormatBuffer(),
			newStructure.muRange.FormatBuffer())
	}
	Central.Log.Debugf("%s current structure parent is %s (%v)", adaType.Name(),
		fieldMap.lastStructure.Name(), fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved))
	newStructure.SubTypes = []IAdaType{}
	fieldMap.evaluateTopLevelStructure(newStructure.Level())
	fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newStructure)
	Central.Log.Debugf("%s -> %s part flag %v", adaType.Name(), fieldMap.lastStructure.Name(), fieldMap.lastStructure.HasFlagSet(FlagOptionPart))
	if fieldMap.lastStructure.HasFlagSet(FlagOptionPart) {
		newStructure.AddFlag(FlagOptionPart)
		Central.Log.Debugf("Set %s part flag %v", newStructure.Name(), newStructure.HasFlagSet(FlagOptionPart))
	}
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
	switch {
	case newStructure.peRange.IsSingleIndex() && newStructure.HasFlagSet(FlagOptionPE):
		Central.Log.Debugf("%s/%s: set single index due to PE range", newStructure.Name(), newStructure.ShortName())
		newStructure.AddFlag(FlagOptionSingleIndex)
	case newStructure.muRange.IsSingleIndex() && newStructure.Type() == FieldTypeMultiplefield:
		Central.Log.Debugf("%s/%s: set single index due to MU range and Type MU", newStructure.Name(), newStructure.ShortName())
		newStructure.AddFlag(FlagOptionSingleIndex)
	default:
	}
	//	if newStructure.peRange.IsSingleIndex() || newStructure.muRange.IsSingleIndex() {
	Central.Log.Debugf("%s: set flag single index pe=%v(%s) mu=%v(%s)", adaType.Name(),
		newStructure.peRange.IsSingleIndex(), newStructure.peRange.FormatBuffer(),
		newStructure.muRange.IsSingleIndex(), newStructure.muRange.FormatBuffer())
	//		newStructure.AddFlag(FlagOptionSingleIndex)
	//	}
	Central.Log.Debugf("Add structure for active tree %d >%s< remove=%v parent %d >%s<", newStructure.Level(),
		adaType.Name(), newStructure.HasFlagSet(FlagOptionToBeRemoved), fieldMap.lastStructure.Level(), fieldMap.lastStructure.Name())
	fieldMap.lastStructure = newStructure
	fieldMap.stackStructure.Push(fieldMap.lastStructure)
	fieldMap.strCount[adaType.Name()] = newStructure
	Central.Log.Debugf("Create structure %s value=%p to %p parent=%p remove=%v", newStructure.Name(), newStructure,
		fieldMap.lastStructure, newStructure.parentType, newStructure.HasFlagSet(FlagOptionToBeRemoved))

}

// searchFieldToSetRemoveFlagTrav traverer method search for fields which are not part of the query
// defined by the `fieldMap` structure.
// In addition the range is set and additional single range flags are set
func searchFieldToSetRemoveFlagTrav(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	fieldMap := x.(*fieldMap)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Check remove field on type %s with parent %s(parent remove=%v)", adaType.Name(), parentType.Name(),
			parentType.HasFlagSet(FlagOptionToBeRemoved))
	}
	// Check if field is in request
	fq, ok := fieldMap.set[adaType.Name()]
	if ok {
		delete(fieldMap.set, adaType.Name())
		fieldMap.definition.activeFields[adaType.Name()] = adaType
	}
	// Structure need to be copied each time because of tree to nodes of fields
	switch {
	case adaType.Type() == FieldTypeRedefinition:
		Central.Log.Debugf("Check redefintion %s", adaType.Name())
		if ok {
			redType := adaType.(*RedefinitionType)
			adaType.RemoveFlag(FlagOptionToBeRemoved)
			for _, s := range redType.SubTypes {
				delete(fieldMap.set, s.Name())
			}
			fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, redType)
		}
	case adaType.IsStructure():
		if adaType.Type() == FieldTypeMultiplefield && !ok && fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved) {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Skip removing MU field %s", adaType.Name())
			}
			return nil
		}
		removeStructure(adaType, fieldMap, fq, ok, parentType.Name() != "" && fieldMap.lastStructure.Name() == parentType.Name())
	default:
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Field %s in map=%v Level=%d < %d", adaType.Name(), ok, fieldMap.lastStructure.Level(),
				adaType.Level())
		}
		fieldMap.evaluateTopLevelStructure(adaType.Level())

		if fq != nil && len(fq.fieldRange) > 0 {
			Central.Log.Debugf("Field range for %s -> %s", adaType.Name(), fq.fieldRange[0].FormatBuffer())
			index := 0
			t := adaType.(*AdaType)
			if adaType.HasFlagSet(FlagOptionPE) {
				t.peRange = *fq.fieldRange[index]
				index++
				if adaType.HasFlagSet(FlagOptionMUGhost) {
					pt := parentType.(*AdaType)
					pt.peRange = t.peRange
				}
			}
			if len(fq.fieldRange) > index && adaType.HasFlagSet(FlagOptionMUGhost) {
				t.muRange = *fq.fieldRange[index]
				pt := parentType.(*AdaType)
				pt.muRange = t.muRange
			}
		}
		Central.Log.Debugf("Field range %s peRange=%s muRange=%s",
			adaType.Name(), adaType.PeriodicRange().FormatBuffer(), adaType.MultipleRange().FormatBuffer())

		// Skip MU field type if parent is not available
		if parentType.Type() == FieldTypeMultiplefield && fieldMap.lastStructure.Name() != parentType.Name() {
			Central.Log.Debugf("Skip MU field %s", adaType.Name())
			return nil
		}

		// Needed to check if not group is selected in query
		remove := fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved)
		// But not if root node
		if fieldMap.lastStructure.Name() == "" {
			remove = true
		}
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Parent node %s has %v", fieldMap.lastStructure.Name(), remove)
		}
		if !ok && remove {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Skip copy to active field, because field %s is not part of map map=%v remove=%v",
					adaType.Name(), ok, remove)
			}
			var p IAdaType
			p = fieldMap.lastStructure
			for {
				if p.GetParent() == nil || p.GetParent().Name() == "" {
					break
				}
				p = p.GetParent()
			}
			if p.Name() != "" {
				p.(*StructureType).addPart()
			}
		} else {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Current parent %d %s -> %d %s map=%v remove=%v", fieldMap.lastStructure.Level(), fieldMap.lastStructure.Name(),
					adaType.Level(), adaType.Name(), ok, remove)
			}
			// Dependent on type create copy of field
			switch adaType.Type() {
			case FieldTypeSuperDesc:
				newType := &AdaSuperType{}
				oldType := adaType.(*AdaSuperType)
				*newType = *oldType
				newType.SetParent(fieldMap.lastStructure)
				newType.peRange = fieldMap.lastStructure.peRange
				newType.muRange = fieldMap.lastStructure.muRange
				fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newType)
				newType.RemoveFlag(FlagOptionToBeRemoved)
			case FieldTypeHyperDesc, FieldTypePhonetic, FieldTypeCollation, FieldTypeReferential, FieldTypeRedefinition:
			default:
				newType := &AdaType{}
				oldType := adaType.(*AdaType)
				*newType = *oldType
				if fq != nil && fq.reference {
					newType.fieldType = FieldTypeString
					newType.length = 0
					newType.AddFlag(FlagOptionReference)
				}
				newType.SetParent(fieldMap.lastStructure)
				newType.peRange = fieldMap.lastStructure.peRange
				newType.muRange = fieldMap.lastStructure.muRange
				if fq != nil {
					newType.partialRange = fq.partialRange
					for i, r := range fq.fieldRange {
						if i == 0 && newType.HasFlagSet(FlagOptionPE) {
							newType.peRange = *r
						} else {
							newType.muRange = *r
						}
						Central.Log.Debugf("Field range peRange=%#v", r)
						if r.IsSingleIndex() {
							newType.AddFlag(FlagOptionSingleIndex)
						}
					}
					Central.Log.Debugf("FB range for field=%s peRange=%s muRange=%s", newType.name, newType.peRange.FormatBuffer(), newType.muRange.FormatBuffer())
				}
				fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newType)
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Add type to %s value=%p count=%d", fieldMap.lastStructure.Name(), fieldMap.lastStructure, fieldMap.lastStructure.NrFields())
					Central.Log.Debugf("Add type entry in structure %s", newType.Name())
				}
				newType.RemoveFlag(FlagOptionToBeRemoved)
				if fieldMap.lastStructure.HasFlagSet(FlagOptionPart) {
					newType.AddFlag(FlagOptionPart)
				}
			}
		}
	}
	return nil
}

// ResetRestrictToFields reset restriction to all field of tree
func (def *Definition) ResetRestrictToFields() {
	Central.Log.Debugf("Reset active tree to complete")
	x := &StructureType{fieldMap: make(map[string]IAdaType)}
	def.activeFieldTree = x
	def.activeFields = make(map[string]IAdaType)
	t := TraverserMethods{EnterFunction: traverseCacheCopy}
	_ = def.fileFieldTree.Traverse(t, 0, def)
}

// ShouldRestrictToFields this method restrict the query to a given comma-separated list
// of fields. If the fields is set to '*', then all fields are read.
// A field definition may contain index information. The index information need to be set
// in square brackets. For example AA[1] will provide the first entry of a multiple field
// or all entries in the first occurence of the period group.
// BB[1,2] will provide the first entry of the period group and the second entry of the
// multiple field.
func (def *Definition) ShouldRestrictToFields(fields string) (err error) {
	def.ResetRestrictToFields()
	if fields == "*" {
		return
	}
	var field []string
	if fields != "" {
		var re = regexp.MustCompile(`(?P<field>[^\[\(\]\),]+(\[[\dN]+,?[\dN]*\])?(\(\d+,\d+\))?),?`)
		mt := re.FindAllStringSubmatch(fields, -1)
		for _, f := range mt {
			field = append(field, f[1])
		}
	}
	Central.Log.Debugf("Split field into slice to %#v", field)
	return def.ShouldRestrictToFieldSlice(field)
}

// RemoveSpecialDescriptors Remove special descriptors from query
func (def *Definition) RemoveSpecialDescriptors() (err error) {
	newTypes := make([]IAdaType, 0)
	for _, s := range def.activeFieldTree.SubTypes {
		if !s.IsSpecialDescriptor() || (s.IsSpecialDescriptor() && !s.IsOption(FieldOptionPE)) {
			newTypes = append(newTypes, s)
		}
	}
	def.activeFieldTree.SubTypes = newTypes
	def.DumpTypes(true, true, "After remove descriptor")
	return nil
}

// newFieldMap create a new `fieldMap` instance used to restrict
// query field set.
func (def *Definition) newFieldMap(field []string) (*fieldMap, error) {
	// BUG(tkn) Check if fields are valid!!!!
	fieldMap := &fieldMap{definition: def}
	fieldMap.set = make(map[string]*fieldQuery)
	fieldMap.strCount = make(map[string]*StructureType)
	fieldMap.stackStructure = NewStack()
	fieldMap.parentStructure = NewStructure()
	fieldMap.parentStructure.AddFlag(FlagOptionToBeRemoved)
	fieldMap.lastStructure = fieldMap.parentStructure
	if len(field) != 0 {
		for _, f := range field {
			Central.Log.Debugf("Add to new Field %s to hash field", f)
			fl := strings.Trim(f, " ")
			if fl != "" {
				var re = regexp.MustCompile(`(?P<field>[^\[\(\]\)]+)(\[(?P<if>[\dN]+),?(?P<it>[\dN]*)\])?(\((?P<ps>\d+),(?P<pt>\d+)\))?`)
				mt := re.FindStringSubmatch(fl)

				fl = mt[1]
				s := mt[3]
				if fl != "" {
					rf := false
					switch {
					case strings.ToLower(f) == "#isn" || strings.ToLower(f) == "#isnquantity" || strings.ToLower(f) == "#key":
					case f[0] == '#':
						// def.IsPeriodGroup(fl)
						//fl = f[1:]
						var adaType IAdaType
						var ok bool
						if adaType, ok = def.fileFields[fl[1:]]; !ok {
							return nil, NewGenericError(0)
						}
						lenType := NewType(FieldTypeFieldLength, fl)
						switch adaType.Type() {
						case FieldTypePeriodGroup, FieldTypeMultiplefield:
							lenType.AddFlag(FlagOptionLengthPE)
						}
						fieldMap.parentStructure.SubTypes = append(fieldMap.parentStructure.SubTypes, lenType)
					case f[0] == '@':
						fl = f[1:]
						rf = true
						fallthrough
					default:
						// if _, ok := def.fileFields[fl]; !ok {
						// 	fmt.Println(fl, "unknown field")
						// 	return nil, NewGenericError(0)
						// }
						fq := newFieldQuery(fl, rf, s, mt[4], mt[6], mt[7])
						fieldMap.set[fl] = fq
					}
				}
			}
		}
	}
	Central.Log.Debugf("initialized field hash map")
	return fieldMap, nil
}

// RestrictFieldSlice Restrict the tree to contain only the given nodes
func (def *Definition) RestrictFieldSlice(field []string) (err error) {
	err = def.ShouldRestrictToFieldSlice(field)
	if err != nil {
		return
	}
	def.fileFieldTree = def.activeFieldTree
	def.fileFields = def.activeFields
	return nil
}

// ShouldRestrictToFieldSlice  this method restrict the query to a given string slice
// of fields. If one field slice entry is set to '*', then all fields are read.
// A field definition may contain index information. The index information need to be set
// in square brackets. For example AA[1] will provide the first entry of a multiple field
// or all entries in the first occurrence of the period group.
// BB[1,2] will provide the first entry of the period group and the second entry of the
// multiple field.
func (def *Definition) ShouldRestrictToFieldSlice(field []string) (err error) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Should restrict fields to %#v", field)
		def.DumpTypes(true, false, "Start status before restrict")
	}
	def.Values = nil
	def.activeFields = make(map[string]IAdaType)

	fieldMap, ferr := def.newFieldMap(field)
	if ferr != nil {
		err = ferr
		return
	}
	// Traverse through field tree to reduce tree to fields which
	// are part of the query
	t := TraverserMethods{EnterFunction: searchFieldToSetRemoveFlagTrav}
	err = def.TraverseTypes(t, true, fieldMap)
	if err != nil {
		return
	}

	if len(fieldMap.set) > 0 {
		Central.Log.Debugf("Field map not empty, unknown fields found ... %v", fieldMap.set)
		for f := range fieldMap.set {
			err = NewGenericError(50, f)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Error restict fieldMap ... %v", err)
				def.DumpTypes(true, false, "error restrict 50")
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
		def.DumpTypes(true, false, "Not active restricted")
		def.DumpTypes(true, true, "final active restricted")
	}
	return
}

// removeFromTree Search for field to be removed and set remove flag
func removeFromTree(value *StructureType) {
	if !value.HasFlagSet(FlagOptionToBeRemoved) {
		Central.Log.Debugf("Field %s already removed", value.Name())
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
			Central.Log.Debugf("Contains %s", t.Name())
		}
	}
}

// SetValueData dependent to the struct interface field the corresponding
// reflection struct the value will be set with the struct value.
func SetValueData(s reflect.Value, v IAdaValue) error {
	Central.Log.Debugf("%s = %s", v.Type().Name(), s.Type().Name())
	switch s.Interface().(type) {
	case *int, *int8, *int16, *int32, *int64:
		vi, err := v.Int64()
		if err != nil {
			return err
		}
		s.Elem().SetInt(vi)
	case *uint, *uint8, *uint16, *uint32, *uint64:
		vui, err := v.UInt64()
		if err != nil {
			return err
		}
		s.Elem().SetUint(vui)
	case int, int8, int16, int32, int64:
		vi, err := v.Int64()
		if err != nil {
			return err
		}
		s.SetInt(vi)
	case uint, uint8, uint16, uint32, uint64:
		vui, err := v.UInt64()
		if err != nil {
			return err
		}
		s.SetUint(vui)
	case string:
		s.SetString(v.String())
	case *string:
		s.Elem().SetString(v.String())
	default:
		Central.Log.Errorf("Unknown conversion %s/%s", s.Type().String(), v.Type().Name())
		return NewGenericError(80, s.Type(), v.Type().Name())
	}
	return nil
}

// newFieldQuery new field query instance is generated using the given parameter
func newFieldQuery(fl string, rf bool,
	s, fRange, pRangeFrom, pRangeTo string) *fieldQuery {
	fq := &fieldQuery{name: fl, reference: rf}
	if s != "" {
		fq.fieldRange = []*AdaRange{NewRangeParser(s)}
		if fRange != "" {
			fq.fieldRange = append(fq.fieldRange, NewRangeParser(fRange))
		}
	}

	ps := 0
	pt := 0
	if pRangeFrom != "" {
		var err error
		ps, err = strconv.Atoi(pRangeFrom)
		if err != nil {
			return nil
		}
		pt, err = strconv.Atoi(pRangeTo)
		if err != nil {
			return nil
		}
		fq.partialRange = NewRange(ps, pt)
		Central.Log.Debugf("Partial field range %d:%d\n", fq.partialRange.from, fq.partialRange.to)
	}
	if Central.IsDebugLevel() {
		r := ""
		if fq.fieldRange != nil && len(fq.fieldRange) > 0 {
			if fq.fieldRange[0] == nil {
				return nil
			}
			r = fq.fieldRange[0].FormatBuffer()
		}
		Central.Log.Debugf("Init field range for field %s -> %s reference=%v", fq.name, r, rf)
	}
	return fq
}
