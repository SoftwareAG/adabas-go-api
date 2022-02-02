/*
* Copyright © 2021-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// UpdateLOBRecord update lob records in an stream, call will write segment to offset of LOB
func (request *StoreRequest) UpdateLOBRecord(isn adatypes.Isn, field string, offset uint64, data []byte) (err error) {
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Store LOB record initiated ...")
	}
	err = request.Open()
	if err != nil {
		return
	}
	err = request.StoreFields(field)
	if err != nil {
		adatypes.Central.Log.Debugf("Store fields error ...%#v", err)
		return err
	}
	if debug {
		adatypes.Central.Log.Debugf("LOB Definition generated ...BlockSize=%d", len(data))
	}
	var record *Record
	record, err = request.CreateRecord()
	if err != nil {
		return
	}
	record.Isn = isn
	err = record.SetPartialValue(field, uint32(offset+1), data)
	if err != nil {
		adatypes.Central.Log.Debugf("Set partial value error ...%#v", err)
		return err
	}
	if debug {
		adatypes.Central.Log.Debugf("Update LOB with ...%#v", field)
	}

	adabasRequest, prepareErr := request.prepareRequest(false)
	if prepareErr != nil {
		return prepareErr
	}
	err = request.update(adabasRequest, record)
	if debug {
		adatypes.Central.Log.Debugf("Error reading %v", err)
	}

	return err
}
