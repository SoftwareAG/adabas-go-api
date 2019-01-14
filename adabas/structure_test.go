/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestABD(t *testing.T) {
	log.Debug("TEST: ", t.Name())
	adabasBuffer := NewBuffer(AbdAQFb)
	assert.Equal(t, uint8('F'), adabasBuffer.abd.Abdid)
	adabasBuffer.Allocate(10)
	adabasBuffer.WriteString("012345")
	assert.Equal(t, 6, adabasBuffer.offset)
	adabasBuffer.WriteString("012345")
	assert.Equal(t, 12, adabasBuffer.offset)
	assert.Equal(t, 5, adabasBuffer.position(5))
	adabasBuffer.WriteString("432")
	assert.Equal(t, 8, adabasBuffer.offset)
	assert.Equal(t, []byte{48, 49, 50, 51, 52, 52, 51, 50, 50, 51, 52, 53}, adabasBuffer.buffer)
	adabasBuffer.WriteBinary([]byte{1, 2, 3, 4})
	assert.Equal(t, 12, adabasBuffer.offset)
	assert.Equal(t, []byte{48, 49, 50, 51, 52, 52, 51, 50, 1, 2, 3, 4}, adabasBuffer.buffer)
}

func TestAcbxReset(t *testing.T) {
	log.Debug("TEST: ", t.Name())
	var acbx Acbx
	// fmt.Println(acbx)
	assert.Equal(t, [2]byte{0, 0}, acbx.Acbxver)
	acbx.resetAcbx()
	// fmt.Println(acbx)
	assert.Equal(t, [2]byte{acbxEyecatcher, acbxVersion}, acbx.Acbxver)
}

func ExampleAdabas_resetAcbx() {
	var acbx Acbx
	acbx.resetAcbx()
	acbx.Acbxcmd = et.code()
	fmt.Println(acbx.String())
	// Output:
	// ACBX:
	//   CmdCode: ET  CmdId: 00000000
	//   Dbid: 0  Filenr: 0  Responsecode: 148 Subcode: 0
	//   Isn:  0  ISN Lower Limit:  0 ISN Quantity:  0
	//   CmdOption: 20 20 20 20 20 20 20 20  [        ] [........]
	//   Add1: 20 20 20 20 20 20 20 20  [        ] [........]
	//   Add2: 20 20 20 20  [    ] [....]
	//   Add3: 00 00 00 00 00 00 00 00  [........] [........]
	//   Add4: 00 00 00 00 00 00 00 00  [........] [........]
	//   Add5: 00 00 00 00 00 00 00 00  [........] [........]
	//   Add6: 00 00 00 00 00 00 00 00  [........] [........]
	//   User Area: 00000000000000000000000000000000 [................] [................]
}
