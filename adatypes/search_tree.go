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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math" //"encoding/binary"
	"regexp"
	"strconv"
	"strings"
)

// ConstantIndicator constant indicator is replaced with constants
const ConstantIndicator = "#"

type comparator int

const (
	// EQ Equals value comparisons
	EQ comparator = iota
	// LT Less than comparisons
	LT
	// LE Less equal comparisons
	LE
	// GT Greater than comparisons
	GT
	// GE Greater equal comparisons
	GE
	// NE Not equal comparison
	NE
	// NONE No comparison (Not needed)
	NONE
)

func (comp comparator) String() string {
	switch comp {
	case EQ:
		return "EQ"
	case LT:
		return "LT"
	case LE:
		return "LE"
	case GT:
		return "GT"
	case GE:
		return "GE"
	case NE:
		return "NE"
	case NONE:
		return ""
	}
	return "UNKNOWN"
}

type logicBound int

const (
	// EMPTY Empty (not needed)
	EMPTY logicBound = iota
	// AND AND logic
	AND
	// OR Adabas OR logic
	OR
	// MOR Adabas OR logic with same descriptor
	MOR
	// RANGE  Range for a value
	RANGE
	// NOT NOT logic
	NOT
)

var logicString = []string{"EMPTY", "AND", "OR", "MOR", "RANGE", "NOT"}

func (logic logicBound) String() string {
	return logicString[logic]
}

var logicAdabas = []string{"", ",D", ",R", ",O", ",S", ",N"}

func (logic logicBound) sb() string {
	return logicAdabas[logic]
}

// ISearchNode interface for adding search tree or nodes into tree
type ISearchNode interface {
	addNode(*SearchNode)
	addValue(*SearchValue)
	String() string
	Platform() *Platform
}

// SearchInfo structure containing search parameters
type SearchInfo struct {
	search     string
	constants  []string
	platform   *Platform
	Definition *Definition
	NeedSearch bool
}

// SearchTree tree entry point
type SearchTree struct {
	platform          *Platform
	node              *SearchNode
	value             *SearchValue
	uniqueDescriptors []string
}

// String provide string of search tree
func (tree *SearchTree) String() string {
	if tree.node != nil {
		return fmt.Sprintf("Tree by node: \n%s", tree.node.String())

	}
	return fmt.Sprintf("Tree by value: %s", tree.value.String())
}

// SearchBuffer returns search buffer of the search tree
func (tree *SearchTree) SearchBuffer() []byte {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Create search buffer ...")
	}
	var buffer bytes.Buffer
	if tree.node != nil {
		tree.node.searchBuffer(&buffer)
	} else {
		tree.value.searchBuffer(&buffer)
	}
	buffer.WriteRune('.')
	return buffer.Bytes()
}

// ValueBuffer returns value buffer of the search tree
func (tree *SearchTree) ValueBuffer(buffer *bytes.Buffer) {
	if tree.node != nil {
		tree.node.valueBuffer(buffer)
		return
	}
	var intBuffer []byte
	helper := NewHelper(intBuffer, math.MaxInt8, endian())
	helper.search = true
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Tree value value buffer %s", tree.value.value.String())
	}
	tree.value.value.StoreBuffer(helper, nil)
	if tree.platform.IsMainframe() && tree.value.comp == EQ {
		_ = tree.value.value.StoreBuffer(helper, nil)
	}
	buffer.Write(helper.buffer)
	//buffer.Write(tree.value.value.Bytes())
}

func (tree *SearchTree) addNode(node *SearchNode) {
	tree.node = node
	node.platform = tree.platform
}

func (tree *SearchTree) addValue(value *SearchValue) {
	tree.value = value
	value.platform = tree.platform
}

// OrderBy provide list of descriptor names for this search
func (tree *SearchTree) OrderBy() []string {
	return tree.uniqueDescriptors
}

func (tree *SearchTree) evaluateDescriptors(fields map[string]bool) bool {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Evaluate node descriptors")
	}
	needSearch := false
	for k, v := range fields {
		if v {
			tree.uniqueDescriptors = append(tree.uniqueDescriptors, k)
		} else {
			needSearch = true
		}
	}
	return needSearch || (len(tree.uniqueDescriptors) != 1)
}

