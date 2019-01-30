package adabas

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
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
		col, cerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100000:11101000]")
		assert.NoError(t, cerr)
		fmt.Println("Read next cursor record...")
		record, rerr := col.NextRecord()
		assert.NoError(t, rerr)
		assert.NotNil(t, record)
		fmt.Println("Record received:")
		record.DumpValues()
	}
}
