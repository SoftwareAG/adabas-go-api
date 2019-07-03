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
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mainframe = NewPlatform(0x00)
var opensystem = NewPlatform(0x20)

func TestSearchSimpleTree(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA='XXXXX' AND BC='2'")
	assert.Equal(t, "XXXXX", searchInfo.constants[0])
	assert.Equal(t, "2", searchInfo.constants[1])
	assert.Equal(t, "AA=#{1} AND BC=#{2}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchExtraTree(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA='XX\\'XXX' AND BC='2'")
	assert.Equal(t, "XX\\'XXX", searchInfo.constants[0])
	assert.Equal(t, "2", searchInfo.constants[1])
	assert.Equal(t, "AA=#{1} AND BC=#{2}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchSecondTree(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 AND BC='12342'")
	searchValue := &SearchValue{}
	adaType := NewType(FieldTypeString, "BC")
	adaType.SetLength(0)
	searchValue.value, _ = adaType.Value()

	assert.Equal(t, "12342", searchInfo.constants[0])
	assert.Equal(t, "AA=1 AND BC=#{1}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)
	xerr := searchInfo.expandConstants(searchValue, "#{1}")
	assert.Equal(t, []byte{0x31, 0x32, 0x33, 0x34, 0x32}, searchValue.value.Bytes())
	assert.NoError(t, xerr)

}

func TestSearchExtractAndBinding(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchExtractOrBinding(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,R,BC,1,B,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchStringValue(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AD='ABCDEF' AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AD,6,A,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, uint8('A'), valueBuffer[0])
	assert.Equal(t, uint8('F'), valueBuffer[5])
	assert.Equal(t, uint8(2), valueBuffer[6])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchVarStringValue(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AD='ABCDEF' AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AD,6,A,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, uint8('A'), valueBuffer[0])
	assert.Equal(t, uint8('F'), valueBuffer[5])
	assert.Equal(t, uint8(2), valueBuffer[6])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchTwoStringValue(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AE='ABCDEF' AND AD='X123'")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AE,0,A,EQ,D,AD,6,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, uint8(7), valueBuffer[0])
	assert.Equal(t, uint8('A'), valueBuffer[1])
	assert.Equal(t, uint8('F'), valueBuffer[6])
	assert.Equal(t, uint8('X'), valueBuffer[7])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchFails(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "ABCDEF")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)

}

func TestSearchRange(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=[12:44]")
	searchInfo.Definition = tDefinition()
	assert.NotNil(t, searchInfo.platform)
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.NotNil(t, tree.platform)
	assert.Equal(t, "AA,8,B,GE,S,AA,8,B,LE.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchRangeMf(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=[12:44]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 16) {
		return
	}
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoLower(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=(12:44]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 24) {
		return
	}
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
		assert.Equal(t, uint8(12), valueBuffer[23])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
		assert.Equal(t, uint8(12), valueBuffer[16])
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoHigher(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=[12:44)")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 24) {
		return
	}
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
		assert.Equal(t, uint8(44), valueBuffer[23])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
		assert.Equal(t, uint8(44), valueBuffer[16])
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeAlpha(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "SA=[10111011:10111101]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "SA,8,A,GE,S,SA,8,A,LE.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	if !assert.Len(t, valueBuffer, 16) {
		return
	}
	assert.Equal(t, "10111011", string(valueBuffer[0:8]))
	assert.Equal(t, "10111101", string(valueBuffer[8:]))
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoHigherAlpha(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(mainframe, "SA=[10111011:10111013)")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{platform: searchInfo.platform}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "SA,8,A,S,SA,8,A,N,SA,8,A.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 24) {
		return
	}
	assert.Equal(t, "10111011", string(valueBuffer[0:8]))
	assert.Equal(t, "10111013", string(valueBuffer[8:16]))
	assert.Equal(t, "10111013", string(valueBuffer[16:]))
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoBorder(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=(12:44)")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B,D,AA,8,B,NE.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	if !assert.Len(t, valueBuffer, 32) {
		return
	}
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
		assert.Equal(t, uint8(12), valueBuffer[23])
		assert.Equal(t, uint8(44), valueBuffer[31])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
		assert.Equal(t, uint8(12), valueBuffer[16])
		assert.Equal(t, uint8(44), valueBuffer[24])
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchValue(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=123")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.Equal(t, "AA,8,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println("Value Buffer: ", valueBuffer)
	assert.Equal(t, 8, len(valueBuffer))
	if len(valueBuffer) == 8 {
		if bigEndian() {
			assert.Equal(t, uint8(123), valueBuffer[7])
		} else {
			assert.Equal(t, uint8(123), valueBuffer[0])
		}
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])

	fields := tree.SearchFields()
	assert.Equal(t, 1, len(fields))
	assert.Equal(t, "AA", fields[0])
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchFields(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	searchInfo := NewSearchInfo(opensystem, "AA=123 AND BC=1")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.Equal(t, "AA,8,B,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, 9, len(valueBuffer))
	if len(valueBuffer) == 9 {
		if bigEndian() {
			assert.Equal(t, uint8(123), valueBuffer[7])
		} else {
			assert.Equal(t, uint8(123), valueBuffer[0])
		}
	}
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])

	fields := tree.SearchFields()
	assert.Equal(t, 2, len(fields))
	assert.Equal(t, "AA", fields[0])
	assert.Equal(t, "BC", fields[1])
	assert.True(t, searchInfo.NeedSearch)

}

func tDefinition() *Definition {
	groupLayout := []IAdaType{
		NewTypeWithLength(FieldTypeString, "AE", 0),
		NewTypeWithLength(FieldTypeString, "AD", 6),
		NewType(FieldTypePacked, "AC"),
	}
	st := NewSuperType("S1", 0)
	st.AddSubEntry("AD", 1, 2)
	st.FdtFormat = 'A'

	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "BC"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "AB", OccNone, groupLayout),
		NewTypeWithLength(FieldTypeString, "SA", 8),
		NewType(FieldTypeUInt8, "AA"),
		NewTypeWithLength(FieldTypeString, "SD", 20),
		st,
	}
	layout[7].AddOption(FieldOptionUQ)
	layout[7].AddOption(FieldOptionDE)
	layout[8].AddOption(FieldOptionDE)

	testDefinition := NewDefinitionWithTypes(layout)
	for _, t := range layout {
		testDefinition.Register(t)
	}
	for _, t := range groupLayout {
		testDefinition.Register(t)
	}
	return testDefinition
}

func TestRegularSearch(t *testing.T) {
	var re = regexp.MustCompile(`(?m)(\w+)=(\w+|'.+'|\[\w+:\w+\])( AND | OR )?`)
	var str = `XX=123
CC=3443
DD=ABC
EE='1232 3232'
AA=ABC AND BB=XDE AND CC='123' OR AA=[123:123]`

	for i, match := range re.FindAllString(str, -1) {
		fmt.Println(match, "found at index", i)
	}

}

func TestSearchComplex(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=(12:44] AND AD='SMITH'")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	Central.Log.Debugf("Search Tree:", tree.String())

	assert.Equal(t, "AA,8,B,GT,S,AA,8,B,LE,D,AD,6,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	if !assert.Len(t, valueBuffer, 22) {
		return
	}
	if bigEndian() {
		assert.Equal(t, uint8(12), valueBuffer[7])
		assert.Equal(t, uint8(44), valueBuffer[15])
	} else {
		assert.Equal(t, uint8(12), valueBuffer[0])
		assert.Equal(t, uint8(44), valueBuffer[8])
	}
	assert.Equal(t, "SMITH ", string(valueBuffer[16:]))
	descriptors := tree.OrderBy()
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchExtractOr2Binding(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2 OR AC=1")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,R,BC,1,B,EQ,R,AC,1,P,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)
}

func TestSearchExtractOr2BindingError(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2 OR CC=1 OR  DD=2")
	searchInfo.Definition = tDefinition()
	_, err = searchInfo.GenerateTree()
	assert.Error(t, err)
	assert.Equal(t, "ADG0000042: No field type CC found in file definition", err.Error())
}

func TestSearchMixedValue(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=0x00'ABCD'0")
	searchValue := &SearchValue{}
	adaType := NewType(FieldTypeString, "BC")
	adaType.SetLength(0)
	searchValue.value, _ = adaType.Value()
	assert.Equal(t, "ABCD", searchInfo.constants[0])
	assert.Equal(t, "AA=0x00#{1}0", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)
	xerr := searchInfo.expandConstants(searchValue, "0x00#{1}0")
	assert.Equal(t, []byte{0, 65, 66, 67, 68, 0}, searchValue.value.Bytes())
	assert.NoError(t, xerr)

}

func TestSuperDescriptor(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "S1=EMPL")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "S1,4,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	assert.Equal(t, "EMPL", buffer.String())
	assert.False(t, searchInfo.NeedSearch)
}

func TestTwoFields(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "S1=EMPL OR S1=ABC")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "S1,4,A,EQ,R,S1,3,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	assert.Equal(t, "EMPLABC", buffer.String())
	assert.False(t, searchInfo.NeedSearch)
}

