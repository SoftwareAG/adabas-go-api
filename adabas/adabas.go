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
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// BuildDate build date
var BuildDate string

// BuildVersion build version
var BuildVersion string

// MaxDatabasesID maximum valid database id
const MaxDatabasesID = 65536

const adaEmptOpt = ' '
const adaFdtXOpt = 'X'

//type adabasOption uint32

// Transaction flags to synchronize and manage different requests
type transactions struct {
	connection   interface{}
	clusterNodes []*URL
}

// CallStatistic statistic of one Adabas call
type CallStatistic struct {
	code       string
	calls      uint64
	timeNeeded time.Duration
}

// Statistics Adabas call statistic of all calls with counting remote calls
type Statistics struct {
	calls         uint64
	success       uint64
	statMap       map[string]*CallStatistic
	remote        uint64
	remoteClosed  uint64
	remoteSend    uint64
	remoteReceive uint64
}

// Adabas is an main Adabas structure containing all call specific parameters
type Adabas struct {
	URL           *URL
	ID            *ID
	status        *Status
	Acbx          *Acbx
	AdabasBuffers []*Buffer
	transactions  *transactions
	statistics    *Statistics
	lock          *sync.Mutex
}

var statistics = false

func init() {
	s := os.Getenv("ADASTATISTICS")
	if strings.ToUpper(s) == "YES" {
		statistics = true
	}
}

// func (option adabasOption) Bit() uint32 {
// 	return (1 << option)
// }

// NewClonedAdabas create a cloned Adabas struct instance
func NewClonedAdabas(clone *Adabas) *Adabas {
	acbx := newAcbx(clone.Acbx.Acbxdbid)

	return &Adabas{
		ID:           clone.ID,
		status:       clone.ID.status(clone.URL.String()),
		Acbx:         acbx,
		URL:          clone.URL,
		transactions: clone.transactions,
		statistics:   clone.statistics,
		lock:         &sync.Mutex{},
	}
}

// NewAdabas create a new Adabas struct instance
func NewAdabas(p ...interface{}) (ada *Adabas, err error) {
	if len(p) == 0 {
		return nil, adatypes.NewGenericError(86)
	}
	var url *URL
	switch u := p[0].(type) {
	case int:
		url = NewURLWithDbid(Dbid(u))
	case Dbid:
		url = NewURLWithDbid(u)
	case string:
		url, err = NewURL(u)
		if err != nil {
			return
		}
	case *URL:
		url = u
	default:
		return nil, adatypes.NewGenericError(87)
	}
	var adaID *ID
	if len(p) > 1 {
		adaID = p[1].(*ID)
	} else {
		adaID = NewAdabasID()
	}
	adatypes.Central.Log.Debugf("Implicit created Adabas instance dbid with ID: %s", adaID.String())
	if (url.Dbid < 1) || (url.Dbid > MaxDatabasesID) {
		err = adatypes.NewGenericError(67, url.Dbid, 1, MaxDatabasesID)
		return nil, err
	}

	acbx := newAcbx(url.Dbid)
	ada = &Adabas{
		ID:           adaID,
		status:       adaID.status(url.String()),
		URL:          url,
		Acbx:         acbx,
		transactions: &transactions{},
		statistics:   newStatistics(),
		lock:         &sync.Mutex{},
	}
	adaID.setAdabas(ada)
	return ada, nil

}

func newStatistics() *Statistics {
	if statistics {
		return &Statistics{statMap: make(map[string]*CallStatistic)}
	}
	return nil
}

// DumpStatistics dump statistics of service
func (adabas *Adabas) DumpStatistics() {
	if adabas.statistics != nil {
		adatypes.Central.Log.Infof("Adabas statistics:")
		for o, s := range adabas.statistics.statMap {
			adatypes.Central.Log.Infof("%s[%s] = %v (%d)", s.code, o, s.timeNeeded, s.calls)
		}
		adatypes.Central.Log.Infof("Remote opened  : %d", adabas.statistics.remote)
		adatypes.Central.Log.Infof("Remote closed  : %d", adabas.statistics.remoteClosed)
		adatypes.Central.Log.Infof("Remote send    : %d", adabas.statistics.remoteSend)
		adatypes.Central.Log.Infof("Remote received: %d", adabas.statistics.remoteReceive)
	}
}

// Open opens a session to the database
func (adabas *Adabas) Open() (err error) {
	return adabas.OpenUser("")
}

