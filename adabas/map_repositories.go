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
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// DatabaseURL defines the target URL of a database file. Might be a database data file or a map repository
type DatabaseURL struct {
	URL URL `json:"Target"`
	Fnr Fnr `json:"File"`
}

// dbid reply the database id
func (repURL *DatabaseURL) dbid() (dbid Dbid, err error) {
	adatypes.Central.Log.Debugf("repURL=%#v", repURL)
	dbid = repURL.URL.Dbid
	return
}

// map name flags contains map related ISN and if found in search
type mapNameFlags struct {
	isn   adatypes.Isn
	found bool
}

// Repository Adabas Map repository container
type Repository struct {
	sync.Mutex
	DatabaseURL
	online     bool
	mapNames   map[string]*mapNameFlags
	cacheLock  *sync.Mutex
	cachedMaps map[string]*Map
	cacheTime  time.Time
}

func init() {
	queryMaps := os.Getenv("QUERY_MAPFILES")
	adatypes.Central.Log.Debugf("QUERY_MAPFILES: %s" + queryMaps)
	var re = regexp.MustCompile(`(?m)\(((\d+|\d+\(.*\)),\d+)\)`)

	for _, match := range re.FindAllStringSubmatch(queryMaps, -1) {
		adatypes.Central.Log.Debugf("Add to global repository search: %s", match[1])
		err := AddGlobalMapRepositoryReference(match[1])
		if err != nil {
			adatypes.Central.Log.Debugf("Error adding global map %v", err)
		}
	}
}

func evaluateURL(i interface{}) *URL {
	var url *URL
	switch i.(type) {
	case *Adabas:
		a := i.(*Adabas)
		url = a.URL
	case *URL:
		url = i.(*URL)
	default:
		return nil
	}
	return url
}

// NewMapRepository new map repository created
func NewMapRepository(i interface{}, fnr Fnr) *Repository {
	url := evaluateURL(i)
	mr := &Repository{DatabaseURL: DatabaseURL{URL: *url, Fnr: fnr}, online: true}
	mr.cacheLock = &sync.Mutex{}
	mr.cachedMaps = make(map[string]*Map)
	return mr
}

// NewMapRepositoryWithURL new map repository created
func NewMapRepositoryWithURL(url DatabaseURL) *Repository {
	mr := &Repository{DatabaseURL: url, online: true}
	mr.cacheLock = &sync.Mutex{}
	mr.cachedMaps = make(map[string]*Map)
	return mr
}

// AddMapToCache add map to cache
func (repository *Repository) AddMapToCache(name string, adabasMap *Map) {
	repository.cacheLock.Lock()
	defer repository.cacheLock.Unlock()

	repository.cachedMaps[name] = adabasMap
}

// GetMapFromCache get map from cache
func (repository *Repository) GetMapFromCache(name string) (*Map, bool) {
	repository.cacheLock.Lock()
	defer repository.cacheLock.Unlock()

	m, ok := repository.cachedMaps[name]
	return m, ok
}

// SearchMapInRepository search map name in specific map repository
func (repository *Repository) SearchMapInRepository(adabas *Adabas, mapName string) (adabasMap *Map, err error) {
	adatypes.Central.Log.Debugf("Map repository: %#v", repository)
	var dbid Dbid
	dbid, err = repository.DatabaseURL.dbid()
	if err != nil {
		fmt.Printf("Error dbid %v\n", err)
		return
	}
	if adabas == nil {
		err = adatypes.NewGenericError(48)
		return
	}
	adatypes.Central.Log.Debugf("Database id %d got map repository dbid %d", adabas.Acbx.Acbxdbid, dbid)
	if adabas.Acbx.Acbxdbid == 0 && adabas.Acbx.Acbxdbid != dbid {
		adabas.Close()
		adabas.Acbx.Acbxdbid = dbid
		adatypes.Central.Log.Debugf("Set new dbid after map %s to %d", mapName, dbid)
		adatypes.Central.Log.Debugf("Call search in repository using Adabas %s/%03d", adabas.URL.String(), adabas.Acbx.Acbxfnr)
	}
	adatypes.Central.Log.Debugf("Load repository of %s/%d", repository.URL.String(), repository.Fnr)
	err = repository.LoadMapRepository(adabas)
	if err != nil {
		adatypes.Central.Log.Debugf("Error loading repository %v\n", err)
		return
	}
	adatypes.Central.Log.Debugf("Read map in repository of %s/%d", repository.URL.String(), repository.Fnr)
	adabasMap, err = repository.readAdabasMap(adabas, mapName)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Found map <%s> in repository of %s/%d", adabasMap.Name, repository.URL.String(), repository.Fnr)
	repository.AddMapToCache(mapName, adabasMap)
	return
}

