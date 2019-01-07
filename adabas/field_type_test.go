package adabas

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFieldTypeStore(t *testing.T) {
	f := initTestLogWithFile(t, "field_type.log")
	defer f.Close()

	// cErr := clearFile(16)
	// if !assert.NoError(t, cErr) {
	// 	return
	// }

	storeRequest := NewStoreRequest("23", 270)
	defer storeRequest.Close()
	err := storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr := storeRequest.CreateRecord()
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRecord.SetValue("S1", "-1")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U1", "1")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S2", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U2", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S4", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U4", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S8", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U8", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("BR", []byte{0x0, 0x10, 0x20})
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("B1", []byte{0xff, 0x10, 0x5, 0x0, 0x10, 0x20})
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("A1", "X")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AS", "NORMALSTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("A2", "LARGESTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AB", "LOBSTRING")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("AF", "FIELD-TYPE-TEST")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord.DumpValues()
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRequest.EndTransaction()
}

func TestFieldType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection.log")
	defer f.Close()

	log.Debug("TEST: ", t.Name())
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	assert.NoError(t, openErr)
	request, err := connection.CreateReadRequest(270)
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, cerr) {
		return
	}
	request.Limit = 0
	request.RecordBufferShift = 64000
	result, rerr := request.ReadLogicalWith("AF=FIELD-TYPE-TEST")
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 4, len(result.Values))
		assert.Equal(t, 4, result.NrRecords())
		// err = result.DumpValues()
		// assert.NoError(t, err)
		kaVal := result.Values[0].HashFields["AA"]
		kaVal = result.Values[3].HashFields["KA"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "ಸೆನಿಓರ್ ಪ್ರೋಗ್ೃಾಮ್ಮೇರ್  ", kaVal.String())
		}
	}
}
