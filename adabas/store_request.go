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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// StoreRequest request instance handling data store and update
type StoreRequest struct {
	commonRequest
}

// NewStoreRequest create a new store Request instance
func NewStoreRequest(param ...interface{}) (*StoreRequest, error) {
	if len(param) == 0 {
		return nil, adatypes.NewGenericError(78)
	}
	switch param[0].(type) {
	case string:
		if len(param) > 1 {
			url := param[0].(string)
			switch param[1].(type) {
			case *Adabas:
				ada := param[1].(*Adabas)
				repo := param[2].(*Repository)
				adaMap, err := repo.SearchMapInRepository(ada, url)
				if err != nil {
					return nil, err
				}
				dataRepository := &Repository{DatabaseURL: *adaMap.Data}
				request := &StoreRequest{commonRequest: commonRequest{MapName: url,
					adabas: ada, adabasMap: adaMap, repository: dataRepository}}
				return request, nil
			default:
				fnr, err := evaluateFnr(param[1])
				if err != nil {
					return nil, err
				}
				var adabas *Adabas
				if dbid, aerr := strconv.Atoi(url); aerr == nil {
					adabas, err = NewAdabas(Dbid(dbid))
					if err != nil {
						return nil, err
					}
				} else {
					return nil, aerr
				}
				return &StoreRequest{commonRequest: commonRequest{adabas: adabas,
					repository: &Repository{DatabaseURL: DatabaseURL{Fnr: Fnr(fnr)}}}}, nil
			}
		}
	default:
		ti := reflect.TypeOf(param[0])
		adatypes.Central.Log.Debugf("It's a struct %s", ti.Name())
		if ti.Kind() == reflect.Ptr {
			ti = ti.Elem()
		}
		if ti.Kind() == reflect.Struct {
			adatypes.Central.Log.Debugf("It's a struct %s", ti.Name())
			mapName := ti.Name()
			if len(param) < 2 {
				return nil, errors.New("Not enough parameters for NewReadRequest")
			}
			var request *StoreRequest
			ada := param[1].(*Adabas)
			if len(param) == 2 {
				adabasMap, _, err := SearchMapRepository(ada, mapName)
				if err != nil {
					return nil, err
				}
				dataRepository := &Repository{DatabaseURL: *adabasMap.Data}
				request = &StoreRequest{commonRequest: commonRequest{MapName: mapName,
					adabas:    ada,
					adabasMap: adabasMap, repository: dataRepository}}
			} else {
				rep := param[2].(*Repository)
				var adabasMap *Map
				var err error
				if rep == nil {
					adabasMap, _, err = SearchMapRepository(ada, mapName)
					if err != nil {
						return nil, err
					}
				} else {
					adabasMap, err = rep.SearchMap(ada, mapName)
					if err != nil {
						return nil, err
					}
				}
				dataRepository := &Repository{DatabaseURL: *adabasMap.Data}
				request = &StoreRequest{commonRequest: commonRequest{MapName: mapName,
					adabas:    ada,
					adabasMap: adabasMap, repository: dataRepository}}
			}
			request.createDynamic(param[0])
			return request, nil
		}
		adatypes.Central.Log.Debugf("Unknown kind: %s", reflect.TypeOf(param[0]).Kind())
	}

	return nil, adatypes.NewGenericError(79)
}

