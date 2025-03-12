/*
* Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

type entry struct {
	libName     string
	ddmName     string
	rootField   *field
	parentField *field
	stack       *adatypes.Stack
	adabasMap   *Map
}

type fieldType int

const (
	rootField fieldType = iota
	normalField
	packField
	unpackField
	groupField
	superField
	periodField
)

const (
	optMultiple   = (1 << 0)
	optNullField  = (1 << 1)
	optDescriptor = (1 << 2)
	optFixField   = (1 << 3)
)

type field struct {
	level           int
	shortname       string
	longname        string
	length          int
	fType           fieldType
	formatType      string
	options         uint64
	fractionalShift string
	childs          []*field
	comment         string
	// parent          *field
}

func (f *field) isGroup() bool {
	switch f.fType {
	case groupField, periodField:
		return true
	default:
	}
	return false
}

func (f *field) String() string {
	var buffer bytes.Buffer
	for i := 0; i < f.level; i++ {
		buffer.WriteString(" ")
	}
	buffer.WriteString(fmt.Sprintf("%d %s %s %d %s\n", f.level, f.shortname, f.longname, f.length, f.formatType))
	for _, sf := range f.childs {
		buffer.WriteString(sf.String())
	}
	return buffer.String()
}

// ImportMapRepository import map by file import
func (repository *Repository) ImportMapRepository(adabas *Adabas, filter string,
	fileName string, mapURL *DatabaseURL) (maps []*Map, err error) {
	adatypes.Central.Log.Debugf("Import map repository of %s using filter %s", fileName, filter)
	if mapURL != nil {
		adatypes.Central.Log.Debugf("Importing map repository to redefined %s", mapURL.URL.String())
	}
	var file *os.File
	file, err = os.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	err = repository.LoadMapRepository(adabas)
	if err != nil {
		return nil, err
	}
	suffixCheck := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(suffixCheck, ".json"):
		maps, err = ParseJSONFileForFields(file)
		if err != nil {
			return
		}
	case strings.HasSuffix(suffixCheck, ".systrans"):
		maps, err = parseSystransFileForFields(file)
		if err != nil {
			return
		}
	case strings.HasSuffix(suffixCheck, ".xml"):
		maps, err = parseXMLFileForFields(file)
		if err != nil {
			return
		}
	default:
		return nil, adatypes.NewGenericError(55, fileName)
	}

	var dataRepository *Repository
	if mapURL != nil {
		dataRepository = &Repository{DatabaseURL: *mapURL}
	}
	maps = repository.filterMaps(filter, maps)
	for _, m := range maps {
		if dataRepository != nil {
			m.Data = &dataRepository.DatabaseURL
		}
		m.Repository = &repository.DatabaseURL
	}

	return
}

func (repository *Repository) filterMaps(filter string, maps []*Map) []*Map {
	if filter == "" || filter == "*" {
		return maps
	}
	tmpMaps := make([]*Map, 0)
	for _, m := range maps {
		if matched, _ := regexp.MatchString(filter, m.Name); matched {
			tmpMaps = append(tmpMaps, m)
			m.Repository = &repository.DatabaseURL
		}
	}
	return tmpMaps
}

func (curEntry *entry) searchParent(f *field) (err error) {

	switch {
	case f.level == 1:
		curEntry.rootField.childs = append(curEntry.rootField.childs, f)
		curEntry.stack.Clear()
	case f.level-1 == curEntry.parentField.level:
		curEntry.parentField.childs = append(curEntry.parentField.childs, f)
	default:
		parent := curEntry.parentField
		for parent.level != f.level-1 {
			t, serr := curEntry.stack.Pop()
			parent = t.(*field)
			if serr != nil {
				err = serr
				return
			}
		}
		curEntry.stack.Push(parent)
		parent.childs = append(parent.childs, f)
		curEntry.parentField = parent
	}

	curEntry.stack.Push(f)
	curEntry.parentField = f
	return
}

/**
* Parse XML map export file
**/
func parseXMLFileForFields(file *os.File) (maps []*Map, err error) {
	return nil, nil
}

