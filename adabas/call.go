// +build adalnk,cgo

/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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
	"os"
	"os/user"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
)

/*
#if !defined(__unix__) && !defined(__hpux) && !defined(__APPLE__) &&	\
  !defined(__IBMC__) && !defined(__sun)
#define __unix__ 0
#else
#ifndef __unix__
#define __unix__ 1
#endif
#endif
#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <string.h>
#if !__unix__
#include <errno.h>
#else
#include <sys/errno.h>
#endif
#include "adabasx.h"
long flags = 0;

// Initialize ABD array with number of ABD
PABD *create_abd(int num_abd)
{
	int i;
	PABD *pabd = (PABD *)malloc(num_abd * sizeof(PABD *));
	for (i = 0; i < num_abd; i++)
	{
		pabd[i] = NULL;
	}
	return pabd;
}

// Destroy ABD array
void destroy_abd(PABD *pabd, int num_abd)
{
	int i;
	for (i = 0; i < num_abd; i++)
	{
		if (pabd[i] != NULL)
		{
			if (pabd[i]->abdaddr != NULL)
			{
				free(pabd[i]->abdaddr);
			}
			free(pabd[i]);
		}
	}
	free(pabd);
}

// Adabas interface for Go to call ACBX Adabas calls
int go_eadabasx(ADAID_T *adabas_id, PACBX acbx, int num_abd, PABD *abd)
{
	register int i;
	int rsp;
	char *buffer;
	uint32_t flag;
	uint32_t timeOut;
	char user[9];
	char node[9];
	// Here I call the ACBX enabled Adabas function of adabasx
	{
		lnk_set_adabas_id((unsigned char *)(adabas_id));
		rsp = adabasx(acbx, num_abd, abd);
	}
	return (rsp);
}

// Malloc C based memory and copy Go based ABD to C based because of pointer references
void copy_to_abd(PABD *pabd, int index, PABD x, char *data, uint32_t size)
{
	PABD dest_pabd = pabd[index] = malloc(L_ABD);
	if (dest_pabd == NULL)
	{
		exit(10);
	}
	if (x == NULL)
	{
		exit(10);
	}
	memcpy(dest_pabd, x, L_ABD);
	if (data != NULL)
	{
		dest_pabd->abdaddr = malloc(size);
		memcpy(dest_pabd->abdaddr, data, size);
		dest_pabd->abdsize = size;
	}
	else
	{
		dest_pabd->abdsize = 0;
		dest_pabd->abdaddr = NULL;
	}
}

// Copy C based ABD to Go based because of pointer references and free memory
void copy_from_abd(PABD *pabd, int index, PABD x, char *data, uint32_t size)
{
	PABD dest_pabd = pabd[index];
	memcpy(x, dest_pabd, L_ABD);
	if ((data != NULL) && (dest_pabd->abdrecv > 0))
	{
		memcpy(data, dest_pabd->abdaddr, size);
	}
	if (dest_pabd->abdaddr != NULL)
	{
		free(dest_pabd->abdaddr);
		dest_pabd->abdaddr = NULL;
	}
	free(pabd[index]);
	pabd[index] = NULL;
}
*/
import "C"

var idCounter uint32

// NewAdabasID create a new Adabas ID instance
func NewAdabasID() *ID {
	AdaID := AID{level: 3, size: adabasIDSize}
	aid := ID{AdaID: &AdaID, connectionMap: make(map[string]*Status)}
	//	C.lnk_get_adabas_id(adabasIDSize, (*C.uchar)(unsafe.Pointer(&AdaID)))
	curUser, err := user.Current()
	adatypes.Central.Log.Debugf("Create new ID(local) with %s", curUser.Username)
	if err != nil {
		copy(AdaID.User[:], ([]byte("Unknown"))[:8])
	} else {
		copy(AdaID.User[:], ([]byte(curUser.Username + "        "))[:8])
	}
	host, err := os.Hostname()
	adatypes.Central.Log.Debugf("Current host is %s", curUser)
	if err != nil {
		copy(AdaID.Node[:], ([]byte("Unknown"))[:8])
	} else {
		copy(AdaID.Node[:], ([]byte(host + "        "))[:8])
	}
	id := atomic.AddUint32(&idCounter, 1)
	adatypes.Central.Log.Debugf("Create new ID(local) with %v", AdaID.Node)
	AdaID.Timestamp = uint64(time.Now().Unix())
	AdaID.Pid = uint32((AdaID.Timestamp - (AdaID.Timestamp % 100)) + uint64(id))
	// fmt.Printf("Create Adabas ID: %d -> %s", AdaID.Pid, aid.String())
	adatypes.Central.Log.Infof("Create Adabas ID: %d -> %s", AdaID.Pid, aid.String())
	return &aid
}

