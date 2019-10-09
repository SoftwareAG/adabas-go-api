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

import "strings"

const (
	// AdaNormal Normal successful completion (Adabas response code 0)
	AdaNormal = 0

	// AdaISNNotSorted ISN list not sorted (Adabas response code 1)
	AdaISNNotSorted = 1

	// AdaFuncMP Function not completely executed (Adabas response code 2)
	AdaFuncMP = 2

	// AdaEOF Adabas End of File reached (Adabas response code 3)
	AdaEOF = 3

	// AdaSExpandedFiles S2/S9 is not allowed for expanded files
	AdaSExpandedFiles = 4

	// AdaVCompression Error in system view compression
	AdaVCompression = 5

	// AdaSXInterrupted SX command interrupted because of timeout
	AdaSXInterrupted = 7

	// AdaTransactionAborted Transaction aborted (Adabas response code 9)
	AdaTransactionAborted = 9

	// AdaTooManyOccurrencesPeriod Too many occurrences for a periodic group
	AdaTooManyOccurrencesPeriod = 10

	// AdaDDLCommandFailed DDL command has failed (Adabas response code 15)
	AdaDDLCommandFailed = 15

	/* ----- used only for internal interface */

	// AdaSubCommandFailed A subcommand of the MC call has failed (Adabas response code 16)
	AdaSubCommandFailed = 16

	// AdaInvalidFileNumber Invalid or unauthorized file number (Adabas response code 17)
	AdaInvalidFileNumber = 17

	// AdaFileChanged File number changed during command sequence (Adabas response code 18)
	AdaFileChanged = 18

	// AddACCCNotAllowed Command not allowed for ACC user (Adabas response code 19)
	AddACCCNotAllowed = 19

	// AdaInvalidCID Invalid command identification (CID) value (Adabas response code 20)
	AdaInvalidCID = 20

	// AdaInconsistentCmd Inconsistent usage of a command (Adabas response code 21)
	AdaInconsistentCmd = 21

	// AdaCmdInvalid Invalid command code (Adabas response code 22)
	AdaCmdInvalid = 22

	// AdaInvalidStartISN Invalid ISN starting value for L2/L5 (Adabas response code 23)
	AdaInvalidStartISN = 23

	// AdaInvalidIsnBuf Invalid ISN found in ISN-buffer (Adabas response code 24)
	AdaInvalidIsnBuf = 24

	// AdaISNLL ISN specified in ISN-LL for subsequent S1/S2 not found (Adabas response
	// code 25)
	AdaISNLL = 25

	// AdaInvalidISNBufferLength Invalid ISN-buffer length or invalid ISN-Quantity (Adabas response code
	// 26)
	AdaInvalidISNBufferLength = 26

	// AdaLPWSmall LWP parameter too small (for given SBL/VBL) (Adabas response code 28)
	AdaLPWSmall = 27

	// AdaInvalidADD1 Invalid ADDITION-1 contents for L3/L6/S9 (Adabas response code 28)
	AdaInvalidADD1 = 28

	// AdaMissingVOPT Missing V option during forced value start during L3/L6 (Adabas response
	// code 29)
	AdaMissingVOPT = 29

	// AdaInvalidCOP An invalid command option has been detected.
	AdaInvalidCOP = 34

	// AdaSyntax Syntax error in format buffer (Adabas response code 40)
	AdaSyntax = 40

	// AdaErrorFB Error in format buffer (Adabas response code 41)
	AdaErrorFB = 41

	// AdaIFBSmall Internal format buffer too small to store format (Adabas response code 42)
	AdaIFBSmall = 42

	// AdaIncosistentDE Inconsistent Descriptor definition for L9 (Adabas response code 43)
	AdaIncosistentDE = 43

	// AdaFBNotUsableUpdate Format buffer cannot be used for update (Adabas response code 44)
	AdaFBNotUsableUpdate = 44

	// AdaFieldCountOverflow Field count for PE or MU overflowed when using N-option for update
	// (Adabas response code 45)
	AdaFieldCountOverflow = 45

	// AdaMismatchFB Mismatch of format buffer usage for supplied command ID (Adabas response
	// code 46)
	AdaMismatchFB = 46

	// AdaHoldIsnOverflow Maximum number of ISNs held by a single user are reached (Adabas response
	// code 48)
	AdaHoldIsnOverflow = 47

	// AdaFUNotAvaiable File(s) / user ID not available at open time (Adabas response code 48)
	AdaFUNotAvaiable = 48

	// AdaCompressTooLong Compressed record too long (Adabas response code 49)
	AdaCompressTooLong = 49

	// AdaSYRBO Syntax error in record buffer for open (Adabas response code 50)
	AdaSYRBO = 50

	// AdaInvalidRBOpen Invalid record buffer Contents During Open (Adabas response code 51)
	AdaInvalidRBOpen = 51

	// AdaInvalidRBVB Invalid data in record buffer or value buffer (Adabas response code 52)
	AdaInvalidRBVB = 52

	// AdaRbts Record buffer too short (Adabas response code 53)
	AdaRbts = 53

	// AdaRbtl Record buffer too long for C3,C5,ET (Adabas response code 54)
	AdaRbtl = 54

	// AdaIncompFCTE Incompatible format conversion or truncation error (Adabas response code 55)
	AdaIncompFCTE = 55

	// AdaDescrLong Descriptor value too long (Adabas response code 56)
	AdaDescrLong = 56

	// AdaDSpec Unknown Descriptor specification in search buffer for L9 (Adabas response
	// code 57)
	AdaDSpec = 57

	// AdaFNFCR Format not found according to selection criterion (Adabas response code
	// 58)
	AdaFNFCR = 58

	// AdaFCONVSUB Format conversion for subfield not possible (Adabas response code 59)
	AdaFCONVSUB = 59

	// AdaSYSBU Syntax error in search buffer (Adabas response code 60)
	AdaSYSBU = 60

	// AdaERSBU Error in search buffer (Adabas response code 61)
	AdaERSBU = 61

	// AdaLSPEC Inconsistent length specification in search and value buffer (Adabas
	// response code 62)
	AdaLSPEC = 62

	// AdaUCIDS Unknown command identification (CID) in search buffer (Adabas response
	// code 63)
	AdaUCIDS = 63

	// AdaUAOS Error in communication with Adabas utilities or Adabas Online System
	// (AOS) (Adabas response code 64)
	AdaUAOS = 64

	// AdaSCALERR Space calculation error (Adabas response code 65)
	AdaSCALERR = 65

	// AdaICNF Invalid client number specification (Adabas response code 66)
	AdaICNF = 66

	// AdaIEDEC Internal error during decompressing of superfields (Adabas response code
	// 67)
	AdaIEDEC = 67

	// AdaNDSOFF Nondescriptor search issued though facility is off (Adabas response code
	// 78)
	AdaNDSOFF = 68
	// AdaNSSC No space in table of sequential commands (Adabas response code 70)
	AdaNSSC = 70
	// AdaNSSR No space in table of search results (Adabas response code 71)
	AdaNSSR = 71

	// AdaNSUQU No space available for user in user queue (Adabas response code 72)
	AdaNSUQU = 72

	// AdaNSWRK No space available for search result in WORK (Adabas response code 73)
	AdaNSWRK = 73

	// AdaNTWRK No temporary space on WORK for search command (Adabas response code 74)
	AdaNTWRK = 74

	// AdaEXOVFCB Extent overflow in File Control Block (FCB) (Adabas response code 75)
	AdaEXOVFCB = 75

	// AdaOVIDX An overflow occured in an inverted list index (Adabas response code 76)
	AdaOVIDX = 76

	// AdaNSAAD No Space available for ASSO/DATA (Adabas response code 77)
	AdaNSAAD = 77

	// AdaOVFST Free Space Table (FST) overflow (Adabas response code 78)
	AdaOVFST = 78

	// AdaHYXNA Hyperdescriptor not available (Adabas response code 79)
	AdaHYXNA = 79

	// AdaHYISNMF MF: Invalid ISN from hyperexit (Adabas response code 83)
	AdaHYISNMF = 82

	// AdaHYISN OS: Invalid ISN from hyperexit, MF: A hypertable overflow occurred.
	// (Adabas response code 83)
	AdaHYISN = 83

	// AdaWOSUB Workpool overflow during sub/super update (Adabas response code 84)
	AdaWOSUB = 84

	// AdaOVDVT DVT overflow during update command (Adabas response code 85)
	AdaOVDVT = 85

	// AdaHYPERR Hyperdescriptor error (Adabas response code 86)
	AdaHYPERR = 86

	// AdaBPLOCK Hyperdescriptor error (Adabas response code 87)
	AdaBPLOCK = 87

	// AdaINMEM Insufficient memory (Adabas response code 88)
	AdaINMEM = 88

	// AdaUNIQD Unique descriptor already present (Adabas response code 98)
	AdaUNIQD = 98

	// AdaIOERR I/O error (Adabas response code 99)
	AdaIOERR = 99

	// AdaINVIS Invalid ISN for HI,N2 or L1/L4 (Adabas response code 113)
	AdaINVIS = 113

	// AdaINVRF Refresh file not permitted (Adabas response code 114)
	AdaINVRF = 114

	// AdaLOBERR Internal error during LOB file processing (Adabas response code 132)
	AdaLOBERR = 132

	// AdaNLOCK ISN to be updated not held by user (Adabas response code 144)
	AdaNLOCK = 144

	// AdaALOCK ISN already held by some other user (Adabas response code 145)
	AdaALOCK = 145

	// AdaBSPEC Invalid buffer length specification (Adabas response code 146)
	AdaBSPEC = 146

	// AdaUBNAC User buffer not accessible (Adabas response code 147)
	AdaUBNAC = 147

	// AdaAnact Adabas is not active or accessible (Adabas response code 148)
	AdaAnact = 148

	// AdaSysCe System communication error (Adabas response code 149)
	AdaSysCe = 149

	// AdaNUCLI Too many nuclei used in parallel (Adabas response code 150)
	AdaNUCLI = 150

	// AdaNSACQ No space available in command queue (Adabas response code 151)
	AdaNSACQ = 151

	// AdaIUBSZ User buffer greater than IUB size (Adabas response code 152)
	AdaIUBSZ = 152

	// AdaPending Adabas call already pending (Adabas response code 153)
	AdaPending = 153

	// AdaCancel Adabas call canceled (Adabas response code 154)
	AdaCancel = 154

	// AdaBPMFU All buffer pool space is used (Adabas response code 162)
	AdaBPMFU = 162

	// AdaNODESC Error in inverted list - descriptor not found (Adabas response code 165)
	AdaNODESC = 165

	// AdaNODV Error in inverted list - DV not found (Adabas response code 166)
	AdaNODV = 166

	// AdaUQDV Error in inverted list - DV already present (Adabas response code 167)
	AdaUQDV = 167

	// AdaINRAB Invalid RABN (Adabas response code 170)
	AdaINRAB = 170

	// AdaISNVAL ISN value invalid (ISN=0 or ISN&gt;MAXISN) (Adabas response code 172)
	AdaISNVAL = 172

	// AdaDARAB Invalid DATA RABN (Adabas response code 173)
	AdaDARAB = 173

	// AdaINVLIST Error in inverted list (Adabas response code 176)
	AdaINVLIST = 176

	// AdaMISAC Record cannot be located in DATA storage block as indicated by AC (Adabas
	// response code 177)
	AdaMISAC = 177

	// AdaETDAT Necessary ET-data was not found in appropriate WORK block (Adabas
	// response code 182)
	AdaETDAT = 182

	/* === rsp codes for referential integrity === */
	/* ============================================ */

	// AdaSECUR Security violation (Adabas response code 200)
	AdaSECUR = 200
	// AdaINVPWD Invalid password (Adabas response code 201)
	AdaINVPWD = 201
	// AdaNFPWD Invalid password for used file (Adabas response code 202)
	AdaNFPWD = 202
	// AdaPWDINU Password already in use (Adabas response code 204)
	AdaPWDINU = 204
	// AdaSAF SAF security login required (Adabas response code 208)
	AdaSAF = 208
	// AdaINVUSR SAF security invalid user (Adabas response code 208)
	AdaINVUSR = 209
	// AdaBLOST receive buffer lost (Adabas response code 210)
	AdaBLOST = 210
	// AdaRMUTI Only local utility usage allowed (Adabas response code 211)
	AdaRMUTI = 211
	// AdaNOTYET Functionality not yet implemented (Adabas response code 212)
	AdaNOTYET = 212
	// AdaLNKERR This response is issued by an Adabas link routine (Adabas response code
	// 228)
	AdaLNKERR = 228
	// AdaTIMEOUT Connection timeout (Adabas response code 224)
	AdaTIMEOUT = 224
	// AdaXAProtocol Mismatch in the calling protocol (Adabas response code 230)
	AdaXAProtocol = 230

	// AdaLODUEX User exit / SPT load error (Adabas response code 241)
	AdaLODUEX = 241
	// AdaALLOC Double allocation error. (Adabas response code 242)
	AdaALLOC = 242
	// AdaGCBEX Invalid GCB / FCB extent detected (Adabas response code 243)
	AdaGCBEX = 243
	// AdaUTUCB Pending utility entries in UCB (Adabas response code 245)
	AdaUTUCB = 245

	// AdaOVUCB Utility communicaton block (UCB) overflow (Adabas response code 246)
	AdaOVUCB = 246
	// AdaIDUCB Correct Ident not found in UCB (Adabas response code 247)
	AdaIDUCB = 247
	// AdaFCTNY Function not yet implemented (Adabas response code 250)
	AdaFCTNY = 250
	// AdaIUCAL Invalid utility call (Adabas response code 251)
	AdaIUCAL = 251
	// AdaCALLINV Invalid function call - coding error (Adabas response code 252)
	AdaCALLINV = 252
	// AdaSYLOD System file not loaded or inconsistent (Adabas response code 253)
	AdaSYLOD = 253

	// AdaBPOLL Insufficient space in attached buffer (Adabas response code 255)
	AdaBPOLL = 255
)

// Adabas command code definitions, list of valid Adabas calls

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
	ri
)

var commandCodes = []string{"  ", "OP", "CL", "BT", "ET", "LF", "L1", "L2", "L3", "L4", "L5", "L6", "L9",
	"N1", "N2", "A1", "S1", "S2", "S3", "E1", "U1", "U2", "U3", "RC", "RI"}

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
