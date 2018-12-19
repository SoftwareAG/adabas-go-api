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
	"fmt"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const messagePrefix = "ADAGE"

type errorCode int

// getAdabasMessage get the current Adabas message dependent on the response and sub code
func (adabas *Adabas) getAdabasMessage() []string {
	message := make([]string, 2)
	msgCode := fmt.Sprintf("%s%02X%03X", messagePrefix, adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxerrc)
	message[0] = msgCode
	msg := adatypes.Translate("en", msgCode)
	if msg == "" && adabas.Acbx.Acbxerrc > 0 {
		msgCode = fmt.Sprintf("%s%02X%03X", messagePrefix, adabas.Acbx.Acbxrsp, 0)
		msg = adatypes.Translate("en", msgCode)
	}
	if msg == "" {
		msg = fmt.Sprintf("Unknown error response %d subcode %d (%s)", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxerrc, msgCode)
	}
	msg = fmt.Sprintf("%s (rsp=%d,subrsp=%d,dbid=%s,file=%d)", msg, adabas.Acbx.Acbxrsp,
		adabas.Acbx.Acbxerrc, adabas.URL.String(), adabas.Acbx.Acbxfnr)
	message[1] = msg
	return message
}

// Error error message with code and time
type Error struct {
	When        time.Time
	Code        string
	Message     string
	Response    uint16
	SubResponse uint16
	Addition2   [4]byte
}

// NewError create new Adabas errror
func NewError(adbas *Adabas) *Error {
	msgCode := fmt.Sprintf("%s%02X%03X", messagePrefix, adbas.Acbx.Acbxrsp, adbas.Acbx.Acbxerrc)
	msg := adatypes.Translate("en", msgCode)
	if msg == "" && adbas.Acbx.Acbxerrc > 0 {
		msgCode = fmt.Sprintf("%s%02X%03X", messagePrefix, adbas.Acbx.Acbxrsp, 0)
		msg = adatypes.Translate("en", msgCode)
	}
	if msg == "" {
		msg = fmt.Sprintf("Unknown error response %d subcode %d (%s)", adbas.Acbx.Acbxrsp, adbas.Acbx.Acbxerrc, msgCode)
	}
	msg = fmt.Sprintf("%s (rsp=%d,subrsp=%d,dbid=%s,file=%d)", msg, adbas.Acbx.Acbxrsp,
		adbas.Acbx.Acbxerrc, adbas.URL.String(), adbas.Acbx.Acbxfnr)
	//fmt.Println(adbas.Acbx.String())
	return &Error{When: time.Now(), Code: msgCode, Message: msg, Response: adbas.Acbx.Acbxrsp, SubResponse: adbas.Acbx.Acbxerrc, Addition2: adbas.Acbx.Acbxadd2}
}

// Error error main interface function, providing message error code and message
func (e Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}
