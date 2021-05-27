/*
* Copyright © 2019-2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"reflect"
)

type stackData struct {
	field string
	val   reflect.Value
}

type valueInterface struct {
	valStack   *Stack
	curVal     reflect.Value
	fieldNames map[string][]string
}

// evaluateMultipleField evaluate multiple field data into dynamic interface fields
func (tp *valueInterface) evaluateMultipleField(adaValue IAdaValue, v reflect.Value) (result TraverseResult, err error) {
	f, refErr := evaluateReflectValue(v, adaValue, tp)
	if refErr != nil {
		return SkipTree, refErr
	}
	t := v.Type()
	name := tp.evaluateReflectName(adaValue.Type().Name())
	st, ok := t.FieldByName(name)
	Central.Log.Debugf("Found MU field t=%v st=%v ok=%v", t, st.Name, ok)
	if !ok {
		return Continue, nil
	}

	if st.Type.Kind() != reflect.Slice {
		return Continue, nil
	}
	stt := st.Type.Elem()
	Central.Log.Debugf("Use type %s %s %v kind=%v", stt.Name(), stt.String(), stt, stt.Kind())
	sv := adaValue.(*StructureValue)
	elemSlice := reflect.MakeSlice(reflect.SliceOf(stt), sv.NrElements(), 10)
	if stt.Kind() == reflect.Ptr {
		stt = stt.Elem()
	}
	for i := 0; i < sv.NrElements(); i++ {
		entry := reflect.New(stt)
		Central.Log.Debugf("New entry %s -> %s ok=%v i=%v", entry.Type().Name(), entry.Type().String(), entry.IsValid(), entry.CanInterface())
		if entry.Kind() == reflect.Ptr {
			entry = entry.Elem()
		}
		err := SetValueData(entry, sv.Elements[i].Values[0])
		if err != nil {
			return EndTraverser, err
		}
		elemSlice.Index(i).Set(entry)
	}
	f.Set(elemSlice)
	//		tp.curVal = elemSlice
	Central.Log.Debugf("Set %s new base struct value %v", adaValue.Type().Name(), tp.curVal.Type())
	return Continue, nil
}

// evaluatePeriodGroup evaluate period group data into dynamic interface fields
func (tp *valueInterface) evaluatePeriodGroup(adaValue IAdaValue, v reflect.Value) (result TraverseResult, err error) {
	f, refErr := evaluateReflectValue(v, adaValue, tp)
	if refErr != nil {
		Central.Log.Debugf("Error evaluating reflect value %s", adaValue.Type().Name())
		return SkipTree, refErr
	}
	Central.Log.Debugf("Use for PE field f=%v %T", f, f)
	t := v.Type()
	name := tp.evaluateReflectName(adaValue.Type().Name())
	st, ok := t.FieldByName(name)
	if !ok {
		Central.Log.Debugf("Not found PE field %s t=%v v=%s", adaValue.Type().Name(), t, v.Type().Name())
		return Continue, nil
	}
	Central.Log.Debugf("Found PE field %s t=%v st=%v ok=%v", adaValue.Type().Name(), t, st.Name, ok)
	Central.Log.Debugf("Add slice %s %v -> %v %s", st.Name, st.Type, st.Type.String(), st.Type.Kind())
	//stt := st.Type.Elem()
	stt := f.Type().Elem()
	Central.Log.Debugf("Use type %s %s %v kind=%v", stt.Name(), stt.String(), stt, stt.Kind())
	sv := adaValue.(*StructureValue)
	cap := 10
	if sv.NrElements() > cap {
		cap = sv.NrElements()
	}
	elemSlice := reflect.MakeSlice(reflect.SliceOf(stt), sv.NrElements(), cap)
	if stt.Kind() == reflect.Ptr {
		stt = stt.Elem()
	}
	Central.Log.Debugf("Created slice %s", elemSlice.Type().String())
	Central.Log.Debugf("of slice entry %s - %s %v slice %v", stt.Name(), stt.String(), stt.Kind(), elemSlice.Type())

	for i := 0; i < sv.NrElements(); i++ {
		entry := reflect.New(stt)
		Central.Log.Debugf("New entry %s -> %s ok=%v i=%v", entry.Type().Name(), entry.Type().String(), entry.IsValid(), entry.CanInterface())
		elemSlice.Index(i).Set(entry)
	}
	f.Set(elemSlice)
	Central.Log.Debugf("Push slice to stack for %s", adaValue.Type().Name())
	tp.valStack.Push(&stackData{val: tp.curVal, field: adaValue.Type().Name()})
	tp.curVal = elemSlice
	Central.Log.Debugf("Set %s new base PE slice value %v", adaValue.Type().Name(), tp.curVal.Type())
	return Continue, nil
}

// evaluateField evaluate field data into dynamic interface fields
func evaluateField(adaValue IAdaValue, v reflect.Value, tp *valueInterface) (result TraverseResult, err error) {
	f, refErr := evaluateReflectValue(v, adaValue, tp)
	if refErr != nil {
		return SkipTree, refErr
	}
	Central.Log.Debugf("No MU or PE, check kind=%v of %s in %T", f.Kind(), adaValue.Type().Name(), tp.curVal.Type().Name())
	switch f.Kind() {
	case reflect.Slice:
		if !f.IsNil() && f.CanInterface() {
			switch f.Interface().(type) {
			case []byte:
				nv := reflect.ValueOf(adaValue.Bytes())
				f.Set(nv)
				return Continue, nil
			default:
				Central.Log.Errorf("Unknown interface type %T", f.Interface())
			}
		}
		Central.Log.Debugf("Found slice on %s %d,%d", adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex())
		st := f.Type().Elem()
		switch st.Kind() {
		case reflect.Int8:
			Central.Log.Debugf("Go for byte array")
		case reflect.Uint8:
			nv := reflect.ValueOf(adaValue.Bytes())
			f.Set(nv)
		default:
			Central.Log.Errorf("Unknown sub type %s", st.Kind())
		}
	case reflect.Ptr:
		if f.Elem().IsValid() {
			f = f.Elem()
		} else {
			ft := reflect.TypeOf(f.Interface())
			x := reflect.New(ft.Elem())
			f.Set(x)
			f = x
			Central.Log.Debugf("Create new instance for %s is ptr, kind = %v", adaValue.Type().Name(), f.Kind())
		}
		Central.Log.Debugf("Push struct to stack for %s", adaValue.Type().Name())
		tp.valStack.Push(&stackData{val: tp.curVal, field: adaValue.Type().Name()})

		tp.curVal = f.Elem()
		Central.Log.Debugf("Set %s new base Ptr value %v", adaValue.Type().Name(), tp.curVal.Type())
		return Continue, nil

	default:
		if f.IsValid() {
			switch f.Kind() {
			case reflect.Int64, reflect.Int32, reflect.Int:
				i, err := adaValue.Int64()
				if err != nil {
					return EndTraverser, err
				}
				f.SetInt(i)
			case reflect.Int8, reflect.Int16:
				i, err := adaValue.Int64()
				if err != nil {
					return EndTraverser, err
				}
				f.SetInt(i)
			case reflect.Uint64, reflect.Uint32:
				i, err := adaValue.UInt64()
				if err != nil {
					return EndTraverser, err
				}
				f.SetUint(i)
			case reflect.Uint8:
				i, err := adaValue.UInt8()
				if err != nil {
					return EndTraverser, err
				}
				f.SetUint(uint64(i))
			case reflect.Uint16:
				i, err := adaValue.UInt16()
				if err != nil {
					return EndTraverser, err
				}
				f.SetUint(uint64(i))
			case reflect.Float32, reflect.Float64:
				fl, err := adaValue.Float()
				if err != nil {
					return EndTraverser, err
				}
				f.SetFloat(fl)
			case reflect.Bool:
				v, err := adaValue.Int32()
				if err != nil {
					return EndTraverser, err
				}
				b := v > 0
				f.SetBool(b)
			case reflect.String:
				f.SetString(adaValue.String())
			default:
				Central.Log.Errorf("Unknown kind: %v for %s", f.Kind(), adaValue.Type().Name())
			}
			Central.Log.Debugf("%s=%v->%v", adaValue.Type().Name(), v, f)
		} else {
			Central.Log.Debugf("%s is invalid, kind = %v", adaValue.Type().Name(), f.Kind())
		}
	}

	return Continue, nil
}

// evaluateReflectName evaluate field name adaption
func (tp *valueInterface) evaluateReflectName(name string) (reflectName string) {
	reflectName = name
	if fn, ok := tp.fieldNames[name]; ok {
		if len(fn) > 0 {
			reflectName = fn[len(fn)-1]
		}
	}
	return reflectName
}

// evaluateReflectValue evaluate reflect value defined by tag in interface or by Adabas field of map
func evaluateReflectValue(v reflect.Value, adaValue IAdaValue, tp *valueInterface) (reflect.Value, error) {
	var f reflect.Value
	Central.Log.Debugf("Check reflect value %s for tp.curVal=%T", adaValue.Type().Name(), tp.curVal)

	name := tp.evaluateReflectName(adaValue.Type().Name())
	if tp.curVal.Kind() == reflect.Slice {
		// return tp.curVal, nil
		sv := tp.curVal.Index(int(adaValue.PeriodIndex() - 1))
		Central.Log.Debugf("Check final slice value %s (kind=%s)", sv.Type().Name(), sv.Type().Kind())
		if sv.Kind() == reflect.Ptr {
			sv = sv.Elem()
		}
		Central.Log.Debugf("Check final slice value %s (kind=%s)", name, sv.Type().Kind())

		f = sv.FieldByName(name)
	} else {
		f = tp.curVal.FieldByName(name)
	}
	if f.IsValid() {
		Central.Log.Debugf("Check final reflect value %s for %s", adaValue.Type().Name(), f.Type().Name())
	} else {
		Central.Log.Debugf("Check final reflect, got invalid value %s", adaValue.Type().Name())
	}

	return f, nil
}

// traverseValueToInterface traverse through all fields and put value into interface dynamically
func traverseValueToInterface(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	Central.Log.Debugf("Adapt value to interface: %s", adaValue.Type().Name())
	if adaValue.Type().HasFlagSet(FlagOptionMUGhost) {
		Central.Log.Debugf("Skip because in MU ghost %s", adaValue.Type().Name())
		return Continue, nil
	}
	tp := x.(*valueInterface)
	Central.Log.Debugf("Work on value [%s] to interface: %s", adaValue.Type().Name(), tp.curVal.Type().Name())
	v := tp.curVal
	if v.Kind() == reflect.Slice {
		v = v.Index(int(adaValue.PeriodIndex() - 1))
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		Central.Log.Debugf("Slice pe=%d kind=%s", adaValue.PeriodIndex(), v.Kind())
	}
	Central.Log.Debugf("Current struct value %v", v.Type())
	switch adaValue.Type().Type() {
	case FieldTypeMultiplefield:
		return tp.evaluateMultipleField(adaValue, v)
	case FieldTypePeriodGroup:
		return tp.evaluatePeriodGroup(adaValue, v)
	default:
	}
	return evaluateField(adaValue, v, tp)
}

// traverseValueToInterfaceLeave traverse value to interface left
func traverseValueToInterfaceLeave(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	Central.Log.Debugf("Leave value to interface: %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		tp := x.(*valueInterface)
		Central.Log.Debugf("Current %s struct value %v", adaValue.Type().Name(), tp.curVal.Type())
		Central.Log.Debugf("Pop from stack for %s type=%s", adaValue.Type().Name(), adaValue.Type().Type().name())

		if adaValue.Type().Type() != FieldTypeMultiplefield && tp.valStack.Size > 0 {
			sdi, err := tp.valStack.Pop()
			if err != nil {
				return EndTraverser, err
			}
			sd := sdi.(*stackData)
			if sd.field != adaValue.Type().Name() {
				tp.valStack.Push(sd)
			} else {
				tp.curVal = sd.val
			}
			Central.Log.Debugf("Reset %s to struct value %v", adaValue.Type().Name(), tp.curVal.Type())
		}
	}
	return Continue, nil
}

// AdaptInterfaceFields adapt field value to interface field
func (def *Definition) AdaptInterfaceFields(v reflect.Value, fm map[string][]string) error {
	Central.Log.Debugf("Adapt interface")
	if fm == nil {
		return NewGenericError(179)
	}
	tp := &valueInterface{curVal: v.Elem(), valStack: NewStack(), fieldNames: fm}
	t := TraverserValuesMethods{EnterFunction: traverseValueToInterface, LeaveFunction: traverseValueToInterfaceLeave}
	_, err := def.TraverseValues(t, tp)
	Central.Log.Debugf("Adapt interface ready: %v", err)
	return err
}
