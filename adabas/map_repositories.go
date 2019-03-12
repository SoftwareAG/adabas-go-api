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
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// DatabaseURL defines the target URL of a database file. Might be a database data file or a map repository
type DatabaseURL struct {
	URL URL `json:"Target"`
	Fnr Fnr `json:"File"`
}

func (repURL *DatabaseURL) dbid() (dbid Dbid, err error) {
	adatypes.Central.Log.Debugf("repURL=%#v", repURL)
	dbid = repURL.URL.Dbid
	return
}

// Repository Adabas Map repository container
type Repository struct {
	DatabaseURL
	MapNames   map[string]adatypes.Isn
	CachedMaps map[string]*Map
}

var repositories map[string]*Repository

// NewMapRepository new map repository created
func NewMapRepository(adabas *Adabas, fnr Fnr) *Repository {
	mr := &Repository{DatabaseURL: DatabaseURL{URL: *adabas.URL, Fnr: fnr}}
	mr.CachedMaps = make(map[string]*Map)
	return mr
}

// NewMapRepositoryWithURL new map repository created
func NewMapRepositoryWithURL(url DatabaseURL) *Repository {
	mr := &Repository{DatabaseURL: url}
	mr.CachedMaps = make(map[string]*Map)
	return mr
}

// AddGlobalMapRepository add global map repository
func AddGlobalMapRepository(adabas *Adabas, fnr Fnr) {
	if repositories == nil {
		repositories = make(map[string]*Repository)
	}
	rep := NewMapRepository(adabas, fnr)
	reference := fmt.Sprintf("%s/%03d", adabas.URL.String(), fnr)
	adatypes.Central.Log.Debugf("Add global repository >%s<", reference)
	repositories[reference] = rep
}

// DelGlobalMapRepository delete global map repository
func DelGlobalMapRepository(adabas *Adabas, fnr Fnr) {
	if repositories != nil {
		reference := fmt.Sprintf("%s/%03d", adabas.URL.String(), fnr)
		adatypes.Central.Log.Debugf("Remove global repository: %s", reference)
		delete(repositories, reference)
	}
}

// DumpGlobalMapRepositories dump global map repositories
func DumpGlobalMapRepositories() {
	fmt.Println("Dump global registered map repositories:")
	for _, r := range repositories {
		fmt.Printf("Repository at %s map file=%d:\n", r.URL, r.Fnr)
		if r.MapNames == nil || len(r.MapNames) == 0 {
			fmt.Println("    Map repository is empty or not initiated already")
		} else {
			for m := range r.MapNames {
				fmt.Printf("    %s\n", m)
			}
		}
	}
	fmt.Println("Dump global registered map repositories done")
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
	repository.CachedMaps[mapName] = adabasMap
	return
}

// readAdabasMapWithRequest read specific Adabas Map defined by the map name and the repository
func (repository *Repository) readAdabasMapWithRequest(commonRequest *commonRequest, name string) (adabasMap *Map, err error) {
	if repository == nil {
		err = adatypes.NewGenericError(5)
		return
	}
	adatypes.Central.Log.Debugf("Before Prepare Repository %#v\n", *repository)
	url := repository.DatabaseURL
	adabasMap = &Map{Repository: &repository.DatabaseURL, Data: &url}
	request := NewReadRequestCommon(commonRequest)
	// Reset map definition, because
	request.commonRequest.adabasMap = nil
	request.commonRequest.mapName = ""
	adatypes.Central.Log.Debugf("Before Read Repository %#v\n", *repository)

	adatypes.Central.Log.Debugf("Search for Map with name=%s", name)
	// Search for map name in database
	err = request.ReadLogicalWithWithParser(mapFieldName.fieldName()+"="+name, parseMap, adabasMap)
	if err != nil {
		return nil, err
	}
	if adabasMap.Name == "" {
		return nil, adatypes.NewGenericError(66)
	}
	adabasMap.createFieldMap()
	repository.CachedMaps[name] = adabasMap
	var dbid Dbid
	adatypes.Central.Log.Debugf("After Repository %#v\n", *repository)
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
	request := NewReadRequestAdabas(adabas, repository.Fnr)
	adatypes.Central.Log.Debugf("Repository %#v\n", *repository)
	adabasMap, err = repository.readAdabasMapWithRequest(&request.commonRequest, name)
	return
}

