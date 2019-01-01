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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type storeRecordTraverserStructure struct {
	record *ResultRecord
	helper *adatypes.BufferHelper
}

func createStoreRecordBuffer(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	record := x.(*storeRecordTraverserStructure)
	adatypes.Central.Log.Debugf("Store record buffer for %s current helper position is %d/%x",
		adaValue.Type().Name(), record.helper.Offset(), record.helper.Offset())
	return adatypes.Continue, adaValue.StoreBuffer(record.helper)
}

func (record *ResultRecord) createRecordBuffer(helper *adatypes.BufferHelper) (err error) {
	adatypes.Central.Log.Debugf("Create record buffer")
	t := adatypes.TraverserValuesMethods{EnterFunction: createStoreRecordBuffer}
	stRecTraverser := &storeRecordTraverserStructure{record: record, helper: helper}
	_, err = record.traverse(t, stRecTraverser)
	adatypes.Central.Log.Debugf("Create record buffer done len=%d", len(helper.Buffer()))
	return
}

// RequestResult contains the result information of the request
type RequestResult struct {
	XMLName xml.Name        `xml:"Response" json:"-"`
	Values  []*ResultRecord `xml:"Record" json:"Record"`
}

// NrRecords number of records in the result
func (requestResult *RequestResult) NrRecords() int {
	return len(requestResult.Values)
}

func dumpValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue == nil {
		record := x.(*ResultRecord)
		if record == nil {
			return adatypes.EndTraverser, adatypes.NewGenericError(25)
		}
		if record.Isn > 0 {
			fmt.Printf("Record Isn: %04d\n", record.Isn)
		}
		if record.quantity > 0 {
			fmt.Printf("Record Quantity: %04d\n", record.quantity)
		}
	} else {

		y := strings.Repeat(" ", int(adaValue.Type().Level()))

		if x == nil {
			brackets := ""
			switch {
			case adaValue.PeriodIndex() > 0 && adaValue.MultipleIndex() > 0:
				brackets = fmt.Sprintf("[%02d,%02d]", adaValue.PeriodIndex(), adaValue.MultipleIndex())
			case adaValue.PeriodIndex() > 0:
				brackets = fmt.Sprintf("[%02d]", adaValue.PeriodIndex())
			case adaValue.MultipleIndex() > 0:
				brackets = fmt.Sprintf("[%02d]", adaValue.MultipleIndex())
			default:
			}

			if adaValue.Type().IsStructure() {
				structureValue := adaValue.(*adatypes.StructureValue)
				fmt.Println(y+" "+adaValue.Type().Name()+brackets+" = [", structureValue.NrElements(), "]")
			} else {
				fmt.Printf("%s %s%s = > %s <\n", y, adaValue.Type().Name(), brackets, adaValue.String())
			}
		} else {
			buffer := x.(*bytes.Buffer)
			buffer.WriteString(fmt.Sprintln(y, adaValue.Type().Name(), "= >", adaValue.String(), "<"))
		}
	}
	return adatypes.Continue, nil
}

// DumpValues traverse through the tree of values calling a callback method
func (requestResult *RequestResult) DumpValues() (err error) {
	fmt.Println("Dump all result values")
	t := adatypes.TraverserValuesMethods{EnterFunction: dumpValues}
	_, err = requestResult.TraverseValues(t, nil)
	return
}

// TraverseValues traverse through the tree of values calling a callback method
func (requestResult *RequestResult) TraverseValues(t adatypes.TraverserValuesMethods, x interface{}) (ret adatypes.TraverseResult, err error) {
	adatypes.Central.Log.Debugf("Traverse result values")
	if requestResult.Values == nil {
		err = errors.New("no values")
		return
	}
	adatypes.Central.Log.Debugf("Go through records -> ", len(requestResult.Values))
	var tr adatypes.TraverseResult
	for _, record := range requestResult.Values {
		if t.EnterFunction != nil {
			tr, err = t.EnterFunction(nil, record)
			if err != nil || tr == adatypes.SkipStructure {
				return
			}
		}
		ret, err = record.traverse(t, x)
		if err != nil {
			return
		}
	}

	adatypes.Central.Log.Debugf("Ready traverse values")
	return
}

