package adabas

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const recordNamePrefix = "FIELD-TYPE-TEST"

func TestFieldTypeStore(t *testing.T) {
	f := initTestLogWithFile(t, "field_type.log")
	defer f.Close()

	cErr := clearFile(270)
	if !assert.NoError(t, cErr) {
		return
	}

	storeRequest := NewStoreRequest("23", 270)
	defer storeRequest.Close()
	//err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	//	err := storeRequest.StoreFields("S1,U1,S2,U2,S4,U4,S8,U8,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP")
	err := storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr := storeRequest.CreateRecord()
	if !assert.NoError(t, serr) {
		return
	}
	err = storeRecord.SetValue("AF", recordNamePrefix)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.Store(storeRecord)
	if !assert.NoError(t, err) {
		return
	}
	storeRecord, serr = storeRequest.CreateRecord()
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
	err = storeRecord.SetValue("AF", recordNamePrefix)
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
	//err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1")
	//err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1,AS,A2,AB,AF,WU,WL,W4,WF,PA,PF,UP,UF,UE")
	err = request.QueryFields("*")
	if !assert.NoError(t, cerr) {
		return
	}
	request.Limit = 0
	request.RecordBufferShift = 64000
	result, rerr := request.ReadLogicalWith("AF=" + recordNamePrefix)
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.NotNil(t, result) {
		assert.Equal(t, 2, len(result.Values))
		assert.Equal(t, 2, result.NrRecords())
		err = result.DumpValues()
		assert.NoError(t, err)
		kaVal := result.Values[1].HashFields["S1"]
		assert.Equal(t, "-1", kaVal.String())
		kaVal = result.Values[1].HashFields["U1"]
		if assert.NotNil(t, kaVal) {
			assert.Equal(t, "1", kaVal.String())
		}
		kaVal = result.Values[1].HashFields["S2"]
		assert.Equal(t, "-1000", kaVal.String())
		kaVal = result.Values[1].HashFields["S4"]
		assert.Equal(t, "-100000", kaVal.String())
		kaVal = result.Values[1].HashFields["A1"]
		assert.Equal(t, "X", kaVal.String())
	}
}

func dumpFieldTypeValues(adaValue adatypes.IAdaValue, x interface{}) (adatypes.TraverseResult, error) {
	if adaValue == nil {
		record := x.(*ResultRecord)
		if record == nil {
			return adatypes.EndTraverser, adatypes.NewGenericError(25)
		}
		fmt.Printf("Record found:\n")
	} else {

		y := strings.Repeat(" ", int(adaValue.Type().Level()))

		if x == nil {
			brackets := ""
			switch {
			case adaValue.PeriodIndex() > 0 && adaValue.MultipleIndex() > 0:
				brackets = fmt.Sprintf("[%02d,%02d]", adaValue.PeriodIndex(), adaValue.MultipleIndex())
			case adaValue.PeriodIndex() > 0:
				brackets = fmt.Sprintf("[%02d]", adaValue.PeriodIndex())
			case adaValue.MultipleIndex() > 0:
				brackets = fmt.Sprintf("[%02d]", adaValue.MultipleIndex())
			default:
			}

			if adaValue.Type().IsStructure() {
				structureValue := adaValue.(*adatypes.StructureValue)
				fmt.Println(y+" "+adaValue.Type().Name()+brackets+" = [", structureValue.NrElements(), "]")
			} else {
				fmt.Printf("%s %s%s = > %s <\n", y, adaValue.Type().Name(), brackets, adaValue.String())
			}
		} else {
			buffer := x.(*bytes.Buffer)
			buffer.WriteString(fmt.Sprintln(y, adaValue.Type().Name(), "= >", adaValue.String(), "<"))
		}
	}
	return adatypes.Continue, nil
}

func ExampleFieldType() {
	f, err := initLogWithFile("field_type.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	url := "23"
	fmt.Println("Connect to ", url)
	connection, cerr := NewConnection("acj;target=" + url)
	if cerr != nil {
		fmt.Println("Error creating database connection", cerr)
		return
	}
	defer connection.Close()
	fmt.Println(connection)
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error opening database", openErr)
		return
	}
	request, err := connection.CreateReadRequest(270)
	if err != nil {
		fmt.Println("Error creating read request", err)
		return
	}
	//err = request.QueryFields("S1,U1,S2,U2,S4,U4,S8,U8,AF,BR,B1,F4,F8,A1")
	err = request.QueryFields("IT,BB,TY,AA,WC,PI,UI")
	//err = request.QueryFields("*")
	if err != nil {
		fmt.Println("Error query fields", err)
		return
	}
	request.Limit = 0
	request.RecordBufferShift = 64000
	result, rerr := request.ReadLogicalWith("AF=" + recordNamePrefix)
	if rerr != nil {
		fmt.Println("Error reading records", rerr)
		return
	}
	t := adatypes.TraverserValuesMethods{EnterFunction: dumpFieldTypeValues}
	_, err = result.TraverseValues(t, nil)
	//Output: Connect to  23
	// Adabas url=23 fnr=0
	// Record found:
	//   IT = [ 1 ]
	//    S1 = > 0 <
	//    U1 = > 0 <
	//    S2 = > 0 <
	//    U2 = > 0 <
	//    S4 = > 0 <
	//    U4 = > 0 <
	//    S8 = > 0 <
	//    U8 = > 0 <
	//   BB = [ 1 ]
	//    BR = > 0 <
	//    B1 = > 0 <
	//   TY = [ 1 ]
	//    F4 = > 0.000000 <
	//    F8 = > 0.000000 <
	//   AA = [ 1 ]
	//    A1 = >   <
	//    AS = >   <
	//    A2 = >   <
	//    AB = >  <
	//    AF = > FIELD-TYPE-TEST      <
	//   WC = [ 1 ]
	//    WU = >   <
	//    WL = >   <
	//    W4 = >   <
	//    WF = >                                                    <
	//   PI = [ 1 ]
	//    PA = > 0 <
	//    PF = > 0 <
	//   UI = [ 1 ]
	//    UP = > 0 <
	//    UF = > 0 <
	//    UE = > 0 <
	// Record found:
	//   IT = [ 1 ]
	//    S1 = > -1 <
	//    U1 = > 1 <
	//    S2 = > -1000 <
	//    U2 = > 1000 <
	//    S4 = > -100000 <
	//    U4 = > 1000 <
	//    S8 = > -1000 <
	//    U8 = > 1000 <
	//   BB = [ 1 ]
	//    BR = > 0 <
	//    B1 = > 255 <
	//   TY = [ 1 ]
	//    F4 = > 0.000000 <
	//    F8 = > 0.000000 <
	//   AA = [ 1 ]
	//    A1 = > X <
	//    AS = > NORMALSTRING <
	//    A2 = > LARGESTRING <
	//    AB = > LOBST <
	//    AF = > FIELD-TYPE-TEST      <
	//   WC = [ 1 ]
	//    WU = >   <
	//    WL = >   <
	//    W4 = >   <
	//    WF = >                                                    <
	//   PI = [ 1 ]
	//    PA = > 0 <
	//    PF = > 0 <
	//   UI = [ 1 ]
	//    UP = > 0 <
	//    UF = > 0 <
	//    UE = > 0 <

}