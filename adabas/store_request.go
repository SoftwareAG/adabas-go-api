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
	"encoding/binary"
	"strconv"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// StoreRequest request instance handling data store and update
type StoreRequest struct {
	commonRequest
}

// NewStoreRequest create a new store Request instance
func NewStoreRequest(url string, fnr Fnr) (*StoreRequest, error) {
	var adabas *Adabas
	if dbid, err := strconv.Atoi(url); err == nil {
		adabas, err = NewAdabas(Dbid(dbid))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &StoreRequest{commonRequest: commonRequest{adabas: adabas,
		repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}, nil
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

// Open Open the Adabas session
func (request *StoreRequest) Open() (err error) {
	err = request.commonOpen()
	return
}

func (request *StoreRequest) prepareRequest() (adabasRequest *adatypes.Request, err error) {
	adabasRequest, err = request.definition.CreateAdabasRequest(true, false, request.adabas.status.platform.IsMainframe())
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Store FB: %s", adabasRequest.FormatBuffer.String())
	adabasRequest.Definition = request.definition
	return
}

// StoreFields create record field definition for the next store
func (request *StoreRequest) StoreFields(fields string) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	// if request.definition == nil {
	// 	err = request.loadDefinition()
	// 	if err != nil {
	// 		return
	// 	}
	// }
	adatypes.Central.Log.Debugf("Check store fields Definition values %#v", request.definition.Values)
	adatypes.Central.Log.Debugf("Dump all fields")
	request.definition.DumpTypes(true, true)
	adatypes.Central.Log.Debugf("Store restrict fields to %s", fields)
	err = request.definition.ShouldRestrictToFields(fields)
	if err != nil {
		return
	}
	request.definition.DumpTypes(true, true)
	adatypes.Central.Log.Debugf("Definition values %#v", request.definition.Values)
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
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		return prepareErr
	}
	//	storeRecord
	helper := adatypes.NewDynamicHelper(binary.LittleEndian)
	err := storeRecord.createRecordBuffer(helper)
	if err != nil {
		return err
	}

	adabasRequest.RecordBuffer = helper
	err = request.adabas.Store(request.repository.Fnr, adabasRequest)
	// Reset values after storage to reset for next store request
	request.definition.Values = nil
	storeRecord.Isn = adabasRequest.Isn
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
	helper := adatypes.NewDynamicHelper(binary.LittleEndian)
	err := storeRecord.createRecordBuffer(helper)
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
