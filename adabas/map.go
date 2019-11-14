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

type mapField uint

const (
	// mapFieldIndicator Field Indicator
	mapFieldIndicator mapField = iota
	// HOst name the map is generated on
	mapFieldHost
	// Date of generation
	mapFieldDate
	// Version used to generate map
	mapFieldVersion
	// Map name
	mapFieldName
	// Reference file number
	mapFieldReferenceFileNr
	// Referenced database URL
	mapFieldReferenceURL
	// Flag for this map
	mapFieldFlags
	// Directory information
	mapFieldDirectory
	// Data file number reference number
	mapFieldDataFnr
	// Data database reference URL
	mapFieldDataURL
	// Period group containing all fields
	mapFieldFields
	// Short name
	mapFieldShortname
	// Type of the field (extended DDM information)
	mapFieldTypeConversion
	// Long name information
	mapFieldLongname
	// Length (override)
	mapFieldLength
	// Content type
	mapFieldContentType
	// format type
	mapFieldFormatType
	// Remarks
	mapFieldRemarks
	// Data modify time
	mapFieldModifyTime
)

var mapFieldNames = []string{"TA", "AB", "AC", "AD", "RN", "DF", "RD", "RB", "RO",
	"RF", "DD", "MA", "MB", "MC", "MD", "ML", "MT", "MY", "MR", "ZB"}

func (cc mapField) fieldName() string {
	fn := mapFieldNames[cc]
	return fn
}

// MapField Structure to define short name to long name mapping.
// In advance the DDM specific type formater like B for Boolean or
// N for NATDATE are available
type MapField struct {
	ShortName   string `json:"ShortName"`
	LongName    string `json:"LongName"`
	Length      int32  `json:"FormatLength"`
	ContentType string `json:"ContentType"`
	FormatType  string `json:"FormatType"`
	FieldType   string `json:"FieldType"`
	Charset     string `json:"Charset"`
	File        uint32 `json:"File"`
	Remarks     string
}

// Map Adabas map structure defining repository where the Map is stored at
type Map struct {
	Name               string       `json:"Name"`
	Version            string       `json:"Version"`
	Isn                adatypes.Isn `json:"Isn"`
	Repository         *DatabaseURL
	Data               *DatabaseURL `json:"Data"`
	Fields             []*MapField  `json:"Fields"`
	RedefinitionFields []*MapField  `json:"RedefinitionFields"`
	// Time of last modification of the map
	Generated            uint64
	ModificationTime     []uint64
	fieldMap             map[string]*MapField
	redefinitionFieldMap map[string][]*MapField
}

// NewAdabasMap create new Adabas map instance
func NewAdabasMap(param ...interface{}) *Map {
	redefinitionFieldMap := make(map[string][]*MapField)
	switch param[0].(type) {
	case string:
		name := param[0].(string)
		if len(param) == 1 {
			return &Map{Name: name, redefinitionFieldMap: redefinitionFieldMap}
		}
		repository := param[1].(*DatabaseURL)
		return &Map{Name: name, Repository: repository,
			redefinitionFieldMap: redefinitionFieldMap}
	case *DatabaseURL:
		repository := param[0].(*DatabaseURL)
		dataRepository := param[1].(*DatabaseURL)
		redefinitionFieldMap := make(map[string][]*MapField)
		return &Map{Repository: repository, Data: dataRepository,
			redefinitionFieldMap: redefinitionFieldMap}
	}
	return nil
}

// addFields Add a field to the Map
func (adabasMap *Map) addFields(shortName string, longName string) *MapField {
	adatypes.Central.Log.Debugf("Add map name %s to %s", shortName, longName)
	mField := &MapField{ShortName: shortName, LongName: longName}
	if strings.HasPrefix(shortName, "#") {
		adabasMap.addRedefinitionField(mField)
	} else {
		adabasMap.Fields = append(adabasMap.Fields, mField)
	}
	return mField
}

// FieldNames list of fields of map
func (adabasMap *Map) FieldNames() []string {
	fields := make([]string, 0)
	for _, f := range adabasMap.Fields {
		fields = append(fields, f.LongName)
	}
	return fields
}

