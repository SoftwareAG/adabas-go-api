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
	"bytes"
	"fmt"
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const acbxEyecatcher = 'F' /*      F - EYECATCHER              */
const acbxVersion = '2'    /*      2 - VERSION                 */
//const eAcbxEyecatcher = 0xc6 /* EBCDIC F - EYECATCHER            */
//const eAcbxVersion = 0xf2    /* EBCDIC 2 - VERSION               */
const acbxLength = 192

// Dbid Adabas database identifier
type Dbid uint32

// Fnr Adabas file number identifier
type Fnr uint32

// Acbx Adabas Control block extended version
type Acbx struct {
	Acbxtyp  byte         /* +00  ADALNK function code        */
	Acbxrsv1 byte         /* +01  Reserved - must be 0x00     */
	Acbxver  [2]byte      /* +02  Version:                    */
	Acbxlen  uint16       /* +04  ACBX Length                 */
	Acbxcmd  [2]byte      /* +06  Command Code                */
	Acbxrsv2 uint16       /* +08  Reserved - must be 0x00     */
	Acbxrsp  uint16       /* +0A  Response code               */
	Acbxcid  [4]byte      /* +0C  Command ID                  */
	Acbxdbid Dbid         /* +10  Database ID                 */
	Acbxfnr  Fnr          /* +14  File number                 */
	Acbxisn  adatypes.Isn /* +18  ISN                         */
	Acbxisl  uint64       /* +20  ISN Lower Limit             */
	Acbxisq  uint64       /* +28  ISN Quantity                */
	Acbxcop  [8]byte      /* +30  Command option 1-8          */
	Acbxadd1 [8]byte      /* +38  Additions 1                 */
	Acbxadd2 [4]byte      /* +40  Additions 2                 */
	Acbxadd3 [8]byte      /* +44  Additions 3                 */
	Acbxadd4 [8]byte      /* +4C  Additions 4                 */
	Acbxadd5 [8]byte      /* +54  Additions 5 - (0x00)        */
	Acbxadd6 [8]byte      /* +5C  Additions 6                 */
	Acbxrsv3 [4]byte      /* +64  Reserved - must be 0x00     */
	Acbxerra uint64       /* +68  Error offset in buffer (64 bit)*/
	Acbxerrb [2]byte      /* +70  Error char field (FN)       */
	Acbxerrc uint16       /* +72  Error subcode               */
	Acbxerrd byte         /* +74  Error buffer ID             */
	Acbxerre byte         /* +75  Reserved for future use     */
	Acbxerrf uint16       /* +76  Error buffer seq num (per ID)*/
	Acbxsubr uint16       /* +78  Subcomp response code       */
	Acbxsubs uint16       /* +7A  Subcomp response subcode    */
	Acbxsubt [4]byte      /* +7C  Subcomp error text          */
	Acbxlcmp uint64       /* +80  Compressed record length    */
	/*      (negative of length if not  */
	/*      all of record read)         */
	Acbxldec uint64 /* +88  Decompressed length of all  */
	/*      returned data               */
	Acbxcmdt     uint64   /* +90  Command time                */
	Acbxuser     [16]byte /* +98  User field                  */
	Acbxsesstime uint64   /* +A8  Time, part of Adabas Session ID*/
	Acbxrsv4     [16]byte /* +B0  Reserved - must be 0x00     */
}

func newAcbx(dbid Dbid) *Acbx {
	var cb Acbx
	cb = Acbx{
		Acbxdbid: dbid,
		Acbxver:  [2]byte{acbxEyecatcher, acbxVersion},
		Acbxlen:  uint16(unsafe.Sizeof(cb)),
		Acbxcmd:  empty.code(),
		Acbxadd1: [8]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}
	cb.resetAcbx()
	return &cb
}

func resetArray(bArray []byte) {
	for i := 0; i < len(bArray); i++ {
		bArray[i] = ' '
	}
}

func (acbx *Acbx) resetCop() {
	for i := 0; i < len(acbx.Acbxcop); i++ {
		acbx.Acbxcop[i] = ' '
	}
}

func (acbx *Acbx) resetAcbx() {
	/*        memset((char*)((PACBX)pACBX),0,L_ACBX);        */
	acbx.Acbxver[0] = acbxEyecatcher
	acbx.Acbxver[1] = acbxVersion
	acbx.Acbxlen = uint16(acbxLength)
	adatypes.Central.Log.Debugf("Reset acbx ver=%v", acbx.Acbxver)
	adatypes.Central.Log.Debugf("Reset acbx cmd=%v", acbx.Acbxcmd)
	adatypes.Central.Log.Debugf("Reset acbx len=%v", acbx.Acbxlen)
	acbx.Acbxisn = 0
	acbx.Acbxisq = 0
	acbx.Acbxrsp = AdaAnact

	acbx.resetCop()
	for i := 0; i < len(acbx.Acbxadd1); i++ {
		acbx.Acbxadd1[i] = ' '
	}
	resetArray(acbx.Acbxadd2[:])
}

