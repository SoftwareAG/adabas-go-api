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
	"fmt"
	"reflect"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type structure struct {
	structType reflect.Type
	entries    []interface{}
}

// parseMap Adabas read parser of one Map definition used during read
func structParser(adabasRequest *adatypes.AdabasRequest, x interface{}) error {
	structResult := x.(*structure)
	fmt.Println("Got ISN:", adabasRequest.Isn)
	e := reflect.New(structResult.structType)
	s := e.Elem()
	for i := 0; i < structResult.structType.NumField(); i++ {
		fieldName := structResult.structType.Field(i).Tag.Get("adabas")
		if fieldName == "" {
			fieldName = structResult.structType.Field(i).Name
		}
		fmt.Println(i, s.Field(i))
		v, err := adabasRequest.GetValue(fieldName)
		if err != nil {
			return err
		}
		if v != nil {
			fmt.Println(fieldName, v, s.Field(i), s.Field(i).Type())
			switch s.Field(i).Interface().(type) {
			case int8, int32, int64:
				vi, err := v.Int64()
				if err != nil {
					return err
				}
				s.Field(i).SetInt(vi)
			case uint8, uint32, uint64:
				vui, err := v.UInt64()
				if err != nil {
					return err
				}
				s.Field(i).SetUint(vui)
			case string:
				s.Field(i).SetString(v.String())
			default:
				return fmt.Errorf("Type %v for %s not supported", s.Field(i).Type(), fieldName)
			}
		}
	}
	structResult.entries = append(structResult.entries, s.Addr().Interface())
	return nil
}

// ReflectSearch search in map using a structure given
func ReflectSearch(mapName string, t reflect.Type, connection *Connection, search string) ([]interface{}, error) {
	adatypes.Central.Log.Debugf("Structured call, %s - %d", t.Name(), t.NumField())
	var buffer bytes.Buffer
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			buffer.WriteRune(',')
		}
		adatypes.Central.Log.Debugf("Add to query %s", t.Field(i).Name)
		buffer.WriteString(t.Field(i).Name)
	}

	adatypes.Central.Log.Debugf("Add connection with: %s", buffer.String())

	request, err := connection.CreateMapReadRequest(mapName)
	if err != nil {
		return nil, err
	}
	if qErr := request.QueryFields(buffer.String()); qErr != nil {
		return nil, qErr
	}
	structResult := &structure{structType: t}
	adatypes.Central.Log.Debugf("Read logical with search=%s", search)
	if err = request.ReadLogicalWithWithParser(search, structParser, structResult); err != nil {
		return nil, err
	}
	adatypes.Central.Log.Debugf("Return result entries")
	return structResult.entries, nil
}

// ReflectStore use reflect map to store data
func ReflectStore(entries interface{}, connection *Connection, mapName string) error {
	t := reflect.TypeOf(entries)
	adatypes.Central.Log.Debugf("Store type = %s", t.String())
	switch reflect.TypeOf(entries).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(entries)
		if s.Len() == 0 {
			return adatypes.NewGenericError(54)
		}
		storeRequest, err := connection.CreateMapStoreRequest(mapName)
		if err != nil {
			return err
		}
		ri := reflect.TypeOf(s.Index(0).Elem().Interface())
		var buffer bytes.Buffer
		fieldNames := make(map[string]string)
		for fi := 0; fi < ri.NumField(); fi++ {
			if fi > 0 {
				buffer.WriteRune(',')
			}
			fieldName := ri.Field(fi).Name
			adabasFieldName := fieldName
			tag := ri.Field(fi).Tag.Get("adabas")
			adatypes.Central.Log.Debugf("X: %s", tag)
			if tag != "" {
				adabasFieldName = tag
			}
			fieldNames[adabasFieldName] = fieldName
			buffer.WriteString(ri.Field(fi).Name)
		}
		storeRequest.StoreFields(buffer.String())
		for si := 0; si < s.Len(); si++ {
			storeRecord, serr := storeRequest.CreateRecord()
			if serr != nil {
				return serr
			}
			record := s.Index(si)
			if record.Kind() == reflect.Ptr {
				record = record.Elem()
			}
			index := s.Index(si)
			ti := reflect.TypeOf(entries).Elem()
			adatypes.Central.Log.Debugf("Index: %v %v %v", index, ti, ri)
			for an, fn := range fieldNames {
				v := record.FieldByName(fn)
				err = storeRecord.SetValue(an, v.Interface())
				if err != nil {
					return adatypes.NewGenericError(52, err.Error())
				}
				adatypes.Central.Log.Debugf("%s: %s = %v", an, fn, "=", v)
			}
			adatypes.Central.Log.Debugf("RECORD ADA: %s", storeRecord.String())
			err = storeRequest.Store(storeRecord)
			if err != nil {
				return adatypes.NewGenericError(53, err.Error())
			}
		}
	default:
		adatypes.Central.Log.Debugf("Unkown type %v", reflect.TypeOf(entries).Kind())
	}
	return nil

}
