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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const (
	maxReadRecordLimit = 20
	defaultBlockSize   = adatypes.PartialLobSize
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
	Start             uint64
	Limit             uint64
	Multifetch        uint32
	RecordBufferShift uint32
	fields            map[string]*queryField
	HoldRecords       adatypes.HoldType
	queryFunction     func(string, string) (*Response, error)
	cursoring         *Cursoring
	PartialRead       bool
	BlockSize         uint32
}

// NewReadRequest create a request defined by a dynamic list of parameters.
// This constructor is used internally. Use the `Connection` instance to
// use read requests to Adabas.
func NewReadRequest(param ...interface{}) (request *ReadRequest, err error) {
	if len(param) == 0 {
		return nil, errors.New("not enough parameters for NewReadRequest")
	}
	switch param[0].(type) {
	case *StoreRequest:
		sr := param[0].(*StoreRequest)
		request, err = createNewReadRequestCommon(&sr.commonRequest)
		if err != nil {
			return nil, err
		}
		request.commonRequest.adabasMap = sr.commonRequest.adabasMap
		request.commonRequest.MapName = sr.commonRequest.MapName
		request.commonRequest.dynamic = nil
		request.commonRequest.definition = adatypes.NewDefinitionClone(sr.commonRequest.definition)
		return
	case *commonRequest:
		cr := param[0].(*commonRequest)
		return createNewReadRequestCommon(cr)
	case *Adabas:
		ada := param[0].(*Adabas)
		switch p := param[1].(type) {
		case string:
			return createNewMapReadRequest(p, ada)
		case *Map:
			return createNewMapPointerReadRequest(ada, p)
		case Fnr:
			return createNewReadRequestAdabas(ada, p), nil
		case int:
			return createNewReadRequestAdabas(ada, Fnr(p)), nil
		}
	case string:
		mapName := param[0].(string)
		if len(param) < 2 {
			return nil, errors.New("not enough parameters for NewReadRequest")
		}
		ada := param[1].(*Adabas)
		if len(param) == 2 {
			return createNewMapReadRequest(mapName, ada)
		}
		rep := param[2].(*Repository)
		return createNewMapReadRequestRepo(mapName, ada, rep)
	default:
		ti := reflect.TypeOf(param[0])
		if ti.Kind() == reflect.Ptr {
			adatypes.Central.Log.Debugf("It's a pointer %s", ti.Name())
			ti = ti.Elem()
		}
		if ti.Kind() == reflect.Struct {
			adatypes.Central.Log.Debugf("It's a struct %s", ti.Name())
			mapName := ti.Name()
			if len(param) < 2 {
				return nil, errors.New("not enough parameters for NewReadRequest")
			}
			ada := param[1].(*Adabas)
			if len(param) == 2 {
				request, err = createNewMapReadRequest(mapName, ada)
			} else {
				rep := param[2].(*Repository)
				request, err = createNewMapReadRequestRepo(mapName, ada, rep)
			}
			if err != nil {
				adatypes.Central.Log.Debugf("Error creating dynamic read request: %v", err)
				return nil, err
			}
			adatypes.Central.Log.Debugf("Success creating dynamic read request: %v", param[0])
			request.createDynamic(param[0])
			adatypes.Central.Log.Debugf("Create dynamic %v", request.dynamic)
			return
		}
		adatypes.Central.Log.Errorf("Unknown request parameter: %v %T", reflect.TypeOf(param[0]).Kind(), param[0])
	}
	return nil, adatypes.NewGenericError(73)
}

// NewReadRequestCommon create a request defined by another request (not even ReadRequest required)
func createNewReadRequestCommon(commonRequest *commonRequest) (*ReadRequest, error) {
	request := &ReadRequest{HoldRecords: adatypes.HoldNone, PartialRead: false, BlockSize: defaultBlockSize}
	request.commonRequest = *commonRequest
	request.commonRequest.adabasMap = nil
	request.commonRequest.MapName = ""
	return request, nil
}

