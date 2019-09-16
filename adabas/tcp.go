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

// AdaTCPHeader Adabas TCP Header ADATCP
type AdaTCPHeader struct {
	Eyecatcher     [6]byte
	Version        [2]byte
	Length         uint32
	BufferType     BufferType
	Identification adaUUID
	ErrorCode      uint32
	Reserved       uint32
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
}

const adatcpDataHeaderEyecatcher = "DATA"

const adatcpDataHeaderVersion = "0001"

const (
	adabasRequest = uint32(1)

//	adabasReply   = uint32(2)
)

// AdaTCPDataHeader Adabas TCP header
type AdaTCPDataHeader struct {
	Eyecatcher      [4]byte
	Version         [4]byte
	Length          uint32
	DataType        uint32
	NumberOfBuffers uint32
	ErrorCode       uint32
}

// AdaTCPDataHeaderLength length of AdaTCPDataHeader structure
const AdaTCPDataHeaderLength = 24

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

func newAdatcpDataHeader(dataType uint32) AdaTCPDataHeader {
	header := AdaTCPDataHeader{DataType: dataType}
	copy(header.Eyecatcher[:], adatcpDataHeaderEyecatcher)
	copy(header.Version[:], adatcpDataHeaderVersion)
	return header
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
	pair := URL.searchCertificate()
	t := &AdaTCP{URL: URL, order: order, pair: pair,
		id: adaTCPID{pid: pid, timestamp: timestamp}}
	copy(t.id.user[:], user[:])
	copy(t.id.node[:], node[:])
	return t
}

// Connect establish connection to ADATCP server
func (connection *AdaTCP) Connect() (err error) {
	url := fmt.Sprintf("%s:%d", connection.URL.Host, connection.URL.Port)
	adatypes.Central.Log.Debugf("Open TCP connection to %s", url)
	addr, _ := net.ResolveTCPAddr("tcp", url)
	switch connection.URL.Driver {
	case "adatcp":
		tcpConn, tcpErr := net.DialTCP("tcp", nil, addr)
		err = tcpErr
		if err != nil {
			adatypes.Central.Log.Debugf("Connect error : %v", err)
			return
		}
		adatypes.Central.Log.Debugf("Connect dial passed ...")
		connection.connection = tcpConn
		tcpConn.SetNoDelay(true)
	case "adatcps":
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
			adatypes.Central.Log.Debugf("client: mutual: %v", state.NegotiatedProtocolIsMutual)

			adatypes.Central.Log.Debugf("Connect dial passed ...")

		}
		connection.connection = tcpConn
	default:
		return adatypes.NewGenericError(131)
	}
	var buffer bytes.Buffer
	header := NewAdatcpHeader(ConnectRequest)
	payload := AdaTCPConnectPayload{Charset: adatcpASCII8, Floatingpoint: adatcpFloatIEEE}
	copy(payload.Userid[:], connection.id.user[:])
	copy(payload.Nodeid[:], connection.id.node[:])
	payload.ProcessID = connection.id.pid
	payload.TimeStamp = connection.id.timestamp

	header.Length = uint32(AdaTCPHeaderLength + unsafe.Sizeof(payload))
	err = binary.Write(&buffer, binary.BigEndian, header)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		return
	}
	if bigEndian() {
		adatypes.Central.Log.Debugf("Write TCP payload for big endian")
		payload.Endianness = adatcpBigEndian
	} else {
		adatypes.Central.Log.Debugf("Write TCP payload for little endian")
		payload.Endianness = adatcpLittleEndian
	}
	adatypes.Central.Log.Debugf("Buffer size after header=%d", buffer.Len())

	// Send payload in big endian needed until remote knows the endianess of the client
	err = binary.Write(&buffer, binary.BigEndian, payload)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP connect payload in buffer error %s", err)
		return
	}
	adatypes.Central.Log.Debugf("Buffer size after payload=%d", buffer.Len())

	send := buffer.Bytes()
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("PAYLOAD:", send, len(send), len(send), 8, true))
	}
	_, err = connection.connection.Write(send)
	if err != nil {
		adatypes.Central.Log.Debugf("Error writing data %s", err)
		return
	}
	rcvBuffer := make([]byte, buffer.Len())
	_, err = io.ReadFull(connection.connection, rcvBuffer)
	//	_, err = connection.connection.Read(rcvBuffer)
	if err != nil {
		adatypes.Central.Log.Debugf("Error reading data %v", err)
		return
	}

	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("RCV PAYLOAD:", rcvBuffer, len(rcvBuffer), len(rcvBuffer), 8, true))
	}

	buf := bytes.NewBuffer(rcvBuffer)
	err = binary.Read(buf, binary.BigEndian, &header)
	if err != nil {
		adatypes.Central.Log.Debugf("Error parsing header %v", err)
		return
	}

	err = binary.Read(buf, binary.BigEndian, &payload)
	if err != nil {
		adatypes.Central.Log.Debugf("Error parsing payload %v", err)
		return
	}

	connection.adauuid = header.Identification
	connection.serverEndianness = payload.Endianness
	connection.serverCharset = payload.Charset
	connection.serverFloatingpoint = payload.Floatingpoint
	connection.databaseVersion = payload.DatabaseVersion
	connection.databaseName = payload.DatabaseName
	connection.databaseID = payload.DatabaseID

	return
}

