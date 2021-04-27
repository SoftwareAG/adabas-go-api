/*
* Copyright © 2018-2021 Software AG, Darmstadt, Germany and/or its licensors
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

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// ExportMapRepository import map by file import
func (repository *Repository) ExportMapRepository(ada *Adabas, filter string,
	fileName string) (err error) {
	var file *os.File
	file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	err = repository.LoadMapRepository(ada)
	if err != nil {
		return
	}

	mapExport := MapFile{}
	var maps []*Map
	maps, err = repository.LoadAllMaps(ada)
	if err != nil {
		return
	}
	if filter == "" || filter == "*" {
		mapExport.Maps = maps
	} else {
		for _, m := range maps {
			if matched, _ := regexp.MatchString(filter, m.Name); matched {
				mapExport.Maps = append(mapExport.Maps, m)
			}
		}
	}
	if len(mapExport.Maps) == 0 {
		panic("Adabas Map export list is empty")
	}
	var buffer []byte
	buffer, err = json.Marshal(mapExport)
	if err != nil {
		adatypes.Central.Log.Debugf("Generate JSON error: %v", err)
		return err
	}
	_, err = file.Write(buffer)
	return
}