// createNewMapReadRequestRepo create a new Request instance
func createNewMapReadRequestRepo(mapName string, adabas *Adabas, repository *Repository) (request *ReadRequest, err error) {
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
		return nil, serr
	}
	dataAdabas, nerr := NewAdabas(&adabasMap.Data.URL, adabas.ID)
	if nerr != nil {
		return nil, nerr
	}
	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: adatypes.DefaultMultifetchLimit,
		BlockSize: defaultBlockSize, PartialRead: false,
		commonRequest: commonRequest{MapName: mapName, adabas: dataAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// createNewMapReadRequest create a new Request instance defined by a Adabas Map
// and a Adabas `instance`
func createNewMapReadRequest(mapName string, adabas *Adabas) (request *ReadRequest, err error) {
	var adabasMap *Map
	if adabas == nil {
		return nil, adatypes.NewGenericError(0)
	}
	// Search for map in repository
	adabasMap, _, err = SearchMapRepository(adabas.ID, mapName)
	if err != nil {
		return
	}
	adabas.SetURL(&adabasMap.Repository.URL)
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d", mapName, adabasMap.Data.Fnr)

	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: adatypes.DefaultMultifetchLimit,
		BlockSize: defaultBlockSize, PartialRead: false,
		commonRequest: commonRequest{MapName: mapName, adabas: adabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// createNewMapPointerReadRequest create a new Request instance
func createNewMapPointerReadRequest(adabas *Adabas, adabasMap *Map) (request *ReadRequest, err error) {
	if adabasMap == nil {
		err = adatypes.NewGenericError(22, "")
		return
	}
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d -> %#v", adabasMap.Name,
		adabasMap.Data.Fnr, adabas.ID.platform)
	cloneAdabas := NewClonedAdabas(adabas)

	dataRepository := NewMapRepository(adabas.URL, adabasMap.Data.Fnr)
	request = &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: adatypes.DefaultMultifetchLimit,
		BlockSize: defaultBlockSize,
		commonRequest: commonRequest{MapName: adabasMap.Name, adabas: cloneAdabas, adabasMap: adabasMap,
			repository: dataRepository, dynamic: adabasMap.dynamic}}
	return
}

// createNewReadRequestAdabas create a new Request instance using an `Adabas`
// instance and a file number
func createNewReadRequestAdabas(adabas *Adabas, fnr Fnr) *ReadRequest {
	clonedAdabas := NewClonedAdabas(adabas)

	return &ReadRequest{HoldRecords: adatypes.HoldNone, Limit: maxReadRecordLimit, Multifetch: adatypes.DefaultMultifetchLimit,
		BlockSize: defaultBlockSize, PartialRead: false,
		commonRequest: commonRequest{adabas: clonedAdabas,
			repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// Open call Adabas session and open a user queue entry in the database.
func (request *ReadRequest) Open() (opened bool, err error) {
	return request.commonOpen()
}

// Prepare read request for special parts in read
func (request *ReadRequest) prepareRequest(descriptorRead bool) (adabasRequest *adatypes.Request, err error) {
	if request.definition == nil {
		adatypes.Central.Log.Debugf("Prepare request creating definition")
		err = request.loadDefinition()
		if err != nil {
			return
		}
		if request.dynamic != nil {
			q := request.dynamic.CreateQueryFields()
			err = request.QueryFields(q)
			if err != nil {
				return
			}
		}
	}
	parameter := &adatypes.AdabasRequestParameter{Store: false,
		DescriptorRead: descriptorRead, SecondCall: 0,
		Mainframe: request.adabas.status.platform.IsMainframe(),
		BlockSize: request.BlockSize, PartialRead: request.PartialRead}
	adatypes.Central.Log.Debugf("Prepare request creating Adabas request")
	adabasRequest, err = request.definition.CreateAdabasRequest(parameter)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Prepare decriptor read %v", descriptorRead)
	adabasRequest.Definition = request.definition
	adabasRequest.PartialLobSize = request.BlockSize
	adabasRequest.RecordBufferShift = request.RecordBufferShift
	adatypes.Central.Log.Debugf("Partial LOB size: %d", adabasRequest.PartialLobSize)
	adatypes.Central.Log.Debugf("Record shift set to: %d", adabasRequest.RecordBufferShift)
	adabasRequest.HoldRecords = request.HoldRecords
	adabasRequest.Multifetch = request.Multifetch
	if request.Limit != 0 && request.Limit < uint64(request.Multifetch) {
		adabasRequest.Multifetch = uint32(request.Limit)
	}
	adatypes.Central.Log.Debugf("Got dynamic part: %v", request.dynamic)
	if request.dynamic != nil {
		adabasRequest.DataType = request.dynamic
	}

	return
}

// SetHoldRecords set hold record flag. All read operations will be done
// setting the record in hold for atomic reads and possible update operations
// afterwards.
func (request *ReadRequest) SetHoldRecords(hold adatypes.HoldType) {
	request.HoldRecords = hold
}

// parses the read record record
func parseReadToRecord(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Parse read to record")
	result := x.(*Response)

	if adabasRequest.Option.StreamCursor == 0 {
		isn := adabasRequest.Isn
		isnQuantity := adabasRequest.IsnQuantity
		record, xerr := NewRecordIsn(isn, isnQuantity, adabasRequest.Definition)
		if xerr != nil {
			return xerr
		}
		if adabasRequest.Parameter != nil {
			record.adabasMap = adabasRequest.Parameter.(*Map)
		}
		result.Values = append(result.Values, record)

		record.fields = result.fields
		adatypes.Central.Log.Debugf("Got ISN=%d Quantity=%d record", record.Isn, record.Quantity)
	}

	return
}

// parses the read record record
func parseReadToInterface(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Parse read to interface")
	result := x.(*Response)

	ti := adabasRequest.DataType.DataType
	if ti.Kind() == reflect.Ptr {
		ti = ti.Elem()
	}
	newInstance := reflect.New(ti)
	adatypes.Central.Log.Debugf("Kind: %v Elem: %v", reflect.TypeOf(newInstance).Kind(), newInstance.Elem())
	err = adabasRequest.Definition.AdaptInterfaceFields(newInstance, adabasRequest.DataType.FieldNames)
	if err != nil {
		return err
	}
	err = adabasRequest.DataType.ExamineIsnField(newInstance, adabasRequest.Isn)
	if err != nil {
		return err
	}
	// if f, ok := adabasRequest.DataType.FieldNames["#isn"]; ok {
	// 	isnField := newInstance.Elem().FieldByName(f[0])
	// 	if !isnField.IsValid() || isnField.Kind() != reflect.Uint64 {
	// 		return adatypes.NewGenericError(113)
	// 	}
	// 	isnField.SetUint(uint64(adabasRequest.Isn))
	// }
	adatypes.Central.Log.Debugf("Parse read to interface %v <%s> -> %d", newInstance, newInstance.String(), len(result.Data))
	result.Data = append(result.Data, newInstance.Interface())
	adatypes.Central.Log.Debugf("After read to interface %v", len(result.Data))

	return
}

// ReadPhysicalSequence the Adabas records will be read in physical order. The
// physical read is an I/O optimal read with physical order of the records.
func (request *ReadRequest) ReadPhysicalSequence() (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadPhysicalSequenceWithParser(nil, result)
	if err != nil {
		return nil, err
	}
	return
}

func (request *ReadRequest) readPhysical(search, descriptors string) (result *Response, err error) {
	return request.ReadPhysicalSequence()
}

// ReadPhysicalSequenceStream the Adabas records will be read in physical order. The
// physical read is an I/O optimal read with physical order of the records.
// For each read record a callback function defined by `streamFunction` will be called.
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

// ReadPhysicalInterface the Adabas records will be read in physical order. The
// physical read is an I/O optimal read with physical order of the records.
// For each read record a callback function defined by `interfaceFunction` will be called.
// This variant is used if dynamic interface are used.
func (request *ReadRequest) ReadPhysicalInterface(interfaceFunction InterfaceFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{interfaceFunction: interfaceFunction, result: &Response{Definition: request.definition}, x: x}
	err = request.ReadPhysicalSequenceWithParser(streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return
}

// ReadPhysicalSequenceWithParser read records in physical order using a
// special parser metod. This function will be removed in further version.
func (request *ReadRequest) ReadPhysicalSequenceWithParser(resultParser adatypes.RequestParser, x interface{}) (err error) {
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		_, err = request.Open()
		if err != nil {
			return
		}
		adabasRequest, prepareErr := request.prepareRequest(false)
		if prepareErr != nil {
			adatypes.Central.Log.Debugf("Prepare failed: %v", prepareErr)
			err = prepareErr
			return
		}
		switch {
		case resultParser != nil:
			adabasRequest.Parser = resultParser
		case adabasRequest.DataType != nil:
			adatypes.Central.Log.Debugf("Set parseReadToInterface for dynamic %v", adabasRequest.DataType)
			adabasRequest.Parser = parseReadToInterface
		default:
			adabasRequest.Parser = parseReadToRecord
		}
		adabasRequest.Limit = request.Limit
		adabasRequest.Multifetch = request.Multifetch
		if request.Limit != 0 && request.Limit < uint64(request.Multifetch) {
			adabasRequest.Multifetch = uint32(request.Limit)
		}
		adabasRequest.Parameter = request.adabasMap
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		err = request.adabas.ReadPhysical(request.repository.Fnr, adabasRequest, x)
	} else {
		adatypes.Central.Log.Debugf("read physical with ...cursoring")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}
	return
}

// ReadISN this method reads a records defined by a given ISN. Ths ISN may
// be read by an search query before.
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
	_, err = request.Open()
	if err != nil {
		return
	}
	adabasRequest, prepareErr := request.prepareRequest(false)
	if prepareErr != nil {
		err = prepareErr
		return
	}
	switch {
	case resultParser != nil:
		adabasRequest.Parser = resultParser
	case adabasRequest.DataType != nil:
		adabasRequest.Parser = parseReadToInterface
	default:
		adabasRequest.Parser = parseReadToRecord
	}
	adabasRequest.Limit = 1
	adabasRequest.Multifetch = 1
	adabasRequest.Isn = isn
	adabasRequest.Parameter = request.adabasMap
	err = request.adabas.readISN(request.repository.Fnr, adabasRequest, x)
	return
}

type stream struct {
	streamFunction    StreamFunction
	interfaceFunction InterfaceFunction
	result            *Response
	x                 interface{}
}

func streamRecord(adabasRequest *adatypes.Request, x interface{}) (err error) {
	stream := x.(*stream)
	switch {
	case stream.interfaceFunction != nil:
		// ti := reflect.TypeOf(adabasRequest.DataType)
		// if ti.Kind() == reflect.Ptr {
		// 	ti = ti.Elem()
		// }
		newInstance := reflect.New(adabasRequest.DataType.DataType)
		adatypes.Central.Log.Debugf("Kind: %v Elem: %v", reflect.TypeOf(newInstance).Kind(), newInstance.Elem())
		err = adabasRequest.Definition.AdaptInterfaceFields(newInstance, adabasRequest.DataType.FieldNames)
		if err != nil {
			return err
		}
		err = adabasRequest.DataType.ExamineIsnField(newInstance, adabasRequest.Isn)
		if err != nil {
			return err
		}
		adatypes.Central.Log.Debugf("Parse read calling interface function %v <%s>", newInstance, newInstance.String())
		err = stream.interfaceFunction(newInstance.Interface(), stream.x)
	case stream.streamFunction != nil:
		isn := adabasRequest.Isn
		isnQuantity := adabasRequest.IsnQuantity
		adatypes.Central.Log.Debugf("Got ISN %d record", isn)
		record, xerr := NewRecordIsn(isn, isnQuantity, adabasRequest.Definition)
		if xerr != nil {
			return xerr
		}
		err = stream.streamFunction(record, stream.x)
	default:
	}
	return
}

// ReadLogicalWithStream this method does an logical read using a search operation.
// The read records have  a logical order given by a search string. The result records will
// be provied to the callback stream function
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

// ReadLogicalWithInterface this method does an logical read using a search operation.
// The read records have  a logical order given by a search string. The result records will
// be provied to the callback interface function
func (request *ReadRequest) ReadLogicalWithInterface(search string, interfaceFunction InterfaceFunction,
	x interface{}) (result *Response, err error) {
	s := &stream{interfaceFunction: interfaceFunction, result: &Response{Definition: request.definition}, x: x}
	err = request.ReadLogicalWithWithParser(search, streamRecord, s)
	if err != nil {
		return nil, err
	}
	result = s.result
	return
}

// ReadLogicalWith this method does an logical read using a search operation.
// The read records have  a logical order given by a search string. The result records will
// be provied in the result `Response` structure value slice.
func (request *ReadRequest) ReadLogicalWith(search string) (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadLogicalWithWithParser(search, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

func (request *ReadRequest) readLogicalWith(search, descriptors string) (result *Response, err error) {
	return request.ReadLogicalWith(search)
}

func (request *ReadRequest) readLogicalBy(search, descriptors string) (result *Response, err error) {
	return request.ReadLogicalBy(descriptors)
}

// ReadByISN read records with a logical order given by a ISN sequence.
// The ISN is to be set by the `Start` `ReadRequest` parameter.
func (request *ReadRequest) ReadByISN() (result *Response, err error) {
	result = &Response{Definition: request.definition, fields: request.fields}
	err = request.ReadLogicalWithWithParser("", nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalWithWithParser read records with a logical order given by a search string.
// The given parser will parse the corresponding data.
func (request *ReadRequest) ReadLogicalWithWithParser(search string, resultParser adatypes.RequestParser, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Read logical with parser")
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
				if request.dynamic != nil {
					adatypes.Central.Log.Debugf("Dynamic query fields Definition ...")
					q := request.dynamic.CreateQueryFields()
					request.QueryFields(q)
				}
				adatypes.Central.Log.Debugf("Loaded Definition ...")
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
		}

		adatypes.Central.Log.Debugf("Definition generated ...")
		adabasRequest, prepareErr := request.prepareRequest(false)
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adatypes.Central.Log.Debugf("Read logical prepare done ...")
		switch {
		case resultParser != nil:
			adabasRequest.Parser = resultParser
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
		_ = request.adaptDescriptorMap(adabasRequest)
		// err = request.adaptDescriptorMap(adabasRequest)
		// if err != nil {
		// 	return err
		// }
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		if searchInfo == nil {
			adatypes.Central.Log.Debugf("read in ISN order ...from %d", request.Start)
			adabasRequest.Isn = adatypes.Isn(request.Start)
			err = request.adabas.ReadISNOrder(request.repository.Fnr, adabasRequest, x)
		} else {
			adabasRequest.Isn = 0
			if searchInfo.NeedSearch {
				adatypes.Central.Log.Debugf("search logical with ...%#v", adabasRequest.Descriptors)
				err = request.adabas.SearchLogicalWith(request.repository.Fnr, adabasRequest, x)
			} else {
				adatypes.Central.Log.Debugf("read logical with ...%#v", adabasRequest.Descriptors)
				err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
			}
		}
	} else {
		adatypes.Central.Log.Debugf("read logical with ...cursoring")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}
	adatypes.Central.Log.Debugf("Read finished")
	return
}

// ReadLogicalBy this method read Adabas records in logical order given by the descriptor argument.
// The access in the database will be reduce to I/O to the ASSO container.
func (request *ReadRequest) ReadLogicalBy(descriptors string) (result *Response, err error) {
	result = &Response{Definition: request.definition}
	err = request.ReadLogicalByWithParser(descriptors, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalByStream this method read Adabas records in logical order given by the descriptor argument.
// The access in the database will NOT be reduce to I/O to the ASSO container.
// The result set will be called using the `streamFunction` method given.
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
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		_, err = request.Open()
		if err != nil {
			return
		}
		if x == nil {
			err = adatypes.NewGenericError(23)
			return
		}
		adatypes.Central.Log.Debugf("Prepare read logical by request ... %s", descriptors)
		adabasRequest, prepareErr := request.prepareRequest(false)
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adabasRequest.Multifetch = request.Multifetch
		if request.Limit != 0 && request.Limit < uint64(request.Multifetch) {
			adabasRequest.Multifetch = uint32(request.Limit)
		}
		switch {
		case resultParser != nil:
			adabasRequest.Parser = resultParser
		case adabasRequest.DataType != nil:
			adabasRequest.Parser = parseReadToInterface
		default:
			adabasRequest.Parser = parseReadToRecord
		}
		adabasRequest.Limit = request.Limit
		adabasRequest.Descriptors, err = request.definition.Descriptors(descriptors)
		if err != nil {
			return
		}
		adabasRequest.Parameter = request.adabasMap
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		adatypes.Central.Log.Debugf("Read logical by ...%d", request.repository.Fnr)
		err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
	} else {
		adatypes.Central.Log.Debugf("read logical by ...cursoring")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}
	return
}

// HistogramBy this method read Adabas records in logical order given by the descriptor argument.
// The access in the database will be reduce to I/O to the ASSO container.
func (request *ReadRequest) HistogramBy(descriptor string) (result *Response, err error) {
	adatypes.Central.Log.Debugf("Read histogram of %s", descriptor)
	// Check if cursoring uninitialized or normal first call then enter
	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		_, err = request.Open()
		if err != nil {
			return
		}
		err = request.QueryFields(descriptor)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Prepare histogram read")
		adabasRequest, prepareErr := request.prepareRequest(true)
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adatypes.Central.Log.Debugf("Prepared histogram read")
		adabasRequest.DescriptorRead = true
		adabasRequest.Parser = parseReadToRecord
		adabasRequest.Limit = request.Limit
		adabasRequest.Descriptors, err = request.definition.Descriptors(descriptor)
		if err != nil {
			return
		}
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}
		adabasRequest.Parameter = request.adabasMap

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

func (request *ReadRequest) histogramBy(search, descriptor string) (result *Response, err error) {
	return request.HistogramBy(descriptor)
}

// HistogramWithStream this method read Adabas records in logical order given by the descriptor-based
// search argument.
// The access in the database will be reduce to I/O to the ASSO container.
// The result set will be called using the `streamFunction` method given.
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

// HistogramWith this method read Adabas records in logical order given by the descriptor-based
// search argument.
// The access in the database will be reduce to I/O to the ASSO container.
func (request *ReadRequest) HistogramWith(search string) (result *Response, err error) {
	response := &Response{Definition: request.definition, fields: request.fields}
	err = request.histogramWithWithParser(search, parseReadToRecord, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (request *ReadRequest) histogramWith(search, descriptors string) (result *Response, err error) {
	return request.HistogramWith(search)
}

// histogramWithWithParser read a descriptor given by a search criteria
func (request *ReadRequest) histogramWithWithParser(search string, resultParser adatypes.RequestParser,
	x interface{}) (err error) {

	if request.cursoring == nil || request.cursoring.adabasRequest == nil {
		_, err = request.Open()
		if err != nil {
			return
		}
		if request.definition == nil {
			_ = request.loadDefinition()
			// err = request.loadDefinition()
			// if err != nil {
			// 	return err
			// }
		}
		searchInfo := adatypes.NewSearchInfo(request.adabas.ID.platform(request.adabas.URL.String()), search)
		searchInfo.Definition = request.definition
		tree, gerr := searchInfo.GenerateTree()
		if gerr != nil {
			err = gerr
			return
		}
		fields := tree.SearchFields()
		if len(fields) != 1 {
			adatypes.Central.Log.Debugf("Fields: %v", fields)
			err = adatypes.NewGenericError(24, len(fields))
			return
		}
		err = request.definition.ShouldRestrictToFieldSlice(fields)
		if err != nil {
			return err
		}
		adatypes.Central.Log.Debugf("Before value creation --------")
		adabasRequest, prepareErr := request.prepareRequest(true)
		adatypes.Central.Log.Debugf("Prepare done --------")
		adabasRequest.SearchTree = tree
		adabasRequest.Descriptors = adabasRequest.SearchTree.OrderBy()
		_ = request.adaptDescriptorMap(adabasRequest)
		// err = request.adaptDescriptorMap(adabasRequest)
		// if err != nil {
		// 	return
		// }
		if prepareErr != nil {
			err = prepareErr
			return
		}
		adabasRequest.Parser = resultParser
		adabasRequest.Limit = request.Limit
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		err = request.adabas.Histogram(request.repository.Fnr, adabasRequest, x)
	} else {
		adatypes.Central.Log.Debugf("read histograp with ...cursoring")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}
	return
}

// SearchAndOrder performs a search call and orders the result using defined descriptors
//  A search term will
// be used to search for and a descriptor defines the final result order.
func (request *ReadRequest) SearchAndOrder(search, descriptors string) (result *Response, err error) {
	result = &Response{Definition: request.definition}
	err = request.SearchAndOrderWithParser(search, descriptors, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// SearchAndOrderWithParser search and order with parser. A search term will
// be used to search for and a descriptor defines the final result order.
func (request *ReadRequest) SearchAndOrderWithParser(search, descriptors string, resultParser adatypes.RequestParser, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Search and order with descriptors")
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
		adatypes.Central.Log.Debugf("Search logical, open done ...%#v with search=%s", request.adabas.ID.platform, search)
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
				if request.dynamic != nil {
					adatypes.Central.Log.Debugf("Dynamic query fields Definition ...")
					q := request.dynamic.CreateQueryFields()
					request.QueryFields(q)
				}
				adatypes.Central.Log.Debugf("Loaded Definition ...")
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
		}

		adatypes.Central.Log.Debugf("Definition generated ...")
		adabasRequest, prepareErr := request.prepareRequest(false)
		if prepareErr != nil {
			err = prepareErr
			return
		}

		adatypes.Central.Log.Debugf("Search prepare done ... request is %#v", adabasRequest)
		switch {
		case resultParser != nil:
			adabasRequest.Parser = resultParser
		case adabasRequest.DataType != nil:
			adabasRequest.Parser = parseReadToInterface
		default:
			adabasRequest.Parser = parseReadToRecord
		}
		adabasRequest.Limit = request.Limit
		//searchInfo.Definition = adabasRequest.Definition
		if descriptors != "" {
			adabasRequest.Descriptors, err = request.definition.Descriptors(descriptors)
			if err != nil {
				return
			}
		}
		if tree != nil {
			adabasRequest.SearchTree = tree
			if descriptors == "" {
				adabasRequest.Descriptors = tree.OrderBy()
			}
		}
		err = request.adaptDescriptorMap(adabasRequest)
		adatypes.Central.Log.Errorf("Adapt Descriptor map error: %v", err)
		// if err != nil {
		// 	return err
		// }
		if request.cursoring != nil {
			request.cursoring.adabasRequest = adabasRequest
		}

		if searchInfo == nil && descriptors == "" {
			adatypes.Central.Log.Debugf("read in ISN order ...from %d", request.Start)
			adabasRequest.Isn = adatypes.Isn(request.Start)
			err = request.adabas.ReadISNOrder(request.repository.Fnr, adabasRequest, x)
		} else {
			adabasRequest.Isn = 0

			if searchInfo != nil && searchInfo.NeedSearch || searchInfo != nil && descriptors != "" {
				adatypes.Central.Log.Debugf("search logical with ...%#v", adabasRequest.Descriptors)
				err = request.adabas.SearchLogicalWith(request.repository.Fnr, adabasRequest, x)
			} else {
				adatypes.Central.Log.Debugf("read logical with ...%#v", adabasRequest.Descriptors)
				err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
			}
		}
	} else {
		adatypes.Central.Log.Debugf("read logical with ...cursoring")
		err = request.adabas.loopCall(request.cursoring.adabasRequest, x)
	}
	adatypes.Central.Log.Debugf("Read finished")
	return
}

func (request *ReadRequest) adaptDescriptorMap(adabasRequest *adatypes.Request) error {
	if request.adabasMap != nil {
		adabasRequest.Parameter = request.adabasMap
		for i := 0; i < len(adabasRequest.Descriptors); i++ {
			t, err := request.definition.SearchType(adabasRequest.Descriptors[i])
			if err != nil {
				return err
			}
			adatypes.Central.Log.Debugf("Found search descriptor %s and got %#v", adabasRequest.Descriptors[i], t)
			if t == nil {
				if adatypes.Central.IsDebugLevel() {
					request.definition.DumpTypes(false, false, "Global Tree")
					request.definition.DumpTypes(false, false, "Active Tree")
				}
				return adatypes.NewGenericError(76)
			}
			adabasRequest.Descriptors[i] = t.ShortName()
		}
	}
	return nil
}

type evaluateFieldMap struct {
	queryFields map[string]*queryField
	fields      map[string]int
	definition  *adatypes.Definition
}

// initFieldSubTypes initialize field sub types.
func initFieldSubTypes(st *adatypes.StructureType, queryFields map[string]*queryField, current *int) {
	adatypes.Central.Log.Debugf("Init field sub types")
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

// traverseFieldMap traverse through field ,aps checking sub types
func traverseFieldMap(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	ev := x.(*evaluateFieldMap)
	s := adaType.Name()
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Traverse field map, search: %s", s)
	}
	if index, ok := ev.fields[s]; ok {
		if _, okq := ev.queryFields[s]; !okq {
			if adatypes.Central.IsDebugLevel() {
				adatypes.Central.Log.Debugf("Check field map, search: %s", s)
			}
			switch {
			case adaType.Type() == adatypes.FieldTypeSuperDesc:
				if adatypes.Central.IsDebugLevel() {
					adatypes.Central.Log.Debugf("Adapt subtypes: %s", s)
				}
				superType := adaType.(*adatypes.AdaSuperType)
				err := superType.InitSubTypes(ev.definition)
				if err != nil {
					return err
				}
			case adaType.IsStructure() && adaType.Type() != adatypes.FieldTypeRedefinition:
				if adatypes.Central.IsDebugLevel() {
					adatypes.Central.Log.Debugf("Adapt subtypes: %s", s)
				}
				st := adaType.(*adatypes.StructureType)
				current := index
				initFieldSubTypes(st, ev.queryFields, &current)
				if current > index {
					d := current - index - 1
					for s, i := range ev.fields {
						if i > index {
							ev.fields[s] = i + d
						}
						if adatypes.Central.IsDebugLevel() {
							adatypes.Central.Log.Debugf("New order %s -> %d", s, ev.fields[s])
						}
					}
				}
			default:
				ev.queryFields[s] = &queryField{field: s, index: index}
			}
		}
	}
	return nil
}

// QueryFields this method define the set of fields which are part of
// the query. It restrict the fields and tree to the needed set.
// If the parameter is set to "*" all fields are part of the request.
// If the parameter is set to "" no field is returned and only the
// ISN an quantity information are provided.
// Following fields are keywords: #isn, #ISN, #key
// Fields started with '#' will provide only field data length information.
func (request *ReadRequest) QueryFields(fieldq string) (err error) {
	if request.dynamic != nil {
		adatypes.Central.Log.Debugf("Query fields of dynamic interface")
		if fieldq == "*" {
			fieldq = request.dynamic.CreateQueryFields()
		}
	} else {
		adatypes.Central.Log.Debugf("Query fields NO dynamic interface")
	}
	adatypes.Central.Log.Debugf("Query fields to %s", fieldq)
	_, err = request.Open()
	if err != nil {
		adatypes.Central.Log.Debugf("Query fields open error: %v", err)
		return
	}

	err = request.loadDefinition()
	if err != nil {
		adatypes.Central.Log.Debugf("Query fields definition error: %v", err)
		return
	}
	request.definition.ResetRestrictToFields()
	if fieldq == "*" {
		err = request.definition.RemoveSpecialDescriptors()
		if err != nil {
			return err
		}
	} else {
		if !(request.adabasMap != nil && fieldq == "*") {
			err = request.definition.ShouldRestrictToFields(fieldq)
			if err != nil {
				adatypes.Central.Log.Debugf("Query fields restrict error: %v", err)
				return err
			}
		}
	}

	// Could not recreate field content of a request!!!
	if request.fields == nil {
		adatypes.Central.Log.Debugf("Create field content")
		if fieldq != "*" && fieldq != "" {
			f := make(map[string]int)
			ev := &evaluateFieldMap{queryFields: make(map[string]*queryField), definition: request.definition, fields: f}
			for i, s := range strings.Split(fieldq, ",") {
				sl := strings.ToLower(s)
				if sl == "#isn" || sl == "#isnquantity" {
					ev.queryFields[sl] = &queryField{field: sl, index: i}
				} else {
					f[s] = i
				}
			}
			tm := adatypes.NewTraverserMethods(traverseFieldMap)
			err = request.definition.TraverseTypes(tm, true, ev)
			if err != nil {
				return err
			}
			request.fields = ev.queryFields
		}
	}

	adatypes.Central.Log.Debugf("Query fields ready %v", request.fields)
	return
}

// scanFieldsTraverser create parameter list of values
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

// Scan this method can be used to scan a number of parameters filled
// by values in the result set. See README.md explanation
func (request *ReadRequest) Scan(dest ...interface{}) error {
	if request.cursoring.HasNextRecord() {
		record, err := request.cursoring.NextRecord()
		if err != nil {
			return err
		}
		return record.Scan(dest...)
	}
	return adatypes.NewGenericError(130)
}