func (adabasBuffer *Buffer) createCAbd(pabdArray *C.PABD, index int) {
	adatypes.Central.Log.Debugf("Copy Metadata index=%d", index)
	var pbuffer *C.char
	if len(adabasBuffer.buffer) == 0 {
		pbuffer = nil
		adabasBuffer.abd.Abdsize = 0
	} else {
		pbuffer = (*C.char)(unsafe.Pointer(&adabasBuffer.buffer[0]))
		adabasBuffer.abd.Abdsize = uint64(len(adabasBuffer.buffer))
		adatypes.Central.Log.Debugf("Set Adabas buffer size=%d", adabasBuffer.abd.Abdsize)
	}
	cabd := adabasBuffer.abd
	C.copy_to_abd(pabdArray, C.int(index), (C.PABD)(unsafe.Pointer(&cabd)),
		pbuffer, C.uint32_t(uint32(adabasBuffer.abd.Abdsize)))
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("C ABD %c: send=%d", cabd.Abdid, cabd.Abdsend)
		adatypes.LogMultiLineString(adatypes.FormatByteBuffer("Buffer content:", adabasBuffer.buffer))
	}

}

func (adabasBuffer *Buffer) putCAbd(pabdArray *C.PABD, index int) {
	adatypes.Central.Log.Debugf("%d: receive index %c len=%d", index, adabasBuffer.abd.Abdid, len(adabasBuffer.buffer))
	var pbuffer *C.char
	adatypes.Central.Log.Debugf("Got buffer len=%d", adabasBuffer.abd.Abdsize)
	if adabasBuffer.abd.Abdsize == 0 {
		pbuffer = nil
	} else {
		pbuffer = (*C.char)(unsafe.Pointer(&adabasBuffer.buffer[0]))
	}
	cabd := adabasBuffer.abd
	var pabd C.PABD
	pabd = (C.PABD)(unsafe.Pointer(&cabd))

	adatypes.Central.Log.Debugf("Work on %c. buffer of size=%d", cabd.Abdid, len(adabasBuffer.buffer))
	C.copy_from_abd(pabdArray, C.int(index), pabd,
		pbuffer, C.uint32_t(uint32(adabasBuffer.abd.Abdsize)))
	adabasBuffer.abd = cabd
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("C ABD %c: recv=%d", cabd.Abdid, cabd.Abdrecv)
		adatypes.LogMultiLineString(adatypes.FormatByteBuffer("Buffer content:", adabasBuffer.buffer))
	}

}

// CallAdabas this method sends the call to the database
func (adabas *Adabas) CallAdabas() (err error) {
	defer adatypes.TimeTrack(time.Now(), "CallAdabas "+string(adabas.Acbx.Acbxcmd[:]))

	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Send calling CC %c%c adabasp=%p URL=%s Adabas ID=%v",
			adabas.Acbx.Acbxcmd[0], adabas.Acbx.Acbxcmd[1],
			adabas, adabas.URL.String(), adabas.ID.String())
		adatypes.LogMultiLineString(adabas.Acbx.String())
	}

	if !validAcbxCommand(adabas.Acbx.Acbxcmd) {
		return adatypes.NewGenericError(2, string(adabas.Acbx.Acbxcmd[:]))
	}
	if adabas.IsRemote() {
		err = adabas.callRemoteAdabas()
		if err != nil {
			return
		}
		if adabas.Acbx.Acbxrsp != 0 {
			if adabas.Acbx.Acbxrsp == 60 {
				fmt.Println(adabas.Acbx.String())
				for index := range adabas.AdabasBuffers {
					fmt.Println(adabas.AdabasBuffers[index].String())
				}
			}
		}
	} else {
		adatypes.Central.Log.Debugf("Call Adabas using native link: %v", adatypes.Central.IsDebugLevel())
		pabdArray := C.create_abd(C.int(len(adabas.AdabasBuffers)))
		adatypes.Central.Log.Debugf("Create ABD : %v", log.GetLevel())
		for index := range adabas.AdabasBuffers {
			adabas.AdabasBuffers[index].abd.Abdrecv = adabas.AdabasBuffers[index].abd.Abdsize
			adabas.AdabasBuffers[index].createCAbd(pabdArray, index)
		}
		ret := int(C.go_eadabasx((*C.ADAID_T)(unsafe.Pointer(adabas.ID.AdaID)),
			(*C.ACBX)(unsafe.Pointer(adabas.Acbx)), C.int(len(adabas.AdabasBuffers)), pabdArray))
		if adatypes.Central.IsDebugLevel() {
			adatypes.Central.Log.Debugf("Send calling CC %c%c adabasp=%p URL=%s Adabas ID=%v",
				adabas.Acbx.Acbxcmd[0], adabas.Acbx.Acbxcmd[1],
				adabas, adabas.URL.String(), adabas.ID.String())
			adatypes.Central.Log.Debugf("Local Adabas call returns: %d", ret)
			adatypes.LogMultiLineString(adabas.Acbx.String())
		}
		for index := range adabas.AdabasBuffers {
			//	adatypes.Central.Log.Debugf(index, ".ABD out : ", adabas.AdabasBuffers[index].abd.Abdsize)
			adabas.AdabasBuffers[index].putCAbd(pabdArray, index)
			if adatypes.Central.IsDebugLevel() {
				adatypes.LogMultiLineString(adabas.AdabasBuffers[index].String())
			}
		}
		adatypes.Central.Log.Debugf("Destroy temporary ABD")
		C.destroy_abd(pabdArray, C.int(len(adabas.AdabasBuffers)))
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