func (adabasMap *Map) addRedefinitionField(mField *MapField) {
	adabasMap.RedefinitionFields = append(adabasMap.RedefinitionFields, mField)
	adatypes.Central.Log.Debugf("%s add redefinition -> %s", mField.ShortName, mField.ShortName[1:])
	if rf, ok := adabasMap.redefinitionFieldMap[mField.ShortName[1:]]; ok {
		adatypes.Central.Log.Debugf("Insert %s", mField.ShortName[1:])
		rf = append(rf, mField)
		adabasMap.redefinitionFieldMap[mField.ShortName[1:]] = rf
	} else {
		adatypes.Central.Log.Debugf("Add %s", mField.ShortName[1:])
		rf = make([]*MapField, 0)
		rf = append(rf, mField)
		adabasMap.redefinitionFieldMap[mField.ShortName[1:]] = rf
	}

}

// FieldShortNames list of fields of map
func (adabasMap *Map) FieldShortNames() []string {
	fields := make([]string, 0)
	for _, f := range adabasMap.Fields {
		fields = append(fields, f.ShortName)
	}
	return fields
}

// String report the Map repository, data reference and the fields mapping of a map
func (adabasMap *Map) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("MapName: " + adabasMap.Name + "\n")
	if adabasMap.Repository == nil {
		buffer.WriteString("Map Repository: not defined\n")
	} else {
		buffer.WriteString("Map Repository: URL=" + adabasMap.Repository.URL.String() + " fnr=" + strconv.Itoa(int(adabasMap.Repository.Fnr)) + "\n")
	}
	if adabasMap.Data == nil {
		buffer.WriteString("Data Repository: not defined\n")
	} else {
		buffer.WriteString("Data Repository: URL=" + adabasMap.Data.URL.String() + " fnr=" + strconv.Itoa(int(adabasMap.Data.Fnr)) + "\n")
	}
	buffer.WriteString("Fields:\n")
	for _, f := range adabasMap.Fields {
		buffer.WriteString("  sn=" + f.ShortName + " ln=" + f.LongName + " len=" + strconv.Itoa(int(f.Length)) + "\n")
		buffer.WriteString("  contenttype=" + f.ContentType + " ft=" + f.FormatType + " rem=" + f.Remarks + "\n")
	}
	return buffer.String()
}

// createFieldMap create a field map to find fields very quick
func (adabasMap *Map) createFieldMap() {
	adabasMap.fieldMap = make(map[string]*MapField)
	for _, f := range adabasMap.Fields {
		if f.ShortName != "" {
			sn := f.ShortName[0:2]
			adatypes.Central.Log.Debugf("Add %s/%s %s to hash map", f.ShortName, f.LongName, sn)
			adabasMap.fieldMap[sn] = f
		}
	}
	adatypes.Central.Log.Debugf("Number of hash map entries %d", len(adabasMap.fieldMap))
}

// URL current map data URL reference
func (adabasMap *Map) URL() *URL {
	return &adabasMap.Data.URL
}

