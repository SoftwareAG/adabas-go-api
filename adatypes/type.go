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

package adatypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// OccByte Occurence identifier indicating that the occurence is defined as byte
const OccByte = -12

// OccUInt2 Occurence identifier indicating that the occurence is defined as uint32
const OccUInt2 = -11

// OccNone Occurence identifier indicating that the occurence is not used
const OccNone = -10

// OccSingle Occurence identifier indicating that the occurence single
const OccSingle = -9

// OccCapacity Occurence identifier indicating that the occurence capactity of MU or PE fields
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
}

// AdaType data type structure for field types, no structures
type AdaType struct {
	CommonType
	SysField  byte
	EditMask  byte
	SubOption byte
}

// FieldCondition field condition reference using for parser length management
type FieldCondition struct {
	lengthFieldIndex int
	refField         int
	conditionMatrix  map[byte][]byte
}

// NewFieldCondition creates a new field condition
func NewFieldCondition(index int, ref int, condition map[byte][]byte) FieldCondition {
	Central.Log.Debugf("New field condition lengthFieldIndex=%d refField=%d", index, ref)
	return FieldCondition{
		lengthFieldIndex: index,
		refField:         ref,
		conditionMatrix:  condition,
	}
}

// StructureType creates a new structure type
type StructureType struct {
	CommonType
	//	fieldType FieldType
	//	name      string
	//	length    uint32
	occ       int
	condition FieldCondition
	SubTypes  []IAdaType
}

// NewType Define new type with length equal 1
func NewType(fType FieldType, name string) *AdaType {
	length := uint32(1)
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
		name:      name,
		flags:     uint8(1 << FlagOptionToBeRemoved),
		shortName: name,
		length:    length,
	}}
}

// NewLongNameType Define new type with length equal 1
func NewLongNameType(fType FieldType, name string, shortName string) *AdaType {
	t := NewType(fType, name)
	t.shortName = shortName
	return t
}

// NewTypeWithLength Definen new type
func NewTypeWithLength(fType FieldType, name string, length uint32) *AdaType {
	return &AdaType{CommonType: CommonType{
		fieldType: fType,
		level:     1,
		name:      name,
		flags:     uint8(1 << FlagOptionToBeRemoved),
		shortName: name,
		length:    length,
	}}
}

// NewLongNameTypeWithLength Definen new type
func NewLongNameTypeWithLength(fType FieldType, name string, shortName string, length uint32) *AdaType {
	t := NewTypeWithLength(fType, name, length)
	t.shortName = shortName
	return t
}

func (commonType *CommonType) flagString() string {
	flags := fmt.Sprintf(" PE=%v MU=%v REMOVE=%v", commonType.HasFlagSet(FlagOptionPE),
		commonType.HasFlagSet(FlagOptionMU), commonType.HasFlagSet(FlagOptionToBeRemoved))
	return flags
}

// String return the name of the field
func (adaType *AdaType) String() string {
	y := strings.Repeat(" ", int(adaType.level))
	options := adaType.Option()
	if options != "" {
		options = "," + options
	}
	return fmt.Sprintf("%s%d, %s, %d, %s %s ; %s %s", y, adaType.level, adaType.shortName, adaType.length,
		adaType.fieldType.FormatCharacter(), options, adaType.name, adaType.flagString())
	// if adaType.shortName == adaType.name {
	// 	return fmt.Sprintf("%s%d %s %d %s %s -> %s", y, adaType.level, adaType.shortName, adaType.length,
	// 		adaType.fieldType.FormatCharacter(), adaType.Option(), adaType.flagString())
	// }
	// return fmt.Sprintf("%s%d %s %s %d %s %s -> %s", y, adaType.level, adaType.name, adaType.shortName, adaType.length,
	// 	adaType.fieldType.FormatCharacter(), adaType.Option(), adaType.flagString())
}

// Length return the length of the field
func (adaType *AdaType) Length() uint32 {
	return adaType.length
}

// SetLength set the length of the field
func (adaType *AdaType) SetLength(length uint32) {
	adaType.length = length
}

