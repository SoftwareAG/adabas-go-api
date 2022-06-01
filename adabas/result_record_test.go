/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	cErr := testDefinition.CreateValues(false)
	assert.NoError(t, cErr)

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

func TestRecord_Marshal(t *testing.T) {
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
	err = testDefinition.CreateValues(false)
	assert.NoError(t, err)
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

func TestRecord_MarshalLink(t *testing.T) {
	initTestLogWithFile(t, "Record.log")

	resultNil, err := NewRecord(nil)
	assert.Error(t, err)
	assert.Nil(t, resultNil)

	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt4, "U4", "UInt4"),
		adatypes.NewType(adatypes.FieldTypeLBString, "S4", "@Link", 100),
		adatypes.NewType(adatypes.FieldTypeUInt8, "I8", "Int8"),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	err = testDefinition.CreateValues(false)
	assert.NoError(t, err)
	result, err := NewRecord(testDefinition)
	if assert.NoError(t, err) {
		verr := result.SetValue("U4", 100)
		assert.Error(t, verr)
		verr = result.SetValue("UInt4", 100)
		assert.NoError(t, verr)
		verr = result.SetValue("Int8", "1234567")
		assert.NoError(t, verr)
		verr = result.SetValue("S4", "1234567")
		assert.Error(t, verr)
		verr = result.SetValue("@Link", "4200")
		assert.NoError(t, verr)

		xout, xerr := xml.Marshal(result)
		assert.NoError(t, xerr)
		fmt.Println("XML:", string(xout))
		assert.Equal(t, "<Record><UInt4>100</UInt4><Link type=\"link\">4200</Link><Int8>1234567</Int8></Record>", string(xout))
		jout, jerr := json.Marshal(result)
		assert.NoError(t, jerr)
		fmt.Println("JSON:", string(jout))
		assert.Equal(t, "{\"@Link\":\"4200\",\"Int8\":1234567,\"UInt4\":100}", string(jout))
		result.adabasMap = NewAdabasMap("ABC", &DatabaseURL{})
		xout, xerr = xml.Marshal(result)
		assert.NoError(t, xerr)
		fmt.Println("XML:", string(xout))
		assert.Equal(t, "<ABC><UInt4>100</UInt4><Link type=\"link\">4200</Link><Int8>1234567</Int8></ABC>", string(xout))
		jout, jerr = json.Marshal(result)
		assert.NoError(t, jerr)
		assert.Equal(t, "{\"@Link\":\"4200\",\"Int8\":1234567,\"UInt4\":100}", string(jout))
		fmt.Println("JSON:", string(jout))

	}
}

func TestRecordGroupValues(t *testing.T) {
	initTestLogWithFile(t, "Record.log")

	resultNil, err := NewRecord(nil)
	assert.Error(t, err)
	assert.Nil(t, resultNil)

	groupLayoutLevel3 := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUInt2, "U3"),
		adatypes.NewType(adatypes.FieldTypeInt8, "I3"),
	}

	groupLayoutLevel2 := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUByte, "U2"),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "G3", adatypes.OccNone, groupLayoutLevel3),
		adatypes.NewType(adatypes.FieldTypeUInt2, "I2"),
	}

	periodGroupLayoutLevel2 := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeUByte, "UP"),
		adatypes.NewType(adatypes.FieldTypeUInt2, "IP"),
	}

	layout := []adatypes.IAdaType{
		adatypes.NewType(adatypes.FieldTypeString, "UI", 20),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "G2", adatypes.OccNone, groupLayoutLevel2),
		adatypes.NewStructureList(adatypes.FieldTypePeriodGroup, "P2", adatypes.OccCapacity, periodGroupLayoutLevel2),
		adatypes.NewType(adatypes.FieldTypeByte, "BI"),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	err = testDefinition.CreateValues(false)
	assert.NoError(t, err)
	result, err := NewRecord(testDefinition)
	_ = result.SetValue("UP[1]", 100)
	_ = result.SetValue("IP[1]", 1023)
	_ = result.SetValue("BI", 1)
	_ = result.SetValue("UI", "ANCXXX")
	_ = result.SetValue("I2", 1231)
	if assert.NoError(t, err) {
		assert.NotNil(t, result)
		assert.Equal(t, "ISN=0 quantity=0\n UI=\"ANCXXX              \"\n G2=\"\"\n U2=\"0\"\n G3=\"\"\n U3=\"0\"\n I3=\"0\"\n I2=\"1231\"\n P2=\"\"\n UP=\"100\"\n IP=\"1023\"\n BI=\"1\"\n", result.String())
		v, verr := result.SearchValue("G3")
		assert.NoError(t, verr)
		assert.NotNil(t, v)
		assert.Equal(t, "", v.String())
		v, verr = result.SearchValue("G2")
		assert.NoError(t, verr)
		assert.Equal(t, "G2", v.Type().Name())
		v, verr = result.SearchValue("U3")
		assert.NoError(t, verr)
		assert.Equal(t, "U3", v.Type().Name())
		fieldNames := []string{"UI", "BI", "G2", "U2", "G3", "U3"}
		for _, s := range fieldNames {
			fmt.Println("Check field", s)
			_, ok := result.HashFields[s]
			assert.True(t, ok)
		}
	}
	//result.definition.DumpTypes(false, true)
	//result.DumpValues()
	x, err := xml.Marshal(result)
	assert.NoError(t, err)
	assert.Equal(t, `<Record><UI>ANCXXX</UI><Group sn="G2"><U2>0</U2><Group sn="G3"><U3>0</U3><I3>0</I3></Group><I2>1231</I2></Group><Period sn="P2"><Entry><UP>100</UP><IP>1023</IP></Entry></Period><BI>1</BI></Record>`, string(x))
	j, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.Equal(t, `{"BI":1,"G2":{"G3":{"I3":0,"U3":0},"I2":1231,"U2":0},"P2":[{"IP":1023,"UP":100}],"UI":"ANCXXX"}`, string(j))

}

