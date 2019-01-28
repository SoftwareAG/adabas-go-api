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
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// ResultRecord one result record of the result
type ResultRecord struct {
	Isn        adatypes.Isn `xml:"Isn,attr"`
	quantity   uint64
	Value      []adatypes.IAdaValue
	HashFields map[string]adatypes.IAdaValue `xml:"-" json:"-"`
	definition *adatypes.Definition          `xml:"-" json:"-"`
}

func hashValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	record := x.(*ResultRecord)
	if _, ok := record.HashFields[adaValue.Type().Name()]; !ok {
		record.HashFields[adaValue.Type().Name()] = adaValue
	}

	return adatypes.Continue, nil
}

// NewResultRecord new result record
func NewResultRecord(definition *adatypes.Definition) (*ResultRecord, error) {
	if definition == nil {
		adatypes.Central.Log.Debugf("Definition values empty")
		return nil, fmt.Errorf("Field list empty")
	}
	if definition.Values == nil {
		err := definition.CreateValues(false)
		if err != nil {
			return nil, err
		}
	}
	record := &ResultRecord{Value: definition.Values, definition: definition}
	definition.Values = nil
	record.HashFields = make(map[string]adatypes.IAdaValue)
	t := adatypes.TraverserValuesMethods{EnterFunction: hashValues}
	record.traverse(t, record)
	return record, nil
}

// NewResultRecordIsn new result record with ISN or ISN quantity
func NewResultRecordIsn(isn adatypes.Isn, isnQuantity uint64, definition *adatypes.Definition) (*ResultRecord, error) {
	record, err := NewResultRecord(definition)
	if err != nil {
		return nil, err
	}
	record.Isn = isn
	record.quantity = isnQuantity
	return record, nil
}

func recordValuesTraverser(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	buffer := x.(*bytes.Buffer)
	buffer.WriteString(fmt.Sprintf(" %s=%#v\n", adaValue.Type().Name(), adaValue.String()))
	return adatypes.Continue, nil
}

func (record *ResultRecord) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ISN=%d quanity=%d\n", record.Isn, record.quantity))
	t := adatypes.TraverserValuesMethods{EnterFunction: recordValuesTraverser}
	record.traverse(t, &buffer)
	// for _, v := range record.Value {
	// 	buffer.WriteString(fmt.Sprintf("value=%#v\n", v))
	// }
	return buffer.String()
}

func (record *ResultRecord) traverse(t adatypes.TraverserValuesMethods, x interface{}) (ret adatypes.TraverseResult, err error) {
	if record == nil {
		return adatypes.EndTraverser, adatypes.NewGenericError(33)
	}

	for _, value := range record.Value {
		adatypes.Central.Log.Debugf("Go through value %s %d", value.Type().Name(), value.Type().Type())
		if t.EnterFunction != nil {
			adatypes.Central.Log.Debugf("Enter field=%s Type=%d", value.Type().Name(), value.Type().Type())
			ret, err = t.EnterFunction(value, x)
			if err != nil {
				return
			}
		}
		if value.Type().IsStructure() {
			adatypes.Central.Log.Debugf("Go through structure %s %d", value.Type().Name(), value.Type().Type())
			ret, err = value.(*adatypes.StructureValue).Traverse(t, x)
			if err != nil || ret == adatypes.EndTraverser {
				return
			}
		}
		if t.LeaveFunction != nil {
			adatypes.Central.Log.Debugf("Leave %s %d", value.Type().Name(), value.Type().Type())
			ret, err = t.LeaveFunction(value, x)
			if err != nil || ret == adatypes.EndTraverser {
				return
			}
		}
	}
	adatypes.Central.Log.Debugf("Traverse ended")
	return
}

// DumpValues traverse through the tree of values calling a callback method
func (record *ResultRecord) DumpValues() {
	fmt.Println("Dump all result values")
	t := adatypes.TraverserValuesMethods{PrepareFunction: prepareResultRecordDump,
		EnterFunction: dumpResultRecord}
	record.traverse(t, nil)
}

func (record *ResultRecord) searchValue(field string) (adatypes.IAdaValue, bool) {
	if adaValue, ok := record.HashFields[field]; ok {
		return adaValue, true
	}
	return nil, false
}

// SetValue set the value for a specific field
func (record *ResultRecord) SetValue(field string, value interface{}) (err error) {
	if strings.ContainsRune(field, '[') {
		i := strings.IndexRune(field, '[')
		e := strings.IndexRune(field, ']')
		index, xerr := strconv.Atoi(field[i+1 : e])
		if xerr != nil {
			return xerr
		}
		eField := field[:i]
		f := field[e+1:]
		i = strings.IndexRune(f, '[')
		if i == -1 {
			return record.SetValueWithIndex(eField, []uint32{uint32(index)}, value)
		}
		e = strings.IndexRune(f, ']')
		muindex, merr := strconv.Atoi(f[i+1 : e])
		if merr != nil {
			return merr
		}

		return record.SetValueWithIndex(eField, []uint32{uint32(index), uint32(muindex)}, value)
	}
	if adaValue, ok := record.searchValue(field); ok {
		err = adaValue.SetValue(value)
		adatypes.Central.Log.Debugf("Set %s [%T] value err=%v", field, adaValue, err)
	} else {
		err = adatypes.NewGenericError(28, field)
		adatypes.Central.Log.Debugf("Field %s not found err=%v", field, adaValue, err)
	}
	return
}

// SetValueWithIndex Add value to an node element
func (record *ResultRecord) SetValueWithIndex(name string, index []uint32, x interface{}) error {
	// TODO why specific?
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SetValueWithIndex(name, index, x)
}

// SearchValue search value in the tree
func (record *ResultRecord) SearchValue(name string) (adatypes.IAdaValue, error) {
	return record.SearchValueIndex(name, []uint32{0, 0})
}

// SearchValueIndex search value in the tree with a given index
func (record *ResultRecord) SearchValueIndex(name string, index []uint32) (adatypes.IAdaValue, error) {
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SearchByIndex(name, index, false)
}
