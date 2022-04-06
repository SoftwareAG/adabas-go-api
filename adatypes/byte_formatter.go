/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
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

package adatypes

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/text/encoding/charmap"
)

const maximumFormatLength = 4096

// FormatByteBuffer formats the byte array to an output with a hexadecimal part, a ASCII part and
// a EBCDIC converted part of the same data
func FormatByteBuffer(header string, b []byte) string {
	nRows := 16
	var buffer bytes.Buffer
	buffer.WriteString(header)
	buffer.WriteString(":")
	buffer.WriteString(fmt.Sprintf(" Dump len=%d(0x%x)", len(b), len(b)))

	var x [16][2]byte
	r := 0
	newLine := 0
	nrLine := 0
	rlen := len(b)

	byteIndex := 0
	for byteIndex < rlen {
		// for (int i = 0; i < rlen; nrLine++) {
		fi := byteIndex
		ti := byteIndex + nRows
		// if (ti >= rlen) {
		// ti = rlen - 1;
		// }
		isEqual := false
		if nrLine > 0 {
			isEqual = true
			for j := 0; j < nRows; j++ {
				if byteIndex < rlen {
					x[j][r] = b[byteIndex]
					if x[j][r] != x[j][(r+1)%2] {
						isEqual = false
					}
				} else {
					x[j][r] = 0
				}
				byteIndex++
			}
		}
		if isEqual {
			newLine++
		} else {
			newLine = 0
		}

		if ti >= rlen {
			newLine = 0
		}

		if (nrLine > 0) && (newLine > 0) {
			if newLine == 1 {
				buffer.WriteString(fmt.Sprintf("\n%04x ... ", fi))
			}
		} else {
			byteIndex = fi
			for j := 0; j < nRows; j++ {
				if byteIndex%nRows == 0 {
					buffer.WriteString(fmt.Sprintf("\n%04x %02x",
						byteIndex, b[byteIndex]))
				} else {
					if byteIndex < rlen {
						buffer.WriteString(fmt.Sprintf("%02x", b[byteIndex]))
					} else {
						buffer.WriteString("  ")
					}
				}
				if (byteIndex-1)%2 == 0 {
					buffer.WriteString(" ")
				}
				byteIndex++
			}
			byteIndex = fi
			buffer.WriteString(" ")
			for j := 0; j < nRows; j++ {
				if byteIndex < rlen {
					if b[byteIndex] > 31 {
						buffer.WriteString(fmt.Sprintf("%c", b[byteIndex]))
					} else {
						buffer.WriteString(".")
					}
				} else {
					buffer.WriteString(" ")
				}
				byteIndex++
			}
			byteIndex = fi
			buffer.WriteString(" ")
			for j := 0; j < nRows; j++ {
				if byteIndex < rlen {
					a := convertToASCII(b[byteIndex])
					if a > 31 {
						buffer.WriteString(fmt.Sprintf("%c", a))
					} else {
						buffer.WriteString(".")
					}
				}
				byteIndex++
			}
			r = (r + 1) % 2
		}
		byteIndex = ti
		// if (!noLimit) && (nrLine > MAX_LINES) {
		// 	buffer.WriteString("\n------- rest skipped  -----")
		// 	break
		// }
		nrLine++
	}
	buffer.WriteString("\n")
	return buffer.String()

}

func convertToASCII(b byte) byte {
	dec := charmap.CodePage037.NewDecoder()
	bebcdic := []byte{b}
	bascii := make([]byte, 1)
	_, _, err := dec.Transform(bascii, bebcdic, false)
	if err != nil {
		return 0
	}
	return bascii[0]
}

// FormatBytes formats a given byte array and modulo space operator. The modulo space defines the
// the possition a space is added to the output. The maximum give the maximum characters per line.
// This function enhance the display with showing the length if showLength is set to true
func FormatBytes(header string, b []byte, bufferLength int, modSpace int, max int, showLength bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(header)

	formatLength := bufferLength
	if os.Getenv("ADABAS_DUMP_BIG") == "" && formatLength > maximumFormatLength {
		formatLength = maximumFormatLength
	}

	if showLength {
		buffer.WriteString(fmt.Sprintf(" length=%d", bufferLength))
	}

	if max != -1 {
		buffer.WriteString("\n")
	}
	lineCr := max
	var lastLine []byte
	if max < 1 {
		lineCr = formatLength
		lastLine = make([]byte, 0)
	} else {
		lastLine = make([]byte, max)
	}
	noticed := true
	for offset := 0; offset < formatLength; offset += lineCr {
		if offset+lineCr < formatLength {
			if max > 0 && offset > 0 && offset+lineCr < formatLength && bytes.Equal(lastLine, b[offset:offset+lineCr]) {
				if noticed {
					buffer.WriteString(fmt.Sprintf("%04X skipped equal lines\n", offset))
				}
				noticed = false
				continue
			}
			noticed = true
			if max > 0 {
				copy(lastLine, b[offset:offset+lineCr])
			}
		}

		for j := 0; j < lineCr; j++ {
			if max > -1 && j == 0 {
				buffer.WriteString(fmt.Sprintf("%04x ", offset+j))
			}
			if formatLength > (offset + j) {
				buffer.WriteString(fmt.Sprintf("%02x", b[offset+j]))
			} else {
				buffer.WriteString("  ")
			}
			if modSpace > 0 && (j+1)%modSpace == 0 {
				buffer.WriteString(" ")
			}
		}
		buffer.WriteString(" [")
		for j := 0; j < lineCr; j++ {
			if formatLength > (offset + j) {
				if b[offset+j] > 31 {
					buffer.WriteString(fmt.Sprintf("%c", b[offset+j]))
				} else {
					buffer.WriteString(".")
				}
			} else {
				buffer.WriteString(" ")
			}
		}
		buffer.WriteString("] [")
		for j := 0; j < lineCr; j++ {
			if formatLength > (offset + j) {
				a := convertToASCII(b[offset+j])
				if a > 31 {
					buffer.WriteString(fmt.Sprintf("%c", a))
				} else {
					buffer.WriteString(".")
				}
			} else {
				buffer.WriteString(" ")
			}
		}

		buffer.WriteString("]\n")
	}
	return buffer.String()
}
