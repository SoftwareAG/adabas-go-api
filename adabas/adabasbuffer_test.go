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

package adabas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdabasBuffer(t *testing.T) {
	a := NewBuffer(AbdAQFb)
	assert.Equal(t, AbdAQFb, a.ID())
	assert.Equal(t, uint64(0), a.Received())
	a.WriteString("AA.")
	assert.Len(t, a.Bytes(), 3)
	assert.Equal(t, uint64(3), a.Size())
	a.Clear()
	assert.Len(t, a.Bytes(), 0)
	assert.Equal(t, uint64(0), a.Size())

}

func TestAdabasBufferSize(t *testing.T) {
	a := NewBufferWithSize(AbdAQRb, 100)
	assert.Len(t, a.Bytes(), 100)
	assert.Equal(t, uint64(100), a.Size())
	assert.Equal(t, AbdAQRb, a.ID())
	a.grow(150)
	assert.Len(t, a.Bytes(), 150)
	assert.Equal(t, uint64(150), a.Size())
	a.extend(50)
	assert.Len(t, a.Bytes(), 200)
	assert.Equal(t, uint64(200), a.Size())
	assert.Equal(t, 5, a.Position(5))
}
