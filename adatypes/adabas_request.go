/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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

package adatypes

import (
	"bytes"
	"fmt"
)

// RequestParser function callback used to go through the list of received buffer
type RequestParser func(adabasRequest *Request, x interface{}) error

// HoldType hold enum type
type HoldType uint32

const (
	// HoldNone no hold
	HoldNone HoldType = iota
	// HoldWait wait for hold released
	HoldWait
	// HoldResponse receive response code
	HoldResponse
)

// Request contains all relevant buffer and parameters for a Adabas call
type Request struct {
	FormatBuffer       bytes.Buffer
	RecordBuffer       *BufferHelper
	RecordBufferLength uint32
	RecordBufferShift  uint32
	PeriodLength       uint32
	SearchTree         *SearchTree
	Parser             RequestParser
	HoldRecords        HoldType
	Limit              uint64
	Multifetch         uint32
	Descriptors        []string
	Definition         *Definition
	Response           uint16
	Isn                Isn
	IsnQuantity        uint64
	Option             *BufferOption
	Parameter          interface{}
}

func (adabasRequest *Request) reset() {
	adabasRequest.SearchTree = nil
	adabasRequest.Definition = nil
}

type valueSearch struct {
	name     string
	adaValue IAdaValue
}

func searchRequestValue(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	vs := x.(*valueSearch)
	if adaValue.Type().Name() == vs.name {
		vs.adaValue = adaValue
		return EndTraverser, nil
	}
	return Continue, nil
}

// GetValue get the value for string with name
func (adabasRequest *Request) GetValue(name string) (IAdaValue, error) {
	vs := &valueSearch{name: name}
	tm := TraverserValuesMethods{EnterFunction: searchRequestValue}
	if adabasRequest.Definition == nil {
		return nil, NewGenericError(26)
	}
	_, err := adabasRequest.Definition.TraverseValues(tm, vs)
	if err != nil {
		return nil, err
	}
	return vs.adaValue, nil
}

// Traverser callback to create format buffer per field type
func formatBufferTraverserEnter(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	adabasRequest := x.(*Request)
	if adaValue.Type().HasFlagSet(FlagOptionReference) {
		return Continue, nil
	}
	Central.Log.Debugf("Add format buffer for %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		// Reset if period group starts
		if adaValue.Type().Level() == 1 && adaValue.Type().Type() == FieldTypePeriodGroup {
			adabasRequest.PeriodLength = 0
		}
		len := adaValue.FormatBuffer(&(adabasRequest.FormatBuffer), adabasRequest.Option)
		adabasRequest.RecordBufferLength += len
		adabasRequest.PeriodLength += len
	} else {
		len := adaValue.FormatBuffer(&(adabasRequest.FormatBuffer), adabasRequest.Option)
		adabasRequest.RecordBufferLength += len
		adabasRequest.PeriodLength += len
	}
	if adabasRequest.Option.SecondCall &&
		adaValue.Type().Type() == FieldTypeMultiplefield && adaValue.Type().HasFlagSet(FlagOptionPE) {
		return SkipTree, nil
	}
	Central.Log.Debugf("After %s current Record length %d -> %s", adaValue.Type().Name(), adabasRequest.RecordBufferLength,
		adabasRequest.FormatBuffer.String())
	return Continue, nil
}

// Traverse callback function to create format buffer and record buffer length
func formatBufferTraverserLeave(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	Central.Log.Debugf("Leave structure %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		// Reset if period group starts
		if adaValue.Type().Level() == 1 && adaValue.Type().Type() == FieldTypePeriodGroup {
			fb := x.(*Request)
			if fb.PeriodLength == 0 {
				fb.PeriodLength += 10
			}
			Central.Log.Debugf("Increase period buffer 10 times with %d", fb.PeriodLength)
			fb.RecordBufferLength += (10 * fb.PeriodLength)
			fb.PeriodLength = 0
		}
	}
	Central.Log.Debugf("Leave %s", adaValue.Type().Name())
	return Continue, nil
}