func (requestResult *RequestResult) String() string {
	var buffer bytes.Buffer
	t := adatypes.TraverserValuesMethods{EnterFunction: dumpValues}
	requestResult.TraverseValues(t, &buffer)
	return buffer.String()
}

func traverseMarshalXML(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	enc := x.(*xml.Encoder)
	start := xml.StartElement{Name: xml.Name{Local: adaValue.Type().Name()}}
	enc.EncodeToken(start)
	if !adaValue.Type().IsStructure() {
		enc.EncodeToken(xml.CharData([]byte(adaValue.String())))
		enc.EncodeToken(start.End())
	}
	return adatypes.Continue, nil
}

func traverseMarshalXMLEnd(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue.Type().IsStructure() {
		enc := x.(*xml.Encoder)
		end := xml.EndElement{Name: xml.Name{Local: adaValue.Type().Name()}}
		enc.EncodeToken(end)
	}
	return adatypes.Continue, nil
}

// MarshalXML provide XML
func (requestResult *RequestResult) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	x := xml.StartElement{Name: xml.Name{Local: "Response"}}
	e.EncodeToken(x)
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalXML, LeaveFunction: traverseMarshalXMLEnd}
	adatypes.Central.Log.Debugf("Go through records -> ", len(requestResult.Values))
	if requestResult.Values != nil {
		for _, record := range requestResult.Values {
			rec := xml.StartElement{Name: xml.Name{Local: "Record"}}
			rec.Attr = []xml.Attr{xml.Attr{Name: xml.Name{Local: "ISN"}, Value: strconv.Itoa(int(record.Isn))}}
			e.EncodeToken(rec)
			// e.EncodeToken(xml.Attr{Name: xml.Name{Local: "ISN"}, Value: strconv.Itoa(int(record.Isn))})
			record.traverse(tm, e)
			e.EncodeToken(rec.End())
		}
	}

	//	requestResult.TraverseValues(tm, e)
	// e.EncodeToken(xml.CharData([]byte("abc")))
	e.EncodeToken(x.End())
	return nil
}

func traverseMarshalXML2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	enc := x.(*xml.Encoder)
	start := xml.StartElement{Name: xml.Name{Local: adaValue.Type().Name()}}
	enc.EncodeToken(start)
	if !adaValue.Type().IsStructure() {
		enc.EncodeToken(xml.CharData([]byte(adaValue.String())))
		enc.EncodeToken(start.End())
	}
	return adatypes.Continue, nil
}

func traverseMarshalXMLEnd2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue.Type().IsStructure() {
		enc := x.(*xml.Encoder)
		end := xml.EndElement{Name: xml.Name{Local: adaValue.Type().Name()}}
		enc.EncodeToken(end)
	}
	return adatypes.Continue, nil
}

// MarshalXML provide XML
func (record *ResultRecord) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	x := xml.StartElement{Name: xml.Name{Local: "Response"}}
	e.EncodeToken(x)
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalXML2, LeaveFunction: traverseMarshalXMLEnd2}
	rec := xml.StartElement{Name: xml.Name{Local: "Record"}}
	rec.Attr = []xml.Attr{xml.Attr{Name: xml.Name{Local: "ISN"}, Value: strconv.Itoa(int(record.Isn))}}
	e.EncodeToken(rec)
	record.traverse(tm, e)
	e.EncodeToken(rec.End())

	e.EncodeToken(x.End())
	return nil
}

type dataValue struct {
	Isn   adatypes.Isn `json:"ISN"`
	Value map[string]string
}

type request struct {
	Values  []*map[string]interface{} `json:"Record"`
	dataMap *map[string]interface{}
	stack   *adatypes.Stack
	buffer  bytes.Buffer
}

func traverseMarshalJSON(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	req := x.(*request)
	if adaValue.Type().IsStructure() {
		req.stack.Push(req.dataMap)
		dataMap := make(map[string]interface{})
		oldMap := req.dataMap
		req.dataMap = &dataMap
		(*oldMap)[adaValue.Type().Name()] = req.dataMap
	} else {
		switch adaValue.Type().Type() {
		case adatypes.FieldTypePacked, adatypes.FieldTypeByte:
			v, err := adaValue.Int64()
			if err != nil {
				fmt.Println("Error ", adaValue.Type().Name(), " -> ", err)
				return adatypes.EndTraverser, err
			}
			(*req.dataMap)[adaValue.Type().Name()] = v
		default:
			(*req.dataMap)[adaValue.Type().Name()] = adaValue.String()
		}
	}
	return adatypes.Continue, nil
}

