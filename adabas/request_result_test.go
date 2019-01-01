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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
)

type fields struct {
	values []*ResultRecord
}

func generateOneFields(t *testing.T) fields {
	oneField := fields{}
	v, err := adatypes.NewType(adatypes.FieldTypeByte, "AA").Value()
	assert.NoError(t, err)
	record := &ResultRecord{Value: []adatypes.IAdaValue{v}}
	oneField.values = append(oneField.values, record)
	return oneField
}

func TestRequestResult_NrRecords(t *testing.T) {
	noFields := fields{}
	oneField := generateOneFields(t)
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"NoField", noFields, 0},
		{"OneField", oneField, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestResult := &RequestResult{
				Values: tt.fields.values,
			}
			if got := requestResult.NrRecords(); got != tt.want {
				t.Errorf("RequestResult.NrRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestResult_String(t *testing.T) {
	noFields := fields{}
	oneField := generateOneFields(t)
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"NoFields", noFields, ""},
		{"OneField", oneField, "  AA = > 0 <\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestResult := &RequestResult{
				Values: tt.fields.values,
			}
			if got := requestResult.String(); got != tt.want {
				t.Errorf("RequestResult.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func generateDefinitionTest() *adatypes.Definition {
	groupLayout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeCharacter, "G1"),
		adatypes.NewType(adatypes.FieldTypeString, "GX"),
		adatypes.NewType(adatypes.FieldTypePacked, "PA"),
	}
	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "AA"),
		adatypes.NewType(adatypes.FieldTypeByte, "B1"),
		adatypes.NewType(adatypes.FieldTypeUByte, "UB"),
		adatypes.NewType(adatypes.FieldTypeUInt2, "I2"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "U8"),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "GR", adatypes.OccNone, groupLayout),
		adatypes.NewType(adatypes.FieldTypeUInt8, "I8"),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	err := testDefinition.CreateValues(false)
	if err != nil {
		fmt.Println("Error creating values")
		return nil
	}
	return testDefinition
}

func generateResult() *RequestResult {
	d := generateDefinitionTest()
	result := &RequestResult{}
	record, err := NewResultRecord(d)
	if err != nil {
		fmt.Println("Error generating result record", err)
		return nil
	}
	record.Isn = 10
	err = record.SetValue("AA", 10)
	if err != nil {
		fmt.Println("Error setting value AA record", err)
		return nil
	}
	err = record.SetValue("PA", 9)
	if err != nil {
		fmt.Println("Error setting value PA record", err)
		return nil
	}
	result.Values = append(result.Values, record)
	record, err = NewResultRecord(d)
	if err != nil {
		fmt.Println("Error generating result record", err)
		return nil
	}
	record.Isn = 11
	record.SetValue("AA", 20)
	record.SetValue("PA", 3)
	result.Values = append(result.Values, record)
	return result
}

func TestJson(t *testing.T) {
	r := []byte("{\"Record\":[{\"AA\":\"11100301\",\"AB\":{\"AC\":\"HANS                \",\"AD\":\"WILHELM             \",\"AE\":\"BERGMANN            \"},\"ISN\":251},{\"AA\":\"11100302\",\"AB\":{\"AC\":\"ROSWITHA            \",\"AD\":\"ELLEN               \",\"AE\":\"HAIBACH             \"},\"ISN\":383}]}")
	result := &RequestResult{}
	err := json.Unmarshal(r, result)
	if !assert.NoError(t, err) {
		return
	}
	res, jerr := json.Marshal(result)
	if !assert.NoError(t, jerr) {
		return
	}
	fmt.Println(string(res))
	assert.Equal(t, r, res)
}

func ExampleRequestResult_JsonMarshal() {
	result := generateResult()
	res, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error generating JSON", err)
		return
	}
	fmt.Println(string(res))
	// Output: {"Record":[{"AA":"10","B1":0,"GR":{"G1":"0","GX":" ","PA":9},"I2":"0","I8":"0","ISN":10,"U8":"0","UB":"0"},{"AA":"20","B1":0,"GR":{"G1":"0","GX":" ","PA":3},"I2":"0","I8":"0","ISN":11,"U8":"0","UB":"0"}]}
}

func ExampleRequestResult_XmlMarshal() {
	f, ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}
	defer f.Close()

	result := generateResult()
	res, err := xml.Marshal(result)
	if err != nil {
		fmt.Println("Error generating XML", err)
		return
	}
	fmt.Println(string(res))
	// Output: <Response><Record ISN="10"><AA>10</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>9</PA></GR><I8>0</I8></Record><Record ISN="11"><AA>20</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>3</PA></GR><I8>0</I8></Record></Response>
}
