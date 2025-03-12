/*
* Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"reflect"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// createTypeInterface structure used to traverse through the field tree and create list of
// dynamic interface fields types
type traverseCreateDynamicTypeFields struct {
	fields     []reflect.StructField
	fieldHash  map[string]bool
	fieldNames map[string][]string
}

// traverseCreateTypeInterface traverser method create type interface
func traverseCreateTypeInterface(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	cti := x.(*traverseCreateDynamicTypeFields)
	name := strings.ReplaceAll(adaType.Name(), "-", "")
	adatypes.Central.Log.Debugf("Add field %s/%s of %v", adaType.Name(), name, cti.fieldHash[name])
	cti.fieldNames[name] = []string{adaType.Name()}
	switch adaType.Type() {
	case adatypes.FieldTypeString, adatypes.FieldTypeUnicode,
		adatypes.FieldTypeLAString, adatypes.FieldTypeLBString,
		adatypes.FieldTypeLAUnicode, adatypes.FieldTypeLBUnicode:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(string(""))})
	case adatypes.FieldTypePacked, adatypes.FieldTypeUnpacked:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(int64(0))})
	case adatypes.FieldTypeByte, adatypes.FieldTypeInt2, adatypes.FieldTypeInt4:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(int32(0))})
	case adatypes.FieldTypeInt8:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(int64(0))})
	case adatypes.FieldTypeUByte, adatypes.FieldTypeUInt2, adatypes.FieldTypeUInt4:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(uint32(0))})
	case adatypes.FieldTypeUInt8:
		cti.fields = append(cti.fields, reflect.StructField{Name: name,
			Type: reflect.TypeOf(uint64(0))})
	case adatypes.FieldTypeMultiplefield:
	case adatypes.FieldTypeGroup, adatypes.FieldTypePeriodGroup, adatypes.FieldTypeStructure:
	case adatypes.FieldTypeSuperDesc, adatypes.FieldTypeHyperDesc, adatypes.FieldTypeCollation:
	case adatypes.FieldTypePhonetic, adatypes.FieldTypeReferential:
	default:
		fmt.Println("Field Type", name, adaType.Type())
		return adatypes.NewGenericError(175, fmt.Sprintf("%T", adaType.Type()))
	}
	return nil
}

// createDynamic create the dynamic interface needed for usage of dynamic
func (request *commonRequest) createDynamic(i interface{}) {
	request.dynamic = adatypes.CreateDynamicInterface(i)
}

// createInterface create a interface of the type
func (request *ReadRequest) createInterface(fieldList string) (err error) {
	err = request.loadDefinition()
	if err != nil {
		return
	}
	err = request.QueryFields(fieldList)
	if err != nil {
		return
	}

	var structType reflect.Type
	tm := adatypes.NewTraverserMethods(traverseCreateTypeInterface)
	cti := traverseCreateDynamicTypeFields{fields: make([]reflect.StructField, 0), fieldNames: make(map[string][]string)}
	isnField := reflect.StructField{Name: "ISN",
		Type: reflect.TypeOf(uint64(0))}
	cti.fields = append(cti.fields, isnField)
	err = request.definition.TraverseTypes(tm, true, &cti)
	if err != nil {
		return err
	}
	structType = reflect.StructOf(cti.fields)
	dynamic := &adatypes.DynamicInterface{DataType: structType, FieldNames: cti.fieldNames}
	dynamic.FieldNames["#isn"] = []string{"ISN"}
	adatypes.Central.Log.Debugf("Create final field names map: %v", dynamic.FieldNames)
	request.dynamic = dynamic
	return nil
}
