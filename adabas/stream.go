/*
* Copyright © 2020 Software AG, Darmstadt, Germany and/or its licensors
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

// ReadLobStream initialize the read stream of a field using cursoring
func (request *ReadRequest) ReadLobStream(search, field string) (cursor *Cursoring, err error) {
	err = request.QueryFields(field)
	if err != nil {
		return
	}
	request.definition.DumpTypes(true, true, "Lob stream restrict")
	v := request.definition.Search(field)
	vs := v.(adatypes.PartialValue)
	vs.SetPartial(0, 4096)
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.ReadFieldStream(search)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.request = request
	request.queryFunction = request.ReadFieldStream
	return request.cursoring, nil
}

// ReadFieldStream read records with a field stream
func (request *ReadRequest) ReadFieldStream(search string) (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	adatypes.Central.Log.Debugf("Read stream with parser")
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		opened, oErr := request.Open()
		if oErr != nil {
			err = oErr
			return
		}
		if opened {
			if request.dynamic != nil && request.definition == nil {
				q := request.dynamic.CreateQueryFields()
				request.QueryFields(q)
			}
			adatypes.Central.Log.Debugf("Query fields Definition ...")
		}
		adatypes.Central.Log.Debugf("Read logical, open done ...%#v with search=%s", request.adabas.ID.platform, search)
		var searchInfo *adatypes.SearchInfo
		var tree *adatypes.SearchTree
		if search != "" {
			searchInfo = adatypes.NewSearchInfo(request.adabas.ID.platform(request.adabas.URL.String()), search)
			adatypes.Central.Log.Debugf("New search info ... %#v", searchInfo)
			if request.definition == nil {
				adatypes.Central.Log.Debugf("Load Definition (read logical)...")
				err = request.loadDefinition()
				if err != nil {
					return
				}
				adatypes.Central.Log.Debugf("Loaded Definition ...")
				if request.dynamic != nil {
					q := request.dynamic.CreateQueryFields()
					request.QueryFields(q)
					adatypes.Central.Log.Debugf("Query fields Definition ...")
				}
				searchInfo.Definition = request.definition
				tree, err = searchInfo.GenerateTree()
				if err != nil {
					return
				}
			} else {
				adatypes.Central.Log.Debugf("Use Definition ...")
				searchInfo.Definition = request.definition
				tree, err = searchInfo.GenerateTree()
				if err != nil {
					return
				}
			}
		} else {
			adatypes.Central.Log.Debugf("No search ...")
			err = adatypes.NewGenericError(136)
			return
		}

		adatypes.Central.Log.Debugf("Definition generated ...")
		adabasRequest, prepareErr := request.prepareRequest()
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adatypes.Central.Log.Debugf("Prepare done ...")
		switch {
		// case resultParser != nil:
		// 	adabasRequest.Parser = resultParser
		case adabasRequest.DataType != nil:
			adabasRequest.Parser = parseReadToInterface
		default:
			adabasRequest.Parser = parseReadToRecord
		}
		adabasRequest.Limit = request.Limit
		//searchInfo.Definition = adabasRequest.Definition
		if tree != nil {
			adabasRequest.SearchTree = tree
			adabasRequest.Descriptors = tree.OrderBy()
		}
		request.adaptDescriptorMap(adabasRequest)
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		if searchInfo == nil {
			err = adatypes.NewGenericError(136)
			return
		}
		adabasRequest.Isn = 0
		if searchInfo.NeedSearch {
			adatypes.Central.Log.Debugf("search logical with ...%#v", adabasRequest.Descriptors)
			err = request.adabas.SearchLogicalWith(request.repository.Fnr, adabasRequest, result)
		} else {
			adatypes.Central.Log.Debugf("read logical with ...%#v", adabasRequest.Descriptors)
			err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, result)
		}
		if len(result.Values) != 1 {
			if len(result.Values) > 1 {
				err = adatypes.NewGenericError(137)
				return
			}
			err = adatypes.NewGenericError(138)
			return
		}

	} else {
		adatypes.Central.Log.Debugf("read logical with ...cursoring")
		//err = request.adabas.loopCall(request.cursoring.adabasRequest, result)
		//err = request.adabas.loopCall(request.cursoring.adabasRequest, nil)
	}
	adatypes.Central.Log.Debugf("Read finished")
	return
}
