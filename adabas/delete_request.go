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
	"strconv"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// DeleteRequest request instance handling data store and update
type DeleteRequest struct {
	commonRequest
}

// NewDeleteRequestDeprecated create a new store Request instance
func NewDeleteRequestDeprecated(url string, fnr Fnr) (*DeleteRequest, error) {
	var adabas *Adabas
	if dbid, err := strconv.Atoi(url); err == nil {
		if (dbid < 0) || dbid > 65536 {
			err = adatypes.NewGenericError(70, dbid)
			return nil, err
		}
		adabas, err = NewAdabas(Dbid(dbid))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &DeleteRequest{commonRequest: commonRequest{adabas: adabas,
		repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}, nil
}

// NewDeleteRequestAdabas create a new Request instance
func NewDeleteRequestAdabas(adabas *Adabas, fnr Fnr) *DeleteRequest {
	clonedAdabas := NewClonedAdabas(adabas)

	return &DeleteRequest{commonRequest: commonRequest{adabas: clonedAdabas,
		repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// NewMapDeleteRequest create a new Map Delete Request instance
func NewMapDeleteRequest(adabas *Adabas, adabasMap *Map) (request *DeleteRequest, err error) {
	mapName := adabasMap.Name
	adatypes.Central.Log.Debugf("Delete: Adabas new map reference to %d", adabasMap.Data.Fnr)
	var dataAdabas *Adabas
	if adabas.URL.String() == adabasMap.Data.URL.String() {
		dataAdabas = NewClonedAdabas(adabas)
	} else {
		dataAdabas, err = NewAdabas(&adabasMap.Data.URL, adabas.ID)
		if err != nil {
			return nil, err
		}
	}
	dataAdabas.Acbx.Acbxfnr = adabasMap.Data.Fnr
	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &DeleteRequest{commonRequest: commonRequest{MapName: mapName, adabas: dataAdabas, adabasMap: adabasMap,
		repository: dataRepository}}
	adatypes.Central.Log.Debugf("Delete per map to %s/%d", request.adabas.String(), request.repository.Fnr)
	return
}

// NewMapNameDeleteRequest create a new Request instance
func NewMapNameDeleteRequest(adabas *Adabas, mapName string) (request *DeleteRequest, err error) {
	var adabasMap *Map
	adabasMap, _, err = SearchMapRepository(adabas.ID, mapName)
	if err != nil {
		return
	}
	dbid, repErr := adabasMap.Data.dbid()
	if repErr != nil {
		err = repErr
		return
	}
	clonedAdabas := NewClonedAdabas(adabas)
	adabas.SetDbid(dbid)
	adatypes.Central.Log.Debugf("Delete: Adabas new map reference to %d", adabasMap.Data.Fnr)

	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &DeleteRequest{commonRequest: commonRequest{MapName: mapName, adabas: clonedAdabas, adabasMap: adabasMap,
		repository: dataRepository}}
	return
}

// NewMapNameDeleteRequestRepo create a new delete Request instance
func NewMapNameDeleteRequestRepo(mapName string, adabas *Adabas, mapRepository *Repository) (request *DeleteRequest, err error) {
	var adabasMap *Map
	adabasMap, err = mapRepository.SearchMap(adabas, mapName)
	if err != nil {
		return
	}
	dbid, repErr := adabasMap.Data.dbid()
	if repErr != nil {
		err = repErr
		return
	}
	clonedAdabas := NewClonedAdabas(adabas)
	adabas.SetDbid(dbid)
	adatypes.Central.Log.Debugf("Delete: Adabas new map reference to %d", adabasMap.Data.Fnr)

	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &DeleteRequest{commonRequest: commonRequest{MapName: mapName, adabas: clonedAdabas, adabasMap: adabasMap,
		repository: dataRepository}}
	return
}

// Open Open the Adabas session
func (deleteRequest *DeleteRequest) Open() (opened bool, err error) {
	return deleteRequest.commonOpen()
}

// Delete delete a specific isn
func (deleteRequest *DeleteRequest) Delete(isn adatypes.Isn) (err error) {
	_, err = deleteRequest.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Deleting...%s -> URL: %s fnr=%d", deleteRequest.adabas.String(),
		deleteRequest.repository.URL.String(), deleteRequest.repository.Fnr)
	return deleteRequest.adabas.DeleteIsn(deleteRequest.repository.Fnr, isn)
}

// DeleteList delete a list of isns
func (deleteRequest *DeleteRequest) DeleteList(isnList []adatypes.Isn) (err error) {
	for _, isn := range isnList {
		err = deleteRequest.Delete(isn)
		if err != nil {
			return err
		}
	}
	return nil
}
