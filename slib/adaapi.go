package main

// #include <stdint.h>
import "C"

import (
	"fmt"
	"github.com/SoftwareAG/adabas-go-api/adabas"
	"sync/atomic"
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
	fmt.Println("Got connection")
	c := string(*conn)
	ac, err := adabas.NewConnection(c)
	if err != nil {
		fmt.Println("Error creating connection", err)
		return C.uint64_t(0)
	}
	id := atomic.AddUint64(&idindex, 1)
	connections[id] = &cConnection{connection: ac}
	return C.uint64_t(id)
}

//export ada_close_connection
func ada_close_connection(hdl C.uint64_t) {
	cConn := connections[uint64(hdl)]
	cConn.connection.Close()
	delete(connections, uint64(hdl))
}

//export ada_send_search
func ada_send_search(hdl C.uint64_t, file C.int, fields, search *C.char) C.int {
	cConn := connections[uint64(hdl)]
	readRequest, rrerr := cConn.connection.CreateReadRequest(uint32(file))
	if rrerr != nil {
		return 1
	}
	err := readRequest.QueryFields(string(*fields))
	if err != nil {
		return 1
	}
	result, rerr := readRequest.ReadLogicalWith(string(*search))
	if rerr != nil {
		return 1
	}
	cConn.result = result
	return 0
}

func main() {}
