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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAcbx(t *testing.T) {
	acbx := newAcbx(1)
	assert.Equal(t, Dbid(1), acbx.Acbxdbid)
	acbx.Acbxcid = [4]byte{'A', 'B', 'x', 'n'}
	acbx.Acbxcop = [8]byte{'I', 'J', 0, 0, 0, 0, 0xff, 0x1}
	acbx.Acbxisn = 1234543
	assert.Equal(t, "ACBX:\n  CmdCode:     CmdId: 41 42 78 6e  [ABxn] [...>]\n  Dbid: 1  Filenr: 0  Responsecode: 148 Subcode: 0\n  Isn:  1234543  ISN Lower Limit:  0 ISN Quantity:  0\n  CmdOption: 49 4a 00 00 00 00 ff 01  [IJ....ÿ.] [........]\n  Add1: 20 20 20 20 20 20 20 20  [        ] [........]\n  Add2: 20 20 20 20  [    ] [....]\n  Add3: 00 00 00 00 00 00 00 00  [........] [........]\n  Add4: 00 00 00 00 00 00 00 00  [........] [........]\n  Add5: 00 00 00 00 00 00 00 00  [........] [........]\n  Add6: 00 00 00 00 00 00 00 00  [........] [........]\n  User Area: 00000000000000000000000000000000 [................] [................]\n", acbx.String())
	acbx.resetCop()
	acbx.Acbxrsp = AdaSECUR
	assert.Equal(t, "ACBX:\n  CmdCode:     CmdId: 41 42 78 6e  [ABxn] [...>]\n  Dbid: 1  Filenr: 0  Responsecode: 200 Subcode: 0\n  Isn:  1234543  ISN Lower Limit:  0 ISN Quantity:  0\n  CmdOption: 20 20 20 20 20 20 20 20  [        ] [........]\n  Add1: 20 20 20 20 20 20 20 20  [        ] [........]\n  Add2: 20 20 20 20  [    ] [....]\n  Add3: 00 00 00 00 00 00 00 00  [........] [........]\n  Add4: 00 00 00 00 00 00 00 00  [........] [........]\n  Add5: 00 00 00 00 00 00 00 00  [........] [........]\n  Add6: 00 00 00 00 00 00 00 00  [........] [........]\n  User Area: 00000000000000000000000000000000 [................] [................]\n", acbx.String())
	acbx.resetAcbx()
	assert.Equal(t, "ACBX:\n  CmdCode:     CmdId: 41 42 78 6e  [ABxn] [...>]\n  Dbid: 1  Filenr: 0  Responsecode: 148 Subcode: 0\n  Isn:  0  ISN Lower Limit:  0 ISN Quantity:  0\n  CmdOption: 20 20 20 20 20 20 20 20  [        ] [........]\n  Add1: 20 20 20 20 20 20 20 20  [        ] [........]\n  Add2: 20 20 20 20  [    ] [....]\n  Add3: 00 00 00 00 00 00 00 00  [........] [........]\n  Add4: 00 00 00 00 00 00 00 00  [........] [........]\n  Add5: 00 00 00 00 00 00 00 00  [........] [........]\n  Add6: 00 00 00 00 00 00 00 00  [........] [........]\n  User Area: 00000000000000000000000000000000 [................] [................]\n", acbx.String())
}

func TestAID(t *testing.T) {
	aid := NewAdabasID()
	aid.AddCredential("abc", "def")
	assert.Equal(t, "abc", aid.user)
	assert.Equal(t, "def", aid.pwd)
	fmt.Println(aid)
	aid.isOpen("abc")
}

func TestAIDClone(t *testing.T) {
	aid := NewAdabasID()
	aid.AddCredential("abc", "def")
	time.Sleep(10 * time.Second)
	caid := aid.Clone()
	assert.Equal(t, caid.user, aid.user)
	assert.Equal(t, caid.pwd, aid.pwd)
	assert.Equal(t, caid.AdaID.User, aid.AdaID.User)
	assert.NotEqual(t, caid.AdaID.Pid, aid.AdaID.Pid)
	assert.NotEqual(t, caid.AdaID.Timestamp, aid.AdaID.Timestamp)
}

func TestAdabasOpenParameter(t *testing.T) {
	isq := uint64(101122305)
	assert.Equal(t, "6.7.1.1", parseVersion(isq))
	isq = uint64(101122305)
	assert.Equal(t, "6.7.1.1", parseVersion(isq))

}
