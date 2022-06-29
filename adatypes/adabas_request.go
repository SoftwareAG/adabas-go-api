/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"strings"
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
	// HoldResponse receive response code if record is in hold state
	HoldResponse
	// HoldAccess check during read that the record is not in hold (shared lock 'C')
	HoldAccess
	// HoldRead use shared lock until next read operation (shared lock 'Q')
	HoldRead
	// HoldTransaction use shared lock until end of transaction (shared lock 'S')
	HoldTransaction
)

var holdOption = []byte{' ', ' ', ' ', 'C', 'Q', 'S'}

// HoldOption return hold option for Adabas option 3
func (ht HoldType) HoldOption() byte {
	return holdOption[ht]
}

// IsHold check if hold type is hold
func (ht HoldType) IsHold() bool {
	return ht != HoldNone
}

const (
	// DefaultMultifetchLimit default number of multifetch entries
	DefaultMultifetchLimit = 10
	// AdaNormal Adabas success response code
	AdaNormal = 0
)

// IAdaCallInterface caller interface
type IAdaCallInterface interface {
	SendSecondCall(adabasRequest *Request, x interface{}) (err error)
	CallAdabas() (err error)
}

// Request contains all relevant buffer and parameters for a Adabas call
type Request struct {
	Caller             IAdaCallInterface
	FormatBuffer       bytes.Buffer
	RecordBuffer       *BufferHelper
	MultifetchBuffer   *BufferHelper
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
	CmdCode            [2]byte
	IsnIncrease        bool
	StoreIsn           bool
	DescriptorRead     bool
	CbIsn              Isn
	Isn                Isn
	IsnQuantity        uint64
	IsnLowerLimit      uint64
	Option             *BufferOption
	Parameter          interface{}
	Reference          string
	DataType           *DynamicInterface
	PartialLobSize     uint32
}

// func (adabasRequest *Request) reset() {
// 	adabasRequest.SearchTree = nil
// 	adabasRequest.Definition = nil
// }

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

// Traverser callback to create format buffer per field type. This method is called on each value entry working to generate
// Format Buffer dependent on the corresponding field types
func formatBufferTraverserEnter(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	adabasRequest := x.(*Request)
	if adaValue.Type().HasFlagSet(FlagOptionReadOnly) || adaValue.Type().HasFlagSet(FlagOptionReference) {
		return Continue, nil
	}
	Central.Log.Debugf("Add format buffer for %s(%s)", adaValue.Type().Name(), adaValue.Type().Type().name())
	switch {
	// In case of structure generate
	case adaValue.Type().IsStructure():
		// Reset if period group starts
		if adaValue.Type().Level() == 1 && adaValue.Type().Type() == FieldTypePeriodGroup {
			adabasRequest.PeriodLength = 0
		}
	// special formater
	case strings.HasPrefix(adaValue.Type().Name(), "#") || strings.HasPrefix(adaValue.Type().Name(), "@"):
	case adabasRequest.Definition != nil && !adabasRequest.Definition.CheckField(adaValue.Type().Name()):
		Central.Log.Debugf("Skip format buffer for %s", adaValue.Type().Name())
		return Continue, nil

	default:
	}

	Central.Log.Debugf("Generate format buffer %s %s", adaValue.Type().Name(), adaValue.Type().Type().name())
	len := adaValue.FormatBuffer(&(adabasRequest.FormatBuffer), adabasRequest.Option)
	adabasRequest.RecordBufferLength += len
	adabasRequest.PeriodLength += len
	if adabasRequest.Option.SecondCall > 0 &&
		adaValue.Type().Type() == FieldTypeMultiplefield && adaValue.Type().HasFlagSet(FlagOptionPE) {
		return SkipTree, nil
	}
	if adaValue.Type().Type() == FieldTypeRedefinition {
		return SkipTree, nil
	}
	Central.Log.Debugf("After %s current Record length %d -> %s", adaValue.Type().Name(), adabasRequest.RecordBufferLength,
		adabasRequest.FormatBuffer.String())
	return Continue, nil
}

type traverseSubType struct {
	enable  bool
	field   string
	request *Request
}

