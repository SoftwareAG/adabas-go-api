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

// Package adatypes is used to define types use to parse field definitions tables.
// It converts Adabas types to GO types and vice versa
package adatypes

import (
	"bytes"
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
	node  *SearchNode
	value *SearchValue
}

// String provide string of search tree
func (tree *SearchTree) String() string {
	if tree.node != nil {
		return fmt.Sprintf("Tree by node: \n%s", tree.node.String())

	}
	return fmt.Sprintf("Tree by value: %s", tree.value.String())
}

// SearchBuffer returns search buffer of the search tree
func (tree *SearchTree) SearchBuffer() string {
	var buffer bytes.Buffer
	if tree.node != nil {
		tree.node.searchBuffer(&buffer)
	} else {
		tree.value.searchBuffer(&buffer)
	}
	buffer.WriteRune('.')
	Central.Log.Debugf("Search buffer %s", buffer.String())
	return buffer.String()
}

// ValueBuffer returns value buffer of the search tree
func (tree *SearchTree) ValueBuffer(buffer *bytes.Buffer) {
	if tree.node != nil {
		tree.node.valueBuffer(buffer)
		return
	}
	var intBuffer []byte
	helper := NewHelper(intBuffer, math.MaxInt8, endian())
	Central.Log.Debugf("Tree value value buffer %s", tree.value.value.String())
	tree.value.value.StoreBuffer(helper)
	buffer.Write(helper.buffer)
	//buffer.Write(tree.value.value.Bytes())
}

func (tree *SearchTree) addNode(node *SearchNode) {
	tree.node = node
}

func (tree *SearchTree) addValue(value *SearchValue) {
	tree.value = value
}

// OrderBy provide list of descriptor names for this search
func (tree *SearchTree) OrderBy() []string {
	var uniqueDescriptors []string
	if tree.node != nil {
		Central.Log.Debugf("Search node descriptor")
		descriptors := tree.node.orderBy()
		Central.Log.Debugf("Descriptor list: %v", descriptors)
		for _, d := range descriptors {
			add := true
			for _, ud := range uniqueDescriptors {
				Central.Log.Debugf("Check descriptor %s to unique descriptor %s", d, ud)
				if d == ud {
					add = false
					break
				}
			}
			if add {
				Central.Log.Debugf("Add node descriptor : %s", d)
				uniqueDescriptors = append(uniqueDescriptors, d)
			}
		}
	} else {
		Central.Log.Debugf("Empty node use value descriptor")
		descriptor := tree.value.orderBy()
		if descriptor != "" {
			Central.Log.Debugf("Add value descriptor : %s", descriptor)
			uniqueDescriptors = append(uniqueDescriptors, descriptor)
		}
	}
	Central.Log.Debugf("Unique descriptor list: %v", uniqueDescriptors)
	return uniqueDescriptors
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
	nodes  []*SearchNode
	values []*SearchValue
	logic  logicBound
}

func (node *SearchNode) addNode(childNode *SearchNode) {
	node.nodes = append(node.nodes, childNode)
}

func (node *SearchNode) addValue(value *SearchValue) {
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
	Central.Log.Debugf("Before node %s in %s", buffer.String(), node.logic.String())

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
	Central.Log.Debugf("After node %s in %s", buffer.String(), node.logic.String())
}

func (node *SearchNode) valueBuffer(buffer *bytes.Buffer) {
	Central.Log.Debugf("Tree Node value buffer")
	Central.Log.Debugf("Values %d", len(node.values))
	if len(node.nodes) > 0 && (node.logic == AND || node.logic == OR) {
		node.nodes[0].valueBuffer(buffer)
	}
	for i, v := range node.values {
		var intBuffer []byte
		helper := NewHelper(intBuffer, math.MaxInt8, endian())
		Central.Log.Debugf("Tree value value buffer %s", v.value.String())
		v.value.StoreBuffer(helper)
		buffer.Write(helper.buffer)
		Central.Log.Debugf("%d Len buffer %d", i, buffer.Len())
	}
	for i, n := range node.nodes {
		if i > 0 || !(node.logic == AND || node.logic == OR) {
			n.valueBuffer(buffer)
		}
	}
}

