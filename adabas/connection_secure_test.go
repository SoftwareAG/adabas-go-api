package adabas

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConnectionSecure_fail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_secure.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=25;auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}

	request, rerr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("AA")
	assert.Error(t, err)
	assert.Equal(t, "ADAGEC801F: Security violation: Authentication error (rsp=200,subrsp=31,dbid=25,file=0)", err.Error())
}

func TestConnectionSecure_pwd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_secure.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=25;auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	connection.AddCredential("hkaf", "dummy1")

	request, rerr := connection.CreateFileReadRequest(11)
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err = request.QueryFields("AA,AE")
	assert.NoError(t, err)
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
