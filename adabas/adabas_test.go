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
	"bytes"
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adabasModDBID   = 23
	adabasModDBIDs  = "23"
	adabasStatDBID  = 24
	adabasStatDBIDs = "24"
)

func TestAdabasFailure(t *testing.T) {
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(0)
	assert.Nil(t, adabas)
	assert.Error(t, err)
	adabas, err = NewAdabas(1)
	assert.NotNil(t, adabas)
	assert.NoError(t, err)

	var abds []*Buffer
	abds = append(abds, NewBuffer(AbdAQFb))
	abds = append(abds, NewBuffer(AbdAQRb))

	abds[0].WriteString("A")
	abds[1].WriteString(".")

	adabas.Acbx.Acbxcmd = op.code()
	adabas.SetAbd(abds)

	retb := adabas.CallAdabas()
	if retb == nil {
		t.Fatal("Adabas call return value not correct", retb)
	}
	assert.Error(t, retb)

	adatypes.Central.Log.Debugf("acbx ver=%v", adabas.Acbx.Acbxver)
	adatypes.Central.Log.Debugf("acbx cmd=%v", adabas.Acbx.Acbxcmd)
	adatypes.Central.Log.Debugf("acbx len=%v", adabas.Acbx.Acbxlen)

	adabas.Acbx.Acbxcmd = cl.code()

	adatypes.Central.Log.Debugf("acbx ver=%v", adabas.Acbx.Acbxver)
	adatypes.Central.Log.Debugf("acbx cmd=%v", adabas.Acbx.Acbxcmd)
	adatypes.Central.Log.Debugf("acbx len=%v", adabas.Acbx.Acbxlen)

	retb = adabas.CallAdabas()
	if retb == nil {
		t.Fatal("Adabas call return value not correct", retb)
	}
	assert.Error(t, retb)
	if adabas.Acbx.Acbxrsp != 148 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	assert.Equal(t, uint16(148), adabas.Acbx.Acbxrsp)
	adabas.Acbx.resetAcbx()
	assert.Equal(t, uint16(148), adabas.Acbx.Acbxrsp)
}

func TestAdabasOk(t *testing.T) {
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
	assert.False(t, adabas.IsRemote())

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

const CDVTReq = uint64(0x98BADCFE)

func TestAdabasOpen(t *testing.T) {
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

func TestAdabasFdt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("fdt")

	fmt.Println("Open database")
	errOpen := adabas.Open()
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	assert.Equal(t, nil, errOpen)

	fmt.Println("Read file definition")
	definition, err := adabas.ReadFileDefinition(11)
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	if err != nil {
		t.Fatal("Adabas error incorrect", err)
	}
	fmt.Println(definition)

	fmt.Println("Close database")
	adabas.Close()
	fmt.Println("test done")
}

func ExampleAdabas_readFileDefinitionFile11() {
	err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("fdt")

	fmt.Println("Open database")
	err = adabas.Open()
	if err != nil || adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Error: ", err, " ", adabas.Acbx.Acbxrsp)
	}
	defer adabas.Close()
	fmt.Println("Read file definition")
	var definition *adatypes.Definition
	definition, err = adabas.ReadFileDefinition(11)
	if adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Resonse code : ", adabas.Acbx.Acbxrsp)
		return
	}
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	definition.DumpTypes(false, false)
	// Output: Open database
	// Read file definition
	// Dump all file field types:
	//   1, AA, 8, A ,UQ,DE ; AA
	//   1, AB  ; AB
	//     2, AC, 20, A ,NU ; AC
	//     2, AE, 20, A ,DE ; AE
	//     2, AD, 20, A ,NU ; AD
	//   1, AF, 1, A ,FI ; AF
	//   1, AG, 1, A ,FI ; AG
	//   1, AH, 4, P ,DE,NC ; AH
	//   1, A1  ; A1
	//     2, AI, 20, A ,NU,MU; AI
	//       3, AI, 20, A ,NU,MU ; AI
	//     2, AJ, 20, A ,NU,DE ; AJ
	//     2, AK, 10, A ,NU ; AK
	//     2, AL, 3, A ,NU ; AL
	//   1, A2  ; A2
	//     2, AN, 6, A ,NU ; AN
	//     2, AM, 15, A ,NU ; AM
	//   1, AO, 6, A ,DE ; AO
	//   1, AP, 25, A ,NU,DE ; AP
	//   1, AQ ,PE ; AQ
	//     2, AR, 3, A ,NU ; AR
	//     2, AS, 5, P ,NU ; AS
	//     2, AT, 5, P ,NU,MU; AT
	//       3, AT, 5, P ,NU,MU ; AT
	//   1, A3  ; A3
	//     2, AU, 2, U  ; AU
	//     2, AV, 2, U ,NU ; AV
	//   1, AW ,PE ; AW
	//     2, AX, 8, U ,NU ; AX
	//     2, AY, 8, U ,NU ; AY
	//   1, AZ, 3, A ,NU,DE,MU; AZ
	//     2, AZ, 3, A ,NU,DE,MU ; AZ
	//  PH=PHON(AE) ; PH
	//  H1=AU(1,2),AV(1,2) ; H1
	//  S1=AO(1,4) ; S1
	//  S2=AO(1,6),AE(1,20) ; S2
	//  S3=AR(1,3),AS(1,9) ; S3

}