// Traverse callback function to create format buffer and record buffer length
func formatBufferTraverserLeave(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	Central.Log.Debugf("Leave structure %s %v", adaValue.Type().Name(), adaValue.Type().Type())
	if adaValue.Type().IsStructure() {
		// Reset if period group starts
		if adaValue.Type().Level() == 1 && adaValue.Type().Type() == FieldTypePeriodGroup {
			request := x.(*Request)
			if request.PeriodLength == 0 {
				request.PeriodLength += 10
			}
			Central.Log.Debugf("Increase period buffer 10 times with %d", request.PeriodLength)
			request.RecordBufferLength += (10 * request.PeriodLength)
			request.PeriodLength = 0

			sv := adaValue.(*StructureValue)
			if !request.Option.StoreCall && len(sv.Elements) == 0 {
				Central.Log.Debugf("Traverse single index")
				t := TraverserMethods{EnterFunction: formatBufferPeriodReadTraverser}
				err := request.Definition.TraverseTypes(t, true,
					&traverseSubType{false, adaValue.Type().Name(), request})
				if err != nil {
					return EndTraverser, err
				}
			}
		}
	}
	Central.Log.Debugf("Leave %s", adaValue.Type().Name())
	return Continue, nil
}

func formatBufferPeriodReadTraverser(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	tst := x.(*traverseSubType)
	if level == 1 {
		if tst.field == adaType.Name() {
			tst.enable = true
		} else {
			tst.enable = false
		}
	} else if tst.enable {
		return formatBufferReadTraverser(adaType, parentType, level, tst.request)
	}
	return nil
}

func generateFormatBufferPeriodGroup(adabasRequest *Request, adaType IAdaType) {
	if !adabasRequest.DescriptorRead {
		Central.Log.Debugf(" FOSI: %v", adaType.HasFlagSet(FlagOptionSingleIndex))
		if !adaType.HasFlagSet(FlagOptionSingleIndex) {
			buffer := &(adabasRequest.FormatBuffer)
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			structureType := adaType.(*StructureType)
			r := structureType.peRange.FormatBuffer()
			Central.Log.Debugf("------->>>>>> Range %s=%s%s %p", structureType.name, structureType.shortName, r, structureType)
			buffer.WriteString(adaType.ShortName() + "C,4,B")
			adabasRequest.RecordBufferLength += 4
			if !adaType.HasFlagSet(FlagOptionAtomicFB) && !adaType.HasFlagSet(FlagOptionPart) {
				Central.Log.Debugf("No MU field, use general range group query")
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				buffer.WriteString(fmt.Sprintf("%s%s", adaType.ShortName(), r))
				adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
			}
		}
	}
}

func generateFormatBufferMultipleField(adabasRequest *Request, adaType IAdaType) {
	buffer := &(adabasRequest.FormatBuffer)
	if adaType.HasFlagSet(FlagOptionPE) {
		structureType := adaType.(*StructureType)
		// r := structureType.peRange.FormatBuffer()
		// buffer.WriteString(adaType.ShortName() + r + "C,4,B")
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Periodic range FB: %s", structureType.PeriodicRange().FormatBuffer())
			Central.Log.Debugf("Multiple range FB: %s", structureType.MultipleRange().FormatBuffer())
			Central.Log.Debugf("Partial  range FB: %s", structureType.PartialRange().FormatBuffer())
			Central.Log.Debugf("Single index: %v", structureType.HasFlagSet(FlagOptionSingleIndex))
			Central.Log.Debugf("Part: %v", structureType.HasFlagSet(FlagOptionPart))
		}
		if adaType.PeriodicRange().IsSingleIndex() {
			structureType := adaType.(*StructureType)
			// fmt.Println("PE Range:", structureType.peRange.FormatBuffer())
			// fmt.Println("MU Range:", structureType.muRange.FormatBuffer())
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			at := structureType.SubTypes[0]
			if !at.MultipleRange().IsSingleIndex() {
				buffer.WriteString(adaType.ShortName() + structureType.peRange.FormatBuffer() + "C,4,B,")
			}
			buffer.WriteString(fmt.Sprintf("%s%s(%s),%d,%s",
				at.ShortName(), at.PeriodicRange().FormatBuffer(), at.MultipleRange().FormatBuffer(),
				at.Length(), at.Type().FormatCharacter()))
		}
		if adaType.Type() == FieldTypeMultiplefield && adaType.MultipleRange().IsSingleIndex() {
			Central.Log.Debugf("Single index, skip MU FormatBuffer")
			return
		}
	} else {
		structureType := adaType.(*StructureType)
		at := structureType.SubTypes[0]
		Central.Log.Debugf("Multiple range FB CS: %s", structureType.MultipleRange().FormatBuffer())
		Central.Log.Debugf("Multiple range FB C: %s", at.MultipleRange().FormatBuffer())
		if !structureType.MultipleRange().IsSingleIndex() {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(adaType.ShortName() + "C,4,B")
		}
	}
	adabasRequest.RecordBufferLength += 4
	if !adaType.HasFlagSet(FlagOptionPE) {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		strType := adaType.(*StructureType)
		subType := strType.SubTypes[0]
		r := strType.muRange.FormatBuffer()
		Central.Log.Debugf("Multiple range FB: %s", r)
		buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", adaType.ShortName(), r, subType.Length(), subType.Type().FormatCharacter()))
		adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
	}
	Central.Log.Debugf("FB MU %s", buffer.String())
}

