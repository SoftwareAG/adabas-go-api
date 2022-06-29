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
	"encoding/binary"
)

// Version version of current build
var Version = "v1.6.19"

// FieldType indicate a field type of the field
type FieldType uint

const (
	// FieldTypeUndefined field type undefined
	FieldTypeUndefined FieldType = iota
	// FieldTypeUByte field type unsigned byte
	FieldTypeUByte
	// FieldTypeByte field type signed byte
	FieldTypeByte
	// FieldTypeUInt2 field type unsigned integer of 2 bytes
	FieldTypeUInt2
	// FieldTypeInt2 field type signed integer of 2 bytes
	FieldTypeInt2
	// FieldTypeShort field type signed short
	FieldTypeShort
	// FieldTypeUInt4 field type unsigned integer of 4 bytes
	FieldTypeUInt4
	// FieldTypeUInt4Array field type array unsigned integer of 4 bytes
	FieldTypeUInt4Array
	// FieldTypeInt4 field type signed integer of 4 bytes
	FieldTypeInt4
	// FieldTypeUInt8 field type unsigned integer of 8 bytes
	FieldTypeUInt8
	// FieldTypeInt8 field type signed integer of 8 bytes
	FieldTypeInt8
	// FieldTypeLong field type signed long
	FieldTypeLong
	// FieldTypePacked field type packed
	FieldTypePacked
	// FieldTypeUnpacked field type unpacked
	FieldTypeUnpacked
	// FieldTypeDouble field type double
	FieldTypeDouble
	// FieldTypeFloat field type float
	FieldTypeFloat
	// FieldTypeFiller field type for fill gaps between struct types
	FieldTypeFiller
	// FieldTypeString field type string
	FieldTypeString
	// FieldTypeByteArray field type byte array
	FieldTypeByteArray
	// FieldTypeCharacter field type character
	FieldTypeCharacter
	// FieldTypeLength field type for length definitions
	FieldTypeLength
	// FieldTypeUnicode field type unicode string
	FieldTypeUnicode
	// FieldTypeLAUnicode field type unicode large objects
	FieldTypeLAUnicode
	// FieldTypeLBUnicode field type unicode LOB
	FieldTypeLBUnicode
	// FieldTypeLAString field type string large objects
	FieldTypeLAString
	// FieldTypeLBString field type string LOB
	FieldTypeLBString
	// FieldTypeFieldLength field length
	FieldTypeFieldLength
	// FieldTypePeriodGroup field type period group
	FieldTypePeriodGroup
	// FieldTypeMultiplefield field type multiple fields
	FieldTypeMultiplefield
	// FieldTypeStructure field type of structured types
	FieldTypeStructure
	// FieldTypeGroup field type group
	FieldTypeGroup
	// FieldTypeRedefinition field type group
	FieldTypeRedefinition
	// FieldTypePackedArray field type packed array
	FieldTypePackedArray
	// FieldTypePhonetic field type of phonetic descriptor
	FieldTypePhonetic
	// FieldTypeSuperDesc field type of super descriptors
	FieldTypeSuperDesc
	// FieldTypeLiteral field type of literal data send to database
	FieldTypeLiteral
	// FieldTypeFieldCount field type to defined field count of MU or PE fields
	FieldTypeFieldCount
	// FieldTypeHyperDesc field type of Hyper descriptors
	FieldTypeHyperDesc
	// FieldTypeReferential field type for referential integrity
	FieldTypeReferential
	// FieldTypeCollation field type of collation descriptors
	FieldTypeCollation
	// FieldTypeFunction field type to define functions working on result list
	FieldTypeFunction
)

var typeName = []string{"Undefined", "UByte", "Byte", "UInt2", "Int2", "Short", "UInt4", "UInt4Array", "Int4", "UInt8", "Int8",
	"Long", "Packed", "Unpacked", "Double", "Float", "Filler", "String", "ByteArray", "Character", "Length",
	"Unicode", "LAUnicode", "LBUnicode", "LAString", "LBString", "FieldLength", "PeriodGroup", "Multiplefield",
	"Structure", "Group", "PackedArray", "Phonetic", "SuperDesc", "Literal", "FieldCount", "HyperDesc",
	"Referential", "Collation", "Function"}

func (fieldType FieldType) name() string {
	return typeName[fieldType]
}

// FormatCharacter format character use to output FDT
func (fieldType FieldType) FormatCharacter() string {
	switch fieldType {
	case FieldTypeCharacter, FieldTypeString, FieldTypeLAString, FieldTypeLBString:
		return "A"
	case FieldTypeUnicode, FieldTypeLAUnicode, FieldTypeLBUnicode:
		return "W"
	case FieldTypeUByte, FieldTypeUInt2, FieldTypeUInt4, FieldTypeUInt8, FieldTypeShort, FieldTypeByteArray:
		return "B"
	case FieldTypePacked:
		return "P"
	case FieldTypeUnpacked:
		return "U"
	case FieldTypeByte, FieldTypeInt2, FieldTypeInt4, FieldTypeInt8:
		return "F"
	case FieldTypeFloat:
		return "G"
	default:
	}
	return " "
}

