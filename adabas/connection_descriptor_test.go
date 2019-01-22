package adabas

import (
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConnectionComplexSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=23;auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(16)
	assert.NoError(t, rErr)
	readRequest.QueryFields("AA,AB")

	adatypes.Central.Log.Debugf("Test Search complex with ...")
	result, rerr := readRequest.ReadLogicalWith("AA=[11100301:11100305] AND AE='SMITH'")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Complex search done")
	fmt.Println(result)
}

func TestConnectionSuperDescriptor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=24")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(11)
	assert.NoError(t, rErr)
	readRequest.QueryFields("AU,AV")

	adatypes.Central.Log.Debugf("Test Search complex with ...")
	result, rerr := readRequest.ReadLogicalBy("S1")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Super Descriptor read done")
	fmt.Println(result.String())
}
