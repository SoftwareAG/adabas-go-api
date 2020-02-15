/*
* Copyright © 2019-2020 Software AG, Darmstadt, Germany and/or its licensors
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

// Cursoring the structure support cursor instance handling reading record list in
// chunks defined by a search or descriptor
type Cursoring struct {
	offset        uint32
	maxLength     uint32
	bufferSize    uint32
	search        string
	result        *Response
	request       *ReadRequest
	adabasRequest *adatypes.Request
	err           error
}

// ReadLogicalWithCursoring initialize the read records using cursoring
func (request *ReadRequest) ReadLogicalWithCursoring(search string) (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.ReadLogicalWith(search)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.request = request
	request.queryFunction = request.ReadLogicalWith
	return request.cursoring, nil
}

// HistogramByCursoring initialize the read records using cursoring
func (request *ReadRequest) HistogramByCursoring(descriptor string) (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.HistogramBy(descriptor)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = ""
	request.cursoring.request = request
	request.queryFunction = request.HistogramBy
	return request.cursoring, nil
}

// HistogramWithCursoring initialize the read records using cursoring
func (request *ReadRequest) HistogramWithCursoring(search string) (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.HistogramWith(search)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.request = request
	request.queryFunction = request.HistogramWith
	return request.cursoring, nil
}

// HasNextRecord check cursoring if a next record exist in the query
func (cursor *Cursoring) HasNextRecord() (hasNext bool) {
	adatypes.Central.Log.Debugf("Check next record: %v offset=%d values=%d", hasNext, cursor.offset+1, len(cursor.result.Values))
	if cursor.offset+1 > uint32(len(cursor.result.Values)) {
		if cursor.adabasRequest == nil || (cursor.adabasRequest.Response != AdaNormal && cursor.adabasRequest.Option.StreamCursor == 0) {
			if cursor.adabasRequest != nil {
				adatypes.Central.Log.Debugf("Error adabas request empty of not normal response, may be EOF %#v\n", cursor.adabasRequest.Response)
			} else {
				adatypes.Central.Log.Debugf("Error adabas request empty %#v\n", cursor.adabasRequest)
			}
			return false
		}

		cursor.result, cursor.err = cursor.request.queryFunction(cursor.search)
		if cursor.err != nil || cursor.result == nil {
			adatypes.Central.Log.Debugf("Error query function %v %#v\n", cursor.err, cursor.result)
			return false
		}
		hasNext = len(cursor.result.Values) > 0
		cursor.offset = 0
	} else {
		hasNext = true
	}
	adatypes.Central.Log.Debugf("Has next record: %v", hasNext)
	return
}

// NextRecord cursoring to next record, if current chunk contains record, no call is send. If
// the chunk is not in memory, the next chunk is read in memory
func (cursor *Cursoring) NextRecord() (record *Record, err error) {
	if cursor.err != nil {
		adatypes.Central.Log.Debugf("Error next record: %v", err)
		return nil, cursor.err
	}
	adatypes.Central.Log.Debugf("Get next record offset=%d/%d\n", cursor.offset, len(cursor.result.Values))
	if cursor.offset+1 > uint32(len(cursor.result.Values)) {
		if !cursor.HasNextRecord() {
			return nil, nil
		}
	}
	cursor.offset++
	adatypes.Central.Log.Debugf("ISN=%d ISN quantity=%d\n", cursor.result.Values[cursor.offset-1].Isn,
		cursor.result.Values[cursor.offset-1].Quantity)
	return cursor.result.Values[cursor.offset-1], nil
}
