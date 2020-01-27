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

type fields struct {
	values []*Record
}

func generateOneFields(t *testing.T) fields {
	oneField := fields{}
	v, err := adatypes.NewType(adatypes.FieldTypeByte, "AA").Value()
	assert.NoError(t, err)
	record := &Record{Value: []adatypes.IAdaValue{v}}
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
			Response := &Response{
				Values: tt.fields.values,
			}
			if got := Response.NrRecords(); got != tt.want {
				t.Errorf("Response.NrRecords() = %v, want %v", got, tt.want)
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
			Response := &Response{
				Values: tt.fields.values,
			}
			if got := Response.String(); got != tt.want {
				t.Errorf("Response.String() = %v, want %v", got, tt.want)
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

func generateMUDefinitionTest() *adatypes.Definition {
	muLayout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "MU", 10),
	}
	groupLayout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypePacked, "PA"),
	}
	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "AA"),
		adatypes.NewType(adatypes.FieldTypeByte, "B1"),
		adatypes.NewType(adatypes.FieldTypeUByte, "UB"),
		adatypes.NewType(adatypes.FieldTypeUInt2, "I2"),
		adatypes.NewType(adatypes.FieldTypeUInt4, "U4"),
		adatypes.NewType(adatypes.FieldTypeUInt8, "U8"),
		adatypes.NewStructureList(adatypes.FieldTypeMultiplefield, "MU", adatypes.OccNone, muLayout),
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

func generatePEMUDefinitionTest() *adatypes.Definition {
	muLayout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "MU"),
	}
	groupLayout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypePacked, "PA"),
		adatypes.NewType(adatypes.FieldTypeUInt4, "PG"),
	}
	periodgroupLayout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypePacked, "PP", 2),
		adatypes.NewStructureList(adatypes.FieldTypeMultiplefield, "MU", adatypes.OccNone, muLayout),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "GR", adatypes.OccNone, groupLayout),
		adatypes.NewType(adatypes.FieldTypeUInt8, "G8"),
	}
	periodgroup2Layout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypePacked, "PX", 8),
		adatypes.NewType(adatypes.FieldTypeUInt8, "PY"),
	}
	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "AA"),
		adatypes.NewStructureList(adatypes.FieldTypePeriodGroup, "PE", adatypes.OccNone, periodgroupLayout),
		adatypes.NewType(adatypes.FieldTypeUInt8, "U8"),
		adatypes.NewStructureList(adatypes.FieldTypePeriodGroup, "P2", adatypes.OccNone, periodgroup2Layout),
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

func generateResult() *Response {
	d := generateDefinitionTest()
	result := &Response{}
	record, err := NewRecord(d)
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
	record, err = NewRecord(d)
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

// func TestJson(t *testing.T) {
// 	r := []byte("{\"Record\":[{\"AA\":\"11100301\",\"AB\":{\"AC\":\"HANS                \",\"AD\":\"WILHELM             \",\"AE\":\"BERGMANN            \"},\"ISN\":251},{\"AA\":\"11100302\",\"AB\":{\"AC\":\"ROSWITHA            \",\"AD\":\"ELLEN               \",\"AE\":\"HAIBACH             \"},\"ISN\":383}]}")
// 	result := &Response{}
// 	err := json.Unmarshal(r, result)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	res, jerr := json.Marshal(result)
// 	if !assert.NoError(t, jerr) {
// 		return
// 	}
// 	fmt.Println(string(res))
// 	assert.Equal(t, r, res)
// }

func ExampleResponse_jsonMarshal() {
	result := generateResult()
	res, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error generating JSON", err)
		return
	}
	fmt.Println(string(res))
	// Output: {"Records":[{"AA":10,"B1":0,"GR":{"G1":"0","GX":"","PA":9},"I2":0,"I8":0,"ISN":10,"U8":0,"UB":0},{"AA":20,"B1":0,"GR":{"G1":"0","GX":"","PA":3},"I2":0,"I8":0,"ISN":11,"U8":0,"UB":0}]}
}

func ExampleResponse_xmlMarshal() {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	result := generateResult()
	res, err := xml.Marshal(result)
	if err != nil {
		fmt.Println("Error generating XML", err)
		return
	}
	fmt.Println(string(res))
	// Output: <Response><Record ISN="10"><AA>10</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>9</PA></GR><I8>0</I8></Record><Record ISN="11"><AA>20</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>3</PA></GR><I8>0</I8></Record></Response>
}

func TestRequestResult(t *testing.T) {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	result := generateResult()
	record := result.Isn(10)
	assert.NotNil(t, record)
	record = result.Isn(100)
	assert.Nil(t, record)
	j, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.Equal(t, "{\"Records\":[{\"AA\":10,\"B1\":0,\"GR\":{\"G1\":\"0\",\"GX\":\"\",\"PA\":9},\"I2\":0,\"I8\":0,\"ISN\":10,\"U8\":0,\"UB\":0},{\"AA\":20,\"B1\":0,\"GR\":{\"G1\":\"0\",\"GX\":\"\",\"PA\":3},\"I2\":0,\"I8\":0,\"ISN\":11,\"U8\":0,\"UB\":0}]}", string(j))
	x, err := xml.Marshal(result)
	assert.NoError(t, err)
	assert.Equal(t, "<Response><Record ISN=\"10\"><AA>10</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>9</PA></GR><I8>0</I8></Record><Record ISN=\"11\"><AA>20</AA><B1>0</B1><UB>0</UB><I2>0</I2><U8>0</U8><GR><G1>0</G1><GX> </GX><PA>3</PA></GR><I8>0</I8></Record></Response>", string(x))
}