// IsStructure return if it is an structure
func (adaType *AdaType) IsStructure() bool {
	return false
}

// Value return type specific value structure object
func (adaType *AdaType) Value() (adaValue IAdaValue, err error) {
	Central.Log.Debugf("Create field type of %v", adaType.fieldType)
	switch adaType.fieldType {
	case FieldTypeByte:
		Central.Log.Debugf("Return byte value")
		adaValue = newByteValue(adaType)
		return
	case FieldTypeByteArray:
		Central.Log.Debugf("Return byte array value")
		adaValue = newByteArrayValue(adaType)
		return
	case FieldTypeLength, FieldTypeUByte, FieldTypeCharacter:
		Central.Log.Debugf("Return byte value")
		adaValue = newUByteValue(adaType)
		return
	case FieldTypeString, FieldTypeLAString, FieldTypeLBString:
		Central.Log.Debugf("Return string value")
		adaValue = newStringValue(adaType)
		return
	case FieldTypeUnicode, FieldTypeLAUnicode, FieldTypeLBUnicode:
		Central.Log.Debugf("Return unicode value")
		adaValue = newUnicodeValue(adaType)
		return
	case FieldTypeUInt2:
		Central.Log.Debugf("Return UInt2 value")
		adaValue = newUInt2Value(adaType)
		return
	case FieldTypeInt2:
		Central.Log.Debugf("Return Int2 value")
		adaValue = newInt2Value(adaType)
		return
	case FieldTypeUInt4:
		Central.Log.Debugf("Return UInt4 value")
		adaValue = newUInt4Value(adaType)
		return
	case FieldTypeInt4:
		Central.Log.Debugf("Return Int4 value")
		adaValue = newInt4Value(adaType)
		return
	case FieldTypeUInt8:
		Central.Log.Debugf("Return UInt8 value")
		adaValue = newUInt8Value(adaType)
		return
	case FieldTypeInt8:
		Central.Log.Debugf("Return Int8 value")
		adaValue = newInt8Value(adaType)
		return
	case FieldTypeUnpacked:
		Central.Log.Debugf("Return Unpacked value")
		adaValue = newUnpackedValue(adaType)
	case FieldTypePacked:
		Central.Log.Debugf("Return Packed value")
		adaValue = newPackedValue(adaType)
	case FieldTypeFloat:
		Central.Log.Debugf("Return Float value")
		switch adaType.length {
		case (4):
			adaValue = newFloatValue(adaType)
		case (8):
			adaValue = newDoubleValue(adaType)
		default:
			err = NewGenericError(110, adaType.length, adaType.String())
		}
	case FieldTypeFiller:
		Central.Log.Debugf("Return filler value")
		adaValue = newFillerValue(adaType)
	case FieldTypePhonetic:
		adaValue = newPhoneticValue(adaType)
	// Should not come here for structure types
	//	case FieldTypeStructure,FieldTypeGroup:
	//		Central.Log.Debugf("Return Structure value")
	//		return 0, newStructure(adaType)
	case FieldTypeCollation:
		adaValue = newCollationValue(adaType)
	case FieldTypeReferential:
		adaValue = newReferentialValue(adaType)
	default:
		Central.Log.Debugf("Return nil value %v %s", adaType.fieldType, adaType.String())
		return nil, NewGenericError(102, adaType.fieldType.name(), adaType.Name())
	}
	return
}

// NewStructure Creates a new object of structured list types
func NewStructure() *StructureType {
	Central.Log.Debugf("Create structure list")
	return &StructureType{
		CommonType: CommonType{
			flags: uint8(1 << FlagOptionToBeRemoved),
		},
		condition: FieldCondition{
			lengthFieldIndex: -1,
			refField:         NoReferenceField,
		},
	}
}

