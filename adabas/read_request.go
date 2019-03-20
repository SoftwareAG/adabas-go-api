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
	"fmt"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const (
	maxReadRecordLimit     = 20
	defaultMultifetchLimit = 10
)

type queryField struct {
	field string
	index int
}

type scanFields struct {
	fields    map[string]*queryField
	parameter []interface{}
}

// ReadRequest request instance handling field query information
type ReadRequest struct {
	commonRequest
	Limit             uint64
	Multifetch        uint32
	RecordBufferShift uint32
	fields            map[string]*queryField
	HoldRecords       adatypes.HoldType
	queryFunction     func(string) (*Response, error)
	cursoring         *Cursoring
}

// NewReadRequestCommon create a request defined by another request (not even ReadRequest required)
func NewReadRequestCommon(commonRequest *commonRequest) *ReadRequest {
	request := &ReadRequest{HoldRecords: adatypes.HoldNone}
	request.commonRequest = *commonRequest
	request.commonRequest.adabasMap = nil
	request.commonRequest.mapName = ""
	return request
}

// NewMapReadRequestRepo create a new Request instance
func NewMapReadRequestRepo(mapName string, adabas *Adabas, repository *Repository) (request *ReadRequest, err error) {
	if repository == nil {
		err = adatypes.NewGenericError(20)
		return
	}
	err = repository.LoadMapRepository(adabas)
	if err != nil {
		return
	}
	adabasMap, serr := repository.SearchMap(adabas, mapName)
	if serr != nil {
		// err = adatypes.NewGenericError(21, mapName)
		return nil, serr
	}
	dataAdabas, nerr := NewAdabasWithURL(&adabasMap.Data.URL, adabas.ID)
	if nerr != nil {
		return nil, nerr
	}
	dataRepository := NewMapRepository(adabas, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: mapName, adabas: dataAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewMapReadRequest create a new Request instance
func NewMapReadRequest(adabas *Adabas, mapName string) (request *ReadRequest, err error) {
	var adabasMap *Map
	adabasMap, err = SearchMapRepository(adabas, mapName)
	if err != nil {
		return
	}
	dbid, repErr := adabasMap.Repository.dbid()
	if repErr != nil {
		err = repErr
		return
	}
	adabas.SetDbid(dbid)
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d", mapName, adabasMap.Data.Fnr)

	dataRepository := NewMapRepository(adabas, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: mapName, adabas: adabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewMapReadRequestByMap create a new Request instance
func NewMapReadRequestByMap(adabas *Adabas, adabasMap *Map) (request *ReadRequest, err error) {
	if adabasMap == nil {
		err = adatypes.NewGenericError(22, adabasMap.Name)
		return
	}
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d -> %#v", adabasMap.Name,
		adabasMap.Data.Fnr, adabas.ID.platform)
	cloneAdabas := NewClonedAdabas(adabas)

	dataRepository := NewMapRepository(adabas, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: adabasMap.Name, adabas: cloneAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewReadRequestAdabas create a new Request instance
func NewReadRequestAdabas(adabas *Adabas, fnr Fnr) *ReadRequest {
	clonedAdabas := NewClonedAdabas(adabas)

	return &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{adabas: clonedAdabas,
			repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// Open Open the Adabas session
func (request *ReadRequest) Open() (err error) {
	err = request.commonOpen()
	return
}

// Prepare read request for special parts in read
func (request *ReadRequest) prepareRequest() (adabasRequest *adatypes.Request, err error) {
	if request.definition == nil {
		err = request.loadDefinition()
		if err != nil {
			return
		}
	}
	adabasRequest, err = request.definition.CreateAdabasRequest(false, false, request.adabas.status.platform.IsMainframe())
	if err != nil {
		return
	}
	adabasRequest.Definition = request.definition
	adabasRequest.RecordBufferShift = request.RecordBufferShift
	adabasRequest.HoldRecords = request.HoldRecords
	return
}

// SetHoldRecords set hold record done
func (request *ReadRequest) SetHoldRecords(hold adatypes.HoldType) {
	request.HoldRecords = hold
}

func parseRead(adabasRequest *adatypes.Request, x interface{}) (err error) {
	result := x.(*Response)

	isn := adabasRequest.Isn
	isnQuantity := adabasRequest.IsnQuantity
	record, xerr := NewRecordIsn(isn, isnQuantity, adabasRequest.Definition)
	if xerr != nil {
		return xerr
	}
	result.Values = append(result.Values, record)
	record.fields = result.fields
	adatypes.Central.Log.Debugf("Got ISN=%d Quantity=%d record", record.Isn, record.Quantity)

	return
}

// ReadPhysicalSequence read records in physical order
func (request *ReadRequest) ReadPhysicalSequence() (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadPhysicalSequenceWithParser(nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadPhysicalSequenceStream read records in physical order
func (request *ReadRequest) ReadPhysicalSequenceStream(streamFunction StreamFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{streamFunction: streamFunction, result: &Response{Definition: request.definition, fields: request.fields}, x: x}
	err = request.ReadPhysicalSequenceWithParser(streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return result, nil
}

// ReadPhysicalSequenceWithParser read records in physical order
func (request *ReadRequest) ReadPhysicalSequenceWithParser(resultParser adatypes.RequestParser, x interface{}) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		err = prepareErr
		return
	}
	if resultParser == nil {
		adabasRequest.Parser = parseRead
	} else {
		adabasRequest.Parser = resultParser
	}
	adabasRequest.Limit = request.Limit
	if request.Multifetch > 1 {
		if request.Limit < uint64(request.Multifetch) {
			adabasRequest.Multifetch = uint32(request.Limit)
		} else {
			adabasRequest.Multifetch = request.Multifetch
		}

	} else {
		adabasRequest.Multifetch = 1
	}

	err = request.adabas.ReadPhysical(request.repository.Fnr, adabasRequest, x)
	return
}

// ReadISN read records defined by a given ISN
func (request *ReadRequest) ReadISN(isn adatypes.Isn) (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadISNWithParser(isn, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadISNWithParser read record defined by a ISN using request parser
func (request *ReadRequest) ReadISNWithParser(isn adatypes.Isn, resultParser adatypes.RequestParser, x interface{}) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		err = prepareErr
		return
	}
	if resultParser == nil {
		adabasRequest.Parser = parseRead
	} else {
		adabasRequest.Parser = resultParser
	}
	adabasRequest.Limit = 1
	adabasRequest.Isn = isn
	err = request.adabas.readISN(request.repository.Fnr, adabasRequest, x)
	return
}

type stream struct {
	streamFunction StreamFunction
	result         *Response
	x              interface{}
}

func streamRecord(adabasRequest *adatypes.Request, x interface{}) (err error) {
	stream := x.(*stream)

	isn := adabasRequest.Isn
	isnQuantity := adabasRequest.IsnQuantity
	adatypes.Central.Log.Debugf("Got ISN %d record", isn)
	record, xerr := NewRecordIsn(isn, isnQuantity, adabasRequest.Definition)
	if xerr != nil {
		return xerr
	}
	err = stream.streamFunction(record, stream.x)
	return
}

// ReadLogicalWithStream read records with a logical order given by a search string and calls stream function
func (request *ReadRequest) ReadLogicalWithStream(search string, streamFunction StreamFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{streamFunction: streamFunction, result: &Response{Definition: request.definition}, x: x}
	err = request.ReadLogicalWithWithParser(search, streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return
}

// ReadLogicalWith read records with a logical order given by a search string
func (request *ReadRequest) ReadLogicalWith(search string) (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadLogicalWithWithParser(search, parseRead, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalWithWithParser read records with a logical order given by a search string
func (request *ReadRequest) ReadLogicalWithWithParser(search string, resultParser adatypes.RequestParser, x interface{}) (err error) {

	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		err = request.Open()
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Read logical, open done ...%#v", request.adabas.ID.platform)
		searchInfo := adatypes.NewSearchInfo(request.adabas.ID.platform(request.adabas.URL.String()), search)
		adatypes.Central.Log.Debugf("New search info ... %#v", searchInfo)
		var tree *adatypes.SearchTree
		if request.definition == nil {
			adatypes.Central.Log.Debugf("Load Definition ...")
			err = request.loadDefinition()
			if err != nil {
				return
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
		adatypes.Central.Log.Debugf("Definition generated ...")
		adabasRequest, prepareErr := request.prepareRequest()
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adatypes.Central.Log.Debugf("Prepare done ...")
		if resultParser == nil {
			adabasRequest.Parser = parseRead
		} else {
			adabasRequest.Parser = resultParser
		}
		adabasRequest.Limit = request.Limit
		searchInfo.Definition = adabasRequest.Definition
		adabasRequest.SearchTree = tree
		adabasRequest.Descriptors = tree.OrderBy()
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		if searchInfo.NeedSearch {
			adatypes.Central.Log.Debugf("search logical with ...%#v", adabasRequest.Descriptors)
			err = request.adabas.SearchLogicalWith(request.repository.Fnr, adabasRequest, x)
		} else {
			adatypes.Central.Log.Debugf("read logical with ...%#v", adabasRequest.Descriptors)
			err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
		}
	} else {
		request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}

	return
}

// ReadLogicalBy read in logical order given by the descriptor argument
func (request *ReadRequest) ReadLogicalBy(descriptors string) (result *Response, err error) {
	result = &Response{Definition: request.definition}
	err = request.ReadLogicalByWithParser(descriptors, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalByStream read records with a logical order given by a descriptor sort and calls stream function
func (request *ReadRequest) ReadLogicalByStream(descriptor string, streamFunction StreamFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{streamFunction: streamFunction, result: &Response{Definition: request.definition, fields: request.fields}, x: x}
	err = request.ReadLogicalByWithParser(descriptor, streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return
}

// ReadLogicalByWithParser read in logical order given by the descriptor argument
func (request *ReadRequest) ReadLogicalByWithParser(descriptors string, resultParser adatypes.RequestParser, x interface{}) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	if x == nil {
		err = adatypes.NewGenericError(23)
		return
	}
	adatypes.Central.Log.Debugf("Prepare read logical by request ... %s", descriptors)
	adabasRequest, prepareErr := request.prepareRequest()
	if prepareErr != nil {
		err = prepareErr
		return
	}
	adabasRequest.Multifetch = request.Multifetch
	if request.Limit != 0 && request.Limit < uint64(request.Multifetch) {
		adabasRequest.Multifetch = uint32(request.Limit)
	}
	if resultParser == nil {
		adabasRequest.Parser = parseRead
	} else {
		adabasRequest.Parser = resultParser
	}
	adabasRequest.Limit = request.Limit
	adabasRequest.Descriptors, err = request.definition.Descriptors(descriptors)
	if err != nil {
		return
	}

	adatypes.Central.Log.Debugf("Read logical by ...%d", request.repository.Fnr)
	err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
	return
}

// HistogramBy read a descriptor in a descriptor order
func (request *ReadRequest) HistogramBy(descriptor string) (result *Response, err error) {
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		err = request.Open()
		if err != nil {
			return
		}
		err = request.QueryFields(descriptor)
		if err != nil {
			return
		}
		adabasRequest, prepareErr := request.prepareRequest()
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adabasRequest.Parser = parseRead
		adabasRequest.Limit = request.Limit
		adabasRequest.Descriptors = []string{descriptor}
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		Response := &Response{Definition: request.definition, fields: request.fields}

		err = request.adabas.Histogram(request.repository.Fnr, adabasRequest, Response)
		if err == nil {
			result = Response
		}
	} else {
		result = &Response{Definition: request.definition, fields: request.fields}
		adatypes.Central.Log.Debugf("Read next chunk")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, result)
		adatypes.Central.Log.Debugf("Read next chunk done %v", err)
		request.cursoring.result = result
	}
	return
}

// HistogramWithStream read a descriptor given by a search criteria
func (request *ReadRequest) HistogramWithStream(search string, streamFunction StreamFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{streamFunction: streamFunction, result: &Response{Definition: request.definition}, x: x}
	err = request.histogramWithWithParser(search, streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return result, nil
}

// HistogramWith read a descriptor given by a search criteria
func (request *ReadRequest) HistogramWith(search string) (result *Response, err error) {
	response := &Response{Definition: request.definition, fields: request.fields}
	err = request.histogramWithWithParser(search, parseRead, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// histogramWithWithParser read a descriptor given by a search criteria
func (request *ReadRequest) histogramWithWithParser(search string, resultParser adatypes.RequestParser,
	x interface{}) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	if request.definition == nil {
		request.loadDefinition()
	}
	searchInfo := adatypes.NewSearchInfo(request.adabas.ID.platform(request.adabas.URL.String()), search)
	searchInfo.Definition = request.definition
	tree, err := searchInfo.GenerateTree()
	if err != nil {
		return
	}
	fields := tree.SearchFields()
	if len(fields) != 1 {
		err = adatypes.NewGenericError(24, len(fields))
		return
	}
	request.definition.ShouldRestrictToFieldSlice(fields)
	adatypes.Central.Log.Debugf("Before value creation --------")
	adabasRequest, prepareErr := request.prepareRequest()
	adatypes.Central.Log.Debugf("Prepare done --------")
	adabasRequest.SearchTree = tree
	adabasRequest.Descriptors = adabasRequest.SearchTree.OrderBy()
	if prepareErr != nil {
		err = prepareErr
		return
	}
	adabasRequest.Parser = resultParser
	adabasRequest.Limit = request.Limit

	err = request.adabas.Histogram(request.repository.Fnr, adabasRequest, x)
	return
}

type evaluateFieldMap struct {
	queryFields map[string]*queryField
	fields      map[string]int
}

func initFieldSubTypes(st *adatypes.StructureType, queryFields map[string]*queryField, current *int) {
	for _, sub := range st.SubTypes {
		if sub.IsStructure() {
			sst := sub.(*adatypes.StructureType)
			initFieldSubTypes(sst, queryFields, current)
		} else {
			adatypes.Central.Log.Debugf("Sub field %s = %d", sub.Name(), *current)
			queryFields[sub.Name()] = &queryField{field: sub.Name(), index: *current}
			*current++
		}
	}

}

func traverseFieldMap(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	ev := x.(*evaluateFieldMap)
	s := adaType.Name()
	if index, ok := ev.fields[s]; ok {
		if _, okq := ev.queryFields[s]; !okq {
			if adaType.IsStructure() {
				st := adaType.(*adatypes.StructureType)
				current := index
				initFieldSubTypes(st, ev.queryFields, &current)
				if current > index {
					d := current - index - 1
					for s, i := range ev.fields {
						if i > index {
							ev.fields[s] = i + d
						}
						adatypes.Central.Log.Debugf("New order %s -> %d", s, ev.fields[s])
					}
				}
			} else {
				ev.queryFields[s] = &queryField{field: s, index: index}
			}
		}
	}
	return nil
}

// QueryFields define the fields queried in that request
func (request *ReadRequest) QueryFields(fieldq string) (err error) {
	adatypes.Central.Log.Debugf("Query fields to %s", fieldq)
	err = request.Open()
	if err != nil {
		return
	}
	// // Could not recreate field content of a request!!!
	// if request.fields == nil {
	// 	if fieldq != "*" && fieldq != "" {
	// 		request.fields = make(map[string]*queryField)
	// 		adatypes.Central.Log.Debugf("Check Query field %s", fieldq)
	// 		for i, s := range strings.Split(fieldq, ",") {
	// 			adatypes.Central.Log.Debugf("Add Query field %s=%d", s, i)
	// 			request.fields[s] = &queryField{field: s, index: i}
	// 		}
	// 	} else {
	// 		adatypes.Central.Log.Debugf("General Query")
	// 	}
	// }

	err = request.loadDefinition()
	if err != nil {
		return
	}
	err = request.definition.ShouldRestrictToFields(fieldq)

	// Could not recreate field content of a request!!!
	if request.fields == nil {
		if fieldq != "*" && fieldq != "" {
			f := make(map[string]int)
			for i, s := range strings.Split(fieldq, ",") {
				f[s] = i
			}
			ev := &evaluateFieldMap{queryFields: make(map[string]*queryField), fields: f}
			tm := adatypes.NewTraverserMethods(traverseFieldMap)
			request.definition.TraverseTypes(tm, true, ev)
			request.fields = ev.queryFields
		}
	}

	return
}

func scanFieldsTraverser(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	sf := x.(*scanFields)
	adatypes.Central.Log.Debugf("Scan field of %s", adaValue.Type().Name())
	if f, ok := sf.fields[adaValue.Type().Name()]; ok {
		adatypes.Central.Log.Debugf("Part field: %s", adaValue.Type().Name())
		if f.index < len(sf.parameter) && sf.parameter[f.index] != nil {
			switch sf.parameter[f.index].(type) {
			case *int:
				v32, err := adaValue.Int32()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*int)) = int(v32)
			case *int32:
				v32, err := adaValue.Int32()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*int32)) = v32
			case *int64:
				v64, err := adaValue.Int64()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*int64)) = v64
			case *uint32:
				v32, err := adaValue.UInt32()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*uint32)) = v32
			case *uint64:
				v64, err := adaValue.UInt64()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*uint64)) = v64
			case *float32:
				v64, err := adaValue.Float()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*float32)) = float32(v64)
			case *float64:
				v64, err := adaValue.Float()
				if err != nil {
					return adatypes.EndTraverser, err
				}
				*(sf.parameter[f.index].(*float64)) = v64
			case *string:
				s := strings.Trim(adaValue.String(), " ")
				if *(sf.parameter[f.index].(*string)) == "" {
					*(sf.parameter[f.index].(*string)) = s
				} else {
					os := *(sf.parameter[f.index].(*string))
					*(sf.parameter[f.index].(*string)) = os + "," + s

				}
			case *[]string:
				s := strings.Trim(adaValue.String(), " ")
				x := sf.parameter[f.index].(*[]string)
				if adaValue.PeriodIndex() > 0 && uint32(len(*x)) >= adaValue.PeriodIndex() {
					if adaValue.MultipleIndex() == 1 {
						(*x)[adaValue.PeriodIndex()-1] = s
					} else {
						(*x)[adaValue.PeriodIndex()-1] = (*x)[adaValue.PeriodIndex()-1] + "," + s
					}
				} else {
					*x = append(*x, s)
				}
			default:
				return adatypes.EndTraverser, adatypes.NewGenericError(150, fmt.Sprintf("%T", sf.parameter[f.index]))
			}
		}
	}
	return adatypes.Continue, nil
}

// Scan scan for different field entries
func (request *ReadRequest) Scan(dest ...interface{}) error {
	if request.cursoring.HasNextRecord() {
		record, err := request.cursoring.NextRecord()
		if err != nil {
			return err
		}
		return record.Scan(request.fields, dest)
	}
	return adatypes.NewGenericError(130)
}
