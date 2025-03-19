/*
* Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// BufferType type of buffer following
type BufferType uint32

const (
	adatcpHeaderEyecatcher = "ADATCP"

	adatcpHeaderVersion = "01"
	// ConnectRequest connect
	ConnectRequest = BufferType(1)
	// ConnectReply reply after first connect
	ConnectReply = BufferType(2)
	// ConnectError connection errror
	ConnectError = BufferType(3)
	// DisconnectRequest disconnect request
	DisconnectRequest = BufferType(4)
	// DisconnetReply disconnect reply
	DisconnetReply = BufferType(5)
	// DisconnectError disconnect error
	DisconnectError = BufferType(6)
	// DataRequest data request
	DataRequest = BufferType(7)
	// DataReply data reply
	DataReply = BufferType(8)
	// DataError data error
	DataError = BufferType(9)
)

type adaUUID [16]byte

const (
	adaTCPNoCluster = 'C'
	adaTCPCluster   = 'G'
)

// AdaTCPHeader Adabas TCP Header ADATCP
type AdaTCPHeader struct {
	Eyecatcher     [6]byte
	Version        [2]byte
	Length         uint32
	BufferType     BufferType
	Identification adaUUID
	ErrorCode      uint32
	DatabaseType   byte
	Reserved       [3]byte
}

// AdaTCPHeaderLength length of AdaTCPHeader structure
const AdaTCPHeaderLength = 40

const (
	adatcpBigEndian    = byte(1)
	adatcpLittleEndian = byte(2)

	adatcpASCII8 = byte(1)
	//adatcpEBCDIC = byte(2)

	adatcpFloatIEEE = byte(1)
)

// AdaTCPConnectPayload Adabas TCP connect payload
type AdaTCPConnectPayload struct {
	DatabaseVersion [16]byte
	DatabaseName    [16]byte
	Userid          [8]byte
	Nodeid          [8]byte
	ProcessID       uint32
	DatabaseID      uint32
	TimeStamp       uint64
	Endianness      byte
	Charset         byte
	Floatingpoint   byte
	Filler          [5]byte
}

// AdaTCPConnectPayloadLength ADATCP connect payload
const AdaTCPConnectPayloadLength = 72

type adatcpDisconnectPayload struct {
	Dummy uint64
}

type adaTCPID struct {
	user      [8]byte
	node      [8]byte
	pid       uint32
	timestamp uint64
}

// AdaTCP TCP connection handle (for internal use only)
type AdaTCP struct {
	connection          net.Conn
	URL                 *URL
	order               binary.ByteOrder
	adauuid             adaUUID
	serverEndianness    byte
	serverCharset       byte
	serverFloatingpoint byte
	databaseVersion     [16]byte
	databaseName        [16]byte
	databaseID          uint32
	pair                []string
	id                  adaTCPID
	stats               *Statistics
	databaseType        byte
	clusterNodes        []*URL
}

func adatcpTCPClientHTON8(l uint64) uint64 {
	return uint64(
		((uint64(l) >> 56) & uint64(0x00000000000000ff)) | ((uint64(l) >> 40) & uint64(0x000000000000ff00)) | ((uint64(l) >> 24) & uint64(0x0000000000ff0000)) | ((uint64(l) >> 8) & uint64(0x00000000ff000000)) | ((uint64(l) << 8) & uint64(0x000000ff00000000)) | ((uint64(l) << 24) & uint64(0x0000ff0000000000)) | ((uint64(l) << 40) & uint64(0x00ff000000000000)) | ((uint64(l) << 56) & uint64(0xff00000000000000)))
}

// func adatcpTCPClientHTON4(l uint32) uint32 {
// 	return uint32(
// 		((uint32(l) >> 24) & uint32(0x000000ff)) | ((uint32(l) >> 8) & uint32(0x0000ff00)) | ((uint32(l) << 8) & uint32(0x00ff0000)) | ((uint32(l) << 24) & uint32(0xff000000)))
// }

// NewAdatcpHeader new Adabas TCP header
func NewAdatcpHeader(bufferType BufferType) AdaTCPHeader {
	header := AdaTCPHeader{BufferType: BufferType(uint32(bufferType))}
	copy(header.Eyecatcher[:], adatcpHeaderEyecatcher)
	copy(header.Version[:], adatcpHeaderVersion)
	return header
}

// Valid check AdaTcp Header valid eyecatcher
func (header AdaTCPHeader) Valid() bool {
	return string(header.Eyecatcher[:]) == adatcpHeaderEyecatcher
}

func bigEndian() (ret bool) {
	i := 0x1
	bs := (*[4]byte)(unsafe.Pointer(&i))
	return bs[0] == 0
}

// Endian current byte order of the client system
func Endian() binary.ByteOrder {
	if bigEndian() {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// NewAdaTCP create new ADATCP connection to remote TCP/IP Adabas nucleus
func NewAdaTCP(URL *URL, order binary.ByteOrder, user [8]byte, node [8]byte,
	pid uint32, timestamp uint64) *AdaTCP {
	if URL == nil {
		return nil
	}
	pair := URL.searchCertificate()
	t := &AdaTCP{URL: URL, order: order, pair: pair,
		id: adaTCPID{pid: pid, timestamp: timestamp}}
	copy(t.id.user[:], user[:])
	copy(t.id.node[:], node[:])
	return t
}

// Send Send the TCP/IP request to remote Adabas database
func (connection *AdaTCP) Send(adaInstance *Adabas) (err error) {
	var buffer bytes.Buffer
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Call Adabas using ADATCP")
	}
	err = adaInstance.WriteBuffer(&buffer, Endian(), false)
	if err != nil {
		adatypes.Central.Log.Debugf("Buffer transmit preparation error ", err)
		return
	}
	if debug {
		adatypes.Central.Log.Debugf("Send buffer of length=%d lenBuffer=%d", buffer.Len(), len(adaInstance.AdabasBuffers))
		adatypes.LogMultiLineString(true, adaInstance.Acbx.String())
	}
	err = connection.SendData(buffer, uint32(len(adaInstance.AdabasBuffers)))
	if err != nil {
		adatypes.Central.Log.Debugf("Transmit Adabas call error: %v", err)
		_ = connection.Disconnect()
		adaInstance.transactions.connection = nil
		return
	}
	buffer.Reset()
	var nrAbdBuffers uint32
	nrAbdBuffers, err = connection.ReceiveData(&buffer, adabasReply)
	if err != nil {
		adatypes.Central.Log.Debugf("Receive Adabas call error: %v", err)
		return
	}
	err = adaInstance.ReadBuffer(&buffer, Endian(), nrAbdBuffers, false)
	if err != nil {
		adatypes.Central.Log.Debugf("Read buffer error, destroy context ... %v", err)
		_ = connection.Disconnect()
		return
	}

	if debug {
		adatypes.Central.Log.Debugf("Remote Adabas call returns successfully")
	}
	if adaInstance.Acbx.Acbxcmd == cl.code() {
		if debug {
			adatypes.Central.Log.Debugf("Close called, destroy context ...")
		}
		if connection != nil {
			_ = connection.Disconnect()
			adaInstance.transactions.connection = nil
		}
	}
	if connection.clusterNodes != nil {
		adaInstance.transactions.clusterNodes = connection.clusterNodes
	}
	if debug {
		adatypes.Central.Log.Debugf("Got context for %s %p ", adaInstance.String(),
			adaInstance.transactions.connection)
		adatypes.LogMultiLineString(true, adaInstance.Acbx.String())
	}
	return
}

// Connect establish connection to ADATCP server
func (connection *AdaTCP) Connect(adabas *Adabas) (err error) {
	connection.stats = adabas.statistics
	return connection.tcpConnect()
}

// tcpConnect establish connection to ADATCP server
func (connection *AdaTCP) tcpConnect() (err error) {
	url := fmt.Sprintf("%s:%d", connection.URL.Host, connection.URL.Port)
	debug := adatypes.Central.IsDebugLevel()
	if debug {
		adatypes.Central.Log.Debugf("Open/Connect TCP connection to %s", url)
	}
	addr, _ := net.ResolveTCPAddr("tcp", url)

	switch connection.URL.Driver {
	case "adatcp":
		tcpConn, tcpErr := net.DialTCP("tcp", nil, addr)
		err = tcpErr
		if err != nil {
			adatypes.Central.Log.Debugf("Connect error : %v", err)
			return
		}
		if connection.stats != nil {
			connection.stats.remote++
		}
		if debug {
			adatypes.Central.Log.Debugf("Connect dial passed ...")
		}
		connection.connection = tcpConn
		_ = tcpConn.SetNoDelay(true)
	case "adatcps":
		err = connection.createSSLConnection(url)
		if err != nil {
			adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
			return
		}
	default:
		return adatypes.NewGenericError(131)
	}
	header := NewAdatcpHeader(ConnectRequest)
	payload := AdaTCPConnectPayload{Charset: adatcpASCII8, Floatingpoint: adatcpFloatIEEE}
	copy(payload.Userid[:], connection.id.user[:])
	copy(payload.Nodeid[:], connection.id.node[:])
	payload.ProcessID = connection.id.pid
	payload.TimeStamp = connection.id.timestamp

	header.Length = uint32(AdaTCPHeaderLength + unsafe.Sizeof(payload))
	var buffer bytes.Buffer
	err = binary.Write(&buffer, binary.BigEndian, header)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}
	if bigEndian() {
		if debug {
			adatypes.Central.Log.Debugf("Write TCP payload for big endian")
		}
		payload.Endianness = adatcpBigEndian
	} else {
		if debug {
			adatypes.Central.Log.Debugf("Write TCP payload for little endian")
		}
		payload.Endianness = adatcpLittleEndian
	}
	if debug {
		adatypes.Central.Log.Debugf("Buffer size after header=%d", buffer.Len())
	}

	// Send payload in big endian needed until remote knows the endianess of the client
	err = binary.Write(&buffer, binary.BigEndian, payload)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP generate payload buffer error: %v", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}
	if debug {
		adatypes.Central.Log.Debugf("Buffer size after payload=%d", buffer.Len())
	}

	send := buffer.Bytes()
	if debug {
		adatypes.LogMultiLineString(true, adatypes.FormatBytes("Connect PAYLOAD:", send, len(send), 8, 16, true))
	}
	_, err = connection.connection.Write(send)
	if err != nil {
		adatypes.Central.Log.Debugf("Error TCP writing data %s", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}
	rcvBuffer := make([]byte, buffer.Len())
	_, err = io.ReadFull(connection.connection, rcvBuffer)
	//	_, err = connection.connection.Read(rcvBuffer)
	if err != nil {
		adatypes.Central.Log.Debugf("Error TCP reading data %v", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}

	if debug {
		adatypes.LogMultiLineString(true, adatypes.FormatBytes("RCV Reply PAYLOAD:", rcvBuffer, len(rcvBuffer), 8, 16, true))
	}

	buf := bytes.NewBuffer(rcvBuffer)
	err = binary.Read(buf, binary.BigEndian, &header)
	if err != nil {
		adatypes.Central.Log.Debugf("Error TCP buffer parsing header %v", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}
	if !header.Valid() {
		connection.connection.Close()
		connection.connection = nil
		return adatypes.NewGenericError(160)
	}
	if header.BufferType != ConnectReply {
		connection.connection.Close()
		connection.connection = nil
		return adatypes.NewGenericError(161)
	}

	err = binary.Read(buf, binary.BigEndian, &payload)
	if err != nil {
		adatypes.Central.Log.Debugf("Error parsing payload %v", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}

	connection.adauuid = header.Identification
	connection.serverEndianness = payload.Endianness
	connection.serverCharset = payload.Charset
	connection.serverFloatingpoint = payload.Floatingpoint
	connection.databaseVersion = payload.DatabaseVersion
	connection.databaseName = payload.DatabaseName
	connection.databaseID = payload.DatabaseID
	connection.databaseType = header.DatabaseType
	if header.DatabaseType == adaTCPCluster {
		adatypes.Central.Log.Debugf("Database cluster found %v", header.DatabaseType)
		err = connection.receiveNodeList()
	}
	return
}

// createSSLConnection create SSL TCP connection
func (connection *AdaTCP) createSSLConnection(url string) (err error) {
	//		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	config := tls.Config{InsecureSkipVerify: true}
	if len(connection.pair) == 2 {
		adatypes.Central.Log.Debugf("Load key pair")
		cert, cerr := tls.LoadX509KeyPair(connection.pair[0], connection.pair[1])
		if cerr != nil {
			adatypes.Central.Log.Debugf("server: loadkeys: %s", cerr)
			return cerr
		}
		config.Certificates = []tls.Certificate{cert}
		config.InsecureSkipVerify = false
	} else {
		adatypes.Central.Log.Debugf("No key pair defined")
	}
	tcpConn, tcpErr := tls.Dial("tcp", url, &config)
	err = tcpErr
	if err != nil {
		adatypes.Central.Log.Debugf("Connect error : %v", err)
		return
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("client: connected to: %v", tcpConn.RemoteAddr())
		state := tcpConn.ConnectionState()
		for _, v := range state.PeerCertificates {
			adatypes.Central.Log.Debugf("Client: Server public key is:")
			adatypes.Central.Log.Debugf("Remote Certificate Issuer: %v", v.Issuer.String())
			//			x, _ := x509.MarshalPKIXPublicKey(v.PublicKey)
			//			adatypes.Central.Log.Debugf("%s %v -> %v", v.Issuer.CommonName, x, pkerr)
		}
		adatypes.Central.Log.Debugf("client: handshake: %v", state.HandshakeComplete)
		adatypes.Central.Log.Debugf("Connect dial passed ...")

	}
	connection.connection = tcpConn
	return
}

// Disconnect disconnect remote TCP/IP Adabas nucleus
func (connection *AdaTCP) Disconnect() (err error) {
	if connection.connection == nil {
		return adatypes.NewGenericError(114)
	}

	adatypes.Central.Log.Debugf("Disconnect connection to %s", connection.URL.String())
	var buffer bytes.Buffer
	header := NewAdatcpHeader(DisconnectRequest)
	header.Identification = connection.adauuid
	payload := adatcpDisconnectPayload{}
	header.Length = uint32(AdaTCPHeaderLength + unsafe.Sizeof(payload))

	// Write structures to buffer
	err = binary.Write(&buffer, binary.BigEndian, header)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		return
	}
	err = binary.Write(&buffer, binary.BigEndian, payload)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		return
	}

	// Send the data to network
	_, err = connection.connection.Write(buffer.Bytes())
	if err != nil {
		return
	}
	rcvBuffer := make([]byte, buffer.Len())
	_, err = io.ReadFull(connection.connection, rcvBuffer)
	if err != nil {
		return
	}
	// Parse buffer from network into structure
	buf := bytes.NewBuffer(rcvBuffer)
	err = binary.Read(buf, connection.order, &header)
	if err != nil {
		return
	}
	if !header.Valid() {
		return adatypes.NewGenericError(160)
	}
	err = binary.Read(buf, connection.order, &payload)
	if err != nil {
		return
	}

	err = connection.connection.Close()
	connection.connection = nil
	if connection.stats != nil {
		connection.stats.remoteClosed++
	}

	return
}

// Generate error code specific error
func generateError(errorCode uint32) error {
	return adatypes.NewGenericError(91, errorCode)
}