// Platform returns current os platform
func (tree *SearchTree) Platform() *Platform {
	return tree.platform
}

// SearchFields provide list of field names for this search
func (tree *SearchTree) SearchFields() []string {
	var uniqueFields []string
	if tree.node != nil {
		fields := tree.node.searchFields()
		for _, d := range fields {
			add := true
			for _, ud := range uniqueFields {
				if d == ud {
					add = false
					break
				}
			}
			if add {
				uniqueFields = append(uniqueFields, d)
			}
		}
	} else {
		descriptor := tree.value.orderBy()
		if descriptor != "" {
			uniqueFields = append(uniqueFields, descriptor)
		}
	}
	return uniqueFields
}

// SearchNode node entry in the searchtree
type SearchNode struct {
	platform *Platform
	nodes    []*SearchNode
	values   []*SearchValue
	logic    logicBound
}

func (node *SearchNode) addNode(childNode *SearchNode) {
	childNode.platform = node.platform
	node.nodes = append(node.nodes, childNode)
}

func (node *SearchNode) addValue(value *SearchValue) {
	value.platform = node.platform
	node.values = append(node.values, value)
}

func (node *SearchNode) String() string {
	var buffer bytes.Buffer
	if node == nil {
		return "ERROR nil node"
	}

	buffer.WriteString("  Nodes: " + node.logic.String() + "\n")
	for i, v := range node.values {
		buffer.WriteString(fmt.Sprintf("    Values: %d:%s", i, node.logic.String()))
		if i > 0 {
			buffer.WriteString(node.logic.String())
		}
		buffer.WriteString(fmt.Sprintf(" -> %d. value = %s\n", i, v.String()))
	}
	for i, n := range node.nodes {
		buffer.WriteString(fmt.Sprintf("    SubNode: %d:%s", i, node.logic.String()))
		if i > 0 {
			buffer.WriteString(node.logic.String())
		}
		buffer.WriteString(fmt.Sprintf(" \n-> %d. node = %s\n", i, n.String()))
	}
	buffer.WriteString(node.logic.String() + " end\n")
	return buffer.String()
}

func (node *SearchNode) searchBuffer(buffer *bytes.Buffer) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Before node %s in %s", buffer.String(), node.logic.String())
	}

	if len(node.nodes) > 0 && (node.logic == AND || node.logic == OR) {
		node.nodes[0].searchBuffer(buffer)
	}
	for _, v := range node.values {
		if buffer.Len() > 0 {
			buffer.WriteString(node.logic.sb())
		}
		v.searchBuffer(buffer)
	}
	for i, n := range node.nodes {
		if i > 0 || !(node.logic == AND || node.logic == OR) {
			// if buffer.Len() > 0 {
			// 	buffer.WriteString(n.logic.sb())
			// }
			n.searchBuffer(buffer)
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("After node %s in %s", buffer.String(), node.logic.String())
	}
}

func (node *SearchNode) valueBuffer(buffer *bytes.Buffer) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Tree Node value buffer")
		Central.Log.Debugf("Values %d", len(node.values))
	}
	if len(node.nodes) > 0 && (node.logic == AND || node.logic == OR) {
		node.nodes[0].valueBuffer(buffer)
	}
	for i, v := range node.values {
		var intBuffer []byte
		helper := NewHelper(intBuffer, math.MaxInt8, endian())
		helper.search = true
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Tree value value buffer %s", v.value.String())
		}
		err := v.value.StoreBuffer(helper, nil)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Error store buffer: %v", err)
		}
		if node.platform.IsMainframe() && v.comp == EQ {
			err = v.value.StoreBuffer(helper, nil)
			if Central.IsDebugLevel() {
				Central.Log.Debugf("Error store buffer (MF): %v", err)
			}
		}
		buffer.Write(helper.buffer)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("%d Len buffer %d", i, buffer.Len())
		}
	}
	for i, n := range node.nodes {
		if i > 0 || !(node.logic == AND || node.logic == OR) {
			n.valueBuffer(buffer)
		}
	}
}

