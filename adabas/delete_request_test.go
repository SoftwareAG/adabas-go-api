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
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

func testCallback(adabasRequest *adatypes.AdabasRequest, x interface{}) (err error) {
	deleteRequest := x.(*DeleteRequest)
	isn := adabasRequest.Isn
	fmt.Printf("Delete ISN: %d on %s/%d\n", adabasRequest.Isn, deleteRequest.repository.URL.String(), deleteRequest.repository.Fnr)
	deleteRequest.Delete(isn)
	return nil
}

func clearFile(file uint32) error {
	connection, err := NewConnection("acj;target=23")
	if err != nil {
		return err
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(file)
	if err != nil {
		return rErr
	}
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(file)
	if dErr != nil {
		return dErr
	}
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	if err != nil {
		return err
	}
	err = deleteRequest.EndTransaction()
	if err != nil {
		return err
	}

	return nil
}

func TestDeleteRequestRefreshFile16(t *testing.T) {
	f := initTestLogWithFile(t, "delete.log")
	defer f.Close()

	cErr := clearFile(16)
	assert.NoError(t, cErr)

}
