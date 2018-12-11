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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloating(t *testing.T) {
	adaType := NewType(FieldTypeFloat, "FL")
	adaType.length = 4
	fl := newFloatValue(adaType)
	assert.Equal(t, float32(0), fl.Value())
	fl.SetStringValue("0.1")
	assert.Equal(t, float32(0.1), fl.Value())
	fl.SetStringValue("10.1")
	assert.Equal(t, float32(10.1), fl.Value())
	fl.SetValue(0.5)
	assert.Equal(t, float32(0.5), fl.Value())

}