// NewStoreRequestAdabas create a new Request instance
func NewStoreRequestAdabas(adabas *Adabas, fnr Fnr) *StoreRequest {
	clonedAdabas := NewClonedAdabas(adabas)
	return &StoreRequest{commonRequest: commonRequest{adabas: clonedAdabas,
		repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// NewAdabasMapNameStoreRequest create new map name store request
func NewAdabasMapNameStoreRequest(adabas *Adabas, adabasMap *Map) (request *StoreRequest, err error) {
	clonedAdabas := NewClonedAdabas(adabas)
	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &StoreRequest{commonRequest: commonRequest{MapName: adabasMap.Name,
		adabas:    clonedAdabas,
		adabasMap: adabasMap, repository: dataRepository}}
	return
}

func evaluateFnr(p interface{}) (Fnr, error) {
	switch p.(type) {
	case int:
		i := p.(int)
		return Fnr(i), nil
	case int32:
		i := p.(int32)
		return Fnr(i), nil
	case int64:
		i := p.(int64)
		return Fnr(i), nil
	case Fnr:
		return p.(Fnr), nil
	default:
	}
	return 0, fmt.Errorf("Cannot evaluate Fnr")
}

// Open Open the Adabas session
func (request *StoreRequest) Open() (err error) {
	_, err = request.commonOpen()
	return
}

func (request *StoreRequest) prepareRequest() (adabasRequest *adatypes.Request, err error) {
	adabasRequest, err = request.definition.CreateAdabasRequest(true, 0, request.adabas.status.platform.IsMainframe())
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Store FB: %s", adabasRequest.FormatBuffer.String())
	adabasRequest.Definition = request.definition
	return
}

func (request *StoreRequest) prepareSecondRequest(secondCall uint8) (adabasRequest *adatypes.Request, err error) {
	adabasRequest, err = request.definition.CreateAdabasRequest(true, secondCall, request.adabas.status.platform.IsMainframe())
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Second store FB: %s", adabasRequest.FormatBuffer.String())
	adabasRequest.Definition = request.definition
	return
}

// StoreFields create record field definition for the next store
func (request *StoreRequest) StoreFields(param ...interface{}) (err error) {
	if len(param) == 0 {
		return adatypes.NewGenericError(0)
	}
	err = request.Open()
	if err != nil {
		return
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Check store fields Definition values %#v", request.definition.Values)
		adatypes.Central.Log.Debugf("Dump all fields")
		request.definition.DumpTypes(true, true)
	}
	switch f := param[0].(type) {
	case string:
		adatypes.Central.Log.Debugf("Store restrict fields to %s", f)
		err = request.definition.ShouldRestrictToFields(f)
		if err != nil {
			return
		}
	case []string:
		adatypes.Central.Log.Debugf("Store restrict fields to %v", f)
		err = request.definition.ShouldRestrictToFieldSlice(f)
		if err != nil {
			return
		}
	}
	if request.dynamic != nil {
		// If interface used, check after restriction if the corresponding fields are part of the
		// query. Remove field names which are not part.
		for k := range request.dynamic.FieldNames {
			if request.definition.Search(k) == nil {
				delete(request.dynamic.FieldNames, k)
			}
		}
	}
	if adatypes.Central.IsDebugLevel() {
		request.definition.DumpTypes(true, true)
		adatypes.Central.Log.Debugf("Definition values %#v", request.definition.Values)
	}
	return
}

// CreateRecord create a record for a special store request
func (request *StoreRequest) CreateRecord() (record *Record, err error) {
	err = request.definition.CreateValues(true)
	if err != nil {
		adatypes.Central.Log.Debugf("Error creating values %v\n", err)
		return
	}
	adatypes.Central.Log.Debugf("Create record Definitons %#v\n", request.definition)
	record, xerr := NewRecord(request.definition)
	if xerr != nil {
		adatypes.Central.Log.Debugf("Error creating record %v\n", xerr)
		err = adatypes.NewGenericError(27, xerr.Error())
	}
	return
}

// Store store a record
func (request *StoreRequest) Store(storeRecord *Record) error {
	request.definition.Values = storeRecord.Value
	adatypes.Central.Log.Debugf("Prepare store request")
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		return prepareErr
	}
	adatypes.Central.Log.Debugf("Prepared store request done need second = %v", adabasRequest.Option.NeedSecondCall)
	//	storeRecord
	helper := adatypes.NewDynamicHelper(Endian())
	err := storeRecord.createRecordBuffer(helper, adabasRequest.Option)
	if err != nil {
		return err
	}
	adatypes.Central.Log.Debugf("Create store request done need second = %v", adabasRequest.Option.NeedSecondCall)

	adabasRequest.RecordBuffer = helper
	err = request.adabas.Store(request.repository.Fnr, adabasRequest)
	if err != nil {
		return err
	}
	storeRecord.Isn = adabasRequest.Isn
	// Reset values after storage to reset for next store request
	adatypes.Central.Log.Debugf("After store request done need second = %v", adabasRequest.Option.NeedSecondCall)
	needSecondCall := adabasRequest.Option.NeedSecondCall
	for needSecondCall != adatypes.NoneSecond {
		adabasRequest.Option.SecondCall++
		adabasRequest, prepareErr := request.prepareSecondRequest(adabasRequest.Option.SecondCall)
		if prepareErr != nil {
			return prepareErr
		}
		adabasRequest.Isn = storeRecord.Isn
		adatypes.Central.Log.Debugf("Prepared update request done need second = %v", adabasRequest.Option.NeedSecondCall)
		helper := adatypes.NewDynamicHelper(Endian())
		err := storeRecord.createRecordBuffer(helper, adabasRequest.Option)
		if err != nil {
			return err
		}
		adabasRequest.RecordBuffer = helper
		err = request.adabas.Update(request.repository.Fnr, adabasRequest)
		if err != nil {
			return err
		}
		needSecondCall = adabasRequest.Option.NeedSecondCall
		adatypes.Central.Log.Debugf("After update request done need second = %v", adabasRequest.Option.NeedSecondCall)
	}
	request.definition.Values = nil
	return err
}

// Update update a record
func (request *StoreRequest) Update(storeRecord *Record) error {
	request.definition.Values = storeRecord.Value
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		return prepareErr
	}
	return request.update(adabasRequest, storeRecord)
}

// Exchange exchange a record
func (request *StoreRequest) Exchange(storeRecord *Record) error {
	request.definition.Values = storeRecord.Value
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		return prepareErr
	}
	return request.update(adabasRequest, storeRecord)
}

