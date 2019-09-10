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
	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type commonRequest struct {
	adabas      *Adabas
	MapName     string
	adabasMap   *Map
	repository  *Repository
	definition  *adatypes.Definition
	initialized bool
	dynamic     *adatypes.DynamicInterface
}

// StreamFunction function callback used to go through the list of received records
type StreamFunction func(record *Record, x interface{}) error

// InterfaceFunction function callback used to go through the list of received interface instances
type InterfaceFunction func(data interface{}, x interface{}) error

func (request *commonRequest) loadDefinition() (err error) {
	if request.definition == nil {
		adatypes.Central.Log.Debugf("Load file Definition ....")
		request.definition, err = request.adabas.ReadFileDefinition(request.repository.Fnr)
		if err != nil {
			return
		}
		if adatypes.Central.IsDebugLevel() {
			adatypes.Central.Log.Debugf("Finish loading file definition ....")
			request.definition.DumpTypes(true, false, "load definition")
		}
	}
	//debug.PrintStack()
	return
}

// Close closes the Adabas session
func (request *commonRequest) Close() {
	if request == nil {
		return
	}
	if request.adabas != nil {
		request.adabas.Close()
	}
	request.definition = nil
	request.initialized = false
}

// Endtransaction end the transaction of the Adabas session
func (request *commonRequest) EndTransaction() error {
	return request.adabas.EndTransaction()
}

// Backout closes the Adabas session
func (request *commonRequest) BackoutTransaction() error {
	return request.adabas.BackoutTransaction()
}

// Open Open the Adabas session
func (request *commonRequest) commonOpen() (opened bool, err error) {
	opened = false
	adatypes.Central.Log.Debugf("Open read request")
	err = request.adabas.Open()
	if err != nil {
		return
	}
	if request.initialized {
		return
	}
	if request.MapName != "" {
		adatypes.Central.Log.Debugf("Open Adabas with map %s for %d", request.MapName, request.repository.Fnr)
		if request.adabasMap == nil {
			request.adabasMap, err = request.repository.readAdabasMapWithRequest(request, request.MapName)
			if err != nil {
				adatypes.Central.Log.Debugf("Error reading Adabas map request ", err)
				return
			}
		}
		var dbid Dbid
		dbid, err = request.adabasMap.Repository.dbid()
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Reset database to new database: %d current: %d", dbid, request.adabas.Acbx.Acbxdbid)
		if dbid != 0 {
			request.adabas.SetDbid(dbid)
		}
		adatypes.Central.Log.Debugf("Load definition on fnr=%d/%d for map %s", request.repository.Fnr, request.adabasMap.Repository.Fnr, request.adabasMap.Name)
		err = request.loadDefinition()
		if err != nil {
			return
		}
		if request.definition == nil {
			adatypes.Central.Log.Debugf("Error request definition empty")
			err = adatypes.NewGenericError(26)
			return
		}
		err = request.adabasMap.adaptFieldType(request.definition, request.dynamic)
		if err != nil {
			adatypes.Central.Log.Debugf("Adapt fields error request definition %v", err)
			return
		}
	} else {
		adatypes.Central.Log.Debugf("Open database without map")
		err = request.loadDefinition()
		if err != nil {
			return
		}
	}
	if adatypes.Central.IsDebugLevel() {
		request.definition.DumpTypes(true, true, "Database open")
		adatypes.Central.Log.Debugf("Database open complete")
	}
	request.initialized = true
	opened = true
	return
}

// IsOpen provide True if the database connection is opened
func (request *commonRequest) IsOpen() bool {
	return request.adabas.status.open
}

// TraverseFields traverse through all request fields
func (request *commonRequest) TraverseFields(t adatypes.TraverserMethods, p interface{}) error {
	return request.definition.TraverseTypes(t, true, p)
}
