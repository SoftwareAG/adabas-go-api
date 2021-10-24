//go:build adalnk && cgo
// +build adalnk,cgo

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
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
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

typedef struct credential {
  char *user;
  char *pwd;
} CREDENTIAL;
#define SIZEOF_CREDENTIAL (sizeof(CREDENTIAL))

CREDENTIAL *create_credentials(char *user,char* pwd) {
	CREDENTIAL *credential = malloc(SIZEOF_CREDENTIAL);
	credential->user = user;
	credential->pwd = pwd;
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Create credentials %p (%p,%p)\n",credential,user,pwd);
#endif
	return credential;
}
void release_credentials(CREDENTIAL* credential) {
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Free credentials %p\n",credential);
#endif
	free(credential->user);
	free(credential->pwd);
	free(credential);
}
// Initialize ABD array with number of ABD
PABD *create_abd(int num_abd)
{
	int i;
	PABD *pabd = (PABD *)malloc(num_abd * sizeof(PABD *));
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Alloc ABD %p\n",pabd);
#endif
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
#if defined(ADA_MEM_TRACE)
				fprintf(stdout,"Free %i. ABD buffer %p\n",i,pabd[i]->abdaddr);
#endif
				free(pabd[i]->abdaddr);
			} else {
#if defined(ADA_MEM_TRACE)
				fprintf(stdout,"Free %i. ABD buffer NULL!!!!\n",i);
#endif
			}
#if defined(ADA_MEM_TRACE)
			fprintf(stdout,"Free %i. ABD %p\n",i,pabd[i]);
#endif
			free(pabd[i]);
		} else {
#if defined(ADA_MEM_TRACE)
			fprintf(stdout,"Free %i. ABD NULL!!!!\n",i);
#endif
		}
	}
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Free ABD %p\n",pabd);
#endif
	free(pabd);
}

// Adabas interface for Go to call ACBX Adabas calls
int go_eadabasx(ADAID_T *adabas_id, PACBX acbx, int num_abd, PABD *abd, CREDENTIAL *c)
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
		if ((c!=NULL)&&(c->user!=NULL)) {
#if defined(ADA_MEM_TRACE)
			fprintf(stdout,"%c%c %p User: %s PWD: %s\n",acbx->acbxcmd[0],acbx->acbxcmd[1],
			   c,c->user,c->pwd);
#endif
#if 0
			fprintf(stdout,"user %p %s\n",c->user,c->user);
			fprintf(stdout,"pwd  %p\n",c->pwd);
#endif
			lnk_set_uid_pw(acbx->acbxdbid, c->user, c->pwd);
		}
		rsp = adabasx(acbx, num_abd, abd);
#if defined(ADA_MEM_TRACE)
		fprintf(stdout,"%c%c rsp: %d rsp: %d ID:%s/%d/%lu\n",acbx->acbxcmd[0],acbx->acbxcmd[1],
		       rsp,acbx->acbxrsp,adabas_id->s_user,adabas_id->s_pid,adabas_id->s_timestamp);
#endif
	}
	return (rsp);
}

// Malloc C based memory and copy Go based ABD to C based because of pointer references
void copy_to_abd(PABD *pabd, int index, PABD x, char *data, uint32_t size)
{
	PABD dest_pabd = pabd[index] = malloc(L_ABD);
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Alloc %d. ABD %p\n",index,dest_pabd);
#endif
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
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Alloc %d. ABD buffer %p\n",index,dest_pabd->abdaddr);
#endif
		memcpy(dest_pabd->abdaddr, data, size);
		dest_pabd->abdsize = size;
	}
	else
	{
		dest_pabd->abdsize = 0;
		dest_pabd->abdaddr = NULL;
#if defined(ADA_MEM_TRACE)
	fprintf(stdout,"Alloc %d. ABD buffer empty\n",index);
#endif
	}
}

// Copy C based ABD to Go based because of pointer references and free memory
void copy_from_abd(PABD *pabd, int index, PABD x, char *data, uint32_t size)
{
	PABD dest_pabd = pabd[index];
	memcpy(x, dest_pabd, L_ABD);
	if ((data != NULL) && (dest_pabd->abdrecv > 0))
	{
#if defined(ADA_MEM_TRACE)
		fprintf(stdout,"Recopy %d.ABD %d(%d)\n",size,dest_pabd->abdrecv);
#endif
		memcpy(data, dest_pabd->abdaddr, size);
	}
	if (dest_pabd->abdaddr != NULL)
	{
#if defined(ADA_MEM_TRACE)
		fprintf(stdout,"Free %d ABD buffer %p\n",index,dest_pabd->abdaddr);
#endif
		free(dest_pabd->abdaddr);
		dest_pabd->abdaddr = NULL;
	}
#if defined(ADA_MEM_TRACE)
		fprintf(stdout,"Free %d ABD %p\n",index,pabd[index]);
#endif
	free(pabd[index]);
	pabd[index] = NULL;
}
*/
import "C"