func (node *SearchNode) searchFields() []string {
	var fields []string
	for _, n := range node.nodes {
		subFields := n.searchFields()
		fields = append(fields, subFields...)
	}
	for _, v := range node.values {
		subFields := v.searchFields()
		if subFields != "" {
			fields = append(fields, subFields)
		}
	}
	return fields
}

// Platform returns current os platform
func (node *SearchNode) Platform() *Platform {
	return node.platform
}

// SearchValue value endpoint
type SearchValue struct {
	platform *Platform
	field    string
	adaType  IAdaType
	value    IAdaValue
	comp     comparator
}

// String shows the current value of the search value
func (value *SearchValue) String() string {
	if value == nil {
		return "nil"
	}
	if value.value == nil {
		return fmt.Sprintf("%s %s undefined", value.field, value.comp.String())
	}
	return fmt.Sprintf("%s %s %s(%d)", value.field, value.comp.String(),
		value.value.String(), value.value.Type().Length())
}

// Platform returns current os platform
func (value *SearchValue) Platform() *Platform {
	return value.platform
}

func (value *SearchValue) orderBy() string {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Order by %s", value.adaType.Name())
	}
	if value.value.Type().IsOption(FieldOptionDE) || value.value.Type().IsSpecialDescriptor() {
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Found descriptor %s", value.adaType.Name())
		}
		return value.value.Type().Name()
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Not a descriptor %s %T", value.adaType.Name(), value.value)
	}
	return ""
}

func (value *SearchValue) searchFields() string {
	return value.value.Type().Name()
}

func (value *SearchValue) searchBuffer(buffer *bytes.Buffer) {
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Before value %s", buffer.String())
	}
	curLen := buffer.Len()
	if curLen > 0 {
		curLen++
	}
	value.value.FormatBuffer(buffer, &BufferOption{StoreCall: true})
	if value.comp != NONE {
		if value.platform.IsMainframe() && value.comp == EQ {
			buffer.WriteString(",S," + buffer.String()[curLen:])
		} else {
			buffer.WriteByte(',')
			buffer.WriteString(value.comp.String())
		}
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("After value %s", buffer.String())
	}
}

func checkComparator(comp string) comparator {
	switch comp {
	case "=", "==":
		return EQ
	case "!=":
		return NE
	case "<=":
		return LE
	case ">=":
		return GE
	case "<>":
		return NE
	case "<":
		return LT
	case ">":
		return GT
	}
	return NONE
}

// NewSearchInfo new search info base to create search tree
func NewSearchInfo(platform *Platform, search string) *SearchInfo {
	searchInfo := SearchInfo{platform: platform, NeedSearch: false}
	searchString := search
	searchWithConstants := searchString
	debug := Central.IsDebugLevel()
	if debug {
		Central.Log.Debugf("Search constants: %s", searchWithConstants)
	}
	index := 1
	startConstants := strings.IndexByte(searchWithConstants, '\'')

	/*
	 * Extract constant values out of string. Please leave it at String
	 * because
	 * charset problems on different charsets
	 */
	for startConstants != -1 {
		endConstants := startConstants
		partStartConstants := startConstants
		searchWithConstants = searchString[partStartConstants+1:]
		if debug {
			Central.Log.Debugf("for: %s", searchWithConstants)
		}
		for {
			if debug {
				Central.Log.Debugf("loop: %s", searchWithConstants)
			}
			endConstants = partStartConstants + strings.IndexByte(searchWithConstants, '\'') + 1
			searchWithConstants = searchWithConstants[endConstants-startConstants:]
			partStartConstants = endConstants
			if debug {
				Central.Log.Debugf("start: %d end:%d rest=%s", startConstants, endConstants, searchWithConstants)
				Central.Log.Debugf("Check %d", endConstants-startConstants-1)
			}
			if searchString[endConstants-1] != '\\' {
				break
			}
		}
		if debug {
			Central.Log.Debugf("after for: %s", searchWithConstants)
			Central.Log.Debugf("[%d,%d]",
				startConstants, endConstants)
		}
		searchWithConstants = searchString
		if debug {
			Central.Log.Debugf("Constant %s [%d,%d]", searchWithConstants[startConstants+1:endConstants],
				startConstants, endConstants)
		}
		constant := searchString[startConstants+1 : endConstants]
		constant = strings.Replace(constant, "\\\\", "", -1)
		if debug {
			Central.Log.Debugf("Register constant: %s", constant)
		}
		searchInfo.constants = append(searchInfo.constants, constant)
		searchString = searchWithConstants[0:startConstants] + ConstantIndicator + "{" + strconv.Itoa(index) + "}" + searchWithConstants[endConstants+1:]
		if debug {
			Central.Log.Debugf("New search constants: %s", searchString)
		}
		index++
		for {
			startConstants = strings.IndexByte(searchString, '\'')
			if debug {
				Central.Log.Debugf("Current index: %d", startConstants)
			}
			if !(startConstants > 0 && searchString[startConstants-1] == '\\') {
				break
			}
		}
	}
	searchInfo.search = searchString
	if debug {
		Central.Log.Debugf("Result search formel: %s and %#v", searchString, searchInfo)
	}
	//        return new SearchInfo(searchWithConstants,
	//            constants.toArray(new String[0]))
	return &searchInfo
}