// traverseExtractMapField pass the field definitions of the Map to a Adabas record needed to
// write the definition into the Adabas Map repository file
func traverseExtractMapField(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	adabasMap := x.(*Map)
	adatypes.Central.Log.Debugf("Extract map field=%s >%s< pe index=%d", adaValue.Type().Name(), adaValue.String(), adaValue.PeriodIndex())
	if adaValue.PeriodIndex() > 0 {

		var mapField *MapField
		adatypes.Central.Log.Debugf("Adabas Map index=%d of %d", adaValue.PeriodIndex(), len(adabasMap.Fields))
		if len(adabasMap.Fields) < int(adaValue.PeriodIndex()) {
			adatypes.Central.Log.Debugf("Create new map")
			mapField = &MapField{}
			adabasMap.Fields = append(adabasMap.Fields, mapField)
		} else {
			adatypes.Central.Log.Debugf("Take index field %v", adaValue.PeriodIndex())
			mapField = adabasMap.Fields[adaValue.PeriodIndex()-1]
		}
		switch adaValue.Type().Name() {
		case mapFieldShortname.fieldName():
			mapField.ShortName = strings.TrimSpace(adaValue.String())
			if strings.HasPrefix(mapField.ShortName, "#") {
				adabasMap.addRedefinitionField(mapField)
			}
		case mapFieldLongname.fieldName():
			mapField.LongName = adaValue.String()
		case mapFieldLength.fieldName():
			mapField.Length = int32(adaValue.Value().(uint32))
		case mapFieldContentType.fieldName():
			mapField.ContentType = adaValue.String()
		case mapFieldFormatType.fieldName():
			mapField.FormatType = adaValue.String()
		case mapFieldRemarks.fieldName():
			mapField.Remarks = adaValue.String()
		case mapFieldTypeConversion.fieldName():
			adatypes.Central.Log.Debugf("Type conversion : %s", adaValue.String())
		case mapFieldFields.fieldName():
			adatypes.Central.Log.Debugf("Got field name")
		default:
			fmt.Printf("Unknown : %s:%s\n", adaValue.Type().Name(), adaValue.String())
		}
	} else {
		switch adaValue.Type().Name() {
		case mapFieldVersion.fieldName():
			adabasMap.Version = adaValue.String()
			switch adabasMap.Version {
			case "1", "2":
			default:
				return adatypes.EndTraverser, adatypes.NewGenericError(94, adabasMap.Version)
			}
		case mapFieldName.fieldName():
			adabasMap.Name = adaValue.String()
		case mapFieldReferenceURL.fieldName():
			url := adaValue.String()
			adatypes.Central.Log.Debugf("Got data reference URL >%s<", url)
			if strings.Trim(url, " ") == "" {
				adatypes.Central.Log.Debugf("Use reference Adabas %#v", adabasMap.Repository)
				adabasMap.Data.URL = adabasMap.Repository.URL
			} else {
				adatypes.Central.Log.Debugf("Create new data reference URL with >%s<", url)
				URL, err := NewURL(url)
				if err != nil {
					return adatypes.EndTraverser, err
				}
				adabasMap.Data.URL = *URL
			}
		case mapFieldDataFnr.fieldName():
			adabasMap.Data.Fnr = Fnr(adaValue.Value().(uint32))
			adatypes.Central.Log.Debugf("Got data FNR=%d", adabasMap.Data.Fnr)
		case mapFieldDate.fieldName():
			adabasMap.Generated = adaValue.Value().(uint64)
			adatypes.Central.Log.Debugf("Got date=%d", adabasMap.Generated)
		case mapFieldModifyTime.fieldName():
			if adaValue.Type().Type() != adatypes.FieldTypeMultiplefield {
				adabasMap.ModificationTime = append(adabasMap.ModificationTime, adaValue.Value().(uint64))
			} else {
				muTime := adaValue.(*adatypes.StructureValue)
				for _, values := range muTime.Elements {
					for _, v := range values.Values {
						adabasMap.ModificationTime = append(adabasMap.ModificationTime, v.Value().(uint64))
					}
				}
			}
		}
	}
	return adatypes.Continue, nil
}

// parseMap Adabas read parser of one Map definition used during read
func parseMap(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adabasMap := x.(*Map)
	isn := adabasRequest.Isn
	adabasMap.Isn = isn

	adatypes.Central.Log.Debugf("Got Map ISN %d record", isn)
	tm := adatypes.TraverserValuesMethods{EnterFunction: traverseExtractMapField}
	adabasRequest.Definition.TraverseValues(tm, adabasMap)
	return
}