func ExampleAdabas_readFileDefinition9() {
	err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("fdt")

	fmt.Println("Open database")
	err = adabas.Open()
	if err != nil || adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Error: ", err, " ", adabas.Acbx.Acbxrsp)
	}
	defer adabas.Close()
	fmt.Println("Read file definition")
	var definition *adatypes.Definition
	definition, err = adabas.ReadFileDefinition(9)
	if adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Resonse code : ", adabas.Acbx.Acbxrsp)
		return
	}
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	definition.DumpTypes(false, false)
	// Output:Open database
	// Read file definition
	// Dump all file field types:
	//   1, A0  ; A0
	//     2, AA, 8, A ,UQ,DE,NC,NN ; AA
	//     2, AB  ; AB
	//       3, AC, 4, F ,DE ; AC
	//       3, AD, 8, B ,NU,HF ; AD
	//       3, AE, 0, A ,NU,NV,NB ; AE
	//   1, B0  ; B0
	//     2, BA, 40, W ,NU ; BA
	//     2, BB, 40, W ,NU ; BB
	//     2, BC, 50, W ,NU,DE ; BC
	//   1, CA, 1, A ,FI ; CA
	//   1, DA, 1, A ,FI ; DA
	//   1, EA, 4, P ,DE,NC ; EA
	//   1, F0 ,PE ; F0
	//     2, FA, 60, W ,NU,MU; FA
	//       3, FA, 60, W ,NU,MU ; FA
	//     2, FB, 40, W ,NU,DE ; FB
	//     2, FC, 10, A ,NU ; FC
	//     2, FD, 3, A ,NU ; FD
	//     2, F1  ; F1
	//       3, FE, 6, A ,NU ; FE
	//       3, FF, 15, A ,NU ; FF
	//       3, FG, 15, A ,NU ; FG
	//       3, FH, 15, A ,NU ; FH
	//       3, FI, 80, A ,NU,DE,MU; FI
	//         4, FI, 80, A ,NU,DE,MU ; FI
	//   1, I0 ,PE ; I0
	//     2, IA, 40, W ,NU,MU; IA
	//       3, IA, 40, W ,NU,MU ; IA
	//     2, IB, 40, W ,NU,DE ; IB
	//     2, IC, 10, A ,NU ; IC
	//     2, ID, 3, A ,NU ; ID
	//     2, IE, 5, A ,NU ; IE
	//     2, I1  ; I1
	//       3, IF, 6, A ,NU ; IF
	//       3, IG, 15, A ,NU ; IG
	//       3, IH, 15, A ,NU ; IH
	//       3, II, 15, A ,NU ; II
	//       3, IJ, 80, A ,NU,DE,MU; IJ
	//         4, IJ, 80, A ,NU,DE,MU ; IJ
	//   1, JA, 6, A ,DE ; JA
	//   1, KA, 66, W ,NU,DE ; KA
	//   1, L0 ,PE ; L0
	//     2, LA, 3, A ,NU ; LA
	//     2, LB, 6, P ,NU ; LB
	//     2, LC, 6, P ,NU,DE,MU; LC
	//       3, LC, 6, P ,NU,DE,MU ; LC
	//   1, MA, 4, G ,NU ; MA
	//   1, N0  ; N0
	//     2, NA, 2, U  ; NA
	//     2, NB, 3, U ,NU ; NB
	//   1, O0 ,PE ; O0
	//     2, OA, 8, U ,NU,DT=E(DATE) ; OA
	//     2, OB, 8, U ,NU,DT=E(DATE) ; OB
	//   1, PA, 3, A ,NU,DE,MU; PA
	//     2, PA, 3, A ,NU,DE,MU ; PA
	//   1, QA, 7, P  ; QA
	//   1, RA, 0, A ,NU,NV,NB ; RA
	//   1, S0 ,PE ; S0
	//     2, SA, 80, W ,NU ; SA
	//     2, SB, 3, A ,NU ; SB
	//     2, SC, 0, A ,NU,NV,NB,MU; SC
	//       3, SC, 0, A ,NU,NV,NB,MU ; SC
	//   1, TC, 20, U ,SY=TIME,DT=E(TIMESTAMP) ; TC
	//   1, TU, 20, U ,MU,SY=TIME,DT=E(TIMESTAMP); TU
	//     2, TU, 20, U ,MU,SY=TIME,DT=E(TIMESTAMP) ; TU
	//  CN,HE=COLLATING(BC,'de@collation=phonebook',PRIMAR) ; CN
	//  H1=NA(1,2),NB(1,3) ; H1
	//  S1=JA(1,2) ; S1
	//  S2=JA(1,6),BC(1,40) ; S2
	//  S3=LA(1,3),LB(1,6) ; S3
	//  HO=REFINT(A,12,A/DC) ; HO
}