// SearchMap search map name in specific map repository
func (repository *Repository) SearchMap(adabas *Adabas, mapName string) (adabasMap *Map, err error) {
	if repository.MapNames == nil {
		err = repository.LoadMapRepository(adabas)
		if err != nil {
			return
		}
	}
	if _, ok := repository.MapNames[mapName]; !ok {
		err = adatypes.NewGenericError(14, mapName)
		return
	}
	// Need a Adabas instance to work with corresponding ID, else return error
	if adabas == nil {
		return nil, adatypes.NewGenericError(64)
	}
	adatypes.Central.Log.Debugf("Search map: %s", mapName)
	if m, ok := repository.CachedMaps[mapName]; ok {
		adabasMap = m
		return
	}

	request := NewReadRequestAdabas(adabas, repository.Fnr)
	request.Limit = 0
	err = request.ReadLogicalWithWithParser(mapFieldName.fieldName()+"="+mapName, parseMaps, repository)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read map repoistory searching %s", mapName)
	if m, ok := repository.CachedMaps[mapName]; ok {
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
	adatypes.Central.Log.Debugf("Got map adabas to %s/%d", adabasMap.Repository.URL.String(), adabasMap.Repository.Fnr)
	adatypes.Central.Log.Debugf("with data adabas to %s/%d", adabasMap.Data.URL.String(), adabasMap.Data.Fnr)
	adatypes.Central.Log.Debugf("out of %s/%d", repository.URL.String(), repository.Fnr)
	return
}

// LoadAllMaps load all map out of specific map repository
func (repository *Repository) LoadAllMaps(adabas *Adabas) (adabasMaps []*Map, err error) {
	if repository.MapNames == nil {
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
	request := NewReadRequestAdabas(adabas, repository.Fnr)
	request.Limit = 0
	err = request.ReadPhysicalSequenceWithParser(parseMaps, repository)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read all map entries in repository")
	for _, m := range repository.CachedMaps {
		adabasMaps = append(adabasMaps, m)
		adatypes.Central.Log.Debugf("Got map %s adabas to %s/%d", m.Name, m.Repository.URL.String(), m.Repository.Fnr)
		adatypes.Central.Log.Debugf("with data adabas to %s/%d", m.Data.URL.String(), m.Data.Fnr)
		adatypes.Central.Log.Debugf("out of %s/%d", repository.URL.String(), repository.Fnr)
	}
	return
}

// SearchMapRepository search in map repository for a specific map name
func SearchMapRepository(adabas *Adabas, mapName string) (adabasMap *Map, err error) {
	for _, mr := range repositories {
		adatypes.Central.Log.Debugf("Search in repository using Adabas %s for %s/%03d",
			adabas.URL.String(), mr.DatabaseURL.URL.String(), mr.Fnr)
		var serr error
		adabas.SetDbid(mr.DatabaseURL.URL.Dbid)
		adabasMap, serr = mr.SearchMapInRepository(adabas, mapName)
		if serr != nil {
			adatypes.Central.Log.Debugf("Continue in next repository because of error %v\n", serr)
		} else {
			if adabasMap != nil {
				adatypes.Central.Log.Debugf("Result map found: %s", adabasMap.String())
				adatypes.Central.Log.Debugf("in repository %s/%d", mr.URL.String(), mr.Fnr)
				return
			}
		}
		adatypes.Central.Log.Debugf("Not found in repository using Adabas %s/%03d", adabas.URL.String(), mr.Fnr)
	}
	adatypes.Central.Log.Debugf("No map found error\n")
	err = adatypes.NewGenericError(16, mapName)
	return
}

// GloablMaps search in map repository all maps
func GloablMaps(adabas *Adabas) (maps []*Map, err error) {
	mm := make(map[string]string)
	for mn, mr := range repositories {
		adabas.SetDbid(mr.DatabaseURL.URL.Dbid)
		adatypes.Central.Log.Debugf("Read in repository using Adabas %s for %s/%03d in %s",
			adabas.URL.String(), mr.DatabaseURL.URL.String(), mr.Fnr, mn)
		adabasMaps, serr := mr.LoadAllMaps(adabas)
		if serr != nil {
			adatypes.Central.Log.Debugf("Continue in next repository because of error %v\n", serr)
		} else {
			for _, m := range adabasMaps {
				if _, ok := mm[m.Name]; !ok {
					mm[m.Name] = m.Name
					maps = append(maps, m)
				}
			}
		}
		adatypes.Central.Log.Debugf("Found %d in repository using Adabas %s/%03d", len(maps), adabas.URL.String(), mr.Fnr)
	}
	return
}

// LoadMapRepository create a new repository
func (repository *Repository) LoadMapRepository(adabas *Adabas) (err error) {
	return repository.LoadRepositoryMapsWithAdabas(adabas)
}

// parseMap Adabas read parser of the Map names used during read
func parseMapNames(adabasRequest *adatypes.Request, x interface{}) (err error) {
	repository := x.(*Repository)
	v := adabasRequest.Definition.Search(mapFieldName.fieldName())
	name := v.String()
	repository.MapNames[name] = adabasRequest.Isn
	return
}

// parseMap Adabas read parser of the Map definition used during read
func parseMaps(adabasRequest *adatypes.Request, x interface{}) (err error) {
	repository := x.(*Repository)
	adabasMap := &Map{Repository: &repository.DatabaseURL, Data: &DatabaseURL{}}
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

	repository.CachedMaps[adabasMap.Name] = adabasMap
	repository.MapNames[adabasMap.Name] = adabasRequest.Isn
	return
}

// LoadRepositoryMapsWithAdabas read on index the names of all Adabas maps into memory
func (repository *Repository) LoadRepositoryMapsWithAdabas(adabas *Adabas) (err error) {
	adatypes.Central.Log.Debugf("Read all data from dbid=%d(%s) of %s/%d\n",
		adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.DatabaseURL.URL.String(), repository.Fnr)
	repository.MapNames = make(map[string]adatypes.Isn)

	adabas.Acbx.Acbxdbid = repository.DatabaseURL.URL.Dbid
	request := NewReadRequestAdabas(adabas, repository.Fnr)
	request.Limit = 0
	request.QueryFields(mapFieldName.fieldName())
	err = request.ReadLogicalByWithParser(mapFieldName.fieldName(), parseMapNames, repository)
	if err != nil {
		adatypes.Central.Log.Debugf("Err %v Read all data from dbid=%d(%s) / %d\n", err, adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.Fnr)
		return err
	}
	adatypes.Central.Log.Debugf("Done Read all data from dbid=%d(%s) / %d\n", adabas.Acbx.Acbxdbid, adabas.URL.String(), repository.Fnr)

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
