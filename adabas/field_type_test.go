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
	err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	//	err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP")
	//err := storeRequest.StoreFields("*")
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
	x1, _ := storeRecord.searchValue("S1")
	if !assert.Equal(t, "-1", x1.String()) {
		return
	}
	err = storeRecord.SetValue("U1", "1")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S2", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	x2, _ := storeRecord.searchValue("S2")
	if !assert.Equal(t, "-1000", x2.String()) {
		return
	}
	err = storeRecord.SetValue("U2", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S4", "-100000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U4", "1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("S8", "-1000")
	if !assert.NoError(t, err) {
		return
	}
	err = storeRecord.SetValue("U8", "1000")
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

func TestFieldTypeRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "field_type.log")
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
	err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1")
	//	err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	//err = request.QueryFields("*")
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
		assert.Equal(t, 1, len(result.Values))
		assert.Equal(t, 1, result.NrRecords())
		err = result.DumpValues()
		assert.NoError(t, err)
		kaVal := result.Values[0].HashFields["S1"]
		assert.Equal(t, "-1", kaVal.String())
		kaVal = result.Values[0].HashFields["U1"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "1", kaVal.String())
		}
		kaVal = result.Values[0].HashFields["S2"]
		assert.Equal(t, "-100000", kaVal.String())
		kaVal = result.Values[0].HashFields["S4"]
		assert.Equal(t, "-1000", kaVal.String())
		kaVal = result.Values[0].HashFields["A1"]
		assert.Equal(t, "X", kaVal.String())
	}
}