func (node *SearchNode) orderBy() []string {
	var descriptors []string
	for _, n := range node.nodes {
		subDescriptors := n.orderBy()
		descriptors = append(descriptors, subDescriptors...)
	}
	for _, v := range node.values {
		subDescriptor := v.orderBy()
		if subDescriptor != "" {
			descriptors = append(descriptors, subDescriptor)
		}
	}
	return descriptors
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

// SearchValue value endpoint
type SearchValue struct {
	field   string
	adaType IAdaType
	value   IAdaValue
	comp    comparator
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

func (value *SearchValue) orderBy() string {
	Central.Log.Debugf("Check descriptor %s", value.adaType.Name())
	if value.value.Type().IsOption(FieldOptionDE) {
		Central.Log.Debugf("Found descriptor %s", value.adaType.Name())
		return value.value.Type().ShortName()
	}
	return ""
}

func (value *SearchValue) searchFields() string {
	return value.value.Type().Name()
}

func (value *SearchValue) searchBuffer(buffer *bytes.Buffer) {
	Central.Log.Debugf("Before value %s", buffer.String())
	value.value.FormatBuffer(buffer, &BufferOption{})
	if value.comp != NONE {
		buffer.WriteByte(',')
		buffer.WriteString(value.comp.String())
	}
	Central.Log.Debugf("After value %s", buffer.String())
}

func checkComparator(comp string) comparator {
	switch comp {
	case "=":
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
	Central.Log.Debugf("start: %s", searchWithConstants)
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
		Central.Log.Debugf("for: %s", searchWithConstants)
		for {
			Central.Log.Debugf("loop: %s", searchWithConstants)
			endConstants = partStartConstants + strings.IndexByte(searchWithConstants, '\'') + 1
			searchWithConstants = searchWithConstants[endConstants-startConstants:]
			partStartConstants = endConstants
			Central.Log.Debugf("start: %d end:%d rest=%s", startConstants, endConstants, searchWithConstants)
			Central.Log.Debugf("Check %d", endConstants-startConstants-1)
			if searchString[endConstants-1] != '\\' {
				break
			}
		}
		Central.Log.Debugf("after for: %s", searchWithConstants)
		Central.Log.Debugf("[%d,%d]",
			startConstants, endConstants)
		searchWithConstants = searchString
		Central.Log.Debugf("Constant %s [%d,%d]", searchWithConstants[startConstants+1:endConstants],
			startConstants, endConstants)
		constant := searchString[startConstants+1 : endConstants]
		constant = strings.Replace(constant, "\\\\", "", -1)
		Central.Log.Debugf("Register constant: %s", constant)
		searchInfo.constants = append(searchInfo.constants, constant)
		searchString = searchWithConstants[0:startConstants] + ConstantIndicator + "{" + strconv.Itoa(index) + "}" + searchWithConstants[endConstants+1:]
		Central.Log.Debugf("New search constants: %s", searchString)
		index++
		for {
			startConstants = strings.IndexByte(searchString, '\'')
			Central.Log.Debugf("Current index: %d", startConstants)
			if !(startConstants > 0 && searchString[startConstants-1] == '\\') {
				break
			}
		}
	}
	searchInfo.search = searchString
	Central.Log.Debugf("Result search formel: %s and %#v", searchString, searchInfo)
	//        return new SearchInfo(searchWithConstants,
	//            constants.toArray(new String[0]))
	return &searchInfo
}

// ParseSearch parse search tree
func (searchInfo *SearchInfo) ParseSearch() (tree *SearchTree, err error) {
	Central.Log.Debugf("Parse search info: %#v", searchInfo)
	tree = &SearchTree{}
	err = searchInfo.extractBinding(tree, searchInfo.search)
	if err != nil {
		return nil, err
	}
	return
}

// GenerateTree generate tree search information
func (searchInfo *SearchInfo) GenerateTree() (tree *SearchTree, err error) {
	tree = &SearchTree{}
	err = searchInfo.extractBinding(tree, searchInfo.search)
	if err != nil {
		return nil, err
	}
	return
}

func (searchInfo *SearchInfo) extractBinding(parentNode ISearchNode, bind string) (err error) {
	var node *SearchNode

	Central.Log.Debugf("Extract binding of: %s in parent Node: %s", bind, parentNode.String())

	binds := regexp.MustCompile(" AND | and ").Split(bind, -1)
	if len(binds) > 1 {
		Central.Log.Debugf("Found AND binds: %d", len(binds))
		node = &SearchNode{logic: AND}
		searchInfo.NeedSearch = true
	} else {
		Central.Log.Debugf("Check or bindings")
		binds = regexp.MustCompile(" OR | or ").Split(bind, -1)
		if len(binds) > 1 {
			Central.Log.Debugf("Found OR binds: %d", len(binds))
			node = &SearchNode{logic: OR}
			searchInfo.NeedSearch = true
		}
	}
	if node != nil {
		Central.Log.Debugf("Go through nodes")
		parentNode.addNode(node)
		for _, bind := range binds {
			Central.Log.Debugf("Go through bind: %s", bind)
			err = searchInfo.extractBinding(node, bind)
			if err != nil {
				return
			}
		}
	} else {
		Central.Log.Debugf("Go through value bind: %s", bind)
		err = searchInfo.extractComparator(bind, parentNode)
		if err != nil {
			return
		}
	}
	return
}

func (searchInfo *SearchInfo) extractComparator(search string, node ISearchNode) (err error) {
	Central.Log.Debugf("Search: %s", search)
	parameter := regexp.MustCompile("!=|=|<=|>=|<>|<|>").Split(search, -1)
	field := parameter[0]
	value := parameter[len(parameter)-1]
	lowerLevel := &SearchValue{field: field}
	Central.Log.Debugf("Field: %s Value: %s from %v", lowerLevel.field, value, parameter)

	/* Check for range information */
	if regexp.MustCompile("^[\\[\\(].*:.*[\\]\\)]$").MatchString(value) {
		/* Found range definition, will add lower and upper limit */
		Central.Log.Debugf("Range found")
		rangeNode := &SearchNode{logic: RANGE}

		/*
		 * Check for lower level and upper level comparator
		 * Mainframe don't like comparator in range
		 */
		var minimumRange comparator
		var maximumRange comparator
		//             isMainframe := false;
		//            if (request != null && request.getTarget().isMainframe()) {
		//                isMainframe = true;
		//            }
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
		//   lowerLevel = request.createSearchNode(field.trim(), minimumRange);
		columnIndex := strings.IndexByte(value, ':')
		startValue := value[1:columnIndex]
		Central.Log.Debugf("Search range start value %s %v", startValue, minimumRange)

		err = searchInfo.searchFieldValue(lowerLevel, startValue)
		if err != nil {
			return
		}
		//searchInfo.extractSearchNodeValue(startValue, lowerLevel)
		rangeNode.addValue(lowerLevel)

		/* Generate upper level value */
		upperLevel := &SearchValue{field: strings.TrimSpace(field), comp: maximumRange}
		//            SearchTree upperLevel =
		//                request.createSearchNode(field.trim(), maximumRange);
		endValue := value[columnIndex+1 : len(value)-1]
		Central.Log.Debugf("Search range end value: %s", startValue)

		//searchInfo.extractSearchNodeValue(endValue, upperLevel)
		err = searchInfo.searchFieldValue(upperLevel, endValue)
		if err != nil {
			return
		}
		// lowerLevel.bound(upperLevel, Logic.RANGE)

		/* On mainframe add NOT operator to exclude ranges */
		if searchInfo.platform.IsMainframe() {
			searchInfo.NeedSearch = true
			//                SearchTree notLowerLevel = null;
			var notLowerLevel *SearchValue
			if value[0] == '(' {
				//                if (value.charAt(0) == '(') {
				//                    LOGGER.debug("Mainframe NOT operator minimum Range");
				notLowerLevel = &SearchValue{field: strings.TrimSpace(field), comp: NONE}
				//                    notLowerLevel =
				//                        request.createSearchNode(field.trim(), C.NONE);
				//                    extractSearchNodeValue(searchInfo, startValue,
				//                        notLowerLevel);
				err = searchInfo.searchFieldValue(notLowerLevel, startValue)
				if err != nil {
					return
				}

				//                    lowerLevel.bound(notLowerLevel, Logic.NOT);
				notRangeNode := &SearchNode{logic: NOT}
				notRangeNode.addValue(notLowerLevel)
				rangeNode.addNode(notRangeNode)
			}
			if value[len(value)-1] == ')' {
				notUpperLevel := &SearchValue{field: strings.TrimSpace(field), comp: NONE}
				err = searchInfo.searchFieldValue(notUpperLevel, endValue)
				if err != nil {
					return
				}
				if notLowerLevel == nil {
					notRangeNode := &SearchNode{logic: NOT}
					notRangeNode.addValue(notUpperLevel)
					rangeNode.addNode(notRangeNode)
					//                    LOGGER.debug("Mainframe NOT operator maximum Range");
					//                    if (notLowerLevel == null) {
					//                        SearchTree notLevel =
					//                            request.createSearchNode(field.trim(), C.NONE);
					//                        extractSearchNodeValue(searchInfo, endValue, notLevel);
					//                        lowerLevel.bound(notLevel, Logic.NOT);
				} else {
					//                        SearchTree notLevel =
					//                            request.createSearchNode(field.trim(), C.NE);
					//                        extractSearchNodeValue(searchInfo, endValue, notLevel);
					//                        notLowerLevel.bound(notLevel, Logic.AND);
					notRangeNode := &SearchNode{logic: AND}
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
			// throw new QueryException(QueryExceptionInfo.ACJ00113, search);
			return
		}
		comparer := search[len(field) : len(search)-len(value)]
		Central.Log.Debugf("Comparer extracted: %s", comparer)
		lowerLevel.comp = checkComparator(comparer)
		Central.Log.Debugf("Search value: %#v", lowerLevel)
		Central.Log.Debugf("Value: %s", value)
		err = searchInfo.searchFieldValue(lowerLevel, value)
		if err != nil {
			return
		}
		node.addValue(lowerLevel)
	}
	return
}

func (searchInfo *SearchInfo) searchFieldValue(searchValue *SearchValue, value string) (err error) {
	Central.Log.Debugf("Search for type %s", searchValue.field)
	searchValue.adaType, err = searchInfo.Definition.SearchType(searchValue.field)
	if err != nil {
		Central.Log.Debugf("Search error ", err)
		return
	}
	searchValue.value, err = searchValue.adaType.Value()
	if err != nil {
		return
	}
	expandedValue, subErr := searchInfo.expandConstants(value)
	if subErr != nil {
		err = subErr
		return
	}
	Central.Log.Debugf("Expanded value >%s<", expandedValue)
	searchValue.value.SetStringValue(expandedValue)
	return
}

func (searchInfo *SearchInfo) expandConstants(value string) (expandConstant string, err error) {
	expandedValue := value
	for strings.Contains(expandedValue, ConstantIndicator) {
		posIndicator := strings.IndexByte(expandedValue, ConstantIndicator[0])
		constantString := regexp.MustCompile(".*#{").ReplaceAllString(expandedValue, "")
		Central.Log.Debugf("Constant without indicator id: %s", constantString)
		constantString = regexp.MustCompile("}.*").ReplaceAllString(constantString, "")
		Central.Log.Debugf("Constant id: %s", constantString)
		postIndicator := strings.IndexByte(expandedValue, '}') + 1
		index, error := strconv.Atoi(constantString)
		if error != nil {
			err = error
			return
		}
		expandedValue = value[:posIndicator] + searchInfo.constants[index-1] + value[postIndicator:]
		Central.Log.Debugf("%d->%s", posIndicator, expandedValue)
	}
	expandConstant = expandedValue
	return
}

func (searchInfo *SearchInfo) extractBinarySearchNodeValue(value string, searchTreeNode *SearchValue) int {
	valuesTrimed := strings.TrimSpace(value)
	values := strings.Split(valuesTrimed, " ")
	var binaryValues [][]byte
	for _, part := range values {
		/* Check if parser constant found */
		if strings.Contains(part, ConstantIndicator) {
			var output bytes.Buffer
			restString := part
			Central.Log.Debugf("Work on part : %s", part)
			for {
				binaryInterpretation := false
				if regexp.MustCompile("^-?H#.*").MatchString(restString) {
					Central.Log.Debugf("Binary value found")
					binaryInterpretation = true
				}
				constantString := regexp.MustCompile("[-H]*#\\{").ReplaceAllString(restString, "")
				constantString = regexp.MustCompile("}.*").ReplaceAllString(constantString, "")
				//                         constantString := restString
				//                            .replaceAll("[-H]*#\\{", "").replaceAll("}.*", "");
				restString = regexp.MustCompile("#\\{[0-9]*\\} *").ReplaceAllString(restString, "")
				//                        restString =
				//                            restString.replaceFirst("#\\{[0-9]*\\} *", "");
				Central.Log.Debugf("Constant string : ",
					constantString)
				Central.Log.Debugf("Rest string : ", restString)

				intTrimed := strings.TrimSpace(constantString)
				index, err := strconv.Atoi(intTrimed)
				if err != nil {
					return -1
				}
				index--
				var binaryValue []byte
				if binaryInterpretation {
					binaryValue = []byte(searchInfo.constants[index])
				} else {
					//					binaryValue = searchTreeNode.binaryValue(		searchInfo.constants[index])
				}
				output.Write(binaryValue)
				//				Central.Log.Debugf("Search field = <{}> of index {}",
				//						java.util.Arrays.toString(binaryValue), index)
				//					Central.Log.Debugf(
				//						DatatypeConverter.printHexBinary(binaryValue))
				if !strings.Contains(restString, ConstantIndicator) {
					break
				}
			}
			binaryValues = append(binaryValues, output.Bytes())
		} else {
			Central.Log.Debugf("Set value: ", value)

			//			binaryValues.add(searchTreeNode.binaryValue(part))
		}
	}
	if len(values) > 1 {
		Central.Log.Debugf("Print binary list: ")
		//		searchTreeNode.SetValue(binaryValues)
	} else {
		//		searchTreeNode.SetValue(binaryValues.get(0))
	}
	return 0
}
