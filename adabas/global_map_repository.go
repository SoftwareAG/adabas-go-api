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
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

var (
	repositories       map[string]*Repository
	mapHash            map[string]*Repository
	mapLoopRunning     bool
	mapMaxMinutesCache = 10
)

func init() {
	mapHash = make(map[string]*Repository)
	mapCacheLoop := os.Getenv("MAP_CACHE_LOOP")
	if mapCacheLoop != "" {
		StartAsynchronousMapCache(10)
	}
}

// StartAsynchronousMapCache this method starts a thread which update the map cache
// periodically. The asynchronous Map cache is a repository Adabas Map cache
// which is asked for Adabas Maps defined. The given interval is the interval
// the cache is updated.
func StartAsynchronousMapCache(interval int) {
	go loopMapCache(interval)
}

// EndAsynchronousMapCache the Map cache update thread is initiated to end asynchronous
// cache update.
func EndAsynchronousMapCache() {
	mapLoopRunning = false
}

// loopMapCache Adabas Map cache loop thread
func loopMapCache(interval int) {
	adatypes.Central.Log.Debugf("Init loop map cache check")
	mapLoopRunning = true
	defer EndAsynchronousMapCache()
	for mapLoopRunning {
		adatypes.Central.Log.Debugf("Start loop map cache check")
		ada, err := NewAdabas(1)
		if err != nil {
			adatypes.Central.Log.Infof("Error loop map cache %v", err)
			return
		}
		_, err = readAllGlobalMapNames(ada)
		if err != nil {
			adatypes.Central.Log.Infof("Some map cache name error %v", err)
		}
		adatypes.Central.Log.Debugf("Number of Hashed maps: %d", len(mapHash))
		ada.Close()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// AddGlobalMapRepositoryReference this method adds a Adabas Map
// repository used to search for a used Adabas Map. This repository is
// defined global.
func AddGlobalMapRepositoryReference(reference string) error {
	url, fnr, err := extractReference(reference)
	if err != nil {
		return err
	}
	AddGlobalMapRepository(url, fnr)
	return nil
}

// AddGlobalMapRepository this method adds a Adabas Map
// repository used to search for a used Adabas Map. This repository is
// defined global.
func AddGlobalMapRepository(i interface{}, fnr Fnr) {
	var url *URL
	switch i.(type) {
	case *URL:
		url = i.(*URL)
	case *Adabas:
		a := i.(*Adabas)
		url = a.URL
	default:
		fmt.Println("Error adding global repository with", i)
		return
	}
	if repositories == nil {
		repositories = make(map[string]*Repository)
	}
	rep := NewMapRepository(url, fnr)
	reference := fmt.Sprintf("%s/%03d", url.String(), fnr)
	adatypes.Central.Log.Debugf("Add global repository >%s<", reference)
	repositories[reference] = rep
}

// DelGlobalMapRepositoryReference this method removes a global map repository
// entry.
func DelGlobalMapRepositoryReference(reference string) error {
	url, fnr, err := extractReference(reference)
	if err != nil {
		return err
	}
	DelGlobalMapRepository(url, fnr)
	return nil
}

// DelGlobalMapRepository this method removes a global map repository
// entry.
func DelGlobalMapRepository(i interface{}, fnr Fnr) {
	url := evaluateURL(i)
	if repositories != nil {
		reference := fmt.Sprintf("%s/%03d", url.String(), fnr)
		adatypes.Central.Log.Debugf("Remove global repository: %s", reference)
		r := repositories[reference]
		delete(repositories, reference)
		for k, v := range mapHash {
			if v == r {
				delete(mapHash, k)
			}
		}

	}
}

// CleanGlobalMapRepository this methods removes all global defined repositories
// in the Adabas Map repository.
func CleanGlobalMapRepository() {
	mapHash = make(map[string]*Repository)
	repositories = nil
}

// DumpGlobalMapRepositories dump global map repositories
func DumpGlobalMapRepositories() {
	fmt.Println("Dump global registered map repositories:")
	id := NewAdabasID()
	for _, r := range repositories {
		fmt.Printf("Repository at %s map file=%d:\n", r.URL, r.Fnr)
		if r.mapNames == nil || len(r.mapNames) == 0 {
			if a, err := NewAdabas(&r.DatabaseURL.URL, id); err == nil {
				err = r.LoadMapRepository(a)
				if err != nil {
					fmt.Println("    Map repository access problem", err)
				}
			} else {
				fmt.Println("    Map repository is empty or not initiated already", err)
			}
		}
		for m := range r.mapNames {
			fmt.Printf("    %s\n", m)
		}

	}
	fmt.Printf("Number of entries in cache=%d\n", len(mapHash))
}

// AllGlobalMaps provides all Adabas Maps defined in the in global defined repositories
func AllGlobalMaps(adabas *Adabas) (maps []*Map, err error) {
	mm := make(map[string]string)
	for mn, mr := range repositories {
		adabas.SetDbid(mr.DatabaseURL.URL.Dbid)
		adatypes.Central.Log.Debugf("Read maps in repository using Adabas %s for %s/%03d in %s",
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

// AllGlobalMapNames provides all Adabas Map names defined in the in global defined repositories
func AllGlobalMapNames(adabas *Adabas) (maps []string, err error) {
	// First map loop running using the cache map
	if mapLoopRunning {
		maps = make([]string, 0)
		for m, mr := range mapHash {
			if mr.online {
				maps = append(maps, m)
			}
		}
		adatypes.Central.Log.Debugf("All global maps number %d", len(maps))

		return maps, nil
	}
	return readAllGlobalMapNames(adabas)
}

// readAllGlobalMapNames search in map repositories global defined, all map names
func readAllGlobalMapNames(ada *Adabas) (maps []string, err error) {
	defer ada.Close()
	maxCacheTime := time.Now().Add(time.Duration(-mapMaxMinutesCache) * time.Minute)
	// If no map loop running, read through all repositories
	for ref, mr := range repositories {
		mr.ClearCache(maxCacheTime)
		ada.SetDbid(mr.DatabaseURL.URL.Dbid)
		adatypes.Central.Log.Debugf("Read map names in repository using Adabas %s for %s/%03d in %s",
			ada.URL.String(), mr.DatabaseURL.URL.String(), mr.Fnr, ref)
		err = mr.LoadMapRepository(ada)
		if err != nil {
			adatypes.Central.Log.Infof("Skip repository %s/%d due to error %v define offline", mr.DatabaseURL.URL.String(), mr.Fnr, err)
			mr.online = false
			continue
		}
		mr.online = true
		for mn := range mr.mapNames {
			maps = append(maps, mn)
			mapHash[mn] = mr
		}
		adatypes.Central.Log.Debugf("Found %d map names in repository using Adabas %s/%03d hash=%d", len(maps), ada.URL.String(), mr.Fnr, len(mapHash))
	}
	adatypes.Central.Log.Debugf("Found %d map names in all repositories", len(maps))
	return
}

// SearchMapRepository searchs in global-defined Adabas Map repositories for a specific map name
func SearchMapRepository(adabas *Adabas, mapName string) (adabasMap *Map, repository *Repository, err error) {
	// Check if hash is defined
	if r, ok := mapHash[mapName]; ok {
		if r.online {
			adatypes.Central.Log.Debugf("Found in map hash, query map...")
			adabas.SetURL(&r.DatabaseURL.URL)
			adabasMap, err = r.SearchMap(adabas, mapName)
			if err == nil {
				repository = r
				return
			}
			adatypes.Central.Log.Debugf("Error searching in repository: %v", err)
		}
	}
	adatypes.Central.Log.Debugf("Not found in map hash or error accessing repository, go through all repositories len=%d", len(repositories))
	// Not in hash search repository
	for _, mr := range repositories {
		if mr.online {
			adatypes.Central.Log.Debugf("Search in repository using Adabas %s for %s,%03d",
				adabas.URL.String(), mr.DatabaseURL.URL.String(), mr.Fnr)
			var serr error
			adabas.SetURL(&mr.DatabaseURL.URL)
			//adabas.SetDbid(mr.DatabaseURL.URL.Dbid)
			adabasMap, serr = mr.SearchMapInRepository(adabas, mapName)
			if serr != nil {
				adatypes.Central.Log.Debugf("Continue in next repository because of error %v\n", serr)
			} else {
				if adabasMap != nil {
					adatypes.Central.Log.Debugf("Result map found: %s", adabasMap.String())
					adatypes.Central.Log.Debugf("in repository %s/%d", mr.URL.String(), mr.Fnr)
					mapHash[mapName] = mr
					return adabasMap, mr, nil
				}
			}
			adatypes.Central.Log.Debugf("Not found in repository using Adabas %s/%03d", adabas.URL.String(), mr.Fnr)
		} else {
			adatypes.Central.Log.Debugf("Repository offline: %s-%d", mr.DatabaseURL.URL.String(), mr.DatabaseURL.Fnr)
		}
	}
	adatypes.Central.Log.Debugf("No map found error\n")
	err = adatypes.NewGenericError(16, mapName)
	return
}

// extractReference extract string representation to URL and fnr
func extractReference(reference string) (url *URL, fnr Fnr, err error) {
	v := strings.Split(reference, ",")
	if len(v) < 2 {
		return nil, 0, adatypes.NewGenericError(132)
	}
	url, err = NewURL(v[0])
	if err != nil {
		return
	}
	f, ferr := strconv.ParseUint(v[1], 10, 0)
	if ferr != nil {
		err = ferr
		return
	}
	if f < 0 || f > math.MaxUint32 {
		return
	}
	fnr = Fnr(f)
	return
}