// NewStructureEmpty Creates a new object of structured list types
func NewStructureEmpty(fType FieldType, name string, occByteShort int16,
	level uint8) *StructureType {
	Central.Log.Debugf("Create empty structure list %s with type %d ", name, fType)
	st := &StructureType{
		CommonType: CommonType{
			fieldType: fType,
			name:      name,
			flags:     uint8(1 << FlagOptionToBeRemoved),
			shortName: name,
			length:    0,
			level:     level,
		},
		occ: int(occByteShort),
		condition: FieldCondition{
			lengthFieldIndex: -1,
			refField:         NoReferenceField,
		},
	}
	st.adaptSubFields()
	return st
}

// NewStructureList Creates a new object of structured list types
func NewStructureList(fType FieldType, name string, occByteShort int16, subFields []IAdaType) *StructureType {
	Central.Log.Debugf("Create new structure list %s types=%d type=%d", name, len(subFields), fType)
	st := &StructureType{
		CommonType: CommonType{fieldType: fType,
			name:      name,
			shortName: name,
			flags:     uint8(1 << FlagOptionToBeRemoved),
			level:     1,
			length:    0},
		occ: int(occByteShort),
		condition: FieldCondition{
			lengthFieldIndex: -1,
			refField:         NoReferenceField,
		},
		SubTypes: subFields,
	}
	st.adaptSubFields()

	return st
}

// NewLongNameStructureList Creates a new object of structured list types
func NewLongNameStructureList(fType FieldType, name string, shortName string, occByteShort int16, subFields []IAdaType) *StructureType {
	st := NewStructureList(fType, name, occByteShort, subFields)
	st.shortName = shortName
	return st
}

// NewStructureCondition Creates a new object of structured list types
func NewStructureCondition(fType FieldType, name string, subFields []IAdaType, condition FieldCondition) *StructureType {
	Central.Log.Debugf("Create new structure with condition %s types=%d type=%d", name, len(subFields), fType)
	for _, t := range subFields {
		t.SetLevel(2)
	}
	return &StructureType{
		CommonType: CommonType{fieldType: fType,
			name:      name,
			shortName: name,
			flags:     uint8(1 << FlagOptionToBeRemoved),
			level:     1,
			length:    0},
		condition: condition,
		SubTypes:  subFields,
	}
}

func (adaType *StructureType) adaptSubFields() {
	if adaType.Type() == FieldTypePeriodGroup {
		Central.Log.Debugf("%s: set PE flag", adaType.Name())
		adaType.AddFlag(FlagOptionPE)
		adaType.occ = OccCapacity
	}
	if adaType.Type() == FieldTypeMultiplefield {
		Central.Log.Debugf("%s: set MU flag", adaType.Name())
		adaType.AddFlag(FlagOptionMU)
		adaType.occ = OccCapacity
	}
	for _, s := range adaType.SubTypes {
		s.SetParent(adaType)
		if adaType.Type() == FieldTypePeriodGroup {
			s.AddFlag(FlagOptionPE)
		}
		if adaType.Type() == FieldTypeMultiplefield {
			Central.Log.Debugf("%s: set MU flag", adaType.Name())
			adaType.AddFlag(FlagOptionMU)
			s.AddFlag(FlagOptionMUGhost)
		}

	}
}

// String return the name of the field
func (adaType *StructureType) String() string {

	y := strings.Repeat(" ", int(adaType.level))
	Central.Log.Debugf("FS: %s -> %d", adaType.Name(), len(adaType.SubTypes))
	if adaType.fieldType == FieldTypeMultiplefield {
		if len(adaType.SubTypes) == 0 {
			return fmt.Sprintf("%s%d %s deleted", y, adaType.level, adaType.shortName)
		}
		return fmt.Sprintf("%s%d, %s, %d, %s %s,MU; %s %s", y, adaType.level, adaType.shortName, adaType.SubTypes[0].Length(),
			adaType.SubTypes[0].Type().FormatCharacter(), adaType.SubTypes[0].Option(), adaType.name, adaType.flagString())

	}
	options := adaType.Option()
	if options != "" {
		options = "," + options
	}
	return fmt.Sprintf("%s%d, %s %s ; %s %s", y, adaType.level, adaType.shortName, options,
		adaType.name, adaType.flagString())
}