func TestMainframeEqual(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "SA='ABC'")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "SA,3,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	assert.Equal(t, "ABC", buffer.String())
	assert.False(t, searchInfo.NeedSearch)

	searchInfo = NewSearchInfo(mainframe, "SA='ABC'")
	searchInfo.Definition = tDefinition()
	tree, serr = searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "SA,3,A,S,SA,3,A.", tree.SearchBuffer())
	buffer = bytes.Buffer{}
	tree.ValueBuffer(&buffer)
	assert.Equal(t, "ABCABC", buffer.String())
	assert.False(t, searchInfo.NeedSearch)

}

func TestSingleLessLength(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())

	searchInfo := NewSearchInfo(opensystem, "SA='ABC'")
	searchInfo.Definition = tDefinition()
	tree, serr := searchInfo.GenerateTree()
	if !assert.NoError(t, serr) {
		return
	}
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "SA,3,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	assert.Equal(t, "ABC", buffer.String())
	assert.True(t, searchInfo.NeedSearch)

}

var searchTest = []struct {
	platform     *Platform
	searchQuery  string
	searchBuffer string
	valueBuffer  string
	needSearch   bool
}{
	{opensystem, "SA='1'", "SA,1,A,EQ.", "1", true},
	{opensystem, "SD='1' AND SD='2'", "SD,1,A,EQ,D,AA,1,A,EQ.", "12", false},
	{mainframe, "SD='1' AND SD='2'", "SD,1,A,S,SD,1,A,D,AA,1,A,S,SD,1,A.", "1122", false},
	{opensystem, "SD='1' OR SD='2'", "SD,1,A,EQ,O,AA,1,A,EQ.", "12", false},
	{mainframe, "SD='1' OR SD='2'", "SD,1,A,S,SD,1,A,O,AA,1,A,S,SD,1,A.", "1122", false},
	{opensystem, "SD<='SMITH'", "SD,5,A,LE.", "SMITH", false},
	{opensystem, "SA='ABC'", "SA,3,A,EQ.", "ABC", true},
	{mainframe, "SA='ABC'", "SA,3,A,S,SA,3,A.", "ABC", true},
	{opensystem, "S1=EMPL OR S1=ABC", "S1,3,A,EQ,R,S1,3,A,EQ.", "EMPLABC", false},
	{mainframe, "S1=EMPL OR S1=ABC", "S1,3,A,S,S1,3,A,R,S1,3,A,S,S1,3,A.", "EMPLEMPLABCABC", false},
	{opensystem, "SD=['A':'B']", "SD,1,A,GE,S,SD,1,A,LE.", "AB", false},
	{mainframe, "SD=['A':'B']", "SD,1,A,S,SD,1,A.", "AB", false},
	{opensystem, "SD=('A':'B']", "SD,1,A,GT,S,SD,1,A,LE.", "AB", false},
	{mainframe, "SD=('A':'B']", "SD,1,A,S,SD,1,A,D,SD,1,A,NE.", "AB", false},
	{opensystem, "SD=['A':'B')", "SD,1,A,GE,S,SD,1,A,LT.", "AB", false},
	{mainframe, "SD=['A':'B')", "SD,1,A,S,SD,1,A,D,SD,1,A,NE.", "AB", false},
}