func formatBufferReadTraverser(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	Central.Log.Debugf("Format Buffer Read traverser: %s-%s level=%d/%d", adaType.Name(), adaType.ShortName(),
		adaType.Level(), level)
	if adaType.HasFlagSet(FlagOptionReference) {
		return nil
	}
	adabasRequest := x.(*Request)
	Central.Log.Debugf("Curent Record Buffer length : %d", adabasRequest.RecordBufferLength)
	buffer := &(adabasRequest.FormatBuffer)
	switch adaType.Type() {
	case FieldTypePeriodGroup:
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		structureType := adaType.(*StructureType)
		r := structureType.peRange.FormatBuffer()
		Central.Log.Debugf("------->>>>>> Range %s=%s%s %p", structureType.name, structureType.shortName, r, structureType)
		buffer.WriteString(adaType.ShortName() + "C,4,B")
		adabasRequest.RecordBufferLength += 4
		if !adaType.HasFlagSet(FlagOptionMU) {
			Central.Log.Debugf("No MU field, use general range group query")
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s%s", adaType.ShortName(), r))
			adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
		}
	case FieldTypeMultiplefield:
		if adaType.HasFlagSet(FlagOptionPE) {
			// structureType := adaType.(*StructureType)
			// r := structureType.peRange.FormatBuffer()
			// buffer.WriteString(adaType.ShortName() + r + "C,4,B")
		} else {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(adaType.ShortName() + "C,4,B")
		}
		adabasRequest.RecordBufferLength += 4
		if !adaType.HasFlagSet(FlagOptionPE) {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			strType := adaType.(*StructureType)
			subType := strType.SubTypes[0]
			r := strType.muRange.FormatBuffer()
			Central.Log.Debugf("Multiple range: %s", r)
			buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", adaType.ShortName(), r, subType.Length(), subType.Type().FormatCharacter()))
			adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
		}
	case FieldTypeSuperDesc, FieldTypeHyperDesc:
		if !(adaType.IsOption(FieldOptionPE) || adaType.IsOption(FieldOptionPE)) {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s,%d", adaType.ShortName(),
				adaType.Length()))
			adabasRequest.RecordBufferLength += adaType.Length()
		}
	case FieldTypePhonetic, FieldTypeCollation, FieldTypeReferential:
	default:
		if !adaType.IsStructure() {
			if !adaType.HasFlagSet(FlagOptionMUGhost) && (!adaType.HasFlagSet(FlagOptionPE) ||
				(adaType.HasFlagSet(FlagOptionPE) && adaType.HasFlagSet(FlagOptionMU))) {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				fieldIndex := ""
				if adaType.Type() == FieldTypeLBString {
					buffer.WriteString(fmt.Sprintf("%sL,4,%s%s(1,%d)", adaType.ShortName(), adaType.ShortName(), fieldIndex,
						PartialLobSize))
					adabasRequest.RecordBufferLength += (4 + PartialLobSize)
				} else {
					if adaType.HasFlagSet(FlagOptionPE) {
						t := adaType.(*AdaType)
						// fieldIndex = "1-N"
						fieldIndex = t.peRange.FormatBuffer()
						adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
					} else {
						if adaType.Length() == uint32(0) {
							adabasRequest.RecordBufferLength += 512
						} else {
							adabasRequest.RecordBufferLength += adaType.Length()
						}
					}
					buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", adaType.ShortName(), fieldIndex,
						adaType.Length(), adaType.Type().FormatCharacter()))
				}
			}
		}
	}
	Central.Log.Debugf("Final type generated Format Buffer : %s", buffer.String())
	Central.Log.Debugf("Final Record Buffer length : %d", adabasRequest.RecordBufferLength)
	return nil
}

// CreateAdabasRequest creates format buffer out of defined metadata tree
func (def *Definition) CreateAdabasRequest(store bool, secondCall bool, mainframe bool) (adabasRequest *Request, err error) {
	adabasRequest = &Request{FormatBuffer: bytes.Buffer{}, Option: NewBufferOption3(store, secondCall, mainframe)}

	Central.Log.Debugf("Create format buffer. Init Buffer: %s second=%v", adabasRequest.FormatBuffer.String(), secondCall)
	if store || secondCall {
		t := TraverserValuesMethods{EnterFunction: formatBufferTraverserEnter, LeaveFunction: formatBufferTraverserLeave}
		_, err = def.TraverseValues(t, adabasRequest)
		if err != nil {
			return
		}
	} else {
		t := TraverserMethods{EnterFunction: formatBufferReadTraverser}
		err = def.TraverseTypes(t, true, adabasRequest)
		if err != nil {
			return nil, err
		}
	}
	_, err = adabasRequest.FormatBuffer.WriteString(".")
	if err != nil {
		return nil, err
	}
	Central.Log.Debugf("Generated FB: %s", adabasRequest.FormatBuffer.String())
	Central.Log.Debugf("RB size=%d", adabasRequest.RecordBufferLength)
	return
}
