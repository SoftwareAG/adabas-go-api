package adabas

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConnectionCursing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_map.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println("Connection : ", connection)
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if assert.NoError(t, rerr) {
		fmt.Println("Limit query data:")
		request.QueryFields("NAME,PERSONNEL-ID")
		request.Limit = 0
		fmt.Println("Init cursor data...")
		col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100110:11100304]")
		if !assert.NoError(t, cerr) {
			return
		}
		fmt.Println("Read next cursor record...")
		for col.HasNextRecord() {
			record, rerr := col.NextRecord()
			assert.NoError(t, rerr)
			assert.NotNil(t, record)
			fmt.Println("Record received:")
			record.DumpValues()

		}
	}
}

func ExampleCursoring() {
	f, ferr := initLogWithFile("connection_map.log")
	if ferr != nil {
		fmt.Println("Error initializing log", ferr)
		return
	}
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error creating new connection", cerr)
		return
	}
	defer connection.Close()
	request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if rerr != nil {
		fmt.Println("Error creating read request", cerr)
		return
	}
	// Define fields to be part of the request
	request.QueryFields("NAME,PERSONNEL-ID")
	// Define chunks of cursoring requests
	request.Limit = 5

	// Init cursoring using search
	col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100110:11100120]")
	if cerr != nil {
		fmt.Println("Error init cursoring", cerr)
		return
	}
	for col.HasNextRecord() {
		record, rerr := col.NextRecord()
		if rerr != nil {
			fmt.Println("Error getting next record", rerr)
			return
		}
		fmt.Printf("New record received: ISN=%d\n", record.Isn)
		record.DumpValues()
	}

	// Output: New record received: ISN=210
	// Dump all record values
	//   PERSONNEL-ID = > 11100110 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > BUNGERT              <
	// New record received: ISN=211
	// Dump all record values
	//   PERSONNEL-ID = > 11100111 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THIELE               <
	// New record received: ISN=212
	// Dump all record values
	//   PERSONNEL-ID = > 11100112 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > THOMA                <
	// New record received: ISN=213
	// Dump all record values
	//   PERSONNEL-ID = > 11100113 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > TREIBER              <
	// New record received: ISN=214
	// Dump all record values
	//   PERSONNEL-ID = > 11100114 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > UNGER                <
	// New record received: ISN=1102
	// Dump all record values
	//   PERSONNEL-ID = > 11100115 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > VETTER               <
	// New record received: ISN=215
	// Dump all record values
	//   PERSONNEL-ID = > 11100116 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > VOGEL                <
	// New record received: ISN=216
	// Dump all record values
	//   PERSONNEL-ID = > 11100117 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > WABER                <
	// New record received: ISN=217
	// Dump all record values
	//   PERSONNEL-ID = > 11100118 <
	//   FULL-NAME = [ 1 ]
	//    NAME = > WAGNER               <

}
