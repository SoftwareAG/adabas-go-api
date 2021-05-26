/*
* Copyright © 2018-2020 Software AG, Darmstadt, Germany and/or its licensors
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
	"reflect"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

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
	cti := createTypeInterface{fields: make([]reflect.StructField, 0), fieldNames: make(map[string][]string)}
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
