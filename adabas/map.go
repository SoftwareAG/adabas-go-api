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
	Remarks     string
}

// Map Adabas map structure defining repository where the Map is stored at
type Map struct {
	Name       string       `json:"Name"`
	Isn        adatypes.Isn `json:"Isn"`
	Repository *DatabaseURL
	Data       *DatabaseURL `json:"Data"`
	Fields     []*MapField  `json:"Fields"`
	// Time of last modification of the map
	ModificationTime []uint64
	fieldMap         map[string]*MapField
}

// NewAdabasMap create new Adabas map instance
func NewAdabasMap(name string, repository *DatabaseURL) *Map {
	return &Map{Name: name, Repository: repository}
}

// addFields Add a field to the Map
func (adabasMap *Map) addFields(shortName string, longName string) *MapField {
	adatypes.Central.Log.Debugf("Add map name %s to %s", shortName, longName)
	mField := &MapField{ShortName: shortName, LongName: longName}
	adabasMap.Fields = append(adabasMap.Fields, mField)
	return mField
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
}

// URL current map data URL reference
func (adabasMap *Map) URL() *URL {
	return &adabasMap.Data.URL
}

// extractMapField pass the field definitions of the Map to a Adabas record needed to
// write the definition into the Adabas Map repository file
func extractMapField(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
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
				URL, err := newURL(url)
				if err != nil {
					return adatypes.EndTraverser, err
				}
				adabasMap.Data.URL = *URL
			}
		case mapFieldDataFnr.fieldName():
			adabasMap.Data.Fnr = Fnr(adaValue.Value().(uint32))
			adatypes.Central.Log.Debugf("Got data FNR=%d", adabasMap.Data.Fnr)
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
	tm := adatypes.TraverserValuesMethods{EnterFunction: extractMapField}
	adabasRequest.Definition.TraverseValues(tm, adabasMap)
	return
}

// adaptType called per field of the Map and adapt fields the type of a Map to the correct type
// in case a DDM type is used
func adaptType(adaType adatypes.IAdaType, parentType adatypes.IAdaType, level int, x interface{}) error {
	adatypes.Central.Log.Debugf("Adapt type %s", adaType.Name())
	adabasMap := x.(*Map)
	sn := adaType.Name()[0:2]
	f := adabasMap.fieldMap[sn]
	if f == nil {
		adaType.AddFlag(adatypes.FlagOptionToBeRemoved)
		adatypes.Central.Log.Debugf("Field map does not contain %s", adaType.Name())
		return nil
	}
	adatypes.Central.Log.Debugf("Field map does contains %s", f.LongName)
	adaType.RemoveFlag(adatypes.FlagOptionToBeRemoved)
	adaType.SetName(f.LongName)
	if !adaType.IsStructure() {
		if f.Length < 0 {
			adatypes.Central.Log.Debugf("Set length to 0")
			adaType.SetLength(0)

		} else {
			adatypes.Central.Log.Debugf("Set length to %d", f.Length)
			adaType.SetLength(uint32(f.Length))
		}
	}
	adatypes.Central.Log.Debugf("Set long name %s for %s", f.LongName, adaType.Name())
	return nil
}

// adaptFieldType base class starting the traverser through the fields to adapt field types
func (adabasMap *Map) adaptFieldType(definition *adatypes.Definition) (err error) {
	if definition == nil {
		return adatypes.NewGenericError(19)
	}
	adatypes.Central.Log.Debugf("Adapt map long names to type definition %#v", adabasMap)
	tm := adatypes.NewTraverserMethods(adaptType)
	err = definition.TraverseTypes(tm, true, adabasMap)
	return
}

// Store stores the Adabas map in the given repository
func (adabasMap *Map) Store() error {
	ID := NewAdabasID()
	if adabasMap.Repository == nil {
		return adatypes.NewGenericError(65)
	}
	adabas, err := NewAdabasWithURL(&adabasMap.Repository.URL, ID)
	if err != nil {
		return err
	}
	repository := NewMapRepository(adabas, adabasMap.Repository.Fnr)
	return repository.writeAdabasMapsWithAdabas(adabas, adabasMap)
}