// Disconnect disconnect remote TCP/IP Adabas nucleus
func (connection *AdaTCP) Disconnect() (err error) {
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
	err = binary.Read(buf, connection.order, &payload)
	if err != nil {
		return
	}

	err = connection.connection.Close()

	return
}

// SendData send data to remote TCP/IP Adabas nucleus
func (connection *AdaTCP) SendData(buffer bytes.Buffer, nrAbdBuffers uint32) (err error) {
	header := NewAdatcpHeader(DataRequest)
	dataHeader := newAdatcpDataHeader(adabasRequest)
	dataHeader.NumberOfBuffers = nrAbdBuffers
	header.Identification = connection.adauuid
	header.Length = uint32(AdaTCPHeaderLength + AdaTCPDataHeaderLength + buffer.Len())
	dataHeader.Length = uint32(AdaTCPDataHeaderLength + buffer.Len())
	var headerBuffer bytes.Buffer
	err = binary.Write(&headerBuffer, binary.BigEndian, header)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		return
	}
	err = binary.Write(&headerBuffer, Endian(), dataHeader)
	if err != nil {
		adatypes.Central.Log.Debugf("Write TCP header in buffer error %s", err)
		return
	}
	headerBuffer.Write(buffer.Bytes())
	send := headerBuffer.Bytes()
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("SND:", send, len(send), len(send), 8, true))
	}
	var n int
	adatypes.Central.Log.Debugf("Write TCP data of length=%d capacity=%d netto bytes send=%d", headerBuffer.Len(), headerBuffer.Cap(), len(send))
	n, err = connection.connection.Write(send)
	if err != nil {
		return
	}

	adatypes.Central.Log.Debugf("Send data completed buffer send=%d really send=%d", buffer.Len(), n)
	return
}

// Generate error code specific error
func generateError(errorCode uint32) error {
	return adatypes.NewGenericError(91, errorCode)
}

// ReceiveData receive data from remote TCP/IP Adabas nucleus
func (connection *AdaTCP) ReceiveData(buffer *bytes.Buffer) (nrAbdBuffers uint32, err error) {
	adatypes.Central.Log.Debugf("Receive data .... size=%d", buffer.Len())

	header := NewAdatcpHeader(DataReply)
	dataHeader := newAdatcpDataHeader(adabasRequest)
	header.Identification = connection.adauuid
	headerLength := uint32(AdaTCPHeaderLength)
	dataHeaderLength := uint32(AdaTCPDataHeaderLength)

	hl := int(headerLength + dataHeaderLength)
	rcvHeaderBuffer := make([]byte, headerLength+dataHeaderLength)
	var n int
	//	n, err = io.ReadFull(connection.connection, rcvHeaderBuffer)
	n, err = io.ReadAtLeast(connection.connection, rcvHeaderBuffer, hl)
	if err != nil {
		return
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Receive got header .... size=%d/%d", n, len(rcvHeaderBuffer))
		adatypes.LogMultiLineString(adatypes.FormatBytes("RCV Header BUFFER:", rcvHeaderBuffer, len(rcvHeaderBuffer), len(rcvHeaderBuffer), 8, true))
	}
	if n < hl {
		return 0, adatypes.NewGenericError(92)
	}
	headerBuffer := bytes.NewBuffer(rcvHeaderBuffer)
	err = binary.Read(headerBuffer, binary.BigEndian, &header)
	if err != nil {
		return
	}

	//header.Length = header.Length
	adatypes.Central.Log.Debugf("Receive got header length .... size=%d error=%d", header.Length, header.ErrorCode)
	err = binary.Read(headerBuffer, Endian(), &dataHeader)
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Receive got data length .... size=%d nrBuffer=%d", dataHeader.Length, dataHeader.NumberOfBuffers)
	nrAbdBuffers = dataHeader.NumberOfBuffers
	if header.Length == headerLength+dataHeaderLength {
		return 0, generateError(header.ErrorCode)
	}
	if header.Length < headerLength+dataHeaderLength {
		return 0, adatypes.NewGenericError(90, header.Length)
	}
	adatypes.Central.Log.Debugf("Current size of buffer=%d", buffer.Len())
	adatypes.Central.Log.Debugf("Receive %d number of bytes of %d", n, header.Length)
	_, err = buffer.Write(rcvHeaderBuffer[hl:])
	if err != nil {
		return
	}
	adatypes.Central.Log.Debugf("Received header size of buffer=%d", buffer.Len())
	if header.Length > uint32(n) {
		dataBytes := make([]byte, int(header.Length)-n)
		adatypes.Central.Log.Debugf("Create buffer of size %d to read rest of missingdata", len(dataBytes))
		n, err = io.ReadFull(connection.connection, dataBytes)
		// _, err = connection.connection.Read(dataBytes)
		if err != nil {
			return
		}
		adatypes.Central.Log.Debugf("Extra read receive %d number of bytes", n)
		buffer.Write(dataBytes)
		adatypes.Central.Log.Debugf("Current size of buffer=%d", buffer.Len())
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("RCV DATA BUFFER:", buffer.Bytes(), buffer.Len(), buffer.Len(), 8, false))
	}

	return
}
