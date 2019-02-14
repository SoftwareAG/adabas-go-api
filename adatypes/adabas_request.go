package adatypes

import (
	"bytes"
	"fmt"
)

// RequestParser function callback used to go through the list of received buffer
type RequestParser func(adabasRequest *Request, x interface{}) error

// Request contains all relevant buffer and parameters for a Adabas call
type Request struct {
	FormatBuffer       bytes.Buffer
	RecordBuffer       *BufferHelper
	RecordBufferLength uint32
	RecordBufferShift  uint32
	PeriodLength       uint32
	SearchTree         *SearchTree
	Parser             RequestParser
	Limit              uint64
	Multifetch         uint32
	Descriptors        []string
	Definition         *Definition
	Response           uint16
	Isn                Isn
	IsnQuantity        uint64
	Option             *BufferOption
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
	Central.Log.Debugf("After %s current Record length %d", adaValue.Type().Name(), adabasRequest.RecordBufferLength)
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
	Central.Log.Debugf("Format Buffer Read traverser: %s level=%d/%d", adaType.Name(), adaType.Level(), level)
	adabasRequest := x.(*Request)
	Central.Log.Debugf("Curent Record Buffer length : %d", adabasRequest.RecordBufferLength)
	buffer := &(adabasRequest.FormatBuffer)
	switch adaType.Type() {
	case FieldTypePeriodGroup:
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		structureType := adaType.(*StructureType)
		r := structureType.Range.FormatBuffer()
		Central.Log.Debugf("------->>>>>> Range %s=%s", structureType.name, r)
		buffer.WriteString(adaType.ShortName() + "C,4")
		adabasRequest.RecordBufferLength += 4
		if !adaType.HasFlagSet(FlagOptionMU) {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("%s%s", adaType.ShortName(), r))
			adabasRequest.RecordBufferLength += adabasRequest.Option.multipleSize
		}
	case FieldTypeMultiplefield:
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		if adaType.HasFlagSet(FlagOptionPE) {
			buffer.WriteString(adaType.ShortName() + "1-NC,4")
		} else {
			buffer.WriteString(adaType.ShortName() + "C,4")
		}
		adabasRequest.RecordBufferLength += 4
		if !adaType.HasFlagSet(FlagOptionPE) {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			strType := adaType.(*StructureType)
			subType := strType.SubTypes[0]
			buffer.WriteString(fmt.Sprintf("%s1-N,%d,%s", adaType.ShortName(), subType.Length(), subType.Type().FormatCharacter()))
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
						fieldIndex = "1-N"
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
func (def *Definition) CreateAdabasRequest(store bool, secondCall bool) (adabasRequest *Request, err error) {
	adabasRequest = &Request{FormatBuffer: bytes.Buffer{}, Option: NewBufferOption(store, secondCall)}

	Central.Log.Debugf("Create format buffer. Init Buffer: %s", adabasRequest.FormatBuffer.String())
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
