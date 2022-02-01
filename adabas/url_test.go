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

package adabas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLGeneral(t *testing.T) {
	lErr := initLogWithFile("url.log")
	if !assert.NoError(t, lErr) {
		return
	}
	dbidURL := NewURLWithDbid(123)
	assert.Equal(t, "123", dbidURL.String())
	assert.Equal(t, ":0", dbidURL.URL())
	URL, err := NewURL("124(adatcp://host:1234)")
	assert.NoError(t, err)
	assert.Equal(t, "124(adatcp://host:1234)", URL.String())
	assert.Equal(t, "host:1234", URL.URL())
	assert.Equal(t, "adatcp", URL.Driver)
	URL, err = NewURL("124(adatcp://host:xx)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000072: 'xx' is no valid port number", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("444(tcpip://host:xx)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000072: 'xx' is no valid port number", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("222(tcpip://host:1234)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000099: Given driver 'tcpip' is not supported", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("333(adatcp://host:123)")
	assert.NoError(t, err)
	assert.Equal(t, "333(adatcp://host:123)", URL.String())
	assert.Equal(t, "host:123", URL.URL())
	_, err = NewURL("(abc://host:123)")
	assert.Error(t, err)
	_, err = NewURL("a(xxx://abc:)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: 'a(xxx://abc:)' is no valid database id", err.Error())
	_, err = NewURL("123(xxx://abc:a)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000072: 'a' is no valid port number", err.Error())
	URL, err = NewURL("adatcp://host:123")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "host", URL.Host)
	assert.Equal(t, "host:123", URL.URL())

}

func TestURLDirect(t *testing.T) {
	lErr := initLogWithFile("url.log")
	if !assert.NoError(t, lErr) {
		return
	}
	URL, err := NewURL("adatcp://host:1230")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	URL, err = NewURL("adatcp://host:0")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000072: '0' is no valid port number", err.Error())
	if !assert.Nil(t, URL) {
		return
	}
	URL, err = NewURL("tcpip://host:0")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000070: 'tcpip://host:0' is no valid database id", err.Error())
	if !assert.Nil(t, URL) {
		return
	}
	URL, err = NewURL("201(tcpip://wcphost:30011)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000099: Given driver 'tcpip' is not supported", err.Error())
	if !assert.Nil(t, URL) {
		return
	}
	URL, err = NewURL("adatcp://abchost:1230")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "1(adatcp://abchost:1230)", URL.String())
	assert.Equal(t, "abchost", URL.Host)
	assert.Equal(t, uint32(1230), URL.Port)
	assert.Equal(t, Dbid(1), URL.Dbid)
	URL, err = NewURL("adatcp://host:0")
	assert.Error(t, err)
	if !assert.Nil(t, URL) {
		return
	}
	assert.Equal(t, "ADG0000072: '0' is no valid port number", err.Error())
	URL, err = NewURL("201(tcpip://localhost:50001)")
	assert.Error(t, err)
	if !assert.Nil(t, URL) {
		return
	}
	assert.Equal(t, "ADG0000099: Given driver 'tcpip' is not supported", err.Error())

	URL, err = NewURL("001(adatcp://abchost:1230)")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "1(adatcp://abchost:1230)", URL.String())
	assert.Equal(t, "abchost", URL.Host)
	assert.Equal(t, uint32(1230), URL.Port)
	assert.Equal(t, Dbid(1), URL.Dbid)
	assert.Equal(t, "", URL.GetOption("adb"))

}

func TestURLOptions(t *testing.T) {
	lErr := initLogWithFile("url.log")
	if !assert.NoError(t, lErr) {
		return
	}
	URL, err := NewURL("adatcp://abchost:1230?check=true")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "1(adatcp://abchost:1230)", URL.String())
	assert.Equal(t, "abchost", URL.Host)
	assert.Equal(t, uint32(1230), URL.Port)
	assert.Equal(t, Dbid(1), URL.Dbid)
	assert.Equal(t, "check=true", URL.Options)
	URL, err = NewURL("adatcp://abchost:1230?check=true&test=true&adapt=false")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "1(adatcp://abchost:1230)", URL.String())
	assert.Equal(t, "abchost", URL.Host)
	assert.Equal(t, uint32(1230), URL.Port)
	assert.Equal(t, Dbid(1), URL.Dbid)
	assert.Equal(t, "check=true&test=true&adapt=false", URL.Options)
	assert.Equal(t, "true", URL.GetOption("check"))
	assert.Equal(t, "true", URL.GetOption("test"))
	assert.Equal(t, "false", URL.GetOption("adapt"))
	assert.Equal(t, "", URL.GetOption("adb"))
	URL, err = NewURL("1(adatcp://abchost:1230?check=true&test=true&adapt=false)")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "1(adatcp://abchost:1230)", URL.String())
	assert.Equal(t, "abchost", URL.Host)
	assert.Equal(t, uint32(1230), URL.Port)
	assert.Equal(t, Dbid(1), URL.Dbid)
	assert.Equal(t, "check=true&test=true&adapt=false", URL.Options)
	assert.Equal(t, "true", URL.GetOption("check"))
	assert.Equal(t, "true", URL.GetOption("test"))
	assert.Equal(t, "false", URL.GetOption("adapt"))
	assert.Equal(t, "", URL.GetOption("adb"))
}

func TestURLSecured(t *testing.T) {
	URL, err := NewURL("124(adatcps://host:1234)")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "124(adatcps://host:1234)", URL.String())
	assert.Equal(t, "host:1234", URL.URL())
	assert.Equal(t, "adatcps", URL.Driver)
	URL, err = NewURL("124(adatcps://host:xx)")
	assert.Error(t, err)
	assert.Equal(t, "ADG0000072: 'xx' is no valid port number", err.Error())
	assert.Nil(t, URL)
	URL, err = NewURL("333(adatcps://host:123)")
	assert.NoError(t, err)
	assert.Equal(t, "333(adatcps://host:123)", URL.String())
	assert.Equal(t, "host:123", URL.URL())
	URL, err = NewURL("adatcps://host:123")
	assert.NoError(t, err)
	if !assert.NotNil(t, URL) {
		return
	}
	assert.Equal(t, "host", URL.Host)
	assert.Equal(t, "host:123", URL.URL())
	assert.Equal(t, "adatcps", URL.Driver)

}
