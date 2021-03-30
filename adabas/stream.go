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

// ReadLobStream reads field data using partial lob reads and provide it
// stream-like to the user. It is possible to use it to read big LOB data
// or stream a video.
// The value of the field is reused, so that the value does not need to
// evaluate/searched in the result instance once more after the first
// call.
// This method initialize the first call by
// - searching the record (it must be a unique result)
// - prepare partial lob query
// Important parameter is the blocksize in the `ReadRequest` which
// defines the size of one block to be read
func (request *ReadRequest) ReadLobStream(search, field string) (cursor *Cursoring, err error) {
	err = request.QueryFields("#" + field)
	if err != nil {
		return
	}
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

	// Define partial range on field value
	v = request.definition.Search(field)
	vs := v.(adatypes.PartialValue)
	vs.SetPartial(0, request.BlockSize)

	// Init first read of the stream used in the cursoring
	result, rerr = request.ReadFieldStream(search)
	if rerr != nil {
		return nil, rerr
	}

	// Prepare cursoring
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.request = request
	request.cursoring.offset = 0
	request.cursoring.FieldLength = recLen
	request.queryFunction = request.readFieldStream
	return request.cursoring, nil
}

// ReadFieldStream reads field data using partial lob reads and provide it
// stream-like to the user. It is possible to use it to read big LOB data
// or stream a video.
// The value of the field is reused, so that the value does not need to
// evaluate/searched in the result instance once more after the first
// call.
// The stream will be read in blocks. The blocksize is defined in the
// `ReadRequest`. The last block will be filled up with space by Adabas.
// To examine the length of the field, the `Cursoring` instance provide
// the current block number in `StreamCursor` beginning by 0 and the
// maximum field length in `FieldLength`.
func (request *ReadRequest) ReadFieldStream(search string) (result *Response, err error) {
	adatypes.Central.Log.Debugf("Read stream with parser")
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		request.cursoring = &Cursoring{}
		result = &Response{Definition: request.definition, fields: request.fields}

		adabasRequest, prepareErr := request.prepareRequest()
		if prepareErr != nil {
			err = prepareErr
			return
		}

		// Define parser parameters
		adabasRequest.Parser = parseReadToRecord
		adabasRequest.Limit = request.Limit
		request.cursoring.result = result
		request.adaptDescriptorMap(adabasRequest)
		request.cursoring.adabasRequest = adabasRequest

		// Call first
		err = request.adabas.ReadStream(request.cursoring.adabasRequest, 0, request.cursoring.result)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("read with ...streaming ISN=%d", request.cursoring.adabasRequest.Isn)
		request.cursoring.adabasRequest.Option.StreamCursor++
	} else {
		if uint32(request.cursoring.adabasRequest.Option.StreamCursor)*request.BlockSize > request.cursoring.FieldLength {
			return nil, adatypes.NewGenericError(168)
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

func (request *ReadRequest) readFieldStream(search, descriptors string) (result *Response, err error) {
	return request.ReadFieldStream(search)
}
