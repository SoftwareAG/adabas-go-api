/*
* Copyright © 2018-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"fmt"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// Adabas field name length maximum
const fieldNameLength = 2

// Field identifier used in the FDT call
type fieldIdentifier uint

const (
	fieldIdentifierField fieldIdentifier = iota
	fieldIdentifierSub
	fieldIdentifierSuper
	fieldIdentifierPhonetic
	fieldIdentifierCollation
	fieldIdentifierReferential
	fieldIdentifierHyperexit
)

var fieldIdentifiers = []byte{'F', 'S', 'T', 'P', 'C', 'R', 'H'}

func (fdt fieldIdentifier) code() byte {
	return byte(fieldIdentifiers[fdt])
}

// Adabas FDT call fdt options used for getting field information
type option byte

const (
	fdtFlagOption2NC     option = iota // (1<<0) /* 0x01 Not counted field (SQL) */
	fdtFlagOption2NN                   // (1<<1) /* 0x02 Field must not contain NULL */
	fdtFlagOption2LB                   // (1<<2) /* 0x04 4 bytes inclusive length for var.len fields */
	fdtFlagOption2LA                   // (1<<3) /* 0x08 Long Alpha (up to 16 K)*/
	fdtFlagOption2UNUSED               // (1<<3) /* 0x10 Unused */
	fdtFlagOption2HF                   // (1<<4) /* 0x20 HF option active */
	fdtFlagOption2NV                   // (1<<5) /* 0x40 No conversion over net-work */
	fdtFlagOption2NB                   // (1<<6) /* 0x89 NB option active */
)

const (
	fdtFlagOption1UQ option = iota // (1<<0) /* 0x01 UQ option (unique descriptor) */
	fdtFlagOption1SB               // (1<<1) /* 0x02 Field is sub descriptor */
	fdtFlagOption1PH               // (1<<2) /* 0x04 Field is phonetic descriptor */
	fdtFlagOption1PE               // (1<<3) /* 0x08 PE (Group-level) */
	fdtFlagOption1NU               // (1<<4) /* 0x10 NU option (zero-suppression) */
	fdtFlagOption1MU               // (1<<5) /* 0x20 Multiple value field */
	fdtFlagOption1FI               // (1<<6) /* 0x40 FI option (fixed length) */
	fdtFlagOption1DE               // (1<<7) /* 0x80 Field is descriptor */
)

const (
	fdtFlagMfOption1DE option = iota // (1<<0) /* 0x01 descriptor */
	fdtFlagMfOption1FI               // (1<<1) /* 0x02 FI option (fixed length) */
	fdtFlagMfOption1MU               // (1<<2) /* 0x04 Multiple value field */
	fdtFlagMfOption1NU               // (1<<3) /* 0x08 NU option (zero-suppression) */
	fdtFlagMfOption1PE               // (1<<4) /* 0x10 PE (Group-level) */
	fdtFlagMfOption1PH               // (1<<5) /* 0x20 Field is phonetic descriptor */
	fdtFlagMfOption1SB               // (1<<6) /* 0x40 Field is sub descriptor */
	fdtFlagMfOption1UQ               // (1<<7) /* 0x80 UQ option (unique descriptor) */
)

const (
	fdtIdentifier = "FieldIdentifier"
	fdtLength     = "fdtLength"
	fdtStrLevel   = "fdtStrLevel"
	fdtFlag       = "fdtFlag"
	fdtCount      = "fdtCount"
	fdtTime       = "fdtTime"
)

func (cc option) iv() int {
	return int(cc)
}

// FDT field entry structures
var fdtFieldEntry = []adatypes.IAdaType{
	adatypes.NewTypeWithLength(adatypes.FieldTypeString, "fieldName", 2),
	adatypes.NewType(adatypes.FieldTypeUInt2, "fieldFrom"),
	adatypes.NewType(adatypes.FieldTypeUInt2, "fieldTo"),
}

// FDT hyper field entry structures
var fdtHyperFieldEntry = []adatypes.IAdaType{
	adatypes.NewTypeWithLength(adatypes.FieldTypeString, "fieldName", 2),
}