func ExampleAdabas_readFileDefinition9RestrictF0() {
	err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("fdt")

	fmt.Println("Open database")
	err = adabas.Open()
	if err != nil || adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Error: ", err, " ", adabas.Acbx.Acbxrsp)
	}
	defer adabas.Close()
	fmt.Println("Read file definition")
	var definition *adatypes.Definition
	definition, err = adabas.ReadFileDefinition(9)
	if adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Resonse code : ", adabas.Acbx.Acbxrsp)
		return
	}
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	definition.ShouldRestrictToFields("AA,F0")
	definition.DumpTypes(false, true)
	// Output:Open database
	// Read file definition
	// Dump all active field types:
	//   1, A0  ; A0
	//     2, AA, 8, A ,UQ,DE,NC,NN ; AA
	//   1, F0 ,PE ; F0
	//     2, FA, 60, W ,NU,MU; FA
	//       3, FA, 60, W ,NU,MU ; FA
	//     2, FB, 40, W ,NU,DE ; FB
	//     2, FC, 10, A ,NU ; FC
	//     2, FD, 3, A ,NU ; FD
	//     2, F1  ; F1
	//       3, FE, 6, A ,NU ; FE
	//       3, FF, 15, A ,NU ; FF
	//       3, FG, 15, A ,NU ; FG
	//       3, FH, 15, A ,NU ; FH
	//       3, FI, 80, A ,NU,DE,MU; FI
	//         4, FI, 80, A ,NU,DE,MU ; FI
}

