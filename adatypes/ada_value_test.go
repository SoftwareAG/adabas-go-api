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

func TestAdaValueInt64(t *testing.T) {
	adaType := NewType(FieldTypeInt8, "XX")
	value := adaValue{adatype: adaType}

	x := 1
	v, err := value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), v)
	x = 10024432
	v, err = value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(10024432), v)
	x = -23
	v, err = value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(-23), v)
	s := "3409340"
	v, err = value.commonInt64Convert(s)
	assert.NoError(t, err)
	assert.Equal(t, int64(3409340), v)
}

func TestAdaValueInt(t *testing.T) {
	adaType := NewType(FieldTypeInt2, "XX")
	value := adaValue{adatype: adaType}

	x := 1
	v, err := value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), v)
	x = 10024432
	v, err = value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(10024432), v)
	x = -23
	v, err = value.commonInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, int64(-23), v)
	s := "3409340"
	v, err = value.commonInt64Convert(s)
	assert.NoError(t, err)
	assert.Equal(t, int64(3409340), v)
}

func TestAdaValueUint(t *testing.T) {
	adaType := NewType(FieldTypeUInt2, "XX")
	value := adaValue{adatype: adaType}

	x := 1
	v, err := value.commonUInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), v)
	x = 10024432
	v, err = value.commonUInt64Convert(x)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10024432), v)
	x = -23
	v, err = value.commonUInt64Convert(x)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), v)
	ui16 := uint16Value{adaValue: value}
	err = ui16.SetValue(x)
	assert.Error(t, err)
	s := "3409340"
	v, err = value.commonUInt64Convert(s)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3409340), v)
}