// readAdabasMapWithRequest read specific Adabas Map defined by the map name and the repository
func (repository *Repository) readAdabasMapWithRequest(commonRequest *commonRequest, name string) (adabasMap *Map, err error) {
	if repository == nil {
		//debug.PrintStack()
		err = adatypes.NewGenericError(5)
		return
	}
	adatypes.Central.Log.Debugf("Before Prepare Repository %#v\n", repository)
	url := repository.DatabaseURL
	adabasMap = NewAdabasMap(&repository.DatabaseURL, &url)
	request, _ := NewReadRequest(commonRequest)
	// Reset map definition, because
	request.commonRequest.adabasMap = nil
	request.commonRequest.MapName = ""

	adatypes.Central.Log.Debugf("Search for Map with name=%s", name)
	// Search for map name in database
	err = request.ReadLogicalWithWithParser(mapFieldName.fieldName()+"="+name, parseMap, adabasMap)
	if err != nil {
		return nil, err
	}
	if adabasMap.Name == "" {
		return nil, adatypes.NewGenericError(66, name)
	}
	adatypes.Central.Log.Debugf("Got Adabas map %s", adabasMap.Name)
	adabasMap.createFieldMap()
	repository.AddMapToCache(name, adabasMap)
	var dbid Dbid
	adatypes.Central.Log.Debugf("After Repository %#v\n", repository)
	if adabasMap.Repository.URL.Dbid == 0 {
		err = adatypes.NewGenericError(18)
		return
	}
	dbid, err = adabasMap.Repository.dbid()
	adatypes.Central.Log.Debugf("Set data reference to map database %d before %d", dbid, request.adabas.Acbx.Acbxdbid)
	if dbid > 0 {
		request.adabas.Acbx.Acbxdbid = dbid
	}

	// Reset definition because Adabas Map is loaded and not needed any more
	adatypes.Central.Log.Debugf("Loaded map in repository of %s/%d", repository.URL.String(), repository.Fnr)
	request.definition = nil
	return
}

// readAdabasMap read Adabas map defined by repository and name
func (repository *Repository) readAdabasMap(adabas *Adabas, name string) (adabasMap *Map, err error) {
	request, _ := NewReadRequest(adabas, repository.Fnr)
	adatypes.Central.Log.Debugf("Read map %s in repository %#v\n", name, repository)
	adabasMap, err = repository.readAdabasMapWithRequest(&request.commonRequest, name)
	return
}

// SearchMap search map name in specific map repository
func (repository *Repository) SearchMap(adabas *Adabas, mapName string) (adabasMap *Map, err error) {
	adatypes.Central.Log.Debugf("Search map %s in repository", mapName)
	if repository.mapNames == nil {
		err = repository.LoadMapRepository(adabas)
		if err != nil {
			return
		}
	}
	repository.Lock()
	if _, ok := repository.mapNames[mapName]; !ok {
		err = adatypes.NewGenericError(14, mapName)
		repository.Unlock()
		return
	}
	repository.Unlock()
	// Need a Adabas instance to work with corresponding ID, else return error
	if adabas == nil {
		return nil, adatypes.NewGenericError(64)
	}
	adatypes.Central.Log.Debugf("Search map in cache: %s", mapName)
	if m, ok := repository.GetMapFromCache(mapName); ok {
		adatypes.Central.Log.Debugf("Found map in cache: %s", mapName)
		adabasMap = m
		return
	}

	adatypes.Central.Log.Debugf("Not found in cache read map: %s", mapName)
	request, _ := NewReadRequest(adabas, repository.Fnr)
	request.Limit = 0
	err = request.ReadLogicalWithWithParser(mapFieldName.fieldName()+"="+mapName, parseMaps, repository)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read map repository searching %s", mapName)
	if m, ok := repository.GetMapFromCache(mapName); ok {
		// adabasMap, err = repository.readAdabasMap(adabas, mapName)
		// if err != nil {
		// 	return nil, err
		// }
		// } else {
		adabasMap = m
		// 	adatypes.Central.Log.Debugf("Found map with ISN=%v", adabasMap.Isn)
	} else {
		return nil, adatypes.NewGenericError(82, mapName)
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Got map adabas to %s/%d", adabasMap.Repository.URL.String(), adabasMap.Repository.Fnr)
		adatypes.Central.Log.Debugf("with data adabas to %s/%d", adabasMap.Data.URL.String(), adabasMap.Data.Fnr)
		adatypes.Central.Log.Debugf("out of %s/%d", repository.URL.String(), repository.Fnr)
	}
	return
}

