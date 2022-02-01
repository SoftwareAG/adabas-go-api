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
	"bytes"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// Connection Adabas connection context
type Connection struct {
	ID           *ID
	adabasToData *Adabas
	adabasMap    *Map
	adabasToMap  *Adabas
	fnr          Fnr
	repository   *Repository
}

var once sync.Once

const inmapMapName = "<inmap>"

var onceBody = func() {
	adatypes.Central.Log.Infof("Adabas GO API version %s", adatypes.Version)
}

// NewConnection create new Adabas connection instance
// The target url will look like <dbid>(<driver prefix>://<host>:<port>).
// Examples are:
//   - Database id: 23
//   - Adabas TCP on port 60023:  23(adatcp://pchost:60023)
//   - Adabas Entire Network (Java only): 23(tcpip://pchost:50001)
// The connection string must contain:
//   - To access database classic targets
//     acj;target=<database url>
//   - Map usage
//     acj;map;config=[<dbid>,<file>]
func NewConnection(connectionString string) (*Connection, error) {
	adabasID := NewAdabasID()
	return NewConnectionID(connectionString, adabasID)
}

// NewConnectionID create new Adabas connection instance providing a Adabas ID
// The target url will look like <dbid>(<driver prefix>://<host>:<port>).
// Examples are:
//   - Database id: 23
//   - Adabas TCP on port 60023:  23(adatcp://pchost:60023)
//   - Adabas Entire Network (Java only): 23(tcpip://pchost:50001)
// The connection string must contain:
//   - To access database classic targets
//     acj;target=<database url>
//   - Map usage
//     acj;map;config=[<dbid>,<file>]
func NewConnectionID(connectionString string, adabasID *ID) (connection *Connection, err error) {
	once.Do(onceBody)
	parts := strings.Split(connectionString, ";")
	if parts[0] != "acj" && parts[0] != "ada" {
		return nil, adatypes.NewGenericError(51)
	}
	var adabasToData *Adabas
	var adabasToMap *Adabas
	var mapName string
	var adabasMap *Map

	var repositoryParameter [][]string
	var repository *Repository
	adatypes.Central.Log.Debugf("New connection to %s", connectionString)
	for _, p := range parts {
		adatypes.Central.Log.Debugf("Work on %s", p)
		switch {
		case p == "acj" || p == "ada":
		case strings.HasPrefix(p, "target="):
			target := strings.Split(parts[1], "=")
			adatypes.Central.Log.Debugf("Connection to target : %s", target[1])
			url, err := NewURL(target[1])
			if err != nil {
				return nil, err
			}
			adabasToData = adabasID.getAdabas(url)
		case strings.HasPrefix(p, "map"):
			if strings.Contains(p, "=") {
				maps := strings.Split(parts[1], "=")
				adatypes.Central.Log.Debugf("Connection to map : %v", maps)
				mapName = maps[1]
			}
		case strings.HasPrefix(p, "inmap"):
			if strings.Contains(p, "=") {
				maps := strings.Split(parts[1], "=")
				adatypes.Central.Log.Debugf("Connection to map : %v", maps)
				// mapName = inmapMapName
				adabasMap = NewAdabasMap(inmapMapName)
				ref := strings.Split(maps[1], ",")
				url, err := NewURL(ref[0])
				if err != nil {
					return nil, err
				}
				adabasMap.Data = &DatabaseURL{URL: *url}
				fnr := 0
				if len(ref) > 1 {
					fnr, err = strconv.Atoi(ref[1])
					if err != nil {
						return nil, err
					}
				}
				if fnr < 0 || fnr > 32000 {
					return nil, adatypes.NewGenericError(116, fnr)
				}
				adabasMap.Data.Fnr = Fnr(fnr)
				adatypes.Central.Log.Debugf("inmap %s,%d", url, fnr)
				adabasToData = adabasID.getAdabas(url)
				// NewAdabas(url, adabasID)
				// if err != nil {
				// 	return nil, err
				// }
			}
		case strings.HasPrefix(p, "config="):
			e := strings.Index(p, "]")
			a := strings.Index(p, "[") + 1
			config := p[a:e]
			re := regexp.MustCompile(`(?m)([^,]*),([[:digit:]]*)\|?`)
			rr := re.FindAllStringSubmatch(config, 10)
			for _, r1 := range rr {
				var r = []string{r1[1], r1[2]}
				repositoryParameter = append(repositoryParameter, r)
			}
		case strings.HasPrefix(p, "auth="):
			x := strings.Index(p, ",")
			if x != -1 {
				x++
				err := parseAuth(adabasID, p[x:])
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, adatypes.NewGenericError(84, p)
		}
	}

	if len(repositoryParameter) > 0 {
		for _, r := range repositoryParameter {
			adatypes.Central.Log.Debugf("Add repository search of dbid=%s fnr=%s", r[0], r[1])
			fnr, serr := strconv.Atoi(r[1])
			if serr != nil {
				return nil, serr
			}
			if fnr < 0 || fnr > 32000 {
				return nil, adatypes.NewGenericError(116, fnr)
			}
			url, err := NewURL(r[0])
			//NewAdabas(r[0], adabasID)
			if err != nil {
				return nil, err
			}
			adabasToMap = adabasID.getAdabas(url)
			adatypes.Central.Log.Debugf("Created adabas reference")
			repository = NewMapRepository(adabasToMap.URL, Fnr(fnr))
			adatypes.Central.Log.Debugf("Created repository")
		}
	} else {
		// TODO check id
		if adabasToData == nil {
			adabasToData, _ = NewAdabas(1, adabasID)
		}
		adabasToMap = adabasToData
	}

	connection = &Connection{adabasToData: adabasToData, ID: adabasID,
		adabasToMap: adabasToMap, adabasMap: adabasMap, repository: repository}
	if mapName != "" {
		err = connection.searchRepository(adabasID, repository, mapName)
		if err != nil {
			return nil, err
		}
	}

	adatypes.Central.Log.Debugf("Ready created connection handle %s", connection.String())
	return
}

// searchRepository search a Adabas Map by name in the Adabas Map repository
func (connection *Connection) searchRepository(adabasID *ID, repository *Repository,
	mapName string) (err error) {
	if repository == nil {
		adatypes.Central.Log.Debugf("Search in global repositories")
		connection.adabasMap, _, err = SearchMapRepository(adabasID, mapName)
		if err != nil {
			adatypes.Central.Log.Debugf("Search in global repositories fail: %v", err)
			return err
		}
		if connection.adabasMap == nil {
			return adatypes.NewGenericError(85, mapName)
		}
	} else {
		adatypes.Central.Log.Debugf("Search in given repository %v: %s", repository, repository.DatabaseURL.URL.String())
		connection.adabasToMap = adabasID.getAdabas(&repository.DatabaseURL.URL)
		// NewAdabas(repository.DatabaseURL.URL.String(), adabasID)
		if err != nil {
			adatypes.Central.Log.Debugf("New Adabas to map ID error: %v", err)
			return err
		}
		connection.adabasMap, err = repository.SearchMap(connection.adabasToMap, mapName)
		if err != nil {
			adatypes.Central.Log.Debugf("Search map error: %v", err)
			return err
		}
	}
	adatypes.Central.Log.Debugf("Found map %s", connection.adabasMap.Name)
	if connection.adabasMap.URL().String() != connection.adabasToMap.URL.String() {
		connection.adabasToMap = connection.ID.getAdabas(connection.adabasMap.URL())
	}
	if connection.adabasMap.URL().String() == connection.adabasMap.Data.URL.String() {
		adatypes.Central.Log.Debugf("Same URL %v -> %v", connection.adabasMap.URL().String(), connection.adabasToMap.URL)
		connection.adabasToData = connection.adabasToMap
	} else {
		adatypes.Central.Log.Debugf("Create new Adabas URL %v!=%v", connection.adabasMap.URL().String(), connection.adabasMap.Data.URL.String())
		connection.adabasToData = adabasID.getAdabas(&connection.adabasMap.Data.URL)
		// NewAdabas(&connection.adabasMap.Data.URL, adabasID)
		// if err != nil {
		// 	adatypes.Central.Log.Debugf("Error new ADabas URL %v", err)
		// 	return err
		// }
	}
	adatypes.Central.Log.Debugf("Final error: %v", err)
	return
}

// parseAuth parse the authentication credentials in the connection string
func parseAuth(id *ID, value string) error {
	re := regexp.MustCompile(`(\w+)=(\w+|'.+'|".+")(,)?`)
	match := re.FindAllString(value, -1)
	for _, x := range match {
		l := len(x)

		if strings.HasSuffix(x, ",") {
			l--
		}
		i := strings.Index(x, "=")
		n := strings.ToLower(x[:i])
		v := x[i+1 : l]
		switch n {
		case "host":
			id.SetHost(v)
		case "user":
			id.SetUser(v)
		case "id":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if i < 0 || i > math.MaxInt32 {
				return adatypes.NewGenericError(119, i)
			}
			id.SetID(uint32(i))
		}
	}
	return nil
}

// String provide the string representation of the connection
func (connection *Connection) String() string {
	if connection == nil {
		return "Connection <nil>"
	}
	var buffer bytes.Buffer
	if connection.adabasMap != nil {
		buffer.WriteString("Map=" + connection.adabasMap.Name + " ")
	}
	if connection.adabasToData == nil {
		buffer.WriteString("Target not defined")
	} else {
		buffer.WriteString(connection.adabasToData.String())
	}
	if connection.fnr != 0 {
		buffer.WriteString(" connection file=" + strconv.Itoa(int(connection.fnr)))
	}
	return buffer.String()
}

// Open open Adabas session
func (connection *Connection) Open() error {
	if connection.adabasToData == nil {
		return nil
	}
	err := connection.adabasToData.Open()
	return err
}

// Close the Adabas session will be closed. An Adabas session/user queue entry
// in the database will be removed. If transaction are open, the backout of the
// transaction is called. All open transaction a rolled back and data restored.
// The backout transaction error is not provided. It is not sure if BT is done.
// Do extra BackoutTransaction() call to be sure and track error.
func (connection *Connection) Close() {
	if connection.adabasToData != nil {
		_ = connection.adabasToData.BackoutTransaction()
		connection.adabasToData.Close()
	}
	if connection.adabasToMap != nil {
		_ = connection.adabasToMap.BackoutTransaction()
		connection.adabasToMap.Close()
	}
}

// BackoutTransaction all current transaction will be backout and reset to before
// the transaction in the Adabas database.
func (connection *Connection) BackoutTransaction() error {
	if connection.adabasToData != nil {
		err := connection.adabasToData.BackoutTransaction()
		if err != nil {
			return err
		}
	}
	if connection.adabasToMap != nil {
		err := connection.adabasToMap.BackoutTransaction()
		if err != nil {
			return err
		}
	}
	return nil
}

// EndTransaction all current transaction will be finally stored in the
// Adabas database.
func (connection *Connection) EndTransaction() error {
	if connection.adabasToData != nil {
		err := connection.adabasToData.EndTransaction()
		if err != nil {
			return err
		}
	}
	if connection.adabasToMap != nil {
		err := connection.adabasToMap.EndTransaction()
		if err != nil {
			return err
		}
	}
	return nil
}

// Release any database hold record resources, like command id caches assigned to a user
func (connection *Connection) Release() error {
	if connection.adabasToData != nil {
		err := connection.adabasToData.ReleaseHold(connection.fnr)
		if err != nil {
			return err
		}
	}
	if connection.adabasToMap != nil && connection.adabasMap != nil {
		err := connection.adabasToMap.ReleaseHold(connection.adabasMap.Data.Fnr)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReleaseCID any database command id resources, like command id caches assigned to a user
// are released on the database.
func (connection *Connection) ReleaseCID() error {
	if connection.adabasToData != nil {
		err := connection.adabasToData.ReleaseCmdID()
		if err != nil {
			return err
		}
	}
	if connection.adabasToMap != nil {
		err := connection.adabasToMap.ReleaseCmdID()
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCredential this method adds user id and password credentials to the called.
// The credentials are needed if the Adabas security is active in the database.
func (connection *Connection) AddCredential(user string, pwd string) {
	connection.ID.AddCredential(user, pwd)
}

// ReadRequest this method create a read request defined by the given list of
// parameters.
// First parameter is used to indicate target. It might be a file number or
// map string name.
// Second parameter is used to indicate query fields.
func (connection *Connection) ReadRequest(p ...interface{}) (request *ReadRequest, err error) {
	if len(p) == 0 {
		return connection.CreateReadRequest()
	}
	switch t := p[0].(type) {
	case int:
		request, err = connection.CreateFileReadRequest(Fnr(t))
	case Fnr:
		request, err = connection.CreateFileReadRequest(t)
	case string:
		request, err = connection.CreateMapReadRequest(t)
	default:
		return nil, fmt.Errorf("target parameter for read request wrong")
	}
	if err != nil {
		return nil, err
	}
	if len(p) > 1 {
		switch q := p[1].(type) {
		case string:
			err = request.QueryFields(q)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("query fields parameter for read request wrong")
		}
	}
	return request, nil
}

// CreateReadRequest this method create a read request defined by the given map in
// the `Connection` creation. If no map is given, an error 83 is returned.
func (connection *Connection) CreateReadRequest() (request *ReadRequest, err error) {
	if connection.adabasMap == nil {
		adatypes.Central.Log.Debugf("Map empty: %#v", connection)
		return nil, adatypes.NewGenericError(83)
	}
	connection.fnr = connection.adabasMap.Data.Fnr
	adatypes.Central.Log.Debugf("Map referenced : %#v", connection.adabasMap)
	request, err = NewReadRequest(connection.adabasToData, connection.adabasMap)
	return
}

// CreateFileReadRequest this method creates a read request using a given Adabas
// file number. The file number request will be used with Adabas short names, not
// long names.
func (connection *Connection) CreateFileReadRequest(fnr Fnr) (*ReadRequest, error) {
	adatypes.Central.Log.Debugf("Connection: %#v", connection)
	adatypes.Central.Log.Debugf("Data referenced : %#v", connection.adabasToData)
	return NewReadRequest(connection.adabasToData, fnr)
}

// CreateMapReadRequest this method creates a read request using a given Adabas Map
// definition. The Map will be searched in an globally defined Map repository only.
func (connection *Connection) CreateMapReadRequest(param ...interface{}) (request *ReadRequest, err error) {
	t := reflect.TypeOf(param[0])
	switch t.Kind() {
	case reflect.Ptr, reflect.Struct:
		if connection.repository == nil {
			if connection.adabasMap != nil && connection.adabasMap.Name == inmapMapName {
				// Lock this Adabas Map for ReadRequest creation process of dynamic part
				connection.adabasMap.lock.Lock()
				defer connection.adabasMap.lock.Unlock()
				adatypes.Central.Log.Debugf("InMap used %s", connection.adabasMap.Name)
				err = connection.adabasMap.defineByInterface(param[0])
				if err != nil {
					return nil, err
				}
				request, err = NewReadRequest(connection.adabasToData, connection.adabasMap)
			} else {
				request, err = NewReadRequest(param[0], connection.adabasToMap)
			}
		} else {
			request, err = NewReadRequest(param[0], connection.adabasToMap, connection.repository)
		}
		if err != nil {
			return
		}
		if len(param) > 1 {
			switch t := param[1].(type) {
			case int:
				connection.fnr = Fnr(t)
				request.repository.Fnr = Fnr(t)
			case Fnr:
				connection.fnr = t
				request.repository.Fnr = t
			}
		} else {
			connection.fnr = request.adabasMap.Data.Fnr
		}
		connection.adabasMap = request.adabasMap
	case reflect.String:
		m := param[0].(string)
		err = connection.prepareMapUsage(m)
		if err != nil {
			return
		}
		connection.adabasMap.lock.Lock()
		defer connection.adabasMap.lock.Unlock()

		connection.fnr = connection.adabasMap.Data.Fnr
		adatypes.Central.Log.Debugf("Map referenced : %#v", connection.adabasMap)
		request, err = NewReadRequest(connection.adabasToData, connection.adabasMap)
		if len(param) > 1 {
			l := param[1].(string)
			ierr := request.createInterface(l)
			if ierr != nil {
				return nil, ierr
			}
		}
	default:
		return nil, adatypes.NewGenericError(0)
	}
	return
}

// CreateStoreRequest this method creates a store request for a Adabas file number.
// The store will be used with Adabas short names only.
func (connection *Connection) CreateStoreRequest(fnr Fnr) (*StoreRequest, error) {
	return NewStoreRequestAdabas(connection.adabasToData, fnr), nil
}

// CreateMapWithInterface this method create a Adabas Map request using the Map name
// and a list of fields defined in the dynamic interface
func (connection *Connection) CreateMapWithInterface(mapName string, fieldList string) (request *ReadRequest, err error) {
	err = connection.prepareMapUsage(mapName)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Create interface: %#v", connection.adabasMap)
	// i, err := connection.adabasMap.createInterface(fieldList)
	// if err != nil {
	// 	return
	// }
	// adatypes.Central.Log.Debugf("Create interface-based map request")

	return connection.CreateMapReadRequest(mapName, fieldList)
}

// prepareMapUsage prepare Map usage
func (connection *Connection) prepareMapUsage(mapName string) (err error) {
	return connection.searchRepository(connection.ID, connection.repository, mapName)
}

// CreateMapStoreRequest this method creates a store request using an Go struct which
// struct field names fit to an Adabas Map field. The struct name will be used to search
// the Adabas Map.
func (connection *Connection) CreateMapStoreRequest(param ...interface{}) (request *StoreRequest, err error) {
	t := reflect.TypeOf(param[0])
	switch t.Kind() {
	case reflect.Ptr, reflect.Struct:
		if connection.repository == nil {
			if connection.adabasMap != nil && connection.adabasMap.Name == inmapMapName {
				adatypes.Central.Log.Debugf("InMap used: %s", connection.adabasMap.Name)
				// Lock this Adabas Map for ReadRequest creation process of dynamic part
				connection.adabasMap.lock.Lock()
				defer connection.adabasMap.lock.Unlock()
				err = connection.adabasMap.defineByInterface(param[0])
				if err != nil {
					return nil, err
				}
				request, err = NewStoreRequest(connection.adabasToData, connection.adabasMap)
				if err != nil {
					return
				}

				if request.definition == nil {
					err = request.loadDefinition()
					if err != nil {
						adatypes.Central.Log.Debugf("Load definition error: %v", err)
						return
					}
				}
				request.dynamic = request.adabasMap.dynamic
				err = request.adabasMap.adaptFieldType(request.definition, request.dynamic)
				if err != nil {
					adatypes.Central.Log.Debugf("Adapt fields error request definition %v", err)
					return
				}
			} else {
				adatypes.Central.Log.Debugf("No repository used: %#v", connection.adabasToMap)
				request, err = NewStoreRequest(param[0], connection.adabasToMap)
				if err != nil {
					return
				}
			}
		} else {
			adatypes.Central.Log.Debugf("With repository used: %#v", connection.adabasMap)
			request, err = NewStoreRequest(param[0], connection.adabasToMap, connection.repository)
			if err != nil {
				return
			}
		}
		if len(param) > 1 {
			switch t := param[1].(type) {
			case int:
				connection.fnr = Fnr(t)
				request.repository.Fnr = Fnr(t)
			case Fnr:
				connection.fnr = t
				request.repository.Fnr = t
			}
		} else {
			if request.adabasMap == nil || request.adabasMap.Data == nil ||
				request.adabasMap.Data.Fnr == 0 {
				err = adatypes.NewGenericError(83)
				return
			}
			connection.fnr = request.adabasMap.Data.Fnr
		}
		connection.adabasMap = request.adabasMap
	case reflect.String:
		mapName := param[0].(string)
		err = connection.prepareMapUsage(mapName)
		if err != nil {
			return
		}
		request, err = NewAdabasMapNameStoreRequest(connection.adabasToData, connection.adabasMap)
	}
	return
}

// CreateDeleteRequest this method create a delete request using Adabas file numbers.
func (connection *Connection) CreateDeleteRequest(fnr Fnr) (*DeleteRequest, error) {
	return NewDeleteRequestAdabas(connection.adabasToData, fnr), nil
}

func evaluateInterface(mapReference interface{}) string {
	ti := reflect.TypeOf(mapReference)
	adatypes.Central.Log.Debugf("It's a struct %s", ti.Name())
	if ti.Kind() == reflect.Ptr {
		ti = ti.Elem()
	}
	return ti.Name()
}

// CreateMapDeleteRequest this method creates a delete request using a given Adabas Map name
func (connection *Connection) CreateMapDeleteRequest(param ...interface{}) (request *DeleteRequest, err error) {
	t := reflect.TypeOf(param[0])
	switch t.Kind() {
	case reflect.Ptr, reflect.Struct:
		if connection.repository == nil {
			if connection.adabasMap != nil && connection.adabasMap.Name == inmapMapName {
				// Lock this Adabas Map for ReadRequest creation process of dynamic part
				connection.adabasMap.lock.Lock()
				defer connection.adabasMap.lock.Unlock()
				adatypes.Central.Log.Debugf("InMap used: %s", connection.adabasMap.Name)
				err = connection.adabasMap.defineByInterface(param[0])
				if err != nil {
					return nil, err
				}
				request, err = NewMapDeleteRequest(connection.adabasToData, connection.adabasMap)
			} else {
				adatypes.Central.Log.Debugf("No repository used: %#v", connection.adabasToMap)
				request, err = NewMapNameDeleteRequest(connection.adabasToMap, evaluateInterface(param[0]))
			}
		} else {
			adatypes.Central.Log.Debugf("With repository used: %#v", connection.adabasMap)
			request, err = NewMapNameDeleteRequestRepo(evaluateInterface(param[0]), connection.adabasToMap, connection.repository)
			if err != nil {
				return
			}
		}
		if len(param) > 1 {
			switch t := param[1].(type) {
			case int:
				connection.fnr = Fnr(t)
				request.repository.Fnr = Fnr(t)
			case Fnr:
				connection.fnr = t
				request.repository.Fnr = t
			}
		} else {
			connection.fnr = request.adabasMap.Data.Fnr
		}
		connection.adabasMap = request.adabasMap
	case reflect.String:
		mapName := param[0].(string)
		err = connection.prepareMapUsage(mapName)
		if err != nil {
			return
		}
		// if connection.repository == nil {
		// 	err = adatypes.NewGenericError(9)
		// 	return
		// }
		// connection.repository.SearchMapInRepository(connection.adabasToMap, mapName)
		if connection.adabasMap == nil {
			err = adatypes.NewGenericError(8, mapName)
			return
		}
		connection.adabasToData = connection.ID.getAdabas(connection.adabasMap.URL())
		// NewAdabas(connection.adabasMap.URL(), connection.ID)
		// if err != nil {
		// 	return
		// }
		connection.fnr = connection.adabasMap.Data.Fnr
		adatypes.Central.Log.Debugf("Connection FNR=%d, Map referenced : %#v", connection.fnr, connection.adabasMap)
		request, err = NewMapDeleteRequest(connection.adabasToData, connection.adabasMap)
	}
	return
}

// IsCluster response status of remote node Adabas cluster part or not
func (connection *Connection) IsCluster() bool {
	if connection == nil || connection.adabasToData == nil {
		return false
	}
	if connection.adabasToData.transactions == nil {
		return false
	}
	return len(connection.adabasToData.transactions.clusterNodes) > 0
}

// GetClusterNodes retrieve cluster node list
func (connection *Connection) GetClusterNodes() []*URL {
	if connection.IsCluster() {
		return connection.adabasToData.transactions.clusterNodes
	}
	return make([]*URL, 0)
}

// GetAdabasInformation get Adabas information
func (connection *Connection) GetAdabasInformation() string {
	if connection.adabasToData == nil {
		return ""
	}
	return "v" + connection.adabasToData.Version() + "," +
		connection.adabasToData.Platform()
}