// OpenUser opens a session to the database using a user session.
// A USERID must be provided if the user intends to store and/or read user data, and the user wants this data to be available during a subsequent user– or Adabas session.
//    The user intends to store and/or read user data, and the user wants this data to be available during a subsequent user- or Adabas session;
//    The user is to be assigned a special processing priority;
//
// The value provided for the USERID must be unique for this user (not used by any other user), and must begin with a digit or an uppercase letter.
//
// Users for whom none of the above conditions are true should set this field to blanks.
func (adabas *Adabas) OpenUser(user string) (err error) {
	url := adabas.URL.String()
	if adabas.ID.isOpen(url) {
		adatypes.Central.Log.Debugf("Database %s already opened by ID %#v", url, adabas.ID)
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Open database %d %s", adabas.Acbx.Acbxdbid, adabas.ID.String())
	adabas.AdabasBuffers = append(adabas.AdabasBuffers, NewBuffer(AbdAQFb))
	adabas.AdabasBuffers = append(adabas.AdabasBuffers, NewBuffer(AbdAQRb))

	adabas.Acbx.Acbxcmd = op.code()
	copy(adabas.Acbx.Acbxcop[:], []byte{0, 0, 0, 0, 0, 0, 0, 0})
	copy(adabas.Acbx.Acbxcid[0:4], []byte{0, 0, 0, 0})
	if user != "" {
		l := len(user)
		if l > 8 {
			l = 8
		}
		copy(adabas.Acbx.Acbxadd1[:l], user[:l])
	}

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
		adabas.ID.changeOpenState(adabas.URL.String(), true)
		adabas.status.open = true
		adabas.status.platform = adatypes.NewPlatformIsl(adabas.Acbx.Acbxisl)
	} else {
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("Error calling open", err)
		adabas.status.open = false
		adabas.ID.changeOpenState(adabas.URL.String(), false)
	}
	return err
}

// Close A session to the database will be closed
func (adabas *Adabas) Close() {
	url := adabas.URL.String()
	adatypes.Central.Log.Debugf("Close Adabas call %s", url)
	if !adabas.ID.isOpen(url) {
		adatypes.Central.Log.Debugf("Database %s already closed by ID %#v", url, adabas.ID)
		return
	}
	if adabas.ID.transactions(adabas.URL.String()) > 0 {
		err := adabas.BackoutTransaction()
		adatypes.Central.Log.Infof("Error backout during close: %v", err)
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adabas.AdabasBuffers = nil
	adabas.Acbx.Acbxcmd = cl.code()
	ret := adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Close call response ret=%v %s", ret, adabas.ID.String())
	adabas.ID.changeOpenState(adabas.URL.String(), false)
}

// ReleaseCmdID Release any command id resource in the database of the session are released
func (adabas *Adabas) ReleaseCmdID() (err error) {
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adabas.AdabasBuffers = nil
	adabas.Acbx.Acbxcmd = rc.code()
	adabas.Acbx.resetCop()
	err = adabas.CallAdabas()
	return
}

// ReleaseHold Any hold record resource in the database of the session are released
func (adabas *Adabas) ReleaseHold(fileNr Fnr) (err error) {
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adabas.AdabasBuffers = nil
	adabas.Acbx.Acbxcmd = ri.code()
	adabas.Acbx.Acbxfnr = fileNr
	adabas.Acbx.resetCop()
	err = adabas.CallAdabas()
	return
}

func (adabas *Adabas) String() string {
	if adabas == nil {
		return "Adabas <nil>"
	}
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
		return adatypes.NewGenericError(68)
	case "adatcp", "adatcps":
		return adabas.sendTCP()
	case "":
		adabas.Acbx.Acbxrsp = AdaSysCe
		return adatypes.NewGenericError(49)
	}
	return adatypes.NewGenericError(1, adabas.URL.Driver)
}