func TestSearchTest(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	tdefinition := tDefinition()
	fmt.Println(tdefinition)
	for _, search := range searchTest {
		searchInfo := NewSearchInfo(search.platform, search.searchQuery)
		searchInfo.Definition = tdefinition
		tree, serr := searchInfo.GenerateTree()
		if !assert.NoError(t, serr) {
			return
		}
		Central.Log.Debugf(tree.String())
		assert.Equal(t, search.searchBuffer, tree.SearchBuffer())
		var buffer bytes.Buffer
		tree.ValueBuffer(&buffer)
		assert.Equal(t, search.valueBuffer, buffer.String())
		assert.Equal(t, search.needSearch, searchInfo.NeedSearch)
	}

}

var complesSearchTest = []struct {
	platform     *Platform
	searchQuery  string
	searchBuffer string
	valueBuffer  []byte
	needSearch   bool
}{
	{opensystem, "S1=['BVEHICLE '0x00:'BVEHICLE '0xFF]", "S1,10,A,GE,S,S1,10,A,LE.", []byte{0x42, 0x56, 0x45, 0x48, 0x49, 0x43, 0x4c, 0x45, 0x20, 0x0, 0x42, 0x56, 0x45, 0x48, 0x49, 0x43, 0x4c, 0x45, 0x20, 0xff}, false},
	{opensystem, "SD=[0x00'VEHICLE '0x01(10):0x00'VEHICLE '0xFF(5)]", "SD,10,A,GE,S,SD,10,A,LE.", []byte{0x01, 0x42, 0x56, 0x45, 0x48, 0x49, 0x43, 0x4c, 0x45, 0x20, 0x0, 0x42, 0x56, 0x45, 0x48, 0x49, 0x43, 0x4c, 0x45, 0x20, 0xff}, false},
}

func TestComplexSearchTest(t *testing.T) {
	err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	tdefinition := tDefinition()
	fmt.Println(tdefinition.String())
	for _, search := range complesSearchTest {
		searchInfo := NewSearchInfo(search.platform, search.searchQuery)
		searchInfo.Definition = tdefinition
		tree, serr := searchInfo.GenerateTree()
		if !assert.NoError(t, serr) {
			return
		}
		Central.Log.Debugf(tree.String())
		assert.Equal(t, search.searchBuffer, tree.SearchBuffer())
		var buffer bytes.Buffer
		tree.ValueBuffer(&buffer)
		assert.Equal(t, search.valueBuffer, buffer.Bytes())
		assert.Equal(t, search.needSearch, searchInfo.NeedSearch)
	}

}
