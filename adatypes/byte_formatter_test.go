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
	"fmt"
)

func ExampleFormatBytes() {
	err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	b := [5]byte{23, 44, 12, 33, 45}
	fmt.Println(FormatBytes("XXX : ", b[:], len(b[:]), 4, -1, false))
	s := []byte("ABCDEFGHIC")
	fmt.Println(FormatBytes("ABC : ", s, len(s), 0, -1, false))
	e := [5]byte{0x81, 0x82, 0xc3, 0xc4, 0x86}
	fmt.Println(FormatBytes("EBCDIC : ", e[:], len(e[:]), 5, -1, false))
	// Output:
	// XXX : 172c0c21 2d [.,.!-] [.....]
	//
	// ABC : 41424344454647484943 [ABCDEFGHIC] [..........]
	//
	// EBCDIC : 8182c3c486  [ÃÄ] [abCDf]
	//
}

func ExampleFormatBytes_x() {
	err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	s := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYT")
	fmt.Println(FormatBytes("ABC :", s, len(s), 4, 8, false))
	// Output:
	// ABC :
	// 0000 41424344 45464748  [ABCDEFGH] [........]
	// 0008 494a4b4c 4d4e4f50  [IJKLMNOP] [...<(+|&]
	// 0010 51525354 55565758  [QRSTUVWX] [........]
	// 0018 5954               [YT      ] [..      ]
}

func ExampleFormatBytes_omitDoubles() {
	err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	s := []byte("ABCDEFGHABCDEFGHABCDEFGHABCDEFGH")
	fmt.Println(FormatBytes("ABC :", s, len(s), 4, 8, false))
	// Output:
	// ABC :
	// 0000 41424344 45464748  [ABCDEFGH] [........]
	// 0008 skipped equal lines
	// 0018 41424344 45464748  [ABCDEFGH] [........]
}

func ExampleFormatByteBuffer_output() {
	err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	s := []byte("ABCDEFGHABCDEFGHABCDEFGHABCDEFGH")
	fmt.Println(FormatByteBuffer("ABC :", s))
	// Output:
	// ABC :: Dump len=32(0x20)
	// 0000 4142 4344 4546 4748 4142 4344 4546 4748  ABCDEFGHABCDEFGH ................
	// 0010 4142 4344 4546 4748 4142 4344 4546 4748  ABCDEFGHABCDEFGH ................
}