// EvaluateFieldType evaluate field type of format string
func EvaluateFieldType(fieldType rune, length int32) FieldType {
	switch fieldType {
	case 'A':
		if length == 1 {
			return FieldTypeByte
		}
		return FieldTypeString
	case 'P':
		return FieldTypePacked
	case 'U':
		return FieldTypeUnpacked
	case 'G':
		return FieldTypeFloat
	case 'B':
		switch length {
		case 1:
			return FieldTypeUByte
		case 2:
			return FieldTypeUInt2
		case 4:
			return FieldTypeUInt4
		case 8:
			return FieldTypeUInt8
		}
		return FieldTypeByteArray
	case 'F':
		switch length {
		case 1:
			return FieldTypeByte
		case 2:
			return FieldTypeInt2
		case 4:
			return FieldTypeInt4
		case 8:
			return FieldTypeInt8
		}
		return FieldTypeByteArray
	default:
	}
	return FieldTypeUndefined
}

// CommonType common data type structure defined for all types
type CommonType struct {
	fieldType           FieldType
	name                string
	shortName           string
	length              uint32
	level               uint8
	flags               uint32
	parentType          IAdaType
	options             uint32
	Charset             string
	endian              binary.ByteOrder
	peRange             AdaRange
	muRange             AdaRange
	partialRange        *AdaRange
	FormatTypeCharacter rune
	FormatLength        uint32
	SubTypes            []IAdaType
	convert             ConvertUnicode
}

// Type returns field type of the field
func (commonType *CommonType) Type() FieldType {
	return commonType.fieldType
}

// Name return the name of the field
func (commonType *CommonType) Name() string {
	return commonType.name
}

// ShortName return the short name of the field
func (commonType *CommonType) ShortName() string {
	return commonType.shortName
}

// SetName set the name of the field
func (commonType *CommonType) SetName(name string) {
	commonType.name = name
}

// Level Type return level of the field
func (commonType *CommonType) Level() uint8 {
	return commonType.level
}

// SetLevel Set Adabas level of the field
func (commonType *CommonType) SetLevel(level uint8) {
	commonType.level = level
}

// Endian Get data endian
func (commonType *CommonType) Endian() binary.ByteOrder {
	if commonType.endian == nil {
		commonType.endian = endian()
	}
	return commonType.endian
}

// SetEndian Set data endian
func (commonType *CommonType) SetEndian(endian binary.ByteOrder) {
	commonType.endian = endian
}

// SetRange set Adabas range
func (commonType *CommonType) SetRange(r *AdaRange) {
	commonType.peRange = *r
}

// SetParent set the parent of the type
func (commonType *CommonType) SetParent(parentType IAdaType) {
	if parentType != nil {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("%s parent is set to %s", commonType.name, parentType.Name())
		}
		if parentType.HasFlagSet(FlagOptionPE) {
			commonType.AddFlag(FlagOptionPE)
		}
		if commonType.HasFlagSet(FlagOptionAtomicFB) {
			p := parentType
			for p != nil {
				if p.GetParent() != nil {
					p.AddFlag(FlagOptionAtomicFB)
				}
				p = p.GetParent()
			}
		}
	} else {
		if commonType.parentType != nil {
			pType := commonType.parentType.(*StructureType)
			pType.RemoveField(commonType)
		}
	}
	commonType.parentType = parentType
}

// GetParent get the parent defined to this type
func (commonType *CommonType) GetParent() IAdaType {
	return commonType.parentType
}

// IsStructure return if the type is of structure types
func (commonType *CommonType) IsStructure() bool {
	return false
}

// AddOption add the option to the field
func (commonType *CommonType) AddOption(fieldOption FieldOption) {
	commonType.options |= (1 << fieldOption)
}

// ClearOption clear the option to the field
func (commonType *CommonType) ClearOption(fieldOption FieldOption) {
	commonType.options &^= (1 << fieldOption)
}

// IsOption Check if the option of the field is set
func (commonType *CommonType) IsOption(fieldOption FieldOption) bool {
	return (commonType.options & (1 << fieldOption)) != 0
}

// SetOption Set all options of the field
func (commonType *CommonType) SetOption(option uint32) {
	commonType.options = option
}

// IsSpecialDescriptor return true if it is a special descriptor
func (commonType *CommonType) IsSpecialDescriptor() bool {
	switch commonType.fieldType {
	case FieldTypeCollation, FieldTypePhonetic, FieldTypeSuperDesc,
		FieldTypeHyperDesc, FieldTypeReferential:
		return true
	default:

	}
	return false
}

// FieldOption type for field option
type FieldOption uint32