// sendTCP Send the TCP/IP request to remote Adabas database
func (adabas *Adabas) sendTCP() (err error) {
	var tcpConn *AdaTCP
	// Check if connection is already available
	if adabas.transactions.connection == nil {
		adatypes.Central.Log.Debugf("Establish new context for %p", adabas)

		tcpConn = NewAdaTCP(adabas.URL, Endian(), adabas.ID.AdaID.User,
			adabas.ID.AdaID.Node, adabas.ID.AdaID.Pid, adabas.ID.AdaID.Timestamp)
		tcpConn.stats = adabas.statistics
		err = tcpConn.Connect()
		if err != nil {
			adabas.Acbx.Acbxrsp = AdaSysCe
			adatypes.Central.Log.Infof("Establish TCP context error %v", err)
			err = NewError(adabas)
			return
		}
		adabas.transactions.connection = tcpConn
	} else {
		adatypes.Central.Log.Debugf("Use context for %p %p ", adabas, adabas.transactions.connection)
		tcpConn = adabas.transactions.connection.(*AdaTCP)
	}
	var buffer bytes.Buffer
	err = adabas.WriteBuffer(&buffer, Endian(), false)
	if err != nil {
		adatypes.Central.Log.Infof("Buffer transmit preparation error ", err)
		return
	}
	adatypes.Central.Log.Debugf("Send buffer of length=%d lenBuffer=%d", buffer.Len(), len(adabas.AdabasBuffers))
	err = tcpConn.SendData(buffer, uint32(len(adabas.AdabasBuffers)))
	if err != nil {
		adatypes.Central.Log.Infof("Transmit Adabas call error: %v", err)
		_ = tcpConn.Disconnect()
		adabas.transactions.connection = nil
		return
	}
	buffer.Reset()
	var nrAbdBuffers uint32
	nrAbdBuffers, err = tcpConn.ReceiveData(&buffer, adabasReply)
	if err != nil {
		adatypes.Central.Log.Infof("Receive Adabas call error: %v", err)
		return
	}
	err = adabas.ReadBuffer(&buffer, Endian(), nrAbdBuffers, false)
	if err != nil {
		adatypes.Central.Log.Infof("Read buffer error, destroy context ... %v", err)
		_ = tcpConn.Disconnect()
		return
	}

	adatypes.Central.Log.Debugf("Remote Adabas call returns successfully")
	if adabas.Acbx.Acbxcmd == cl.code() {
		adatypes.Central.Log.Debugf("Close called, destroy context ...")
		if tcpConn != nil {
			_ = tcpConn.Disconnect()
			adabas.transactions.connection = nil
		}
	}
	if tcpConn.clusterNodes != nil {
		adabas.transactions.clusterNodes = tcpConn.clusterNodes
	}
	adatypes.Central.Log.Debugf("Got context for %s %p ", adabas.String(),
		adabas.transactions.connection)
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
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Read file definition with %s", lf.command())
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
	adatypes.Central.Log.Debugf("Read file definition error=%v rsp=%d", err, adabas.Acbx.Acbxrsp)
	if err == nil {
		/* Create new helper to parse returned buffer */
		helper := adatypes.NewHelper(adabas.AdabasBuffers[1].buffer, int(adabas.AdabasBuffers[1].abd.Abdrecv), Endian())
		fdtDefinition := createFdtDefintion()
		fdtDefinition.Values = nil
		_, err = fdtDefinition.ParseBuffer(helper, adatypes.NewBufferOption(false, 0), "")
		if err != nil {
			adatypes.Central.Log.Debugf("ERROR parse FDT: %v", err)
			return
		}
		adatypes.Central.Log.Debugf("Format read field definition")
		definition, err = createFieldDefinitionTable(fdtDefinition)
		if err != nil {
			adatypes.Central.Log.Debugf("ERROR create FDT: %v", err)
			return
		}
		definition.PutCache(cacheName)
		if adatypes.Central.IsDebugLevel() {
			definition.DumpTypes(true, true, "FDT read")
		}
		adatypes.Central.Log.Debugf("Ready parse Format read field definition")
	}
	// Check response to indicate error reading field definition
	if adabas.Acbx.Acbxrsp != 0 {
		adatypes.Central.Log.Infof("Error reading definition: %s", adabas.getAdabasMessage())
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
	adabas.AdabasBuffers[0].abd.Abdrecv = 0
	adatypes.Central.Log.Debugf("ABD init F send %d\n", adabas.AdabasBuffers[0].abd.Abdsend)

	// Create record buffer for the call
	adabas.AdabasBuffers[1] = NewBufferWithSize(AbdAQRb,
		multifetch*(adabasRequest.RecordBufferLength+adabasRequest.RecordBufferShift))
	adabas.AdabasBuffers[1].abd.Abdsend = 0
	adabas.AdabasBuffers[1].abd.Abdrecv = adabas.AdabasBuffers[0].abd.Abdsize
	adatypes.Central.Log.Debugf("ABD init R send %d buffer length %d\n", adabas.AdabasBuffers[1].abd.Abdsend, adabas.AdabasBuffers[1].abd.Abdsize)

	// Define search and value buffer to search
	if adabasRequest.SearchTree != nil {
		adatypes.Central.Log.Debugf("Search logical added")
		adabas.AdabasBuffers[2] = SearchAdabasBuffer(adabasRequest.SearchTree)
		adatypes.Central.Log.Debugf("ABD init S send %d\n", adabas.AdabasBuffers[2].abd.Abdsend)
		adabas.AdabasBuffers[3] = ValueAdabasBuffer(adabasRequest.SearchTree)
		adatypes.Central.Log.Debugf("ABD init V send %d\n", adabas.AdabasBuffers[3].abd.Abdsend)
	}
	if adabasRequest.Multifetch > 1 {
		adatypes.Central.Log.Debugf("Create multifetch buffer for %d multifetch entries", adabasRequest.Multifetch)
		index := len(adabas.AdabasBuffers) - 1
		adabas.AdabasBuffers[index] = NewBufferWithSize(AbdAQMb, 4+(adabasRequest.Multifetch*16))
	}

}

// ReadPhysical read data in physical order
func (adabas *Adabas) ReadPhysical(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Physical read file ... %s", l2.command())
	if adabasRequest.HoldRecords.IsHold() {
		adabas.Acbx.Acbxcmd = l5.code()
	} else {
		adabas.Acbx.Acbxcmd = l2.code()
	}
	nrMultifetch := 2
	adabas.Acbx.resetCop()
	if adabasRequest.Multifetch > 1 {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'O'
		} else {
			adabas.Acbx.Acbxcop[0] = 'M'
		}
		nrMultifetch = 3
	} else {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'R'
		}
	}
	adabas.Acbx.Acbxcop[2] = adabasRequest.HoldRecords.HoldOption()
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
		adabas.Acbx.Acbxisl = uint64(multifetch)
	}

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// read a specific ISN out of Adabas file
func (adabas *Adabas) readISN(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	if adabasRequest.HoldRecords.IsHold() {
		adatypes.Central.Log.Debugf("Read ISN %d ... %s dbid=%d fnr=%d", adabasRequest.Isn, l4.command(), adabas.Acbx.Acbxdbid, fileNr)
		adabas.Acbx.Acbxcmd = l4.code()
	} else {
		adatypes.Central.Log.Debugf("Read ISN %d ... %s dbid=%d fnr=%d", adabasRequest.Isn, l1.command(), adabas.Acbx.Acbxdbid, fileNr)
		adabas.Acbx.Acbxcmd = l1.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxisn = adabasRequest.Isn
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}
	adabas.Acbx.Acbxfnr = fileNr
	if adabasRequest.HoldRecords == adatypes.HoldResponse {
		adabas.Acbx.Acbxcop[0] = 'R'
	}
	// adabas.Acbx.Acbxcop[2] = adabasRequest.HoldRecords.HoldOption()
	switch adabasRequest.HoldRecords {
	case adatypes.HoldResponse, adatypes.HoldNone, adatypes.HoldWait:
	default:
		return adatypes.NewGenericError(95)
	}

	adabas.prepareBuffers(adabasRequest)

	err = adabas.loopCall(adabasRequest, x)
	return
}

