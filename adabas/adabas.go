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

// Package adabas contains Adabas specific Adabas buffer conversion and call functions.
// The Adabas file metadata will be read and requested field content is returned.
// The package provides three type of access to the database.
//
//  1. The local access using the Adabas client native library. This uses the classic
//     inter process communication method
//  2. The Entire Network remote data access using the Entire Network server and corresponding
//     infrastructure
//  3. The new Adabas TCP/IP communication for a direct point-to-point access to the database
//
package adabas

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
)

// MaxDatabasesID maximum valid database id
const MaxDatabasesID = 255

const adaEmptOpt = ' '
const adaFdtXOpt = 'X'

const (
	// AdaNormal Adabas success response code
	AdaNormal = 0
	// AdaEOF Adabas End of File reached (End of data received)
	AdaEOF = 3
	// AdaRbts Adabas record buffer too short
	AdaRbts = 53
	// AdaAnact Adabas offline
	AdaAnact = 148
	// AdaSysCe Adabas remote connection problem
	AdaSysCe = 149
)

type adabasOption uint32

// Options for operation and synchronization of different requests
const (
	adabasOptionOP adabasOption = iota
	adabasOptionXX
)

// Transaction flags to synchronize and manage different requests
type transactions struct {
	flags            uint32
	openTransactions uint32
	connection       interface{}
}

// Adabas is an main Adabas structure containing all call specific parameters
type Adabas struct {
	URL           *URL
	ID            *ID
	status        *Status
	Acbx          *Acbx
	AdabasBuffers []*Buffer
	transactions  *transactions
}

func (option adabasOption) Bit() uint32 {
	return (1 << option)
}

// NewClonedAdabas create a cloned Adabas struct instance
func NewClonedAdabas(clone *Adabas) *Adabas {
	acbx := newAcbx(clone.Acbx.Acbxdbid)

	return &Adabas{
		ID:           clone.ID,
		status:       clone.ID.status(clone.URL.String()),
		Acbx:         acbx,
		URL:          clone.URL,
		transactions: clone.transactions,
	}
}

// NewAdabas create a new Adabas struct instance
func NewAdabas(dbid Dbid) (*Adabas, error) {
	ID := NewAdabasID()
	adatypes.Central.Log.Debugf("Implicit created Adabas instance dbid with ID: %s", ID.String())
	if (dbid < 1) || (dbid > MaxDatabasesID) {
		err := adatypes.NewGenericError(67, dbid, 1, MaxDatabasesID)
		return nil, err
	}
	acbx := newAcbx(dbid)
	URL := newURLWithDbid(dbid)
	return &Adabas{
		ID:           ID,
		status:       ID.status(URL.String()),
		URL:          URL,
		Acbx:         acbx,
		transactions: &transactions{},
	}, nil
}

// NewAdabass create a new Adabas struct instance using string parameter
// func NewAdabass(target string) (*Adabas, error) {
// 	ID := NewAdabasID()
// 	adatypes.Central.Log.Debugf("Implicit created Adabas instance target with ID %s", ID.String())
// 	URL, err := newURL(target)
// 	if err != nil {
// 		return nil, err
// 	}
// 	acbx := newAcbx(URL.Dbid)
// 	adatypes.Central.Log.Debugf("Created ACBX")
// 	return &Adabas{
// 		ID:           ID,
// 		status:       ID.status(URL.String()),
// 		URL:          URL,
// 		Acbx:         acbx,
// 		transactions: &transactions{},
// 	}, nil
// }

// NewAdabasWithID create a new Adabas struct instance using string parameter
func NewAdabasWithID(target string, ID *ID) (*Adabas, error) {
	if ID == nil {
		return nil, adatypes.NewGenericError(60)
	}
	adatypes.Central.Log.Debugf("Use new Adabas with Adabas ID: %s", ID.String())
	// fmt.Println("Create URL", target)
	URL, err := newURL(target)
	if err != nil {
		return nil, err
	}
	if (URL.Dbid < 1) || (URL.Dbid > MaxDatabasesID) {
		err = adatypes.NewGenericError(67, URL.Dbid, 1, MaxDatabasesID)
		return nil, err
	}

	acbx := newAcbx(URL.Dbid)
	return &Adabas{
		ID:           ID,
		status:       ID.status(URL.String()),
		URL:          URL,
		Acbx:         acbx,
		transactions: &transactions{},
	}, nil
}

// NewAdabasWithURL create a new Adabas struct instance
func NewAdabasWithURL(URL *URL, ID *ID) (*Adabas, error) {
	adatypes.Central.Log.Debugf("Use new Adabas instance with Adabas ID: %s", ID.String())
	if (URL.Dbid < 1) || (URL.Dbid > MaxDatabasesID) {
		err := adatypes.NewGenericError(67, URL.Dbid, 1, MaxDatabasesID)
		return nil, err
	}
	acbx := newAcbx(URL.Dbid)
	return &Adabas{
		URL:          URL,
		ID:           ID,
		status:       ID.status(URL.String()),
		Acbx:         acbx,
		transactions: &transactions{},
	}, nil
}

