/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdabasFailure(t *testing.T) {
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(0)

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

	log.Debug("acbx ver=", adabas.Acbx.Acbxver)
	log.Debug("acbx cmd=", adabas.Acbx.Acbxcmd)
	log.Debug("acbx len=", adabas.Acbx.Acbxlen)

	adabas.Acbx.Acbxcmd = cl.code()

	log.Debug("acbx ver=", adabas.Acbx.Acbxver)
	log.Debug("acbx cmd=", adabas.Acbx.Acbxcmd)
	log.Debug("acbx len=", adabas.Acbx.Acbxlen)

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

func TestAdabasOk23(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)

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

// const CDVTReq = uint64(0xFEDCBA98)

// func TestAdabasCompressedRemoteADATCP(t *testing.T) {
// 	f := initTestLogWithFile(t, "adabas.log")
// 	defer f.Close()

// 	log.Debug("TEST: ", t.Name())
// 	url := "111(adatcp://pctkn10:60001)"
// 	fmt.Println("Connect to ", url)
// 	adabas,err := NewAdabasWithURL(newURL(url))

// 	var abds []*Buffer
// 	abds = append(abds, NewBuffer(AbdAQFb))
// 	abds = append(abds, NewBuffer(AbdAQRb))

// 	abds[0].WriteString("A")
// 	abds[1].WriteString(".")

// 	adabas.Acbx.Acbxcmd = op.code()
// 	adabas.SetAbd(abds)

// 	retb := adabas.CallAdabas()
// 	if retb != nil {
// 		t.Fatal("Adabas call return value not correct", retb)
// 	}

// 	abds[0].Clear()
// 	abds[0].WriteString("C.")
// 	abds[1].Allocate(1024)

// 	adabas.Acbx.Acbxcmd = l1.code()
// 	adabas.Acbx.Acbxfnr = 11
// 	adabas.Acbx.Acbxisn = 1

// 	retb = adabas.CallAdabas()
// 	if retb != nil {
// 		t.Fatal("Adabas call return value not correct", retb)
// 	}

// 	adabas.Acbx.Acbxcmd = n1.code()
// 	adabas.Acbx.Acbxfnr = 111
// 	adabas.Acbx.Acbxisn = 0
// 	adabas.Acbx.Acbxisq = CDVTReq
// 	binary.LittleEndian.PutUint32(adabas.Acbx.Acbxcid[:], uint32(CDVTReq))

// 	retb = adabas.CallAdabas()
// 	if retb != nil {
// 		t.Fatal("Adabas call return value not correct", retb)
// 	}

// 	adabas.Acbx.Acbxcmd = cl.code()
// 	retb = adabas.CallAdabas()
// 	if retb != nil {
// 		t.Fatal("Adabas call return value not correct", retb)
// 	}

// 	if adabas.Acbx.Acbxrsp != 0 {
// 		t.Fatal(adabas.getAdabasMessage(), adabas.Acbx.Acbxrsp)
// 	}
// 	assert.Equal(t, uint16(0), adabas.Acbx.Acbxrsp)
// 	require.NoError(t, retb)
// 	adabas.Acbx.resetAcbx()
// }

func TestAdabasOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)

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
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	adabas.ID.setUser("fdt")

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