// ClearCache clear cache if time frame occur
func (repository *Repository) ClearCache(maxTime time.Time) {
	if repository.cacheTime.Before(maxTime) {
		adatypes.Central.Log.Debugf("Clear caching ... %v -> %v", maxTime, time.Now())
		repository.cachedMaps = make(map[string]*Map)
	}
}

// LoadAllMaps load all map out of specific map repository
func (repository *Repository) LoadAllMaps(adabas *Adabas) (adabasMaps []*Map, err error) {
	if repository.mapNames == nil {
		err = repository.LoadMapRepository(adabas)
		if err != nil {
			return
		}
	}
	if adabas == nil {
		// adabas, err = NewAdabass(repository.URL.String())
		// if err != nil {
		// 	return
		// }
		return nil, adatypes.NewGenericError(64)
	}
	adatypes.Central.Log.Debugf("Load all maps")
	request, _ := NewReadRequest(adabas, repository.Fnr)
	request.Limit = 0
	err = request.ReadPhysicalSequenceWithParser(parseMaps, repository)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read all map entries in repository")
	repository.cacheLock.Lock()
	defer repository.cacheLock.Unlock()
	for _, m := range repository.cachedMaps {
		adabasMaps = append(adabasMaps, m)
		adatypes.Central.Log.Debugf("Got map %s adabas to %s/%d", m.Name, m.Repository.URL.String(), m.Repository.Fnr)
		adatypes.Central.Log.Debugf("with data adabas to %s/%d", m.Data.URL.String(), m.Data.Fnr)
		adatypes.Central.Log.Debugf("out of %s/%d", repository.URL.String(), repository.Fnr)
	}
	return
}

// parseMap Adabas read parser of the Map names used during read
func parseMapNames(adabasRequest *adatypes.Request, x interface{}) (err error) {
	repository := x.(*Repository)
	v := adabasRequest.Definition.Search(mapFieldName.fieldName())
	name := v.String()
	repository.Lock()
	defer repository.Unlock()
	if f, ok := repository.mapNames[name]; ok {
		f.found = true
	} else {
		repository.mapNames[name] = &mapNameFlags{isn: adabasRequest.Isn, found: true}
	}
	return
}

// parseMap Adabas read parser of the Map definition used during read
func parseMaps(adabasRequest *adatypes.Request, x interface{}) (err error) {
	repository := x.(*Repository)
	adabasMap := NewAdabasMap(&repository.DatabaseURL, &DatabaseURL{})
	err = parseMap(adabasRequest, adabasMap)
	if err != nil {
		return
	}

	dbid, dbidErr := adabasMap.Repository.dbid()
	if dbidErr != nil {
		err = dbidErr
		return
	}
	adatypes.Central.Log.Debugf("Got database for map %s is %d", adabasMap.Name, dbid)
	if dbid == 0 {
		dbid, _ = repository.dbid()
		adatypes.Central.Log.Debugf("Repository for map %s is rep dbid %d", adabasMap.Name, dbid)
	}
	adatypes.Central.Log.Debugf("Map map adabas to %s/%d", adabasMap.Repository.URL.String(), adabasMap.Repository.Fnr)
	adatypes.Central.Log.Debugf("Map data adabas to %s/%d", adabasMap.Data.URL.String(), adabasMap.Data.Fnr)
	// Create hashs
	adabasMap.createFieldMap()

	adatypes.Central.Log.Debugf("Add map name %s", adabasMap.Name)
	repository.AddMapToCache(adabasMap.Name, adabasMap)
	repository.Lock()
	defer repository.Unlock()

	repository.mapNames[adabasMap.Name] = &mapNameFlags{isn: adabasRequest.Isn, found: true}
	return
}

