//go:build (adalnk && cgo) || windows
// +build adalnk,cgo windows

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
	AdaID := AID{level: 3, size: adabasIDSize}
	aid := ID{AdaID: &AdaID, connectionMap: make(map[string]*Status)}
	//	C.lnk_get_adabas_id(adabasIDSize, (*C.uchar)(unsafe.Pointer(&AdaID)))
	curUser, err := user.Current()
	if err != nil {
		adatypes.Central.Log.Debugf("Error evaluing user")
		copy(AdaID.User[:], ([]byte("Unknown "))[:8])
	} else {
		adatypes.Central.Log.Debugf("Create new ID(local) with %s", curUser.Username)
		copy(AdaID.User[:], ([]byte(curUser.Username + "        "))[:8])
	}
	host, err := os.Hostname()
	if err != nil {
		adatypes.Central.Log.Debugf("Error evaluing host")
		copy(AdaID.Node[:], ([]byte("Unknown"))[:8])
	} else {
		adatypes.Central.Log.Debugf("Current host is %s", curUser)
		copy(AdaID.Node[:], ([]byte(host + "        "))[:8])
	}
	id := atomic.AddUint32(&idCounter, 1)
	adatypes.Central.Log.Debugf("Create new ID(local) with %v", AdaID.Node)
	AdaID.Timestamp = uint64(time.Now().UnixNano() / 1000)
	AdaID.Pid = uint32((AdaID.Timestamp - (AdaID.Timestamp % 100)) + uint64(id))
	adatypes.Central.Log.Debugf("Create Adabas ID: %d -> %s", AdaID.Pid, aid.String())
	return &aid
}

// CallAdabas this method sends the call to the Adabas database. It uses
// native local Adabas calls because this part is used with native
// AdabasClient library support
func (adabas *Adabas) CallAdabas() (err error) {
	defer TimeTrack(time.Now(), "Call adabas", adabas)
	s := adabas.status
	s.lock.Lock()
	defer s.lock.Unlock()

	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Send calling CC %c%c adabasp=%p URL=%s Adabas ID=%v",
			adabas.Acbx.Acbxcmd[0], adabas.Acbx.Acbxcmd[1],
			adabas, adabas.URL.String(), adabas.ID.String())
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
	// Loop and increase if Record Buffer size is too small
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

	// Clear transactions if response code != EOF or ADANORMAL
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
