package adabas

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkConnection_noMultifetch(b *testing.B) {
	f, ferr := initLogWithFile("connection_bench.log")
	if ferr != nil {
		fmt.Println("Error creating log")
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", b.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(b, err) {
		return
	}
	if !assert.NotNil(b, connection) {
		return
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(11)
	assert.NoError(b, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 1

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(b, qErr)
	result := &Response{}
	err = readRequest.ReadPhysicalSequenceWithParser(nil, result)
	assert.NoError(b, err)
	assert.Equal(b, 1107, len(result.Values))
}

func BenchmarkConnection_Multifetch(b *testing.B) {
	f, ferr := initLogWithFile("connection_bench.log")
	if ferr != nil {
		fmt.Println("Error creating log")
		return
	}
	defer f.Close()

	log.Infof("TEST: %s", b.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(b, err) {
		return
	}
	if !assert.NotNil(b, connection) {
		return
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateReadRequest(11)
	assert.NoError(b, rErr)
	readRequest.Limit = 0
	readRequest.Multifetch = 10

	qErr := readRequest.QueryFields("AA,AB")
	assert.NoError(b, qErr)
	result := &Response{}
	err = readRequest.ReadPhysicalSequenceWithParser(nil, result)
	assert.NoError(b, err)
	assert.Equal(b, 1107, len(result.Values))
}
