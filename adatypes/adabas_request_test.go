package adatypes

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCaller struct {
	count       uint64
	secondCount uint64
}

func (caller *testCaller) CallAdabas() (err error) {
	caller.count++
	return nil
}

func (caller *testCaller) SendSecondCall(adabasRequest *Request, x interface{}) (err error) {
	caller.secondCount++
	return
}

func initDefinition() *Definition {
	groupLayout := []IAdaType{
		NewType(FieldTypeUnpacked, "GS", 8),
		NewType(FieldTypePacked, "GP", 8),
	}
	layout := []IAdaType{
		NewType(FieldTypeString, "A8", 8),
		NewType(FieldTypeUInt4, "U4"),
		NewType(FieldTypeByte, "B1"),
		NewType(FieldTypeUByte, "UB"),
		NewType(FieldTypeUInt2, "I2"),
		NewType(FieldTypeUInt8, "U8"),
		NewStructureList(FieldTypePeriodGroup, "PG", OccCapacity, groupLayout),
		NewType(FieldTypeUInt8, "I8"),
	}
	for _, x := range groupLayout {
		x.AddFlag(FlagOptionPE)
		x.SetLevel(2)
	}
	for _, x := range groupLayout {
		x.SetLevel(1)
	}

	testDefinition := NewDefinitionWithTypes(layout)
	testDefinition.InitReferences()
	return testDefinition
}

func testParser(adabasRequest *Request, x interface{}) (err error) {
	fmt.Printf("Test parser called %T\n", x)
	Central.Log.Debugf("Multifetch offset=%d RecordBuffer offset=%d", adabasRequest.MultifetchBuffer.offset, adabasRequest.RecordBuffer.offset)
	return
}

func TestAdabasRequestParser_withPeriod(t *testing.T) {
	lerr := initLogWithFile("adabas_request.log")
	if !assert.NoError(t, lerr) {
		return
	}

	testDefinition := initDefinition()
	err := testDefinition.ShouldRestrictToFields("A8,PG")
	if !assert.NoError(t, err) {
		return
	}
	testDefinition.CreateValues(false)
	adabasRequest, aerr := testDefinition.CreateAdabasRequest(false, 0, false)
	if !assert.NoError(t, aerr) || !assert.NotNil(t, adabasRequest) {
		return
	}
	testCaller := &testCaller{}
	adabasRequest.Definition = testDefinition
	adabasRequest.Caller = testCaller
	adabasRequest.Parser = testParser
	var multifetchData []byte
	var dataContent []byte
	if endian() == binary.LittleEndian {
		multifetchData = []byte{1, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7, 0, 0, 0}
		dataContent = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 1, 0, 0, 0, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x31, 0x32, 0, 0, 0, 0, 0, 0, 0x12, 0x1d}
	} else {
		multifetchData = []byte{0, 0, 0, 1, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7}
		dataContent = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 0, 0, 0, 1, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x31, 0x32, 0, 0, 0, 0, 0, 0, 0x12, 0x1d}
	}
	adabasRequest.RecordBuffer = NewHelper(dataContent, len(dataContent), endian())
	adabasRequest.MultifetchBuffer = NewHelper(multifetchData, len(multifetchData), endian())
	count := uint64(0)
	responseCode, perr := adabasRequest.ParseBuffer(&count, nil)
	assert.NoError(t, perr)
	assert.Equal(t, 0, int(responseCode))
	assert.Equal(t, uint64(1), count)
	assert.Equal(t, uint32(20), adabasRequest.MultifetchBuffer.offset)
	assert.Equal(t, uint32(28), adabasRequest.RecordBuffer.offset)
	Central.Log.Debugf("Test dump values")
	testDefinition.DumpValues(true)
	v, serr := testDefinition.SearchByIndex("GS", []uint32{1}, false)
	assert.NoError(t, serr)
	assert.NotNil(t, v)
	assert.Equal(t, "12", v.String())
	v, serr = testDefinition.SearchByIndex("GP", []uint32{1}, false)
	assert.NoError(t, serr)
	assert.NotNil(t, v)
	assert.Equal(t, "-121", v.String())
}