func traverseMarshalJSONEnd(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue.Type().IsStructure() {
		req := x.(*request)
		dataMap, err := req.stack.Pop()
		if err != nil {
			return adatypes.EndTraverser, err
		}
		req.dataMap = dataMap.((*map[string]interface{}))
	}
	return adatypes.Continue, nil
}

// MarshalJSON provide JSON
func (requestResult *RequestResult) MarshalJSON() ([]byte, error) {
	req := &request{}
	adatypes.Central.Log.Debugf("Go through records -> ", len(requestResult.Values))
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalJSON, LeaveFunction: traverseMarshalJSONEnd}
	req.stack = adatypes.NewStack()

	for _, record := range requestResult.Values {
		dataMap := make(map[string]interface{})
		req.dataMap = &dataMap
		req.Values = append(req.Values, req.dataMap)
		dataMap["ISN"] = record.Isn
		record.traverse(tm, req)
	}
	return json.Marshal(req)
}

// Isn Search for record with given ISN
func (requestResult *RequestResult) Isn(isn adatypes.Isn) *ResultRecord {
	for _, record := range requestResult.Values {
		if record.Isn == isn {
			return record
		}
	}
	return nil
}

type rrecord struct {
	stack       *adatypes.Stack
	buffer      bytes.Buffer
	hasElements bool
}

func traverseMarshalJSON2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	req := x.(*rrecord)
	if req.hasElements {
		req.buffer.WriteByte(',')
	}
	if adaValue.Type().IsStructure() {
		req.buffer.WriteString("\"" + adaValue.Type().Name() + "\":{")
		req.stack.Push(true)
		req.hasElements = false
	} else {
		switch adaValue.Type().Type() {
		case adatypes.FieldTypePacked, adatypes.FieldTypeByte:
			v, err := adaValue.Int64()
			if err != nil {
				fmt.Println("Error ", adaValue.Type().Name(), " -> ", err)
				return adatypes.EndTraverser, err
			}
			req.buffer.WriteString(fmt.Sprintf("\"%s\":%v", adaValue.Type().Name(), v))
		default:
			req.buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\"", adaValue.Type().Name(), adaValue.String()))
		}
		req.hasElements = true
	}
	return adatypes.Continue, nil
}

func traverseMarshalJSONEnd2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue.Type().IsStructure() {
		req := x.(*rrecord)
		req.buffer.WriteString("}")
	}
	return adatypes.Continue, nil
}

// MarshalJSON provide JSON
func (record *ResultRecord) MarshalJSON() ([]byte, error) {
	rec := &rrecord{hasElements: false}
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalJSON2, LeaveFunction: traverseMarshalJSONEnd2}
	rec.stack = adatypes.NewStack()

	rec.buffer.WriteByte('{')
	rec.buffer.WriteString(fmt.Sprintf("\"%s\":%v", "ISN", record.Isn))
	rec.hasElements = true
	/* ret, err := */ record.traverse(tm, rec)
	rec.buffer.WriteByte('}')

	return rec.buffer.Bytes(), nil
}

// UnmarshalJSON parse JSON
// func (record *ResultRecord) UnmarshalJSON(b []byte) error {
// 	var stuff map[string]interface{}
// 	err := json.Unmarshal(b, &stuff)
// 	if err != nil {
// 		return err
// 	}
// 	if record.Value == nil {
// 		if record.definition.Values == nil {
// 			record.definition.CreateValues(false)
// 		}
// 		record.Value = record.definition.Values
// 	}
// 	for key, value := range stuff {
// 		fmt.Println("JSON:", key, "=", value)
// 		if key == "ISN" {
// 			isn, ierr := strconv.Atoi(value.(string))
// 			if ierr != nil {
// 				return ierr
// 			}
// 			record.Isn = adatypes.Isn(isn)
// 		} else {
// 			switch value.(type) {
// 			case map[string]interface{}:
// 				fmt.Println("JSON:", key, "=", value)
// 			default:
// 				err = record.SetValue(key, value)
// 				if err != nil {
// 					fmt.Println("Error setting key:", key)
// 					return err
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
