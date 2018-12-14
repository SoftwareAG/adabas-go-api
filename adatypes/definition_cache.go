package adatypes

import (
	"time"
)

type cacheEntry struct {
	timestamp     time.Time
	fileFieldTree *StructureType
}

var definitionCache map[string]*cacheEntry

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
	definition.activeFieldTree = e.fileFieldTree
	definition.fileFieldTree = definition.activeFieldTree
	definition.InitReferences()
	return definition
}

// PutCache put cache entry of current definition
func (def *Definition) PutCache(reference string) {
	if definitionCache == nil {
		return
	}
	definitionCache[reference] = &cacheEntry{timestamp: time.Now(), fileFieldTree: def.fileFieldTree}
	Central.Log.Debugf("Put cache entry: %s", reference)
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
