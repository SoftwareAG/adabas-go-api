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
	"bytes"
	"encoding/binary"

	"golang.org/x/text/encoding"
)

// Platform platform definition structure
type Platform struct {
	architecture    byte
	systemCharset   encoding.Decoder
	unicodeEncoding encoding.Decoder
	spaceByte       byte
	order           binary.ByteOrder
}

// PlatformMainframe Mainframe architecture byte
const PlatformMainframe = 0x0

// PlatformLUWHighOrder LUW high order architecture byte
const PlatformLUWHighOrder = 0x20

// PlatformLUWLowOrder LUW low order architecture byte
const PlatformLUWLowOrder = 0x21

// NewPlatformIsl create a new platform instance
func NewPlatformIsl(isl uint64) *Platform {
	arch := byte((isl >> (3 * 8)) & 0xff)
	return NewPlatform(arch)
}

// NewPlatform create a new platform instance
func NewPlatform(arch byte) *Platform {
	space := byte(0x20)
	if arch == PlatformMainframe {
		space = byte(0x40)
	}
	var order binary.ByteOrder
	if arch == PlatformLUWLowOrder {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}
	pl := &Platform{architecture: arch, spaceByte: space, order: order}
	Central.Log.Debugf("New platform mainframe=%v(%d)", pl.IsMainframe(), arch)
	return pl
}

// IsMainframe returns True if the platform is a Mainframe platform
func (platform *Platform) IsMainframe() bool {
	platformIdentifier := platform.architecture & 0xF0
	return platformIdentifier == PlatformMainframe
}

// String representation of platform identifier
func (platform *Platform) String() string {
	var buffer bytes.Buffer
	if platform.IsMainframe() {
		buffer.WriteString("Mainframe")
	} else {
		buffer.WriteString("Open System")
	}
	buffer.WriteRune(',')
	if platform.order == binary.LittleEndian {
		buffer.WriteString("Low Order")
	} else {
		buffer.WriteString("High Order")
	}

	return buffer.String()
}