func TestAdabasRequestParser_osEmptyPeriod(t *testing.T) {
	lerr := initLogWithFile("adabas_request.log")
	if !assert.NoError(t, lerr) {
		return
	}

	testDefinition := initDefinition()
	err := testDefinition.ShouldRestrictToFields("A8,PG")
	if !assert.NoError(t, err) {
		return
	}
	testDefinition.CreateValues(false)
	adabasRequest, aerr := testDefinition.CreateAdabasRequest(false, 0, false)
	if !assert.NoError(t, aerr) || !assert.NotNil(t, adabasRequest) {
		return
	}
	testCaller := &testCaller{}
	adabasRequest.Definition = testDefinition
	adabasRequest.Caller = testCaller
	adabasRequest.Parser = testParser
	var multifetchData []byte
	var dataContent []byte
	if endian() == binary.LittleEndian {
		dataContent = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 0, 0, 0, 0}
		multifetchData = []byte{1, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7, 0, 0, 0}
	} else {
		dataContent = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 0, 0, 0, 0}
		multifetchData = []byte{0, 0, 0, 1, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7}
	}
	adabasRequest.RecordBuffer = NewHelper(dataContent, len(dataContent), endian())
	adabasRequest.MultifetchBuffer = NewHelper(multifetchData, len(multifetchData), endian())
	count := uint64(0)
	responseCode, perr := adabasRequest.ParseBuffer(&count, nil)
	assert.NoError(t, perr)
	assert.Equal(t, 0, int(responseCode))
	assert.Equal(t, uint64(1), count)
	assert.Equal(t, uint32(len(multifetchData)), adabasRequest.MultifetchBuffer.offset)
	assert.Equal(t, uint32(len(dataContent)), adabasRequest.RecordBuffer.offset)
	Central.Log.Debugf("Test dump values")
	testDefinition.DumpValues(true)
	v := testDefinition.Search("PG")
	assert.NotNil(t, v)
	if assert.IsType(t, &StructureValue{}, v) {
		sv := v.(*StructureValue)
		assert.Equal(t, 0, sv.NrElements())
	}
}

func TestAdabasRequestParser_mfEmptyPeriod(t *testing.T) {
	lerr := initLogWithFile("adabas_request.log")
	if !assert.NoError(t, lerr) {
		return
	}

	testDefinition := initDefinition()
	err := testDefinition.ShouldRestrictToFields("A8,PG")
	if !assert.NoError(t, err) {
		return
	}
	testDefinition.CreateValues(false)
	adabasRequest, aerr := testDefinition.CreateAdabasRequest(false, 0, true)
	if !assert.NoError(t, aerr) || !assert.NotNil(t, adabasRequest) {
		return
	}
	testCaller := &testCaller{}
	adabasRequest.Definition = testDefinition
	adabasRequest.Caller = testCaller
	adabasRequest.Parser = testParser
	dataContent := []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	adabasRequest.RecordBuffer = NewHelper(dataContent, len(dataContent), endian())
	var multifetchData []byte
	if endian() == binary.LittleEndian {
		multifetchData = []byte{1, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7, 0, 0, 0}
	} else {
		multifetchData = []byte{0, 0, 0, 1, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, 7}
	}
	adabasRequest.MultifetchBuffer = NewHelper(multifetchData, len(multifetchData), endian())
	count := uint64(0)
	responseCode, perr := adabasRequest.ParseBuffer(&count, nil)
	assert.NoError(t, perr)
	assert.Equal(t, 0, int(responseCode))
	assert.Equal(t, uint64(1), count)
	assert.Equal(t, uint32(len(multifetchData)), adabasRequest.MultifetchBuffer.offset)
	assert.Equal(t, uint32(len(dataContent)), adabasRequest.RecordBuffer.offset)
	Central.Log.Debugf("Test dump values")
	testDefinition.DumpValues(true)
	v := testDefinition.Search("PG")
	assert.NotNil(t, v)
	if assert.IsType(t, &StructureValue{}, v) {
		sv := v.(*StructureValue)
		assert.Equal(t, 0, sv.NrElements())
	}
}
