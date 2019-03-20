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

// Record one result record of the result
type Record struct {
	Isn        adatypes.Isn `xml:"Isn,attr"`
	Quantity   uint64       `xml:"Quantity,attr"`
	Value      []adatypes.IAdaValue
	HashFields map[string]adatypes.IAdaValue `xml:"-" json:"-"`
	fields     map[string]*queryField
	definition *adatypes.Definition
}

func hashValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	record := x.(*Record)
	if _, ok := record.HashFields[adaValue.Type().Name()]; !ok {
		record.HashFields[adaValue.Type().Name()] = adaValue
	}

	return adatypes.Continue, nil
}

// NewRecord new result record
func NewRecord(definition *adatypes.Definition) (*Record, error) {
	if definition == nil {
		adatypes.Central.Log.Debugf("Definition values empty")
		return nil, adatypes.NewGenericError(69)
	}
	if definition.Values == nil {
		err := definition.CreateValues(false)
		if err != nil {
			return nil, err
		}
	}
	record := &Record{Value: definition.Values, definition: definition}
	definition.Values = nil
	record.HashFields = make(map[string]adatypes.IAdaValue)
	t := adatypes.TraverserValuesMethods{EnterFunction: hashValues}
	record.traverse(t, record)
	return record, nil
}

// NewRecordIsn new result record with ISN or ISN quantity
func NewRecordIsn(isn adatypes.Isn, isnQuantity uint64, definition *adatypes.Definition) (*Record, error) {
	record, err := NewRecord(definition)
	if err != nil {
		return nil, err
	}
	record.Isn = isn
	record.Quantity = isnQuantity
	adatypes.Central.Log.Debugf("New record with ISN=%d and ISN quantity=%d", isn, isnQuantity)

	return record, nil
}

func recordValuesTraverser(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	buffer := x.(*bytes.Buffer)
	buffer.WriteString(fmt.Sprintf(" %s=%#v\n", adaValue.Type().Name(), adaValue.String()))
	return adatypes.Continue, nil
}

func (record *Record) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ISN=%d quantity=%d\n", record.Isn, record.Quantity))
	t := adatypes.TraverserValuesMethods{EnterFunction: recordValuesTraverser}
	record.traverse(t, &buffer)
	// for _, v := range record.Value {
	// 	buffer.WriteString(fmt.Sprintf("value=%#v\n", v))
	// }
	return buffer.String()
}

func (record *Record) traverse(t adatypes.TraverserValuesMethods, x interface{}) (ret adatypes.TraverseResult, err error) {
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
func (record *Record) DumpValues() {
	fmt.Println("Dump all record values")
	t := adatypes.TraverserValuesMethods{PrepareFunction: prepareRecordDump,
		EnterFunction: dumpRecord}
	record.traverse(t, nil)
}

func (record *Record) searchValue(field string) (adatypes.IAdaValue, bool) {
	if adaValue, ok := record.HashFields[field]; ok {
		return adaValue, true
	}
	return nil, false
}

// SetValue set the value for a specific field
func (record *Record) SetValue(field string, value interface{}) (err error) {
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
		adatypes.Central.Log.Debugf("Field %s not found %v err=%v", field, adaValue, err)
	}
	return
}

// SetValueWithIndex Add value to an node element
func (record *Record) SetValueWithIndex(name string, index []uint32, x interface{}) error {
	// TODO why specific?
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SetValueWithIndex(name, index, x)
}

// SearchValue search value in the tree
func (record *Record) SearchValue(name string) (adatypes.IAdaValue, error) {
	return record.SearchValueIndex(name, []uint32{0, 0})
}

// SearchValueIndex search value in the tree with a given index
func (record *Record) SearchValueIndex(name string, index []uint32) (adatypes.IAdaValue, error) {
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SearchByIndex(name, index, false)
}

// Scan scan for different field entries
func (record *Record) Scan(dest ...interface{}) (err error) {
	adatypes.Central.Log.Debugf("Scan Record %#v", record.fields)
	if f, ok := record.fields["#ISN"]; ok {
		adatypes.Central.Log.Debugf("Fill Record ISN=%d", record.Isn)
		*(dest[f.index].(*int)) = int(record.Isn)
	}
	if f, ok := record.fields["#ISNQUANTITY"]; ok {
		adatypes.Central.Log.Debugf("Fill Record ISN quantity=%d", record.Quantity)
		*(dest[f.index].(*int)) = int(record.Quantity)
	}
	// Traverse to current entries
	tm := adatypes.TraverserValuesMethods{EnterFunction: scanFieldsTraverser}
	sf := &scanFields{fields: record.fields, parameter: dest}
	_, err = record.traverse(tm, sf)
	if err != nil {
		return err
	}
	return nil

}