func ExampleAdabas_readFileDefinition11() {
	f, err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	adabas.ID.setUser("fdt")

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
	//   1, AA, 8, A ,UQ DE ; AA  PE=false MU=false REMOVE=true
	//   1, AB  ; AB  PE=false MU=false REMOVE=true
	//     2, AC, 20, A ,NU ; AC  PE=false MU=false REMOVE=true
	//     2, AE, 20, A ,DE ; AE  PE=false MU=false REMOVE=true
	//     2, AD, 20, A ,NU ; AD  PE=false MU=false REMOVE=true
	//   1, AF, 1, A ,FI ; AF  PE=false MU=false REMOVE=true
	//   1, AG, 1, A ,FI ; AG  PE=false MU=false REMOVE=true
	//   1, AH, 4, P ,DE NC ; AH  PE=false MU=false REMOVE=true
	//   1, A1  ; A1  PE=false MU=true REMOVE=true
	//     2, AI, 20, A NU MU,MU; AI  PE=false MU=true REMOVE=true
	//       3, AI, 20, A ,NU MU ; AI  PE=false MU=true REMOVE=true
	//     2, AJ, 20, A ,NU DE ; AJ  PE=false MU=true REMOVE=true
	//     2, AK, 10, A ,NU ; AK  PE=false MU=true REMOVE=true
	//     2, AL, 3, A ,NU ; AL  PE=false MU=true REMOVE=true
	//   1, A2  ; A2  PE=false MU=false REMOVE=true
	//     2, AN, 6, A ,NU ; AN  PE=false MU=false REMOVE=true
	//     2, AM, 15, A ,NU ; AM  PE=false MU=false REMOVE=true
	//   1, AO, 6, A ,DE ; AO  PE=false MU=false REMOVE=true
	//   1, AP, 25, A ,NU DE ; AP  PE=false MU=false REMOVE=true
	//   1, AQ ,PE ; AQ  PE=true MU=true REMOVE=true
	//     2, AR, 3, A ,NU ; AR  PE=true MU=true REMOVE=true
	//     2, AS, 5, P ,NU ; AS  PE=true MU=true REMOVE=true
	//     2, AT, 5, P NU MU,MU; AT  PE=true MU=true REMOVE=true
	//       3, AT, 5, P ,NU MU ; AT  PE=true MU=true REMOVE=true
	//   1, A3  ; A3  PE=false MU=false REMOVE=true
	//     2, AU, 2, U  ; AU  PE=false MU=false REMOVE=true
	//     2, AV, 2, U ,NU ; AV  PE=false MU=false REMOVE=true
	//   1, AW ,PE ; AW  PE=true MU=false REMOVE=true
	//     2, AX, 8, U ,NU ; AX  PE=true MU=false REMOVE=true
	//     2, AY, 8, U ,NU ; AY  PE=true MU=false REMOVE=true
	//   1, AZ, 3, A NU DE MU,MU; AZ  PE=false MU=true REMOVE=true
	//     2, AZ, 3, A ,NU DE MU ; AZ  PE=false MU=true REMOVE=true
	//  PH=PHON(AE) ; PH  PE=false MU=false REMOVE=true
	//  H1=AU(1-2),AV(1-2) ; H1  PE=false MU=false REMOVE=true
	//  S1=AO(1-4) ; S1  PE=false MU=false REMOVE=true
	//  S2=AO(1-6),AE(1-20) ; S2  PE=false MU=false REMOVE=true
	//  S3=AR(1-3),AS(1-9) ; S3  PE=false MU=false REMOVE=true

}