func generateFormatBufferSpecialDescriptor(adabasRequest *Request, adaType IAdaType) {
	if !adaType.IsOption(FieldOptionPE) {
		buffer := &(adabasRequest.FormatBuffer)
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(fmt.Sprintf("%s,%d,A", adaType.ShortName(),
			adaType.Length()))
		adabasRequest.RecordBufferLength += adaType.Length()
	} else {
		if adabasRequest.DescriptorRead {
			buffer := &(adabasRequest.FormatBuffer)
			buffer.WriteString(fmt.Sprintf("%s,%d,A", adaType.ShortName(),
				adaType.Length()))
			adabasRequest.RecordBufferLength += adaType.Length()
		} else {
			Central.Log.Debugf("Skip PE Desc %s", adaType.ShortName())
		}
	}
}

func generateFormatBufferField(adabasRequest *Request, adaType IAdaType) {
	if !adaType.IsStructure() {
		buffer := &(adabasRequest.FormatBuffer)
		if adabasRequest.DescriptorRead {
			buffer.WriteString(adaType.ShortName())
			adabasRequest.RecordBufferLength += adaType.Length()
		} else {
			Central.Log.Debugf(" MUGhost: %v", adaType.HasFlagSet(FlagOptionMUGhost))
			Central.Log.Debugf(" SingleIndex: %v", adaType.HasFlagSet(FlagOptionSingleIndex))
			Central.Log.Debugf(" PE: %v", adaType.HasFlagSet(FlagOptionPE))
			Central.Log.Debugf(" AtomicFB: %v", adaType.HasFlagSet(FlagOptionAtomicFB))
			Central.Log.Debugf(" Part: %v", adaType.HasFlagSet(FlagOptionPart))
			Central.Log.Debugf(" PartialRead: %v", adabasRequest.Option.PartialRead)
			if !adaType.HasFlagSet(FlagOptionMUGhost) && (!adaType.HasFlagSet(FlagOptionPE) ||
				(adaType.HasFlagSet(FlagOptionPE) && (adaType.HasFlagSet(FlagOptionAtomicFB) || adaType.HasFlagSet(FlagOptionPart)))) {
				if buffer.Len() > 0 {
					buffer.WriteString(",")
				}
				fieldIndex := ""
				genType := adaType
				if adaType.Type() == FieldTypeRedefinition {
					genType = adaType.(*RedefinitionType).MainType
				}
				if adaType.Type() == FieldTypeLBString {
					indexRange := getIndexRange(adaType)
					partialRange := adaType.PartialRange()
					if partialRange != nil {
						Central.Log.Debugf("Partial Range %d:%d\n", partialRange.from, partialRange.to)
						if partialRange.from == 0 {
							buffer.WriteString(fmt.Sprintf("%s%s(*,%d)", adaType.ShortName(), indexRange, partialRange.to))
						} else {
							buffer.WriteString(fmt.Sprintf("%s%s(%d,%d)", adaType.ShortName(), indexRange, partialRange.from, partialRange.to))
						}
						adabasRequest.RecordBufferLength += uint32(partialRange.to)
					} else {
						if genType.HasFlagSet(FlagOptionPE) {
							t := genType.(*AdaType)
							fieldIndex = t.peRange.FormatBuffer()
						}
						if adabasRequest.Option.PartialRead {
							buffer.WriteString(fmt.Sprintf("%sL%s,4,%s%s(*,%d)", adaType.ShortName(), indexRange, adaType.ShortName(), fieldIndex,
								adabasRequest.PartialLobSize))
						} else {
							buffer.WriteString(fmt.Sprintf("%sL%s,4,%s%s(1,%d)", adaType.ShortName(), indexRange, adaType.ShortName(), fieldIndex,
								adabasRequest.PartialLobSize))
						}
						adabasRequest.RecordBufferLength += (4 + adabasRequest.PartialLobSize)
					}
				} else {
					if genType.HasFlagSet(FlagOptionPE) {
						t := genType.(*AdaType)
						// fieldIndex = "1-N"
						fieldIndex = t.peRange.FormatBuffer()
						adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
					} else {
						if genType.Length() == uint32(0) {
							adabasRequest.RecordBufferLength += 512
						} else {
							adabasRequest.RecordBufferLength += adaType.Length()
						}
					}
					if Central.IsDebugLevel() {
						Central.Log.Debugf("FB generate %T %s -> %s field index=%s", adaType, genType.ShortName(), genType.Type().FormatCharacter(), fieldIndex)
						// TODO check pe range
						ft := adaType.(*AdaType)
						Central.Log.Debugf("FB %s peRange=%s muRange=%s", ft.name, ft.peRange.FormatBuffer(), ft.muRange.FormatBuffer())
					}
					buffer.WriteString(fmt.Sprintf("%s%s,%d,%s", genType.ShortName(), fieldIndex,
						genType.Length(), genType.Type().FormatCharacter()))
				}
			} else {
				Central.Log.Debugf("MU ghost or PE of %s (%s)", adaType.Name(), adaType.Type().name())
				Central.Log.Debugf("PeriodRange %s", adaType.PeriodicRange().FormatBuffer())
				Central.Log.Debugf("MultipleRange %s", adaType.MultipleRange().FormatBuffer())
				if adaType.HasFlagSet(FlagOptionSingleIndex) {
					Central.Log.Debugf("Create single index FB")
				}
			}
		}
	} else {
		Central.Log.Infof("Unknown FB generator: %v", adaType.Name())
	}

}

