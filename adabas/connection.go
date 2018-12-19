/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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
	"regexp"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// Connection Adabas connection context
type Connection struct {
	ID           ID
	adabasToData *Adabas
	adabasMap    *Map
	adabasToMap  *Adabas
	fnr          int
	repository   *Repository
}

// NewConnection create new Adabas connection instance
func NewConnection(connectionString string) (*Connection, error) {
	adabasID := NewAdabasID()
	return NewConnectionID(connectionString, adabasID)
}

// NewConnectionID create new Adabas connection instance providing a Adabas ID
func NewConnectionID(connectionString string, adabasID *ID) (connection *Connection, err error) {
	parts := strings.Split(connectionString, ";")
	if parts[0] != "acj" {
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
		case p == "acj":
		case strings.HasPrefix(p, "target="):
			target := strings.Split(parts[1], "=")
			adatypes.Central.Log.Debugf("Connection to target : %s", target[1])
			adabasToData, err = NewAdabasWithID(target[1], adabasID)
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
			return nil, nil
		}
	}

	if len(repositoryParameter) > 2 {
		adatypes.Central.Log.Debugf("Add repository search of dbid=%s fnr=%s\n", repositoryParameter[1], repositoryParameter[2])
		fnr, serr := strconv.Atoi(repositoryParameter[2])
		if serr != nil {
			return nil, serr
		}
		adabasToMap, err = NewAdabasWithID(repositoryParameter[1], adabasID)
		if err != nil {
			return nil, err
		}
		adatypes.Central.Log.Debugf("Created adabas reference")
		repository = NewMapRepository(adabasToMap, uint32(fnr))
		adatypes.Central.Log.Debugf("Created repository")
		// repository.LoadMapRepository()
	}

	if mapName != "" {
		adatypes.Central.Log.Debugf("Create map for %s\n", mapName)
		adabasMap = NewAdabasMap(mapName, &repository.DatabaseURL)
		adabasToMap = NewAdabasWithURL(adabasMap.URL(), adabasID)
	}
	adatypes.Central.Log.Debugf("Ready creating connection handle")

	connection = &Connection{adabasToData: adabasToData, ID: *adabasID,
		adabasToMap: adabasToMap, adabasMap: adabasMap, repository: repository}
	return
}

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
			id.setHost(v)
		case "user":
			id.setUser(v)
		case "id":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			id.setID(uint32(i))
		}
	}
	return nil
}

func (connection *Connection) String() string {
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

// Open open adabas session
func (connection *Connection) Open() error {
	err := connection.adabasToData.Open()
	return err
}

// Close close adabas session
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

// Release release any database resources, like command id caches assigned to a user
func (connection *Connection) Release() error {
	if connection.adabasToData != nil {
		err := connection.adabasToData.Release()
		if err != nil {
			return err
		}
	}
	if connection.adabasToMap != nil {
		err := connection.adabasToMap.Release()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateReadRequest create a read request
func (connection *Connection) CreateReadRequest(fnr uint32) (*ReadRequest, error) {
	return NewRequestAdabas(connection.adabasToData, fnr), nil
}

// CreateMapReadRequest create a read request using a given map
func (connection *Connection) CreateMapReadRequest(mapName string) (request *ReadRequest, err error) {
	err = connection.prepareMapUsage(mapName)
	if err != nil {
		return
	}
	// if err != nil {
	// 	err = adatypes.NewGenericError(7)
	// 	return
	// }
	connection.fnr = int(connection.adabasMap.Data.Fnr)
	adatypes.Central.Log.Debugf("Map referenced : %#v", connection.adabasMap)
	request, err = NewAdabasMapNameRequest(connection.adabasToData, connection.adabasMap)
	return
}

// CreateStoreRequest create a store request
func (connection *Connection) CreateStoreRequest(fnr uint32) (*StoreRequest, error) {
	return NewStoreRequestAdabas(connection.adabasToData, fnr), nil
}

func (connection *Connection) prepareMapUsage(mapName string) (err error) {
	if connection.repository == nil {
		return adatypes.NewGenericError(5)
	}
	// TODO search global enable
	adatypes.Central.Log.Debugf("Search Map : %s", mapName)
	connection.adabasMap, err = connection.repository.SearchMap(connection.adabasToMap, mapName)
	if err != nil {
		return
	}
	if connection.adabasMap == nil {
		err = adatypes.NewGenericError(6, mapName)
		return
	}
	// Reuse Adabas handle
	if connection.adabasMap.Repository.URL.String() == connection.adabasMap.Data.URL.String() {
		connection.adabasToData = connection.adabasToMap
	}
	adatypes.Central.Log.Debugf("Found Adabas : %p", connection.adabasToData)
	if connection.adabasToData != nil {
		adatypes.Central.Log.Debugf("Found Adabas Map : %s", connection.adabasToData.URL.String())
	}
	adatypes.Central.Log.Debugf("Data Repository : %s", connection.adabasMap.Data.URL.String())
	if connection.adabasToData == nil || connection.adabasToData.URL.String() != connection.adabasMap.Data.URL.String() {
		adatypes.Central.Log.Debugf("Create new Adabas")
		connection.adabasToData = NewAdabasWithURL(connection.adabasMap.URL(), &connection.ID)
	}
	return nil
}

// CreateMapStoreRequest create a store request using map name
func (connection *Connection) CreateMapStoreRequest(mapName string) (request *StoreRequest, err error) {
	err = connection.prepareMapUsage(mapName)
	if err != nil {
		return
	}
	request, err = NewAdabasMapNameStoreRequest(connection.adabasToData, connection.adabasMap)
	return
}

// CreateDeleteRequest create a delete request
func (connection *Connection) CreateDeleteRequest(fnr uint32) (*DeleteRequest, error) {
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
	connection.adabasToData = NewAdabasWithURL(connection.adabasMap.URL(), &connection.ID)
	if err != nil {
		err = adatypes.NewGenericError(10)
		return
	}
	connection.fnr = int(connection.adabasMap.Data.Fnr)
	adatypes.Central.Log.Debugf("Connection FNR=%d, Map referenced : %#v", connection.fnr, connection.adabasMap)
	request, err = NewMapDeleteRequest(connection.adabasToData, connection.adabasMap)
	return
}