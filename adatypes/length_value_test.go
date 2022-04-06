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
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldLength(t *testing.T) {
	err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeFieldLength, "I8")
	up := newLengthValue(adaType)
	fmt.Println("Length value ", up.value)
	assert.Equal(t, uint32(0), up.Value())
	up.SetValue(1000)
	assert.Equal(t, uint32(1000), up.Value())
	var buffer bytes.Buffer
	option := &BufferOption{}
	len := up.FormatBuffer(&buffer, option)
	assert.Equal(t, uint32(4), len)
	assert.Equal(t, "I8L,4,B", buffer.String())
}