func getIndexRange(adaType IAdaType) string {
	indexRange := ""
	if Central.IsDebugLevel() {
		Central.Log.Debugf("PE range %s", adaType.PeriodicRange().FormatBuffer())
		Central.Log.Debugf("MU range %s", adaType.MultipleRange().FormatBuffer())
	}
	if adaType.PeriodicRange().from > 0 && adaType.PeriodicRange().to != LastEntry {
		indexRange = fmt.Sprintf("%d", adaType.PeriodicRange().from)
	}
	if adaType.MultipleRange().from > 0 && adaType.MultipleRange().to != LastEntry {
		if indexRange != "" {
			indexRange += fmt.Sprintf("(%d)", adaType.MultipleRange().from)
		} else {
			indexRange += fmt.Sprintf("%d", adaType.MultipleRange().from)
		}
	}
	Central.Log.Debugf("Index Range %s", indexRange)
	return indexRange
}

func formatBufferReadTraverser(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	Central.Log.Debugf("Format Buffer Read traverser: %s-%s level=%d/%d -> %T", adaType.Name(), adaType.ShortName(),
		adaType.Level(), level, adaType)
	if adaType.HasFlagSet(FlagOptionReference) {
		Central.Log.Debugf("Skip FB reference generation [%s]", adaType.Name())
		return nil
	}
	adabasRequest := x.(*Request)
	Central.Log.Debugf("Curent Record Buffer length : %d", adabasRequest.RecordBufferLength)
	switch adaType.Type() {
	case FieldTypePeriodGroup:
		generateFormatBufferPeriodGroup(adabasRequest, adaType)
	case FieldTypeMultiplefield:
		generateFormatBufferMultipleField(adabasRequest, adaType)
	case FieldTypeSuperDesc, FieldTypeHyperDesc:
		generateFormatBufferSpecialDescriptor(adabasRequest, adaType)
	case FieldTypeFieldLength:
		buffer := &(adabasRequest.FormatBuffer)
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		fn := adaType.ShortName()
		if fn[0] == '#' {
			fn = fn[1:]
		}
		if adaType.HasFlagSet(FlagOptionLengthPE) {
			buffer.WriteString(fmt.Sprintf("%sC,4,B", fn))
		} else {
			buffer.WriteString(fmt.Sprintf("%sL,4,B", fn))
		}
		adabasRequest.RecordBufferLength += 4
	case FieldTypePhonetic, FieldTypeCollation, FieldTypeReferential:
	case FieldTypeRedefinition:
		buffer := &(adabasRequest.FormatBuffer)
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		genType := adaType.(*RedefinitionType).MainType
		buffer.WriteString(fmt.Sprintf("%s,%d,%s", genType.ShortName(),
			genType.Length(), genType.Type().FormatCharacter()))
		adabasRequest.RecordBufferLength += genType.Length()
	default:
		generateFormatBufferField(adabasRequest, adaType)
	}
	if Central.IsDebugLevel() {
		buffer := &(adabasRequest.FormatBuffer)
		Central.Log.Debugf("Final type generated Format Buffer : %s", buffer.String())
		//debug.PrintStack()
		Central.Log.Debugf("Final Record Buffer length : %d", adabasRequest.RecordBufferLength)
	}
	return nil
}

