//go:build !adalnk && !windows
// +build !adalnk,!windows

/*
* Copyright © 2018-2021 Software AG, Darmstadt, Germany and/or its licensors
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

// NewAdaIPC fake nil instance to disable driver
func NewAdaIPC(URL *URL, ID *ID) Driver {
	return nil
}

// NewAdabasID create a new unique Adabas ID instance using static data. Instead
// using the current process id a generate unique time stamp and counter version
// of the pid is used.
func NewAdabasID() *ID {
	id := atomic.AddUint32(&idCounter, 1)
	adaid := AID{level: 3, size: adabasIDSize}
	adaid.Timestamp = uint64(time.Now().UnixNano() / 1000)
	adaid.Pid = uint32((adaid.Timestamp - (adaid.Timestamp % 100)) + uint64(id))
	aid := ID{AdaID: &adaid, connectionMap: make(map[string]*Status)}
	curUser, err := user.Current()
	if err != nil {
		copy(adaid.User[:], ([]byte("Unknown"))[:8])
	} else {
		adatypes.Central.Log.Debugf("Create new ID(remote) with %s", curUser)
		copy(adaid.User[:], ([]byte(curUser.Username + "        "))[:8])
	}
	host, err := os.Hostname()
	if err != nil {
		copy(adaid.Node[:], ([]byte("Unknown"))[:8])
	} else {
		adatypes.Central.Log.Debugf("Current host is %s", curUser)
		copy(adaid.Node[:], ([]byte(host + "        "))[:8])
	}
	return &aid
}

// CallAdabas this method sends the call to the Adabas database. It uses only
// remote Adabas calls with ADATCP because this part is not used with native
// AdabasClient library support
func (adabas *Adabas) CallAdabas() (err error) {
	defer TimeTrack(time.Now(), "RCall adabas", adabas)

	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Call Adabas (local disabled) adabasp=%p  %s%v", adabas, adabas.URL.String(), adabas.ID.String())
		adatypes.LogMultiLineString(true, adabas.Acbx.String())
	}
	// check sending Adabas call
	if !validAcbxCommand(adabas.Acbx.Acbxcmd) {
		return adatypes.NewGenericError(2, string(adabas.Acbx.Acbxcmd[:]))
	}
	adabas.Acbx.Acbxrsp = AdaAnact
	adabas.Acbx.Acbxerrc = 0
	adatypes.Central.Log.Debugf("Input Adabas response = %d", adabas.Acbx.Acbxrsp)
	recordBufferResize := uint8(5)
	for {
		err = adabas.callAdabasDriver()
		if err != nil {
			return
		}
		if adatypes.Central.IsDebugLevel() {
			adatypes.LogMultiLineString(true, adabas.Acbx.String())
			if adabas.Acbx.Acbxrsp != AdaNormal {
				if adabas.Acbx.Acbxrsp == AdaSYSBU {
					adatypes.Central.Log.Debugf("%s", adabas.Acbx.String())
					for index := range adabas.AdabasBuffers {
						adatypes.Central.Log.Debugf("%s", adabas.AdabasBuffers[index].String())
					}
				}
			}
		}
		// check received Adabas call
		if !validAcbxCommand(adabas.Acbx.Acbxcmd) {
			adatypes.Central.Log.Debugf("Invalid Adabas command received: %s", string(adabas.Acbx.Acbxcmd[:]))
			return adatypes.NewGenericError(3, string(adabas.Acbx.Acbxcmd[:]))
		}
		if adabas.Acbx.Acbxrsp != AdaRbts || recordBufferResize == 0 {
			break
		}
		recordBufferResize--
		for index := range adabas.AdabasBuffers {
			if adabas.AdabasBuffers[index].abd.Abdid == AbdAQRb {
				adabas.AdabasBuffers[index].extend(8192)
			}
		}
	}

	switch adabas.Acbx.Acbxrsp {
	case AdaAnact, AdaTransactionAborted, AdaSysCe:
		adabas.ID.clearTransactions(adabas.URL.String())
		adabas.ID.changeOpenState(adabas.URL.String(), false)
	}
	if adabas.Acbx.Acbxrsp > AdaEOF {
		return NewError(adabas)
	}
	return nil
}