type AdaIPC struct {
}

func NewAdaIPC(URL *URL, ID *ID) *AdaIPC {
	return &AdaIPC{}
}

// Send Send the TCP/IP request to remote Adabas database
func (ipc *AdaIPC) Send(adabas *Adabas) (err error) {

	adatypes.Central.Log.Debugf("Call Adabas using native link: %v", adatypes.Central.IsDebugLevel())
	pabdArray := C.create_abd(C.int(len(adabas.AdabasBuffers)))
	for index := range adabas.AdabasBuffers {
		adabas.AdabasBuffers[index].abd.Abdrecv = adabas.AdabasBuffers[index].abd.Abdsize
		adabas.AdabasBuffers[index].createCAbd(pabdArray, index)
	}
	x := (*C.CREDENTIAL)(unsafe.Pointer(C.NULL))
	//CreateCredentials()
	// &C.CREDENTIAL{user: nil, pwd: nil}
	/* For OP calls, initialize the security layer setting the password. The corresponding
	 * Security buffer (Z-Buffer) are generated inside the Adabas client layer.
	 * Under the hood the Z-Buffer will generate one time passwords send with the next call
	 * after OP. */
	if adabas.ID.pwd != "" && adabas.Acbx.Acbxcmd == op.code() {
		adatypes.Central.Log.Debugf("Set user %s password credentials", adabas.ID.user)
		x = C.create_credentials(C.CString(adabas.ID.user), C.CString(adabas.ID.pwd))
		/*cUser := C.CString(adabas.ID.user)
		cPassword := C.CString(adabas.ID.pwd)
		x.user = cUser
		x.pwd = cPassword*/
	}
	ret := int(C.go_eadabasx((*C.ADAID_T)(unsafe.Pointer(adabas.ID.AdaID)),
		(*C.ACBX)(unsafe.Pointer(adabas.Acbx)), C.int(len(adabas.AdabasBuffers)), pabdArray, x))
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Received calling CC %c%c adabasp=%p URL=%s Adabas ID=%v",
			adabas.Acbx.Acbxcmd[0], adabas.Acbx.Acbxcmd[1],
			adabas, adabas.URL.String(), adabas.ID.String())
		adatypes.Central.Log.Debugf("Local Adabas call returns: %d", ret)
		adatypes.LogMultiLineString(true, adabas.Acbx.String())
	}

	// Free the corresponding C based memory
	if adabas.ID.pwd != "" && adabas.Acbx.Acbxcmd == op.code() {
		C.release_credentials(x)
		/*C.free(unsafe.Pointer(x.user))
		C.free(unsafe.Pointer(x.pwd))*/
	}
	for index := range adabas.AdabasBuffers {
		//	adatypes.Central.Log.Debugf(index, ".ABD out : ", adabas.AdabasBuffers[index].abd.Abdsize)
		adabas.AdabasBuffers[index].putCAbd(pabdArray, index)
		if adatypes.Central.IsDebugLevel() {
			adatypes.LogMultiLineString(true, adabas.AdabasBuffers[index].String())
		}
	}
	adatypes.Central.Log.Debugf("Destroy temporary ABD")
	C.destroy_abd(pabdArray, C.int(len(adabas.AdabasBuffers)))
	return nil
}

// createCAbd create C native ABD
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
		adatypes.LogMultiLineString(true, adatypes.FormatByteBuffer("Buffer content:", adabasBuffer.buffer))
	}

}

// putCAbd put ABD array to C native ABD
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
		adatypes.Central.Log.Debugf("C ABD %c: send=%d recv=%d", cabd.Abdid, cabd.Abdsend, cabd.Abdrecv)
		adatypes.LogMultiLineString(true, adatypes.FormatByteBuffer("Buffer content:", adabasBuffer.buffer))
	}

}

func (ipc *AdaIPC) Connect(adabas *Adabas) (err error) {
	return nil
}

// Disconnect disconnect remote TCP/IP Adabas nucleus
func (ipc *AdaIPC) Disconnect() (err error) {
	return nil
}
