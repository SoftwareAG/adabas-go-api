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
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// URL define the desination of the host. Possible types are
//
//  - Local call with Driver="" and Port=0
//  - Entire Network calls with Driver="tcpip" and corresponding host and port
//  - Adabas TCP/IP calls with Driver="adatcp" and corresponding host and port
//
// Dependent on the Driver the corresponding connection is used. To use the local
// call access the Adabas Client native library is used.
type URL struct {
	Dbid   Dbid
	Driver string
	Host   string
	Port   uint32
}

// NewURLWithDbid create a new URL based on the database id only. Simple local access
// to the database
func NewURLWithDbid(dbid Dbid) *URL {
	return &URL{Dbid: dbid}
}

// NewURL create a URL based on a input string
func NewURL(url string) (*URL, error) {
	URL := &URL{}
	err := URL.examineURL(url)
	if err != nil {
		return nil, err
	}
	return URL, nil
}

func (URL *URL) examineURL(url string) error {
	adatypes.Central.Log.Debugf("New Adabas URL %s", url)
	re := regexp.MustCompile(`([0-9]+)\((\w*):\/\/([^:]*?):([0-9]*)\)`)
	match := re.FindStringSubmatch(url)
	if len(match) == 0 {
		dbid, err := strconv.Atoi(url)
		if err != nil {
			adatypes.Central.Log.Debugf("No numeric: %v", err)
			err = adatypes.NewGenericError(70, url)
			return err
		}
		URL.Dbid = Dbid(dbid)
		return nil
	}
	if len(match) < 4 {
		return adatypes.NewGenericError(71)
	}

	dbid, err := strconv.Atoi(match[1])
	if err != nil {
		adatypes.Central.Log.Debugf("Dbid not numeric: %v", err)
		err = adatypes.NewGenericError(70, match[1])
		return err
	}
	port, err := strconv.Atoi(match[4])
	if err != nil {
		adatypes.Central.Log.Debugf("Port not numeric: %v", err)
		err = adatypes.NewGenericError(72, match[4])
		return err
	}
	URL.Dbid = Dbid(dbid)
	URL.Port = uint32(port)
	if URL.Port > 0 {
		URL.Driver = strings.ToLower(match[2])
		switch URL.Driver {
		case "adatcp":
		case "tcpip":
			err = adatypes.NewGenericError(115)
			return err
		default:
			err = adatypes.NewGenericError(99, URL.Driver)
			return err
		}
		URL.Host = match[3]
	}
	return nil
}

// URL URL representation containing the TCP/IP host and port part only
func (URL URL) URL() string {
	return URL.Host + ":" + strconv.Itoa(int(URL.Port))
}

// String Full reference of the URL, like 123(adatcp://hostname:port)
func (URL URL) String() string {
	if URL.Driver == "" || URL.Port == 0 {
		return strconv.Itoa(int(URL.Dbid))
	}
	return strconv.Itoa(int(URL.Dbid)) + "(" + URL.Driver + "://" + URL.URL() + ")"
}

// UnmarshalJSON unmarshal JSON code
func (URL *URL) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	adatypes.Central.Log.Debugf("Got " + v)
	return URL.examineURL(v)
}

func (URL *URL) searchCertificate() []string {
	var pair []string
	cert := os.Getenv("ADABAS_CLIENT_CERT")
	if cert == "" {
		return nil
	}
	adatypes.Central.Log.Debugf("Add certificate file %s", cert)
	pair = append(pair, cert)
	key := os.Getenv("ADABAS_CLIENT_KEY")
	if key == "" {
		return nil
	}
	adatypes.Central.Log.Debugf("Add key file %s", key)
	pair = append(pair, key)
	return pair
}