func TestRequestResultWithMU(t *testing.T) {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generateMUDefinitionTest()
	record, err := NewRecord(d)
	if !assert.NoError(t, err) {
		fmt.Println("Result record generation error", err)
		return
	}

	fmt.Println("Test request result:")

	j, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "{\"AA\":0,\"B1\":0,\"GR\":{\"PA\":0},\"I2\":0,\"I8\":0,\"MU\":[],\"U4\":0,\"U8\":0,\"UB\":0}", string(j))
	x, err := xml.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "<Record><AA>0</AA><B1>0</B1><UB>0</UB><I2>0</I2><U4>0</U4><U8>0</U8><Multiple sn=\"MU\"></Multiple><Group sn=\"GR\"><PA>0</PA></Group><I8>0</I8></Record>", string(x))
}

func TestRequestResultWithMUWithContent(t *testing.T) {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generateMUDefinitionTest()
	record, err := NewRecord(d)
	if !assert.NoError(t, err) {
		fmt.Println("Result record generation error", err)
		return
	}

	for i := uint32(0); i < 5; i++ {
		adatypes.Central.Log.Infof("Set MU entry of %d", (i + 1))
		err = record.SetValueWithIndex("MU", []uint32{i + 1}, fmt.Sprintf("AAX%03d", (i+1)))
		if !assert.NoError(t, err) {
			fmt.Println("Set MU error", err)
			return
		}
	}
	err = record.SetValue("AA", 2)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}
	err = record.SetValue("B1", 3)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}
	err = record.SetValue("PA", 1)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}

	// fmt.Println("Test request result:")
	// record.DumpValues()
	j, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "{\"AA\":2,\"B1\":3,\"GR\":{\"PA\":1},\"I2\":0,\"I8\":0,\"MU\":[\"AAX001\",\"AAX002\",\"AAX003\",\"AAX004\",\"AAX005\"],\"U4\":0,\"U8\":0,\"UB\":0}", string(j))
	x, err := xml.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "<Record><AA>2</AA><B1>3</B1><UB>0</UB><I2>0</I2><U4>0</U4><U8>0</U8><Multiple sn=\"MU\"><MU>AAX001</MU><MU>AAX002</MU><MU>AAX003</MU><MU>AAX004</MU><MU>AAX005</MU></Multiple><Group sn=\"GR\"><PA>1</PA></Group><I8>0</I8></Record>", string(x))
}

func TestRequestResultWithPEMUWithoutContent(t *testing.T) {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generatePEMUDefinitionTest()
	record, err := NewRecord(d)
	if !assert.NoError(t, err) {
		fmt.Println("Result record generation error", err)
		return
	}

	// fmt.Println("Test request result:")
	// record.DumpValues()
	j, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "{\"AA\":0,\"I8\":0,\"P2\":[],\"PE\":[],\"U8\":0}", string(j))
	x, err := xml.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "<Record><AA>0</AA><Period sn=\"PE\"></Period><U8>0</U8><Period sn=\"P2\"></Period><I8>0</I8></Record>", string(x))
}

func TestRequestResultWithPEMUWithContent(t *testing.T) {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generatePEMUDefinitionTest()
	record, err := NewRecord(d)
	if !assert.NoError(t, err) {
		fmt.Println("Result record generation error", err)
		return
	}

	for i := uint32(0); i < 3; i++ {
		adatypes.Central.Log.Infof("Set period group entry of %d", (i + 1))
		err = record.SetValueWithIndex("PP", []uint32{i + 1}, (i + 1))
		if !assert.NoError(t, err) {
			fmt.Println("Set MU error", err)
			return
		}
	}
	err = record.SetValueWithIndex("MU", []uint32{1, 1}, 100)
	if !assert.NoError(t, err) {
		fmt.Println("Set MU error", err)
		return
	}
	err = record.SetValueWithIndex("MU", []uint32{1, 2}, 122)
	if !assert.NoError(t, err) {
		fmt.Println("Set MU error", err)
		return
	}
	err = record.SetValue("AA", 2)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}
	err = record.SetValue("U8", 3)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}
	err = record.SetValue("I8", 1)
	if !assert.NoError(t, err) {
		fmt.Println("Set PA error", err)
		return
	}

	// fmt.Println("Test request result:")
	// record.DumpValues()
	j, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "{\"AA\":2,\"I8\":1,\"P2\":[],\"PE\":[{\"G8\":0,\"GR\":{\"PA\":0,\"PG\":0},\"MU\":[100,122],\"PP\":1},{\"G8\":0,\"GR\":{\"PA\":0,\"PG\":0},\"MU\":[],\"PP\":2},{\"G8\":0,\"GR\":{\"PA\":0,\"PG\":0},\"MU\":[],\"PP\":3}],\"U8\":3}", string(j))
	x, err := xml.Marshal(record)
	assert.NoError(t, err)
	assert.Equal(t, "<Record><AA>2</AA><Period sn=\"PE\"><Entry><PP>1</PP><Multiple sn=\"MU\"><MU>100</MU><MU>122</MU></Multiple><Group sn=\"GR\"><PA>0</PA><PG>0</PG></Group><G8>0</G8></Entry><Entry><PP>2</PP><Multiple sn=\"MU\"></Multiple><Group sn=\"GR\"><PA>0</PA><PG>0</PG></Group><G8>0</G8></Entry><Entry><PP>3</PP><Multiple sn=\"MU\"></Multiple><Group sn=\"GR\"><PA>0</PA><PG>0</PG></Group><G8>0</G8></Entry></Period><U8>3</U8><Period sn=\"P2\"></Period><I8>1</I8></Record>", string(x))
}

