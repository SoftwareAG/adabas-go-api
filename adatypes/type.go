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

package adatypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// OccByte Occurrence identifier indicating that the occurrence is defined as byte
const OccByte = -12

// OccUInt2 Occurrence identifier indicating that the occurrence is defined as uint32
const OccUInt2 = -11

// OccNone Occurrence identifier indicating that the occurrence is not used
const OccNone = -10

// OccSingle Occurrence identifier indicating that the occurrence single
const OccSingle = -9

// OccCapacity Occurrence identifier indicating that the occurrence capactity of 2 or PE fields
const OccCapacity = -8

// NoReferenceField field out of range of given field possibilities
const NoReferenceField = math.MaxInt32

// IAdaType data type interface defined for all types
type IAdaType interface {
	Type() FieldType
	String() string
	Name() string
	ShortName() string
	SetName(string)
	Value() (IAdaValue, error)
	Length() uint32
	SetLength(uint32)
	SetRange(*AdaRange)
	PartialRange() *AdaRange
	SetPartialRange(*AdaRange)
	PeriodicRange() *AdaRange
	MultipleRange() *AdaRange
	IsStructure() bool
	Level() uint8
	SetLevel(uint8)
	Option() string
	SetParent(IAdaType)
	GetParent() IAdaType
	HasFlagSet(FlagOption) bool
	AddFlag(FlagOption)
	RemoveFlag(FlagOption)
	IsOption(FieldOption) bool
	AddOption(FieldOption)
	IsSpecialDescriptor() bool
	SetEndian(binary.ByteOrder)
	Endian() binary.ByteOrder
	SetFractional(uint32)
	SetCharset(string)
	SetFormatType(rune)
	FormatType() rune
	SetFormatLength(uint32)
	Fractional() uint32
	Convert() ConvertUnicode
}

// AdaType data type structure for field types, no structures
type AdaType struct {
	CommonType
	SysField   byte
	EditMask   byte
	SubOption  byte
	FractValue uint32
}

// FieldCondition field condition reference using for parser length management
type FieldCondition struct {
	lengthFieldIndex int
	refField         int
	conditionMatrix  map[byte][]byte
}

//func NewFieldCondition(index int, ref int, condition map[byte][]byte) FieldCondition {

// NewFieldCondition creates a new field condition
func NewFieldCondition(param ...interface{}) FieldCondition {
	lengthFieldIndex := -1
	refField := NoReferenceField
	var conditionMatrix map[byte][]byte
	if len(param) > 0 {
		lengthFieldIndex = param[0].(int)
		refField = param[1].(int)
		conditionMatrix = param[2].(map[byte][]byte)
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("New field condition lengthFieldIndex=%d refField=%d", lengthFieldIndex, refField)
	}
	return FieldCondition{
		lengthFieldIndex: lengthFieldIndex,
		refField:         refField,
		conditionMatrix:  conditionMatrix,
	}
	// return FieldCondition{
	// 	lengthFieldIndex: index,
	// 	refField:         ref,
	// 	conditionMatrix:  condition,
	// }
}

// NewType Define new type with length equal 1
func NewType(param ...interface{}) *AdaType {
	fType := param[0].(FieldType)
	name := param[1].(string)
	longName := name
	length := uint32(1)
	if len(param) > 2 {
		for i := 2; i < len(param); i++ {
			switch p := param[i].(type) {
			case int:
				length = uint32(p)
			case uint32:
				length = p
			case string:
				longName = p
			default:
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Unknown parameter type %T", param[2])
				}
				//panic("Error for type parameter")
				return nil
			}
		}
	}
	flags := FlagOptionToBeRemoved.Bit()
	if len(param) > 3 {
		flags = flags | FlagOptionLengthNotIncluded.Bit()
	}
	switch fType {
	case FieldTypeUByte, FieldTypeByte:
	case FieldTypeUInt2, FieldTypeInt2:
		length = 2
	case FieldTypeUInt4, FieldTypeInt4:
		length = 4
	case FieldTypeUInt8, FieldTypeInt8:
		length = 8
	}
	return &AdaType{CommonType: CommonType{
		fieldType: fType,
		level:     1,
		name:      longName,
		flags:     flags,
		shortName: name,
		peRange:   *NewEmptyRange(),
		muRange:   *NewEmptyRange(),
		length:    length,
	}}
}

// NewLongNameType Define new type with length equal 1
func NewLongNameType(fType FieldType, name string, shortName string) *AdaType {
	return NewType(fType, shortName, name)
}

// NewTypeWithLength Definen new type
func NewTypeWithLength(fType FieldType, name string, length uint32) *AdaType {
	return NewType(fType, name, length)
}

// NewTypeWithFlag Define new type with flag
func NewTypeWithFlag(fType FieldType, name string, flag FlagOption) *AdaType {
	t := NewType(fType, name, 0)
	t.AddFlag(flag)
	return t
}

// NewLongNameTypeWithLength Define new type
func NewLongNameTypeWithLength(fType FieldType, name string, shortName string, length uint32) *AdaType {
	return NewType(fType, shortName, name, length)
}

