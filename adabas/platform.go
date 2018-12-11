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

// NewPlatform create a new platform instance
func NewPlatform(arch byte) Platform {
	return Platform{architecture: arch}
}