// ReadISNOrder Read logical using a descriptor
func (adabas *Adabas) ReadISNOrder(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	if adabasRequest.HoldRecords.IsHold() {
		adatypes.Central.Log.Debugf("Read ISN order ... %s dbid=%d multifetch=%d", l4.command(), adabas.Acbx.Acbxdbid, adabasRequest.Multifetch)
		adabas.Acbx.Acbxcmd = l4.code()
	} else {
		adatypes.Central.Log.Debugf("Read ISN order ... %s dbid=%d multifetch=%d", l1.command(), adabas.Acbx.Acbxdbid, adabasRequest.Multifetch)
		adabas.Acbx.Acbxcmd = l1.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = 'I'
	if adabasRequest.Multifetch > 1 {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'O'
		} else {
			adabas.Acbx.Acbxcop[0] = 'M'
		}
		adabas.Acbx.Acbxisl = uint64(adabasRequest.Multifetch)
	} else {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'R'
		}
	}
	adabas.Acbx.Acbxcop[2] = adabasRequest.HoldRecords.HoldOption()

	adabas.Acbx.Acbxisn = adabasRequest.Isn
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabas.prepareBuffers(adabasRequest)
	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// ReadLogicalWith Read logical using a descriptor
func (adabas *Adabas) ReadLogicalWith(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Read logical ... %s dbid=%d multifetch=%d", l3.command(), adabas.Acbx.Acbxdbid, adabasRequest.Multifetch)
	if adabasRequest.HoldRecords.IsHold() {
		adabas.Acbx.Acbxcmd = l6.code()
	} else {
		adabas.Acbx.Acbxcmd = l3.code()
	}
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxisn = adabasRequest.Isn
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxcop[1] = 'A'
	if adabasRequest.Multifetch > 1 {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'O'
		} else {
			adabas.Acbx.Acbxcop[0] = 'M'
		}
		adabas.Acbx.Acbxisl = uint64(adabasRequest.Multifetch)
	} else {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'R'
		}
	}
	adabas.Acbx.Acbxcop[2] = adabasRequest.HoldRecords.HoldOption()

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
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:8])

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// ReadStream Read partial lob stream
func (adabas *Adabas) ReadStream(adabasRequest *adatypes.Request, offset uint64, x interface{}) (err error) {
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adabas.Acbx.Acbxcmd = l1.code()
	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = 'L'
	adabas.Acbx.Acbxisl = offset
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabasRequest.Multifetch = 1
	adabasRequest.Limit = 1

	adabas.prepareBuffers(adabasRequest)

	err = adabas.loopCall(adabasRequest, x)
	return
}

