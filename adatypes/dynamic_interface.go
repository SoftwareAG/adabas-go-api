package adatypes

import (
	"bytes"
	"reflect"
)

// DynamicInterface dynamic interface
type DynamicInterface struct {
	DataType   reflect.Type
	FieldNames map[string]string
}

// CreateDynamicInterface constructor create dynamic interface
func CreateDynamicInterface(i interface{}) *DynamicInterface {
	ri := reflect.TypeOf(i)
	if ri.Kind() == reflect.Ptr {
		ri = ri.Elem()
	}
	dynamic := &DynamicInterface{DataType: ri, FieldNames: make(map[string]string)}
	Central.Log.Debugf("Dynamic interface %s", ri.Name())
	Central.Log.Debugf("Dynamic interface %v nrFields=%d", ri, ri.NumField())
	for fi := 0; fi < ri.NumField(); fi++ {
		fieldName := ri.Field(fi).Name
		adabasFieldName := fieldName
		tag := ri.Field(fi).Tag.Get("adabas")
		Central.Log.Debugf("fieldName=%s/%s -> tag=%s", adabasFieldName, fieldName, tag)
		if tag != "" {
			adabasFieldName = tag
		}
		dynamic.FieldNames[adabasFieldName] = fieldName
	}
	return dynamic
}

// CreateQueryFields create query field list of dynamic interface given
func (dynamic *DynamicInterface) CreateQueryFields() string {
	var buffer bytes.Buffer
	for fieldName := range dynamic.FieldNames {
		if buffer.Len() > 0 {
			buffer.WriteRune(',')
		}
		buffer.WriteString(fieldName)
	}
	Central.Log.Debugf("Create query fields: %s", buffer.String())

	return buffer.String()
}