// Open opens a session to the database
func (adabas *Adabas) Open() (err error) {
	adatypes.Central.Log.Debugf("Open flag %p %v preopen", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	url := adabas.URL.String()
	if adabas.ID.isOpen(url) {
		adatypes.Central.Log.Debugf("Database %s already opened by ID %#v", url, adabas.ID)
		return
	}
	if adabas.transactions.flags&adabasOptionOP.Bit() != 0 {
		adatypes.Central.Log.Debugf("Database already opened %#v", adabas.ID)
		return
	}
	adatypes.Central.Log.Debugf("Open database %d %s", adabas.Acbx.Acbxdbid, adabas.ID.String())
	adabas.AdabasBuffers = append(adabas.AdabasBuffers, NewBuffer(AbdAQFb))
	adabas.AdabasBuffers = append(adabas.AdabasBuffers, NewBuffer(AbdAQRb))

	adabas.Acbx.Acbxcmd = op.code()

	adabas.AdabasBuffers[0].WriteString(" ")
	adabas.AdabasBuffers[1].WriteString("UPD.")
	adabas.AdabasBuffers[1].abd.Abdsend = adabas.AdabasBuffers[1].abd.Abdsize

	err = adabas.CallAdabas()
	if err != nil {
		adatypes.Central.Log.Debugf("Open call response ret=%v", err)
		return
	}
	if adabas.Acbx.Acbxrsp == AdaNormal {
		adatypes.Central.Log.Debugf("Open call response success")
		adabas.transactions.flags |= adabasOptionOP.Bit()
		adatypes.Central.Log.Debugf("Open flag %p %v normal", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
		adabas.status.open = true
		adabas.status.platform = adatypes.NewPlatformIsl(adabas.Acbx.Acbxisl)
	} else {
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("Error calling open", err)
		adabas.status.open = false
	}
	return err
}

// Close A session to the database will be closed
func (adabas *Adabas) Close() {
	adatypes.Central.Log.Debugf("Open flag %p %v preclose", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	if adabas.transactions.openTransactions > 0 {
		adabas.BackoutTransaction()
	}
	adabas.AdabasBuffers = nil
	adabas.Acbx.Acbxcmd = cl.code()
	ret := adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Close call response ret=%v %s", ret, adabas.ID.String())
	adatypes.Central.Log.Debugf("Open flag %p %v close", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	//if adabas.transactions.flags&adabasOptionOP.Bit() != 0 {
	adabas.transactions.flags &^= adabasOptionOP.Bit()
	//}
	adabas.status.open = false

	adatypes.Central.Log.Debugf("Open flag %p %v afterclose", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	adabas.transactions.openTransactions = 0
}

// Release Any resource in the database of the session are released
func (adabas *Adabas) Release() (err error) {
	adabas.AdabasBuffers = nil
	adabas.Acbx.Acbxcmd = rc.code()
	adabas.Acbx.resetCop()
	err = adabas.CallAdabas()
	return
}

func (adabas *Adabas) String() string {
	return fmt.Sprintf("Adabas url=%s fnr=%d", adabas.URL.String(), adabas.Acbx.Acbxfnr)
}

// ACBX Current used ACBX
func (adabas *Adabas) ACBX() *Acbx {
	return adabas.Acbx
}

// SetAbd Set ABD to adabas structure
func (adabas *Adabas) SetAbd(abd []*Buffer) {
	adabas.AdabasBuffers = abd
}

// IsRemote Indicate if the call uses WCL remote calls
func (adabas *Adabas) IsRemote() bool {
	adatypes.Central.Log.Debugf("Remote usage check of %s", adabas.URL)
	return adabas.URL != nil && adabas.URL.Port > 0
}

func (adabas *Adabas) callRemoteAdabas() (err error) {
	// Call remote database URL
	adatypes.Central.Log.Debugf("Call remote via driver url: %s", adabas.URL)
	switch adabas.URL.Driver {
	case "tcpip":
		return fmt.Errorf("Entire Network client not supported, use port 0 and Entire Network native access")
	case "adatcp":
		return adabas.sendTCP()
	case "":
		return adatypes.NewGenericError(49)
	}
	return adatypes.NewGenericError(1, adabas.URL.Driver)
}

// sendTCP Send the TCP/IP request to remote Adabas database
func (adabas *Adabas) sendTCP() (err error) {
	var tcpConn *adatcp
	// Check if connection is already available
	if adabas.transactions.connection == nil {
		adatypes.Central.Log.Debugf("Establish new context for %p", adabas)

		tcpConn, err = connect(fmt.Sprintf("%s:%d", adabas.URL.Host, adabas.URL.Port), Endian(), adabas.ID.AdaID.User,
			adabas.ID.AdaID.Node, adabas.ID.AdaID.Pid, adabas.ID.AdaID.Timestamp)
		if err != nil {
			adabas.Acbx.Acbxrsp = AdaSysCe
			adatypes.Central.Log.Debugf("Establish TCP context error ", err)
			err = NewError(adabas)
			return
		}
		adabas.transactions.connection = tcpConn
	} else {
		adatypes.Central.Log.Debugf("Use context for %p %p ", adabas, adabas.transactions.connection)
		tcpConn = adabas.transactions.connection.(*adatcp)
	}
	var buffer bytes.Buffer
	err = adabas.WriteBuffer(&buffer, Endian(), false)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Send buffer of length=%d lenBuffer=%d", buffer.Len(), len(adabas.AdabasBuffers))
	err = tcpConn.SendData(buffer, uint32(len(adabas.AdabasBuffers)))
	if err != nil {
		adatypes.Central.Log.Debugf("Transmit Adabas call error ", err)
		tcpConn.Disconnect()
		adabas.transactions.connection = nil
		return
	}
	buffer.Reset()
	var nrAbdBuffers uint32
	nrAbdBuffers, err = tcpConn.ReceiveData(&buffer)
	if err != nil {
		adatypes.Central.Log.Debugf("Transmit Adabas call error ", err)
		return
	}
	err = adabas.ReadBuffer(&buffer, Endian(), nrAbdBuffers, false)
	if err != nil {
		adatypes.Central.Log.Debugf("Read buffer error, destroy context ... %v", err)
		tcpConn.Disconnect()
		return
	}

	adatypes.Central.Log.Debugf("Remote Adabas call returns successfully")
	if adabas.Acbx.Acbxcmd == cl.code() {
		adatypes.Central.Log.Debugf("Close called, destroy context ...")
		if tcpConn != nil {
			tcpConn.Disconnect()
			adabas.transactions.connection = nil
		}
	}
	adatypes.Central.Log.Debugf("Got context for %p %p ", adabas, adabas.transactions.connection)
	return
}

// ReadFileDefinition Read file definition out of Adabas file
func (adabas *Adabas) ReadFileDefinition(fileNr Fnr) (definition *adatypes.Definition, err error) {
	cacheName := adabas.URL.String() + "_" + strconv.Itoa(int(fileNr))
	definition = adatypes.CreateDefinitionByCache(cacheName)
	if definition != nil {
		return
	}

	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read file definition with %v", lf.code())
	adabas.Acbx.Acbxcmd = lf.code()
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[0] = adaEmptOpt
	adabas.Acbx.Acbxcop[1] = adaFdtXOpt
	adabas.Acbx.Acbxisn = 1
	adabas.Acbx.Acbxisq = 0

	adabas.AdabasBuffers = make([]*Buffer, 2)
	adabas.AdabasBuffers[0] = NewBuffer(AbdAQFb)
	adabas.AdabasBuffers[1] = NewBufferWithSize(AbdAQRb, 4096*2)

	adabas.AdabasBuffers[0].WriteString(".")

	adabas.Acbx.Acbxfnr = fileNr
	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Read file definition %v rsp=%d", err, adabas.Acbx.Acbxrsp)
	if err == nil {
		/* Create new helper to parse returned buffer */
		helper := adatypes.NewHelper(adabas.AdabasBuffers[1].buffer, int(adabas.AdabasBuffers[1].abd.Abdrecv), Endian())
		fdtDefinition := createFdtDefintion()
		fdtDefinition.Values = nil
		fdtDefinition.ParseBuffer(helper, adatypes.NewBufferOption(false, false))
		adatypes.Central.Log.Debugf("Format read field definition")
		definition, err = createFieldDefinitionTable(fdtDefinition)
		if err != nil {
			adatypes.Central.Log.Debugf("ERROR create FDT:", err)
			return
		}
		definition.PutCache(cacheName)
		// definition.DumpTypes(true, true)
		adatypes.Central.Log.Debugf("Ready parse Format read field definition")
	}
	// Check response to indicate error reading field definition
	if adabas.Acbx.Acbxrsp != 0 {
		log.Info("Error reading definition: ", adabas.getAdabasMessage())
		adatypes.LogMultiLineString(adabas.Acbx.String())
		err = NewError(adabas)
	}

	return
}

// Prepare Adabas buffer ABD and buffer content for the Adabas request
func (adabas *Adabas) prepareBuffers(adabasRequest *adatypes.Request) {
	bufferCount := 2
	if adabasRequest.SearchTree != nil {
		bufferCount = 4
	}
	multifetch := adabasRequest.Multifetch
	if multifetch > 1 {
		bufferCount++
	} else {
		multifetch = 1
	}

	adabas.AdabasBuffers = make([]*Buffer, bufferCount)
	// Create format buffer for the call
	adabas.AdabasBuffers[0] = NewBuffer(AbdAQFb)
	adabas.AdabasBuffers[0].buffer = adabasRequest.FormatBuffer.Bytes()
	adabas.AdabasBuffers[0].abd.Abdsize = uint64(adabasRequest.FormatBuffer.Len())
	adabas.AdabasBuffers[0].abd.Abdsend = adabas.AdabasBuffers[0].abd.Abdsize
	if adabas.AdabasBuffers[0].abd.Abdver[0] != 'G' {
		adatypes.Central.Log.Infof("ABD init 0 error %p\n", adabas.AdabasBuffers[0])
		os.Exit(100)
	}
	adatypes.Central.Log.Debugf("ABD init 0 %p\n", adabas.AdabasBuffers[0])

	// Create record buffer for the call
	adabas.AdabasBuffers[1] = NewBufferWithSize(AbdAQRb,
		multifetch*(adabasRequest.RecordBufferLength+adabasRequest.RecordBufferShift))
	adatypes.Central.Log.Debugf("ABD init 1 %p\n", adabas.AdabasBuffers[1])

	// Define search and value buffer to search
	if adabasRequest.SearchTree != nil {
		adatypes.Central.Log.Debugf("Search logical added")
		adabas.AdabasBuffers[2] = SearchAdabasBuffer(adabasRequest.SearchTree)
		adabas.AdabasBuffers[3] = ValueAdabasBuffer(adabasRequest.SearchTree)
	}
	if adabasRequest.Multifetch > 1 {
		adatypes.Central.Log.Debugf("Create multifetch buffer for %d multifetch entries", adabasRequest.Multifetch)
		index := len(adabas.AdabasBuffers) - 1
		adabas.AdabasBuffers[index] = NewBufferWithSize(AbdAQMb, 4+(adabasRequest.Multifetch*16))
	}

}

// ReadPhysical read data in physical order
func (adabas *Adabas) ReadPhysical(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Open flag %p %v readpp", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Open flag %p %v readp", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	adatypes.Central.Log.Debugf("Physical read file ... %s", l2.command())
	if adabasRequest.Option.HoldRecords {
		adabas.Acbx.Acbxcmd = l5.code()
	} else {
		adabas.Acbx.Acbxcmd = l2.code()
	}
	nrMultifetch := 2
	adabas.Acbx.resetCop()
	if adabasRequest.Multifetch > 1 {
		adabas.Acbx.Acbxcop[0] = 'M'
		nrMultifetch = 3
	}
	adabas.Acbx.Acbxisn = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	multifetch := adabasRequest.Multifetch
	if multifetch < 1 {
		multifetch = 1
	}
	adabas.AdabasBuffers = make([]*Buffer, nrMultifetch)
	adabas.AdabasBuffers[0] = NewBuffer(AbdAQFb)
	adabas.AdabasBuffers[0].buffer = adabasRequest.FormatBuffer.Bytes()
	adabas.AdabasBuffers[0].abd.Abdsize = uint64(adabasRequest.FormatBuffer.Len())
	adabas.AdabasBuffers[0].abd.Abdsend = adabas.AdabasBuffers[0].abd.Abdsize
	adabas.AdabasBuffers[1] = NewBuffer(AbdAQRb)
	adabas.AdabasBuffers[1].Allocate(multifetch * adabasRequest.RecordBufferLength)
	if multifetch > 1 {
		adabas.AdabasBuffers[2] = NewBuffer(AbdAQMb)
		adabas.AdabasBuffers[2].Allocate(multifetch * 32)
	}

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	adatypes.Central.Log.Debugf("Open flag %p %v readpf", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	return
}

// read a specific ISN out of Adabas file
func (adabas *Adabas) readISN(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Open flag %p %v readisnp", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read ISN %d ... %s dbid=%d fnr=%d", adabasRequest.Isn, l1.command(), adabas.Acbx.Acbxdbid, fileNr)
	if adabasRequest.Option.HoldRecords {
		adabas.Acbx.Acbxcmd = l4.code()
	} else {
		adabas.Acbx.Acbxcmd = l1.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxisn = adabasRequest.Isn
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}
	adabas.Acbx.Acbxfnr = fileNr

	adabas.prepareBuffers(adabasRequest)

	err = adabas.loopCall(adabasRequest, x)
	return
}

// ReadLogicalWith Read logical using a descriptor
func (adabas *Adabas) ReadLogicalWith(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read logical ... %s dbid=%d multifetch=%d", l3.command(), adabas.Acbx.Acbxdbid, adabasRequest.Multifetch)
	if adabasRequest.Option.HoldRecords {
		adabas.Acbx.Acbxcmd = l6.code()
	} else {
		adabas.Acbx.Acbxcmd = l3.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = 'A'
	if adabasRequest.Multifetch > 1 {
		adabas.Acbx.Acbxcop[0] = 'M'
	}

	adabas.Acbx.Acbxisn = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabas.prepareBuffers(adabasRequest)
	var add1 bytes.Buffer
	if len(adabasRequest.Descriptors) == 1 {
		for _, d := range adabasRequest.Descriptors {
			add1.WriteString(d)
		}
	} else {
		err = adatypes.NewGenericError(58)
		return
	}
	add1.WriteString("        ")
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:7])

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// SearchLogicalWith Search logical using a descriptor
func (adabas *Adabas) SearchLogicalWith(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Read logical ... %s dbid=%d", l3.command(), adabas.Acbx.Acbxdbid)
	adabas.Acbx.Acbxcmd = s2.code()
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[0] = 'H'
	if adabasRequest.Option.Ascending {
		adabas.Acbx.Acbxcop[1] = ' '
	} else {
		adabas.Acbx.Acbxcop[1] = 'D'
	}

	adabas.Acbx.Acbxisn = 0
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabas.prepareBuffers(adabasRequest)
	var add1 bytes.Buffer
	if len(adabasRequest.Descriptors) > 0 {
		for _, d := range adabasRequest.Descriptors {
			add1.WriteString(d)
		}
	} else {
		add1.WriteString("ISN")
	}
	add1.WriteString("        ")
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:7])

	adabas.Acbx.Acbxfnr = fileNr
	// Call Adabas
	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Received search response ret=%v", err)
	if err != nil {
		return
	}
	// End of file reached
	if adabas.Acbx.Acbxrsp == AdaEOF || adabas.Acbx.Acbxisq == 0 {
		return
	}

	if adabasRequest.Option.HoldRecords {
		adabas.Acbx.Acbxcmd = l1.code()
	} else {
		adabas.Acbx.Acbxcmd = l4.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = ' '
	adabas.Acbx.Acbxcop[1] = 'N'
	err = adabas.loopCall(adabasRequest, x)
	return
}

// Loop call used to read a sequence of records
func (adabas *Adabas) loopCall(adabasRequest *adatypes.Request, x interface{}) (err error) {

	count := uint64(0)
	var responseCode uint32
	for responseCode == 0 {
		if !adabasRequest.Option.SecondCall {
			//adabasRequest.Definition.Values = nil
			adabasRequest.Definition.CreateValues(false)
		}
		// Call Adabas
		err = adabas.CallAdabas()
		adatypes.Central.Log.Debugf("Received call response ret=%v", err)
		if err != nil {
			return
		}

		adabasRequest.Response = adabas.Acbx.Acbxrsp

		// End of file reached
		if adabas.Acbx.Acbxrsp == AdaEOF {
			return
		}
		// Error received from Adabas
		if adabas.Acbx.Acbxrsp != AdaNormal {
			log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
			err = NewError(adabas)
			return
		}
		adabasRequest.RecordBuffer = adatypes.NewHelper(adabas.AdabasBuffers[1].buffer,
			int(adabas.AdabasBuffers[1].abd.Abdrecv), Endian())
		adabasRequest.Isn = adatypes.Isn(adabas.Acbx.Acbxisn)
		adabasRequest.IsnQuantity = adabas.Acbx.Acbxisq

		// If parser is available, use the parser to extract content
		if adabasRequest.Parser != nil {
			var multifetchHelper *adatypes.BufferHelper
			nrMultifetchEntries := uint32(1)
			if adabasRequest.Multifetch > 1 {
				multifetchHelper, err = adabas.multifetchBuffer()
				if err != nil {
					return
				}
				nrMultifetchEntries, err = multifetchHelper.ReceiveUInt32()
				if err != nil {
					return
				}
				if nrMultifetchEntries > 10000 {
					panic("To many multifetch entries")
				}
				adatypes.Central.Log.Debugf("Nr of multifetch entries %d", nrMultifetchEntries)
			}
			for nrMultifetchEntries > 0 {
				count++
				adatypes.Central.Log.Debugf("Nr of multifetch entries left: %d", nrMultifetchEntries)
				if multifetchHelper != nil {
					recordLength, rErr := multifetchHelper.ReceiveUInt32()
					if rErr != nil {
						err = rErr
						return
					}
					adatypes.Central.Log.Debugf("Record length %d", recordLength)
					responseCode, err = multifetchHelper.ReceiveUInt32()
					if err != nil {
						return
					}
					if responseCode != AdaNormal {
						adabasRequest.Response = adabas.Acbx.Acbxrsp
						break
					}
					adatypes.Central.Log.Debugf("Response code %d", responseCode)
					isn, isnErr := multifetchHelper.ReceiveUInt32()
					if isnErr != nil {
						err = isnErr
						return
					}
					adatypes.Central.Log.Debugf("ISN %d", isn)
					adabasRequest.Isn = adatypes.Isn(isn)
					adabas.Acbx.Acbxisn = adatypes.Isn(isn)
					_, err = multifetchHelper.ReceiveUInt32()
					if err != nil {
						return
					}
				}

				adatypes.Central.Log.Debugf("Parse Buffer .... values avail.=%v", (adabasRequest.Definition.Values == nil))
				_, err = adabasRequest.Definition.ParseBuffer(adabasRequest.RecordBuffer, adabasRequest.Option)
				if err != nil {
					return
				}
				err = adabas.secondCall(adabasRequest, x)
				if err != nil {
					return
				}
				adatypes.Central.Log.Debugf("Found parser .... values avail.=%v", (adabasRequest.Definition.Values == nil))
				err = adabasRequest.Parser(adabasRequest, x)
				if err != nil {
					return
				}
				nrMultifetchEntries--

				// If multifetch on, create values for next parse step, only possible on read calls
				if nrMultifetchEntries > 0 {
					//adabasRequest.Definition.Values = nil
					adabasRequest.Definition.CreateValues(false)
				}
			}

		} else {
			adatypes.Central.Log.Debugf("Found no parser")
			break
		}

		adatypes.Central.Log.Debugf("Limit=%d count=%d", adabasRequest.Limit, count)
		if (adabasRequest.Limit > 0) && (count >= adabasRequest.Limit) {
			break
		}
	}

	return
}

func (adabas *Adabas) secondCall(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Check second call .... values avail.=%v", (adabasRequest.Definition.Values == nil))
	if adabasRequest.Option.NeedSecondCall {
		adatypes.Central.Log.Debugf("Need second call %v", adabasRequest.Option.NeedSecondCall)
		tmpAdabasRequest, err2 := adabasRequest.Definition.CreateAdabasRequest(false, true)
		if err2 != nil {
			err = err2
			return
		}
		acbx := *adabas.Acbx
		abd := adabas.AdabasBuffers
		tmpAdabasRequest.Isn = adabasRequest.Isn
		tmpAdabasRequest.Definition = adabasRequest.Definition
		tmpAdabasRequest.Option.SecondCall = true
		adatypes.Central.Log.Debugf("Got temporary request")
		err = adabas.readISN(adabas.Acbx.Acbxfnr, tmpAdabasRequest, x)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Parse buffer of temporary request")
		_, err = tmpAdabasRequest.Definition.ParseBuffer(tmpAdabasRequest.RecordBuffer, tmpAdabasRequest.Option)
		if err != nil {
			adatypes.Central.Log.Debugf("Parse buffer of temporary request ended with error: ", err)
			return
		}
		adatypes.Central.Log.Debugf("Parse buffer of temporary request ended, reset to old adabas request")
		*adabas.Acbx = acbx
		adabas.AdabasBuffers = abd
		adatypes.Central.Log.Debugf("Second call done")

		adabasRequest.Option.NeedSecondCall = false
	}

	return
}

// Histogram histogram of a specific descriptor
func (adabas *Adabas) Histogram(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Descriptor read file %s", l9.command())
	adabas.Acbx.Acbxcmd = l9.code()
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = 'A'
	adabas.Acbx.Acbxisn = 0
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabas.prepareBuffers(adabasRequest)

	var add1 bytes.Buffer
	for _, d := range adabasRequest.Descriptors {
		add1.WriteString(d)
	}
	add1.WriteString("        ")
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:7])

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// Store store a record into database
func (adabas *Adabas) Store(fileNr Fnr, adabasRequest *adatypes.Request) (err error) {
	adatypes.Central.Log.Debugf("Prepare Store transactions=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Store transactions=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	if adabasRequest.Isn != 0 {
		adatypes.Central.Log.Debugf("Store data ... %s", n2.command())
		adabas.Acbx.Acbxcmd = n2.code()
		adabas.Acbx.Acbxisn = adabasRequest.Isn
	} else {
		adatypes.Central.Log.Debugf("Store data ... %s", n1.command())
		adabas.Acbx.Acbxcmd = n1.code()
		adabas.Acbx.Acbxisn = 0
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0, 0, 0, 0}
	adabas.Acbx.Acbxfnr = fileNr

	adabas.AdabasBuffers = make([]*Buffer, 2)
	adabas.AdabasBuffers[0] = NewBuffer(AbdAQFb)
	adabas.AdabasBuffers[0].buffer = adabasRequest.FormatBuffer.Bytes()
	adabas.AdabasBuffers[0].abd.Abdsize = uint64(adabasRequest.FormatBuffer.Len())
	adabas.AdabasBuffers[0].abd.Abdsend = adabas.AdabasBuffers[0].abd.Abdsize
	adabas.AdabasBuffers[1] = NewBuffer(AbdAQRb)
	adabas.AdabasBuffers[1].buffer = adabasRequest.RecordBuffer.Buffer()
	adabas.AdabasBuffers[1].abd.Abdsize = uint64(len(adabas.AdabasBuffers[1].buffer))
	adabas.AdabasBuffers[1].abd.Abdsend = adabas.AdabasBuffers[1].abd.Abdsize

	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Store call response ret=%v", err)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Store ISN rsp=%d ... %d", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxisn)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		log.Errorf("Error storing data: %s", adabas.getAdabasMessage())
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("%v", err)
		return
	}
	adabas.transactions.openTransactions++
	adabasRequest.Isn = adabas.Acbx.Acbxisn
	return
}

// Update update a record in database
func (adabas *Adabas) Update(fileNr Fnr, adabasRequest *adatypes.Request) (err error) {
	adatypes.Central.Log.Debugf("Prepare Update transactions=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	err = adabas.Open()
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Update transactions=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	adatypes.Central.Log.Debugf("Update data ... %s", a1.command())
	adabas.Acbx.Acbxcmd = a1.code()
	adabas.Acbx.Acbxisn = adabasRequest.Isn
	adabas.Acbx.resetCop()
	if adabasRequest.Option.ExchangeRecord {
		adabas.Acbx.Acbxcop[0] = 'X'
		adabas.Acbx.Acbxcop[1] = 'H'
	} else {
		adabas.Acbx.Acbxcop[0] = 'H'
	}
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0, 0, 0, 0}
	adabas.Acbx.Acbxfnr = fileNr

	adabas.AdabasBuffers = make([]*Buffer, 2)
	adabas.AdabasBuffers[0] = NewBuffer(AbdAQFb)
	adabas.AdabasBuffers[0].buffer = adabasRequest.FormatBuffer.Bytes()
	adabas.AdabasBuffers[0].abd.Abdsize = uint64(adabasRequest.FormatBuffer.Len())
	adabas.AdabasBuffers[0].abd.Abdsend = adabas.AdabasBuffers[0].abd.Abdsize
	adabas.AdabasBuffers[1] = NewBuffer(AbdAQRb)
	adabas.AdabasBuffers[1].buffer = adabasRequest.RecordBuffer.Buffer()
	adabas.AdabasBuffers[1].abd.Abdsize = uint64(len(adabas.AdabasBuffers[1].buffer))
	adabas.AdabasBuffers[1].abd.Abdsend = adabas.AdabasBuffers[1].abd.Abdsize

	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Update call response ret=%v", err)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Update ISN rsp=%d ... %d", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxisn)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		log.Errorf("Error updating data: %s", adabas.getAdabasMessage())
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("%v", err)
		return
	}
	adabas.transactions.openTransactions++
	adabasRequest.Isn = adabas.Acbx.Acbxisn
	return
}

