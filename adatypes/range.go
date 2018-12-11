/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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
)

const (
	lastEntry = -2
	noEntry   = -1
)

// AdaRange Adabas range definition
type AdaRange struct {
	from int
	to   int
}

// NewEmptyRange create new empty range
func NewEmptyRange() *AdaRange {
	return &AdaRange{from: noEntry, to: noEntry}
}

// NewSingleRange new single range
func NewSingleRange(index int) *AdaRange {
	return &AdaRange{from: index, to: index}
}

// NewRange new range from to
func NewRange(from, to int) *AdaRange {
	return &AdaRange{from: from, to: to}
}

// NewLastRange range defining only the last entry
func NewLastRange() *AdaRange {
	return &AdaRange{from: lastEntry, to: lastEntry}
}

// FormatBuffer generate corresponding format buffer
func (adaRange *AdaRange) FormatBuffer() string {
	var buffer bytes.Buffer
	if adaRange.from == lastEntry {
		buffer.WriteRune('N')
	} else if adaRange.from > 0 {
		buffer.WriteString(fmt.Sprintf("%d", adaRange.from))
	}
	if adaRange.to > 0 && adaRange.from != adaRange.to {
		if adaRange.from > 0 {
			buffer.WriteString("-N")
		} else {
			buffer.WriteString(fmt.Sprintf("-%d", adaRange.from))
		}
	}
	return buffer.String()
}
