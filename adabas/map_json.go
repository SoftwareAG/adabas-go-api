/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// MapFile parse map JSON import/export files
type MapFile struct {
	Maps []*Map `json:"Maps"`
}

// ParseJSONFileForFields Parse JSON map export file
func ParseJSONFileForFields(file *os.File) (mapList []*Map, err error) {

	byteValue, _ := ioutil.ReadAll(file)

	var mapFile MapFile
	err = json.Unmarshal([]byte(byteValue), &mapFile)
	if err != nil {
		adatypes.Central.Log.Debugf("Parse JSON error: %v", err)
		return nil, err
	}
	adatypes.Central.Log.Debugf("Number map entries %d", len(mapFile.Maps))
	mapList = mapFile.Maps
	return
}

// LoadJSONMap load JSON Map file and creates Map instance of that
func LoadJSONMap(file string) (maps []*Map, err error) {
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + "/" + file
	fmt.Println("Loading ...." + file)
	f, ferr := os.Open(name)
	if ferr != nil {
		return nil, ferr
	}
	defer f.Close()

	maps, err = ParseJSONFileForFields(f)
	if err != nil {
		fmt.Println("Error parsing file", err)
		return nil, err
	}
	return
}