// GenerateTree generate tree search information
func (searchInfo *SearchInfo) GenerateTree() (tree *SearchTree, err error) {
	Central.Log.Debugf("Generate search tree: %#v", searchInfo)
	tree = &SearchTree{platform: searchInfo.platform}
	fields := make(map[string]bool)
	err = searchInfo.extractBinding(tree, searchInfo.search, fields)
	if err != nil {
		return nil, err
	}
	searchNeeded := tree.evaluateDescriptors(fields)
	if !searchInfo.NeedSearch {
		searchInfo.NeedSearch = searchNeeded
	}
	Central.Log.Debugf("Need search call: %v", searchInfo.NeedSearch)
	return
}

func (searchInfo *SearchInfo) extractBinding(parentNode ISearchNode, bind string, fields map[string]bool) (err error) {
	var node *SearchNode

	Central.Log.Debugf("Extract binding of: %s in parent Node: %s", bind, parentNode.String())

	binds := regexp.MustCompile(" AND | and ").Split(bind, -1)
	if len(binds) > 1 {
		Central.Log.Debugf("Found AND binds: %d", len(binds))
		node = &SearchNode{logic: AND, platform: parentNode.Platform()}
		searchInfo.NeedSearch = true
	} else {
		Central.Log.Debugf("Check or bindings")
		binds = regexp.MustCompile(" OR | or ").Split(bind, -1)
		if len(binds) > 1 {
			Central.Log.Debugf("Found OR binds: %d", len(binds))
			node = &SearchNode{logic: OR, platform: parentNode.Platform()}
			searchInfo.NeedSearch = true
		}
	}
	if node != nil {
		Central.Log.Debugf("Go through nodes")
		parentNode.addNode(node)
		subFields := make(map[string]bool)
		for _, bind := range binds {
			Central.Log.Debugf("Go through bind: %s", bind)
			err = searchInfo.extractBinding(node, bind, subFields)
			if err != nil {
				return
			}
		}
		if node.logic == OR && len(subFields) == 1 {
			node.logic = MOR
		}
		for k, v := range subFields {
			fields[k] = v
		}
	} else {
		Central.Log.Debugf("Go through value bind: %s", bind)
		err = searchInfo.extractComparator(bind, parentNode, fields)
		if err != nil {
			return
		}
	}
	return
}