// AdabasRequestParameter Adabas request parameter defining type of Adabas request
type AdabasRequestParameter struct {
	Store          bool
	SingleRead     bool
	DescriptorRead bool
	PartialRead    bool
	SecondCall     uint32
	Mainframe      bool
	BlockSize      uint32
}

// CreateAdabasRequest creates format buffer out of defined metadata tree
func (def *Definition) CreateAdabasRequest(parameter *AdabasRequestParameter) (adabasRequest *Request, err error) {
	adabasRequest = &Request{FormatBuffer: bytes.Buffer{}, Option: NewBufferOption3(parameter.Store, parameter.SecondCall, parameter.Mainframe),
		Multifetch: DefaultMultifetchLimit, DescriptorRead: parameter.DescriptorRead, PartialLobSize: parameter.BlockSize}
	adabasRequest.Option.DescriptorRead = parameter.DescriptorRead
	adabasRequest.Option.PartialRead = parameter.PartialRead
	adabasRequest.Definition = def
	if parameter.BlockSize == 0 {
		adabasRequest.PartialLobSize = PartialLobSize
		adabasRequest.Option.BlockSize = PartialLobSize
	} else {
		adabasRequest.PartialLobSize = parameter.BlockSize
		adabasRequest.Option.BlockSize = parameter.BlockSize
	}

	if parameter.Store || parameter.SecondCall > 0 || def.Values != nil {
		Central.Log.Debugf("Create defined on values format buffer. Init Buffer: %s second=%v", adabasRequest.FormatBuffer.String(), parameter.SecondCall)
		t := TraverserValuesMethods{EnterFunction: formatBufferTraverserEnter, LeaveFunction: formatBufferTraverserLeave}
		_, err = def.TraverseValues(t, adabasRequest)
		if err != nil {
			return
		}
	} else {
		Central.Log.Debugf("Create defined on types format buffer. Init Buffer: %s second=%v", adabasRequest.FormatBuffer.String(), parameter.SecondCall)
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

// ParseBuffer parse given record buffer and multifetch buffer and put
// all data into the given definition value tree, corresponding to the
// field definition of the concurrent field
func (adabasRequest *Request) ParseBuffer(count *uint64, x interface{}) (responseCode uint32, err error) {
	Central.Log.Debugf("Parse Adabas request buffers avail.=%v", (adabasRequest.Definition.Values != nil))
	// If parser is available, use the parser to extract content
	if adabasRequest.Parser != nil {
		Central.Log.Debugf("Parser method found")
		var multifetchHelper *BufferHelper
		nrMultifetchEntries := uint32(1)
		if adabasRequest.Multifetch > 1 {
			Central.Log.Debugf("Multifetch %d", adabasRequest.Multifetch)
			multifetchHelper = adabasRequest.MultifetchBuffer
			nrMultifetchEntries, err = multifetchHelper.ReceiveUInt32()
			if err != nil {
				Central.Log.Debugf("Error evaluate multifetch entries %v", err)
				return
			}
			if nrMultifetchEntries > 10000 {
				Central.Log.Debugf("multifetch entries mismatch, panic ...")
				return 0, NewGenericError(177)
			}
		}
		Central.Log.Debugf("Nr Multifetch entries %d", nrMultifetchEntries)
		for nrMultifetchEntries > 0 {
			(*count)++
			if multifetchHelper != nil {
				responseCode, err = adabasRequest.readMultifetch(multifetchHelper)
				if err != nil {
					Central.Log.Debugf("Multifetch parse error: %v", err)
					return
				}
				if responseCode != AdaNormal {
					Central.Log.Debugf("Adabas multifetch response received %d at %d", responseCode, *count)
					break
				}
			}

			Central.Log.Debugf("Parse Buffer .... values avail.=%v", (adabasRequest.Definition.Values != nil))
			adabasRequest.Option.LowerLimit = adabasRequest.IsnLowerLimit
			// Parse the received request
			prefix := fmt.Sprintf("/image/%s/%d/", adabasRequest.Reference, adabasRequest.Isn)
			_, err = adabasRequest.Definition.ParseBuffer(adabasRequest.RecordBuffer, adabasRequest.Option, prefix)
			if err != nil {
				return
			}
			adabasRequest.Definition.DumpValues(true)
			Central.Log.Debugf("Ready parse buffer .... %p values avail.=%v", adabasRequest.Definition, (adabasRequest.Definition.Values == nil))
			if adabasRequest.Caller != nil {
				err = adabasRequest.Caller.SendSecondCall(adabasRequest, x)
				if err != nil {
					return
				}
			}
			adabasRequest.Definition.DumpValues(true)
			Central.Log.Debugf("Found parser .... values avail.=%v", (adabasRequest.Definition.Values == nil))
			err = adabasRequest.Parser(adabasRequest, x)
			if err != nil {
				return
			}
			nrMultifetchEntries--

			// If multifetch on, create values for next parse step, only possible on read calls
			if nrMultifetchEntries > 0 {
				Central.Log.Debugf("Create multifetch values")
				//adabasRequest.Definition.Values = nil
				err = adabasRequest.Definition.CreateValues(false)
				if err != nil {
					return
				}
			}
		}
		Central.Log.Debugf("Parser ended")
	} else {
		Central.Log.Debugf("Found no parser")
	}
	return
}

// Parse multifetch values
func (adabasRequest *Request) readMultifetch(multifetchHelper *BufferHelper) (responseCode uint32, err error) {
	recordLength, rErr := multifetchHelper.ReceiveUInt32()
	if rErr != nil {
		err = rErr
		return
	}
	Central.Log.Debugf("Record length %d", recordLength)
	responseCode, err = multifetchHelper.ReceiveUInt32()
	if err != nil {
		Central.Log.Debugf("Response parser error in MF %v", err)
		return
	}
	if responseCode != AdaNormal {
		adabasRequest.Response = uint16(responseCode) // adabas.Acbx.Acbxrsp
		Central.Log.Debugf("Response code in MF %v", adabasRequest.Response)
		return
	}
	Central.Log.Debugf("Response code %d", responseCode)
	isn, isnErr := multifetchHelper.ReceiveUInt32()
	if isnErr != nil {
		err = isnErr
		return
	}
	Central.Log.Debugf("Got ISN %d", isn)
	adabasRequest.Isn = Isn(isn)
	if adabasRequest.StoreIsn {
		adabasRequest.CbIsn = Isn(isn)
	}
	quantity, qerr := multifetchHelper.ReceiveUInt32()
	if qerr != nil {
		Central.Log.Debugf("Quantity buffer error %v", qerr)
		err = qerr
		return
	}
	Central.Log.Debugf("ISN quantity=%d", quantity)
	adabasRequest.IsnQuantity = uint64(quantity)
	return
}