func ExampleAdabas_readFileDefinition9Restricted() {
	err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("fdt")

	fmt.Println("Open database")
	err = adabas.Open()
	if err != nil || adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Error: ", err, " ", adabas.Acbx.Acbxrsp)
	}
	defer adabas.Close()
	fmt.Println("Read file definition")
	var definition *adatypes.Definition
	definition, err = adabas.ReadFileDefinition(9)
	if adabas.Acbx.Acbxrsp != 0 {
		fmt.Println("Resonse code : ", adabas.Acbx.Acbxrsp)
		return
	}
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	definition.ShouldRestrictToFields("A0,DA,L0")
	definition.DumpTypes(false, true)
	// Output: Open database
	// Read file definition
	// Dump all active field types:
	//   1, A0  ; A0
	//     2, AA, 8, A ,UQ,DE,NC,NN ; AA
	//     2, AB  ; AB
	//       3, AC, 4, F ,DE ; AC
	//       3, AD, 8, B ,NU,HF ; AD
	//       3, AE, 0, A ,NU,NV,NB ; AE
	//   1, DA, 1, A ,FI ; DA
	//   1, L0 ,PE ; L0
	//     2, LA, 3, A ,NU ; LA
	//     2, LB, 6, P ,NU ; LB
	//     2, LC, 6, P ,NU,DE,MU; LC
	//       3, LC, 6, P ,NU,DE,MU ; LC
}

func TestAdabasFdtNewEmployee(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("newempl")

	fmt.Println("Open database")
	err := adabas.Open()
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	assert.Equal(t, nil, err)

	fmt.Println("Read file definition")
	definition, err := adabas.ReadFileDefinition(9)
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	if err != nil {
		t.Fatal("Adabas error incorrect", err)
	}
	fmt.Println(definition)

	fmt.Println("Close database")
	adabas.Close()
	fmt.Println("test done")
}

func TestAdabasFdtHyperexit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(adabasModDBID)
	adabas.ID.SetUser("hyper")
	defer adabas.Close()

	fmt.Println("Open database")
	err := adabas.Open()
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	assert.Equal(t, nil, err)

	fmt.Println("Read file definition")
	definition, err := adabas.ReadFileDefinition(50)
	if adabas.Acbx.Acbxrsp != 0 {
		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	}
	if err != nil {
		t.Fatal("Adabas error incorrect", err)
	}
	fmt.Println(definition)

	fmt.Println("test done")
}

func TestAdabasFdtNewEmployeeRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adatypes.Central.Log.Debugf("Network location %s", entireNetworkLocation())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	ID := NewAdabasID()
	adabas, uerr := NewAdabasWithID(url, ID)
	if !assert.NoError(t, uerr) {
		return
	}
	defer adabas.Close()

	adabas.ID.SetUser("newempl")

	fmt.Println("Open database")
	err := adabas.Open()
	assert.Error(t, err)
	assert.Equal(t, "ADG0000068: Entire Network client not supported, use port 0 and Entire Network native access", err.Error())
	// if assert.NoError(t, err) {
	// 	if adabas.Acbx.Acbxrsp != 0 {
	// 		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	// 	}
	// 	assert.Equal(t, uint16(0), ret)

	// 	fmt.Println("Read file definition")
	// 	definition, err := adabas.ReadFileDefinition(9)
	// 	if adabas.Acbx.Acbxrsp != 0 {
	// 		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
	// 	}

	// 	assert.NoError(t, err)
	// 	fmt.Println(definition)
	// }

	fmt.Println("test done")
}

func TestAdabasUnknownDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adatypes.Central.Log.Debugf("Network location %s", entireNetworkLocation())
	url := "201(abc://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	ID := NewAdabasID()
	adabas, uerr := NewAdabasWithID(url, ID)
	if !assert.NoError(t, uerr) {
		return
	}
	defer adabas.Close()

	err := adabas.Open()
	assert.Error(t, err)
	assert.Equal(t, "ADG0000001: Unknown network driver 'abc' given", err.Error())
}

func simpleDefinition() *adatypes.Definition {
	layout := []adatypes.IAdaType{
		adatypes.NewTypeWithLength(adatypes.FieldTypeString, "AA", 8),
	}

	testDefinition := adatypes.NewDefinitionWithTypes(layout)
	return testDefinition
}

