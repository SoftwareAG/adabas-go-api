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
	// FieldLength in streaming mode of a field the field length of the field
	FieldLength   uint32
	offset        uint32
	search        string
	descriptors   string
	empty         bool
	result        *Response
	request       *ReadRequest
	adabasRequest *adatypes.Request
	err           error
}

// ReadLogicalWithCursoring this method provide the search of records in Adabas
// and provide a cursor. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
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
	request.queryFunction = request.readLogicalWith
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// ReadLogicalByCursoring this method provide the descriptor read of records in Adabas
// and provide a cursor. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
func (request *ReadRequest) ReadLogicalByCursoring(descriptor string) (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.ReadLogicalBy(descriptor)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = ""
	request.cursoring.request = request
	request.queryFunction = request.readLogicalBy
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// SearchAndOrderWithCursoring this method provide the search of records in Adabas
// ordered by a descriptor. It provide a cursor. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
func (request *ReadRequest) SearchAndOrderWithCursoring(search, descriptors string) (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	if request.Limit == 0 {
		request.Limit = 10
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.SearchAndOrder(search, descriptors)
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = search
	request.cursoring.descriptors = descriptors
	request.cursoring.request = request
	request.queryFunction = request.SearchAndOrder
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// HistogramByCursoring provides the descriptor read of a field and uses
// cursoring. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
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
	request.cursoring.descriptors = descriptor
	request.cursoring.request = request
	request.queryFunction = request.histogramBy
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// HistogramWithCursoring provides the searched read of a descriptor of a
// field. It uses a cursor to read only a part of the data and read further
// only on request. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
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
	request.queryFunction = request.histogramWith
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// ReadPhysicalWithCursoring this method provide the physical read of records in Adabas
// and provide a cursor. The cursor will read a number of records using Multifetch
// calls. The number of records is defined in `Limit`.
// This method initialize the read records using cursoring.
func (request *ReadRequest) ReadPhysicalWithCursoring() (cursor *Cursoring, err error) {
	request.cursoring = &Cursoring{}
	/* Define the read chunk to 20 if not defined */
	if request.Limit == 0 {
		request.Limit = 20
	}
	request.Multifetch = uint32(request.Limit)
	if request.Multifetch > 20 {
		request.Multifetch = 20
	}
	result, rerr := request.ReadPhysicalSequence()
	if rerr != nil {
		return nil, rerr
	}
	request.cursoring.result = result
	request.cursoring.search = ""
	request.cursoring.request = request
	request.queryFunction = request.readPhysical
	request.cursoring.empty = result.NrRecords() == 0
	return request.cursoring, nil
}

// HasNextRecord with cursoring this method checks if a next record
// or stream entry is available return `true` if it is.
// This method will call Adabas if no entry is available and reads new entries
// using Multifetch or partial LOB.
// If an error during processing occurs, the function will return an false and
// you need to check with cursor Error() methods
func (cursor *Cursoring) HasNextRecord() (hasNext bool) {
	if cursor == nil || cursor.empty {
		return false
	}
	adatypes.Central.Log.Debugf("Check next record: %v offset=%d values=%d data=%d", hasNext, cursor.offset+1, len(cursor.result.Values), len(cursor.result.Data))
	if cursor.offset+1 > uint32(cursor.result.NrRecords()) {
		if cursor.adabasRequest == nil || (cursor.adabasRequest.Response != AdaNormal && cursor.adabasRequest.Option.StreamCursor == 0) {
			if cursor.adabasRequest != nil {
				adatypes.Central.Log.Debugf("Error adabas request empty of not normal response, may be EOF %#v", cursor.adabasRequest.Response)
			} else {
				adatypes.Central.Log.Debugf("Error adabas request empty %#v", cursor.adabasRequest)
			}
			return false
		}

		cursor.result, cursor.err = cursor.request.queryFunction(cursor.search, cursor.descriptors)
		if cursor.err != nil || cursor.result == nil {
			adatypes.Central.Log.Debugf("Error query function %v %#v", cursor.err, cursor.result)
			return false
		}
		adatypes.Central.Log.Debugf("Nr Records cursored %d", cursor.result.NrRecords())
		hasNext = cursor.result.NrRecords() > 0
		cursor.offset = 0
	} else {
		hasNext = true
	}
	adatypes.Central.Log.Debugf("Has next record: %v", hasNext)
	return
}

// NextRecord cursoring to next record, if current chunk contains a record, no call is send. If
// the chunk is not in memory, the next chunk is read into memory. This method may be initiated,
// if `HasNextRecord()` is called before.
func (cursor *Cursoring) NextRecord() (record *Record, err error) {
	if cursor.empty {
		return nil, adatypes.NewGenericError(141)
	}
	if cursor.err != nil {
		adatypes.Central.Log.Debugf("Error next record: %v", err)
		return nil, cursor.err
	}
	adatypes.Central.Log.Debugf("Get next record offset=%d/%d", cursor.offset, len(cursor.result.Values))
	if cursor.offset+1 > uint32(cursor.result.NrRecords()) {
		if !cursor.HasNextRecord() {
			return nil, nil
		}
	}
	cursor.offset++
	adatypes.Central.Log.Debugf("ISN=%d ISN quantity=%d", cursor.result.Values[cursor.offset-1].Isn,
		cursor.result.Values[cursor.offset-1].Quantity)
	if len(cursor.result.Data) > 0 {
		return nil, adatypes.NewGenericError(139)
	}
	return cursor.result.Values[cursor.offset-1], nil
}

// NextData cursoring to next structure representation of the data record, if current chunk contains a
// record, no call is send. If
// the chunk is not in memory, the next chunk is read into memory. This method may be initiated,
// if `HasNextRecord()` is called before.
func (cursor *Cursoring) NextData() (record interface{}, err error) {
	if cursor.empty {
		return nil, adatypes.NewGenericError(141)
	}
	if cursor.err != nil {
		adatypes.Central.Log.Debugf("Error next data record: %v", err)
		return nil, cursor.err
	}
	adatypes.Central.Log.Debugf("Get next data record offset=%d/%d", cursor.offset, len(cursor.result.Values))
	if cursor.offset+1 > uint32(cursor.result.NrRecords()) {
		if !cursor.HasNextRecord() {
			return nil, nil
		}
	}
	cursor.offset++
	if len(cursor.result.Values) > 0 {
		return nil, adatypes.NewGenericError(139)
	}
	return cursor.result.Data[cursor.offset-1], nil
}

// Error Provide the current error state for the cursor
func (cursor *Cursoring) Error() (err error) {
	if cursor == nil {
		return nil
	}
	return cursor.err
}
