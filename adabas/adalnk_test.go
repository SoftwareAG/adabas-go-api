// +build adalnk,cgo

/*
* Copyright © 2018-2021 Software AG, Darmstadt, Germany and/or its licensors
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

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdabasLnkOk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)

	var abds []*Buffer
	abds = append(abds, NewBuffer(AbdAQFb))
	abds = append(abds, NewBuffer(AbdAQRb))

	abds[0].WriteString("A")
	abds[1].WriteString(".")

	adabas.Acbx.Acbxcmd = op.code()
	adabas.SetAbd(abds)

	retb := adabas.CallAdabas()
	if retb != nil {
		t.Fatal("Adabas call return value not correct", retb)
	}

	abds[0].Clear()
	abds[0].WriteString("AA.")
	abds[1].Allocate(8)

	adabas.Acbx.Acbxcmd = l1.code()
	adabas.Acbx.Acbxfnr = 11
	adabas.Acbx.Acbxisn = 1

	retb = adabas.CallAdabas()
	if retb != nil {
		t.Fatal("Adabas call return value not correct", retb)
	}
	assert.Equal(t, "50005800", string(abds[1].Bytes()))
	driver := adabas.URL.Instance(adabas.ID)
	assert.IsType(t, (*AdaIPC)(nil), driver)

	adabas.Acbx.Acbxcmd = cl.code()
	retb = adabas.CallAdabas()
	if retb != nil {
		t.Fatal("Adabas call return value not correct", retb)
	}

	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	assert.Equal(t, uint16(0), adabas.Acbx.Acbxrsp)
	require.NoError(t, retb)
	adabas.Acbx.resetAcbx()
}