/**
 * Parse SYSTRANS file for Entries and there field information
 * @return List of SYSTRANS entries in mapping entry list
 */
func parseSystransFileForFields(file *os.File) (maps []*Map, err error) {
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var curEntry *entry
	maps = make([]*Map, 0)
	shortNames := make(map[string]bool)
	var comment string
	for scanner.Scan() {
		line := scanner.Text()
		// 		while (line != null && !end) {
		// 			try {
		lineNumber++
		//		var ddmEntries []*entry
		phoneticDescriptor := false
		if len(line) > 4 {
			if strings.HasPrefix(line, "*C**") {

				// Only parse DDM views (V)
				if line[76] == 'V' {
					curEntry = &entry{libName: line[36:44], ddmName: line[44 : 44+32], rootField: &field{}, stack: adatypes.NewStack(),
						adabasMap: NewAdabasMap(line[44 : 44+32])}
					shortNames = make(map[string]bool)
					maps = append(maps, curEntry.adabasMap)
				}
			}
			if curEntry != nil {
				if strings.HasPrefix(line, "*S***") {
					if strings.HasPrefix(line, "*S**HD=") {
						comment += line[7:]
					} else {
						comment += line[5:]

					}
				} else {
					c := string(line[4])
					// 							if (!end) {
					if strings.Contains(" GMP", c) {
						if strings.HasPrefix(line, "*S**") {
							if len(line) > 46 && line[46:47] == "P" {
								phoneticDescriptor = true
								adatypes.Central.Log.Debugf("Phonetic: %s -> %v", line[46:47], phoneticDescriptor)
							}
							// field
							sn := line[6:8]
							// a ddm can contain more than
							// one
							// reference to a field, use the
							// first
							// only
							adatypes.Central.Log.Debugf("SN=%s", sn)
							if _, ok := shortNames[sn]; !ok && !phoneticDescriptor {
								adatypes.Central.Log.Debugf("not found SN=%s phonetic=%v", sn, phoneticDescriptor)
								shortNames[sn] = true
								level, cerr := strconv.Atoi(line[5:6])
								if cerr != nil {
									err = cerr
									return
								}
								field := &field{level: level, shortname: sn,
									longname: strings.Trim(line[8:39], " "), fType: normalField, comment: comment}
								if len(line) > 46 && line[46:47] == "S" {
									field.fType = superField
								}
								err = curEntry.searchParent(field)
								if err != nil {
									return
								}
								// 											parentField =
								// 												checkTreeReference(
								// 													parentStack,
								// 													parentField, rootField,
								// 													level, field);
								// 											lastField = field;
								// 											boolean lengthSet = true;
								if field.fType != superField {
									switch c {
									case "G":
										field.fType = groupField
									case "P":
										field.fType = periodField
									default:
										field.fType = normalField
										field.formatType = line[40:41]
									}
								}

								/* Parse length */
								length, cerr :=
									strconv.Atoi(line[42:44])
								if cerr != nil {
									err = cerr
									return
								}
								if (field.fType == packField) || (field.fType == unpackField) {
									length =
										(length / 2) + 1
								}
								field.length = length

								/*
								 * Check through some extra options
								 */
								if c == "M" {
									field.options = field.options | optMultiple
								}
								if len(line) > 44 {
									field.fractionalShift = line[44:45]
								}
								if len(line) > 45 && line[45:46] == "N" {
									field.options |= optNullField
								}

								if len(line) > 46 && line[46:47] == "D" {
									field.options |= optDescriptor
								}
								if field.length < 0 && field.length > 65000 {
									err = adatypes.NewGenericError(165, field.length)
									return
								}
								if field.length > 0 && field.length < 3 && !field.isGroup() {
									field.options ^= optNullField
									field.options |= optFixField
								}
								af := &MapField{ShortName: field.shortname, LongName: field.longname,
									Length:     int32(field.length),
									FormatType: field.formatType}
								curEntry.adabasMap.Fields = append(curEntry.adabasMap.Fields, af)
							}
						}
					}
				}

			}

		}
	}
	if err = scanner.Err(); err != nil {
		return
	}

	return
}