// FDT main field structures
var fdt = []adatypes.IAdaType{
	adatypes.NewType(adatypes.FieldTypeCharacter, fdtIdentifier),                       // 0
	adatypes.NewType(adatypes.FieldTypeLength, "FieldDefLength"),                       // 1
	adatypes.NewTypeWithLength(adatypes.FieldTypeString, "fieldName", fieldNameLength), // 2
	adatypes.NewType(adatypes.FieldTypeCharacter, "fieldFormat"),                       // 3
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldOption"),                           // 4
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldOption2"),                          // 5
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldLevel"),
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldEditMask"),
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldSubOption"),
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldSYfunction"),
	adatypes.NewType(adatypes.FieldTypeUByte, "fieldDeactivate"), // 10
	adatypes.NewType(adatypes.FieldTypeUInt4, "fieldLength"),
	adatypes.NewType(adatypes.FieldTypeUInt2, "superLength"),
	adatypes.NewType(adatypes.FieldTypeByte, "superOption2"),
	adatypes.NewStructureList(adatypes.FieldTypeStructure, "superList", adatypes.OccByte, fdtFieldEntry),
	adatypes.NewType(adatypes.FieldTypeUInt2, "subLength"), // 15
	adatypes.NewType(adatypes.FieldTypeUByte, "subOption2"),
	adatypes.NewTypeWithLength(adatypes.FieldTypeFiller, "FILL1", 2),
	adatypes.NewTypeWithLength(adatypes.FieldTypeString, "parentName", 2),
	adatypes.NewType(adatypes.FieldTypeUInt2, "subFrom"),
	adatypes.NewType(adatypes.FieldTypeUInt2, "subTo"), // 20
	adatypes.NewType(adatypes.FieldTypeUInt2, "colLength"),
	adatypes.NewType(adatypes.FieldTypeString, "colParentName"),
	adatypes.NewType(adatypes.FieldTypeUInt2, "colInternalLength"),
	adatypes.NewType(adatypes.FieldTypeUByte, "colOption2"),
	adatypes.NewTypeWithLength(adatypes.FieldTypeString, "colAttribute", 0), // 25
	adatypes.NewType(adatypes.FieldTypeUInt2, "hyperLength"),
	adatypes.NewType(adatypes.FieldTypeUByte, "hyperFExit"),
	adatypes.NewType(adatypes.FieldTypeUByte, "hyperOption2"),
	adatypes.NewTypeWithLength(adatypes.FieldTypeFiller, "FILL2", 1),
	adatypes.NewStructureList(adatypes.FieldTypeStructure, "hyperList",
		adatypes.OccByte, fdtHyperFieldEntry), // 30
	adatypes.NewType(adatypes.FieldTypeUInt4, "refFile"),
	adatypes.NewType(adatypes.FieldTypeString, "refPrimaryKey"),
	adatypes.NewType(adatypes.FieldTypeString, "refForeignKey"),
	adatypes.NewType(adatypes.FieldTypeUByte, "refType"),
	adatypes.NewType(adatypes.FieldTypeUByte, "refUpdateAction"), // 35
	adatypes.NewType(adatypes.FieldTypeUByte, "refDeleteAction"),
}

