// +build !adalnk

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
	"os"
	"os/user"
	"time"
)

// NewAdabasID create a new Adabas ID instance
func NewAdabasID() *ID {
	adaid := ID{level: 3, Pid: uint32(os.Getpid()), size: adabasIDSize}
	curUser, err := user.Current()
	if err != nil {
		copy(adaid.User[:], ([]byte("Unknown"))[:8])
	} else {
		copy(adaid.User[:], ([]byte(curUser.Username))[:8])
	}
	host, err := os.Hostname()
	if err != nil {
		copy(adaid.Node[:], ([]byte("Unknown"))[:8])
	} else {
		copy(adaid.Node[:], ([]byte(host))[:8])
	}
	return &adaid
}

// CallAdabas this method sends the call to the database
func (adabas *Adabas) CallAdabas() (err error) {
	defer adatypes.TimeTrack(time.Now(), "CallAdabas "+string(adabas.Acbx.Acbxcmd[:]))

	adatypes.Central.Log.Debugf("Call Adabas %p %s\n%v", adabas, adabas.URL.String(), adabas.ID.String())
	adatypes.LogMultiLineString(adabas.Acbx.String())
	if !validAcbxCommand(adabas.Acbx.Acbxcmd) {
		return adatypes.NewGenericError(2, string(adabas.Acbx.Acbxcmd[:]))
	}
	err = adabas.callRemoteAdabas()
	if err != nil {
		return
	}
	if !validAcbxCommand(adabas.Acbx.Acbxcmd) {
		adatypes.Central.Log.Debugf("Invalid Adabas command received: %s", string(adabas.Acbx.Acbxcmd[:]))
		return adatypes.NewGenericError(3, string(adabas.Acbx.Acbxcmd[:]))
	}
	if adabas.Acbx.Acbxrsp > AdaEOF {
		return NewError(adabas)
	}
	return nil
}