func testParser(adabasRequest *adatypes.Request, x interface{}) (err error) {
	switch x.(type) {
	case *uint32:
		counter := x.(*uint32)
		(*counter)++
	case []uint32:
		isns := x.([]adatypes.Isn)
		for i := range isns {
			if adatypes.Isn(i) == adabasRequest.Isn {
				return
			}
		}
		err = fmt.Errorf("ISN %d not found in list", adabasRequest.Isn)
	default:
		fmt.Println(adabasRequest.Isn)
	}
	return
}

func TestAdabasReadPhysical(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	err := adabas.Open()
	if !assert.NoError(t, err) {
		return
	}
	defer adabas.Close()
	var fb bytes.Buffer
	fb.WriteString("AA.")
	req := &adatypes.Request{Option: &adatypes.BufferOption{}, Definition: simpleDefinition(),
		FormatBuffer: fb, Multifetch: 1, RecordBufferLength: 200, Parser: testParser, Limit: 5}
	counter := uint32(0)
	//, RecordBuffer: adatypes.NewHelper(make([]byte, 199), 200, binary.LittleEndian)}
	rerr := adabas.ReadPhysical(11, req, &counter)
	if !assert.NoError(t, rerr) {
		return
	}
	assert.Equal(t, uint32(5), counter)
}

func TestAdabasReadLogical(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	err := adabas.Open()
	if !assert.NoError(t, err) {
		return
	}
	defer adabas.Close()
	var fb bytes.Buffer
	fb.WriteString("AA.")
	req := &adatypes.Request{Option: &adatypes.BufferOption{}, Definition: simpleDefinition(),
		Descriptors: []string{"AA"}, FormatBuffer: fb, Multifetch: 1,
		RecordBufferLength: 200, Parser: testParser, Limit: 5}
	counter := uint32(0)
	rerr := adabas.ReadLogicalWith(11, req, &counter)
	if !assert.NoError(t, rerr) {
		return
	}
	assert.Equal(t, uint32(5), counter)
}

func TestAdabasReadIsn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	err := adabas.Open()
	if !assert.NoError(t, err) {
		return
	}
	defer adabas.Close()
	var fb bytes.Buffer
	fb.WriteString("AA.")
	req := &adatypes.Request{Option: &adatypes.BufferOption{}, Definition: simpleDefinition(),
		FormatBuffer: fb, Isn: 100, Multifetch: 1, RecordBufferLength: 200}
	counter := []adatypes.Isn{11}
	rerr := adabas.readISN(11, req, &counter)
	if !assert.NoError(t, rerr) {
		return
	}
}

func TestAdabasSearchLogical(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	err := adabas.Open()
	if !assert.NoError(t, err) {
		return
	}
	defer adabas.Close()
	var fb bytes.Buffer
	fb.WriteString("AA.")

	//	searchInfo := adatypes.NewSearchInfo(adatypes.NewPlatform(0x20), "AA=[11100110:11100115] OR AA=11100304")
	searchInfo := adatypes.NewSearchInfo(adatypes.NewPlatform(0x20), "AA=[11100110:11100115]")
	searchInfo.Definition = simpleDefinition()
	tree, terr := searchInfo.GenerateTree()
	if !assert.NoError(t, terr) {
		return
	}

	req := &adatypes.Request{Option: &adatypes.BufferOption{}, Definition: simpleDefinition(),
		Descriptors: []string{"AA"}, FormatBuffer: fb, Multifetch: 1, SearchTree: tree,
		RecordBufferLength: 200, Parser: testParser, Limit: 5}
	counter := uint32(0)
	rerr := adabas.SearchLogicalWith(11, req, &counter)
	if !assert.NoError(t, rerr) {
		return
	}
	assert.Equal(t, uint32(5), counter)
}

func TestAdabasCloned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "adabas.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(adabasModDBID)
	clonedAdabas := NewClonedAdabas(adabas)

	assert.Equal(t, adabas.ID, clonedAdabas.ID)
	assert.False(t, adabas.Acbx == clonedAdabas.Acbx)
	assert.True(t, adabas.status == clonedAdabas.status)
}
