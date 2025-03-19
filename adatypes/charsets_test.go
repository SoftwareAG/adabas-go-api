/*
* Copyright © 2019-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type testCharset struct {
	charsetName string
	testString  string
	validate    []byte
}

var charsetList = []*testCharset{
	{"ISO-8859-1", "abc$üäö()!+#", []byte{97, 98, 99, 36, 252, 228, 246, 40, 41, 33, 43, 35}},
	{"ISO-8859-15", "abc$€üäö()!+#", []byte{97, 98, 99, 36, 164, 252, 228, 246, 40, 41, 33, 43, 35}},
	{"windows-1251", "Покупатели", []byte{207, 238, 234, 243, 239, 224, 242, 229, 235, 232}},
	// {"x-iscii91", "जाधव", []byte{186, 218, 197, 212, 0, 0, 0, 0}},
	{"ibm852", "Đorđe Balašević", []byte{209, 111, 114, 208, 101, 32, 66, 97, 108, 97, 231, 101, 118, 105, 134}},
	{"shift_jis", "明伯", []byte{150, 190, 148, 140}},
	{"US-ASCII", "ABCabcRSTUVXYZxyz$!-", []byte{65, 66, 67, 97, 98, 99, 82, 83, 84, 85, 86, 88, 89, 90, 120, 121, 122, 36, 33, 45}},
	{"CP037", "ABCabcRSTUVXYZxyz$!-", []byte{193, 194, 195, 129, 130, 131, 217, 226, 227, 228, 229, 231, 232, 233, 167, 168, 169, 91, 90, 96}},
	{"IBM037", "ABCabcRSTUVXYZxyz$!-", []byte{193, 194, 195, 129, 130, 131, 217, 226, 227, 228, 229, 231, 232, 233, 167, 168, 169, 91, 90, 96}},
	{"ISO-8859-1", "ABCabcRSTUVXYZxyz$!-", []byte{65, 66, 67, 97, 98, 99, 82, 83, 84, 85, 86, 88, 89, 90, 120, 121, 122, 36, 33, 45}},
	{"ISO-8859-15", "ABCabcRSTUVXYZxyz$!-", []byte{65, 66, 67, 97, 98, 99, 82, 83, 84, 85, 86, 88, 89, 90, 120, 121, 122, 36, 33, 45}},
	{"US-ASCII", "ABCabcRSTUVXYZxyz$!-", []byte{65, 66, 67, 97, 98, 99, 82, 83, 84, 85, 86, 88, 89, 90, 120, 121, 122, 36, 33, 45}},
	{"windows-1252", "ABCabcRSTUVXYZxyz$!-", []byte{65, 66, 67, 97, 98, 99, 82, 83, 84, 85, 86, 88, 89, 90, 120, 121, 122, 36, 33, 45}},
	{"ISO-8859-1", "Gérard Depardieu", []byte{71, 233, 114, 97, 114, 100, 32, 68, 101, 112, 97, 114, 100, 105, 101, 117}},
}

func TestLookupCharset(t *testing.T) {
	initTestLogWithFile(t, "charset.log")

	Central.Log.Infof("TEST: %s", t.Name())
	ck := struct {
		name     string
		encoding encoding.Encoding
	}{name: "ISO-8859-1", encoding: charmap.Windows1252}
	ne := lookupCharset(ck.name)
	if !assert.NotNil(t, ne) {
		return
	}
	assert.Equal(t, ck.encoding, ne)

}

func TestMapCharset(t *testing.T) {
	initTestLogWithFile(t, "charset.log")

	Central.Log.Infof("TEST: %s", t.Name())

	e, n := charset.Lookup("ISO8859-1")
	d, nus := charset.Lookup("US-ASCII")
	fmt.Println("Lookup:", n, nus)
	germanTests := []string{"���", "���", "���", "���"}
	g2 := []int8{-28, -10, -4}
	gb := make([]byte, 3)
	for i, b := range g2 {
		gb[i] = byte(b)
	}
	{
		dst := make([]byte, 20)
		enc := d.NewDecoder()
		nd, ns, err := enc.Transform(dst, gb, false)
		fmt.Println(nus, "->", g2, gb)
		fmt.Println("G1 error ->", err)
		fmt.Println("G1->", nd, ns, dst, string(dst))

	}
	for _, g := range g2 {
		fmt.Printf("G2 %d\n", g)
	}
	{
		gb := make([]byte, 3)
		for i, b := range g2 {
			gb[i] = byte(b)
		}
		dst := make([]byte, 10)
		enc := e.NewDecoder()
		nd, ns, err := enc.Transform(dst, gb, false)
		fmt.Println("G2->", g2)
		fmt.Println("G2->", gb)
		fmt.Println("G2 error ->", err)
		fmt.Println("G2->", nd, ns, dst, string(dst))

	}
	for _, g := range []byte(germanTests[0]) {
		fmt.Printf("gx2 %d\n", g)
	}

	for _, g := range germanTests {
		fmt.Println("Origin:", []byte(g), g)
		fmt.Println("Lookup", n)
		enc := e.NewEncoder()
		dst := make([]byte, 10)
		nd, ns, err := enc.Transform(dst, []byte(g), false)
		fmt.Println(nd, ns, err, dst, string(dst))
	}

	// m := NewAdabasMap("testmap")
	// m.setDefaultOptions("charset=ISO8859-1")
	// assert.Equal(t, "testmap", m.Name)
	// assert.Equal(t, "ISO8859-1", m.DefaultCharset)
	// m.setDefaultOptions("charset=US-ASCII")
	// assert.Equal(t, "testmap", m.Name)
	// assert.Equal(t, "US-ASCII", m.DefaultCharset)
	// e, cname := charset.Lookup(m.DefaultCharset)
	assert.NotNil(t, e)
	fmt.Println("Check convert")
	s := convrtToUTF8([]byte(gb), "ISO-8859-1")
	assert.Equal(t, "äöü", s)
	// assert.Equal(t, "windows-1252", cname)
	e, cname := charset.Lookup("ISO-8859-1")
	assert.NotNil(t, e)
	assert.Equal(t, "windows-1252", cname)
	cm := charmap.ISO8859_1
	assert.Equal(t, "ISO 8859-1", cm.String())
}

func TestMapCharsetCheck(t *testing.T) {
	for _, c := range charsetList {
		e, n := charset.Lookup(c.charsetName)
		if e == nil {
			fmt.Println(c.charsetName, "not found")
			continue
		}
		ne := lookupCharset(c.charsetName)
		if !assert.NotNil(t, ne) {
			return
		}
		dst := make([]byte, len(c.validate))
		tb := []byte(c.testString)
		fmt.Println(c.charsetName, n, tb)
		nd, x, err := e.NewEncoder().Transform(dst, tb, false)
		if err == nil {
			if !assert.Equal(t, c.validate, dst) {
				fmt.Println("Error transforming string", nd, x, c.testString)
			}
		} else {
			fmt.Println("Error", err)
		}

		dst = make([]byte, len([]byte(c.testString)))
		nd, x, err = e.NewDecoder().Transform(dst, c.validate, false)
		if err == nil {
			if !assert.Equal(t, c.testString, string(dst)) {
				fmt.Println("Error transforming ...", nd, x, string(dst))
			}
		} else {
			fmt.Println("Error", err)
		}
	}
}

func TestUnicodeConverter(t *testing.T) {
	initTestLogWithFile(t, "charset.log")

	Central.Log.Infof("TEST: %s", t.Name())

	for _, c := range charsetList {
		x := NewUnicodeConverter(c.charsetName)
		if !assert.NotNil(t, x, "Error lookup "+c.charsetName) {
			continue
		}
		xsBytes, _ := x.Decode(c.validate)
		if !assert.Equal(t, c.testString, string(xsBytes)) {
			fmt.Println("Error coding", c.charsetName, c.testString)
			return
		}
		orgBytes, _ := x.Encode(xsBytes)
		if !assert.Equal(t, c.validate, orgBytes) {
			fmt.Println("Error coding", c.charsetName, c.testString)
			return
		}

	}

}

func convrtToUTF8(strBytes []byte, origEncoding string) string {
	byteReader := bytes.NewReader(strBytes)
	reader, _ := charset.NewReaderLabel(origEncoding, byteReader)
	strBytes, _ = io.ReadAll(reader)
	return string(strBytes)
}
