/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhonetic(t *testing.T) {
	adaType := NewType(FieldTypePhonetic, "PH")
	ph := newPhoneticValue(adaType)
	assert.Equal(t, "", ph.String())
	err := ph.SetValue(1234)
	assert.Error(t, err)
	i64, i64err := ph.Int64()
	assert.Error(t, i64err)
	assert.Equal(t, int64(0), i64)

	ui64, ui64err := ph.UInt64()
	assert.Error(t, ui64err)
	assert.Equal(t, uint64(0), ui64)

}
