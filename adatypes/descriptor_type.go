/*
* Copyright © 2019-2021 Software AG, Darmstadt, Germany and/or its licensors
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
		flags: uint32(1<<FlagOptionToBeRemoved | 1<<FlagOptionReadOnly),
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

// SetFractional set fractional part
func (adaType *AdaSuperType) SetFractional(x uint32) {
}

// Fractional get fractional part
func (adaType *AdaSuperType) Fractional() uint32 {
	return 0
}

// // SetCharset set fractional part
// func (adaType *AdaSuperType) SetCharset(x string) {
// }

// SetFormatType set format type
func (adaType *AdaSuperType) SetFormatType(x rune) {
}

// FormatType get format type
func (adaType *AdaSuperType) FormatType() rune {
	return adaType.FormatTypeCharacter
}

// SetFormatLength set format length
func (adaType *AdaSuperType) SetFormatLength(x uint32) {
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
		buffer.WriteString(fmt.Sprintf("%s(%d,%d)", s.Name, s.From, s.To))
	}
	buffer.WriteString(fmt.Sprintf(" ; %s", adaType.name))
	return buffer.String()
}

// Value value of the sub or super descriptor
func (adaType *AdaSuperType) Value() (adaValue IAdaValue, err error) {
	Central.Log.Debugf("Return super descriptor value")
	adaValue = newSuperDescriptorValue(adaType)
	return
}

// InitSubTypes init Adabas super/sub types with adabas definition
func (adaType *AdaSuperType) InitSubTypes(definition *Definition) (err error) {
	Central.Log.Debugf("Init super descriptor types of %s", adaType.name)
	for _, s := range adaType.Entries {
		v := definition.fileShortFields[string(s.Name[:])]
		if v == nil {
			return fmt.Errorf("Error init sub type %s", string(s.Name[:]))
		}
		t := NewType(v.Type(), string(s.Name[:]), int(s.To-s.From+1))
		adaType.SubTypes = append(adaType.SubTypes, t)
	}
	return nil
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
		flags:     uint32(1<<FlagOptionToBeRemoved | 1<<FlagOptionReadOnly),
		shortName: name}},
		descriptorLength: descriptorLength, parentName: code}
}

// String string representation of the phonetic type
func (fieldType *AdaPhoneticType) String() string {
	return fmt.Sprintf("%s=PHON(%s) ; %s", fieldType.shortName, fieldType.parentName, fieldType.name)
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
		flags: uint32(1<<FlagOptionToBeRemoved | 1<<FlagOptionReadOnly),
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
	return fmt.Sprintf("%s%s=COLLATING(%s,%s) ; %s", fieldType.shortName, options, fieldType.parentName,
		fieldType.collAttribute, fieldType.name)
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
		flags: uint32(1<<FlagOptionToBeRemoved | 1<<FlagOptionReadOnly),
		name:  name, shortName: name, length: length}},
		fdtFormat: fdtFormat, nr: nr, parentNames: parentNames}
}

// String string representation of the hyper exit type
func (fieldType *AdaHyperExitType) String() string {
	options := fieldType.Option()
	if len(options) > 0 {
		options = "," + strings.Replace(options, " ", ",", -1)
	}
	parents := ""
	for _, p := range fieldType.parentNames {
		if len(parents) > 0 {
			parents += ","
		}
		parents += p
	}
	return fmt.Sprintf("%s %d %c%s=HYPER(%d,%s) ; %s", fieldType.shortName, fieldType.length, fieldType.fdtFormat,
		options, fieldType.nr, parents, fieldType.name)
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
		flags: uint32(1<<FlagOptionToBeRemoved | 1<<FlagOptionReadOnly),
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
	buffer.WriteString(fmt.Sprintf(" ; %s", fieldType.name))
	return buffer.String()
}
