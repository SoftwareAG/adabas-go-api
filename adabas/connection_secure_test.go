package adabas

import (
	"fmt"
)

func ExampleConnection_secureRead() {
	f, err := initLogWithFile("connection_secure.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	connection, cerr := NewConnection("acj;target=25")
	if cerr != nil {
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateFileReadRequest(11)
	if rerr != nil {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("AA")
	if err != nil {
		fmt.Println("Error query fields for request", err)
		return
	}
	request.Limit = 0
	fmt.Println("Read logigcal data:")
	var result *Response
	result, err = request.ReadLogicalWith("AA=[11100315:11100316]")
	if err != nil {
		fmt.Println("Error read logical data", err)
		return
	}
	result.DumpValues()
	// Output: XX
}