// Length returns the length of the field
func (adaType *StructureType) Length() uint32 {
	return adaType.length
}

// SetLength set the length of the field
func (adaType *StructureType) SetLength(length uint32) {
	adaType.length = length
}

// IsStructure return the structure of the field
func (adaType *StructureType) IsStructure() bool {
	return true
}

// NrFields number of fields contained in the structure
func (adaType *StructureType) NrFields() int {
	return len(adaType.SubTypes)
}

func (adaType *StructureType) parseBuffer(helper *BufferHelper, option *BufferOption) {
	Central.Log.Debugf("Parse Structure type offset=%d", helper.offset)
}

// Traverse Traverse through the definition tree calling a callback method for each node
func (adaType *StructureType) Traverse(t TraverserMethods, level int, x interface{}) (err error) {
	Central.Log.Debugf("Current structure -> %s", adaType.name)
	Central.Log.Debugf("Nr of structure types -> %v", len(adaType.SubTypes))
	for _, v := range adaType.SubTypes {
		Central.Log.Debugf("Traverse on %s", v.Name())
		err = t.EnterFunction(v, adaType, level, x)
		if err != nil {
			return
		}
		if v.IsStructure() {
			Central.Log.Debugf("Traverse into structure %s", v.Name())
			err = v.(*StructureType).Traverse(t, level+1, x)
			if err != nil {
				return
			}
			if t.leaveFunction != nil {
				err = t.leaveFunction(v, adaType, level, x)
				if err != nil {
					return
				}
			}
		}
	}
	return nil
}

// AddField add a new field type into the structure type
func (adaType *StructureType) AddField(fieldType IAdaType) {
	Central.Log.Debugf("Add sub field %s on parent %s", fieldType.Name(), adaType.Name())
	fieldType.SetLevel(adaType.level + 1)
	fieldType.SetParent(adaType)
	Central.Log.Debugf("Parent of %s is %s ", fieldType.Name(), fieldType.GetParent())
	if adaType.HasFlagSet(FlagOptionPE) {
		Central.Log.Debugf("Add sub field PE of parent %s field %s", adaType.Name(), fieldType.Name())
		fieldType.AddFlag(FlagOptionPE)
	}
	adaType.SubTypes = append(adaType.SubTypes, fieldType)
}

// RemoveField remote field of the structure type
func (adaType *StructureType) RemoveField(fieldType *CommonType) {
	Central.Log.Debugf("Remove field %s out of %s nrFields=%d", fieldType.Name(), adaType.Name(), adaType.NrFields())
	// if adaType.NrFields() < 2 && adaType.GetParent() != nil {
	// 	Central.Log.Debugf("Only one left, remove last ", fieldType.Name())
	// 	commonType := &adaType.CommonType
	// 	adaType.GetParent().(*StructureType).RemoveField(commonType)
	// } else {
	Central.Log.Debugf("Rearrange, left=%d", adaType.NrFields())
	var newTypes []IAdaType
	for _, t := range adaType.SubTypes {
		if t.Name() != fieldType.Name() {
			newTypes = append(newTypes, t)
		}
	}
	adaType.SubTypes = newTypes
	// }
}

// Option return structure option as a string
func (adaType *StructureType) Option() string {
	switch adaType.fieldType {
	case FieldTypeMultiplefield:
		return "MU"
	case FieldTypePeriodGroup:
		return "PE"
	default:
	}
	return ""
}

