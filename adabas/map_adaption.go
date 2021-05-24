/*
* Copyright © 2021 Software AG, Darmstadt, Germany and/or its licensors
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

package adabas

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type mapAdaption struct {
	adabasMap  *Map
	definition *adatypes.Definition
}

// traverseAdaptType called per field of the Map and adapt fields the type of a Map to the correct type
// in case a DDM type is used
func traverseAdaptType(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	mapAdapt := x.(*mapAdaption)
	adabasMap := mapAdapt.adabasMap
	sn := adaType.ShortName()[0:2]
	adatypes.Central.Log.Debugf("Adapt type %s/%s", adaType.Name(), sn)
	f := adabasMap.fieldMap[sn]
	if f == nil {
		adaType.AddFlag(adatypes.FlagOptionToBeRemoved)
		adatypes.Central.Log.Debugf("Field %s flag to be removed", adaType.Name())
		return nil
	}
	adatypes.Central.Log.Debugf("Field map does contains %s -> %s/%s format type=>%s<", f.LongName, adaType.Name(), adaType.ShortName(), f.FormatType)
	adaType.RemoveFlag(adatypes.FlagOptionToBeRemoved)
	mapAdapt.definition.AdaptName(adaType, f.LongName)
	adaType.SetName(f.LongName)

	switch strings.Trim(f.FormatType, " ") {
	case "#":
		adatypes.Central.Log.Debugf("Replace %s on parent %s", adaType.ShortName(), parentType.ShortName())
		nt := adatypes.NewRedefinitionType(adaType)
		adaptRedefintionFields(nt, adabasMap.redefinitionFieldMap[adaType.ShortName()])
		st := parentType.(*adatypes.StructureType)
		err := st.ReplaceType(adaType, nt)
		if err != nil {
			return err
		}
		return nil
	case "":
	default:
		adaType.SetFormatType([]rune(f.FormatType)[0])
	}
	adaType.SetFormatLength(uint32(f.Length))
	if !adaType.IsStructure() {
		if f.Length < 0 {
			adatypes.Central.Log.Debugf("Set %s length to 0 formerly was %d or %d", adaType.Name(), adaType.Length(), f.Length)
			adaType.SetLength(0)
		} else {
			adatypes.Central.Log.Debugf("Set %s length to %d, check content type=%s",
				adaType.Name(), f.Length, f.ContentType)
			adaType.SetLength(uint32(f.Length))
		}
		ct := strings.Split(f.ContentType, ",")
		for _, c := range ct {
			p := strings.Split(c, "=")
			if len(p) > 1 {
				adatypes.Central.Log.Debugf("%s=%s", p[0], p[1])
				s := strings.ToLower(p[0])
				switch s {
				case "fractionalshift":
					fs, ferr := strconv.Atoi(p[1])
					if ferr != nil {
						return ferr
					}
					if fs < 0 && fs > math.MaxUint32 {
						return adatypes.NewGenericError(166, fs)
					}
					adaType.SetFractional(uint32(fs))
				case "charset":
					adatypes.Central.Log.Debugf("Set charset to %s", p[1])
					adaType.SetCharset(p[1])
				case "formattype":
					if p[1] != "" {
						adaType.SetFormatType(rune(p[1][0]))
					}
				case "length":
					fs, ferr := strconv.Atoi(p[1])
					if ferr != nil {
						return ferr
					}
					if fs < 0 && fs > math.MaxUint32 {
						return adatypes.NewGenericError(166, fs)
					}
					adaType.SetFormatLength(uint32(fs))
				default:
					fmt.Println("Unknown paramteter", p[0])
				}
			}
		}
	}
	adatypes.Central.Log.Debugf("Set long name %s for %s/%s", f.LongName, adaType.Name(), adaType.ShortName())
	return nil
}

// adaptRedefintionFields read Map entry adapt the redefinition definition during
// Map read process.
func adaptRedefintionFields(redType *adatypes.RedefinitionType, fields []*MapField) {
	adatypes.Central.Log.Debugf("Fields: %#v", fields)
	for _, f := range fields {
		adatypes.Central.Log.Debugf("%s %s %s %d", f.ShortName, f.LongName, f.FormatType, f.Length)
		fieldType := adatypes.EvaluateFieldType([]rune(f.FormatType)[0], f.Length)
		subType := adatypes.NewType(fieldType, f.ShortName, uint32(f.Length))
		subType.SetName(f.LongName)
		adatypes.Central.Log.Debugf("%s:%s %s %d %d", f.LongName, f.ShortName, f.FormatType, f.Length, subType.Length())
		redType.AddSubType(subType)
	}
}

// adaptFieldType base class starting the traverser through the fields to adapt field types.
// The long name is adapted to the field entries in the overall file definition
func (adabasMap *Map) adaptFieldType(definition *adatypes.Definition, dynamic *adatypes.DynamicInterface) (err error) {
	if definition == nil {
		return adatypes.NewGenericError(19)
	}
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "before adapt field types")
		adatypes.Central.Log.Debugf("Adapt map long names to type definition %#v", adabasMap.String())
	}
	adaptMap := &mapAdaption{adabasMap: adabasMap, definition: definition}
	tm := adatypes.NewTraverserMethods(traverseAdaptType)
	err = definition.TraverseTypes(tm, true, adaptMap)
	if err != nil {
		return
	}
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "before restrict slice")
	}
	// Restrict fields to the fields included in the map
	fields := adabasMap.FieldNames()
	if dynamic != nil {
		adatypes.Central.Log.Debugf("Check dynamic adapt fields %v", fields)
		newFields := make([]string, 0)
		for _, f := range fields {
			if _, ok := dynamic.FieldNames[f]; ok {
				adatypes.Central.Log.Debugf("Checked %s", f)
				newFields = append(newFields, f)
			} else {
				adatypes.Central.Log.Debugf("Not Checked %s", f)
			}
		}
		fields = newFields
		adatypes.Central.Log.Debugf("Redefine dynamic fields %v", fields)
	} else {
		adatypes.Central.Log.Debugf("Check non-dynamic adapt fields %v", fields)
		// Subdivide field list to the definition fields
		newFields := make([]string, 0)
		for _, f := range fields {
			if definition.CheckField(f) {
				adatypes.Central.Log.Debugf("Checked %s", f)
				newFields = append(newFields, f)
			} else {
				adatypes.Central.Log.Debugf("Not Checked %s", f)
			}
		}
		fields = newFields
		adatypes.Central.Log.Debugf("Rework adapted fields %v", fields)
	}
	// Restrict to final fields list
	err = definition.RestrictFieldSlice(fields)
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "after restrict slice")
	}
	return
}
