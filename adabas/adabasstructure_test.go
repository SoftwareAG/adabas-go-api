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

package adabas

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcbx(t *testing.T) {
	acbx := newAcbx(1)
	assert.Equal(t, Dbid(1), acbx.Acbxdbid)
	fmt.Println(acbx)
}

func TestAID(t *testing.T) {
	aid := NewAdabasID()
	aid.AddCredential("abc", "def")
	assert.Equal(t, "abc", aid.user)
	assert.Equal(t, "def", aid.pwd)
	fmt.Println(aid)
	aid.isOpen("abc")
}
