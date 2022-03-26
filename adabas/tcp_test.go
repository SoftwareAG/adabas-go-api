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

// Package tcp contains Adabas ADATCP call functions
package adabas

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcpHTON(t *testing.T) {
	x := adatcpTCPClientHTON8(88)
	fmt.Println("X: ", x)
	y := adatcpTCPClientHTON8(x)
	fmt.Println("Y: ", y)
}

func TestNoConnect(t *testing.T) {
	var user [8]byte
	var node [8]byte
	// No URL returns nil connection instance
	connection := NewAdaTCP(nil, binary.LittleEndian, user, node, 0, 0)
	assert.Nil(t, connection)
	copy(user[:], []byte("User_001"))
	copy(node[:], []byte("Node_001"))

	// No Connect() called and a error is returned if Disconnect() is called
	url, _ := NewURL("1(adatcp://localhost:12345)")
	connection = NewAdaTCP(url, binary.LittleEndian, user, node, 0, 0)
	err := connection.Disconnect()
	assert.Error(t, err)
}

func TestFailConnect(t *testing.T) {
	var user [8]byte
	var node [8]byte

	// No Connect() called and a error is returned if Disconnect() is called
	url, _ := NewURL("1(adatcp://xx:12345)")
	connection := NewAdaTCP(url, binary.LittleEndian, user, node, 0, 0)
	err := connection.tcpConnect()
	assert.Error(t, err)
	err = connection.Disconnect()
	assert.Error(t, err)
}
