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
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
	// Options information
	mapFieldOptions
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
	DefaultCharset       string
	fieldMap             map[string]*MapField
	redefinitionFieldMap map[string][]*MapField
	dynamic              *adatypes.DynamicInterface
	lock                 *sync.Mutex
}

// NewAdabasMap create new Adabas map instance containing the long name
// to short name definition. The definition is enhance to include extra
// charset and dynamic length definition like the Natural DDM provides it.
// In advance redefinition of fields to a subset of fields is possible.
func NewAdabasMap(param ...interface{}) *Map {
	redefinitionFieldMap := make(map[string][]*MapField)
	redefinitionFields := make([]*MapField, 0)
	switch param[0].(type) {
	case string:
		name := param[0].(string)
		if len(param) == 1 {
			return &Map{Name: name, redefinitionFieldMap: redefinitionFieldMap,
				RedefinitionFields: redefinitionFields, lock: &sync.Mutex{}}
		}
		repository := param[1].(*DatabaseURL)
		return &Map{Name: name, Repository: repository, DefaultCharset: "US-ASCII",
			redefinitionFieldMap: redefinitionFieldMap,
			RedefinitionFields:   redefinitionFields, lock: &sync.Mutex{}}
	case *DatabaseURL:
		repository := param[0].(*DatabaseURL)
		dataRepository := param[1].(*DatabaseURL)
		redefinitionFieldMap := make(map[string][]*MapField)
		return &Map{Repository: repository, Data: dataRepository, DefaultCharset: "US-ASCII",
			redefinitionFieldMap: redefinitionFieldMap,
			RedefinitionFields:   redefinitionFields, lock: &sync.Mutex{}}
	}
	return nil
}

// addFields Add a field shortname/long name definition to the Map
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

// FieldNames list of fields of the map is returned
func (adabasMap *Map) FieldNames() []string {
	fields := make([]string, 0)
	for _, f := range adabasMap.Fields {
		if strings.Trim(f.LongName, " ") != "" {
			fields = append(fields, f.LongName)
		}
	}
	return fields
}

// addRedefinitionField add a sub redefinition of an field of the Adabas file.
// In Cobol and Natural it is possible to redefine Alpha fields to contain a
// subset of other possible field types.
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

func (adabasMap *Map) setDefaultOptions(options string) {
	if options != "" {
		re := regexp.MustCompile(`(\w+)=([^,]*),?`)
		rr := re.FindAllStringSubmatch(options, -1)
		for _, r1 := range rr {
			switch strings.ToLower(r1[1]) {
			case "charset":
				adabasMap.DefaultCharset = r1[2]
			default:
			}
		}
	}
}

// FieldShortNames list of field short names of the Map
func (adabasMap *Map) FieldShortNames() []string {
	fields := make([]string, 0)
	for _, f := range adabasMap.Fields {
		fields = append(fields, f.ShortName)
	}
	return fields
}

// String reports the Map repository, data reference and the fields mapping of a map
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

// createFieldMap create a field hash map to find fields very quick
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

// URL current map data URL reference. This is the database URL where the
// Map data is taken from.
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
		case mapFieldOptions.fieldName():
			adabasMap.setDefaultOptions(adaValue.String())
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
	_, err = adabasRequest.Definition.TraverseValues(tm, adabasMap)
	return
}

// Store stores the Adabas map in the given Adabas Map repository. It generates
// a new entry if it is new or update current entry
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

// Delete deletes the Adabas map in the given Adabas Map repository.
func (adabasMap *Map) Delete() error {
	ID := NewAdabasID()
	if adabasMap.Repository == nil {
		return adatypes.NewGenericError(65)
	}
	adabas, err := NewAdabas(&adabasMap.Repository.URL, ID)
	if err != nil {
		return err
	}
	repository := NewMapRepository(adabas.URL, adabasMap.Repository.Fnr)
	return repository.DeleteMap(adabas, adabasMap.Name)
}

// define the map definition using interface tags
func (adabasMap *Map) defineByInterface(i interface{}) error {
	adabasMap.dynamic = adatypes.CreateDynamicInterface(i)
	adatypes.Central.Log.Debugf("Create dynamic interface %v", adabasMap.dynamic)
	for index, f := range adabasMap.dynamic.FieldNames {
		adatypes.Central.Log.Debugf("Define field %s: %s", index, f)
		adabasMap.addFields(index, index)
	}
	return nil
}