// update update a record
func (request *StoreRequest) update(adabasRequest *adatypes.Request, storeRecord *Record) error {
	//	storeRecord
	helper := adatypes.NewDynamicHelper(Endian())
	err := storeRecord.createRecordBuffer(helper, adabasRequest.Option)
	if err != nil {
		return err
	}

	adabasRequest.RecordBuffer = helper
	adabasRequest.Isn = storeRecord.Isn
	err = request.adabas.Update(request.repository.Fnr, adabasRequest)
	// Reset values after storage to reset for next store request
	request.definition.Values = nil
	return err
}

// EndTransaction end of Adabas database transaction
func (request *StoreRequest) EndTransaction() error {
	return request.adabas.EndTransaction()
}

func searchDynamicValue(value reflect.Value, fn []string) (v reflect.Value, ok bool) {
	adatypes.Central.Log.Debugf("Search dynamic interface value %v %d", fn, len(fn))
	v = value
	ok = false
	for _, f := range fn {
		adatypes.Central.Log.Debugf("FieldName search %s", f)
		v = v.FieldByName(f)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		case reflect.Slice:
			return v, true
		}
		adatypes.Central.Log.Debugf("New value %v kind=%s", v, v.Kind())
		ok = v.IsValid()
	}
	return v, ok
}

func (request *StoreRequest) storeValue(record reflect.Value, store bool) error {
	if request.definition == nil {
		q := request.dynamic.CreateQueryFields()
		request.StoreFields(q)
	}

	if record.Kind() == reflect.Ptr {
		record = record.Elem()
	}
	storeRecord, serr := request.CreateRecord()
	if serr != nil {
		return serr
	}
	if adatypes.Central.IsDebugLevel() {
		for k, v := range request.dynamic.FieldNames {
			adatypes.Central.Log.Debugf("FN: %s=%v\n", k, v)
		}
		adatypes.Central.Log.Debugf("Slice index: %v", record)
		request.definition.DumpTypes(true, true, "Active store entries")
	}
	for an, fn := range request.dynamic.FieldNames {
		if !strings.HasPrefix(an, "#") {
			v, ok := searchDynamicValue(record, fn)
			if ok { //&& v.IsValid() {
				if adatypes.Central.IsDebugLevel() {
					adatypes.Central.Log.Debugf("Set dynamic value %v = %v", an, v.Interface())
				}
				err := storeRecord.SetValue(an, v.Interface())
				if err != nil {
					return adatypes.NewGenericError(52, err.Error())
				}
				if adatypes.Central.IsDebugLevel() {
					adatypes.Central.Log.Debugf("Set value %s: %s = %v", an, fn, v)
				}
			}
		}
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("store/update=%v record ADA: %s", store, storeRecord.String())
	}
	storeRecord.Isn = request.dynamic.ExtractIsnField(record)
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("store/update to ISN=%d", storeRecord.Isn)
	}
	if store {
		err := request.Store(storeRecord)
		if err != nil {
			return adatypes.NewGenericError(53, err.Error())
		}
		request.dynamic.PutIsnField(record, storeRecord.Isn)
	} else {
		if storeRecord.Isn == 0 {
			err := request.evaluateKeyIsn(record, storeRecord)
			if err != nil {
				return err
			}
		}

		if storeRecord.Isn == 0 {
			return adatypes.NewGenericError(97)
		}
		err := request.Update(storeRecord)
		if err != nil {
			return adatypes.NewGenericError(133, err.Error())
		}

	}
	return nil
}

