/*
* Copyright © 2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCode(t *testing.T) {
	cc := s4
	assert.Equal(t, "S4", cc.command())
	cc = v3
	assert.Equal(t, "V3", cc.command())
	assert.Equal(t, byte('V'), cc.code()[0])
	assert.Equal(t, byte('3'), cc.code()[1])

	assert.True(t, validAcbxCommand(cc.code()))
	assert.False(t, validAcbxCommand([2]byte{'X', '1'}))
}
