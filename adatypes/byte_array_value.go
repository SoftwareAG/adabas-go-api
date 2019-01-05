/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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

package adatypes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type byteArrayValue struct {
	adaValue
	value []byte
}

func newByteArrayValue(initType IAdaType) *byteArrayValue {
	value := byteArrayValue{adaValue: adaValue{adatype: initType}}
	value.value = make([]byte, initType.Length())
	return &value
}

func (value *byteArrayValue) ByteValue() byte {
	if len(value.value) > 0 {
		return byte(value.value[0])
	}
	return 0
}

func (value *byteArrayValue) String() string {
	return strconv.Itoa(int(value.ByteValue()))
}

func (value *byteArrayValue) Value() interface{} {
	return value.value
}

func (value *byteArrayValue) Bytes() []byte {
	return value.value
}

func (value *byteArrayValue) SetStringValue(stValue string) {
	if strings.HasPrefix(stValue, "0x") {
		decoded, err := hex.DecodeString(stValue[2:])
		if err == nil {
			value.value = decoded
		}
	} else {
		iv, err := strconv.ParseUint(stValue, 10, 64)
		if err == nil {
			binary.LittleEndian.PutUint64(value.value, iv)
		}
	}
}

func (value *byteArrayValue) SetValue(v interface{}) error {
	switch v.(type) {
	case []byte:
		copy(value.value, v.([]byte))
		return nil
	case string:
		value.value = []byte(v.(string))
		return nil
	}
	return fmt.Errorf("Value interface not supported %T", v)
}

func (value *byteArrayValue) FormatBuffer(buffer *bytes.Buffer, option *BufferOption) uint32 {
	len := value.commonFormatBuffer(buffer, option)
	if len == 0 {
		len = 126
	}
	return len
}

func (value *byteArrayValue) StoreBuffer(helper *BufferHelper) error {
	return helper.putBytes(value.value)
}

func (value *byteArrayValue) parseBuffer(helper *BufferHelper, option *BufferOption) (res TraverseResult, err error) {
	len := uint8(value.Type().Length())
	if len == 0 {
		len, err = helper.ReceiveUInt8()
		if err != nil {
			return
		}
	}
	value.value, err = helper.ReceiveBytes(uint32(len))
	Central.Log.Debugf("Buffer get bytes offset=%v %s", helper.offset, value.Type().String())
	return
}
