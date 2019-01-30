/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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

// Cursoring cursor instance handling cursoring
type Cursoring struct {
	offset        uint32
	search        string
	result        *Response
	request       *ReadRequest
	adabasRequest *adatypes.Request
	err           error
}

// ReadLogicalWithCursoring read records using cursoring
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
	return request.cursoring, nil
}

// HasNextRecord check cursoring if next record available
func (cursor *Cursoring) HasNextRecord() (hasNext bool) {
	if cursor.offset+1 > uint32(len(cursor.result.Values)) {
		if cursor.adabasRequest.Response != AdaNormal {
			return false
		}
		cursor.result, cursor.err = cursor.request.ReadLogicalWith(cursor.search)
		if cursor.err != nil {
			return false
		}
		hasNext = len(cursor.result.Values) > 0
		cursor.offset = 0
	} else {
		hasNext = true
	}
	return
}

// NextRecord cursoring to next record
func (cursor *Cursoring) NextRecord() (record *Record, err error) {
	if cursor.err != nil {
		return nil, cursor.err
	}
	adatypes.Central.Log.Debugf("offset=%d/%d\n", cursor.offset, len(cursor.result.Values))
	if cursor.offset+1 > uint32(len(cursor.result.Values)) {
		if !cursor.HasNextRecord() {
			return nil, nil
		}
	}
	cursor.offset++
	adatypes.Central.Log.Debugf("ISN=%d\n", cursor.result.Values[cursor.offset-1].Isn)
	return cursor.result.Values[cursor.offset-1], nil
}
