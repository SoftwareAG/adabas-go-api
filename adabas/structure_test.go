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
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestABD(t *testing.T) {
	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabasBuffer := NewBuffer(AbdAQFb)
	assert.Equal(t, uint8('F'), adabasBuffer.abd.Abdid)
	adabasBuffer.Allocate(10)
	adabasBuffer.WriteString("012345")
	assert.Equal(t, 6, adabasBuffer.offset)
	adabasBuffer.WriteString("012345")
	assert.Equal(t, 12, adabasBuffer.offset)
	assert.Equal(t, 5, adabasBuffer.Position(5))
	adabasBuffer.WriteString("432")
	assert.Equal(t, 8, adabasBuffer.offset)
	assert.Equal(t, []byte{48, 49, 50, 51, 52, 52, 51, 50, 50, 51, 52, 53}, adabasBuffer.buffer)
	adabasBuffer.WriteBinary([]byte{1, 2, 3, 4})
	assert.Equal(t, 12, adabasBuffer.offset)
	assert.Equal(t, []byte{48, 49, 50, 51, 52, 52, 51, 50, 1, 2, 3, 4}, adabasBuffer.buffer)
	assert.Equal(t, 12, adabasBuffer.Position(13335))
	var x uint64
	err := binary.Read(adabasBuffer, binary.LittleEndian, &x)
	assert.Error(t, err)
	assert.Equal(t, 3, adabasBuffer.Position(3))
	adabasBuffer.WriteBinary([]byte{1, 2, 0, 0})
	assert.Equal(t, []byte{48, 49, 50, 1, 2, 0, 0, 50, 1, 2, 3, 4}, adabasBuffer.buffer)
	err = binary.Read(adabasBuffer, binary.LittleEndian, &x)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), x)
	assert.Equal(t, 3, adabasBuffer.Position(3))
	var in uint32
	err = binary.Read(adabasBuffer, binary.LittleEndian, &in)
	assert.NoError(t, err)
	assert.Equal(t, uint32(513), in)
	in = 10002
	assert.Equal(t, 4, adabasBuffer.Position(4))
	err = binary.Write(adabasBuffer, binary.LittleEndian, &in)
	assert.NoError(t, err)
	assert.Equal(t, []byte{48, 49, 50, 1, 0x12, 0x27, 0, 0, 1, 2, 3, 4}, adabasBuffer.buffer)
	x = 888
	err = binary.Write(adabasBuffer, binary.LittleEndian, &x)
	assert.Error(t, err)
	assert.Equal(t, []byte{48, 49, 50, 1, 0x12, 0x27, 0, 0, 1, 2, 3, 4}, adabasBuffer.buffer)
}

func TestAcbxReset(t *testing.T) {
	adatypes.Central.Log.Infof("TEST: %s", t.Name())
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
	//   CmdCode: ET  CmdId: 00 00 00 00  [....] [....]
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
