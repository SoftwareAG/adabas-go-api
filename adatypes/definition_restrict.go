/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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

func removeStructure(adaType IAdaType, fieldMap *fieldMap, fq *fieldQuery, ok bool, parentLast bool) {
	oldStructure := adaType.(*StructureType)
	newStructure := NewStructure()
	*newStructure = *oldStructure
	if fq != nil && fq.fieldRange != nil {
		switch adaType.Type() {
		case FieldTypeMultiplefield:
			if adaType.HasFlagSet(FlagOptionMUGhost) {
				newStructure.peRange = *fq.fieldRange[0]
				newStructure.muRange = *fq.fieldRange[1]
				Central.Log.Debugf("-------<<<< PE/MU Range %s=[%s,%s]", adaType.Name(),
					fq.fieldRange[0].FormatBuffer(), fq.fieldRange[1].FormatBuffer())
			} else {
				newStructure.muRange = *fq.fieldRange[0]
				Central.Log.Debugf("-------<<<< MU Range %s=%s", adaType.Name(), fq.fieldRange[0].FormatBuffer())
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
	Central.Log.Debugf("Structure=%p -> %s", newStructure, newStructure.Name())
	newStructure.SubTypes = []IAdaType{}
	fieldMap.evaluateTopLevelStructure(newStructure.Level())
	fieldMap.lastStructure.SubTypes = append(fieldMap.lastStructure.SubTypes, newStructure)
	Central.Log.Debugf("%s part flag %v", fieldMap.lastStructure.Name(), fieldMap.lastStructure.HasFlagSet(FlagOptionPart))
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
	Central.Log.Debugf("Add structure for active tree %d >%s< remove=%v parent %d >%s<", newStructure.Level(),
		adaType.Name(), newStructure.HasFlagSet(FlagOptionToBeRemoved), fieldMap.lastStructure.Level(), fieldMap.lastStructure.Name())
	fieldMap.lastStructure = newStructure
	fieldMap.stackStructure.Push(fieldMap.lastStructure)
	fieldMap.strCount[adaType.Name()] = newStructure
	Central.Log.Debugf("Create structure %s value=%p to %p parent=%p remove=%v", newStructure.Name(), newStructure,
		fieldMap.lastStructure, newStructure.parentType, newStructure.HasFlagSet(FlagOptionToBeRemoved))

}

func removeFieldEnterTrav(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
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
			Central.Log.Debugf("%s -> %s", adaType.Name(), fq.fieldRange[0].FormatBuffer())
			index := 0
			t := adaType.(*AdaType)
			if adaType.HasFlagSet(FlagOptionPE) {
				t.peRange = *fq.fieldRange[index]
				index++
			}
			if len(fq.fieldRange) > index && adaType.HasFlagSet(FlagOptionMUGhost) {
				t.muRange = *fq.fieldRange[index]
			}
			Central.Log.Debugf("%s peRange=%s muRange=%s", t.name, t.peRange.FormatBuffer(), t.muRange.FormatBuffer())

		}

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
			case FieldTypeHyperDesc:
			case FieldTypePhonetic:
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
						Central.Log.Debugf("FB FQ peRange=%#v", r)
					}
					Central.Log.Debugf("FB %s peRange=%s muRange=%s", newType.name, newType.peRange.FormatBuffer(), newType.muRange.FormatBuffer())
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

// ShouldRestrictToFields Restrict the tree to contain only the given nodes, remove the value tree
func (def *Definition) ShouldRestrictToFields(fields string) (err error) {
	if fields == "*" {
		def.activeFieldTree = def.fileFieldTree
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
		//fmt.Println(s.Name(), s.IsSpecialDescriptor())
		if !s.IsSpecialDescriptor() {
			newTypes = append(newTypes, s)
		}
	}
	def.activeFieldTree.SubTypes = newTypes
	return nil
}

func (def *Definition) newFieldMap(field []string) (*fieldMap, error) {
	// BUG(tkn) Check if fields are valid!!!!
	fieldMap := &fieldMap{definition: def}
	fieldMap.set = make(map[string]*fieldQuery)
	fieldMap.strCount = make(map[string]*StructureType)
	fieldMap.stackStructure = NewStack()
	if len(field) != 0 {
		for _, f := range field {
			Central.Log.Debugf("Map new Field %s", f)
			fl := strings.Trim(f, " ")
			var re = regexp.MustCompile(`(?P<field>[^\[\(\]\)]+)(\[(?P<if>[\dN]+),?(?P<it>[\dN]*)\])?(\((?P<ps>\d+),(?P<pt>\d+)\))?`)
			mt := re.FindStringSubmatch(fl)

			// fmt.Printf("FindAllString %#v\n", mt)
			// fmt.Printf("%q\n", re.SubexpNames())
			// fmt.Printf("Got %s %s\n--------\n", mt[1], mt[3])
			fl = mt[1]
			s := mt[3]
			// for i, match := range re.FindAllString(fl, -1) {
			// 	fmt.Printf("Found at index %#v at %d\n", match, i)
			// }
			// m := re.FindAllStringSubmatch(fl, -1)
			// fmt.Printf("%#v\n", re.FindStringSubmatch(fl))
			// fmt.Printf("SubExpNames %#v\n", re.SubexpNames())
			// fmt.Printf("%#v\n", m)
			// fl = m[0][1]
			// s := m[0][2]
			t := mt[4]
			ps := 0
			pt := 0
			if fl != "" && !strings.HasPrefix(fl, "#ISN") {
				rf := false
				if f[0] == '@' {
					fl = f[1:]
					rf = true
				}
				Central.Log.Debugf("%s=>%s index=[%s,%s](%d,%d)", f, fl, s, t, ps, pt)
				fq := &fieldQuery{name: fl, reference: rf}
				if s != "" {
					fq.fieldRange = []*AdaRange{NewRangeParser(s)}
					if t != "" {
						fq.fieldRange = append(fq.fieldRange, NewRangeParser(t))
					}
				}

				if mt[6] != "" {
					var err error
					ps, err = strconv.Atoi(mt[6])
					if err != nil {
						return nil, err
					}
					pt, err = strconv.Atoi(mt[7])
					if err != nil {
						return nil, err
					}
					fq.partialRange = NewRange(ps, pt)
					Central.Log.Debugf("Partial %d:%d\n", fq.partialRange.from, fq.partialRange.to)
				}
				if Central.IsDebugLevel() {
					r := ""
					if fq.fieldRange != nil && len(fq.fieldRange) > 0 {
						if fq.fieldRange[0] == nil {
							panic("field Range nil")
						}
						r = fq.fieldRange[0].FormatBuffer()
					}
					Central.Log.Debugf("Add to map: %s -> %s reference=%v", fq.name, r, rf)
				}
				fieldMap.set[fl] = fq
			}

			// if fl != "" && !strings.HasPrefix(fl, "#ISN") {
			// 	rf := false
			// 	if f[0] == '@' {
			// 		fl = f[1:]
			// 		rf = true
			// 	}
			// 	b := strings.Index(fl, "[")
			// 	var r *AdaRange
			// 	if b > 0 {
			// 		fn := fl[:b]
			// 		e := strings.Index(fl, "]")
			// 		r = NewRangeParser(fl[b+1 : e])
			// 		if r == nil {
			// 			return nil, NewGenericError(129, f)
			// 		}
			// 		Central.Log.Debugf("Add to map: %s -> %s reference=%v", fn, r.FormatBuffer(), rf)
			// 		fieldMap.set[fn] = &fieldQuery{name: fn, fieldRange: []*AdaRange{r}, reference: rf}
			// 	} else {
			// 		Central.Log.Debugf("Add to map: %s reference=%v", fl, rf)
			// 		fieldMap.set[fl] = &fieldQuery{name: fl, reference: rf}
			// 	}
			// }
		}
	}
	fieldMap.parentStructure = NewStructure()
	fieldMap.parentStructure.AddFlag(FlagOptionToBeRemoved)
	fieldMap.lastStructure = fieldMap.parentStructure
	Central.Log.Debugf("Init parent structure %v", fieldMap.lastStructure.HasFlagSet(FlagOptionToBeRemoved))
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

// ShouldRestrictToFieldSlice Restrict the tree to contain only the given nodes
// the corresponding remove flag is set to all fields which are not part of the query
func (def *Definition) ShouldRestrictToFieldSlice(field []string) (err error) {
	Central.Log.Debugf("Should restrict fields to %#v", field)
	if Central.IsDebugLevel() {
		def.DumpTypes(true, false, "before restrict")
	}
	def.Values = nil
	def.activeFields = make(map[string]IAdaType)

	fieldMap, ferr := def.newFieldMap(field)
	if ferr != nil {
		err = ferr
		return
	}
	if Central.IsDebugLevel() {
		def.DumpTypes(true, false, "enter restrict")
	}
	t := TraverserMethods{EnterFunction: removeFieldEnterTrav}
	err = def.TraverseTypes(t, true, fieldMap)
	if err != nil {
		return
	}
	if Central.IsDebugLevel() {
		def.DumpTypes(true, false, "remove restrict restrict")
	}

	if len(fieldMap.set) > 0 {
		Central.Log.Debugf("Field map ... %v", fieldMap.set)
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
		def.DumpTypes(true, false, "Not init restricted")
		def.DumpTypes(true, true, "final restricted")
	}
	return
}

// removeFromTree remove field from tree because of given remove flag
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
			Central.Log.Debugf("Contains %s", t.Name())
		}
	}
}

// SetValueData dependent to the struct interface field the corresponding
// reflection value will be set
func SetValueData(s reflect.Value, v IAdaValue) error {
	Central.Log.Debugf("%s = %s", v.Type().Name(), s.Type().Name())
	switch s.Interface().(type) {
	case *int8, *int32, *int64:
		vi, err := v.Int64()
		if err != nil {
			return err
		}
		s.Elem().SetInt(vi)
	case *uint8, *uint32, *uint64:
		vui, err := v.UInt64()
		if err != nil {
			return err
		}
		s.Elem().SetUint(vui)
	case int8, int32, int64:
		vi, err := v.Int64()
		if err != nil {
			return err
		}
		s.SetInt(vi)
	case uint8, uint32, uint64:
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
		Central.Log.Debugf("Unknown conversion %s/%s", s.Type().String(), v.Type().Name())
		return NewGenericError(80, s.Type(), v.Type().Name())
	}
	return nil
}