// SetDbid set new database id
func (adabas *Adabas) SetDbid(dbid Dbid) {
	if dbid == adabas.Acbx.Acbxdbid {
		return
	}
	adabas.Close()
	adabas.Acbx.Acbxdbid = dbid
}

// DeleteIsn delete a single isn
func (adabas *Adabas) DeleteIsn(fileNr Fnr, isn adatypes.Isn) (err error) {
	adatypes.Central.Log.Debugf("Open flag %p %v delete", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	adatypes.Central.Log.Debugf("Delete ISN transactions=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	adatypes.Central.Log.Debugf("Delete Isn ...%s on dbid %d and file %d", e1.command(), adabas.Acbx.Acbxdbid, fileNr)
	adabas.Acbx.Acbxcmd = e1.code()
	adabas.Acbx.Acbxisn = isn
	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.CallAdabas()
	if err != nil {
		adatypes.Central.Log.Debugf("Delete isn call response error=%v", err)
		return
	}
	adabas.transactions.openTransactions++
	adatypes.Central.Log.Debugf("Delete ISN error ...%d transactions=%d adabas=%p", adabas.Acbx.Acbxrsp,
		adabas.transactions.openTransactions, adabas)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
		log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// BackoutTransaction backout transaction initiated
func (adabas *Adabas) BackoutTransaction() (err error) {
	adatypes.Central.Log.Debugf("Open flag %p %v bt", adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	if adabas.transactions.flags&adabasOptionOP.Bit() == 0 || adabas.transactions.openTransactions == 0 {
		return
	}
	adatypes.Central.Log.Debugf("Backout transaction ... %s", bt.command())
	adabas.Acbx.Acbxcmd = bt.code()
	adabas.AdabasBuffers = nil

	ret := adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Backout transaction rsp ... ret=%d rsp=%d", ret, adabas.Acbx.Acbxrsp)
	adabas.transactions.openTransactions = 0

	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
		log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// EndTransaction end of transaction initiated
func (adabas *Adabas) EndTransaction() (err error) {
	adatypes.Central.Log.Debugf("End of transaction=%d adabas=%p open=%v", adabas.transactions.openTransactions,
		adabas, adabas.transactions.flags&adabasOptionOP.Bit())
	if adabas.transactions.flags&adabasOptionOP.Bit() == 0 ||
		adabas.transactions.openTransactions == 0 {
		adatypes.Central.Log.Debugf("End of transaction ... no open transactions")
		return
	}
	adatypes.Central.Log.Debugf("End of transaction ... %s", et.command())
	adabas.Acbx.Acbxcmd = et.code()
	adabas.AdabasBuffers = nil

	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("End of transction response ret=%v", err)
	if err != nil {
		return
	}
	adabas.transactions.openTransactions = 0
	adatypes.Central.Log.Debugf("End of transaction rsp ... rsp=%d", adabas.Acbx.Acbxrsp)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		log.Errorf("Error end transaction: %s", adabas.getAdabasMessage())
		log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// WriteBuffer write adabas call to buffer
func (adabas *Adabas) WriteBuffer(buffer *bytes.Buffer, order binary.ByteOrder, serverMode bool) (err error) {
	defer adatypes.TimeTrack(time.Now(), "Adabas Write buffer "+string(adabas.Acbx.Acbxcmd[:]))
	adatypes.Central.Log.Debugf("Adabas write buffer, add  ACBX: ")
	err = binary.Write(buffer, Endian(), adabas.Acbx)
	if err != nil {
		adatypes.Central.Log.Debugf("Write ACBX in buffer error %v", err)
		return
	}
	adatypes.Central.Log.Debugf("Create ADABAS ABD %d", len(adabas.AdabasBuffers))
	adatypes.Central.Log.Debugf("Buffer len= %d", buffer.Len())
	for index, abd := range adabas.AdabasBuffers {
		var tempBuffer bytes.Buffer
		if !serverMode {
			abd.abd.Abdrecv = abd.abd.Abdsize
		}
		adatypes.Central.Log.Debugf("Add %d ABD header", index)

		if abd.abd.Abdver[0] != 'G' {
			adatypes.Central.Log.Debugf("ABD error %p\n", abd)
			fmt.Println("ABD content error", index)
			os.Exit(100)
		}
		err = binary.Write(&tempBuffer, Endian(), abd.abd)
		if err != nil {
			adatypes.Central.Log.Debugf("Write ABD in buffer error %v", err)
			return
		}
		b := tempBuffer.Bytes()
		if b[2] != 'G' {
			fmt.Println("ABD buffer error")
		}
		buffer.Write(b)
		adatypes.Central.Log.Debugf("Add ADABAS ABD: %d to len buffer=%d", index, buffer.Len())
	}
	for index, abd := range adabas.AdabasBuffers {
		var transferSize uint64
		if serverMode {
			transferSize = abd.abd.Abdrecv
		} else {
			transferSize = abd.abd.Abdsend
		}
		if transferSize > 0 {
			var n int
			n, err = buffer.Write(abd.buffer)
			adatypes.Central.Log.Debugf("Add ADABAS Buffer index=%d %c -> send=%d (%d)", index, abd.abd.Abdid, transferSize, n)
			if err != nil {
				return
			}
		} else {
			adatypes.Central.Log.Debugf("Add ADABAS Buffer index=%d %c -> skipped", index, abd.abd.Abdid)
		}
	}
	return
}

// ReadBuffer read buffer and parse call
func (adabas *Adabas) ReadBuffer(buffer *bytes.Buffer, order binary.ByteOrder, nCalBuf uint32, serverMode bool) (err error) {
	defer adatypes.TimeTrack(time.Now(), "Adabas Read buffer")
	if buffer == nil {
		err = adatypes.NewGenericError(4)
		return
	}
	adatypes.Central.Log.Debugf("Read buffer, read  ACBX: ")
	err = binary.Read(buffer, binary.LittleEndian, adabas.Acbx)
	if err != nil {
		adatypes.Central.Log.Debugf("ACBX read error : %v", err)
		return
	}

	adatypes.Central.Log.Debugf("Received ACBX rsp=%d cc=%c%c\n", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxcmd[0], adabas.Acbx.Acbxcmd[1])
	adatypes.Central.Log.Debugf("Receive number of ABD: %d rsp=%d", nCalBuf, adabas.Acbx.Acbxrsp)
	if serverMode || (adabas.Acbx.Acbxrsp <= 3 && nCalBuf > 0) {
		if serverMode {
			adatypes.Central.Log.Debugf("Check nr ABDs current=%d should=%d", len(adabas.AdabasBuffers), nCalBuf)
			if nCalBuf < uint32(len(adabas.AdabasBuffers)) {
				adatypes.Central.Log.Debugf("Reduce number buffers from %d / %d", len(adabas.AdabasBuffers), nCalBuf)
				adabas.AdabasBuffers = adabas.AdabasBuffers[:nCalBuf]
			} else if nCalBuf > uint32(len(adabas.AdabasBuffers)) {
				adatypes.Central.Log.Debugf("Init number buffers to %d", nCalBuf)
				for i := uint32(len(adabas.AdabasBuffers)); i < nCalBuf; i++ {
					abd := NewBuffer(0)
					adabas.AdabasBuffers = append(adabas.AdabasBuffers, abd)
				}
			}
		}
		adatypes.Central.Log.Debugf("Parse %d ABD buffers headers Number of current ABDS=%d len=%d", nCalBuf, len(adabas.AdabasBuffers), buffer.Len())
		for index, abd := range adabas.AdabasBuffers {
			if adatypes.Central.IsDebugLevel() {
				adatypes.Central.Log.Debugf("Parse %d.ABD got %c len=%d\n", index, abd.abd.Abdid, buffer.Len())
				adatypes.LogMultiLineString(adatypes.FormatBytes("Rest ABD:", buffer.Bytes(), buffer.Len(), 8))
			}
			err = binary.Read(buffer, Endian(), &abd.abd)
			if err != nil {
				adatypes.Central.Log.Debugf("ABD read header error: %v", err)
				return
			}
			if abd.abd.Abdver[0] != 'G' {
				adatypes.Central.Log.Debugf("ABD error %p\n", abd)
				os.Exit(100)
			}
			adatypes.Central.Log.Debugf("%d.ABD got send=%d rcv=%d size=%d\n",
				index, abd.abd.Abdsend, abd.abd.Abdrecv, abd.abd.Abdsize)
			if serverMode {
				// Check if size is correct
				abd.Allocate(uint32(abd.abd.Abdsize))
			}
		}
		adatypes.Central.Log.Debugf("Parse ABD buffer data")
		for index, abd := range adabas.AdabasBuffers {
			var transferSize uint64
			if serverMode {
				transferSize = abd.abd.Abdsend
			} else {
				transferSize = abd.abd.Abdrecv
			}
			if transferSize > 0 {
				if abd.abd.Abdsize != transferSize {
					p := make([]byte, transferSize)
					buffer.Read(p)
					copy(abd.buffer, p)
				} else {
					_, err = buffer.Read(abd.buffer)
					if err != nil {
						return
					}
				}
				if adatypes.Central.IsDebugLevel() {
					adatypes.LogMultiLineString(adatypes.FormatBytes(fmt.Sprintf("Got data of ABD %d :", index), abd.buffer, 8, 16))
				}
			}
		}
	} else {
		adatypes.Central.Log.Debugf("Skip parse ABD buffers")
	}
	adatypes.Central.Log.Infof("Got adabas call " + string(adabas.Acbx.Acbxcmd[:]))
	return
}

func (adabas *Adabas) multifetchBuffer() (helper *adatypes.BufferHelper, err error) {
	for _, abd := range adabas.AdabasBuffers {
		if abd.abd.Abdid == 'M' {
			helper = adatypes.NewHelper(abd.buffer,
				int(abd.abd.Abdrecv), Endian())
			return
		}
	}
	return
}
