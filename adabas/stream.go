/*
* Copyright © 2020-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	//err = request.QueryFields("#" + field)
	err = request.QueryFields("")
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
	request.definition.ResetRestrictToFields()

	adatypes.Central.Log.Debugf("Found record ...streaming ISN=%d BlockSize=%d",
		result.Values[0].Isn, request.BlockSize)
	request.cursoring.result, err = request.ReadLOBRecord(result.Values[0].Isn, field, uint64(request.BlockSize))
	if err != nil {
		return nil, err
	}
	adatypes.Central.Log.Debugf("Read record ...streaming ISN=%d BlockSize=%d",
		result.Values[0].Isn, request.BlockSize)
	return request.cursoring, nil
}

// ReadLOBSegment reads field data using partial lob reads and provide it
// stream-like to the user. It is possible to use it to read big LOB data
// or stream a video.
// The value of the field is reused, so that the value does not need to
// evaluate/searched in the result instance once more after the first
// call.
// This method initialize the first call by
// - offset 0
// - prepare partial lob query of given blocksize
// Important parameter is the blocksize in the `ReadRequest` which
// defines the size of one block to be read
func (request *ReadRequest) ReadLOBSegment(isn adatypes.Isn, field string, blocksize uint64) (segment []byte, err error) {
	result, err := request.ReadLOBRecord(isn, field, blocksize)
	if err != nil {
		return nil, err
	}
	// If last segment is reached and EOF is received, record is not available
	if result.NrRecords() != 1 {
		segment = make([]byte, 0)
		return
	}
	// Search record containing the segment
	v, found := result.Values[0].searchValue(field)
	if found {
		segment = v.Bytes()
		return
	}
	err = adatypes.NewGenericError(183)
	return
}

// ReadLOBRecord read lob records in an stream, repeated call will read next segment of LOB
func (request *ReadRequest) ReadLOBRecord(isn adatypes.Isn, field string, blocksize uint64) (result *Response, err error) {
	debug := adatypes.Central.IsDebugLevel()
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		if debug {
			adatypes.Central.Log.Debugf("Read LOB record initiated ...")
		}
		request.cursoring = &Cursoring{}
		request.BlockSize = uint32(blocksize)
		request.PartialRead = true
		_, oErr := request.Open()
		if oErr != nil {
			err = oErr
			return
		}
		err = request.QueryFields(field)
		if err != nil {
			adatypes.Central.Log.Debugf("Query fields error ...%#v", err)
			return nil, err
		}
		if debug {
			adatypes.Central.Log.Debugf("LOB Definition generated ...BlockSize=%d", request.BlockSize)
		}
		err = request.definition.CreateValues(false)
		if err != nil {
			return
		}
		if debug {
			adatypes.Central.Log.Debugf("LOB create values, types defined")
			request.definition.DumpTypes(true, true)
			adatypes.Central.Log.Debugf("LOB list of values")
			request.definition.DumpValues(true)
			adatypes.Central.Log.Debugf("Search field: %s", field)
		}

		fieldName, index := parseField(field)
		fieldValue, ferr := request.definition.SearchByIndex(fieldName, index, true)
		if ferr != nil {
			return nil, ferr
		}
		if fieldValue == nil {
			return nil, adatypes.NewGenericError(184, field)
		}
		if debug {
			adatypes.Central.Log.Debugf("LOB after defined")
			request.definition.DumpValues(true)
			adatypes.Central.Log.Debugf("Found field: %s for %d,%d", fieldValue.Type().Name(), fieldValue.MultipleIndex(), fieldValue.PeriodIndex())
		}
		var lob adatypes.ILob
		switch t := fieldValue.(type) {
		case adatypes.ILob:
			lob = t
		default:
			return nil, adatypes.NewGenericError(185, field)
		}
		lob.SetLobBlockSize(blocksize)
		lob.SetLobPartRead(true)
		if debug {
			adatypes.Central.Log.Debugf("Read LOB with ...%#v", field)
		}

		adabasRequest, prepareErr := request.prepareRequest(false)
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adabasRequest.Parser = parseReadToRecord
		adabasRequest.Limit = 1
		request.cursoring.adabasRequest = adabasRequest
		if debug {
			adatypes.Central.Log.Debugf("Query field LOB values ...%#v", field)
			adatypes.Central.Log.Debugf("Create LOB values ...%#v", field)
		}
		adabasRequest.Limit = 1
		adabasRequest.Multifetch = 1
		adabasRequest.Isn = isn
		adabasRequest.Option.PartialRead = true
		request.cursoring.adabasRequest = adabasRequest
		request.cursoring.search = field
		request.queryFunction = request.readSteamSegment
		request.cursoring.request = request
		result = &Response{Definition: request.definition, fields: request.fields}
		err = request.adabas.readISN(request.repository.Fnr, adabasRequest, result)
	} else {
		if debug {
			adatypes.Central.Log.Debugf("Read next LOB segment with ...cursoring")
		}
		/*
			request.definition.DumpTypes(false, false, "All")
			request.definition.DumpTypes(false, true, "Active")
			request.definition.DumpValues(true)
			request.definition.DumpValues(false)*/
		err = request.definition.CreateValues(false)
		if err != nil {
			return
		}
		/*
			request.definition.DumpTypes(false, false, "All")
			request.definition.DumpTypes(false, true, "Active")
			request.definition.DumpValues(true)
			request.definition.DumpValues(false)
			fmt.Println("Search", field, request.definition.Fieldnames())
		*/
		fieldValue := request.definition.Search(field)
		lob := fieldValue.(adatypes.ILob)
		lob.SetLobBlockSize(blocksize)
		lob.SetLobPartRead(true)
		request.cursoring.adabasRequest.Option.PartialRead = true
		/*
			request.definition.DumpValues(true)
			request.definition.DumpValues(false)
		*/
		result = &Response{Definition: request.definition, fields: request.fields}
		if debug {
			adatypes.Central.Log.Debugf("Call next LOB read %v/%d", request.cursoring.adabasRequest.Option.PartialRead, request.BlockSize)
		}
		err = request.adabas.loopCall(request.cursoring.adabasRequest, result)
	}
	if debug {
		adatypes.Central.Log.Debugf("Error reading %v", err)
	}

	return result, err
}

func (request *ReadRequest) readSteamSegment(search, descriptors string) (result *Response, err error) {
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("read LOB >1 segments")
	}
	return request.ReadLOBRecord(0, search, uint64(request.BlockSize))
}