func TestRecordStoreTest(t *testing.T) {
	definitionEmployees := employeeDefinition()
	fmt.Println(definitionEmployees.String())
	req, err := CreateTestRequest(true, definitionEmployees)
	assert.NoError(t, err)
	err = definitionEmployees.CreateValues(true)
	if !assert.NoError(t, err) {
		return
	}
	record, xerr := NewRecord(definitionEmployees)
	if !assert.NoError(t, xerr) {
		return
	}
	assert.Equal(t, "ISN=0 quantity=0\n AA=\"        \"\n AB=\"\"\n AC=\"                    \"\n AD=\"                    \"\n AE=\"                    \"\n A2=\"\"\n AN=\"      \"\n AM=\"               \"\n AQ=\"\"\n", record.String())
	assert.Equal(t, "AA,8,A,AC,20,A,AD,20,A,AE,20,A,AN,6,A,AM,15,A.", req.FormatBuffer.String())
	definitionEmployees.RestrictFieldSlice([]string{"AS"})
	record, xerr = NewRecord(definitionEmployees)
	if !assert.NoError(t, xerr) {
		return
	}
	err = record.SetValueWithIndex("AS", []uint32{1}, 12345)
	if !assert.NoError(t, err) {
		return
	}
	err = record.SetValueWithIndex("AS", []uint32{2}, 112345)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "ISN=0 quantity=0\n AQ=\"\"\n AS=\"12345\"\n AS=\"112345\"\n", record.String())
	record.DumpValues()
	req, err = CreateTestRequest(true, definitionEmployees)
	assert.NoError(t, err)
	err = definitionEmployees.CreateValues(true)
	if !assert.NoError(t, err) {
		return
	}
	req.Definition.Values = record.Value

	assert.Equal(t, "AS1,5,P,AS2,5,P.", req.FormatBuffer.String())

}

func CreateTestRequest(store bool, testDefinition *adatypes.Definition) (*adatypes.Request, error) {
	if testDefinition == nil {
		return nil, fmt.Errorf("test definition not defined")
	}
	adabasParameter := &adatypes.AdabasRequestParameter{Store: store, DescriptorRead: false,
		SecondCall: 0, Mainframe: false}
	req, err := testDefinition.CreateAdabasRequest(adabasParameter)
	if err != nil {
		fmt.Println("Create request", err)
		return nil, err
	}
	helper := adatypes.NewDynamicHelper(Endian())
	req.RecordBuffer = helper
	req.Parser = testParser
	req.Limit = 1
	req.Multifetch = 1
	req.Isn = 10
	req.Definition = testDefinition
	req.RecordBuffer.PutInt32(2)
	req.RecordBuffer.PutInt32(10)
	return req, nil
}

func employeeDefinition() *adatypes.Definition {
	multipleLayout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypePacked, "AT", 5),
	}
	for _, l := range multipleLayout {
		l.SetLevel(2)
		l.AddFlag(adatypes.FlagOptionMUGhost)
	}
	multipleLayout[0].AddOption(adatypes.FieldOptionMU)
	multipleLayout[0].AddOption(adatypes.FieldOptionNU)
	multipleLayout[0].AddOption(adatypes.FieldOptionNB)
	multipleLayout[0].AddOption(adatypes.FieldOptionNV)
	peGroupLayout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AR", 3),
		adatypes.NewTypeWithLength(adatypes.FieldTypePacked, "AS", 5),
		adatypes.NewStructureList(adatypes.FieldTypeMultiplefield, "AT", adatypes.OccNone, multipleLayout),
	}
	groupLayout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AC", 20),
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AD", 20),
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AE", 20),
	}
	groupLayout2 := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AN", 6),
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AM", 15),
	}
	for _, l := range groupLayout {
		l.SetLevel(2)
		l.AddOption(adatypes.FieldOptionNU)
	}
	for _, l := range groupLayout2 {
		l.SetLevel(2)
		l.AddOption(adatypes.FieldOptionNU)
	}
	// groupLayout[1].AddOption(FieldOptionMU)
	layout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AA", 8),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "AB", adatypes.OccNone, groupLayout),
		adatypes.NewStructureList(adatypes.FieldTypeGroup, "A2", adatypes.OccNone, groupLayout2),
		adatypes.NewStructureList(adatypes.FieldTypePeriodGroup, "AQ", adatypes.OccNone, peGroupLayout),
	}
	for i, l := range layout {
		l.SetLevel(1)
		if i != 0 {
			l.AddOption(adatypes.FieldOptionNU)
		}
	}
	layout[0].AddOption(adatypes.FieldOptionDE)
	layout[0].AddOption(adatypes.FieldOptionUQ)
	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}
