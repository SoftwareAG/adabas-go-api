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
	"bytes"
	"reflect"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type structure struct {
	structType reflect.Type
	entries    []interface{}
}

// parseMap Adabas read parser of one Map definition used during read
func structParser(adabasRequest *adatypes.Request, x interface{}) error {
	structResult := x.(*structure)
	adatypes.Central.Log.Debugf("Got ISN: %d", adabasRequest.Isn)
	e := reflect.New(structResult.structType)
	s := e.Elem()
	for i := 0; i < structResult.structType.NumField(); i++ {
		fieldName := structResult.structType.Field(i).Tag.Get("adabas")
		if fieldName == "" {
			fieldName = structResult.structType.Field(i).Name
		}
		adatypes.Central.Log.Debugf("Work on %d %v:%s", i, s.Field(i), fieldName)
		lfn := strings.ToLower(fieldName)
		if lfn == "#isn" {
			s.Field(i).SetUint(uint64(adabasRequest.Isn))
		} else {
			v, err := adabasRequest.GetValue(fieldName)
			if err != nil {
				return err
			}

			if v != nil {
				err = adatypes.SetValueData(s.Field(i), v)
				if err != nil {
					return err
				}
			}
		}

	}
	adatypes.Central.Log.Debugf("Ready parsing ISN: %d", adabasRequest.Isn)
	structResult.entries = append(structResult.entries, s.Addr().Interface())
	return nil
}

// ReflectSearch search in map using a interface type given and a search query
func (connection *Connection) ReflectSearch(mapName string, t reflect.Type, search string) ([]interface{}, error) {
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Structured call, %s - %d", t.Name(), t.NumField())
	}
	var buffer bytes.Buffer
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			buffer.WriteRune(',')
		}
		ft := t.Field(i)
		tag := ft.Tag.Get("adabas")
		doAdd := true
		if tag != "" {
			tags := strings.Split(tag, ":")
			for _, t := range tags {
				lt := strings.ToLower(t)
				lt = strings.Trim(lt, " ")
				if lt != "" && lt[0] == '#' {
					doAdd = false
				}
			}
		}
		if doAdd {
			if debug {
				adatypes.Central.Log.Debugf("Add to query %s", t.Field(i).Name)
			}
			buffer.WriteString(ft.Name)
		}
	}

	if debug {
		adatypes.Central.Log.Debugf("Add connection with: %s", buffer.String())
	}

	request, err := connection.CreateMapReadRequest(mapName)
	if err != nil {
		return nil, err
	}

	if qErr := request.QueryFields(buffer.String()); qErr != nil {
		return nil, qErr
	}
	structResult := &structure{structType: t}
	if debug {
		adatypes.Central.Log.Debugf("Read logical with search=%s", search)
	}
	if err = request.ReadLogicalWithWithParser(search, structParser, structResult); err != nil {
		return nil, err
	}
	if debug {
		adatypes.Central.Log.Debugf("Return result entries")
	}
	return structResult.entries, nil
}

// ReflectStore use reflect map to store data with a dynamic interface and Adabas Map name
func (connection *Connection) ReflectStore(entries interface{}, mapName string) error {
	debug := adatypes.Central.IsDebugLevel()
	t := reflect.TypeOf(entries)
	if debug {
		adatypes.Central.Log.Debugf("Store type = %s", t.String())
	}
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
		var isnsNames []string
		fieldNames := make(map[string]string)
		for fi := 0; fi < ri.NumField(); fi++ {
			if fi > 0 {
				buffer.WriteRune(',')
			}
			fieldName := ri.Field(fi).Name
			adabasFieldName := fieldName
			tag := ri.Field(fi).Tag.Get("adabas")
			if debug {
				adatypes.Central.Log.Debugf("Adabas tag: %s", tag)
			}
			if strings.HasPrefix(tag, "#") {
				switch strings.ToLower(tag) {
				case "#isn":
					isnsNames = append(isnsNames, fieldName)
				default:
				}
			} else {
				if tag != "" {
					s := strings.Split(tag, ":")
					adabasFieldName = s[0]
				}
				if debug {
					adatypes.Central.Log.Debugf("Hash field %s=%s", adabasFieldName, fieldName)
				}
				fieldNames[adabasFieldName] = fieldName
				buffer.WriteString(ri.Field(fi).Name)
			}
		}
		ferr := storeRequest.StoreFields(buffer.String())
		if ferr != nil {
			return ferr
		}
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
			if debug {
				adatypes.Central.Log.Debugf("Index: %v %v %v", index, ti, ri)
			}
			for an, fn := range fieldNames {
				if !strings.HasPrefix(an, "#") {
					v := record.FieldByName(fn)
					err = storeRecord.SetValue(an, v.Interface())
					if err != nil {
						return adatypes.NewGenericError(52, err.Error())
					}
					if debug {
						adatypes.Central.Log.Debugf("%s: %s = %v", an, fn, "=", v)
					}
				}
			}
			if debug {
				adatypes.Central.Log.Debugf("Reflect store record: %s", storeRecord.String())
			}
			err = storeRequest.Store(storeRecord)
			if err != nil {
				return adatypes.NewGenericError(53, err.Error())
			}
			if len(isnsNames) > 0 {
				for _, isnName := range isnsNames {
					record.FieldByName(isnName).SetUint(uint64(storeRecord.Isn))
				}
			}
		}
	default:
		adatypes.Central.Log.Errorf("Unknown reflect store type %v", reflect.TypeOf(entries).Kind())
	}
	return nil

}
