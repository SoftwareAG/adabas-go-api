/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMapRepository(t *testing.T) {
	f := initTestLogWithFile(t, "map_repositories.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	AddMapRepository(NewAdabas(24), 4)
	defer DelMapRepository(NewAdabas(24), 4)
	adabas := NewAdabas(0)
	defer adabas.Close()
	adabasMap, err := SearchMapRepository(adabas, "EMPLOYEES-NAT-DDM")
	assert.NoError(t, err)
	assert.NotNil(t, adabasMap)

}

func TestMapRepositoryReadAll(t *testing.T) {
	f := initTestLogWithFile(t, "map_repositories.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	adabas := NewAdabas(24)
	defer adabas.Close()
	mr := NewMapRepository(adabas, 4)
	adabasMaps, err := mr.LoadAllMaps(adabas)
	assert.NoError(t, err)
	assert.NotNil(t, adabasMaps)
	assert.NotEqual(t, 0, len(adabasMaps))
	for _, m := range adabasMaps {
		fmt.Println(m.Name)
	}
}
