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

package adabas

import (
	"bytes"
	"reflect"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type dynamicInterface struct {
	dataType   interface{}
	fieldNames map[string]string
}

func (request *commonRequest) createDynamic(i interface{}) {
	request.dynamic = &dynamicInterface{dataType: i, fieldNames: make(map[string]string)}
	ri := reflect.TypeOf(i)
	adatypes.Central.Log.Debugf("Dynamic interface %v nrFields=%d", ri, ri.NumField())
	for fi := 0; fi < ri.NumField(); fi++ {
		fieldName := ri.Field(fi).Name
		adabasFieldName := fieldName
		tag := ri.Field(fi).Tag.Get("adabas")
		adatypes.Central.Log.Debugf("fieldName=%s/%s -> tag=%s", adabasFieldName, fieldName, tag)
		if tag != "" {
			adabasFieldName = tag
		}
		request.dynamic.fieldNames[adabasFieldName] = fieldName
	}

}

func (dynamic *dynamicInterface) createQueryFields() string {
	var buffer bytes.Buffer
	for fieldName := range dynamic.fieldNames {
		if buffer.Len() > 0 {
			buffer.WriteRune(',')
		}
		buffer.WriteString(fieldName)
	}
	adatypes.Central.Log.Debugf("Create query fields: %s", buffer.String())

	return buffer.String()
}
