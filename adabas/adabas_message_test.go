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
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

func TestAdabasMessageGeneric(t *testing.T) {
	initTestLogWithFile(t, "messages.log")

	err := adatypes.NewGenericError(2, "XX")
	assert.Equal(t, "ADG0000002: Invalid Adabas command send: XX", err.Error())
	err = adatypes.NewGenericError(21, "TESTMAP")
	assert.Equal(t, "ADG0000021: Map TESTMAP not found", err.Error())

}

func TestAdabasMessage(t *testing.T) {
	initTestLogWithFile(t, "messages.log")

	assert.Equal(t, "Normal successful completion", adatypes.Translate("en", "ADAGE00000"))
	assert.Equal(t, "Invalid command ID value was detected", adatypes.Translate("en", "ADAGE15000"))
	assert.Equal(t, "Insufficient space in attached buffer", adatypes.Translate("en", "ADAGEFF000"))
	assert.Equal(t, "Adabas versuchte eine ISN in Hold zu setzen die gerade bei einem anderen User in Hold gesetzt ist.", adatypes.Translate("de", "ADAGE02001"))
	assert.Equal(t, "Funktion noch nicht implementiert", adatypes.Translate("de", "ADAGEFA000"))
	// Not available in german, use english one
	assert.Equal(t, "Compressed record area too small (internal error)", adatypes.Translate("de", "ADAGE05001"))

	// Not available at all, use empty response
	assert.Equal(t, "", adatypes.Translate("en", "ABC"))

}

func TestAdabasMessageError(t *testing.T) {
	initTestLogWithFile(t, "messages.log")

	ada, err := NewAdabas(21)
	assert.NoError(t, err)

	// Return: Hello, i18n
	assert.Equal(t, "ADAGE94000: Adabas is not active or accessible (rsp=148,subrsp=0,dbid=21,file=0)", NewError(ada).Error())
	ada.Acbx.Acbxrsp = AdaNormal
	assert.Equal(t, "ADAGE00000: Normal successful completion (rsp=0,subrsp=0,dbid=21,file=0)", NewError(ada).Error())
	ada.Acbx.Acbxrsp = AdaInvalidFileNumber
	assert.Equal(t, "ADAGE11000: Invalid or unauthorized file number (rsp=17,subrsp=0,dbid=21,file=0)", NewError(ada).Error())
	ada.Acbx.Acbxerrc = AdaISNNotSorted
	assert.Equal(t, "ADAGE11001: The program tried to access system file 1 or 2, and no OP command was issued. (rsp=17,subrsp=1,dbid=21,file=0)", NewError(ada).Error())
	ada.Acbx.Acbxrsp = 120
	ada.Acbx.Acbxerrc = 0
	assert.Equal(t, "ADAGE78000: Unknown response and subcode (rsp=120,subrsp=0,dbid=21,file=0)", NewError(ada).Error())
	m := []string{"ADAGE78000", "Unknown error response 120 subcode 0 (ADAGE78000) (rsp=120,subrsp=0,dbid=21,file=0)"}
	assert.Equal(t, m, ada.getAdabasMessage())

}

func TestAdabasMessageTranslate(t *testing.T) {
	initTestLogWithFile(t, "messages.log")

	ada, err := NewAdabas(21)
	assert.NoError(t, err)

	// Return: Hello, i18n
	assert.Equal(t, "Adabas ist nicht aktiv oder erreichbar (rsp=148,subrsp=0,dbid=21,file=0)", NewError(ada).Translate("de"))
	ada.Acbx.Acbxrsp = AdaNormal
	assert.Equal(t, "Erfolgreich beendet (rsp=0,subrsp=0,dbid=21,file=0)", NewError(ada).Translate("de"))
	ada.Acbx.Acbxrsp = AdaInvalidFileNumber
	assert.Equal(t, "Falscher oder nicht authorizierte File-Nummer (rsp=17,subrsp=0,dbid=21,file=0)", NewError(ada).Translate("de"))
	ada.Acbx.Acbxerrc = AdaISNNotSorted
	assert.Equal(t, "The program tried to access system file 1 or 2, and no OP command was issued. (rsp=17,subrsp=1,dbid=21,file=0)", NewError(ada).Translate("de"))
	ada.Acbx.Acbxrsp = 120
	ada.Acbx.Acbxerrc = 0
	assert.Equal(t, "Unbekannter Response und Subcode Wert (rsp=120,subrsp=0,dbid=21,file=0)", NewError(ada).Translate("de"))

}
