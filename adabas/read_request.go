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
	"strconv"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const (
	maxReadRecordLimit     = 20
	defaultMultifetchLimit = 10
)

// ReadRequest request instance handling field query information
type ReadRequest struct {
	commonRequest
	Limit             uint64
	Multifetch        uint32
	RecordBufferShift uint32
}

// NewReadRequestCommon create a request defined by another request (not even ReadRequest required)
func NewReadRequestCommon(commonRequest *commonRequest) *ReadRequest {
	request := &ReadRequest{}
	request.commonRequest = *commonRequest
	request.commonRequest.adabasMap = nil
	request.commonRequest.mapName = ""
	return request
}

/*
// NewMapRequest create a new Request instance based on the map name and the url and file number
// should not be used in the code because the new Adabas instance need to be created
func NewMapRequest(mapName string, url string, fnr uint32) (*ReadRequest, error) {
	adatypes.Central.Log.Debugf("Create new request using map name %s and reference to file %s/%d", mapName, url, fnr)
	var adabas *Adabas
	if dbid, err := strconv.Atoi(url); err == nil {
		adabas = NewAdabas(Dbid(dbid))
	} else {
		return nil, err
	}
	mapRepository := NewMapRepository(adabas, fnr)
	adabasMap, err := mapRepository.SearchMap(mapName)
	if err != nil {
		return nil, err
	}
	return &ReadRequest{Limit: maxReadRecordLimit, commonRequest: commonRequest{mapName: mapName, adabas: adabas,
		adabasMap: adabasMap, repository: adabasMap.dataRepository}}, nil
}*/

// NewMapNameRequestRepo create a new Request instance
func NewMapNameRequestRepo(mapName string, adabas *Adabas, repository *Repository) (request *ReadRequest, err error) {
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
	dataAdabas := NewAdabasWithURL(&adabasMap.Data.URL, adabas.ID)
	dataRepository := NewMapRepository(adabas, adabasMap.Data.Fnr)
	request = &ReadRequest{Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: mapName, adabas: dataAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewMapNameRequest create a new Request instance
func NewMapNameRequest(adabas *Adabas, mapName string) (request *ReadRequest, err error) {
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
	request = &ReadRequest{Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: mapName, adabas: adabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewAdabasMapNameRequest create a new Request instance
func NewAdabasMapNameRequest(adabas *Adabas, adabasMap *Map) (request *ReadRequest, err error) {
	if adabasMap == nil {
		err = adatypes.NewGenericError(22, adabasMap.Name)
		return
	}
	adatypes.Central.Log.Debugf("Read: Adabas new map reference for %s to %d -> %#v", adabasMap.Name,
		adabasMap.Data.Fnr, adabas.ID.platform)
	cloneAdabas := NewClonedAdabas(adabas)

	dataRepository := NewMapRepository(adabas, adabasMap.Data.Fnr)
	request = &ReadRequest{Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{mapName: adabasMap.Name, adabas: cloneAdabas, adabasMap: adabasMap,
			repository: dataRepository}}
	return
}

// NewRequest create a new Request instance
func NewRequest(url string, fnr uint32) *ReadRequest {
	var adabas *Adabas
	if dbid, err := strconv.Atoi(url); err == nil {
		adabas = NewAdabas(Dbid(dbid))
	} else {
		return nil
	}
	return &ReadRequest{Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
		commonRequest: commonRequest{adabas: adabas,
			repository: &Repository{DatabaseURL: DatabaseURL{Fnr: fnr}}}}
}

// NewRequestAdabas create a new Request instance
func NewRequestAdabas(adabas *Adabas, fnr uint32) *ReadRequest {
	clonedAdabas := NewClonedAdabas(adabas)

	return &ReadRequest{Limit: maxReadRecordLimit, Multifetch: defaultMultifetchLimit,
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
	adabasRequest, err = request.definition.CreateAdabasRequest(false, false)
	if err != nil {
		return
	}
	adabasRequest.Definition = request.definition
	adabasRequest.RecordBufferShift = request.RecordBufferShift
	return
}

func parseRead(adabasRequest *adatypes.Request, x interface{}) (err error) {
	result := x.(*Response)

	isn := adabasRequest.Isn
	isnQuantity := adabasRequest.IsnQuantity
	adatypes.Central.Log.Debugf("Got ISN %d record", isn)
	Record, xerr := NewRecordIsn(isn, isnQuantity, adabasRequest.Definition)
	if xerr != nil {
		return xerr
	}
	result.Values = append(result.Values, Record)
	return
}

// ReadPhysicalSequence read records in physical order
func (request *ReadRequest) ReadPhysicalSequence() (result *Response, err error) {
	result = &Response{}
	err = request.ReadPhysicalSequenceWithParser(nil, result)
	if err != nil {
		return nil, err
	}
	return
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
	result = &Response{}
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

// ReadLogicalWith read records with a logical order given by a search string
func (request *ReadRequest) ReadLogicalWith(search string) (result *Response, err error) {
	result = &Response{}
	err = request.ReadLogicalWithWithParser(search, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalWithWithParser read records with a logical order given by a search string
func (request *ReadRequest) ReadLogicalWithWithParser(search string, resultParser adatypes.RequestParser, x interface{}) (err error) {
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
		// fields := tree.SearchFields()
		// if len(fields) != 1 {
		// 	err = fmt.Errorf("Please use histogram with single fields only. Nr of fields given is %d", len(fields))
		// 	return
		// }
		// request.definition.ShouldRestrictToFieldSlice(fields)
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

	if searchInfo.NeedSearch {
		adatypes.Central.Log.Debugf("search logical with ...%#v", adabasRequest.Descriptors)
		err = request.adabas.SearchLogicalWith(request.repository.Fnr, adabasRequest, x)
	} else {
		adatypes.Central.Log.Debugf("read logical with ...%#v", adabasRequest.Descriptors)
		err = request.adabas.ReadLogicalWith(request.repository.Fnr, adabasRequest, x)
	}

	return
}

// ReadLogicalBy read in logical order given by the descriptor argument
func (request *ReadRequest) ReadLogicalBy(descriptors string) (result *Response, err error) {
	result = &Response{}
	err = request.ReadLogicalByWithParser(descriptors, nil, result)
	if err != nil {
		return nil, err
	}
	return
}

// ReadLogicalByWithParser read in logical order given by the descriptor argument
func (request *ReadRequest) ReadLogicalByWithParser(descriptors string, resultParser adatypes.RequestParser, x interface{}) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	if x == nil {
		// x = &Response{}
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
	if request.Limit < uint64(request.Multifetch) {
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

	Response := &Response{}

	err = request.adabas.Histogram(request.repository.Fnr, adabasRequest, Response)
	if err == nil {
		result = Response
	}
	return
}

// HistogramWith read a descriptor given by a search criteria
func (request *ReadRequest) HistogramWith(search string) (result *Response, err error) {
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
	adabasRequest.Parser = parseRead
	adabasRequest.Limit = request.Limit

	Response := &Response{}

	err = request.adabas.Histogram(request.repository.Fnr, adabasRequest, Response)
	if err == nil {
		result = Response
	}
	return
}

// QueryFields define the fields queried in that request
func (request *ReadRequest) QueryFields(fields string) (err error) {
	err = request.Open()
	if err != nil {
		return
	}
	err = request.loadDefinition()
	if err != nil {
		return
	}
	err = request.definition.ShouldRestrictToFields(fields)
	return
}
