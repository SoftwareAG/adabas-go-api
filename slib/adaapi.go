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

package main

// #include <stdlib.h>
// #include <stdint.h>
// #include <string.h>
// #ifndef NULL
// #define NULL (void*)(0)
// #endif
import "C"

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/SoftwareAG/adabas-go-api/adabas"
)

type cConnection struct {
	connection *adabas.Connection
	result     *adabas.Response
}

var connections map[uint64]*cConnection
var idindex uint64

func init() {
	connections = make(map[uint64]*cConnection)
}

//export ada_free
func ada_free(p unsafe.Pointer) {
	C.free(p)
}

//export ada_new_connection
func ada_new_connection(conn *C.char) C.uint64_t {
	c := C.GoString(conn)
	fmt.Println("Got connection", c)
	ac, err := adabas.NewConnection(c)
	if err != nil {
		fmt.Println("Error creating connection", err)
		return C.uint64_t(0)
	}
	id := atomic.AddUint64(&idindex, 1)
	connections[id] = &cConnection{connection: ac}
	fmt.Println("New handle", id)
	return C.uint64_t(id)
}

//export ada_close_connection
func ada_close_connection(hdl C.uint64_t) {
	fmt.Println("Close", hdl)
	if cConn, ok := connections[uint64(hdl)]; ok {
		cConn.connection.Close()
		delete(connections, uint64(hdl))
	}
}

//export ada_send_search
func ada_send_search(hdl C.uint64_t, file C.int, fields, search *C.char) C.int {
	cConn := connections[uint64(hdl)]
	readRequest, rrerr := cConn.connection.CreateReadRequest(uint32(file))
	if rrerr != nil {
		fmt.Println("Error creating request", rrerr)
		return 1
	}
	readRequest.Limit = 0
	err := readRequest.QueryFields(C.GoString(fields))
	if err != nil {
		fmt.Println("Error query fields", err)
		return 1
	}
	result, rerr := readRequest.ReadLogicalWith(C.GoString(search))
	if rerr != nil {
		fmt.Println("Error read logical with", rerr)
		return 0
	}
	cConn.result = result
	return C.int(len(result.Values))
}

//export ada_get_record_string_value
func ada_get_record_string_value(hdl C.uint64_t, index C.int, field, value *C.char, strlen C.int) C.int {
	cConn := connections[uint64(hdl)]
	valueIndex := int(index) - 1
	v := cConn.result.Values[valueIndex].HashFields[C.GoString(field)]
	if v == nil {
		fmt.Println("Field not found:", C.GoString(field))
		return C.int(1)
	}
	vs := []byte(v.String())
	C.memcpy(unsafe.Pointer(value), unsafe.Pointer(&vs[0]), C.size_t(len(vs)))
	return C.int(0)
}

//export ada_get_fieldnames
func ada_get_fieldnames(hdl C.uint64_t) **C.char {
	cConn := connections[uint64(hdl)]

	fieldnames := cConn.result.Definition.Fieldnames()
	cArray := C.malloc(C.size_t(len(fieldnames)+1) * C.size_t(unsafe.Sizeof(uintptr(0))))

	// convert the C array to a Go Array so we can index it
	a := (*[1<<30 - 1]*C.char)(cArray)

	for idx, substring := range fieldnames {
		a[idx] = C.CString(substring)
	}
	a[len(fieldnames)] = nil

	return (**C.char)(cArray)
}

//export ada_get_record_int64_value
func ada_get_record_int64_value(hdl C.uint64_t, index C.int, field *C.char, value *C.int64_t) C.int {
	cConn := connections[uint64(hdl)]
	valueIndex := int(index) - 1
	v := cConn.result.Values[valueIndex].HashFields[C.GoString(field)]
	vi, err := v.Int64()
	if err != nil {
		return C.int(1)
	}
	*value = C.int64_t(vi)
	return C.int(0)
}

//export ada_get_record_byte_array_value
func ada_get_record_byte_array_value(hdl C.uint64_t, index C.int, field, value *C.char, blen C.int) C.int {
	cConn := connections[uint64(hdl)]
	valueIndex := int(index) - 1
	v := cConn.result.Values[valueIndex].HashFields[C.GoString(field)]
	vi := v.Bytes()
	C.memcpy(unsafe.Pointer(value), unsafe.Pointer(&vi[0]), C.size_t(len(vi)))
	return C.int(0)
}

//export ada_send_msearch
func ada_send_msearch(hdl C.uint64_t, mapName *C.char, fields, search *C.char) C.int {
	cConn := connections[uint64(hdl)]
	readRequest, rrerr := cConn.connection.CreateMapReadRequest(C.GoString(mapName))
	if rrerr != nil {
		fmt.Println("Error create map request", rrerr)
		return 0
	}
	err := readRequest.QueryFields(C.GoString(fields))
	if err != nil {
		fmt.Println("Error query fields", err)
		return 0
	}
	result, rerr := readRequest.ReadLogicalWith(C.GoString(search))
	if rerr != nil {
		fmt.Println("Error read logical with", rerr)
		return 0
	}
	cConn.result = result
	return C.int(len(result.Values))
}

func main() {}