func (acbx *Acbx) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ACBX:\n  CmdCode: %c%c", acbx.Acbxcmd[0], acbx.Acbxcmd[1]))
	buffer.WriteString(fmt.Sprintf("  CmdId: %X\n", acbx.Acbxcid))

	buffer.WriteString(fmt.Sprintf("  Dbid: %d  Filenr: %d", acbx.Acbxdbid, acbx.Acbxfnr))
	buffer.WriteString(fmt.Sprintf("  Responsecode: %d Subcode: %d\n", acbx.Acbxrsp, acbx.Acbxerrc))
	buffer.WriteString(fmt.Sprintln("  Isn: ", acbx.Acbxisn, " ISN Lower Limit: ", acbx.Acbxisl, "ISN Quantity: ", acbx.Acbxisq))
	buffer.WriteString(adatypes.FormatBytes("  CmdOption: ", acbx.Acbxcop[:], len(acbx.Acbxcop[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add1: ", acbx.Acbxadd1[:], len(acbx.Acbxadd1[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add2: ", acbx.Acbxadd2[:], len(acbx.Acbxadd2[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add3: ", acbx.Acbxadd3[:], len(acbx.Acbxadd3[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add4: ", acbx.Acbxadd4[:], len(acbx.Acbxadd4[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add5: ", acbx.Acbxadd5[:], len(acbx.Acbxadd5[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  Add6: ", acbx.Acbxadd6[:], len(acbx.Acbxadd6[:]), 1, -1, false))
	buffer.WriteString(adatypes.FormatBytes("  User Area: ", acbx.Acbxuser[:], len(acbx.Acbxuser[:]), 0, -1, false))
	return buffer.String()
}

const adabasIDSize = 32

// AID Adabas id
type AID struct {
	level     uint16
	size      uint16
	Node      [8]byte
	User      [8]byte
	Pid       uint32
	Timestamp uint64
}

// Status of the referenced connection
type Status struct {
	ref              string
	open             bool
	openTransactions uint32
	platform         *adatypes.Platform
	adabas           *Adabas
}

// ID Adabas Id
type ID struct {
	connectionMap map[string]*Status
	AdaID         *AID
	user          string
	pwd           string
}

// SetUser set the user id name into the ID, prepare byte array correctly
func (adaid *ID) SetUser(User string) {
	for i := 0; i < 8; i++ {
		adaid.AdaID.User[i] = ' '
	}

	copy(adaid.AdaID.User[:], User)
}

// SetHost set the host id name into the ID, prepare byte array correctly
func (adaid *ID) SetHost(Host string) {
	for i := 0; i < 8; i++ {
		adaid.AdaID.Node[i] = ' '
	}

	copy(adaid.AdaID.Node[:], Host)
}

// SetID set the pid into the ID, prepare byte array correctly
func (adaid *ID) SetID(pid uint32) {
	adaid.AdaID.Pid = pid
}

// AddCredential add user id and password credentials
func (adaid *ID) AddCredential(user string, pwd string) {
	adaid.user = user
	adaid.pwd = pwd
}

// String return string representation of Adabas ID
func (adaid *ID) String() string {
	return fmt.Sprintf("%s:%s [%d] %x/%d", string(adaid.AdaID.Node[0:8]), string(adaid.AdaID.User[0:8]),
		adaid.AdaID.Pid, adaid.AdaID.Timestamp, adaid.AdaID.Timestamp)
}

func (adaid *ID) status(url string) *Status {
	if s, ok := adaid.connectionMap[url]; ok {
		return s
	}
	s := &Status{open: false}
	adaid.connectionMap[url] = s
	return s
}

func (adaid *ID) platform(url string) *adatypes.Platform {
	s := adaid.status(url)
	if s == nil {
		return nil
	}
	return s.platform
}

func (adaid *ID) changeOpenState(url string, open bool) {
	adatypes.Central.Log.Debugf("Register open=%v to url=%s", open, url)
	s := adaid.status(url)
	if s == nil {
		return
	}
	s.open = open
	if !open {
		s.openTransactions = 0
	}
}

func (adaid *ID) getAdabas(url *URL) *Adabas {
	s := adaid.status(url.String())
	if s.adabas == nil {
		NewAdabas(url, adaid)
	}
	return s.adabas
}

func (adaid *ID) setAdabas(a *Adabas) {
	s := adaid.status(a.URL.String())
	if s.adabas == nil {
		s.adabas = a
	} else {
		a.transactions = s.adabas.transactions
	}
}

func (adaid *ID) isOpen(url string) bool {
	s := adaid.status(url)
	if s == nil {
		return false
	}
	open := s.open
	adatypes.Central.Log.Debugf("Check is open=%v to url=%s", open, url)
	return open
}

func (adaid *ID) transactions(url string) uint32 {
	s := adaid.status(url)
	if s == nil {
		return 0
	}
	return s.openTransactions
}

func (adaid *ID) incTransactions(url string) {
	s := adaid.status(url)
	if s == nil {
		return
	}
	s.openTransactions++
}

func (adaid *ID) clearTransactions(url string) {
	s := adaid.status(url)
	if s == nil {
		return
	}
	s.openTransactions = 0
}

// Close close all open adabas instance for this ID
func (adaid *ID) Close() {
	for _, s := range adaid.connectionMap {
		s.adabas.Close()
	}
}
