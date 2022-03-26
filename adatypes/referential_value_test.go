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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferential(t *testing.T) {
	superType := NewReferentialType("AA", 11, [2]string{"AA", "BB"}, 1, 1, 1)

	v, err := superType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "AA=REFINT(AA,1,AA)", v.String())
	sv := v.(*referentialValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	sv.FormatBuffer(&buffer, option)
	assert.Equal(t, "", buffer.String())
	helper := NewHelper([]byte{0x1, 0x2, 0x3, 0x4, 0xff}, 100, endian())
	_, err = sv.parseBuffer(helper, option)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4, 0xff}, helper.Buffer())
	assert.Equal(t, []byte(nil), sv.Bytes())
	err = sv.SetValue("123")
	assert.Error(t, err)
	assert.Equal(t, byte(' '), sv.ByteValue())
	assert.Equal(t, uint32(0), helper.Offset())
	sv.StoreBuffer(helper, nil)
	assert.Equal(t, uint32(0), helper.Offset())
	assert.Equal(t, "", sv.Value())
	_, err = sv.Int32()
	assert.Error(t, err)
	_, err = sv.Int64()
	assert.Error(t, err)
	_, err = sv.UInt32()
	assert.Error(t, err)
	_, err = sv.UInt64()
	assert.Error(t, err)
	_, err = sv.Float()
	assert.Error(t, err)
}