func (request *StoreRequest) evaluateKeyIsn(record reflect.Value, storeRecord *Record) error {
	var sn string
	var keyValue string
	if k, ok := request.dynamic.FieldNames["#key"]; ok {
		adatypes.Central.Log.Debugf("Found Keyfield in hash: %v", k)
		kf := request.dynamic.FieldNames[k[0]]
		keyField := record.FieldByName(kf[0])
		adatypes.Central.Log.Debugf("Check keyfield %v %v %s", k, keyField.IsValid(), record.String())
		if keyField.IsValid() {
			if adatypes.Central.IsDebugLevel() {
				adatypes.Central.Log.Debugf("Keyfields: %v %v %s", k, keyField, keyField.String())
				for k, v := range request.dynamic.FieldNames {
					adatypes.Central.Log.Debugf("%s=%s", k, v)
				}
			}
			sn = k[0]
			keyValue = keyField.String()
		}
	} else {
		adatypes.Central.Log.Debugf("Don't found Keyfield")
		t := record.Type()
		var st reflect.Type
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("adabas")
			if tag != "" {
				s := strings.Split(tag, ":")
				if len(s) > 1 && strings.ToLower(s[1]) == "key" {
					st = t.Field(i).Type
					if s[0] != "" {
						sn = s[0]
					} else {
						sn = t.Field(i).Name
					}
					break
				}
			}
		}
		if st == nil {
			return adatypes.NewGenericError(97)
		}
	}
	iRequest, iErr := NewReadRequest(request)
	if iErr != nil {
		return iErr
	}
	iErr = iRequest.QueryFields("")
	if iErr != nil {
		return iErr
	}
	if keyValue == "" {
		adatypes.Central.Log.Debugf("Query temporary read ok %s", sn)
		if adaValue, ok := storeRecord.searchValue(sn); ok {
			adatypes.Central.Log.Debugf("Search key %s='%s'\n", sn, adaValue.String())
			keyValue = adaValue.String()
		} else {
			return adatypes.NewGenericError(96, sn)
		}
	}
	adatypes.Central.Log.Debugf("Read logical ISN with %s=%s", sn, keyValue)
	resultRead, rErr := iRequest.ReadLogicalWith(sn + "=" + keyValue)
	if rErr != nil {
		return rErr
	}
	if len(resultRead.Values) != 1 {
		return adatypes.NewGenericError(98, sn)
	}
	storeRecord.Isn = resultRead.Values[0].Isn
	adatypes.Central.Log.Debugf("Update ISN", storeRecord.Isn)
	return nil
}

// StoreData store interface data, either struct or array
func (request *StoreRequest) StoreData(data interface{}) error {
	return request.modifyData(data, true)
}

// UpdateData update interface data, either struct or array
func (request *StoreRequest) UpdateData(data interface{}) error {
	return request.modifyData(data, false)
}

// StoreData store interface data, either struct or array
func (request *StoreRequest) modifyData(data interface{}, store bool) error {
	adatypes.Central.Log.Debugf("Store type = %T %v", data, reflect.TypeOf(data).Kind())
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		adatypes.Central.Log.Debugf("Work on slice")
		s := reflect.ValueOf(data)
		if s.Len() == 0 {
			return adatypes.NewGenericError(54)
		}
		if request.dynamic == nil {
			request.createDynamic(s.Index(0))
		}

		for si := 0; si < s.Len(); si++ {
			err := request.storeValue(s.Index(si), store)
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if request.dynamic == nil {
			request.createDynamic(data)
		}
		ti := reflect.ValueOf(data).Elem()
		err := request.storeValue(ti, store)
		if err != nil {
			return err
		}
	case reflect.Struct:
		if request.dynamic == nil {
			request.createDynamic(data)
		}
		adatypes.Central.Log.Debugf("Type data %T", data)
		ti := reflect.ValueOf(data)
		err := request.storeValue(ti, store)
		if err != nil {
			return err
		}
	default:
		adatypes.Central.Log.Debugf("Unkown type %v", reflect.TypeOf(data).Kind())
		return adatypes.NewGenericError(0)
	}
	return nil
}
