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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

func TestAdabasMessageGeneric(t *testing.T) {
	f := initTestLogWithFile(t, "messages.log")
	defer f.Close()

	err := adatypes.NewGenericError(2, "XX")
	assert.Equal(t, "ADG0000002: Invalid Adabas command send: XX", err.Error())
	err = adatypes.NewGenericError(21, "TESTMAP")
	assert.Equal(t, "ADG0000021: Map TESTMAP not found", err.Error())

}

func TestAdabasMessage(t *testing.T) {
	f := initTestLogWithFile(t, "messages.log")
	defer f.Close()

	// Return: Hello, i18n
	assert.Equal(t, "Normal successful completion", adatypes.Translate("en", "ADAGE00000"))
	assert.Equal(t, "Invalid command ID value was detected", adatypes.Translate("en", "ADAGE15000"))
	assert.Equal(t, "Insufficient space in attached buffer", adatypes.Translate("en", "ADAGEFF000"))
	assert.Equal(t, "Unknown message for code: ABC", adatypes.Translate("en", "ABC"))

}
