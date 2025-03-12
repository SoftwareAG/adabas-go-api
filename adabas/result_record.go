/*
* Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// Record one result record of the result list received by
// record list or in the stream callback.
//
// To extract the values in the record you might request the
// value using the SearchValue() methods. Alternatively you
// might use the Traverse() callback method to call a method
// for each Adabas field in the tree. The tree includes group
// nodes of the Adabas record.
type Record struct {
	Isn               adatypes.Isn `xml:"Isn,attr"`
	Quantity          uint64       `xml:"Quantity,attr"`
	Value             []adatypes.IAdaValue
	HashFields        map[string]adatypes.IAdaValue `xml:"-" json:"-"`
	fields            map[string]*queryField
	definition        *adatypes.Definition
	adabasMap         *Map
	LobEndTransaction bool
}

func traverseHashValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	record := x.(*Record)
	//if _, ok := record.HashFields[adaValue.Type().Name()]; !ok {
	if !adaValue.Type().HasFlagSet(adatypes.FlagOptionMUGhost) {
		adatypes.Central.Log.Debugf("Add hash to %s", adaValue.Type().Name())
		record.HashFields[adaValue.Type().Name()] = adaValue
	}
	//}

	return adatypes.Continue, nil
}

// NewRecord create new result record infrastructure based on the given definition
func NewRecord(definition *adatypes.Definition) (*Record, error) {
	adatypes.Central.Log.Debugf("Create new record")
	if definition == nil {
		adatypes.Central.Log.Debugf("Definition empty")
		return nil, adatypes.NewGenericError(69)
	}
	if definition.Values == nil {
		adatypes.Central.Log.Debugf("Definition values empty")
		err := definition.CreateValues(false)
		if err != nil {
			adatypes.Central.Log.Debugf("Error creating Definition values")
			return nil, err
		}
	}
	record := &Record{Value: definition.Values, definition: definition, LobEndTransaction: false}
	definition.Values = nil
	record.HashFields = make(map[string]adatypes.IAdaValue)
	t := adatypes.TraverserValuesMethods{EnterFunction: traverseHashValues}
	_, err := record.Traverse(t, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// NewRecordIsn create a new result record with ISN or ISN quantity
func NewRecordIsn(isn adatypes.Isn, isnQuantity uint64, definition *adatypes.Definition) (*Record, error) {
	record, err := NewRecord(definition)
	if err != nil {
		return nil, err
	}
	record.Isn = isn
	record.Quantity = isnQuantity
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("New record with ISN=%d and ISN quantity=%d", isn, isnQuantity)
	}

	return record, nil
}

// recordValuesTraverser create buffer used to output values
func recordValuesTraverser(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	buffer := x.(*bytes.Buffer)
	buffer.WriteString(fmt.Sprintf(" %s=%#v\n", adaValue.Type().Name(), adaValue.String()))
	return adatypes.Continue, nil
}

// createRecordBuffer create a record buffer
func (record *Record) createRecordBuffer(helper *adatypes.BufferHelper, option *adatypes.BufferOption) (err error) {
	adatypes.Central.Log.Debugf("Create store record buffer")
	t := adatypes.TraverserValuesMethods{EnterFunction: createStoreRecordBufferTraverser}
	stRecTraverser := &storeRecordTraverserStructure{record: record, helper: helper, option: option}
	_, err = record.Traverse(t, stRecTraverser)
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Create record buffer done len=%d", len(helper.Buffer()))
	}
	return
}

// String string representation of the record
func (record *Record) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ISN=%d quantity=%d\n", record.Isn, record.Quantity))
	t := adatypes.TraverserValuesMethods{EnterFunction: recordValuesTraverser}
	_, _ = record.Traverse(t, &buffer)
	return buffer.String()
}

// Traverse step/traverse through all record entries and call methods
func (record *Record) Traverse(t adatypes.TraverserValuesMethods, x interface{}) (ret adatypes.TraverseResult, err error) {
	if record == nil {
		return adatypes.EndTraverser, adatypes.NewGenericError(33)
	}
	debug := adatypes.Central.IsDebugLevel()
	for _, value := range record.Value {
		if debug {
			adatypes.Central.Log.Debugf("Go through value %s %d", value.Type().Name(), value.Type().Type())
		}
		if t.EnterFunction != nil {
			if debug {
				adatypes.Central.Log.Debugf("Enter field=%s Type=%d", value.Type().Name(), value.Type().Type())
			}
			ret, err = t.EnterFunction(value, x)
			if err != nil {
				return
			}
			if ret == adatypes.SkipStructure {
				continue
			}
		}
		if value.Type().IsStructure() {
			if debug {
				adatypes.Central.Log.Debugf("Go through structure %s %d", value.Type().Name(), value.Type().Type())
			}
			ret, err = value.(adatypes.StructureValueTraverser).Traverse(t, x)
			if err != nil || ret == adatypes.EndTraverser {
				return
			}
		}
		if t.LeaveFunction != nil {
			if debug {
				adatypes.Central.Log.Debugf("Leave %s %d", value.Type().Name(), value.Type().Type())
			}
			ret, err = t.LeaveFunction(value, x)
			if err != nil || ret == adatypes.EndTraverser {
				return
			}
		}
	}
	if debug {
		adatypes.Central.Log.Debugf("Traverse ended")
	}
	return
}

// DumpValues traverse through the tree of values calling a callback method
func (record *Record) DumpValues() {
	fmt.Println("Dump all record values")
	var buffer bytes.Buffer
	t := adatypes.TraverserValuesMethods{PrepareFunction: prepareRecordDump,
		EnterFunction: traverseDumpRecord}
	_, _ = record.Traverse(t, &buffer)
	fmt.Printf("%s", buffer.String())
}

// searchValue search a value using a given field name
func (record *Record) searchValue(field string) (adatypes.IAdaValue, bool) {
	fq, err := NewFieldQuery(field)
	if err != nil {
		adatypes.Central.Log.Debugf("Query failure: %v", err)
		return nil, false
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Search field %s (%s)", fq.Name, field)
		for k, v := range record.HashFields {
			adatypes.Central.Log.Debugf("Key: %s %#v", k, v)
		}
	}
	name := fq.Name
	if fq.Prefix != ' ' {
		name = string(fq.Prefix) + fq.Name
	}
	if adaValue, ok := record.HashFields[name]; ok {
		adatypes.Central.Log.Debugf("Found value %s (%v)", adaValue.Type().Name(), adaValue.Type().Type())
		if adaValue.Type().Type() == adatypes.FieldTypeMultiplefield && (fq.MultipleIndex > 0 || fq.PeriodicIndex > 0) {
			sv := adaValue.(*adatypes.StructureValue)
			for _, ve := range sv.Elements {
				for _, v := range ve.Values {
					adatypes.Central.Log.Debugf("PE %d MU %d", v.PeriodIndex(), v.MultipleIndex())
					if fq.PeriodicIndex == v.PeriodIndex() && fq.MultipleIndex == v.MultipleIndex() {
						return v, true
					}
				}
			}
		}
		return adaValue, true
	}
	adatypes.Central.Log.Debugf("Found no value for %s", name)
	if adatypes.Central.IsDebugLevel() {
		for k, v := range record.HashFields {
			adatypes.Central.Log.Debugf("Key valid %s -> %s(%d)", k, v.Type().Name(), v.Type().Type())
		}
	}
	return nil, false
}

// SetValue set the value for a specific field given by the field name. The
// field value is defined by the interface given.
// The field value index of a period group or multiple field might be defined
// using square brackets. For example AA[1,2] will set the first entry of a
// period group and the second entry of the multiple field.
func (record *Record) SetValue(field string, value interface{}) (err error) {
	if field == "" {
		err = adatypes.NewGenericError(172)
		return
	}

	adatypes.Central.Log.Debugf("Set value %s", field)
	if strings.ContainsRune(field, '[') {
		i := strings.IndexRune(field, '[')
		c := strings.IndexRune(field[i:], ',')
		e := strings.IndexRune(field, ']')
		if c > 0 {
			eField := field[:i]
			index1, xerr := strconv.ParseInt(field[i+1:i+c], 0, 64)
			if xerr != nil {
				return xerr
			}
			index2, xerr := strconv.ParseInt(field[i+c+1:e], 0, 64)
			if xerr != nil {
				return xerr
			}
			if index1 < 0 || index1 > math.MaxUint32 {
				return adatypes.NewGenericError(118, index1)
			}
			if index2 < 0 || index2 > math.MaxUint32 {
				return adatypes.NewGenericError(118, index2)
			}
			return record.SetValueWithIndex(eField, []uint32{uint32(index1), uint32(index2)}, value)
		}

		index, xerr := strconv.ParseInt(field[i+1:e], 0, 64)
		if xerr != nil {
			return xerr
		}
		eField := field[:i]
		f := field[e+1:]
		i = strings.IndexRune(f, '[')
		if i == -1 {
			if index < 0 || index > math.MaxUint32 {
				return adatypes.NewGenericError(118, index)
			}
			return record.SetValueWithIndex(eField, []uint32{uint32(index)}, value)
		}
		e = strings.IndexRune(f, ']')
		muindex, merr := strconv.ParseInt(f[i+1:e], 0, 64)
		if merr != nil {
			return merr
		}
		if index < 0 || index > math.MaxUint32 {
			return adatypes.NewGenericError(118, index)
		}
		if muindex < 0 || muindex > math.MaxUint32 {
			return adatypes.NewGenericError(118, muindex)
		}

		return record.SetValueWithIndex(eField, []uint32{uint32(index), uint32(muindex)}, value)
	}
	if adaValue, ok := record.searchValue(field); ok {
		err = adaValue.SetValue(value)
		adatypes.Central.Log.Debugf("Set %s [%T] value err=%v", field, adaValue, err)
		// TODO check if the field which is not found and stored should be checked
	} else {
		adatypes.Central.Log.Debugf("Field %s not found %p", field, adaValue)
		s, e2 := record.definition.SearchType(field)
		adatypes.Central.Log.Debugf("Type %v not found %v", s, e2)
		if e2 == nil {
			if s.IsSpecialDescriptor() {
				err = nil
				adatypes.Central.Log.Debugf("Found %v is super descriptor and is ignored", field)
				return
			}
			adatypes.Central.Log.Debugf("Found %v but no super descriptor", field)
		}
		err = adatypes.NewGenericError(28, field)
	}
	return
}

// SetValueWithIndex set the value for a specific field given by the field name. The
// field value is defined by the interface given.
// The field value index of a period group or multiple field might be defined
// using the uint32 slice.
func (record *Record) SetValueWithIndex(name string, index []uint32, x interface{}) error {
	// TODO why specific?
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SetValueWithIndex(name, index, x)
}

// SetPartialValue set the field value for a partial part of a lob field
func (record *Record) SetPartialValue(name string, offset uint32, data []byte) (err error) {
	v, verr := record.SearchValue(name)
	if verr != nil {
		return verr
	}
	if v.Type().Type() != adatypes.FieldTypeLBString {
		return adatypes.NewGenericError(134, v.Type().Name())
	}
	av := v.(adatypes.PartialValue)
	_ = v.SetValue(data)
	// if err != nil {
	// 	return
	// }
	av.SetPartial(offset, uint32(len(data)))

	return nil
}

// extractIndex extract the index information RA[1] of the field
func extractIndex(name string) []uint32 {
	var index []uint32
	var re = regexp.MustCompile(`(?m)(\w+(\[(\d+),?(\d+)?\])?)`)
	for _, s := range re.FindAllStringSubmatch(name, -1) {
		v, err := strconv.ParseInt(s[3], 10, 0)
		if err != nil {
			return index
		}
		if v < 0 || v > math.MaxUint32 {
			return index
		}
		index = append(index, uint32(v))
		if s[4] != "" {
			v, err = strconv.ParseInt(s[4], 10, 0)
			if err != nil {
				return index
			}
			if v < 0 || v > math.MaxUint32 {
				return index
			}
			index = append(index, uint32(v))
		}
	}
	return index
}

// SearchValue this method search for the value in the tree given by the
// field parameter.
func (record *Record) SearchValue(parameter ...interface{}) (adatypes.IAdaValue, error) {
	name := parameter[0].(string)
	var index []uint32
	if len(parameter) > 1 {
		index = parameter[1].([]uint32)
	} else {
		if strings.ContainsRune(name, '[') {
			index = extractIndex(name)
			name = name[:strings.IndexRune(name, '[')]
		} else {
			index = []uint32{0, 0}
		}
	}

	if v, ok := record.HashFields[name]; ok && !v.Type().HasFlagSet(adatypes.FlagOptionPE) {
		return v, nil
	}
	adatypes.Central.Log.Debugf("Search %s index: %#v", name, index)
	return record.SearchValueIndex(name, index)
}

// SearchValueIndex search value in the tree with a given index uint32 slice
func (record *Record) SearchValueIndex(name string, index []uint32) (adatypes.IAdaValue, error) {
	record.definition.Values = record.Value
	adatypes.Central.Log.Debugf("Record value : %#v", record.Value)
	return record.definition.SearchByIndex(name, index, false)
}

// PeriodGroup return if it is part of an  period group return the period
// group field of level 1. If no period group field, returns nil
func PeriodGroup(v adatypes.IAdaValue) adatypes.IAdaValue {
	if v.Type().HasFlagSet(adatypes.FlagOptionPE) {
		c := v
		for c.Type().Type() != adatypes.FieldTypePeriodGroup {
			c = c.Parent()
		}
		return c
	}
	return nil
}

// ValueQuantity this method provide the number of quantity of an PE
// or MU field. If the field is referenced with a square bracket index,
// the corresponding MU field of an period group is counted.
// The quantity is not the Adabas record quantity. It represents the
// record result quantity only.
func (record *Record) ValueQuantity(param ...interface{}) int32 {
	if len(param) == 0 {
		return -1
	}

	var index []uint32
	fieldName := param[0].(string)
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Field name: %s", fieldName)
	}

	if strings.ContainsRune(fieldName, '[') {
		index = extractIndex(fieldName)
		fieldName = fieldName[:strings.IndexRune(fieldName, '[')]
	} else {
		for i := 1; i < len(param); i++ {
			switch w := param[i].(type) {
			case uint32:
				index = append(index, w)
			case int:
				index = append(index, uint32(w))
			default:
			}
		}
	}
	if debug {
		adatypes.Central.Log.Debugf("Index from parser %#v", index)
	}
	if v, ok := record.HashFields[fieldName]; ok {
		if v.Type().HasFlagSet(adatypes.FlagOptionPE) {
			if debug {
				adatypes.Central.Log.Debugf("Quantity of %s PE", v.Type().Name())
			}
			if len(index) < 1 {
				p := PeriodGroup(v)
				pv := p.(*adatypes.StructureValue)
				return int32(pv.NrElements())
			}
			var err error
			if debug {
				adatypes.Central.Log.Debugf("Search index of PE %s", v.Type().Name())
			}
			v, err = record.SearchValueIndex(fieldName, index)
			if err != nil {
				adatypes.Central.Log.Debugf("Error %s/%v: %v", fieldName, index, err)
				return -1
			}
			switch mv := v.(type) {
			case *adatypes.StructureValue:
				return int32(mv.NrElements())
			default:
			}
		}
		if v.Type().Type() == adatypes.FieldTypeMultiplefield {
			if debug {
				adatypes.Central.Log.Debugf("Quantity of %s MU elements", v.Type().Name())
			}
			mv := v.(*adatypes.StructureValue)
			return int32(mv.NrElements())
		}
		return 1
	}
	return -1
}

// Scan scan for different field entries
func (record *Record) Scan(dest ...interface{}) (err error) {
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Scan Record %#v", record.fields)
	}
	if f, ok := record.fields["#isn"]; ok {
		if debug {
			adatypes.Central.Log.Debugf("Fill Record ISN=%d", record.Isn)
		}
		*(dest[f.index].(*int)) = int(record.Isn)
	}
	if f, ok := record.fields["#isnquantity"]; ok {
		if debug {
			adatypes.Central.Log.Debugf("Fill Record ISN quantity=%d", record.Quantity)
		}
		*(dest[f.index].(*int)) = int(record.Quantity)
	}
	// Traverse to current entries
	tm := adatypes.TraverserValuesMethods{EnterFunction: scanFieldsTraverser}
	sf := &scanFields{fields: record.fields, parameter: dest}
	_, err = record.Traverse(tm, sf)
	if err != nil {
		return err
	}
	return nil

}

// traverseMarshalXML2 traverser used by the XML Marshaller
func traverseMarshalXML2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	enc := x.(*xml.Encoder)
	if adaValue.Type().IsStructure() {
		switch adaValue.Type().Type() {
		case adatypes.FieldTypePeriodGroup:
			peName := "Period"
			start := xml.StartElement{Name: xml.Name{Local: peName}}
			if adaValue.Type().Name() != adaValue.Type().ShortName() {
				start.Name.Local = adaValue.Type().Name()
			} else {
				attrs := make([]xml.Attr, 0)
				attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "sn"}, Value: adaValue.Type().Name()})
				start.Attr = attrs
			}
			_ = enc.EncodeToken(start)
		case adatypes.FieldTypeMultiplefield:
			muName := "Multiple"
			if adaValue.Type().Name() != adaValue.Type().ShortName() {
				muName = adaValue.Type().Name()
				start := xml.StartElement{Name: xml.Name{Local: muName}}
				_ = enc.EncodeToken(start)
			} else {
				start := xml.StartElement{Name: xml.Name{Local: muName}}
				attrs := make([]xml.Attr, 0)
				attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "sn"}, Value: adaValue.Type().Name()})
				start.Attr = attrs
				_ = enc.EncodeToken(start)
			}
		case adatypes.FieldTypeGroup:
			grName := "Group"
			if adaValue.Type().Name() != adaValue.Type().ShortName() {
				grName = adaValue.Type().Name()
				start := xml.StartElement{Name: xml.Name{Local: grName}}
				_ = enc.EncodeToken(start)
			} else {
				start := xml.StartElement{Name: xml.Name{Local: grName}}
				attrs := make([]xml.Attr, 0)
				attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "sn"}, Value: adaValue.Type().Name()})
				start.Attr = attrs
				_ = enc.EncodeToken(start)
			}
		default:
			start := xml.StartElement{Name: xml.Name{Local: adaValue.Type().Name()}}
			_ = enc.EncodeToken(start)
		}
	} else {
		isLink := strings.HasPrefix(adaValue.Type().Name(), "@")
		name := adaValue.Type().Name()
		if isLink {
			name = adaValue.Type().Name()[1:]
		}
		start := xml.StartElement{Name: xml.Name{Local: name}}
		if isLink {
			start.Attr = []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "link"}}
		}
		_ = enc.EncodeToken(start)
		x := adaValue.String()
		x = strings.Trim(x, " ")
		_ = enc.EncodeToken(xml.CharData([]byte(x)))
		_ = enc.EncodeToken(start.End())
	}
	return adatypes.Continue, nil
}

// traverseMarshalXMLEnd2 traverser end function used by the XML Marshaller
func traverseMarshalXMLEnd2(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue.Type().IsStructure() {
		enc := x.(*xml.Encoder)
		sv := adaValue.(*adatypes.StructureValue)
		if adaValue.Type().Type() == adatypes.FieldTypePeriodGroup && sv.NrElements() > 0 {
			end := xml.EndElement{Name: xml.Name{Local: "Entry"}}
			enc.EncodeToken(end)
		}
		if adaValue.Type().Type() == adatypes.FieldTypePeriodGroup {
			end := xml.EndElement{Name: xml.Name{Local: "Period"}}
			enc.EncodeToken(end)
		}
		if adaValue.Type().Type() == adatypes.FieldTypeMultiplefield {
			end := xml.EndElement{Name: xml.Name{Local: "Multiple"}}
			enc.EncodeToken(end)
		}
		if adaValue.Type().Type() == adatypes.FieldTypeGroup {
			end := xml.EndElement{Name: xml.Name{Local: "Group"}}
			enc.EncodeToken(end)
		}
		end := xml.EndElement{Name: xml.Name{Local: adaValue.Type().Name()}}
		enc.EncodeToken(end)
	}
	return adatypes.Continue, nil
}

// traverseMarshalXMLElement traverser element function used by the XML Marshaller
func traverseMarshalXMLElement(adaValue adatypes.IAdaValue, nr, max int, x interface{}) (adatypes.TraverseResult, error) {
	enc := x.(*xml.Encoder)
	if adaValue.Type().Type() == adatypes.FieldTypePeriodGroup {
		if nr > 0 {
			end := xml.EndElement{Name: xml.Name{Local: "Entry"}}
			enc.EncodeToken(end)
		}
		start := xml.StartElement{Name: xml.Name{Local: "Entry"}}
		enc.EncodeToken(start)
	}
	return adatypes.Continue, nil
}

// MarshalXML provide XML marshal method of a record
func (record *Record) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	adatypes.Central.Log.Debugf("Marshal XML record: %d", record.Isn)
	var rec xml.StartElement
	adatypes.Central.Log.Debugf("Map usage: %#v", record.adabasMap)
	if record.adabasMap != nil {
		rec = xml.StartElement{Name: xml.Name{Local: record.adabasMap.Name}}
	} else {
		rec = xml.StartElement{Name: xml.Name{Local: "Record"}}
	}
	if record.Isn > 0 {
		rec.Attr = []xml.Attr{{Name: xml.Name{Local: "ISN"}, Value: strconv.Itoa(int(record.Isn))}}
	}
	if record.Quantity > 0 {
		rec.Attr = []xml.Attr{{Name: xml.Name{Local: "Quantity"}, Value: strconv.Itoa(int(record.Quantity))}}
	}
	_ = e.EncodeToken(rec)
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalXML2, LeaveFunction: traverseMarshalXMLEnd2, ElementFunction: traverseMarshalXMLElement}
	_, err := record.Traverse(tm, e)
	if err != nil {
		return err
	}
	_ = e.EncodeToken(rec.End())
	adatypes.Central.Log.Debugf("Marshal XML record finished")

	return nil
}

// MarshalJSON provide JSON marshal function of a record
func (record *Record) MarshalJSON() ([]byte, error) {
	adatypes.Central.Log.Debugf("Marshal JSON record: %d", record.Isn)
	req := &responseJSON{special: true}
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseMarshalJSON, LeaveFunction: traverseMarshalJSONEnd,
		ElementFunction: traverseElementMarshalJSON}
	req.stack = adatypes.NewStack()

	dataMap := make(map[string]interface{})
	req.dataMap = &dataMap
	req.Values = append(req.Values, req.dataMap)
	if record.Isn > 0 {
		dataMap["ISN"] = record.Isn
	}
	if record.Quantity > 0 {
		dataMap["Quantity"] = record.Quantity
	}

	// Traverse record generating JSON
	_, err := record.Traverse(tm, req)
	if err != nil {
		return nil, err
	}
	adatypes.Central.Log.Debugf("Go JSON response %v -> %s", err, req.buffer.String())

	return json.Marshal(req.dataMap)
}

// TrimString search value and provide a trimmed string/alpha representation.
func (record *Record) TrimString(parameter ...interface{}) string {
	v, err := record.SearchValue(parameter...)
	if err == nil {
		return strings.Trim(v.String(), " ")
	}
	return ""
}

// Bytes search value and provide a raw byte slice representation.
func (record *Record) Bytes(parameter ...interface{}) []byte {
	v, err := record.SearchValue(parameter...)
	if err == nil {
		return v.Bytes()
	}
	return nil
}

// SelectValue provide selected list of value
func (record *Record) SelectValue(definition *adatypes.Definition) []adatypes.IAdaValue {
	activeValues := make([]adatypes.IAdaValue, 0)
	fn := definition.Fieldnames()
	for _, v := range record.Value {
		if contains(fn, v.Type().Name()) {
			activeValues = append(activeValues, v)
		}
	}
	return activeValues
}

func contains(array []string, str string) bool {
	for _, a := range array {
		if a == str {
			return true
		}
	}
	return false
}
