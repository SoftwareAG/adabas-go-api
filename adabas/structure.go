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
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const acbxEyecatcher = 'F'   /*      F - EYECATCHER              */
const acbxVersion = '2'      /*      2 - VERSION                 */
const eAcbxEyecatcher = 0xc6 /* EBCDIC F - EYECATCHER            */
const eAcbxVersion = 0xf2    /* EBCDIC 2 - VERSION               */
const acbxLength = 192

// Dbid Adabas database identifier
type Dbid uint32

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
	Acbxfnr  uint32       /* +14  File number                 */
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

type commandCode uint

const (
	empty commandCode = iota
	op
	cl
	bt
	et
	lf
	l1
	l2
	l3
	l4
	l5
	l6
	l9
	n1
	n2
	a1
	s1
	s2
	s3
	e1
	u1
	u2
	u3
	rc
)

var commandCodes = []string{"  ", "OP", "CL", "BT", "ET", "LF", "L1", "L2", "L3", "L4", "L5", "L6", "L9",
	"N1", "N2", "A1", "S1", "S2", "S3", "E1", "U1", "U2", "U3", "RC"}

func (cc commandCode) code() [2]byte {
	var code [2]byte
	codeConst := []byte(commandCodes[cc])
	copy(code[:], codeConst[0:2])
	return code
}

func (cc commandCode) command() string {
	return commandCodes[cc]
}

func validAcbxCommand(cmd [2]byte) bool {
	checkCmd := strings.ToUpper(string(cmd[:]))
	for _, validCmd := range commandCodes {
		if validCmd == checkCmd {
			return true
		}
	}
	return false
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
	acbx.Acbxrsp = 148

	acbx.resetCop()
	for i := 0; i < len(acbx.Acbxadd1); i++ {
		acbx.Acbxadd1[i] = ' '
	}
	resetArray(acbx.Acbxadd2[:])
	/*    memcpy(((PACBX)pACBX)->acbxcmd, "  "  , sizeof(((PACBX)pACBX)->acbxcmd));
	      memcpy(((PACBX)pACBX)->acbxcid, "    ", sizeof(((PACBX)pACBX)->acbxcid));
	      ((PACBX)pACBX)->acbxcop1   = ' ';
	      ((PACBX)pACBX)->acbxcop2   = ' ';
	      ((PACBX)pACBX)->acbxcop3   = ' ';
	      ((PACBX)pACBX)->acbxcop4   = ' ';
	      ((PACBX)pACBX)->acbxcop5   = ' ';
	      ((PACBX)pACBX)->acbxcop6   = ' ';
	      ((PACBX)pACBX)->acbxcop7   = ' ';
	      ((PACBX)pACBX)->acbxcop8   = ' ';
	      memcpy(((PACBX)pACBX)->acbxadd1, "        ", sizeof(((PACBX)pACBX)->acbxadd1));
	      memcpy(((PACBX)pACBX)->acbxadd2, "    "    , sizeof(((PACBX)pACBX)->acbxadd2));
	      memcpy(((PACBX)pACBX)->acbxadd3, "        ", sizeof(((PACBX)pACBX)->acbxadd3));
	      memcpy(((PACBX)pACBX)->acbxadd4, "        ", sizeof(((PACBX)pACBX)->acbxadd4));
	      memcpy(((PACBX)pACBX)->acbxadd6, "        ", sizeof(((PACBX)pACBX)->acbxadd6));
	      memcpy(((PACBX)pACBX)->acbxerrb, "  "      , sizeof(((PACBX)pACBX)->acbxerrb));
	      ((PACBX)pACBX)->acbxerrd   = ' ';
	      ((PACBX)pACBX)->acbxerre   = ' ';
	      ((PACBX)pACBX)->acbxerrf   = 0;
	      memcpy(((PACBX)pACBX)->acbxsubt, "    "    , sizeof(((PACBX)pACBX)->acbxsubt));
	      memset(((PACBX)pACBX)->acbxuser, ' ', sizeof(((PACBX)pACBX)->acbxuser));
	*/
}

