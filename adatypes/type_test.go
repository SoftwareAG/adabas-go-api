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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeOptions(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeString, "AB")
	assert.Equal(t, "AB", adaType.Name())
	assert.Equal(t, uint32(1), adaType.Length())
	assert.Equal(t, "", adaType.Option())

	adaType.AddOption(FieldOptionDE)
	assert.Equal(t, "DE", adaType.Option())
	adaType.AddOption(FieldOptionUQ)
	assert.Equal(t, "UQ DE", adaType.Option())
	adaType.AddOption(FieldOptionNU)
	adaType.AddOption(FieldOptionNB)
	assert.Equal(t, "UQ NU DE NB", adaType.Option())
	adaType.AddOption(FieldOptionHF)
	adaType.ClearOption(FieldOptionDE)
	assert.Equal(t, "UQ NU HF NB", adaType.Option())
	adaType.SetLevel(1)

	assert.Equal(t, " 1, AB, 1, A ,UQ,NU,HF,NB ; AB", adaType.String())
	adaType.length = 20
	adaType.fieldType = FieldTypeUInt2
	assert.Equal(t, " 1, AB, 20, B ,UQ,NU,HF,NB ; AB", adaType.String())
	adaType.fieldType = FieldTypeInt2
	assert.Equal(t, " 1, AB, 20, F ,UQ,NU,HF,NB ; AB", adaType.String())
	adaType.fieldType = FieldTypeLBString
	assert.Equal(t, " 1, AB, 20, A ,UQ,NU,HF,NB,LB ; AB", adaType.String())
	assert.Equal(t, "UQ NU HF NB LB", adaType.Option())

}

func TestTypeFlags(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}

	Central.Log.Infof("TEST: %s", t.Name())
	assert.Equal(t, uint32(1), FlagOptionPE.Bit())
	assert.Equal(t, uint32(2), FlagOptionAtomicFB.Bit())
}

func TestTypeReferential(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	refType := NewReferentialType("RE", 1, [2]string{"PK", "FK"}, 1, 2, 1)
	assert.Equal(t, "RE=REFINT(FK,1,PK/DC,UN) ; RE", refType.String())
	refType = NewReferentialType("RE", 1, [2]string{"PK", "FK"}, 1, 1, 2)
	assert.Equal(t, "RE=REFINT(FK,1,PK/DN,UC) ; RE", refType.String())
	refType = NewReferentialType("RX", 1, [2]string{"PK", "FK"}, 1, 0, 0)
	assert.Equal(t, "RX=REFINT(FK,1,PK/DX,UX) ; RX", refType.String())
}

func TestTypeLongName(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	ty := NewLongNameType(FieldTypeByte, "ABC", "XX")
	assert.Equal(t, "ABC", ty.Name())
	assert.Equal(t, "XX", ty.ShortName())
	assert.Equal(t, " 1, XX, 1, F  ; ABC", ty.String())

	ty = NewLongNameType(FieldTypeString, "STRING", "ST")
	assert.Equal(t, "STRING", ty.Name())
	assert.Equal(t, "ST", ty.ShortName())
	assert.Equal(t, " 1, ST, 1, A  ; STRING", ty.String())

}

func TestTypeFlagsSetClear(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	ty := NewLongNameType(FieldTypeByte, "ABC", "XX")
	assert.True(t, ty.HasFlagSet(FlagOptionToBeRemoved))
	assert.False(t, ty.HasFlagSet(FlagOptionPart))
	ty.AddFlag(FlagOptionAtomicFB)
	ty.AddFlag(FlagOptionPart)
	assert.True(t, ty.HasFlagSet(FlagOptionPart))
	assert.True(t, ty.HasFlagSet(FlagOptionAtomicFB))
	ty.RemoveFlag(FlagOptionToBeRemoved)
	assert.False(t, ty.HasFlagSet(FlagOptionToBeRemoved))
	ty.RemoveFlag(FlagOptionAtomicFB)
	ty.RemoveFlag(FlagOptionPart)
	assert.False(t, ty.HasFlagSet(FlagOptionToBeRemoved))
	assert.False(t, ty.HasFlagSet(FlagOptionAtomicFB))
	assert.False(t, ty.HasFlagSet(FlagOptionPart))
	assert.Equal(t, uint32(0), ty.flags)
	ty.AddFlag(FlagOptionToBeRemoved)
	assert.True(t, ty.HasFlagSet(FlagOptionToBeRemoved))
	assert.False(t, ty.HasFlagSet(FlagOptionAtomicFB))

}

func TestRedefinitionType(t *testing.T) {
	err := initLogWithFile("type.log")
	if !assert.NoError(t, err) {
		return
	}
	Central.Log.Infof("TEST: %s", t.Name())
	adaType := NewType(FieldTypeString, "AB", 10)
	redType := NewRedefinitionType(adaType)
	assert.Equal(t, uint32(10), redType.Length())
	assert.Equal(t, "A", adaType.fieldType.FormatCharacter())
	assert.Equal(t, "A", redType.MainType.Type().FormatCharacter())
	assert.Equal(t, uint32(10), redType.MainType.Length())

}
