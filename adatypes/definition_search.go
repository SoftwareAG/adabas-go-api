/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"strconv"
	"strings"
)

// type search struct {
// 	name    string
// 	adaType IAdaType
// }

type searchByName struct {
	name    string
	peIndex uint32
	muIndex uint32
	found   IAdaValue
	grFound IAdaValue
}

func traverseSearchValueByName(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	search := x.(*searchByName)
	Central.Log.Debugf("Search value by name %s and index %d:%d, found %s %d/%d %s", search.name, search.peIndex,
		search.muIndex, adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex(), adaValue.Type().Type().name())
	if adaValue.Type().Name() == search.name {
		if search.peIndex == adaValue.PeriodIndex() &&
			search.muIndex == adaValue.MultipleIndex() {
			search.found = adaValue
			Central.Log.Debugf("End traverse, found entry")
			return EndTraverser, nil
		}
		if adaValue.Type().IsStructure() {
			search.grFound = adaValue
		}
	}
	return Continue, nil
}

func traverseSearchValueByNameEnd(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
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
	Central.Log.Debugf("Search field value of %s in definition", fieldName)
	x := &searchByName{name: fieldName}
	fi := strings.Index(fieldName, "[")
	if fi != -1 {
		x.name = fieldName[:fi]
		fi1 := strings.Index(fieldName, "]")
		Central.Log.Debugf("Parse index %s", fieldName[fi+1:fi1])
		index, err := strconv.ParseUint(fieldName[fi+1:fi1], 10, 32)
		if err != nil {
			Central.Log.Debugf("Error parsing search index: %v", err)
			return nil
		}
		x.muIndex = uint32(index)
		Central.Log.Debugf("Parse index %s", fieldName[fi1:])
		rest := fieldName[fi1+1:]
		fi2 := strings.Index(rest, "[")
		if fi2 != -1 {
			fi3 := strings.Index(rest, "]")
			Central.Log.Debugf("Parse index %s %d", rest[fi2+1:], fi3)
			index2, err := strconv.ParseUint(rest[fi2+1:fi3], 10, 32)
			if err != nil {
				Central.Log.Debugf("Error parsing search index: %v", err)
				return nil
			}
			x.peIndex = x.muIndex
			x.muIndex = uint32(index2)
		}
	}
	Central.Log.Debugf("Indexless search of %#v", x)
	t := TraverserValuesMethods{EnterFunction: traverseSearchValueByName}
	_, err := def.TraverseValues(t, x)
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
		Central.Log.Debugf("Search type error: %s", fieldName)
		return
	}

	Central.Log.Debugf("Search field %s index: %#v", fieldName, index)
	// Receive main parent
	parent := t
	for parent.GetParent() != nil && parent.GetParent().Name() != "" {
		parent = parent.GetParent()
	}

	// Main group name if period group use other
	Central.Log.Debugf("Main group parent name : %s", parent.Name())
	if parent.Type() == FieldTypePeriodGroup || parent.Type() == FieldTypeMultiplefield {
		var v IAdaValue

		for _, v = range def.Values {
			Central.Log.Debugf("Search: %s -> %s", parent.Name(), v.Type().Name())
			if v.Type().Name() == parent.Name() {
				break
			}
		}
		strv := v.(*StructureValue)
		if len(index) == 0 {
			err = NewGenericError(121)
			return
		}

		Central.Log.Debugf("Structure value : %s[%d,%d]", strv.Type().Name(), strv.PeriodIndex(), strv.MultipleIndex())
		curIndex := index[0]
		muIndex := uint32(0)
		if parent.Type() == FieldTypeMultiplefield {
			if len(index) < 2 {
				err = NewGenericError(121)
				return
			}
			muIndex = index[1]
			curIndex = index[1]
			Central.Log.Debugf("Use MU index %v for field", curIndex)
		} else {
			Central.Log.Debugf("Use PE index %v for field", curIndex)
		}
		element := strv.elementMap[curIndex]
		if element == nil {
			Central.Log.Debugf("Element not found")
			if create {
				Central.Log.Debugf("Create new Element %d", curIndex)
				strv.initMultipleSubValues(index[0], index[0], muIndex, true)
				element = strv.elementMap[curIndex]
			} else {
				err = NewGenericError(122)
				return
			}
		}
		if element != nil {
			Central.Log.Debugf("Check Element : %#v", element)
			for _, v = range element.Values {
				Central.Log.Debugf("Check value : %s[%d,%d]", v.Type().Name(), v.PeriodIndex(), v.MultipleIndex())
				x := searchByName{name: fieldName}
				switch {
				case index == nil:
				case len(index) > 1:
					x.peIndex = index[0]
					x.muIndex = index[1]
				case len(index) > 0:
					if parent.Type() == FieldTypeMultiplefield {
						x.muIndex = index[0]
					} else {
						x.peIndex = index[0]
					}
				default:
				}
				tvm := TraverserValuesMethods{EnterFunction: traverseSearchValueByName, LeaveFunction: traverseSearchValueByNameEnd}
				_, err = strv.Traverse(tvm, &x)
				if err == nil {
					if x.found != nil {
						Central.Log.Debugf("Found value searching %s under %s", x.found.Type().Name(), strv.Type().Name())
						if x.found.Type().Type() == FieldTypeMultiplefield {
							if len(index) < 2 {
								//return nil, NewGenericError(61)
							} else {
								strv := x.found.(*StructureValue)
								element := strv.elementMap[index[1]]
								if element == nil {
									def.DumpValues(false)
									err = NewGenericError(123)
									return
								}
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
							value.setPeriodIndex(index[0])
							// value.setMultipleIndex(index[1])
							err = strv.addValue(value, index[0], index[1])
							Central.Log.Debugf("New MU value index %d:%d -> %d:%d", index[0], index[1], value.PeriodIndex(), value.MultipleIndex())
							return

						}

					}
				}
				Central.Log.Debugf("Not found or error searching: %v", err)
			}
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

// func traverserFindType(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
// 	search := x.(*search)
// 	Central.Log.Debugf("Check search %s:%s search=%s", adaType.Name(), adaType.ShortName(), search.name)
// 	if adaType.Name() == search.name {
// 		search.adaType = adaType
// 		Central.Log.Debugf("Found type ...return error find")
// 		return NewGenericError(126, search.name)
// 	}
// 	return nil
// }

// SearchType search for a type definition in the tree
func (def *Definition) SearchType(fieldName string) (adaType IAdaType, err error) {
	Central.Log.Debugf("Search type %s", fieldName)
	if af, ok := def.fileFields[fieldName]; ok {
		Central.Log.Debugf("Found file field %s -> %s", af.Name(), af.Type().name())
		return af, nil
	}
	if af, ok := def.activeFields[fieldName]; ok {
		Central.Log.Debugf("Found active field %s", af.Type().name())
		return af, nil
	}
	// search := &search{name: fieldName}
	// level := 1
	// t := TraverserMethods{EnterFunction: traverserFindType}
	// if def.activeFieldTree != nil {
	// 	err = def.activeFieldTree.Traverse(t, level+1, search)
	// } else {
	// 	err = def.fileFieldTree.Traverse(t, level+1, search)
	// }
	// if err == nil {
	// 	err = NewGenericError(41, fieldName)
	// 	return
	// }
	// err = nil
	// if search.adaType == nil {
	// 	Central.Log.Debugf("AdaType not found ", fieldName)
	if Central.IsDebugLevel() {
		Central.Log.Debugf("AdaType %s not found in file fields %#v %#v", fieldName, def.fileFields, def.fileFieldTree)
		for k, v := range def.fileFields {
			Central.Log.Debugf("%s:%s->%s", k, v.ShortName(), v.Name())
		}
		Central.Log.Debugf("AdaType not found in active fields")
		for k, v := range def.activeFields {
			Central.Log.Debugf("%s:%s->%s", k, v.ShortName(), v.Name())
		}
	}
	err = NewGenericError(42, fieldName)
	return
	// }
	// Central.Log.Debugf("Found adaType for search field %s -> %s", fieldName, search.adaType)
	// adaType = search.adaType
	// return
}