// String return the name of the field
func (adaType *AdaType) String() string {
	var b strings.Builder
	b.Grow(60)
	b.WriteString(strings.Repeat(" ", int(adaType.level)))
	options := adaType.Option()
	if options != "" {
		options = "," + strings.Replace(options, " ", ",", -1)
	}
	fmt.Fprintf(&b, "%d, %s, %d, %s %s ; %s", adaType.level, adaType.shortName, adaType.length,
		adaType.fieldType.FormatCharacter(), options, adaType.name)
	return b.String()
}

// Length return the length of the field
func (adaType *AdaType) Length() uint32 {
	return adaType.length
}

// SetLength set the length of the field
func (adaType *AdaType) SetLength(length uint32) {
	if adaType.length == length {
		return
	}
	if (adaType.fieldType != FieldTypeFloat && adaType.fieldType != FieldTypeDouble) || length > 0 {
		if adaType.HasFlagSet(FlagOptionPE) {
			// Period length change, CANNNOT use collected FB entry!!!!
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Length not default and field %s is PE field, need atomic FB", adaType.shortName)
			}
			adaType.AddFlag(FlagOptionAtomicFB)
		}
		adaType.length = length
	} else {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Skip float or double to %d", adaType.length)
		}
	}
}

// IsStructure return if it is an structure
func (adaType *AdaType) IsStructure() bool {
	return false
}

// SetFractional set fractional part
func (adaType *AdaType) SetFractional(x uint32) {
	adaType.FractValue = x
}

// Fractional get fractional part
func (adaType *AdaType) Fractional() uint32 {
	return adaType.FractValue
}

// // SetCharset set fractional part
// func (adaType *AdaType) SetCharset(x string) {
// 	adaType.Charset = x
// }

// SetFormatType set format type
func (adaType *AdaType) SetFormatType(x rune) {
	adaType.FormatTypeCharacter = x
}

// FormatType get format type
func (adaType *AdaType) FormatType() rune {
	return adaType.FormatTypeCharacter
}

// SetFormatLength set format length
func (adaType *AdaType) SetFormatLength(x uint32) {
	adaType.FormatLength = x
}

// Value return type specific value structure object
func (adaType *AdaType) Value() (adaValue IAdaValue, err error) {
	Central.Log.Debugf("Create field type of %v", adaType.fieldType)
	switch adaType.fieldType {
	case FieldTypeByte:
		adaValue = newByteValue(adaType)
		return
	case FieldTypeByteArray:
		if adaType.length > 126 {
			return nil, NewGenericError(111, adaType.length, "binary", adaType.Name())
		}
		adaValue = newByteArrayValue(adaType)
		return
	case FieldTypeFieldLength:
		adaValue = newLengthValue(adaType)
		return
	case FieldTypeLength, FieldTypeUByte, FieldTypeCharacter:
		adaValue = newUByteValue(adaType)
		return
	case FieldTypeString:
		if adaType.length > 253 {
			return nil, NewGenericError(111, adaType.length, "alpha", adaType.Name())
		}
		adaValue = newStringValue(adaType)
		return
	case FieldTypeLAString:
		if adaType.length > 65533 {
			return nil, NewGenericError(111, adaType.length, "large alpha", adaType.Name())
		}
		adaValue = newStringValue(adaType)
		return
	case FieldTypeLBString:
		if adaType.length > 2147483543 {
			return nil, NewGenericError(111, adaType.length, "large object alpha", adaType.Name())
		}
		adaValue = newStringValue(adaType)
		return
	case FieldTypeUnicode:
		if adaType.length > 253 {
			return nil, NewGenericError(111, adaType.length, "unicode", adaType.Name())
		}
		adaValue = newUnicodeValue(adaType)
		return
	case FieldTypeLAUnicode:
		if adaType.length > 16381 {
			return nil, NewGenericError(111, adaType.length, "large unicode", adaType.Name())
		}
		adaValue = newUnicodeValue(adaType)
		return
	case FieldTypeLBUnicode:
		if adaType.length > 16381 {
			return nil, NewGenericError(111, adaType.length, "large object unicode", adaType.Name())
		}
		adaValue = newUnicodeValue(adaType)
		return
	case FieldTypeUInt2:
		adaValue = newUInt2Value(adaType)
		return
	case FieldTypeInt2:
		adaValue = newInt2Value(adaType)
		return
	case FieldTypeUInt4:
		adaValue = newUInt4Value(adaType)
		return
	case FieldTypeInt4:
		adaValue = newInt4Value(adaType)
		return
	case FieldTypeUInt8:
		adaValue = newUInt8Value(adaType)
		return
	case FieldTypeInt8:
		adaValue = newInt8Value(adaType)
		return
	case FieldTypeUnpacked:
		if adaType.length > 29 {
			return nil, NewGenericError(111, adaType.length, "unpacked", adaType.Name())
		}
		adaValue = newUnpackedValue(adaType)
	case FieldTypePacked:
		if adaType.length > 15 {
			return nil, NewGenericError(111, adaType.length, "packed", adaType.Name())
		}
		adaValue = newPackedValue(adaType)
	case FieldTypeFloat:
		switch adaType.length {
		case (4):
			adaValue = newFloatValue(adaType)
		case (8):
			adaValue = newDoubleValue(adaType)
		default:
			err = NewGenericError(110, adaType.length, adaType.Name())
		}
	case FieldTypeFiller:
		adaValue = newFillerValue(adaType)
	case FieldTypePhonetic:
		adaValue = newPhoneticValue(adaType)
	case FieldTypeSuperDesc:
		adaValue = newSuperDescriptorValue(adaType)
	// Should not come here for structure types
	//	case FieldTypeStructure,FieldTypeGroup:
	//		Central.Log.Debugf("Return Structure value")
	//		return 0, newStructure(adaType)
	case FieldTypeCollation:
		adaValue = newCollationValue(adaType)
	case FieldTypeReferential:
		adaValue = newReferentialValue(adaType)
	default:
		Central.Log.Debugf("Return error type value evaluation for %v %s", adaType.fieldType, adaType.String())
		return nil, NewGenericError(102, adaType.fieldType.name(), adaType.Name())
	}
	return
}