// FDT condition matrix defining various parts of the field types needed
var fdtCondition = map[byte][]byte{
	fieldIdentifierField.code(): {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	fieldIdentifierSuper.code(): {1, 2, 3, 4, 12, 13, 14},
	fieldIdentifierSub.code():   {1, 2, 3, 4, 12, 13, 14},
	//		fieldIdentifierSub.code():         []byte{1, 2, 3, 4, 15, 16, 17, 18, 19, 20},
	fieldIdentifierPhonetic.code():    {1, 2, 3, 4, 12, 17, 18},
	fieldIdentifierCollation.code():   {1, 2, 3, 4, 12, 18, 23, 24, 25},
	fieldIdentifierHyperexit.code():   {1, 2, 3, 4, 26, 27, 28, 29, 30},
	fieldIdentifierReferential.code(): {1, 2, 31, 32, 33, 34, 35, 36},
}

// FDT general main level layout for the Adabas LA call
var fdtGeneralLayout = []adatypes.IAdaType{
	adatypes.NewType(adatypes.FieldTypeUInt4, fdtLength),
	adatypes.NewType(adatypes.FieldTypeByte, fdtStrLevel),
	adatypes.NewType(adatypes.FieldTypeUByte, fdtFlag),
	adatypes.NewType(adatypes.FieldTypeUInt2, fdtCount),
	adatypes.NewType(adatypes.FieldTypeUInt8, fdtTime),
	adatypes.NewStructureCondition(adatypes.FieldTypeStructure, "fdt", fdt, adatypes.NewFieldCondition(1, 0, fdtCondition)),
}

// Create used definition to read FDT
func createFdtDefintion() *adatypes.Definition {
	return adatypes.NewDefinitionWithTypes(fdtGeneralLayout)
}

// Traverser to count fields
// func traverserFieldDefinitionCreator(adaValue adatypes.IAdaValue, level int, x interface{}) bool {
// 	number := x.(*int)
// 	(*number)++
// 	return true
// }

// Create field definition table definition useds to parse Adabas LA call
// getting the Adabas file definition out of the FDT
func createFieldDefinitionTable(fdtDef *adatypes.Definition) (definition *adatypes.Definition, err error) {
	definition = adatypes.NewDefinition()
	fdtSearch := fdtDef.Search("fdt")
	fdt := fdtSearch.(*adatypes.StructureValue)
	nrFdtEntries := len(fdt.Elements)
	stack := adatypes.NewStack()
	definition.FileTime = fdtDef.Search(fdtTime)
	var lastStruct adatypes.IAdaType
	for index := 1; index < nrFdtEntries+1; index++ {
		value := fdt.Get(fdtIdentifier, index)

		var fieldType adatypes.IAdaType
		switch value.Value().(byte) {
		case 0:
			break
		case fieldIdentifierField.code():
			fieldType, err = createFieldType(fdt, index)
			if err != nil {
				return
			}
			adatypes.Central.Log.Debugf("Found normal field %s level=%d fieldType=%v", fieldType.Name(), fieldType.Level(), fieldType.Type())
		case fieldIdentifierSub.code(), fieldIdentifierSuper.code():
			adatypes.Central.Log.Debugf("Found Super/Sub field %c\n", value.Value().(byte))
			fieldType, err = createSubSuperDescriptorType(fdt, index)
			if err != nil {
				return
			}
		case fieldIdentifierPhonetic.code():
			adatypes.Central.Log.Debugf("Found Super/Sub field %c\n", value.Value().(byte))
			fieldType, err = createPhoneticType(fdt, index)
			if err != nil {
				return
			}
		case fieldIdentifierCollation.code():
			adatypes.Central.Log.Debugf("Found Collation field %c\n", value.Value().(byte))
			fieldType, err = createCollationType(fdt, index)
			if err != nil {
				return
			}
		case fieldIdentifierHyperexit.code():
			adatypes.Central.Log.Debugf("Found HyperExit field %c\n", value.Value().(byte))
			fieldType, err = createHyperExitType(fdt, index)
			if err != nil {
				return
			}
		case fieldIdentifierReferential.code():
			adatypes.Central.Log.Debugf("Found Referential field %c\n", value.Value().(byte))
			fieldType, err = createReferential(fdt, index)
			if err != nil {
				return
			}
		default:
			fmt.Printf("Not implemented already >%c<\n", value.Value().(byte))
			err = adatypes.NewGenericError(11, value.Value().(byte), value.Value().(byte))
			return
		}
		if fieldType != nil {
			for {
				if lastStruct != nil {
					if lastStruct.Level() == fieldType.Level()-1 {
						if adatypes.Central.IsDebugLevel() {
							adatypes.Central.Log.Debugf("Append to structure %s add %s %d", lastStruct.Name(), fieldType.String(), fieldType.Level())
						}
						lastStruct.(*adatypes.StructureType).AddField(fieldType)
						break
					} else {
						popElement, _ := stack.Pop()
						if popElement == nil {
							lastStruct = nil
							definition.AppendType(fieldType)
							if adatypes.Central.IsDebugLevel() {
								adatypes.Central.Log.Debugf("%s append to main %v", fieldType.Name(), fieldType.Type())
							}
							break
						} else {
							if adatypes.Central.IsDebugLevel() {
								adatypes.Central.Log.Debugf("Pop from Stack %v", popElement)
							}
							lastStruct = popElement.(adatypes.IAdaType)
							if adatypes.Central.IsDebugLevel() {
								adatypes.Central.Log.Debugf("Level equal last=%s %d current=%s %d",
									lastStruct.String(), lastStruct.Level(), fieldType.String(), fieldType.Level())
							}
						}
					}
				} else {
					adatypes.Central.Log.Debugf("Append to main %d %s", fieldType.Level(), fieldType.Name())
					definition.AppendType(fieldType)
					break
				}
			}
			if fieldType.IsStructure() && fieldType.Type() != adatypes.FieldTypeMultiplefield {
				lastStruct = fieldType
				if adatypes.Central.IsDebugLevel() {
					adatypes.Central.Log.Debugf("Pop to MU Stack %v", lastStruct)
				}
				stack.Push(lastStruct)
			}
			if adatypes.Central.IsDebugLevel() {
				adatypes.Central.Log.Debugf("Current structure %v", lastStruct)
			}
			definition.Register(fieldType)
		}
		adatypes.Central.Log.Debugf("Field type DONE")
	}
	definition.InitReferences()

	return
}

// create a common field type for a field
func createFieldType(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	length := fdt.Get("fieldLength", index).Value().(uint32)
	fdtFormat := fdt.Get("fieldFormat", index).Value().(byte)
	option := fdt.Get("fieldOption", index).Value().(uint8)
	option2 := fdt.Get("fieldOption2", index).Value().(uint8)
	level := fdt.Get("fieldLevel", index).Value().(uint8)
	sysf := fdt.Get("fieldSYfunction", index).Value().(uint8)
	editMask := fdt.Get("fieldEditMask", index).Value().(uint8)
	subOption := fdt.Get("fieldSubOption", index).Value().(uint8)

	adatypes.Central.Log.Debugf("Create field type %s check option=%v check containing %v", name, option, fdtFlagOption1PE)
	// Check if field is period element
	if level == 1 && option&(1<<fdtFlagOption1PE) > 0 {
		adatypes.Central.Log.Debugf("%s is PE", name)
		fieldType = adatypes.NewStructureEmpty(adatypes.FieldTypePeriodGroup, name, adatypes.OccUInt2, level)
	} else {
		// Normal field, check format
		var id adatypes.FieldType
		switch fdtFormat {
		case 'A':
			switch {
			case option2&(1<<fdtFlagOption2LA) > 0:
				id = adatypes.FieldTypeLAString
			case option2&(1<<fdtFlagOption2LB) > 0:
				id = adatypes.FieldTypeLBString
			default:
				id = adatypes.FieldTypeString
			}
		case 'W':
			switch {
			case option2&(1<<fdtFlagOption2LA) > 0:
				id = adatypes.FieldTypeLAUnicode
			case option2&(1<<fdtFlagOption2LB) > 0:
				id = adatypes.FieldTypeLBUnicode
			default:
				id = adatypes.FieldTypeUnicode
			}
		case 'P':
			id = adatypes.FieldTypePacked
		case 'U':
			id = adatypes.FieldTypeUnpacked
		case 'B':
			id = evaluateIntegerValue(true, length)
		case 'F':
			id = evaluateIntegerValue(false, length)
		case 'G':
			switch length {
			case 4:
				id = adatypes.FieldTypeFloat
			case 8:
				id = adatypes.FieldTypeFloat
			default:
				err = adatypes.NewGenericError(12, length)
				return
			}
		case ' ':
			adatypes.Central.Log.Debugf("%s created as Group", name)
			fieldType = adatypes.NewStructureEmpty(adatypes.FieldTypeGroup, name, adatypes.OccSingle, level)
			return
		default:
			err = adatypes.NewGenericError(13, fdtFormat)
			return
		}

		// flag option check
		adatypes.Central.Log.Debugf("Id=%d name=%s length=%d format=%c", id, name, length, fdtFormat)
		if (option & (1 << fdtFlagOption1MU)) > 0 {
			newType := adatypes.NewTypeWithLength(id, name, length)
			evaluateOption(newType, option, option2)
			newType.SysField = sysf
			newType.EditMask = editMask
			newType.SubOption = subOption

			newType.AddFlag(adatypes.FlagOptionAtomicFB)
			adatypes.Central.Log.Debugf("%s created as MU  on top of the field MU=%v %p", name, newType.HasFlagSet(adatypes.FlagOptionAtomicFB), newType)

			fieldTypes := []adatypes.IAdaType{newType}
			fieldType = adatypes.NewStructureList(adatypes.FieldTypeMultiplefield, name, adatypes.OccUInt2, fieldTypes)
			adatypes.Central.Log.Debugf("%s MU structure %d -> %p", fieldType.Name(), fieldType.(*adatypes.StructureType).NrFields(), fieldType)
		} else {
			newType := adatypes.NewTypeWithLength(id, name, length)
			adatypes.Central.Log.Debugf("%s created as normal field %p", name, newType)
			evaluateOption(newType, option, option2)
			newType.SysField = sysf
			newType.EditMask = editMask
			newType.SubOption = subOption
			fieldType = newType
		}

	}
	fieldType.SetLevel(level)
	return
}

// Create Super-/Sub- Descriptor types
func createSubSuperDescriptorType(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	superList := fdt.Get("superList", index).(*adatypes.StructureValue)
	fdtFormat := fdt.Get("fieldFormat", index).Value().(byte)
	option := fdt.Get("fieldOption", index).Value().(byte)
	superType := adatypes.NewSuperType(name, option)
	superType.FdtFormat = fdtFormat
	for _, sub := range superList.Elements {
		superType.AddSubEntry(string(sub.Values[0].Value().([]byte)), sub.Values[1].Value().(uint16), sub.Values[2].Value().(uint16))
	}
	fieldType = superType

	return
}

// create phonetic type
func createPhoneticType(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	descriptorLength := fdt.Get("superLength", index).Value().(uint16)
	parentName := string(fdt.Get("parentName", index).Value().([]byte))
	fieldType = adatypes.NewPhoneticType(name, descriptorLength, parentName)
	return
}

// create collation descriptor type
func createCollationType(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	length := fdt.Get("superLength", index).Value().(uint16)
	//	fdtFormat := fdt.Get("fieldFormat", index).Value().(byte)
	parentName := string(fdt.Get("parentName", index).Value().([]byte))
	colAttribute := string(fdt.Get("colAttribute", index).Value().([]byte))
	adatypes.Central.Log.Debugf("Collation attribute : %s", colAttribute)
	collType := adatypes.NewCollationType(name, length, parentName, colAttribute)
	option := fdt.Get("fieldOption", index).Value().(uint8)
	adatypes.Central.Log.Debugf("Option %d", option)

	flags := []byte{0x1, 0x8, 0x10, 0x20}
	optionFlags := []adatypes.FieldOption{adatypes.FieldOptionUQ, adatypes.FieldOptionPE,
		adatypes.FieldOptionNU, adatypes.FieldOptionMU}
	for index, f := range flags {
		if (option & f) > 0 {
			collType.AddOption(optionFlags[index])
		}
	}
	if (option & 0x3) == 0 {
		collType.AddOption(adatypes.FieldOptionHE)
	}

	option2 := fdt.Get("colOption2", index).Value().(uint8)
	if (option2 & 0x4) == 0 {
		collType.AddOption(adatypes.FieldOptionLA)
	}
	if (option2 & 0x8) == 0 {
		collType.AddOption(adatypes.FieldOptionLB)
	}
	if (option2 & 0x80) == 0 {
		collType.AddOption(adatypes.FieldOptionColExit)
	}

	fieldType = collType
	return
}

// create hyperexit type
func createHyperExitType(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	length := fdt.Get("hyperLength", index).Value().(uint16)
	fdtFormat := fdt.Get("fieldFormat", index).Value().(byte)
	nr := fdt.Get("hyperFExit", index).Value().(uint8)

	hyperList := fdt.Get("hyperList", index).(*adatypes.StructureValue)
	var parentFieldNames []string
	for _, hyper := range hyperList.Elements {
		parentFieldNames = append(parentFieldNames, string(hyper.Values[0].Value().([]byte)))
	}

	hyperType := adatypes.NewHyperExitType(name, uint32(length), fdtFormat, nr, parentFieldNames)

	option := fdt.Get("fieldOption", index).Value().(uint8)
	flags := []byte{0x1, 0x4, 0x8, 0x10, 0x20}
	optionFlags := []adatypes.FieldOption{adatypes.FieldOptionUQ, adatypes.FieldOptionHE, adatypes.FieldOptionPE,
		adatypes.FieldOptionNU, adatypes.FieldOptionMU}
	for index, f := range flags {
		if (option & f) > 0 {
			hyperType.AddOption(optionFlags[index])
		}
	}
	fieldType = hyperType

	return
}

// create referential integrity
func createReferential(fdt *adatypes.StructureValue, index int) (fieldType adatypes.IAdaType, err error) {
	name := string(fdt.Get("fieldName", index).Value().([]byte))
	refFile := fdt.Get("refFile", index).Value().(uint32)
	var keys [2]string
	keys[0] = string(fdt.Get("refPrimaryKey", index).Value().([]byte))
	keys[1] = string(fdt.Get("refForeignKey", index).Value().([]byte))
	refType := fdt.Get("refType", index).Value().(uint8)
	refUpdateAction := fdt.Get("refUpdateAction", index).Value().(uint8)
	refDeleteAction := fdt.Get("refDeleteAction", index).Value().(uint8)
	referentialType := adatypes.NewReferentialType(name, refFile, keys,
		refType, refUpdateAction, refDeleteAction)

	fieldType = referentialType
	return
}

// evaluate type of integer dependent on length
func evaluateIntegerValue(binary bool, length uint32) adatypes.FieldType {
	switch {
	case length == 4 && binary:
		return adatypes.FieldTypeUInt4
	case length == 4:
		return adatypes.FieldTypeInt4
	case length == 2 && binary:
		return adatypes.FieldTypeUInt2
	case length == 2:
		return adatypes.FieldTypeInt2
	case length == 1 && binary:
		return adatypes.FieldTypeUByte
	case length == 1:
		return adatypes.FieldTypeByte
	case length == 8 && binary:
		return adatypes.FieldTypeUInt8
	case length == 8:
		return adatypes.FieldTypeInt8
	default:
		return adatypes.FieldTypeByteArray
	}
}

// Evaluate option for a field types
func evaluateOption(fieldType *adatypes.AdaType, option uint8, option2 uint8) {
	flags := [...]int{fdtFlagOption1UQ.iv(), fdtFlagOption1NU.iv(), fdtFlagOption1FI.iv(), fdtFlagOption1DE.iv(), fdtFlagOption1MU.iv()}
	optionFlags := []adatypes.FieldOption{adatypes.FieldOptionUQ, adatypes.FieldOptionNU, adatypes.FieldOptionFI, adatypes.FieldOptionDE, adatypes.FieldOptionMU}
	flags2 := [...]int{fdtFlagOption2NC.iv(), fdtFlagOption2NN.iv(), fdtFlagOption2HF.iv(), fdtFlagOption2NV.iv(), fdtFlagOption2NB.iv()}
	optionFlags2 := []adatypes.FieldOption{adatypes.FieldOptionNC, adatypes.FieldOptionNN, adatypes.FieldOptionHF, adatypes.FieldOptionNV, adatypes.FieldOptionNB}

	adatypes.Central.Log.Debugf("Evaluate Options %x\n", option)
	for i := 0; i < len(flags); i++ {
		if (option & (1 << uint32(flags[i]))) > 0 {
			adatypes.Central.Log.Debugf("%s Option %d", fieldType.String(), i)
			fieldType.AddOption(optionFlags[i])
		}
	}

	adatypes.Central.Log.Debugf("Evaluate Options2 %v", option2)
	for i := 0; i < len(flags2); i++ {
		if (option2 & (1 << uint32(flags2[i]))) > 0 {
			adatypes.Central.Log.Debugf("%s Option2 %d", fieldType.String(), i)
			fieldType.AddOption(optionFlags2[i])
		}
	}
}
