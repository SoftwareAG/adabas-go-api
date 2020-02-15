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
	"fmt"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// ReadLobStream initialize the read stream of a field using cursoring
func (request *ReadRequest) ReadLobStream(search, field string) (cursor *Cursoring, err error) {
	err = request.QueryFields("#" + field)
	if err != nil {
		return
	}
	request.definition.DumpTypes(true, true, "Lob stream preparation")
	result, rerr := request.ReadLogicalWith(search)
	if rerr != nil {
		return nil, rerr
	}
	if len(result.Values) != 1 {
		if len(result.Values) > 1 {
			err = adatypes.NewGenericError(137)
			return
		}
		err = adatypes.NewGenericError(138)
		return
	}
	adatypes.Central.Log.Debugf("Found record ...streaming ISN=%d", result.Values[0].Isn)
	v, verr := result.Values[0].SearchValue("#" + field)
	if verr != nil {
		return nil, verr
	}
	recLen, _ := v.UInt32()

	// Limit records to 3 because more then one record is error case, match should be only
	// one record defining a stream
	request.Limit = 1
	request.Multifetch = 1

	// Prepare search of field to current one
	err = request.QueryFields(field)
	if err != nil {
		return
	}
	request.definition.DumpTypes(true, true, "Lob stream partial")

	v = request.definition.Search(field)
	vs := v.(adatypes.PartialValue)
	vs.SetPartial(0, 4096)

	result, rerr = request.ReadFieldStream(search)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.request = request
	request.cursoring.offset = 0
	request.cursoring.maxLength = recLen
	request.cursoring.bufferSize = 4096
	request.queryFunction = request.ReadFieldStream
	return request.cursoring, nil
}

// ReadFieldStream read records with a field stream
func (request *ReadRequest) ReadFieldStream(search string) (result *Response, err error) {
	adatypes.Central.Log.Debugf("Read stream with parser")
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		request.cursoring = &Cursoring{}
		result = &Response{Definition: request.definition, fields: request.fields}

		adatypes.Central.Log.Debugf("Definition generated ...")
		adabasRequest, prepareErr := request.prepareRequest()
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adatypes.Central.Log.Debugf("Prepare done ...")
		adabasRequest.Parser = parseReadToRecord
		adabasRequest.Limit = request.Limit
		request.cursoring.result = result
		request.adaptDescriptorMap(adabasRequest)
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}
		err = request.adabas.ReadStream(request.cursoring.adabasRequest, 0, request.cursoring.result)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("read with ...streaming ISN=%d", request.cursoring.adabasRequest.Isn)
		request.cursoring.adabasRequest.Option.StreamCursor++
		adatypes.Central.Log.Debugf("After first stream definition values %p avail.=%v", request.cursoring.adabasRequest.Definition.Values, (request.cursoring.adabasRequest.Definition.Values != nil))
	} else {
		adatypes.Central.Log.Debugf("Start offset=%d max=%d", request.cursoring.offset, request.cursoring.maxLength)
		if uint32(request.cursoring.adabasRequest.Option.StreamCursor)*4096 > request.cursoring.maxLength {
			return nil, fmt.Errorf("End reached")
		}
		request.cursoring.adabasRequest.Definition.Values = request.cursoring.result.Values[0].Value
		adatypes.Central.Log.Debugf("Next read with ...streaming ISN=%d avail.=%v", request.cursoring.adabasRequest.Isn, (request.cursoring.adabasRequest.Definition.Values != nil))
		err = request.adabas.loopCall(request.cursoring.adabasRequest, request.cursoring.result)
		result = request.cursoring.result
		adatypes.Central.Log.Debugf("Stream read finished: %v", err)
		request.cursoring.adabasRequest.Option.StreamCursor++
	}
	request.cursoring.offset = uint32(request.cursoring.adabasRequest.Option.StreamCursor) * 4096
	return
}
