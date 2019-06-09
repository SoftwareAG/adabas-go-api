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
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

func TestRecord(t *testing.T) {
	initTestLogWithFile(t, "Record.log")

	resultNil, err := NewRecord(nil)
	assert.Error(t, err)
	assert.Nil(t, resultNil)

	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "U4"),
		adatypes.NewType(adatypes.FieldTypeByte, "B1"),
		adatypes.NewType(adatypes.FieldTypeUByte, "UB"),
		adatypes.NewType(adatypes.FieldTypeUInt2, "I2"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "U8"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "I8"),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	testDefinition.CreateValues(false)
	result, err := NewRecord(testDefinition)
	if assert.NoError(t, err) {
		assert.NotNil(t, result)
		assert.Equal(t, "ISN=0 quantity=0\n U4=\"0\"\n B1=\"0\"\n UB=\"0\"\n I2=\"0\"\n U8=\"0\"\n I8=\"0\"\n", result.String())
		v, verr := result.SearchValue("I2")
		assert.NoError(t, verr)
		assert.NotNil(t, v)
		assert.Equal(t, "0", v.String())
		verr = result.SetValue("I2", 100)
		assert.NoError(t, verr)
		assert.Equal(t, "100", v.String())
		verr = result.SetValue("I2", "100")
		assert.NoError(t, verr)
		assert.Equal(t, "100", v.String())
		v, verr = result.SearchValue("X1")
		assert.Error(t, verr)
		assert.Nil(t, v)
		v, verr = result.SearchValue("I8")
		assert.NoError(t, verr)
		assert.NotNil(t, v)
		assert.Equal(t, "0", v.String())
		v, verr = result.SearchValue("X1[1]")
		assert.Error(t, verr)
		assert.Nil(t, v)
		v, verr = result.SearchValue("X1[1,10]")
		assert.Error(t, verr)
		assert.Nil(t, v)
	}
}

func TestRecord_MarshalXML(t *testing.T) {
	initTestLogWithFile(t, "Record.log")

	resultNil, err := NewRecord(nil)
	assert.Error(t, err)
	assert.Nil(t, resultNil)

	layout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "S4", 100),
		adatypes.NewType(adatypes.FieldTypeUInt4, "U4"),
		adatypes.NewType(adatypes.FieldTypeByte, "B1"),
		adatypes.NewType(adatypes.FieldTypeUByte, "UB"),
		adatypes.NewType(adatypes.FieldTypeUInt2, "I2"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "U8"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "I8"),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	testDefinition.CreateValues(false)
	result, err := NewRecord(testDefinition)
	if assert.NoError(t, err) {
		verr := result.SetValue("I2", 100)
		assert.NoError(t, verr)
		verr = result.SetValue("I8", "1234567")
		assert.NoError(t, verr)
		verr = result.SetValue("U4", "4200")
		assert.NoError(t, verr)
		verr = result.SetValue("B1", "10")
		assert.NoError(t, verr)
		verr = result.SetValue("S4", "ABCabcdfegggggg")
		assert.NoError(t, verr)

		xout, xerr := xml.Marshal(result)
		assert.NoError(t, xerr)
		fmt.Println("XML:", string(xout))
		assert.Equal(t, "<Record><S4>ABCabcdfegggggg</S4><U4>4200</U4><B1>10</B1><UB>0</UB><I2>100</I2><U8>0</U8><I8>1234567</I8></Record>", string(xout))
		jout, jerr := json.Marshal(result)
		assert.NoError(t, jerr)
		fmt.Println("JSON:", string(jout))
		assert.Equal(t, "{\"B1\":10,\"I2\":100,\"I8\":1234567,\"S4\":\"ABCabcdfegggggg\",\"U4\":4200,\"U8\":0,\"UB\":0}", string(jout))
		result.adabasMap = NewAdabasMap("ABC", &DatabaseURL{})
		xout, xerr = xml.Marshal(result)
		assert.NoError(t, xerr)
		fmt.Println("XML:", string(xout))
		assert.Equal(t, "<ABC><S4>ABCabcdfegggggg</S4><U4>4200</U4><B1>10</B1><UB>0</UB><I2>100</I2><U8>0</U8><I8>1234567</I8></ABC>", string(xout))
		jout, jerr = json.Marshal(result)
		assert.NoError(t, jerr)
		assert.Equal(t, "{\"B1\":10,\"I2\":100,\"I8\":1234567,\"S4\":\"ABCabcdfegggggg\",\"U4\":4200,\"U8\":0,\"UB\":0}", string(jout))
		fmt.Println("JSON:", string(jout))

	}
}
