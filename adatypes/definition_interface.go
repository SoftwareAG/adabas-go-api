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
	"fmt"
	"reflect"
)

type valueInterface struct {
	valStack   *Stack
	curVal     reflect.Value
	invalid    string
	fieldNames map[string][]string
}

func traverseValueToInterface(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	Central.Log.Debugf("Adapt value to interface: %s", adaValue.Type().Name())
	if adaValue.Type().HasFlagSet(FlagOptionMUGhost) {
		Central.Log.Debugf("Skip because in MU ghost %s", adaValue.Type().Name())
		return Continue, nil
	}
	tp := x.(*valueInterface)
	if tp.invalid != "" {
		Central.Log.Debugf("Skip because no interface part %s invalid=%s", adaValue.Type().Name(), tp.invalid)
		return Continue, nil
	}
	Central.Log.Debugf("Work on value %s to interface", adaValue.Type().Name())
	v := tp.curVal
	if v.Kind() == reflect.Slice {
		v = v.Index(int(adaValue.PeriodIndex() - 1))
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		Central.Log.Debugf("Slice pe=%d kind=%s", adaValue.PeriodIndex(), v.Kind())
	}
	Central.Log.Debugf("Current struct value %v", v.Type())
	var f reflect.Value
	if fn, ok := tp.fieldNames[adaValue.Type().Name()]; ok {
		if len(fn) == 0 {
			panic(fmt.Sprintf("Format name error %s -> %d", adaValue.Type().Name(), len(fn)))
		}
		f = v.FieldByName(fn[len(fn)-1])
	} else {
		f = v.FieldByName(adaValue.Type().Name())
	}
	if adaValue.Type().Type() == FieldTypeMultiplefield {
		t := v.Type()
		st, ok := t.FieldByName(adaValue.Type().Name())
		Central.Log.Debugf("Found MU field t=%v st=%v ok=%v", t, st.Name, ok)
		if !ok {
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
			err := SetValueData(entry, sv.Elements[i].Values[0])
			if err != nil {
				return EndTraverser, err
			}
			elemSlice.Index(i).Set(entry.Elem())
		}
		f.Set(elemSlice)
		//		tp.curVal = elemSlice
		Central.Log.Debugf("Set %s new base struct value %v", adaValue.Type().Name(), tp.curVal.Type())
		return Continue, nil
	}
	if adaValue.Type().Type() == FieldTypePeriodGroup {
		t := v.Type()
		st, ok := t.FieldByName(adaValue.Type().Name())
		Central.Log.Debugf("Found PE field t=%v st=%v ok=%v", t, st.Name, ok)
		if !ok {
			return Continue, nil
		}
		Central.Log.Debugf("Add slice %s %v -> %v %s", st.Name, st.Type, st.Type.String(), st.Type.Kind())
		stt := st.Type.Elem()
		Central.Log.Debugf("Use type %s %s %v kind=%v", stt.Name(), stt.String(), stt, stt.Kind())
		sv := adaValue.(*StructureValue)
		elemSlice := reflect.MakeSlice(reflect.SliceOf(stt), sv.NrElements(), 10)
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
		tp.valStack.Push(tp.curVal)
		tp.curVal = elemSlice
		Central.Log.Debugf("Set %s new base PE slice value %v", adaValue.Type().Name(), tp.curVal.Type())
		return Continue, nil
	}
	// if adaValue.Type().Type() == FieldTypeGroup {
	// 	t := v.Type()
	// 	st, ok := t.FieldByName(adaValue.Type().Name())
	// 	Central.Log.Debugf("Found PE field t=%v st=%v ok=%v", t, st.Name, ok)
	// 	if !ok {
	// 		return Continue, nil
	// 	}
	// 	Central.Log.Debugf("Add slice %s %v -> %v %s", st.Name, st.Type, st.Type.String(), st.Type.Kind())
	// 	stt := st.Type.Elem()
	// 	Central.Log.Debugf("Use type %s %s %v kind=%v", stt.Name(), stt.String(), stt, stt.Kind())
	// 	//sv := adaValue.(*StructureValue)
	// 	entry := reflect.New(st.Type)
	// 	Central.Log.Debugf("New entry %s -> %s ok=%v i=%v", entry.Type().Name(), entry.Type().String(), entry.IsValid(), entry.CanInterface())
	// 	f.Set(entry)
	// 	Central.Log.Debugf("Push slice to stack for %s", adaValue.Type().Name())
	// 	tp.valStack.Push(tp.curVal)
	// 	tp.curVal = entry
	// 	return Continue, nil
	// }
	Central.Log.Debugf("No MU or PE, check kind=%v of %s", f.Kind(), adaValue.Type().Name())
	switch f.Kind() {
	case reflect.Slice:
		Central.Log.Debugf("Found slice on %s %d,%d", adaValue.Type().Name(), adaValue.PeriodIndex(), adaValue.MultipleIndex())
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
		tp.valStack.Push(tp.curVal)
		tp.curVal = f.Elem()
		Central.Log.Debugf("Set %s new base Ptr value %v", adaValue.Type().Name(), tp.curVal.Type())
		return Continue, nil

	default:
		if f.IsValid() {
			switch f.Kind() {
			case reflect.Int64, reflect.Int32:
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
				Central.Log.Infof("Unkown kind: %v", f.Kind())
			}
			Central.Log.Debugf("%s=%v->%v", adaValue.Type().Name(), v, f)
		} else {
			Central.Log.Debugf("%s is invalid, kind = %v", adaValue.Type().Name(), f.Kind())
			if adaValue.Type().IsStructure() {
				tp.invalid = adaValue.Type().Name()
			}
		}
	}

	return Continue, nil
}

func traverseValueToInterfaceLeave(adaValue IAdaValue, x interface{}) (result TraverseResult, err error) {
	Central.Log.Debugf("Leave value to interface: %s", adaValue.Type().Name())
	if adaValue.Type().IsStructure() {
		tp := x.(*valueInterface)
		if tp.invalid != "" {
			Central.Log.Debugf("Skip leave because no interface part %s invalid=%s", adaValue.Type().Name(), tp.invalid)
			if adaValue.Type().Name() == tp.invalid {
				tp.invalid = ""
			}
			return Continue, nil
		}
		Central.Log.Debugf("Current %s struct value %v", adaValue.Type().Name(), tp.curVal.Type())
		Central.Log.Debugf("Pop from stack for %s type=%s", adaValue.Type().Name(), adaValue.Type().Type().name())
		if adaValue.Type().Type() != FieldTypeMultiplefield {
			v, err := tp.valStack.Pop()
			if err != nil {
				return EndTraverser, err
			}
			tp.curVal = v.(reflect.Value)
			Central.Log.Debugf("Reset %s to struct value %v", adaValue.Type().Name(), tp.curVal.Type())
		}
	}
	return Continue, nil
}

// AdaptInterfaceFields adapt field value to interface field
func (def *Definition) AdaptInterfaceFields(v reflect.Value, fm map[string][]string) error {
	Central.Log.Debugf("Adapt interface")
	if fm == nil {
		panic("Format map not initialized")
	}
	tp := &valueInterface{curVal: v.Elem(), valStack: NewStack(), fieldNames: fm}
	t := TraverserValuesMethods{EnterFunction: traverseValueToInterface, LeaveFunction: traverseValueToInterfaceLeave}
	_, err := def.TraverseValues(t, tp)
	Central.Log.Debugf("Adapt interface ready: %v", err)
	return err
}
