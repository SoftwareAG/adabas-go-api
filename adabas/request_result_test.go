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

package adabas

import (
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fields struct {
	values []*ResultRecord
}

func generateOneFields(t *testing.T) fields {
	oneField := fields{}
	v, err := adatypes.NewType(adatypes.FieldTypeByte, "AA").Value()
	assert.NoError(t, err)
	record := &ResultRecord{Value: []adatypes.IAdaValue{v}}
	oneField.values = append(oneField.values, record)
	return oneField
}

func TestRequestResult_NrRecords(t *testing.T) {
	noFields := fields{}
	oneField := generateOneFields(t)
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"NoField", noFields, 0},
		{"OneField", oneField, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestResult := &RequestResult{
				Values: tt.fields.values,
			}
			if got := requestResult.NrRecords(); got != tt.want {
				t.Errorf("RequestResult.NrRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestResult_String(t *testing.T) {
	noFields := fields{}
	oneField := generateOneFields(t)
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"NoFields", noFields, ""},
		{"OneField", oneField, "  AA = > 0 <\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestResult := &RequestResult{
				Values: tt.fields.values,
			}
			if got := requestResult.String(); got != tt.want {
				t.Errorf("RequestResult.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
