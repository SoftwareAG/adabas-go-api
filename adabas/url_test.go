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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	dbidURL := NewURLWithDbid(123)
	assert.Equal(t, "123", dbidURL.String())
	assert.Equal(t, ":0", dbidURL.URL())
	URL, err := NewURL("124(adatcp://host:1234)")
	assert.NoError(t, err)
	assert.Equal(t, "124(adatcp://host:1234)", URL.String())
	assert.Equal(t, "host:1234", URL.URL())
	URL, err = NewURL("124(adatcp://host:xx)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: '124(adatcp://host:xx)' is no valid database id", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("444(tcpip://host:xx)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: '444(tcpip://host:xx)' is no valid database id", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("222(tcpip://host:1234)")
	assert.NoError(t, err)
	assert.Equal(t, "222(tcpip://host:1234)", URL.String())
	assert.Equal(t, "host:1234", URL.URL())
	URL, err = NewURL("333(adatcp://host:123)")
	assert.NoError(t, err)
	assert.Equal(t, "333(adatcp://host:123)", URL.String())
	assert.Equal(t, "host:123", URL.URL())
	URL, err = NewURL("(abc://host:123)")
	assert.Error(t, err)
	URL, err = NewURL("a(xxx://abc:)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: 'a(xxx://abc:)' is no valid database id", err.Error())
	URL, err = NewURL("123(xxx://abc:a)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: '123(xxx://abc:a)' is no valid database id", err.Error())

}
