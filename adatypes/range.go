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
	"fmt"
	"regexp"
	"strconv"
)

const (
	// LastEntry last N name index for Adabas
	LastEntry  = -2
	noEntry    = -1
	allEntries = -124
)

// AdaRange Adabas range definition
type AdaRange struct {
	from int
	to   int
}

// NewEmptyRange create an empty range
func NewEmptyRange() *AdaRange {
	return &AdaRange{from: noEntry, to: noEntry}
}

// NewRangeParser new range using string parser
func NewRangeParser(r string) *AdaRange {
	var re = regexp.MustCompile(`(?m)^(N|[0-9]*)-?(N|[0-9]*)?$`)

	match := re.FindStringSubmatch(r)
	if match == nil {
		Central.Log.Debugf("Does not match: %s", r)
		return nil
	}

	if Central.IsDebugLevel() {
		Central.Log.Debugf("Got matches %s->%s,%s", r, match[1], match[1])
	}
	from := 0
	to := 0
	var err error
	if len(match) > 1 {
		if match[1] == "N" {
			from = LastEntry
			to = LastEntry
		} else {
			from, err = strconv.Atoi(match[1])
			if err != nil {
				return nil
			}
			to = from
		}
	}
	if len(match) > 2 && match[2] != "" {
		if match[2] == "N" {
			to = LastEntry
		} else {
			if from == LastEntry {
				return nil
			}
			to, err = strconv.Atoi(match[2])
			if err != nil {
				if Central.IsDebugLevel() {
					Central.Log.Debugf("Integer error: %s -> %s", r, match[2])
				}
				return nil
			}
		}
	}
	if to < from {
		if to != LastEntry {
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Last entry error: %s -> %d < %d", r, to, LastEntry)
			}
			return nil
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Create new range %d-%d", from, to)
	}
	return &AdaRange{from: from, to: to}
}

// NewSingleRange new single dimensioned range
func NewSingleRange(index int) *AdaRange {
	return &AdaRange{from: index, to: index}
}

// NewPartialRange new partial range
func NewPartialRange(from, to int) *AdaRange {
	return &AdaRange{from: from, to: to}
}

// NewRange new range from a dimension to a dimension
func NewRange(from, to int) *AdaRange {
	if from > to {
		if to != LastEntry {
			return nil
		}
	}
	return &AdaRange{from: from, to: to}
}

// NewLastRange range defining only the last entry
func NewLastRange() *AdaRange {
	return &AdaRange{from: LastEntry, to: LastEntry}
}

// FormatBuffer generate corresponding format buffer
func (adaRange *AdaRange) FormatBuffer() string {
	if adaRange == nil {
		return ""
	}
	var buffer bytes.Buffer
	if adaRange.from == LastEntry {
		buffer.WriteRune('N')
	} else if adaRange.from > 0 {
		buffer.WriteString(fmt.Sprintf("%d", adaRange.from))
	}
	if adaRange.to != 0 && adaRange.from != adaRange.to {
		if adaRange.to == LastEntry {
			buffer.WriteString("-N")
		} else {
			buffer.WriteString(fmt.Sprintf("-%d", adaRange.to))
		}
	}
	return buffer.String()
}

func (adaRange *AdaRange) multiplier() int {
	if adaRange.to == adaRange.from {
		return 1
	}
	if adaRange.to != LastEntry && adaRange.from != LastEntry {
		return adaRange.to - adaRange.from + 1
	}
	return allEntries
}

func (adaRange *AdaRange) index(pos uint32, max uint32) uint32 {
	if adaRange.from == LastEntry {
		return max
	}
	if adaRange.from > 0 {
		return uint32(adaRange.from) + pos - 1
	}
	return pos
}

// IsSingleIndex is a single index query, although range available
func (adaRange *AdaRange) IsSingleIndex() bool {
	//Central.Log.Debugf("%d to %d", adaRange.from, adaRange.to)
	if adaRange.from == 0 && adaRange.to == 0 {
		return false
	}
	if adaRange.from == noEntry {
		return false
	}
	if adaRange.from == adaRange.to {
		return true
	}
	return false
}