// SearchLogicalWith Search logical using a descriptor
func (adabas *Adabas) SearchLogicalWith(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Search logical ... %s dbid=%d hold=%v", s2.command(), adabas.Acbx.Acbxdbid, adabasRequest.HoldRecords.IsHold())
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
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:8])

	adabas.Acbx.Acbxfnr = fileNr
	// Call Adabas
	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Received search response ret=%v ISN quantity=%d", err, adabas.Acbx.Acbxisq)
	if err != nil {
		return
	}
	// End of file reached
	if adabas.Acbx.Acbxrsp == AdaEOF || adabas.Acbx.Acbxisq == 0 {
		return
	}

	if adabasRequest.HoldRecords.IsHold() {
		adatypes.Central.Log.Debugf("Read logical after search ... %s dbid=%d", l4.command(), adabas.Acbx.Acbxdbid)
		adabas.Acbx.Acbxcmd = l4.code()
	} else {
		adatypes.Central.Log.Debugf("Read logical after search ... %s dbid=%d", l1.command(), adabas.Acbx.Acbxdbid)
		adabas.Acbx.Acbxcmd = l1.code()
	}
	adabas.Acbx.resetCop()
	if adabasRequest.Multifetch > 1 {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'O'
		} else {
			adabas.Acbx.Acbxcop[0] = 'M'
		}
		adabas.Acbx.Acbxisl = uint64(adabasRequest.Multifetch)
	} else {
		if adabasRequest.HoldRecords == adatypes.HoldResponse {
			adabas.Acbx.Acbxcop[0] = 'R'
		}
	}
	adabas.Acbx.Acbxcop[1] = 'N'
	adabas.Acbx.Acbxcop[2] = adabasRequest.HoldRecords.HoldOption()
	err = adabas.loopCall(adabasRequest, x)
	return
}

// Loop call used to read a sequence of records
func (adabas *Adabas) loopCall(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Loop call records avail.=%v", (adabasRequest.Definition.Values != nil))
	count := uint64(0)
	adabasRequest.CmdCode = adabas.Acbx.Acbxcmd
	switch adabas.Acbx.Acbxcmd {
	case l1.code(), l4.code():
		adabasRequest.IsnIncrease = true
		adabasRequest.StoreIsn = true
	case l2.code(), l5.code():
		adabasRequest.IsnIncrease = false
		adabasRequest.StoreIsn = false
	default:
		adabasRequest.IsnIncrease = false
		adabasRequest.StoreIsn = true
	}
	if adabasRequest.Parameter == nil {
		adabasRequest.Reference = fmt.Sprintf("db/%d/%d", adabas.Acbx.Acbxdbid, adabas.Acbx.Acbxfnr)
	} else {
		adabasMap := adabasRequest.Parameter.(*Map)
		adatypes.Central.Log.Debugf("%v -> %#v\n", adabasRequest.Parameter, adabasMap)
		if adabasMap != nil {
			adabasRequest.Reference = fmt.Sprintf("map/%s", adabasMap.Name)
		}
	}
	var responseCode uint32
	for responseCode == 0 {
		if !(adabasRequest.Option.SecondCall > 0 || adabasRequest.Option.StreamCursor > 0) {
			err = adabasRequest.Definition.CreateValues(false)
			if err != nil {
				return
			}
		}
		adabas.resetSendSize()
		// if adabas.Acbx.Acbxcop[0] == 'M' {
		// 	adabas.Acbx.Acbxisl = 0
		// }
		adatypes.Central.Log.Debugf("Send call avail.=%v", (adabasRequest.Definition.Values != nil))
		// Call Adabas
		err = adabas.CallAdabas()
		adatypes.Central.Log.Debugf("Received call response ret=%v", err)
		if err != nil {
			return
		}

		adabasRequest.Caller = adabas
		adabasRequest.Response = adabas.Acbx.Acbxrsp

		// End of file reached
		if adabas.Acbx.Acbxrsp == AdaEOF {
			return
		}
		// Error received from Adabas
		if adabas.Acbx.Acbxrsp != AdaNormal {
			adatypes.Central.Log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
			err = NewError(adabas)
			return
		}
		adabasRequest.Isn = adatypes.Isn(adabas.Acbx.Acbxisn)
		adabasRequest.IsnQuantity = adabas.Acbx.Acbxisq
		adatypes.Central.Log.Debugf("ISN= %d ISN quantity=%d multifetch=%d", adabasRequest.Isn, adabasRequest.IsnQuantity,
			adabasRequest.Multifetch)

		adabasRequest.RecordBuffer = adatypes.NewHelper(adabas.AdabasBuffers[1].buffer,
			int(adabas.AdabasBuffers[1].abd.Abdrecv), Endian())
		adabasRequest.MultifetchBuffer, err = adabas.multifetchBuffer()
		if err != nil {
			return
		}
		if adabasRequest.Parser == nil {
			break
		}
		adabasRequest.CbIsn = adabas.Acbx.Acbxisn
		if adabasRequest.IsnIncrease {
			adabas.Acbx.Acbxisn++
		}
		responseCode, err = adabasRequest.ParseBuffer(&count, x)
		if err != nil {
			return
		}
		adabas.Acbx.Acbxisn = adabasRequest.CbIsn
		adatypes.Central.Log.Debugf("Loop step ended Limit=%d count=%d", adabasRequest.Limit, count)
		if (adabasRequest.Limit > 0) && (count >= adabasRequest.Limit) {
			adatypes.Central.Log.Debugf("Limit reached")
			break
		}
	}
	adatypes.Central.Log.Debugf("Loop call ended count=%d", count)

	return
}