func ExampleAdabas_readFileDefinition9() {
	f, err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	adabas.ID.setUser("fdt")

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
	//   1, A0  ; A0  PE=false MU=false REMOVE=true
	//     2, AA, 8, A ,UQ DE NC NN ; AA  PE=false MU=false REMOVE=true
	//     2, AB  ; AB  PE=false MU=false REMOVE=true
	//       3, AC, 4, F ,DE ; AC  PE=false MU=false REMOVE=true
	//       3, AD, 8, B ,NU HF ; AD  PE=false MU=false REMOVE=true
	//       3, AE, 0, A ,NU NV NB ; AE  PE=false MU=false REMOVE=true
	//   1, B0  ; B0  PE=false MU=false REMOVE=true
	//     2, BA, 40, W ,NU ; BA  PE=false MU=false REMOVE=true
	//     2, BB, 40, W ,NU ; BB  PE=false MU=false REMOVE=true
	//     2, BC, 50, W ,NU DE ; BC  PE=false MU=false REMOVE=true
	//   1, CA, 1, A ,FI ; CA  PE=false MU=false REMOVE=true
	//   1, DA, 1, A ,FI ; DA  PE=false MU=false REMOVE=true
	//   1, EA, 4, P ,DE NC ; EA  PE=false MU=false REMOVE=true
	//   1, F0 ,PE ; F0  PE=true MU=true REMOVE=true
	//     2, FA, 60, W NU MU,MU; FA  PE=true MU=true REMOVE=true
	//       3, FA, 60, W ,NU MU ; FA  PE=true MU=true REMOVE=true
	//     2, FB, 40, W ,NU DE ; FB  PE=true MU=true REMOVE=true
	//     2, FC, 10, A ,NU ; FC  PE=true MU=true REMOVE=true
	//     2, FD, 3, A ,NU ; FD  PE=true MU=true REMOVE=true
	//     2, F1  ; F1  PE=true MU=true REMOVE=true
	//       3, FE, 6, A ,NU ; FE  PE=true MU=true REMOVE=true
	//       3, FF, 15, A ,NU ; FF  PE=true MU=true REMOVE=true
	//       3, FG, 15, A ,NU ; FG  PE=true MU=true REMOVE=true
	//       3, FH, 15, A ,NU ; FH  PE=true MU=true REMOVE=true
	//       3, FI, 80, A NU DE MU,MU; FI  PE=true MU=true REMOVE=true
	//         4, FI, 80, A ,NU DE MU ; FI  PE=true MU=true REMOVE=true
	//   1, I0 ,PE ; I0  PE=true MU=true REMOVE=true
	//     2, IA, 40, W NU MU,MU; IA  PE=true MU=true REMOVE=true
	//       3, IA, 40, W ,NU MU ; IA  PE=true MU=true REMOVE=true
	//     2, IB, 40, W ,NU DE ; IB  PE=true MU=true REMOVE=true
	//     2, IC, 10, A ,NU ; IC  PE=true MU=true REMOVE=true
	//     2, ID, 3, A ,NU ; ID  PE=true MU=true REMOVE=true
	//     2, IE, 5, A ,NU ; IE  PE=true MU=true REMOVE=true
	//     2, I1  ; I1  PE=true MU=true REMOVE=true
	//       3, IF, 6, A ,NU ; IF  PE=true MU=true REMOVE=true
	//       3, IG, 15, A ,NU ; IG  PE=true MU=true REMOVE=true
	//       3, IH, 15, A ,NU ; IH  PE=true MU=true REMOVE=true
	//       3, II, 15, A ,NU ; II  PE=true MU=true REMOVE=true
	//       3, IJ, 80, A NU DE MU,MU; IJ  PE=true MU=true REMOVE=true
	//         4, IJ, 80, A ,NU DE MU ; IJ  PE=true MU=true REMOVE=true
	//   1, JA, 6, A ,DE ; JA  PE=false MU=false REMOVE=true
	//   1, KA, 66, W ,NU DE ; KA  PE=false MU=false REMOVE=true
	//   1, L0 ,PE ; L0  PE=true MU=true REMOVE=true
	//     2, LA, 3, A ,NU ; LA  PE=true MU=true REMOVE=true
	//     2, LB, 6, P ,NU ; LB  PE=true MU=true REMOVE=true
	//     2, LC, 6, P NU DE MU,MU; LC  PE=true MU=true REMOVE=true
	//       3, LC, 6, P ,NU DE MU ; LC  PE=true MU=true REMOVE=true
	//   1, MA, 4, G ,NU ; MA  PE=false MU=false REMOVE=true
	//   1, N0  ; N0  PE=false MU=false REMOVE=true
	//     2, NA, 2, U  ; NA  PE=false MU=false REMOVE=true
	//     2, NB, 3, U ,NU ; NB  PE=false MU=false REMOVE=true
	//   1, O0 ,PE ; O0  PE=true MU=false REMOVE=true
	//     2, OA, 8, U ,NU DT=E(DATE) ; OA  PE=true MU=false REMOVE=true
	//     2, OB, 8, U ,NU DT=E(DATE) ; OB  PE=true MU=false REMOVE=true
	//   1, PA, 3, A NU DE MU,MU; PA  PE=false MU=true REMOVE=true
	//     2, PA, 3, A ,NU DE MU ; PA  PE=false MU=true REMOVE=true
	//   1, QA, 7, P  ; QA  PE=false MU=false REMOVE=true
	//   1, RA, 0, A ,NU NV NB ; RA  PE=false MU=false REMOVE=true
	//   1, S0 ,PE ; S0  PE=true MU=true REMOVE=true
	//     2, SA, 80, W ,NU ; SA  PE=true MU=true REMOVE=true
	//     2, SB, 3, A ,NU ; SB  PE=true MU=true REMOVE=true
	//     2, SC, 0, A NU NV NB MU,MU; SC  PE=true MU=true REMOVE=true
	//       3, SC, 0, A ,NU NV NB MU ; SC  PE=true MU=true REMOVE=true
	//   1, TC, 20, U ,SY=TIME DT=E(TIMESTAMP) ; TC  PE=false MU=false REMOVE=true
	//   1, TU, 20, U MU SY=TIME DT=E(TIMESTAMP),MU; TU  PE=false MU=true REMOVE=true
	//     2, TU, 20, U ,MU SY=TIME DT=E(TIMESTAMP) ; TU  PE=false MU=true REMOVE=true
	//  CN,HE=COLLATING(BC,'de@collation=phonebook',PRIMAR) ; CN  PE=false MU=false REMOVE=true
	//  H1=NA(1-2),NB(1-3) ; H1  PE=false MU=false REMOVE=true
	//  S1=JA(1-2) ; S1  PE=false MU=false REMOVE=true
	//  S2=JA(1-6),BC(1-40) ; S2  PE=false MU=false REMOVE=true
	//  S3=LA(1-3),LB(1-6) ; S3  PE=false MU=false REMOVE=true
	//  HO=REFINT(A,12,A/DC) ; HO  PE=false MU=false REMOVE=true
}