func (searchInfo *SearchInfo) extractComparator(search string, node ISearchNode, fields map[string]bool) (err error) {
	Central.Log.Debugf("Extract comparator %s", search)
	parameter := regexp.MustCompile("!=|=|<=|>=|<>|<|>").Split(search, -1)
	field := parameter[0]
	value := parameter[len(parameter)-1]
	lowerLevel := &SearchValue{field: field, platform: node.Platform()}
	Central.Log.Debugf("Field: %s Value: %s from %v", lowerLevel.field, value, parameter)

	/* Check for range information */
	if regexp.MustCompile(`^[\[\(].*:.*[\]\)]$`).MatchString(value) {
		/* Found range definition, will add lower and upper limit */
		Central.Log.Debugf("Range found")
		rangeNode := &SearchNode{logic: RANGE, platform: node.Platform()}

		/*
		 * Check for lower level and upper level comparator
		 * Mainframe don't like comparator in range
		 */
		var minimumRange comparator
		var maximumRange comparator

		if searchInfo.platform.IsMainframe() {
			minimumRange = NONE
			maximumRange = NONE
		} else {
			if value[0] == '[' {
				minimumRange = GE
			} else {
				minimumRange = GT
			}
			if value[len(value)-1] == ']' {
				maximumRange = LE
			} else {
				maximumRange = LT
			}
		}
		lowerLevel.comp = minimumRange

		/* Generate lower level value */
		columnIndex := strings.IndexByte(value, ':')
		startValue := value[1:columnIndex]
		Central.Log.Debugf("Search range start value %s %v", startValue, minimumRange)

		err = searchInfo.searchFieldValue(lowerLevel, startValue)
		if err != nil {
			return
		}
		rangeNode.addValue(lowerLevel)
		fields[lowerLevel.adaType.Name()] = lowerLevel.adaType.IsSpecialDescriptor() || lowerLevel.adaType.IsOption(FieldOptionDE)

		/* Generate upper level value */
		upperLevel := &SearchValue{field: strings.TrimSpace(field), comp: maximumRange, platform: node.Platform()}
		endValue := value[columnIndex+1 : len(value)-1]
		Central.Log.Debugf("Search range end value: %s", startValue)

		err = searchInfo.searchFieldValue(upperLevel, endValue)
		if err != nil {
			return
		}

		/* On mainframe add NOT operator to exclude ranges */
		if searchInfo.platform.IsMainframe() {
			searchInfo.NeedSearch = true
			var notLowerLevel *SearchValue
			if value[0] == '(' {
				notLowerLevel = &SearchValue{field: strings.TrimSpace(field), comp: NONE, platform: node.Platform()}
				err = searchInfo.searchFieldValue(notLowerLevel, startValue)
				if err != nil {
					return
				}

				notRangeNode := &SearchNode{logic: NOT}
				notRangeNode.addValue(notLowerLevel)
				rangeNode.addNode(notRangeNode)
			}
			if value[len(value)-1] == ')' {
				notUpperLevel := &SearchValue{field: strings.TrimSpace(field), comp: NONE, platform: node.Platform()}
				err = searchInfo.searchFieldValue(notUpperLevel, endValue)
				if err != nil {
					return
				}
				if notLowerLevel == nil {
					notRangeNode := &SearchNode{logic: NOT, platform: node.Platform()}
					notRangeNode.addValue(notUpperLevel)
					rangeNode.addNode(notRangeNode)
				} else {
					notRangeNode := &SearchNode{logic: AND, platform: node.Platform()}
					notUpperLevel.comp = NE
					notRangeNode.addValue(notUpperLevel)
					rangeNode.addNode(notRangeNode)
				}
			}
		}
		rangeNode.addValue(upperLevel)
		node.addNode(rangeNode)
	} else {
		// No range, add common value with corresponding logic operator
		if len(field) > (len(search) - len(value)) {
			Central.Log.Debugf("FL %d sl=%d vl=%d", len(field),
				len(search), len(value))
			err = NewGenericError(170)
			return
		}
		comparer := search[len(field) : len(search)-len(value)]
		Central.Log.Debugf("Comparer extracted: %s", comparer)
		lowerLevel.comp = checkComparator(comparer)
		if Central.IsDebugLevel() {
			Central.Log.Debugf("Search value: %#v", lowerLevel)
			Central.Log.Debugf("Value: %s", value)
		}
		err = searchInfo.searchFieldValue(lowerLevel, value)
		if err != nil {
			return
		}
		if lowerLevel.comp == NE {
			searchInfo.NeedSearch = true
		}
		fields[lowerLevel.adaType.Name()] = lowerLevel.adaType.IsSpecialDescriptor() || lowerLevel.adaType.IsOption(FieldOptionDE)
		node.addValue(lowerLevel)
	}
	return
}

