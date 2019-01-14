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

package adatypes

import (
	"fmt"
)

func ExampleFormatBytes() {
	f, err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	b := [5]byte{23, 44, 12, 33, 45}
	fmt.Println(FormatBytes("XXX : ", b[:], 4, -1))
	s := []byte("ABCDEFGHIC")
	fmt.Println(FormatBytes("ABC : ", s, 0, -1))
	e := [5]byte{0x81, 0x82, 0xc3, 0xc4, 0x86}
	fmt.Println(FormatBytes("EBCDIC : ", e[:], 5, -1))
	// Output:
	// XXX : 172C0C21 2D [.,.!-] [.....]
	//
	// ABC : 41424344454647484943 [ABCDEFGHIC] [..........]
	//
	// EBCDIC : 8182C3C486  [ÃÄ] [abCDf]
	//
}

func ExampleFormatBytes_x() {
	f, err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	s := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYT")
	fmt.Println(FormatBytes("ABC :", s, 4, 8))
	// Output:
	// ABC :
	// 0000 41424344 45464748  [ABCDEFGH] [........]
	// 0008 494A4B4C 4D4E4F50  [IJKLMNOP] [...<(+|&]
	// 0010 51525354 55565758  [QRSTUVWX] [........]
	// 0018 5954               [YT      ] [..      ]
}

func ExampleFormatBytes_omitDoubles() {
	f, err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	s := []byte("ABCDEFGHABCDEFGHABCDEFGHABCDEFGH")
	fmt.Println(FormatBytes("ABC :", s, 4, 8))
	// Output:
	// ABC :
	// 0000 41424344 45464748  [ABCDEFGH] [........]
	// 0008 skipped equal lines
	// 0018 41424344 45464748  [ABCDEFGH] [........]
}

func ExampleFormatByteBuffer_output() {
	f, err := initLogWithFile("formatter.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	s := []byte("ABCDEFGHABCDEFGHABCDEFGHABCDEFGH")
	fmt.Println(FormatByteBuffer("ABC :", s))
	// Output:
	// ABC :: Dump len=32(0x20)
	// 0000 4142 4344 4546 4748 4142 4344 4546 4748  ABCDEFGHABCDEFGH ................
	// 0010 4142 4344 4546 4748 4142 4344 4546 4748  ABCDEFGHABCDEFGH ................
}
