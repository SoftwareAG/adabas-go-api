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

package adatypes

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var mainframe = NewPlatform(0x00)
var opensystem = NewPlatform(0x20)

func TestSearchSimpleTree(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA='XXXXX' AND BC='2'")
	assert.Equal(t, "XXXXX", searchInfo.constants[0])
	assert.Equal(t, "2", searchInfo.constants[1])
	assert.Equal(t, "AA=#{1} AND BC=#{2}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchExtraTree(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA='XX\\'XXX' AND BC='2'")
	assert.Equal(t, "XX\\'XXX", searchInfo.constants[0])
	assert.Equal(t, "2", searchInfo.constants[1])
	assert.Equal(t, "AA=#{1} AND BC=#{2}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchSecondTree(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 AND BC='12342'")
	assert.Equal(t, "12342", searchInfo.constants[0])
	assert.Equal(t, "AA=1 AND BC=#{1}", searchInfo.search)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchExtractAndBinding(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchExtractOrBinding(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,R,BC,1,B,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchStringValue(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AD='ABCDEF' AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
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
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AD='ABCDEF' AND BC=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
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
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AE='ABCDEF' AND AD='X123'")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
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

func TestSearchRange(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=[12:44]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,GE,S,AA,8,B,LE.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.False(t, searchInfo.NeedSearch)

}

func TestSearchRangeMf(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=[12:44]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 16) {
		return
	}
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoLower(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=(12:44]")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 24) {
		return
	}
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	assert.Equal(t, uint8(12), valueBuffer[16])
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoHigher(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=[12:44)")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 24) {
		return
	}
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	assert.Equal(t, uint8(44), valueBuffer[16])
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchRangeMfNoHigherAlpha(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(mainframe, "SA=[10111011:10111013)")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
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
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(mainframe, "AA=(12:44)")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,S,AA,8,B,N,AA,8,B,N,AA,8,B.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	if !assert.Len(t, valueBuffer, 32) {
		return
	}
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	assert.Equal(t, uint8(12), valueBuffer[16])
	assert.Equal(t, uint8(44), valueBuffer[24])
	descriptors := tree.OrderBy()
	fmt.Println("Descriptors ", descriptors)
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchValue(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=123")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println("Value Buffer: ", valueBuffer)
	assert.Equal(t, 8, len(valueBuffer))
	if len(valueBuffer) == 8 {
		assert.Equal(t, uint8(123), valueBuffer[0])
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
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	searchInfo := NewSearchInfo(opensystem, "AA=123 AND BC=1")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	assert.Equal(t, "AA,8,B,EQ,D,BC,1,B,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	fmt.Println(valueBuffer)
	assert.Equal(t, 9, len(valueBuffer))
	if len(valueBuffer) == 9 {
		assert.Equal(t, uint8(123), valueBuffer[0])
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
	layout := []IAdaType{
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "BC"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypeGroup, "AB", OccNone, groupLayout),
		NewTypeWithLength(FieldTypeString, "SA", 8),
		NewType(FieldTypeUInt8, "AA"),
	}
	layout[6].AddOption(FieldOptionUQ)
	layout[6].AddOption(FieldOptionDE)

	testDefinition := NewDefinitionWithTypes(layout)
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
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=(12:44] AND AD='SMITH'")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	searchInfo.extractBinding(tree, searchInfo.search)
	Central.Log.Debugf("Search Tree:", tree.String())

	assert.Equal(t, "AA,8,B,GT,S,AA,8,B,LE,D,AD,6,A,EQ.", tree.SearchBuffer())
	var buffer bytes.Buffer
	tree.ValueBuffer(&buffer)
	valueBuffer := buffer.Bytes()
	if !assert.Len(t, valueBuffer, 22) {
		return
	}
	assert.Equal(t, uint8(12), valueBuffer[0])
	assert.Equal(t, uint8(44), valueBuffer[8])
	assert.Equal(t, "SMITH ", string(valueBuffer[16:]))
	descriptors := tree.OrderBy()
	assert.Equal(t, 1, len(descriptors))
	assert.Equal(t, "AA", descriptors[0])
	assert.True(t, searchInfo.NeedSearch)

}

func TestSearchExtractOr2Binding(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2 OR AC=1")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	err = searchInfo.extractBinding(tree, searchInfo.search)
	assert.NoError(t, err)
	Central.Log.Debugf(tree.String())
	assert.Equal(t, "AA,8,B,EQ,R,BC,1,B,EQ,R,AC,1,P,EQ.", tree.SearchBuffer())
	assert.True(t, searchInfo.NeedSearch)
}

func TestSearchExtractOr2BindingError(t *testing.T) {
	f, err := initLogWithFile("search_tree.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())

	searchInfo := NewSearchInfo(opensystem, "AA=1 OR BC=2 OR CC=1 OR  DD=2")
	searchInfo.Definition = tDefinition()
	tree := &SearchTree{}
	err = searchInfo.extractBinding(tree, searchInfo.search)
	assert.Error(t, err)
	assert.Equal(t, "ADG0000041: No field CC found in file definition", err.Error())
}