func (searchInfo *SearchInfo) searchFieldValue(searchValue *SearchValue, value string) (err error) {
	Central.Log.Debugf("Search for type %s", searchValue.field)
	adaType, xerr := searchInfo.Definition.SearchType(searchValue.field)
	if xerr != nil {
		Central.Log.Debugf("Search error: %v", err)
		return xerr
	}
	switch t := adaType.(type) {
	case *AdaType:
		var xType AdaType
		xType = *t
		searchValue.adaType = &xType
	case *AdaSuperType:
		var xType AdaSuperType
		xType = *t
		searchValue.adaType = &xType
	default:
		return NewGenericError(0)
	}

	if Central.IsDebugLevel() {
		Central.Log.Debugf("Search value type: %T (length=%d)", searchValue.adaType, searchValue.adaType.Length())
	}
	searchValue.value, err = searchValue.adaType.Value()
	if err != nil {
		return
	}
	subErr := searchInfo.expandConstants(searchValue, value)
	if subErr != nil {
		err = subErr
		return
	}
	return
}

func (searchInfo *SearchInfo) expandConstants(searchValue *SearchValue, value string) (err error) {
	debug := Central.IsDebugLevel()
	if debug {
		Central.Log.Debugf("Expand constants %s", value)
	}
	expandedValue := value
	var buffer bytes.Buffer
	posIndicator := 0
	postIndicator := 0
	if !strings.Contains(expandedValue, ConstantIndicator) {
		searchValue.value.SetStringValue(value)
		return
	}
	numPart := false
	for strings.Contains(expandedValue, ConstantIndicator) {
		if debug {
			Central.Log.Debugf("Work on expanded value %s", expandedValue)
		}
		posIndicator = strings.Index(expandedValue, ConstantIndicator+"{")
		//posIndicator = strings.IndexByte(expandedValue, ConstantIndicator[0])
		constantString := expandedValue[posIndicator+2:]
		if debug {
			Central.Log.Debugf("Constant without indicator id: %s", constantString)
		}
		constantString = regexp.MustCompile("}.*").ReplaceAllString(constantString, "")
		postIndicator = strings.IndexByte(expandedValue, '}') + 1
		if debug {
			Central.Log.Debugf("Constant id: %s pos=%d post=%d", constantString, posIndicator, postIndicator)
		}
		index, error := strconv.Atoi(constantString)
		if error != nil {
			err = error
			return
		}
		if posIndicator > 0 {
			if debug {
				Central.Log.Debugf("Check numeric value %s", expandedValue[:posIndicator])
			}
			appendNumericValue(&buffer, expandedValue[:posIndicator])
			numPart = true
		}
		expandedValue = expandedValue[postIndicator:]
		buffer.WriteString(searchInfo.constants[index-1])
		if debug {
			Central.Log.Debugf("Expand end=%s", expandedValue)
		}
	}
	if debug {
		Central.Log.Debugf("Rest value=%s", value[postIndicator:])
	}
	if expandedValue != "" {
		appendNumericValue(&buffer, expandedValue)
		numPart = true
	}
	if numPart {
		if debug {
			Central.Log.Debugf("Numeric part available ....")
		}
		searchValue.value.Type().SetLength(uint32(buffer.Len()))
		err = searchValue.value.SetValue(buffer.Bytes())
	} else {
		if debug {
			Central.Log.Debugf("No Numeric part available ....%s", string(expandedValue))
		}
		searchValue.value.Type().SetLength(uint32(buffer.Len()))
		searchValue.value.SetStringValue(buffer.String())
	}
	return
}