func (adabas *Adabas) resetSendSize() {
	for _, abd := range adabas.AdabasBuffers {
		abd.resetSendSize()
	}
}

// SendSecondCall do second call reading lob data or multiple fields of the period group
func (adabas *Adabas) SendSecondCall(adabasRequest *adatypes.Request, x interface{}) (err error) {
	adatypes.Central.Log.Debugf("Check second call .... values avail.=%v", (adabasRequest.Definition.Values == nil))
	if adabasRequest.Option.NeedSecondCall != adatypes.NoneSecond {
		adatypes.Central.Log.Debugf("Need second call %v", adabasRequest.Option.NeedSecondCall)
		parameter := &adatypes.AdabasRequestParameter{Store: false, SecondCall: 1, Mainframe: adabas.status.platform.IsMainframe()}
		tmpAdabasRequest, err2 := adabasRequest.Definition.CreateAdabasRequest(parameter)
		if err2 != nil {
			err = err2
			return
		}
		acbx := *adabas.Acbx
		abd := adabas.AdabasBuffers
		tmpAdabasRequest.Isn = adabasRequest.Isn
		tmpAdabasRequest.Definition = adabasRequest.Definition
		tmpAdabasRequest.RecordBufferShift = adabasRequest.RecordBufferShift
		tmpAdabasRequest.Multifetch = 1
		tmpAdabasRequest.Option.SecondCall = 1
		adatypes.Central.Log.Debugf("Call second request to ISN=%d only", tmpAdabasRequest.Isn)
		err = adabas.readISN(adabas.Acbx.Acbxfnr, tmpAdabasRequest, x)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Read ISN done, parse buffer of second call")
		_, err = tmpAdabasRequest.Definition.ParseBuffer(tmpAdabasRequest.RecordBuffer, tmpAdabasRequest.Option, "")
		if err != nil {
			adatypes.Central.Log.Debugf("Parse buffer of second call  with error: ", err)
			return
		}
		adatypes.Central.Log.Debugf("Parse buffer of second call ended, reset to old adabas request")
		*adabas.Acbx = acbx
		adabas.AdabasBuffers = abd
		adatypes.Central.Log.Debugf("Second call done")

		adabasRequest.Option.NeedSecondCall = adatypes.NoneSecond
	}

	return
}

// Histogram histogram of a specific descriptor
func (adabas *Adabas) Histogram(fileNr Fnr, adabasRequest *adatypes.Request, x interface{}) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Descriptor read file %s", l9.command())
	adabas.Acbx.Acbxcmd = l9.code()
	adabas.Acbx.Acbxisn = 0
	adabas.Acbx.Acbxisl = 0
	adabas.Acbx.Acbxisq = 0
	adabas.Acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}

	adabas.Acbx.resetCop()
	adabas.Acbx.Acbxcop[1] = 'A'
	if adabasRequest.Multifetch > 1 {
		adabas.Acbx.Acbxcop[0] = 'M'
		adabas.Acbx.Acbxisl = uint64(adabasRequest.Multifetch)
	}

	adabas.prepareBuffers(adabasRequest)

	var add1 bytes.Buffer
	for _, d := range adabasRequest.Descriptors {
		add1.WriteString(d)
	}
	add1.WriteString("        ")
	copy(adabas.Acbx.Acbxadd1[:], add1.Bytes()[0:8])

	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.loopCall(adabasRequest, x)
	return
}

// Store store a record into database
func (adabas *Adabas) Store(fileNr Fnr, adabasRequest *adatypes.Request) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Call store, pending transactions=%d adabas=%p",
		adabas.ID.transactions(adabas.URL.String()), adabas)
	if adabasRequest.Isn != 0 {
		adatypes.Central.Log.Debugf("Store specific ISN ... %s", n2.command())
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
	adatypes.Central.Log.Debugf("Store ISN rsp=%d ... ISN=%d", adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxisn)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		adatypes.Central.Log.Errorf("Error storing data: %s", adabas.getAdabasMessage())
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("%v", err)
		return
	}
	adabas.ID.incTransactions(adabas.URL.String())
	adabasRequest.Isn = adabas.Acbx.Acbxisn
	return
}

