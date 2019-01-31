package main

// #include <stdint.h>
// #include <string.h>
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

//export ada_get_record_value
func ada_get_record_value(hdl C.uint64_t, index C.int, field, value *C.char) C.int {
	cConn := connections[uint64(hdl)]
	valueIndex := int(index) - 1
	v := cConn.result.Values[valueIndex].HashFields[C.GoString(field)]
	vs := []byte(v.String())
	C.memcpy(unsafe.Pointer(value), unsafe.Pointer(&vs[0]), C.ulong(len(vs)))
	return C.int(0)
}

//export ada_send_msearch
func ada_send_msearch(hdl C.uint64_t, mapName *C.char, fields, search *C.char) C.int {
	cConn := connections[uint64(hdl)]
	readRequest, rrerr := cConn.connection.CreateMapReadRequest(C.GoString(mapName))
	if rrerr != nil {
		return 1
	}
	err := readRequest.QueryFields(C.GoString(fields))
	if err != nil {
		return 1
	}
	result, rerr := readRequest.ReadLogicalWith(C.GoString(search))
	if rerr != nil {
		return 1
	}
	cConn.result = result
	return 0
}

func main() {}
