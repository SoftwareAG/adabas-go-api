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
	"bytes"
	"reflect"
	"regexp"
	"strconv"
	"strings"

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
	parts := strings.Split(connectionString, ";")
	if parts[0] != "acj" && parts[0] != "ada" {
		return nil, adatypes.NewGenericError(51)
	}
	var adabasToData *Adabas
	var adabasToMap *Adabas
	var mapName string
	var adabasMap *Map

	var repositoryParameter []string
	var repository *Repository
	adatypes.Central.Log.Debugf("New connection to %s", connectionString)
	for _, p := range parts {
		adatypes.Central.Log.Debugf("Work on %s", p)
		switch {
		case p == "acj" || p == "ada":
		case strings.HasPrefix(p, "target="):
			target := strings.Split(parts[1], "=")
			adatypes.Central.Log.Debugf("Connection to target : %s", target[1])
			adabasToData, err = NewAdabas(target[1], adabasID)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(p, "map"):
			if strings.Contains(p, "=") {
				maps := strings.Split(parts[1], "=")
				adatypes.Central.Log.Debugf("Connection to map : %v", maps)
				mapName = maps[1]
			}
		case strings.HasPrefix(p, "config="):
			re := regexp.MustCompile(`config=\[([^,]*),([[:digit:]]*)\]`)
			repositoryParameter = re.FindStringSubmatch(p)
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

	if len(repositoryParameter) > 2 {
		adatypes.Central.Log.Debugf("Add repository search of dbid=%s fnr=%s\n", repositoryParameter[1], repositoryParameter[2])
		fnr, serr := strconv.Atoi(repositoryParameter[2])
		if serr != nil {
			return nil, serr
		}
		adabasToMap, err = NewAdabas(repositoryParameter[1], adabasID)
		if err != nil {
			return nil, err
		}
		adatypes.Central.Log.Debugf("Created adabas reference")
		repository = NewMapRepository(adabasToMap.URL, Fnr(fnr))
		adatypes.Central.Log.Debugf("Created repository")
	}

	connection = &Connection{adabasToData: adabasToData, ID: adabasID,
		adabasToMap: adabasToMap, adabasMap: adabasMap, repository: repository}

	if mapName != "" {
		connection.searchRepository(adabasID, repository, mapName)
		if err != nil {
			return nil, err
		}
	}

	adatypes.Central.Log.Debugf("Ready created connection handle %s", connection.String())
	return
}

// search the repository for a map name
func (connection *Connection) searchRepository(adabasID *ID, repository *Repository,
	mapName string) (err error) {
	if repository == nil {
		adatypes.Central.Log.Debugf("Search in global repositories")
		connection.adabasToMap, err = NewAdabas("1", adabasID)
		if err != nil {
			return err
		}
		connection.adabasMap, _, err = SearchMapRepository(connection.adabasToMap, mapName)
		if err != nil {
			adatypes.Central.Log.Debugf("Search in global repositories fail: %v", err)
			return err
		}
		if connection.adabasMap == nil {
			return adatypes.NewGenericError(85, mapName)
		}
	} else {
		adatypes.Central.Log.Debugf("Search in given repository %v: %s", repository, repository.DatabaseURL.URL.String())
		connection.adabasToMap, err = NewAdabas(repository.DatabaseURL.URL.String(), adabasID)
		if err != nil {
			adatypes.Central.Log.Debugf("New Adabas to map ID error: %v", err)
			return err
		}
		connection.adabasMap, err = repository.SearchMap(connection.adabasToMap, mapName)
		if err != nil {
			adatypes.Central.Log.Debugf("Search map error: %v", err)
			return err
		}
		// 	connection.adabasMap = NewAdabasMap(mapName, &repository.DatabaseURL)
		// 	if connection.adabasMap == nil {
		// 		return adatypes.NewGenericError(85, mapName)
		// 	}
		// 	connection.adabasToMap, err = NewAdabas(connection.adabasMap.URL(), adabasID)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}
	adatypes.Central.Log.Debugf("Found map %s\n", connection.adabasMap.Name)
	if connection.adabasMap.URL().String() == connection.adabasMap.Data.URL.String() {
		adatypes.Central.Log.Debugf("Different URL %v", connection.adabasMap.URL().String())
		connection.adabasToData = connection.adabasToMap
	} else {
		adatypes.Central.Log.Debugf("Create new Adabas URL %v!=%v", connection.adabasMap.URL().String(), connection.adabasMap.Data.URL.String())
		connection.adabasToMap, err = NewAdabas(&connection.adabasMap.Data.URL, adabasID)
		if err != nil {
			adatypes.Central.Log.Debugf("Error new ADabas URL %v", err)
			return err
		}
	}
	adatypes.Central.Log.Debugf("Final error: %v", err)
	return
}

// parse the authentication credentials in the connection string
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

// Close close Adabas session
func (connection *Connection) Close() {
	if connection.adabasToData != nil {
		connection.adabasToData.BackoutTransaction()
		connection.adabasToData.Close()
	}
	if connection.adabasToMap != nil {
		connection.adabasToMap.BackoutTransaction()
		connection.adabasToMap.Close()
	}
}

// EndTransaction current transaction is finally stored in the database
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
	if connection.adabasToMap != nil {
		err := connection.adabasToMap.ReleaseHold(connection.adabasMap.Data.Fnr)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReleaseCID any database command id resources, like command id caches assigned to a user
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

// AddCredential add user id and password credentials
func (connection *Connection) AddCredential(user string, pwd string) {
	connection.ID.AddCredential(user, pwd)
}

// CreateReadRequest create a read request
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

// CreateFileReadRequest create a read request
func (connection *Connection) CreateFileReadRequest(fnr Fnr) (*ReadRequest, error) {
	adatypes.Central.Log.Debugf("Connection: %#v", connection)
	adatypes.Central.Log.Debugf("Data referenced : %#v", connection.adabasToData)
	return NewReadRequest(connection.adabasToData, fnr)
}

// CreateMapReadRequest create a read request using a given map
func (connection *Connection) CreateMapReadRequest(mapReference interface{}) (request *ReadRequest, err error) {
	t := reflect.TypeOf(mapReference)
	switch t.Kind() {
	case reflect.Ptr, reflect.Struct:
		if connection.repository == nil {
			request, err = NewReadRequest(mapReference, connection.adabasToMap)
		} else {
			request, err = NewReadRequest(mapReference, connection.adabasToMap, connection.repository)
		}
		if err != nil {
			return
		}
		connection.fnr = request.adabasMap.Data.Fnr
		connection.adabasMap = request.adabasMap
	case reflect.String:
		m := mapReference.(string)
		err = connection.prepareMapUsage(m)
		if err != nil {
			return
		}
		connection.fnr = connection.adabasMap.Data.Fnr
		adatypes.Central.Log.Debugf("Map referenced : %#v", connection.adabasMap)
		request, err = NewReadRequest(connection.adabasToData, connection.adabasMap)
	default:
		return nil, adatypes.NewGenericError(0)
	}
	return
}

// CreateStoreRequest create a store request
func (connection *Connection) CreateStoreRequest(fnr Fnr) (*StoreRequest, error) {
	return NewStoreRequestAdabas(connection.adabasToData, fnr), nil
}

func (connection *Connection) prepareMapUsage(mapName string) (err error) {
	return connection.searchRepository(connection.ID, connection.repository, mapName)
	// if err == nil {
	// 	return nil
	// }
	// if connection.repository == nil {
	// 	return adatypes.NewGenericError(5)
	// }
	// // TODO search global enable
	// adatypes.Central.Log.Debugf("Search Map : %s platform: %v", mapName, connection.adabasToMap.ID.platform)
	// connection.adabasMap, err = connection.repository.SearchMap(connection.adabasToMap, mapName)
	// if err != nil {
	// 	return
	// }
	// if connection.adabasMap == nil {
	// 	err = adatypes.NewGenericError(6, mapName)
	// 	return
	// }
	// // Reuse Adabas handle
	// if connection.adabasMap.Repository.URL.String() == connection.adabasMap.Data.URL.String() {
	// 	connection.adabasToData = connection.adabasToMap
	// }
	// adatypes.Central.Log.Debugf("Found Adabas : %p", connection.adabasToData)
	// if connection.adabasToData != nil {
	// 	adatypes.Central.Log.Debugf("Found Adabas Map : %s", connection.adabasToData.URL.String())
	// }
	// adatypes.Central.Log.Debugf("Data Repository : %s", connection.adabasMap.Data.URL.String())
	// if connection.adabasToData == nil || connection.adabasToData.URL.String() != connection.adabasMap.Data.URL.String() {
	// 	adatypes.Central.Log.Debugf("Create new Adabas")
	// 	connection.adabasToData, err = NewAdabas(connection.adabasMap.URL(), connection.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// adatypes.Central.Log.Debugf("Platform Map : %#v", connection.adabasToMap.ID.platform)
	// adatypes.Central.Log.Debugf("Platform Adabas : %#v", connection.adabasToData.ID.platform)
	// return nil
}

// CreateMapStoreRequest create a store request using map name
func (connection *Connection) CreateMapStoreRequest(mapReference interface{}) (request *StoreRequest, err error) {
	t := reflect.TypeOf(mapReference)
	switch t.Kind() {
	case reflect.Ptr, reflect.Struct:
		request, err = NewStoreRequest(mapReference, connection.adabasToMap, connection.repository)
		if err != nil {
			return
		}
		connection.fnr = request.adabasMap.Data.Fnr
		connection.adabasMap = request.adabasMap
	case reflect.String:
		mapName := mapReference.(string)
		err = connection.prepareMapUsage(mapName)
		if err != nil {
			return
		}
		request, err = NewAdabasMapNameStoreRequest(connection.adabasToData, connection.adabasMap)
	}
	return
}

// CreateDeleteRequest create a delete request
func (connection *Connection) CreateDeleteRequest(fnr Fnr) (*DeleteRequest, error) {
	return NewDeleteRequestAdabas(connection.adabasToData, fnr), nil
}

// CreateMapDeleteRequest create a read request using a given map
func (connection *Connection) CreateMapDeleteRequest(mapName string) (request *DeleteRequest, err error) {
	if connection.repository == nil {
		err = adatypes.NewGenericError(9)
		return
	}
	connection.repository.SearchMapInRepository(connection.adabasToMap, mapName)
	if connection.adabasMap == nil {
		err = adatypes.NewGenericError(8, mapName)
		return
	}
	connection.adabasToData, err = NewAdabas(connection.adabasMap.URL(), connection.ID)
	if err != nil {
		//err = adatypes.NewGenericError(10)
		return
	}
	connection.fnr = connection.adabasMap.Data.Fnr
	adatypes.Central.Log.Debugf("Connection FNR=%d, Map referenced : %#v", connection.fnr, connection.adabasMap)
	request, err = NewMapDeleteRequest(connection.adabasToData, connection.adabasMap)
	return
}