// Update update a record in database
func (adabas *Adabas) Update(fileNr Fnr, adabasRequest *adatypes.Request) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Call update, pending transactions=%d adabas=%p",
		adabas.ID.transactions(adabas.URL.String()), adabas)
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
		adatypes.Central.Log.Errorf("Error updating data: %s", adabas.getAdabasMessage())
		err = NewError(adabas)
		adatypes.Central.Log.Debugf("%v", err)
		return
	}
	adabas.ID.incTransactions(adabas.URL.String())
	adabasRequest.Isn = adabas.Acbx.Acbxisn
	return
}

// SetURL set new database URL
func (adabas *Adabas) SetURL(URL *URL) {
	if adabas.URL == URL {
		return
	}
	adabas.Close()
	adabas.Acbx.Acbxdbid = URL.Dbid
	adabas.URL = URL
	// Different adabas instance, need to update status
	adabas.status = adabas.ID.status(adabas.URL.String())
}

// SetDbid set new database id
func (adabas *Adabas) SetDbid(dbid Dbid) {
	if dbid == adabas.Acbx.Acbxdbid {
		return
	}
	adabas.Close()
	adabas.Acbx.Acbxdbid = dbid
	adabas.URL = NewURLWithDbid(dbid)
	// Different adabas instance, need to update status
	adabas.status = adabas.ID.status(adabas.URL.String())
}