// Value return type specific value structure object
func (adaType *StructureType) Value() (adaValue IAdaValue, err error) {
	Central.Log.Debugf("Create structure type of %v", adaType.fieldType.name())
	switch adaType.fieldType {
	case FieldTypeStructure, FieldTypeGroup, FieldTypePeriodGroup, FieldTypeMultiplefield:
		Central.Log.Debugf("Return Structure value")
		adaValue = newStructure(adaType)
		return
	}
	Central.Log.Debugf("Return nil structure", adaType.String())
	err = NewGenericError(104, adaType.String())
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

type subSuperEntries struct {
	Name [2]byte
	From uint16
	To   uint16
}

// AdaSuperType data type structure for super or sub descriptor field types, no structures
type AdaSuperType struct {
	CommonType
	FdtFormat byte
	Entries   []subSuperEntries
}

// NewSuperType new super or sub descriptor field type
func NewSuperType(name string, option byte) *AdaSuperType {

	superType := &AdaSuperType{CommonType: CommonType{fieldType: FieldTypeSuperDesc,
		flags: uint8(1 << FlagOptionToBeRemoved),
		name:  name, shortName: name}}
	if (option & 0x08) > 0 {
		Central.Log.Debugf("%s super/sub descriptor found PE", name)
		superType.AddOption(FieldOptionPE)
	}
	if (option & 0x20) > 0 {
		superType.AddOption(FieldOptionMU)
	}
	return superType
}

// IsStructure return the structure of the field
func (adaType *AdaSuperType) IsStructure() bool {
	return false
}

// AddSubEntry add sub field entry on super or sub descriptors
func (adaType *AdaSuperType) AddSubEntry(name string, from uint16, to uint16) {
	var code [2]byte
	copy(code[:], name)
	entry := subSuperEntries{Name: code, From: from, To: to}
	adaType.Entries = append(adaType.Entries, entry)
	adaType.calcLength()
}

func (adaType *AdaSuperType) calcLength() {
	len := uint32(0)
	for _, entry := range adaType.Entries {
		Central.Log.Debugf("%s: super descriptor entry %s len=%d add [%d:%d] -> %d", adaType.name, entry.Name,
			len, entry.From, entry.To, uint32(entry.To-entry.From+1))
		len += uint32(entry.To - entry.From + 1)
	}
	Central.Log.Debugf("len=%d", len)
	adaType.length = len
}

// Length return the length of the field
func (adaType *AdaSuperType) Length() uint32 {
	return adaType.length
}

// SetLength set the length of the field
func (adaType *AdaSuperType) SetLength(length uint32) {
}

// Option string representation of all option of Sub or super descriptors
func (adaType *AdaSuperType) Option() string {
	return ""
}

// String string representation of the sub or super descriptor
func (adaType *AdaSuperType) String() string {
	var buffer bytes.Buffer
	if adaType.shortName == adaType.name {
		buffer.WriteString(adaType.shortName + "=")
	} else {
		buffer.WriteString(adaType.name + "[" + adaType.shortName + "] =")
	}
	for index, s := range adaType.Entries {
		if index > 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(fmt.Sprintf("%s(%d-%d)", s.Name, s.From, s.To))
	}
	buffer.WriteString(fmt.Sprintf(" ; %s %s", adaType.name, adaType.flagString()))
	return buffer.String()
}

// Value value of the sub or super descriptor
func (adaType *AdaSuperType) Value() (adaValue IAdaValue, err error) {
	Central.Log.Debugf("Return super descriptor value")
	adaValue = newSuperDescriptorValue(adaType)
	return
}

// AdaPhoneticType data type phonetic descriptor for field types, no structures
type AdaPhoneticType struct {
	AdaType
	descriptorLength uint16
	parentName       [2]byte
}

// NewPhoneticType new phonetic descriptor type
func NewPhoneticType(name string, descriptorLength uint16, parentName string) *AdaPhoneticType {
	var code [2]byte
	copy(code[:], parentName)
	return &AdaPhoneticType{AdaType: AdaType{CommonType: CommonType{fieldType: FieldTypePhonetic, name: name,
		flags:     uint8(1 << FlagOptionToBeRemoved),
		shortName: name}},
		descriptorLength: descriptorLength, parentName: code}
}

// String string representation of the phonetic type
func (fieldType *AdaPhoneticType) String() string {
	return fmt.Sprintf("%s=PHON(%s) ; %s %s", fieldType.shortName, fieldType.parentName, fieldType.name, fieldType.flagString())
}

// AdaCollationType data type structure for field types, no structures
type AdaCollationType struct {
	AdaType
	length        uint16
	parentName    [2]byte
	collAttribute string
}

// NewCollationType creates new collation type instance
func NewCollationType(name string, length uint16, parentName string, collAttribute string) *AdaCollationType {
	var code [2]byte
	copy(code[:], parentName)
	return &AdaCollationType{AdaType: AdaType{CommonType: CommonType{fieldType: FieldTypeCollation,
		flags: uint8(1 << FlagOptionToBeRemoved),
		name:  name, shortName: name}}, length: length,
		parentName: code, collAttribute: collAttribute}
}

// String string representation of the collation type
func (fieldType *AdaCollationType) String() string {
	options := ""
	if fieldType.IsOption(FieldOptionLA) {
		options = ",LA"
	} else {
		if fieldType.IsOption(FieldOptionLB) {
			options = ",L4"
		}
	}
	if fieldType.IsOption(FieldOptionHE) {
		options = ",HE"
	}
	if fieldType.IsOption(FieldOptionUQ) {
		options = ",UQ"
	}
	return fmt.Sprintf("%s%s=COLLATING(%s,%s) ; %s %s", fieldType.shortName, options, fieldType.parentName,
		fieldType.collAttribute, fieldType.name, fieldType.flagString())
}

// AdaHyperExitType data type structure for field types, no structures
type AdaHyperExitType struct {
	AdaType
	fdtFormat   byte
	nr          byte
	parentNames []string
}

// NewHyperExitType new hyper exit type
func NewHyperExitType(name string, length uint32, fdtFormat byte, nr uint8, parentNames []string) *AdaHyperExitType {
	return &AdaHyperExitType{AdaType: AdaType{CommonType: CommonType{fieldType: FieldTypeHyperDesc,
		flags: uint8(1 << FlagOptionToBeRemoved),
		name:  name, shortName: name, length: length}},
		fdtFormat: fdtFormat, nr: nr, parentNames: parentNames}
}

// String string representation of the hyper exit type
func (fieldType *AdaHyperExitType) String() string {
	options := fieldType.Option()
	if len(options) > 0 {
		options = "," + options
	}
	parents := ""
	for _, p := range fieldType.parentNames {
		if len(parents) > 0 {
			parents += ","
		}
		parents += p
	}
	return fmt.Sprintf("%s %d %c%s=HYPER(%d,%s) ; %s %s", fieldType.shortName, fieldType.length, fieldType.fdtFormat,
		options, fieldType.nr, parents, fieldType.name, fieldType.flagString())
}

// AdaReferentialType data type structure for referential integrity types, no structures
type AdaReferentialType struct {
	AdaType
	refFile         uint32
	keys            [2]string
	refType         uint8
	refUpdateAction uint8
	refDeleteAction uint8
}

// NewReferentialType new referential integrity type
func NewReferentialType(name string, refFile uint32, keys [2]string, refType uint8, refUpdateAction uint8, refDeleteAction uint8) *AdaReferentialType {
	return &AdaReferentialType{AdaType: AdaType{CommonType: CommonType{fieldType: FieldTypeReferential,
		flags: uint8(1 << FlagOptionToBeRemoved),
		name:  name, shortName: name, length: 0}}, refFile: refFile, keys: keys, refType: refType,
		refUpdateAction: refUpdateAction, refDeleteAction: refDeleteAction}

}

// String string representation of the hyper exit type
func (fieldType *AdaReferentialType) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s=REFINT(%s,%d,%s", fieldType.shortName, fieldType.keys[1], fieldType.refFile, fieldType.keys[0]))
	switch fieldType.refDeleteAction {
	case 0:
		buffer.WriteString("/DX")
	case 1:
		buffer.WriteString("/DC")
	case 2:
		buffer.WriteString("/DN")
	}
	switch fieldType.refUpdateAction {
	case 0:
		buffer.WriteString(",UX")
	case 1:
		buffer.WriteString(",UC")
	case 2:
		buffer.WriteString(",UN")
	}
	buffer.WriteString(")")
	buffer.WriteString(fmt.Sprintf(" ; %s %s", fieldType.name, fieldType.flagString()))
	return buffer.String()
}
