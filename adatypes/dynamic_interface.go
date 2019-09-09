package adatypes

import (
	"bytes"
	"reflect"
	"strings"
)

// DynamicInterface dynamic interface
type DynamicInterface struct {
	DataType   reflect.Type
	FieldNames map[string][]string
}

func generateFieldNames(ri reflect.Type, f map[string][]string, fields []string) {
	Central.Log.Debugf("Generate field names for %s", ri.Name())
	for fi := 0; fi < ri.NumField(); fi++ {
		ct := ri.Field(fi)
		fieldName := ct.Name
		adabasFieldName := fieldName
		tag := ct.Tag.Get("adabas")
		Central.Log.Debugf("fieldName=%s/%s -> tag=%s", adabasFieldName, fieldName, tag)
		if tag != "" {
			s := strings.Split(tag, ":")
			adabasFieldName = s[0]
		}
		subFields := make([]string, len(fields))
		copy(subFields, fields)
		subFields = append(subFields, fieldName)
		Central.Log.Debugf("Set field names to %s -> %v", adabasFieldName, subFields)
		f[adabasFieldName] = subFields
		Central.Log.Debugf("Type struct field = %v", ct.Type.Kind())
		if ct.Type.Kind() == reflect.Ptr {
			Central.Log.Debugf("Pointer found %v %v", ct.Type.Name(), ct.Type.Elem().Name())
			//et := reflect.TypeOf(ct.Type.Elem())
			generateFieldNames(ct.Type.Elem(), f, subFields)
		}
	}

}

// CreateDynamicInterface constructor create dynamic interface
func CreateDynamicInterface(i interface{}) *DynamicInterface {
	ri := reflect.TypeOf(i)
	if ri.Kind() == reflect.Ptr {
		ri = ri.Elem()
	}
	dynamic := &DynamicInterface{DataType: ri, FieldNames: make(map[string][]string)}
	Central.Log.Debugf("Dynamic interface %s", ri.Name())
	Central.Log.Debugf("Dynamic interface %v nrFields=%d", ri, ri.NumField())
	generateFieldNames(ri, dynamic.FieldNames, make([]string, 0))
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
