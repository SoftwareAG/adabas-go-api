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
	"bytes"
	"reflect"
	"strings"
)

// DynamicInterface dynamic interface
type DynamicInterface struct {
	DataType   reflect.Type
	FieldNames map[string][]string
}

func generateFieldNames(ri reflect.Type, f map[string][]string, fields []string) {
	Central.Log.Debugf("Generate field names for %s %s", ri.Name(), ri.Kind())
	if ri.Kind() != reflect.Struct {
		return
	}
	for fi := 0; fi < ri.NumField(); fi++ {
		ct := ri.Field(fi)
		fieldName := ct.Name
		adabasFieldName := fieldName
		tag := ct.Tag.Get("adabas")
		Central.Log.Debugf("fieldName=%s/%s -> tag=%s", adabasFieldName, fieldName, tag)
		if tag != "" {
			s := strings.Split(tag, ":")
			if len(s) > 2 {
				adabasFieldName = s[2]
			} else {
				if s[0] != "" {
					adabasFieldName = s[0]
					if strings.ToLower(adabasFieldName) == "#isn" {
						adabasFieldName = "#isn"
					}
				}
			}

			if len(s) > 1 {
				switch s[1] {
				case "key":
					//fmt.Println(fieldName, adabasFieldName)
					f["#key"] = []string{adabasFieldName}
				case "isn":
					f["#isn"] = []string{adabasFieldName}
					// No sub value and not relevant Adabas field, skip rest
					continue
				case "ignore":
					continue
				case "":
					// this is if the inmap repository-less map is used
				default:
					Central.Log.Errorf("Unknown control tag >%s<", s[1])
					continue
				}
			}
		}
		subFields := make([]string, len(fields))
		copy(subFields, fields)
		subFields = append(subFields, fieldName)
		Central.Log.Debugf("Set field names to %s -> %v", adabasFieldName, subFields)
		f[adabasFieldName] = subFields
		Central.Log.Debugf("Type struct field = %v", ct.Type.Kind())
		switch ct.Type.Kind() {
		case reflect.Ptr:
			Central.Log.Debugf("Pointer found %v %v", ct.Type.Name(), ct.Type.Elem().Name())
			//et := reflect.TypeOf(ct.Type.Elem())
			generateFieldNames(ct.Type.Elem(), f, subFields)
		case reflect.Slice:
			Central.Log.Debugf("Slice found %v %v", ct.Type.Name(), ct.Type.Elem().Name())
			sliceT := ct.Type.Elem()
			if sliceT.Kind() == reflect.Ptr {
				sliceT = sliceT.Elem()
			}
			generateFieldNames(sliceT, f, subFields)
		}
	}

}

// CreateDynamicInterface constructor create dynamic interface
func CreateDynamicInterface(i interface{}) *DynamicInterface {
	ri := reflect.TypeOf(i)
	if ri.Kind() == reflect.Ptr {
		ri = ri.Elem()
	}
	dynamic := &DynamicInterface{DataType: ri, FieldNames: make(map[string][]string)}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Dynamic interface %s", ri.Name())
		Central.Log.Debugf("Dynamic interface %v nrFields=%d", ri, ri.NumField())
	}
	generateFieldNames(ri, dynamic.FieldNames, make([]string, 0))
	return dynamic
}

// CreateQueryFields create query field list of dynamic interface given
func (dynamic *DynamicInterface) CreateQueryFields() string {
	var buffer bytes.Buffer
	for fieldName := range dynamic.FieldNames {
		if buffer.Len() > 0 {
			buffer.WriteRune(',')
		}
		buffer.WriteString(fieldName)
	}
	Central.Log.Debugf("Create query fields: %s", buffer.String())

	return buffer.String()
}

// ExamineIsnField set the interface Isn-tagged field with value for ISN
func (dynamic *DynamicInterface) ExamineIsnField(value reflect.Value, isn Isn) error {
	Central.Log.Debugf("Examine ISN field: %d", isn)
	v := value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if f, ok := dynamic.FieldNames["#isn"]; ok {
		isnField := v.FieldByName(f[0])
		if !isnField.IsValid() || isnField.Kind() != reflect.Uint64 {
			return NewGenericError(113)
		}
		Central.Log.Debugf("Found isn %d", isn)
		isnField.SetUint(uint64(isn))
	} else {
		Central.Log.Debugf("No ISN field found")

	}
	return nil
}

// ExtractIsnField extract out of interface Isn-tagged field with value for ISN
func (dynamic *DynamicInterface) ExtractIsnField(value reflect.Value) Isn {
	v := value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if k, ok := dynamic.FieldNames["#isn"]; ok {
		Central.Log.Debugf("ISNfield: %v", k)
		keyField := v.FieldByName(k[0])
		return Isn(keyField.Uint())
	}
	return 0
}

// PutIsnField put ISN field back into structure
func (dynamic *DynamicInterface) PutIsnField(value reflect.Value, isn Isn) {
	v := value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	Central.Log.Debugf("Check FieldNames %v", dynamic.FieldNames)
	if k, ok := dynamic.FieldNames["#isn"]; ok {
		Central.Log.Debugf("Set ISN field: %s", k)
		for _, kisn := range k {
			iv := v.FieldByName(kisn)
			if iv.Kind() == reflect.Ptr {
				iv = iv.Elem()
			}
			if iv.CanAddr() {
				Central.Log.Debugf("Set ISN for %s to %d", kisn, isn)
				iv.SetUint(uint64(isn))
			} else {
				Central.Log.Debugf("Cannot address ISN: %s", kisn)
			}
		}
	}
}
