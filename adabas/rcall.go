// +build !adalnk !cgo

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
	"os"
	"os/user"
	"sync/atomic"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

var idCounter uint32

// NewAdabasID create a new unique Adabas ID instance using static data. Instead
// using the current process id a generate unique time stamp and counter version
// of the pid is used.
func NewAdabasID() *ID {
	id := atomic.AddUint32(&idCounter, 1)
	adaid := AID{level: 3, size: adabasIDSize}
	adaid.Timestamp = uint64(time.Now().Unix())
	adaid.Pid = uint32((adaid.Timestamp - (adaid.Timestamp % 100)) + uint64(id))
	aid := ID{AdaID: &adaid, connectionMap: make(map[string]*Status)}
	curUser, err := user.Current()
	adatypes.Central.Log.Debugf("Create new ID(remote) with %s", curUser)
	if err != nil {
		copy(adaid.User[:], ([]byte("Unknown"))[:8])
	} else {
		copy(adaid.User[:], ([]byte(curUser.Username))[:8])
	}
	host, err := os.Hostname()
	adatypes.Central.Log.Debugf("Current host is %s", curUser)
	if err != nil {
		copy(adaid.Node[:], ([]byte("Unknown"))[:8])
	} else {
		copy(adaid.Node[:], ([]byte(host))[:8])
	}
	return &aid
}

// CallAdabas this method sends the call to the Adabas database. It uses only
// remote Adabas calls with ADATCP because this part is not used with native
// AdabasClient library support
func (adabas *Adabas) CallAdabas() (err error) {
	defer adatypes.TimeTrack(time.Now(), "CallAdabas "+string(adabas.Acbx.Acbxcmd[:]))

	adatypes.Central.Log.Debugf("Call Adabas (local disabled) adabasp=%p  %s\n%v", adabas, adabas.URL.String(), adabas.ID.String())
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