const (
	// FieldOptionUQ field option for unique descriptors
	FieldOptionUQ FieldOption = iota
	// FieldOptionNU field option for null suppression
	FieldOptionNU
	// FieldOptionFI field option for fixed size
	FieldOptionFI
	// FieldOptionDE field option for descriptors
	FieldOptionDE
	// FieldOptionNC field option for sql
	FieldOptionNC
	// FieldOptionNN field option for non null
	FieldOptionNN
	// FieldOptionHF field option for high order fields
	FieldOptionHF
	// FieldOptionNV field option for null value
	FieldOptionNV
	// FieldOptionNB field option for
	FieldOptionNB
	// FieldOptionHE field option for
	FieldOptionHE
	// FieldOptionPE field option for period
	FieldOptionPE
	// FieldOptionMU field option for multiple fields
	FieldOptionMU
	// FieldOptionLA field option for large alpha
	FieldOptionLA
	// FieldOptionLB field option for large objects
	FieldOptionLB
	// FieldOptionColExit field option for collation exit
	FieldOptionColExit
)

var fieldOptions = []string{"UQ", "NU", "FI", "DE", "NC", "NN", "HF", "NV", "NB", "HE", "PE", "MU"}

// FlagOption flag option used to omit traversal through the tree (example is MU and PE)
type FlagOption uint32

const (
	// FlagOptionPE indicate tree is part of period group
	FlagOptionPE FlagOption = iota
	// FlagOptionAtomicFB indicate tree contains MU fields
	FlagOptionAtomicFB
	// FlagOptionMUGhost ghost field for MU
	FlagOptionMUGhost
	// FlagOptionToBeRemoved should be removed
	FlagOptionToBeRemoved
	// FlagOptionSecondCall Field will need a second call to get the value
	FlagOptionSecondCall
	// FlagOptionReference Field will skip parsing value
	FlagOptionReference
	// FlagOptionReadOnly read only field
	FlagOptionReadOnly
	// FlagOptionLengthNotIncluded length not include in record buffer
	FlagOptionLengthNotIncluded
	// FlagOptionPart structure is request only in parts
	FlagOptionPart
	// FlagOptionSingleIndex single index query
	FlagOptionSingleIndex
	// FlagOptionLengthPE instead of length use period group count
	FlagOptionLengthPE
)

// Bit return the Bit of the option flag
func (flagOption FlagOption) Bit() uint32 {
	return (1 << flagOption)
}

// HasFlagSet check if given flag is set
func (commonType *CommonType) HasFlagSet(flagOption FlagOption) bool {
	//Central.Log.Debugf("Check flag %d set %d=%d -> %v", commonType.flags, flagOption.Bit(), flagOption.Bit(), (commonType.flags & flagOption.Bit()))
	return (commonType.flags & flagOption.Bit()) != 0
}

// AddFlag add the flag to the type flag set
func (commonType *CommonType) AddFlag(flagOption FlagOption) {
	if commonType.HasFlagSet(flagOption) {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Flag %s to %d already done", commonType.shortName, flagOption.Bit())
		}
		return
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set Flag %s to %d", commonType.shortName, flagOption.Bit())
	}
	commonType.flags |= flagOption.Bit()

	if flagOption == FlagOptionAtomicFB || flagOption == FlagOptionSingleIndex {
		p := commonType.GetParent()
		for p != nil && p.ShortName() != "" {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Set Parent Flag %s to %d", p.ShortName(), flagOption.Bit())
			}
			if p.HasFlagSet(flagOption) {
				break
			}
			p.AddFlag(flagOption)
			p = p.GetParent()
		}
		if flagOption == FlagOptionAtomicFB {
			// Only work in period group or group
			//			if !p.HasFlagSet(flagOption) {
			for _, s := range commonType.SubTypes {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Set Children Flag %s to %d", s.ShortName(), flagOption.Bit())
				}
				if !s.HasFlagSet(flagOption) {
					s.AddFlag(flagOption)
				}
			}
			//			}
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Set Flag %s to %d done", commonType.shortName, flagOption.Bit())
	}
}

// RemoveFlag add the flag to the type flag set
func (commonType *CommonType) RemoveFlag(flagOption FlagOption) {
	commonType.flags &= ^flagOption.Bit()
}

// SetPartialRange set partial range
func (commonType *CommonType) SetPartialRange(partial *AdaRange) {
	commonType.partialRange = partial
}

// PartialRange partial range provided
func (commonType *CommonType) PartialRange() *AdaRange {
	return commonType.partialRange
}

// PeriodicRange range of PE field provided
func (commonType *CommonType) PeriodicRange() *AdaRange {
	return &commonType.peRange
}

// MultipleRange range of MU field provided
func (commonType *CommonType) MultipleRange() *AdaRange {
	return &commonType.muRange
}

// Convert convert function if type is Alpha/String
func (commonType *CommonType) Convert() ConvertUnicode {
	return commonType.convert
}

// SetCharset set charset converter
func (commonType *CommonType) SetCharset(name string) {
	commonType.convert = NewUnicodeConverter(name)
}

// SetConvert set convert function if type is Alpha/String
func (commonType *CommonType) SetConvert(c ConvertUnicode) {
	switch commonType.fieldType {
	case FieldTypeString, FieldTypeLAString, FieldTypeLBString:
		commonType.convert = c
	default:
	}
}
