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

func TestCollationValue(t *testing.T) {
	collationType := NewCollationType("AA", 4, "PA", "de@collation=phonebook")

	v, err := collationType.Value()
	assert.NoError(t, err)
	assert.Equal(t, "", v.String())
	sv := v.(*collationValue)
	option := &BufferOption{}
	var buffer bytes.Buffer
	sv.FormatBuffer(&buffer, option)
	assert.Equal(t, "", buffer.String())
	helper := NewHelper([]byte{0x1, 0x2, 0x3, 0x4, 0xff}, 100, endian())
	sv.parseBuffer(helper, option)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4, 0xff}, helper.Buffer())
	assert.Nil(t, sv.Bytes())
	assert.Error(t, sv.SetValue("123"))
	assert.Equal(t, byte(' '), sv.ByteValue())
	assert.Equal(t, uint32(0), helper.Offset())
	sv.StoreBuffer(helper, nil)
	assert.Equal(t, uint32(0), helper.Offset())
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