func appendNumericValue(buffer *bytes.Buffer, v string) {
	Central.Log.Debugf("Append numeric offset=%d v=%s\n", buffer.Len(), v)
	if v != "" {
		// Work on hexadecimal value
		if strings.HasPrefix(v, "0x") {
			multiplier := 1
			bm := strings.Index(v, "(")
			if bm > 0 {
				em := strings.Index(v, ")")
				Central.Log.Debugf("Multiplier %v", v[bm+1:em])
				var err error
				multiplier, err = strconv.Atoi(v[bm+1 : em])
				if err != nil {
					Central.Log.Debugf("Error multiplier %v", err)
					return
				}
			} else {
				bm = len(v)
			}
			Central.Log.Debugf("Range end %d %v", bm, v[2:bm])
			src := []byte(v[2:bm])
			Central.Log.Debugf("Append numeric %s\n", v[2:bm])
			dst := make([]byte, hex.DecodedLen(len(src)))
			n, err := hex.Decode(dst, src)
			if err != nil {
				Central.Log.Fatal(err)
			}

			Central.Log.Debugf("Byte value %v\n", dst[:n])
			for i := 0; i < multiplier; i++ {
				buffer.Write(dst[:n])
			}
		} else {
			va, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				Central.Log.Fatal(err)
			}
			if va > math.MaxUint32 {
				Central.Log.Fatal("value is greate then maximum")
				// TODO add error return
				return
			}
			if va > 0 {
				bs := make([]byte, 4)
				binary.LittleEndian.PutUint32(bs, uint32(va))
				x := len(bs)
				for x > 0 {
					if bs[x-1] > 0 {
						break
					}
					x--
				}
				buffer.Write(bs[:x])
				Central.Log.Debugf("Byte value -> offset=%d\n", buffer.Len())
			} else {
				buffer.WriteByte(0)
			}
		}
	}
}

// func (searchInfo *SearchInfo) extractBinarySearchNodeValue(value string, searchTreeNode *SearchValue) int {
// 	valuesTrimed := strings.TrimSpace(value)
// 	values := strings.Split(valuesTrimed, " ")
// 	var binaryValues [][]byte
// 	for _, part := range values {
// 		/* Check if parser constant found */
// 		if strings.Contains(part, ConstantIndicator) {
// 			var output bytes.Buffer
// 			restString := part
// 			Central.Log.Debugf("Work on part : %s", part)
// 			for {
// 				binaryInterpretation := false
// 				if regexp.MustCompile("^-?H#.*").MatchString(restString) {
// 					Central.Log.Debugf("Binary value found")
// 					binaryInterpretation = true
// 				}
// 				constantString := regexp.MustCompile(`[-H]*#\{`).ReplaceAllString(restString, "")
// 				constantString = regexp.MustCompile("}.*").ReplaceAllString(constantString, "")
// 				restString = regexp.MustCompile(`#\{[0-9]*\} *`).ReplaceAllString(restString, "")
// 				Central.Log.Debugf("Constant string : ",
// 					constantString)
// 				Central.Log.Debugf("Rest string : ", restString)

// 				intTrimed := strings.TrimSpace(constantString)
// 				index, err := strconv.Atoi(intTrimed)
// 				if err != nil {
// 					return -1
// 				}
// 				index--
// 				var binaryValue []byte
// 				if binaryInterpretation {
// 					binaryValue = []byte(searchInfo.constants[index])
// 				// } else {
// 				// 	//					binaryValue = searchTreeNode.binaryValue(		searchInfo.constants[index])
// 				}
// 				output.Write(binaryValue)
// 				if !strings.Contains(restString, ConstantIndicator) {
// 					break
// 				}
// 			}
// 			binaryValues = append(binaryValues, output.Bytes())
// 		// } else {
// 		// 	Central.Log.Debugf("Set value: ", value)

// 		// 	//			binaryValues.add(searchTreeNode.binaryValue(part))
// 		}
// 	}
// 	// if len(values) > 1 {
// 	// 	Central.Log.Debugf("Print binary list: ")
// 	// 	//		searchTreeNode.SetValue(binaryValues)
// 	// } else {
// 	// 	//		searchTreeNode.SetValue(binaryValues.get(0))
// 	// }
// 	return 0
// }
