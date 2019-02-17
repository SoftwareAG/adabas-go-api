/*
* Copyright © 2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"strings"
)

// Traverser api to handle tree traverses for type definitions
type Traverser func(adaType IAdaType, parentType IAdaType, level int, x interface{}) error

// TraverserMethods structure for Traverser types
type TraverserMethods struct {
	EnterFunction Traverser
	leaveFunction Traverser
}

// NewTraverserMethods new traverser methods structure
func NewTraverserMethods(enter Traverser) TraverserMethods {
	return TraverserMethods{EnterFunction: enter}
}

// TraverseResult Traverser result operation
type TraverseResult int

const (
	// Continue continue traversing the tree
	Continue TraverseResult = iota
	// EndTraverser end the traverser
	EndTraverser
	// SkipStructure skip all other elements of an structure
	SkipStructure
)

// PrepareTraverser prepare giving current main object
type PrepareTraverser func(t interface{}, x interface{}) (TraverseResult, error)

// ElementTraverser prepare start of an element
type ElementTraverser func(value IAdaValue, nr, max int, x interface{}) (TraverseResult, error)

// TraverserValues api to handle tree traverses for values
type TraverserValues func(value IAdaValue, x interface{}) (TraverseResult, error)

// TraverseTypes traverse through the tree of definition calling a callback method
func (def *Definition) TraverseTypes(t TraverserMethods, activeTree bool, x interface{}) error {
	level := 0
	if activeTree {
		return def.activeFieldTree.Traverse(t, level+1, x)
	}
	return def.fileFieldTree.Traverse(t, level+1, x)
}

// TraverserValuesMethods structure for Traverser values
type TraverserValuesMethods struct {
	PrepareFunction PrepareTraverser
	EnterFunction   TraverserValues
	LeaveFunction   TraverserValues
	ElementFunction ElementTraverser
}

// TraverseValues traverse through the tree of values calling a callback method
func (def *Definition) TraverseValues(t TraverserValuesMethods, x interface{}) (ret TraverseResult, err error) {
	if def.Values == nil {
		Central.Log.Debugf("Init create values")
		def.CreateValues(false)
		Central.Log.Debugf("Done create values")
	}
	Central.Log.Debugf("Traverse through level 1 values -> %d", len(def.Values))
	for i, value := range def.Values {
		Central.Log.Debugf("Found level %d value name=%s type=%d fieldindex=%d", value.Type().Level(),
			value.Type().Name(), value.Type().Type(), i)
		ret, err = t.EnterFunction(value, x)
		if err != nil || ret == EndTraverser {
			return
		}
		if ret == SkipStructure {
			Central.Log.Debugf("Skip structure")
			continue
		}
		if value.Type().IsStructure() {
			ret, err = value.(*StructureValue).Traverse(t, x)
			if err != nil {
				return
			}
			if ret == SkipStructure {
				continue
			}
		}
		if t.LeaveFunction != nil {
			ret, err = t.LeaveFunction(value, x)
			if err != nil || ret == EndTraverser {
				return
			}
		}
	}

	Central.Log.Debugf("Ready traverse values")
	return
}

func dumpTypeEnterTrav(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	y := strings.Repeat(" ", int(adaType.Level()))
	buffer := x.(*bytes.Buffer)

	buffer.WriteString(y + adaType.String() + "\n")
	return nil
}

// DumpTypes traverse through the tree of definition calling a callback method
func (def *Definition) DumpTypes(doLog bool, activeTree bool) {
	var buffer bytes.Buffer
	if activeTree {
		if def.activeFieldTree == nil {
			Central.Log.Debugf("Type tree empty")
			return
		}
		buffer.WriteString("Dump all active field types:\n")
	} else {
		if def.fileFieldTree == nil {
			Central.Log.Debugf("Type tree empty")
			return
		}
		buffer.WriteString("Dump all file field types:\n")
	}
	if !doLog || Central.IsDebugLevel() {
		t := TraverserMethods{EnterFunction: dumpTypeEnterTrav}
		def.TraverseTypes(t, activeTree, &buffer)
		if doLog {
			LogMultiLineString(buffer.String())
			// Central.Log.Debugf("Dump all types: ", buffer.String())
		} else {
			fmt.Println(buffer.String())
		}
	}
}

func dumpValuesEnterTrav(adaValue IAdaValue, x interface{}) (TraverseResult, error) {
	buffer := x.(*bytes.Buffer)
	y := strings.Repeat(" ", int(adaValue.Type().Level()))

	var name string
	switch {
	case adaValue.PeriodIndex() > 0 && adaValue.MultipleIndex() > 0:
		name = fmt.Sprintf("%s[%d,%d]", adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex())
	case adaValue.PeriodIndex() > 0:
		name = fmt.Sprintf("%s[%d]", adaValue.Type().Name(), adaValue.PeriodIndex())
	case adaValue.MultipleIndex() > 0:
		name = fmt.Sprintf("%s[%d]", adaValue.Type().Name(), adaValue.MultipleIndex())
	default:
		name = adaValue.Type().Name()
	}
	if adaValue.Type().IsStructure() {
		structureValue := adaValue.(*StructureValue)
		buffer.WriteString(fmt.Sprintf("%s%s = [%d]\n", y, name, structureValue.NrElements()))
	} else {
		buffer.WriteString(fmt.Sprintf("%s%s = >%s<\n", y, name, adaValue.String()))
	}
	return Continue, nil
}

// DumpValues traverse through the tree of values calling a callback method
func (def *Definition) DumpValues(doLog bool) {
	var buffer bytes.Buffer
	Central.Log.Debugf("Dump all values")
	t := TraverserValuesMethods{EnterFunction: dumpValuesEnterTrav}
	def.TraverseValues(t, &buffer)
	if doLog {
		Central.Log.Debugf("Dump values : %s", buffer.String())
	} else {
		fmt.Println("Dump values : ", buffer.String())
	}
}