func ExampleRecord_dumpZeroValues() {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generatePEMUDefinitionTest()
	record, err := NewRecord(d)
	if err != nil {
		fmt.Println("Result record generation error", err)
		return
	}

	fmt.Println("Dump request result:")
	record.DumpValues()

	// Output: Dump request result:
	// Dump all record values
	//   AA = > 0 <
	//   PE = [ 0 ]
	//   U8 = > 0 <
	//   P2 = [ 0 ]
	//   I8 = > 0 <
}

func ExampleRecord_setValueWithIndex() {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generatePEMUDefinitionTest()
	record, err := NewRecord(d)
	if err != nil {
		fmt.Println("Result record generation error", err)
		return
	}

	err = record.SetValueWithIndex("PX", []uint32{1, 1}, 122)
	if err == nil {
		fmt.Println("Error setting PX with MU error", err)
		return
	}
	fmt.Println("Correct error:", err)
	err = record.SetValueWithIndex("PX", []uint32{1, 0}, 122)
	if err != nil {
		fmt.Println("Error setting PX error", err)
		return
	}

	fmt.Println("Dump request result:")
	record.DumpValues()

	// Output: Correct error: ADG0000062: Multiple field index on an non-Multiple field
	// Dump request result:
	// Dump all record values
	//   AA = > 0 <
	//   PE = [ 0 ]
	//   U8 = > 0 <
	//   P2 = [ 1 ]
	//    PX[01] = > 122 <
	//    PY[01] = > 0 <
	//   I8 = > 0 <
}

func ExampleRecord_setValue() {
	ferr := initLogWithFile("request_result.log")
	if ferr != nil {
		return
	}

	d := generatePEMUDefinitionTest()

	record, err := NewRecord(d)
	if err != nil {
		fmt.Println("Result record generation error", err)
		return
	}

	for i := uint32(0); i < 3; i++ {
		adatypes.Central.Log.Infof("==> Set period group entry PP of %d", (i + 1))
		err = record.SetValueWithIndex("PP", []uint32{i + 1}, (i + 1))
		if err != nil {
			fmt.Println("Set PP error", err)
			return
		}
		adatypes.Central.Log.Infof("<== Set period group entry PP of %d", (i + 1))
	}
	adatypes.Central.Log.Infof("==> Set period MU")
	err = record.SetValueWithIndex("MU", []uint32{1, 1}, 100)
	if err != nil {
		fmt.Println("Set MU error", err)
		return
	}
	adatypes.Central.Log.Infof("==> Set second period MU")
	err = record.SetValueWithIndex("MU", []uint32{1, 2}, 122)
	if err != nil {
		fmt.Println("Set MU error", err)
		return
	}
	err = record.SetValue("AA", 2)
	if err != nil {
		fmt.Println("Set AA error", err)
		return
	}
	err = record.SetValue("U8", 3)
	if err != nil {
		fmt.Println("Set U8 error", err)
		return
	}
	err = record.SetValue("I8", 1)
	if err != nil {
		fmt.Println("Set I8 error", err)
		return
	}

	fmt.Println("Dump request result:")
	record.DumpValues()

	// Output: Dump request result:
	// Dump all record values
	//   AA = > 2 <
	//   PE = [ 3 ]
	//    PP[01] = > 1 <
	//    MU[01] = [ 1 ]
	//     MU[01,01] = > 100 <
	//     MU[01,02] = > 122 <
	//    GR[01] = [ 1 ]
	//     PA[01] = > 0 <
	//     PG[01] = > 0 <
	//    G8[01] = > 0 <
	//    PP[02] = > 2 <
	//    MU[02] = [ 0 ]
	//    GR[02] = [ 1 ]
	//     PA[02] = > 0 <
	//     PG[02] = > 0 <
	//    G8[02] = > 0 <
	//    PP[03] = > 3 <
	//    MU[03] = [ 0 ]
	//    GR[03] = [ 1 ]
	//     PA[03] = > 0 <
	//     PG[03] = > 0 <
	//    G8[03] = > 0 <
	//   U8 = > 3 <
	//   P2 = [ 0 ]
	//   I8 = > 1 <
}