// traverseAdaptType called per field of the Map and adapt fields the type of a Map to the correct type
// in case a DDM type is used
func traverseAdaptType(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	adabasMap := x.(*Map)
	sn := adaType.ShortName()[0:2]
	adatypes.Central.Log.Debugf("Adapt type %s/%s", adaType.Name(), sn)
	f := adabasMap.fieldMap[sn]
	if f == nil {
		adaType.AddFlag(adatypes.FlagOptionToBeRemoved)
		adatypes.Central.Log.Debugf("Field %s flag to be removed", adaType.Name())
		return nil
	}
	adatypes.Central.Log.Debugf("Field map does contains %s -> %s/%s format type=>%s<", f.LongName, adaType.Name(), adaType.ShortName(), f.FormatType)
	adaType.RemoveFlag(adatypes.FlagOptionToBeRemoved)
	adaType.SetName(f.LongName)

	switch strings.Trim(f.FormatType, " ") {
	case "#":
		adatypes.Central.Log.Debugf("Replace %s on parent %s", adaType.ShortName(), parentType.ShortName())
		nt := adatypes.NewRedefinitionType(adaType)
		adaptRedefintionFields(nt, adabasMap.redefinitionFieldMap[adaType.ShortName()])
		st := parentType.(*adatypes.StructureType)
		err := st.ReplaceType(adaType, nt)
		if err != nil {
			return err
		}
		return nil
	case "":
	default:
		adaType.SetFormatType([]rune(f.FormatType)[0])
	}
	adaType.SetFormatLength(uint32(f.Length))
	if !adaType.IsStructure() {
		if f.Length < 0 {
			adatypes.Central.Log.Debugf("Set %s length to 0", adaType.Name())
			adaType.SetLength(0)
		} else {
			adatypes.Central.Log.Debugf("Set %s length to %d, check content type=%s",
				adaType.Name(), f.Length, f.ContentType)
			adaType.SetLength(uint32(f.Length))
			ct := strings.Split(f.ContentType, ",")
			for _, c := range ct {
				p := strings.Split(c, "=")
				if len(p) > 1 {
					adatypes.Central.Log.Debugf("%s=%s", p[0], p[1])
					s := strings.ToLower(p[0])
					switch s {
					case "fractionalshift":
						fs, ferr := strconv.Atoi(p[1])
						if ferr != nil {
							return ferr
						}
						adaType.SetFractional(uint32(fs))
					case "charset":
						adaType.SetCharset(p[1])
					case "formattype":
						if p[1] != "" {
							adaType.SetFormatType(rune(p[1][0]))
						}
					case "length":
						fs, ferr := strconv.Atoi(p[1])
						if ferr != nil {
							return ferr
						}
						adaType.SetFormatLength(uint32(fs))
					default:
						fmt.Println("Unknown paramteter", p[0])
					}
				}
			}
		}
	}
	adatypes.Central.Log.Debugf("Set long name %s for %s/%s", f.LongName, adaType.Name(), adaType.ShortName())
	return nil
}

func adaptRedefintionFields(redType *adatypes.RedefinitionType, fields []*MapField) {
	adatypes.Central.Log.Debugf("Fields: %#v", fields)
	for _, f := range fields {
		adatypes.Central.Log.Debugf("%s %s %s %d", f.ShortName, f.LongName, f.FormatType, f.Length)
		fieldType := adatypes.EvaluateFieldType([]rune(f.FormatType)[0], f.Length)
		subType := adatypes.NewType(fieldType, f.ShortName, uint32(f.Length))
		subType.SetName(f.LongName)
		adatypes.Central.Log.Debugf("%s:%s %s %d %d", f.LongName, f.ShortName, f.FormatType, f.Length, subType.Length())
		redType.AddSubType(subType)
	}
}

// adaptFieldType base class starting the traverser through the fields to adapt field types.
// The long name is adapted to the field entries in the overall file definition
func (adabasMap *Map) adaptFieldType(definition *adatypes.Definition, dynamic *adatypes.DynamicInterface) (err error) {
	if definition == nil {
		return adatypes.NewGenericError(19)
	}
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "before adapt field types")
		adatypes.Central.Log.Debugf("Adapt map long names to type definition %#v", adabasMap.String())
	}
	tm := adatypes.NewTraverserMethods(traverseAdaptType)
	err = definition.TraverseTypes(tm, true, adabasMap)
	if err != nil {
		return
	}
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "before restrict slice")
	}
	// Restrict fields to the fields included in the map
	fields := adabasMap.FieldNames()
	adatypes.Central.Log.Debugf("Check %v", fields)
	if dynamic != nil {
		newFields := make([]string, 0)
		for _, f := range fields {
			if _, ok := dynamic.FieldNames[f]; ok {
				//adatypes.Central.Log.Debugf("Check %s -> %s ok=%v", f, fn, ok)
				newFields = append(newFields, f)
			}
		}
		fields = newFields
		adatypes.Central.Log.Debugf("Redefine %v", fields)
	}
	// TODO restrict to interface if given
	err = definition.RestrictFieldSlice(fields)
	if adatypes.Central.IsDebugLevel() {
		definition.DumpTypes(true, false, "after restrict slice")
	}
	return
}

// Store stores the Adabas map in the given repository
func (adabasMap *Map) Store() error {
	ID := NewAdabasID()
	if adabasMap.Repository == nil {
		return adatypes.NewGenericError(65)
	}
	adabas, err := NewAdabas(&adabasMap.Repository.URL, ID)
	if err != nil {
		return err
	}
	repository := NewMapRepository(adabas.URL, adabasMap.Repository.Fnr)
	return repository.writeAdabasMapsWithAdabas(adabas, adabasMap)
}
