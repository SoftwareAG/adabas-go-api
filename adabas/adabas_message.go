/*
* Copyright © 2018-2020 Software AG, Darmstadt, Germany and/or its licensors
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

//type errorCode int

// getAdabasMessage get the current Adabas message dependent on the response and sub code
func (adabas *Adabas) getAdabasMessage() []string {
	message := make([]string, 2)
	msgCode := fmt.Sprintf("%s%02X%03X", messagePrefix, adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxerrc)
	message[0] = msgCode
	msg := adatypes.Translate(adatypes.Language(), msgCode)
	if msg == "" && adabas.Acbx.Acbxerrc > 0 {
		msgCode = fmt.Sprintf("%s%02X%03X", messagePrefix, adabas.Acbx.Acbxrsp, 0)
		msg = adatypes.Translate(adatypes.Language(), msgCode)
	}
	if msg == "" {
		msg = fmt.Sprintf("Unknown error response %d subcode %d (%s)", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxerrc, msgCode)
	}
	msg = fmt.Sprintf("%s (rsp=%d,subrsp=%d,dbid=%s,file=%d)", msg, adabas.Acbx.Acbxrsp,
		adabas.Acbx.Acbxerrc, adabas.URL.String(), adabas.Acbx.Acbxfnr)
	if adabas.Acbx.Acbxrsp > 3 {
		adatypes.Central.Log.Infof("Error message %s", msg)
		adatypes.Central.Log.Infof("Add1: %v", adabas.Acbx.Acbxadd1)
		adatypes.Central.Log.Infof("Add2: %v", adabas.Acbx.Acbxadd2)
	}
	message[1] = msg
	return message
}

// Error error message with code and time
type Error struct {
	When    time.Time
	Acbx    Acbx
	URL     URL
	Code    string
	Message string
}

// NewError Create new Adabas errror
func NewError(adbas *Adabas) *Error {
	msgCode := fmt.Sprintf("%s%02X%03X", messagePrefix, adbas.Acbx.Acbxrsp, adbas.Acbx.Acbxerrc)
	e := &Error{When: time.Now(), URL: *adbas.URL, Code: msgCode, Acbx: *adbas.Acbx}
	// msg := adatypes.Translate(adatypes.Language(), msgCode)
	// if msg == "" && adbas.Acbx.Acbxerrc > 0 {
	// 	msgCode = fmt.Sprintf("%s%02X%03X", messagePrefix, adbas.Acbx.Acbxrsp, 0)
	// 	msg = adatypes.Translate(adatypes.Language(), msgCode)
	// }
	// suffix := acbxSuffix(adbas.URL, adbas.Acbx)
	// if msg == "" {
	// 	msg = fmt.Sprintf("Unknown error response %d subcode %d&s", adbas.Acbx.Acbxrsp, adbas.Acbx.Acbxerrc, suffix)
	// }
	// msg = fmt.Sprintf("%s%s", msg, suffix)
	// adatypes.Central.Log.Debugf("Adabas error message created:[%s] %s", msgCode, msg)
	e.Message = e.Translate(adatypes.Language())
	return e
}

func acbxSuffix(URL *URL, acbx *Acbx) string {
	return fmt.Sprintf(" (rsp=%d,subrsp=%d,dbid=%s,file=%d)", acbx.Acbxrsp,
		acbx.Acbxerrc, URL.String(), acbx.Acbxfnr)

}

// Response return the response code of adabas call
func (e Error) Response() uint16 {
	return e.Acbx.Acbxrsp
}

// SubResponse return the sub response code of adabas call
func (e Error) SubResponse() uint16 {
	return e.Acbx.Acbxerrc
}

// Addition2 return the additon 2 sub field of adabas call
func (e Error) Addition2() [4]byte {
	return e.Acbx.Acbxadd2
}

// Error error main interface function, providing message error code and message
func (e Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

// Translate translate to language
func (e Error) Translate(lang string) string {
	msg := adatypes.Translate(lang, e.Code)
	if msg == "" && e.Acbx.Acbxerrc > 0 {
		msg = adatypes.Translate(lang, e.Code)
	}
	suffix := acbxSuffix(&e.URL, &e.Acbx)
	if msg == "" {
		msg = fmt.Sprintf("%s%s", adatypes.Translate(lang, "ADAGEFFFFF"), suffix)
	} else {
		msg = fmt.Sprintf("%s%s", msg, suffix)
	}
	return msg
}
