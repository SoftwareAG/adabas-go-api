/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestByteArray(t *testing.T) {
	f, err := initLogWithFile("byte_array.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()
	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeByteArray, "XX")
	barray := newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0}, barray.value)

	adaType = NewTypeWithLength(FieldTypeByteArray, "XX", 2)
	barray = newByteArrayValue(adaType)
	assert.Equal(t, []byte{0x0, 0x0}, barray.value)
	assert.Equal(t, "[0 0]", barray.String())
	var buffer bytes.Buffer
	len := barray.FormatBuffer(&buffer, NewBufferOption(false, false))
	assert.Equal(t, uint32(2), len)
	assert.Equal(t, "XX,2,B", buffer.String())

}