func ExampleAdabas_readFileDefinition9RestrictF0() {
	f, err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	adabas.ID.setUser("fdt")

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
	//   1, A0  ; A0  PE=false MU=false REMOVE=true
	//     2, AA, 8, A ,UQ DE NC NN ; AA  PE=false MU=false REMOVE=false
	//   1, F0 ,PE ; F0  PE=true MU=true REMOVE=false
	//     2, FA, 60, W NU MU,MU; FA  PE=true MU=true REMOVE=false
	//       3, FA, 60, W ,NU MU ; FA  PE=true MU=true REMOVE=false
	//     2, FB, 40, W ,NU DE ; FB  PE=true MU=true REMOVE=false
	//     2, FC, 10, A ,NU ; FC  PE=true MU=true REMOVE=false
	//     2, FD, 3, A ,NU ; FD  PE=true MU=true REMOVE=false
	//     2, F1  ; F1  PE=true MU=true REMOVE=false
	//       3, FE, 6, A ,NU ; FE  PE=true MU=true REMOVE=false
	//       3, FF, 15, A ,NU ; FF  PE=true MU=true REMOVE=false
	//       3, FG, 15, A ,NU ; FG  PE=true MU=true REMOVE=false
	//       3, FH, 15, A ,NU ; FH  PE=true MU=true REMOVE=false
	//       3, FI, 80, A NU DE MU,MU; FI  PE=true MU=true REMOVE=false
	//         4, FI, 80, A ,NU DE MU ; FI  PE=true MU=true REMOVE=false
}

func ExampleAdabas_readFileDefinition9Restricted() {
	f, err := initLogWithFile("adabas.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	adabas := NewAdabas(23)
	adabas.ID.setUser("fdt")

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
	//   1, A0  ; A0  PE=false MU=false REMOVE=false
	//     2, AA, 8, A ,UQ DE NC NN ; AA  PE=false MU=false REMOVE=false
	//     2, AB  ; AB  PE=false MU=false REMOVE=false
	//       3, AC, 4, F ,DE ; AC  PE=false MU=false REMOVE=false
	//       3, AD, 8, B ,NU HF ; AD  PE=false MU=false REMOVE=false
	//       3, AE, 0, A ,NU NV NB ; AE  PE=false MU=false REMOVE=false
	//   1, DA, 1, A ,FI ; DA  PE=false MU=false REMOVE=false
	//   1, L0 ,PE ; L0  PE=true MU=true REMOVE=false
	//     2, LA, 3, A ,NU ; LA  PE=true MU=true REMOVE=false
	//     2, LB, 6, P ,NU ; LB  PE=true MU=true REMOVE=false
	//     2, LC, 6, P NU DE MU,MU; LC  PE=true MU=true REMOVE=false
	//       3, LC, 6, P ,NU DE MU ; LC  PE=true MU=true REMOVE=false
}

func TestAdabasFdtNewEmployee(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	adabas.ID.setUser("newempl")

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
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(23)
	adabas.ID.setUser("hyper")
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
	f := initTestLogWithFile(t, "adabas.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	log.Debug("Network location ", entireNetworkLocation())
	url := "201(tcpip://" + entireNetworkLocation() + ")"
	fmt.Println("Connect to ", url)
	ID := NewAdabasID()
	adabas, uerr := NewAdabasWithID(url, ID)
	if !assert.NoError(t, uerr) {
		return
	}
	defer adabas.Close()

	adabas.ID.setUser("newempl")

	fmt.Println("Open database")
	err := adabas.Open()
	assert.Error(t, err)
	assert.Equal(t, "Entire Network client not supported, use port 0 and Entire Network native access", err.Error())
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