// LoadMapRepository read an index for names of all Adabas maps in the repository into memory
func (repository *Repository) LoadMapRepository(adabas *Adabas) (err error) {
	// if repository.mapNames != nil {
	// 	return nil
	// }
	repository.mapNames = make(map[string]*mapNameFlags)
	adabas.SetURL(&repository.DatabaseURL.URL)
	adatypes.Central.Log.Debugf("Read all data from dbid=%d(%s) of %s/%d\n",
		adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.DatabaseURL.URL.String(), repository.Fnr)
	//	adabas.Acbx.Acbxdbid = repository.DatabaseURL.URL.Dbid
	request, _ := NewReadRequest(adabas, repository.Fnr)
	request.Limit = 0
	err = request.QueryFields(mapFieldName.fieldName())
	if err != nil {
		repository.online = false
		adatypes.Central.Log.Debugf("Err %v query fields dbid=%d(%s) / %d\n", err, adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.Fnr)
		return err
	}
	err = request.ReadLogicalByWithParser(mapFieldName.fieldName(), parseMapNames, repository)
	if err != nil {
		repository.online = false
		adatypes.Central.Log.Debugf("Err %v Read all map names from dbid=%d(%s) / %d\n", err, adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.Fnr)
		return err
	}
	adatypes.Central.Log.Debugf("Done Read all map names from dbid=%d(%s) / %d\n", adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.Fnr)
	repository.online = true

	return
}

// write Adabas Map into database repository
func (repository *Repository) writeAdabasMapsWithAdabas(adabas *Adabas, adabasMap *Map) (err error) {
	defer adabas.Close()
	request := NewStoreRequestAdabas(adabas, repository.Fnr)
	err = request.StoreFields("*")
	if err != nil {
		adatypes.Central.Log.Debugf("Error store fields: %v", err)
		return
	}
	record, cerr := request.CreateRecord()
	if cerr != nil {
		err = cerr
		adatypes.Central.Log.Debugf("Error store fields: %v", cerr)
		return
	}
	adatypes.Central.Log.Debugf("Write map: %s record=%d", adabasMap.String(), record.Isn)
	record.SetValue(mapFieldName.fieldName(), adabasMap.Name)
	record.SetValue("TA", 77)
	hostname, herr := os.Hostname()
	if herr != nil {
		hostname = "localhost"
	}
	record.SetValue("AB", hostname)
	ut := time.Now().Unix()
	adatypes.Central.Log.Debugf("unit time stamp %d", ut)
	record.SetValue("AC", ut)
	record.SetValue("AD", 1)

	if adabasMap.Data == nil {
		err = adatypes.NewGenericError(17)
		return
	}

	record.SetValue("RF", uint32(adabasMap.Data.Fnr))
	if adabasMap.Data.URL.String() != adabasMap.Repository.URL.String() {
		record.SetValue("RD", adabasMap.Data.URL.String())
	}

	for index, m := range adabasMap.Fields {
		adatypes.Central.Log.Debugf("Store fields: %v - %v", m.ShortName, m.LongName)
		record.SetValueWithIndex("MB", []uint32{uint32(index + 1)}, m.ShortName)
		record.SetValueWithIndex("MC", []uint32{uint32(index + 1)}, 0)
		record.SetValueWithIndex("MD", []uint32{uint32(index + 1)}, m.LongName)
		record.SetValueWithIndex("ML", []uint32{uint32(index + 1)}, m.Length)
		record.SetValueWithIndex("MT", []uint32{uint32(index + 1)}, m.ContentType)
		record.SetValueWithIndex("MY", []uint32{uint32(index + 1)}, m.FormatType)
		adatypes.Central.Log.Debugf("Store Map Record %s", record)
	}
	err = request.Store(record)
	if err != nil {
		return
	}
	err = request.EndTransaction()
	return
}

// RemoveMap remove map from hash
func (repository *Repository) RemoveMap(mapName string) (err error) {
	repository.Lock()
	defer repository.Unlock()

	delete(repository.mapNames, mapName)
	return nil
}
