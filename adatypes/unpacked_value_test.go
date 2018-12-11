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

package adatypes

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUnpacked(t *testing.T) {
	f, err := initLogWithFile("unpacked.log")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adaType := NewType(FieldTypeUnpacked, "UP")
	adaType.length = 4
	up := newUnpackedValue(adaType)
	fmt.Println("Unpacked value ", up.value)
	up.LongToUnpacked(0, 4, false)
	fmt.Println("Unpacked value 0 ", up.value)
}