func (acbx *Acbx) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ACBX:\n  CmdCode: %c%c", acbx.Acbxcmd[0], acbx.Acbxcmd[1]))
	buffer.WriteString(fmt.Sprintf("  CmdId: %X\n", acbx.Acbxcid))

	buffer.WriteString(fmt.Sprintf("  Dbid: %d  Filenr: %d", acbx.Acbxdbid, acbx.Acbxfnr))
	buffer.WriteString(fmt.Sprintf("  Responsecode: %d Subcode: %d\n", acbx.Acbxrsp, acbx.Acbxerrc))
	buffer.WriteString(fmt.Sprintln("  Isn: ", acbx.Acbxisn, " ISN Lower Limit: ", acbx.Acbxisl, "ISN Quantity: ", acbx.Acbxisq))
	buffer.WriteString(adatypes.FormatBytes("  CmdOption: ", acbx.Acbxcop[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add1: ", acbx.Acbxadd1[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add2: ", acbx.Acbxadd2[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add3: ", acbx.Acbxadd3[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add4: ", acbx.Acbxadd4[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add5: ", acbx.Acbxadd5[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  Add6: ", acbx.Acbxadd6[:], 1, -1))
	buffer.WriteString(adatypes.FormatBytes("  User Area: ", acbx.Acbxuser[:], 0, -1))
	return buffer.String()
}

/*
* Internal constants providing various configurations for the Adabas buffer
* block.
 */
const abdEyecatcher = 'G' /*      G - EYECATCHER              */
const abdVersion = '2'    /*      2 - VERSION                 */
//const E_ABD_EYECATCHER = 0xc7 /* EBCDIC G - EYECATCHER            */
//const E_ABD_VERSION = 0xf2    /* EBCDIC 2 - VERSION               */
const (
	AbdAQFb  = ('F') /*      F-Format Buffer             */
	AbdAQRb  = ('R') /*      R-Record Buffer             */
	AbdAQSb  = ('S') /*      S-Search Buffer             */
	AbdAQVb  = ('V') /*      V-Value Buffer              */
	AbdAQIb  = ('I') /*      I-ISN Buffer                */
	AbdAQPb  = ('P') /*      Performance Buffer          */
	AbdAQMb  = ('M') /*      Multifetch  Buffer          */
	AbdAQUi  = ('U') /*      U-User Info                 */
	AbdAQOb  = ('O') /*      I/O Buffer (internal)       */
	AbdAQXb  = ('X') /*      CLEX Info Buffer (internal) */
	AbdAQZb  = ('Z') /*      Security Buffer (internal)  */
	AbdEQFb  = 0xc6  /* EBCDIC F-Format Buffer           */
	AbdEQRb  = 0xd9  /* EBCDIC R-Record Buffer           */
	AbdEQSb  = 0xe2  /* EBCDIC S-Search Buffer           */
	AbdEQVb  = 0xe5  /* EBCDIC V-Value Buffer            */
	AbdEQIb  = 0xc9  /* EBCDIC I-ISN Buffer              */
	AbdEQPb  = 0xd7  /* EBCDIC Performance Buffer        */
	AbdEQMb  = 0xd4  /* EBCDIC Multifetch  Buffer        */
	AbdEQUi  = 0xe4  /* EBCDIC User Info                 */
	AbdEQOb  = 0xd6  /* EBCDIC I/O Buffer (internal)     */
	ABdEQXb  = 0xe7  /* EBCDIC CLEX Info Buffer          */
	AbdEQZb  = 0xe9  /* EBCDIC Security Buffer           */
	abdQStd  = (' ') /*      ' ' -at end of ABD (std)    */
	abdQInd  = ('I') /*      I   -indirectly addressed   */
	eAbdQStd = 0x40  /* EBCDIC ' ' at end of ABD (std)   */
	eABdQInd = 0xc9  /* EBCDIC I  indirectly addressed   */
)
const abdLength = 48

// Abd Adabas Buffer definition. Representation of ABD structure in the GO environment.
type Abd struct {
	Abdlen  uint16  /* +00  ABD Length                  */
	Abdver  [2]byte /* +02  Version:                    */
	Abdid   byte    /* +04  Buffer ID:                  */
	Abdrsv1 byte    /* +05  Reserved - must be 0x00     */
	Abdloc  byte    /* +06  Buffer location flag:       */
	Abdrsv2 [9]byte /* +07  Reserved - must be 0x00     */
	Abdsize uint64  /* +10  Buffer Size                 */
	Abdsend uint64  /* +18  Len to send to database     */
	Abdrecv uint64  /* +20  Len received from database  */

	Abdaddr uint64 /* +28  8 byte aligned 64bit Ptr    */
	/*      indirectly to buffer        */
}

// Buffer Adabas Buffer overlay to combine the buffer itself with
// the Adabas buffer definition. It includes the current offset
// of the buffer.
type Buffer struct {
	abd    Abd
	offset int
	buffer []byte
}

// NewBuffer Create a new buffer with given id
func NewBuffer(id byte) *Buffer {
	return &Buffer{
		abd:    Abd{Abdver: [2]byte{abdEyecatcher, abdVersion}, Abdlen: abdLength, Abdid: id, Abdloc: abdQInd},
		offset: 0,
	}
}

// If needed, grow the buffer size to new size given
func (adabasBuffer *Buffer) grow(newSize int) {
	adatypes.Central.Log.Debugf("Current %c buffer to %d,%d", adabasBuffer.abd.Abdid, len(adabasBuffer.buffer), cap(adabasBuffer.buffer))
	adatypes.Central.Log.Debugf("Resize buffer to %d", newSize)
	newBuffer := make([]byte, newSize)
	copy(newBuffer, adabasBuffer.buffer)
	adabasBuffer.buffer = newBuffer
	adatypes.Central.Log.Debugf("Growed buffer len=%d cap=%d", len(adabasBuffer.buffer), cap(adabasBuffer.buffer))
	adabasBuffer.abd.Abdsize = uint64(len(adabasBuffer.buffer))
}

// WriteString write string intp buffer
func (adabasBuffer *Buffer) WriteString(content string) {
	adatypes.Central.Log.Debugf("Write string in adabas buffer")
	end := adabasBuffer.offset + len(content)
	if adabasBuffer.offset+len(content) > cap(adabasBuffer.buffer) {
		adabasBuffer.grow(adabasBuffer.offset + len(content))
		adabasBuffer.abd.Abdsize = uint64(adabasBuffer.offset + len(content))
	}
	copy(adabasBuffer.buffer[adabasBuffer.offset:end], content)
	adabasBuffer.offset += len(content)
	adabasBuffer.abd.Abdsend = uint64(adabasBuffer.offset)
}

// WriteBinary write binary slice into buffer
func (adabasBuffer *Buffer) WriteBinary(content []byte) {
	adatypes.Central.Log.Debugf("Write binary in adabas buffer")
	end := adabasBuffer.offset + len(content)
	if adabasBuffer.offset+len(content) > cap(adabasBuffer.buffer) {
		adabasBuffer.grow(end)
		adabasBuffer.abd.Abdsize = uint64(end)
	}

	// Copy content into buffer
	adatypes.Central.Log.Debugf("Copy to range", adabasBuffer.offset, end, len(adabasBuffer.buffer), cap(adabasBuffer.buffer))
	copy(adabasBuffer.buffer[adabasBuffer.offset:], content[:])
	adabasBuffer.offset += len(content)
	adabasBuffer.abd.Abdsend = uint64(adabasBuffer.offset)
}

// Allocate allocate buffer of specified size
func (adabasBuffer *Buffer) Allocate(size uint32) {
	if adabasBuffer.buffer == nil || size != uint32(len(adabasBuffer.buffer)) {
		adabasBuffer.buffer = make([]byte, size)
		adabasBuffer.abd.Abdsize = uint64(len(adabasBuffer.buffer))
	}
}

// Bytes receive buffer content
func (adabasBuffer *Buffer) Bytes() []byte {
	return adabasBuffer.buffer
}

// Position offset to another position in the buffer
func (adabasBuffer *Buffer) position(pos int) int {
	switch {
	case pos < 0:
		adabasBuffer.offset = 0
	case pos > len(adabasBuffer.buffer):
		adabasBuffer.offset = len(adabasBuffer.buffer)
	default:
		adabasBuffer.offset = pos
	}
	return adabasBuffer.offset
}

// Received Number of received bytes
func (adabasBuffer *Buffer) Received() uint64 {
	return adabasBuffer.abd.Abdrecv
}

// Clear buffer emptied
func (adabasBuffer *Buffer) Clear() {
	adabasBuffer.buffer = nil
	adabasBuffer.offset = 0
	adabasBuffer.abd.Abdsize = 0
	adabasBuffer.abd.Abdsend = 0
	adabasBuffer.abd.Abdrecv = 0
}

// String common string representation of the Adabas buffer
func (adabasBuffer *Buffer) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ABD ID: %c\n  Size: %d\n", adabasBuffer.abd.Abdid, adabasBuffer.abd.Abdsize))
	buffer.WriteString(fmt.Sprintf(" Send: %d  Received: %d\n", adabasBuffer.abd.Abdsend, adabasBuffer.abd.Abdrecv))
	buffer.WriteString(adatypes.FormatByteBuffer("Buffer", adabasBuffer.buffer))
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
	ref      string
	open     bool
	platform *adatypes.Platform
}

// ID Adabas Id
type ID struct {
	connectionMap map[string]*Status
	adaID         *AID
}

func (adaid *ID) setUser(User string) {
	for i := 0; i < 8; i++ {
		adaid.adaID.User[i] = ' '
	}

	copy(adaid.adaID.User[:], User)
}

func (adaid *ID) setHost(Host string) {
	for i := 0; i < 8; i++ {
		adaid.adaID.Node[i] = ' '
	}

	copy(adaid.adaID.Node[:], Host)
}

func (adaid *ID) setID(pid uint32) {
	adaid.adaID.Pid = pid
}

// String return string representation of Adabas ID
func (adaid *ID) String() string {
	return fmt.Sprintf("%s:%s [%d] %x", string(adaid.adaID.Node[0:8]), string(adaid.adaID.User[0:8]),
		adaid.adaID.Pid, adaid.adaID.Timestamp)
}

func (adaid *ID) connection(url string) *Status{
	if s,ok:= adaid.connectionMap[url];ok {
		return s
	}
	s := &Status{open: false}
	adaid.connectionMap[url] = s
	return s
}

func (adaid *ID) changePlatform(url string, platform *adatypes.Platform) {
	adaid.connection(url).platform = platform
}

func (adaid *ID) platform(url string) *adatypes.Platform {
	return adaid.connection(url).platform
}

func (adaid *ID) changeOpenState(url string, open bool) {
	adaid.connection(url).open = open
}

func (adaid *ID) isOpen(url string) bool {
	return adaid.connection(url).open
}