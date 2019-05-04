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
	"os"
	"time"
)

type cacheEntry struct {
	timestamp     time.Time
	fileFieldTree *StructureType
}

var definitionCache map[string]*cacheEntry

func init() {
	ed := os.Getenv("ENABLE_ADAFDT_CACHE")
	if ed == "1" {
		InitDefinitionCache()
	}
}

// InitDefinitionCache init definition cache
func InitDefinitionCache() {
	definitionCache = make(map[string]*cacheEntry)
	go cacheClearer()
}

// FinitDefinitionCache finit definition cache
func FinitDefinitionCache() {
	definitionCache = nil
}

func traverseCacheCopy(adaType IAdaType, parentType IAdaType, level int, x interface{}) error {
	definition := x.(*Definition)

	f := definition.activeFieldTree.fieldMap[parentType.Name()]
	cst := definition.activeFieldTree
	if f != nil {
		cst = f.(*StructureType)
	}

	var newNode IAdaType
	switch adaType.(type) {
	case *StructureType:
		nst := &StructureType{}
		ost := adaType.(*StructureType)
		*nst = *ost
		Central.Log.Debugf("------->>>>>> Range %s=%s%s -> %p", ost.name, ost.shortName, ost.peRange.FormatBuffer(), ost)
		//		nst.peRange = &AdaRange{}
		nst.peRange = ost.peRange
		//nst.muRange = &AdaRange{}
		nst.muRange = ost.muRange
		Central.Log.Debugf("------->>>>>> Range %s=%s%s %p", nst.name, nst.shortName, nst.peRange.FormatBuffer(), nst)
		nst.SubTypes = make([]IAdaType, 0)
		newNode = nst
	case *AdaPhoneticType:
		npt := &AdaPhoneticType{}
		*npt = *(adaType.(*AdaPhoneticType))
		newNode = npt
	case *AdaSuperType:
		npt := &AdaSuperType{}
		*npt = *(adaType.(*AdaSuperType))
		newNode = npt
	default:
		nat := &AdaType{}
		*nat = *(adaType.(*AdaType))
		newNode = nat
	}
	cst.AddField(newNode)
	definition.activeFields[newNode.Name()]=newNode
	switch adaType.(type) {
	case *StructureType:
		nst := newNode.(*StructureType)
		ost := adaType.(*StructureType)
		nst.peRange = ost.peRange
		nst.muRange = ost.muRange
		Central.Log.Debugf("------->>>>>> Range %s=%s%s %p", nst.name, nst.shortName, nst.peRange.FormatBuffer(), nst)
	case *AdaType:
		nst := newNode.(*AdaType)
		ost := adaType.(*AdaType)
		nst.peRange = ost.peRange
		nst.muRange = ost.muRange
		Central.Log.Debugf("------->>>>>> Range %s=%s%s %p", nst.name, nst.shortName, nst.peRange.FormatBuffer(), nst)
	default:
	}

	return nil
}

// CreateDefinitionByCache create definition out of cache if available
func CreateDefinitionByCache(reference string) *Definition {
	if definitionCache == nil {
		return nil
	}
	e, ok := definitionCache[reference]
	if !ok {
		Central.Log.Debugf("Mis cache entry: %s", reference)
		return nil
	}
	Central.Log.Debugf("Get cache entry: %s", reference)
	definition := NewDefinition()
	x := &StructureType{fieldMap: make(map[string]IAdaType)}
	definition.activeFieldTree = x
	definition.activeFields = make(map[string]IAdaType)
	t := TraverserMethods{EnterFunction: traverseCacheCopy}
	e.fileFieldTree.Traverse(t, 0, definition)
	definition.fileFieldTree = definition.activeFieldTree
	Central.Log.Debugf("ORIG %#v\n", e.fileFieldTree)
	Central.Log.Debugf("COPY TO %#v\n", x)
	definition.InitReferences()
	Central.Log.Debugf("Get copied cache entry: %s", reference)
	definition.DumpTypes(true, false, "copied cache")

	return definition
}

// PutCache put cache entry of current definition
func (def *Definition) PutCache(reference string) {
	if definitionCache == nil {
		return
	}
	Central.Log.Debugf("Put cache entry: %s", reference)
	def.DumpTypes(true, false, "put cache")
	definitionCache[reference] = &cacheEntry{timestamp: time.Now(), fileFieldTree: def.fileFieldTree}
	Central.Log.Debugf("Done put cache entry: %s", reference)
}

func cacheClearer() {
	last := time.Now()
	for {
		time.Sleep(60 * time.Second)
		t := time.Now()
		if definitionCache == nil {
			return
		}
		for r, e := range definitionCache {
			if e.timestamp.Before(last) {
				delete(definitionCache, r)
				Central.Log.Debugf("Remove cache entry: %s", r)
			}
		}
		last = t
	}
}