// Option output all options of a field in an string
func (adaType *AdaType) Option() string {
	var buffer bytes.Buffer
	for i := 0; i < len(fieldOptions); i++ {
		if (adaType.options & (1 << uint(i))) > 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(" ")
			}
			buffer.WriteString(fieldOptions[i])
		}
	}
	switch {
	case adaType.Type() == FieldTypeLBString:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("LB")
	case adaType.Type() == FieldTypeLAString:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("LA")
	default:
	}
	if adaType.fieldType == FieldTypeMultiplefield {
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("MU")
	}
	switch adaType.SysField {
	case 1:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("SY=TIME")
	case 2:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("SY=SESSIONID")
	case 3:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("SY=OPUSER")
	default:
	}
	switch adaType.EditMask {
	case 1:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(DATE)")
	case 2:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(TIME)")
	case 3:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(DATETIME)")
	case 4:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(TIMESTAMP)")
	case 5:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(NATDATE)")
	case 6:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(NATTIME)")
	case 7:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(UNIXTIME)")
	case 8:
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("DT=E(XTIMESTAMP)")
	default:
	}
	return buffer.String()
}

// RedefinitionType creates a new redefinition type
type RedefinitionType struct {
	CommonType
	MainType IAdaType
	SubTypes []IAdaType
	// fieldMap map[string]IAdaType
}

// NewRedefinitionType Creates a new object of redefootopm types
func NewRedefinitionType(mainType IAdaType) *RedefinitionType {
	if mainType == nil {
		//panic("Main type of redefinition nil")
		return nil
	}
	return &RedefinitionType{MainType: mainType,
		CommonType: CommonType{level: mainType.Level(), name: mainType.Name(),
			FormatTypeCharacter: mainType.FormatType(), shortName: mainType.ShortName(),
			length: mainType.Length(), fieldType: FieldTypeRedefinition}}
}

// AddSubType add redefinition sub types used for the field
func (adaType *RedefinitionType) AddSubType(subType IAdaType) {
	subType.SetLevel(adaType.MainType.Level() + 1)
	adaType.SubTypes = append(adaType.SubTypes, subType)
}

// Value return type specific value structure object
func (adaType *RedefinitionType) Value() (adaValue IAdaValue, err error) {
	return newRedefinition(adaType), nil

}

// SetFormatType set format type
func (adaType *RedefinitionType) SetFormatType(x rune) {
}

// FormatType get format type
func (adaType *RedefinitionType) FormatType() rune {
	Central.Log.Debugf("Redefinition format type %c", adaType.MainType.FormatType())
	return adaType.MainType.FormatType()
}

// SetFractional set fractional part
func (adaType *RedefinitionType) SetFractional(x uint32) {
}

// Fractional get fractional part
func (adaType *RedefinitionType) Fractional() uint32 {
	return 0
}

// // SetCharset set fractional part
// func (adaType *RedefinitionType) SetCharset(x string) {
// }

// SetFormatLength set format length
func (adaType *RedefinitionType) SetFormatLength(x uint32) {
}

// Length return the length of the field
func (adaType *RedefinitionType) Length() uint32 {
	return adaType.MainType.Length()
}

// IsStructure return if it is an structure
func (adaType *RedefinitionType) IsStructure() bool {
	return true
}

// Option output all options of a field in an string
func (adaType *RedefinitionType) Option() string {
	return ""
}

// SetLength set the length of the field
func (adaType *RedefinitionType) SetLength(length uint32) {
	adaType.MainType.SetLength(length)
	adaType.CommonType.length = length
}

// String return the name of the field
func (adaType *RedefinitionType) String() string {
	if adaType.MainType == nil {
		return "Main type in redefinition nil"
	}
	return adaType.MainType.String()
}

// Traverse Traverse through the definition tree calling a callback method for each node
func (adaType *RedefinitionType) Traverse(t TraverserMethods, level int, x interface{}) (err error) {
	return nil
}
