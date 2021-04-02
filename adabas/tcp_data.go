/*
* Copyright © 2021 Software AG, Darmstadt, Germany and/or its licensors
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
	"io"
	"strings"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

const adatcpDataHeaderEyecatcher = "DATA"

const adatcpDataHeaderVersion = "0001"

// TransferDataType transfer data type used to check buffer
type TransferDataType uint32

const (
	adabasRequest       = TransferDataType(1)
	adabasReply         = TransferDataType(2)
	clusterNodesRequest = TransferDataType(3)
	clusterNodesReply   = TransferDataType(4)
	clusterNodesError   = TransferDataType(5)
)

// AdaTCPDataHeader Adabas TCP header
type AdaTCPDataHeader struct {
	Eyecatcher      [4]byte
	Version         [4]byte
	Length          uint32
	DataType        TransferDataType
	NumberOfBuffers uint32
	ErrorCode       uint32
}

const defaultNodeListLength = (64 * 1024)

// AdaTCPDataHeaderLength length of AdaTCPDataHeader structure
const AdaTCPDataHeaderLength = 24

func newAdatcpDataHeader(dataType TransferDataType) AdaTCPDataHeader {
	header := AdaTCPDataHeader{DataType: dataType}
	copy(header.Eyecatcher[:], adatcpDataHeaderEyecatcher)
	copy(header.Version[:], adatcpDataHeaderVersion)
	return header
}

func (connection *AdaTCP) receiveNodeList() (err error) {
	defer TimeTrack(time.Now(), "ADATCP request node list", nil)
	header := NewAdatcpHeader(DataRequest)
	dataHeader := newAdatcpDataHeader(clusterNodesRequest)
	dataHeader.NumberOfBuffers = 0
	header.Identification = connection.adauuid
	header.Length = uint32(AdaTCPHeaderLength + AdaTCPDataHeaderLength)
	dataHeader.Length = uint32(AdaTCPDataHeaderLength + defaultNodeListLength)
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
	send := headerBuffer.Bytes()
	//	send = append(send, make([]byte, defaultNodeListLength)...)
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("SND:", send, len(send), 8, 16, true))
	}
	adatypes.Central.Log.Debugf("Write TCP data of length=%d capacity=%d netto bytes send=%d", headerBuffer.Len(), headerBuffer.Cap(), len(send))
	_, err = connection.connection.Write(send)
	if err != nil {
		adatypes.Central.Log.Infof("Send data TCP node list request error: %v", err)
		return
	}
	if connection.stats != nil {
		connection.stats.remoteSend++
	}

	buffer := &bytes.Buffer{}
	_, err = connection.ReceiveData(buffer, clusterNodesReply)
	if err != nil {
		adatypes.Central.Log.Infof("Error TCP reading node list: %v", err)
		connection.connection.Close()
		connection.connection = nil
		return
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("Received node list:", buffer.Bytes(), buffer.Len(), 8, 16, true))
	}
	clusterNodes := strings.Split(buffer.String(), ";")
	for _, n := range clusterNodes {
		u, err := NewURL(n)
		if err != nil {
			return err
		}
		connection.clusterNodes = append(connection.clusterNodes, u)
	}
	if len(connection.clusterNodes) == 0 {
		return adatypes.NewGenericError(163)
	}
	if *connection.URL == *connection.clusterNodes[0] {
		adatypes.Central.Log.Infof("Connection to master already available")
	}
	return
}

// SendData send data to remote TCP/IP Adabas nucleus
func (connection *AdaTCP) SendData(buffer bytes.Buffer, nrAbdBuffers uint32) (err error) {
	defer TimeTrack(time.Now(), "ADATCP Send data", nil)
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
		adatypes.LogMultiLineString(adatypes.FormatBytes("SND:", send, len(send), 8, 16, true))
	}
	var n int
	adatypes.Central.Log.Debugf("Write TCP data of length=%d capacity=%d netto bytes send=%d", headerBuffer.Len(), headerBuffer.Cap(), len(send))
	n, err = connection.connection.Write(send)
	if err != nil {
		adatypes.Central.Log.Infof("Send data TCP data error: %v", err)
		return
	}
	if connection.stats != nil {
		connection.stats.remoteSend++
	}
	adatypes.Central.Log.Debugf("Send data completed buffer send=%d really send=%d", buffer.Len(), n)
	return
}

// ReceiveData receive data from remote TCP/IP Adabas nucleus
func (connection *AdaTCP) ReceiveData(buffer *bytes.Buffer, expected TransferDataType) (nrAbdBuffers uint32, err error) {
	defer TimeTrack(time.Now(), "ADATCP Receive data", nil)
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
		adatypes.Central.Log.Infof("Receive TCP data error: %v", err)
		return
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Receive got header .... size=%d/%d", n, len(rcvHeaderBuffer))
		adatypes.LogMultiLineString(adatypes.FormatBytes("RCV Header BUFFER:", rcvHeaderBuffer, len(rcvHeaderBuffer), 8, 16, true))
	}
	if n < hl {
		return 0, adatypes.NewGenericError(92)
	}
	headerBuffer := bytes.NewBuffer(rcvHeaderBuffer)
	err = binary.Read(headerBuffer, binary.BigEndian, &header)
	if err != nil {
		adatypes.Central.Log.Infof("Read TCP header error: %v", err)
		return
	}

	if header.BufferType != DataReply {
		return 0, adatypes.NewGenericError(164)
	}

	//header.Length = header.Length
	adatypes.Central.Log.Debugf("Receive got header length .... size=%d error=%d", header.Length, header.ErrorCode)
	err = binary.Read(headerBuffer, Endian(), &dataHeader)
	if err != nil {
		adatypes.Central.Log.Infof("Read TCP data header error: %v", err)
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
	if dataHeader.DataType != expected {
		return 0, adatypes.NewGenericError(162)
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
		// n, err = io.ReadFull(connection.connection, dataBytes)
		// if err != nil {
		// 	return
		// }
		n, err = connection.connection.Read(dataBytes)
		if err != nil {
			return
		}
		if n != len(dataBytes) {
			b := make([]byte, len(dataBytes)-n)
			for n != len(dataBytes) {
				n1, nerr := connection.connection.Read(b)
				if nerr != nil {
					return
				}
				copy(dataBytes[n:], b[:n1])
				n += n1
			}
		}
		adatypes.Central.Log.Debugf("Extra read receive %d number of bytes", n)
		buffer.Write(dataBytes)
		adatypes.Central.Log.Debugf("Current size of buffer=%d", buffer.Len())
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.LogMultiLineString(adatypes.FormatBytes("RCV DATA BUFFER:", buffer.Bytes(), buffer.Len(), 8, 16, false))
	}
	if connection.stats != nil {
		connection.stats.remoteReceive++
	}

	return
}
