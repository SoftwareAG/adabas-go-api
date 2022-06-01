/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"reflect"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// StoreRequest request instance handling data store and update
type StoreRequest struct {
	commonRequest
}

// NewStoreRequest creates a new store Request instance using different
// types of parameters. This is only for internal use. Use the `Connection`
// instance to create store requests.
// This constructor is internal.
func NewStoreRequest(param ...interface{}) (*StoreRequest, error) {
	if len(param) == 0 {
		return nil, adatypes.NewGenericError(78)
	}
	switch p := param[0].(type) {
	case string:
		if len(param) > 1 {
			url := p
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
				// TODO should not be numeric only, anyway use Connection
				if dbid, aerr := strconv.Atoi(url); aerr == nil {
					if dbid < 0 || dbid > 65535 {
						return nil, adatypes.NewGenericError(70, dbid)
					}
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
	case *Adabas:
		m := param[1].(*Map)
		return createNewMapPointerStoreRequest(p, m)
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
				return nil, errors.New("not enough parameters for NewStoreRequest")
			}
			var request *StoreRequest
			ada := param[1].(*Adabas)
			if len(param) == 2 {
				adabasMap, _, err := SearchMapRepository(ada.ID, mapName)
				if err != nil {
					return nil, err
				}
				dataRepository := &Repository{DatabaseURL: *adabasMap.Data}
				ada.SetURL(&adabasMap.Data.URL)
				request = &StoreRequest{commonRequest: commonRequest{MapName: mapName,
					adabas:    ada,
					adabasMap: adabasMap, repository: dataRepository}}
			} else {
				rep := param[2].(*Repository)
				var adabasMap *Map
				var err error
				if rep == nil {
					adabasMap, _, err = SearchMapRepository(ada.ID, mapName)
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
				ada.SetURL(&adabasMap.Data.URL)
				request = &StoreRequest{commonRequest: commonRequest{MapName: mapName,
					adabas:    ada,
					adabasMap: adabasMap, repository: dataRepository}}
			}
			adatypes.Central.Log.Debugf("Data reference %s", request.adabas.URL.String())
			adatypes.Central.Log.Debugf("Map reference %s", request.adabasMap.Repository.URL.String())
			request.createDynamic(param[0])
			return request, nil
		}
		adatypes.Central.Log.Errorf("Unknown kind: %s", reflect.TypeOf(param[0]).Kind())
	}

	return nil, adatypes.NewGenericError(79)
}

// createNewMapPointerReadRequest create a new Request instance
func createNewMapPointerStoreRequest(adabas *Adabas, adabasMap *Map) (request *StoreRequest, err error) {
	if adabasMap == nil {
		err = adatypes.NewGenericError(22, "")
		return
	}
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d -> %#v", adabasMap.Name,
		adabasMap.Data.Fnr, adabas.ID.platform)
	cloneAdabas := NewClonedAdabas(adabas)

	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &StoreRequest{
		commonRequest: commonRequest{MapName: adabasMap.Name, adabas: cloneAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewStoreRequestAdabas creates a new store Request instance using an
// Adabas instance and Adabas file number.
// This is only for internal use. Use the `Connection`
// instance to create store requests.
// This constructor is internal.
func NewStoreRequestAdabas(adabas *Adabas, fnr Fnr) *StoreRequest {
	clonedAdabas := NewClonedAdabas(adabas)
	return &StoreRequest{commonRequest: commonRequest{adabas: clonedAdabas,
		repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// NewAdabasMapNameStoreRequest creates a new store Request instance using an
// Adabas instance and Adabas Map.
// This is only for internal use. Use the `Connection`
// instance to create store requests.
// This constructor is internal.
func NewAdabasMapNameStoreRequest(adabas *Adabas, adabasMap *Map) (request *StoreRequest, err error) {
	clonedAdabas := NewClonedAdabas(adabas)
	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &StoreRequest{commonRequest: commonRequest{MapName: adabasMap.Name,
		adabas:    clonedAdabas,
		adabasMap: adabasMap, repository: dataRepository}}
	return
}

// evaluateFnr evalute Adabas file number
func evaluateFnr(p interface{}) (Fnr, error) {
	switch i := p.(type) {
	case int:
		return Fnr(i), nil
	case int32:
		return Fnr(i), nil
	case int64:
		return Fnr(i), nil
	case Fnr:
		return p.(Fnr), nil
	default:
	}
	return 0, adatypes.NewGenericError(167)
}

// Open Open the Adabas session
func (request *StoreRequest) Open() (err error) {
	_, err = request.commonOpen()
	return
}

// prepareRequest prepare a store request create an adabas request information
// like Format Buffer or Record Buffer
func (request *StoreRequest) prepareRequest(descriptorRead bool) (adabasRequest *adatypes.Request, err error) {
	parameter := &adatypes.AdabasRequestParameter{Store: true, DescriptorRead: descriptorRead, SecondCall: 0, Mainframe: request.adabas.status.platform.IsMainframe()}
	adabasRequest, err = request.definition.CreateAdabasRequest(parameter)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Store FB: %s", adabasRequest.FormatBuffer.String())
	adabasRequest.Definition = request.definition
	return
}

func (request *StoreRequest) prepareSecondRequest(secondCall uint32) (adabasRequest *adatypes.Request, err error) {
	parameter := &adatypes.AdabasRequestParameter{Store: true, DescriptorRead: false,
		SecondCall: secondCall, Mainframe: request.adabas.status.platform.IsMainframe()}
	adabasRequest, err = request.definition.CreateAdabasRequest(parameter)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Second store FB: %s", adabasRequest.FormatBuffer.String())
	adabasRequest.Definition = request.definition
	return
}

// StoreFields defines the fields to be part of the store request.
// This is to prepare the create record for the next store
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
				adatypes.Central.Log.Debugf("Remove dynamic field %s", k)
				if k != "" && k[0] != '#' {
					delete(request.dynamic.FieldNames, k)
				}
			}
		}
	}
	if adatypes.Central.IsDebugLevel() {
		request.definition.DumpTypes(true, true)
		adatypes.Central.Log.Debugf("Definition values %#v", request.definition.Values)
	}
	return
}

// CreateRecord create a record for a special store request. The fields
// which are part of the record are defined using the `StoreFields` method.
func (request *StoreRequest) CreateRecord() (record *Record, err error) {
	err = request.definition.CreateValues(true)
	if err != nil {
		adatypes.Central.Log.Debugf("Error creating values %v", err)
		return
	}
	adatypes.Central.Log.Debugf("Create record Definitons %#v", request.definition)
	record, xerr := NewRecord(request.definition)
	if xerr != nil {
		adatypes.Central.Log.Debugf("Error creating record %v", xerr)
		err = adatypes.NewGenericError(27, xerr.Error())
	}
	return
}

// Store store/insert a given record into database.
// Note: the data is stored, but is not final until the end of
// transaction is done.
func (request *StoreRequest) Store(storeRecord *Record) error {
	if request.definition == nil {
		sErr := request.StoreFields(storeRecord)
		if sErr != nil {
			return sErr
		}
	}
	request.definition.Values = storeRecord.SelectValue(request.definition)
	adatypes.Central.Log.Debugf("Prepare store request")
	adabasRequest, prepareErr := request.prepareRequest(false)
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
	return request.secondStore(adabasRequest, storeRecord)
}

func (request *StoreRequest) secondStore(adabasRequest *adatypes.Request, storeRecord *Record) error {

	needSecondCall := adabasRequest.Option.NeedSecondCall
	for needSecondCall != adatypes.NoneSecond {
		adabasRequest.Option.SecondCall++
		adabasRequest, prepareErr := request.prepareSecondRequest(adabasRequest.Option.SecondCall)
		if prepareErr != nil {
			adatypes.Central.Log.Debugf("Error preparing second call: %v", prepareErr)
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
		if storeRecord.LobEndTransaction {
			err = request.EndTransaction()
			if err != nil {
				return err
			}
		}
	}
	request.definition.Values = nil
	return nil
}

// Update update a given record into database. The given record
// need to include the corresponding Isn information provided by
// a previous read call.
// Note: the data is update in Adabas, but is not final until the end of
// transaction is done. Dirty ready may be done.
func (request *StoreRequest) Update(storeRecord *Record) error {
	if request.definition == nil {
		sErr := request.StoreFields(storeRecord)
		if sErr != nil {
			return sErr
		}
	}
	storeRecord.definition = request.definition
	request.definition.Values = storeRecord.SelectValue(request.definition)
	adabasRequest, prepareErr := request.prepareRequest(false)
	if prepareErr != nil {
		return prepareErr
	}
	err := request.update(adabasRequest, storeRecord)
	if err != nil {
		return err
	}
	adatypes.Central.Log.Debugf("After update request done need second = %v", adabasRequest.Option.NeedSecondCall)
	return request.secondStore(adabasRequest, storeRecord)
}

// Exchange exchange a record
func (request *StoreRequest) Exchange(storeRecord *Record) error {
	request.definition.Values = storeRecord.SelectValue(request.definition)
	adabasRequest, prepareErr := request.prepareRequest(false)
	if prepareErr != nil {
		return prepareErr
	}
	adabasRequest.Option.ExchangeRecord = true
	err := request.update(adabasRequest, storeRecord)
	if err != nil {
		return err
	}
	adatypes.Central.Log.Debugf("After update request done need second = %v", adabasRequest.Option.NeedSecondCall)
	return request.secondStore(adabasRequest, storeRecord)
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
	if adabasRequest.Option.NeedSecondCall == adatypes.NoneSecond {
		request.definition.Values = nil
	}
	return err
}

// EndTransaction call an end of Adabas database transaction
func (request *StoreRequest) EndTransaction() error {
	return request.adabas.EndTransaction()
}

// searchDynamicValue search dynamic value in the interface
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

// storeValue used in dynamic interface mode to store records
func (request *StoreRequest) storeValue(record reflect.Value, store, etData bool) error {
	if request.definition == nil {
		q := request.dynamic.CreateQueryFields()
		err := request.StoreFields(q)
		if err != nil {
			return err
		}
	}

	if record.Kind() == reflect.Ptr {
		record = record.Elem()
	}
	storeRecord, serr := request.CreateRecord()
	if serr != nil {
		return serr
	}
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		for k, v := range request.dynamic.FieldNames {
			adatypes.Central.Log.Debugf("FN: %s=%v", k, v)
		}
		adatypes.Central.Log.Debugf("Slice index: %v", record)
		request.definition.DumpTypes(true, true, "Active store entries")
	}
	if debug {
		adatypes.Central.Log.Debugf("Put request dynamic field %v", request.dynamic.FieldNames)
	}
	for an, fn := range request.dynamic.FieldNames {
		if !strings.HasPrefix(an, "#") && an != "" {
			v, ok := searchDynamicValue(record, fn)
			if ok { //&& v.IsValid() {
				if debug {
					adatypes.Central.Log.Debugf("Set dynamic value %v = %v", an, v.Interface())
				}
				err := storeRecord.SetValue(an, v.Interface())
				if err != nil {
					return adatypes.NewGenericError(52, err.Error())
				}
				if debug {
					adatypes.Central.Log.Debugf("Set value %s: %s = %v", an, fn, v)
				}
			}
		}
	}
	if debug {
		adatypes.Central.Log.Debugf("store/update=%v record ADA: %s", store, storeRecord.String())
	}
	storeRecord.Isn = request.dynamic.ExtractIsnField(record)
	storeRecord.LobEndTransaction = etData
	if debug {
		adatypes.Central.Log.Debugf("store/update to ISN=%d", storeRecord.Isn)
	}
	if store {
		err := request.Store(storeRecord)
		if err != nil {
			return adatypes.NewGenericError(53, err.Error())
		}
		adatypes.Central.Log.Debugf("Stored record to %d", storeRecord.Isn)
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

// evaluateKeyIsn evaluate key or Isn information by keywords #isn or #key
func (request *StoreRequest) evaluateKeyIsn(record reflect.Value, storeRecord *Record) error {
	var sn string
	var keyValue string
	debug := adatypes.Central.IsDebugLevel()
	if k, ok := request.dynamic.FieldNames["#key"]; ok {
		adatypes.Central.Log.Debugf("Found Keyfield in hash: %v", k)
		kf := request.dynamic.FieldNames[k[0]]
		keyField := record.FieldByName(kf[0])
		if debug {
			adatypes.Central.Log.Debugf("Check keyfield %v %v %s", k, keyField.IsValid(), record.String())
		}
		if keyField.IsValid() {
			if debug {
				adatypes.Central.Log.Debugf("Keyfields: %v %v %s", k, keyField, keyField.String())
				for k, v := range request.dynamic.FieldNames {
					adatypes.Central.Log.Debugf("%s=%s", k, v)
				}
			}
			sn = k[0]
			keyValue = keyField.String()
		}
	} else {
		if debug {
			adatypes.Central.Log.Debugf("Don't found Keyfield")
		}
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
		if debug {
			adatypes.Central.Log.Debugf("Query temporary read ok %s", sn)
		}
		if adaValue, ok := storeRecord.searchValue(sn); ok {
			if debug {
				adatypes.Central.Log.Debugf("Search key %s='%s'", sn, adaValue.String())
			}
			keyValue = adaValue.String()
		} else {
			return adatypes.NewGenericError(96, sn)
		}
	}
	if debug {
		adatypes.Central.Log.Debugf("Read logical ISN with %s=%s", sn, keyValue)
	}
	resultRead, rErr := iRequest.ReadLogicalWith(sn + "=" + keyValue)
	if rErr != nil {
		return rErr
	}
	if len(resultRead.Values) != 1 {
		return adatypes.NewGenericError(98, sn)
	}
	storeRecord.Isn = resultRead.Values[0].Isn
	if debug {
		adatypes.Central.Log.Debugf("Update ISN", storeRecord.Isn)
	}
	return nil
}

// StoreData store interface data, either struct or array
func (request *StoreRequest) StoreData(data ...interface{}) error {
	etData := true
	if len(data) > 1 {
		etData = data[1].(bool)
	}
	return request.modifyData(data[0], true, etData)
}

// UpdateData update interface data, either struct or array
func (request *StoreRequest) UpdateData(data ...interface{}) error {
	etData := true
	if len(data) > 1 {
		etData = data[1].(bool)
	}
	return request.modifyData(data[0], false, etData)
}

// StoreData store interface data, either struct or array
func (request *StoreRequest) modifyData(data interface{}, store, etData bool) error {
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Store type = %T %v", data, reflect.TypeOf(data).Kind())
	}
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		if debug {
			adatypes.Central.Log.Debugf("Work on slice")
		}
		s := reflect.ValueOf(data)
		if s.Len() == 0 {
			return adatypes.NewGenericError(54)
		}
		if request.dynamic == nil {
			request.createDynamic(s.Index(0))
		}

		for si := 0; si < s.Len(); si++ {
			if debug {
				adatypes.Central.Log.Debugf("Store slice entry %d", si)
			}
			err := request.storeValue(s.Index(si), store, etData)
			if err != nil {
				return err
			}
		}
		if debug {
			adatypes.Central.Log.Debugf("Store all slice entries")
		}
	case reflect.Ptr:
		if request.dynamic == nil {
			request.createDynamic(data)
		}
		if debug {
			adatypes.Central.Log.Debugf("Type data %T", data)
		}
		ti := reflect.ValueOf(data).Elem()
		err := request.storeValue(ti, store, etData)
		if err != nil {
			return err
		}
	case reflect.Struct:
		if request.dynamic == nil {
			request.createDynamic(data)
		}
		if debug {
			adatypes.Central.Log.Debugf("Type data %T", data)
		}
		ti := reflect.ValueOf(data)
		err := request.storeValue(ti, store, etData)
		if err != nil {
			return err
		}
	default:
		adatypes.Central.Log.Errorf("Unknown Store type %v", reflect.TypeOf(data).Kind())
		return adatypes.NewGenericError(0)
	}
	return nil
}