// DeleteIsn delete a single isn
func (adabas *Adabas) DeleteIsn(fileNr Fnr, isn adatypes.Isn) (err error) {
	err = adabas.Open()
	if err != nil {
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Delete ISN transactions=%d adabas=%p", adabas.ID.transactions(adabas.URL.String()),
		adabas)
	adatypes.Central.Log.Debugf("Delete Isn ...%s on dbid %d and file %d", e1.command(), adabas.Acbx.Acbxdbid, fileNr)
	adabas.Acbx.Acbxcmd = e1.code()
	adabas.Acbx.Acbxisn = isn
	adabas.Acbx.Acbxfnr = fileNr

	err = adabas.CallAdabas()
	if err != nil {
		adatypes.Central.Log.Debugf("Delete isn call response error=%v", err)
		return
	}
	adabas.ID.incTransactions(adabas.URL.String())
	adatypes.Central.Log.Debugf("Delete ISN error ...%d transactions=%d adabas=%p", adabas.Acbx.Acbxrsp,
		adabas.ID.transactions(adabas.URL.String()), adabas)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		adatypes.Central.Log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
		adatypes.Central.Log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// BackoutTransaction backout transaction initiated
func (adabas *Adabas) BackoutTransaction() (err error) {
	url := adabas.URL.String()
	if !adabas.ID.isOpen(url) {
		adatypes.Central.Log.Debugf("Database %s already opened by ID %#v", url, adabas.ID)
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("Open flag %p bt", adabas)
	if adabas.ID.transactions(adabas.URL.String()) == 0 {
		return
	}
	adatypes.Central.Log.Debugf("Backout transaction ... %s", bt.command())
	adabas.Acbx.Acbxcmd = bt.code()
	adabas.AdabasBuffers = nil

	ret := adabas.CallAdabas()
	adatypes.Central.Log.Debugf("Backout transaction rsp ... ret=%d rsp=%d", ret, adabas.Acbx.Acbxrsp)
	adabas.ID.clearTransactions(adabas.URL.String())

	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		adatypes.Central.Log.Errorf("Error reading data: %s", adabas.getAdabasMessage())
		adatypes.Central.Log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// EndTransaction end of transaction initiated
func (adabas *Adabas) EndTransaction() (err error) {
	adatypes.Central.Log.Debugf("End of transaction pending=%d adabas=%p",
		adabas.ID.transactions(adabas.URL.String()), adabas)

	if adabas.ID.transactions(adabas.URL.String()) == 0 {
		adatypes.Central.Log.Debugf("End of transaction ... not pending transactions")
		return
	}
	adabas.lock.Lock()
	defer adabas.lock.Unlock()
	adatypes.Central.Log.Debugf("End of transaction ... %s", et.command())
	adabas.Acbx.Acbxcmd = et.code()
	adabas.AdabasBuffers = nil

	err = adabas.CallAdabas()
	adatypes.Central.Log.Debugf("End of transction response ret=%v", err)
	if err != nil {
		return
	}
	adabas.ID.clearTransactions(adabas.URL.String())
	adatypes.Central.Log.Debugf("End of transaction rsp ... rsp=%d", adabas.Acbx.Acbxrsp)
	// Error received from Adabas
	if adabas.Acbx.Acbxrsp != AdaNormal {
		adatypes.Central.Log.Errorf("Error end transaction: %s", adabas.getAdabasMessage())
		adatypes.Central.Log.Errorf("CB: %s", adabas.Acbx.String())
		err = NewError(adabas)
		return
	}
	return
}

// WriteBuffer write adabas call to buffer
func (adabas *Adabas) WriteBuffer(buffer *bytes.Buffer, order binary.ByteOrder, serverMode bool) (err error) {
	defer TimeTrack(time.Now(), "Adabas Write buffer", adabas)
	// xx fmt.Sprintf("Adabas Write buffer %s rsp=%d subrsp=%d",
	// 	string(adabas.Acbx.Acbxcmd[:]), adabas.Acbx.Acbxrsp, adabas.Acbx.Acbxerrc))
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
			return adatypes.NewGenericError(74, index)
		}
		err = binary.Write(&tempBuffer, Endian(), abd.abd)
		if err != nil {
			adatypes.Central.Log.Debugf("Write ABD in buffer error %v", err)
			return
		}
		b := tempBuffer.Bytes()
		if b[2] != 'G' {
			return adatypes.NewGenericError(75)
		}
		buffer.Write(b)
		adatypes.Central.Log.Debugf("Add ADABAS ABD: %d to len buffer=%d", index, buffer.Len())
	}
	adatypes.Central.Log.Debugf("Index of end ABD: %d/%X", buffer.Len(), buffer.Len())
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
	defer TimeTrack(time.Now(), "Adabas Read buffer", adabas)
	if buffer == nil {
		err = adatypes.NewGenericError(4)
		return
	}
	adatypes.Central.Log.Debugf("Read buffer, read  ACBX: ")
	err = binary.Read(buffer, Endian(), adabas.Acbx)
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
		adatypes.Central.Log.Debugf("Got nCalBuf=%d Number of current ABDS=%d len=%d", nCalBuf, len(adabas.AdabasBuffers), buffer.Len())
		for index, abd := range adabas.AdabasBuffers {
			if adatypes.Central.IsDebugLevel() {
				adatypes.Central.Log.Debugf("Parse %d.ABD got %c rest len=%d\n", index, abd.abd.Abdid, buffer.Len())
				adatypes.LogMultiLineString(adatypes.FormatBytes("Rest ABD:", buffer.Bytes(), buffer.Len(), 8, 16, false))
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
					_, err = buffer.Read(p)
					if err != nil {
						return
					}
					copy(abd.buffer, p)
				} else {
					_, err = buffer.Read(abd.buffer)
					if err != nil {
						return
					}
				}
				if adatypes.Central.IsDebugLevel() {
					adatypes.LogMultiLineString(adatypes.FormatBytes(fmt.Sprintf("Got data of ABD %d :", index), abd.buffer, len(abd.buffer), 8, 16, false))
				}
			}
		}
	} else {
		adatypes.Central.Log.Debugf("Skip parse ABD buffers")
	}
	adatypes.Central.Log.Debugf("Got adabas call " + string(adabas.Acbx.Acbxcmd[:]))
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

// TimeTrack defer function measure the difference end log it to log management, like
//    defer TimeTrack(time.Now(), "CallAdabas "+string(adabas.Acbx.Acbxcmd[:]))
func TimeTrack(start time.Time, name string, adabas *Adabas) {
	elapsed := time.Since(start)
	if adabas == nil {
		adatypes.Central.Log.Debugf("%s took %s", name, elapsed)
		return
	}
	acbx := adabas.Acbx
	if adabas.statistics != nil && name == "Call adabas" {
		adabas.statistics.calls++
		if acbx.Acbxrsp == 0 {
			adabas.statistics.success++
		}
		if s, ok := adabas.statistics.statMap[string(acbx.Acbxcmd[:])]; ok {
			s.timeNeeded = s.timeNeeded + elapsed
			s.calls++
			adatypes.Central.Log.Debugf("%s: Call statistics %s took %d", name, s.code, s.calls)
		} else {
			sNew := &CallStatistic{timeNeeded: elapsed, calls: 1, code: string(acbx.Acbxcmd[:])}
			adabas.statistics.statMap[sNew.code] = sNew
			adatypes.Central.Log.Debugf("%s: Call statistics %s took %d", name, sNew.code, sNew.calls)
		}
	}
	adatypes.Central.Log.Debugf("%s took %s, %s file=%d rsp=%d subrsp=%d add2=%#v", name, elapsed,
		string(acbx.Acbxcmd[:]), acbx.Acbxfnr, acbx.Acbxrsp, acbx.Acbxerrc, []byte(acbx.Acbxadd2[:]))
}
