package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const destMap = "EmployeeMap"

type streamStruct struct {
	store *adabas.StoreRequest
}

func initLogLevelWithFile(fileName string, level zapcore.Level) (err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + fileName

	rawJSON := []byte(`{
		"level": "error",
		"encoding": "console",
		"outputPaths": [ "XXX"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		fmt.Printf("Initial logging JSON configuration error: %v\n", err)
		os.Exit(1)
	}
	cfg.Level.SetLevel(level)
	cfg.OutputPaths = []string{name}
	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("Initial logging error: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	adatypes.Central.SetDebugLevel(true)

	sugar.Infof("Start logging with level %v", level)
	adatypes.Central.Log = sugar

	return
}

func updateStream(record *adabas.Record, x interface{}) error {
	tc := x.(*streamStruct)
	updateRecord, err := tc.store.CreateRecord()
	if err != nil {
		return err
	}
	updateRecord.Isn = record.Isn
	record.DumpValues()
	last := record.ValueQuantity("AS")
	fmt.Println("Quantity", last)
	for i := uint32(0); i < uint32(last); i++ {
		vi, err := record.SearchValueIndex("AS", []uint32{i + 1})
		if err != nil {
			return err
		}
		sv, _ := vi.Int32()
		sv += 1000
		err = updateRecord.SetValueWithIndex("AS", []uint32{i + 1}, sv)
		if err != nil {
			return err
		}
	}
	fmt.Println("Read", record)
	fmt.Println("Update", updateRecord)
	return nil // tc.store.Update(updateRecord)
}

func main() {
	initLogLevelWithFile("employees.log", zapcore.DebugLevel)
	adabasModDBIDs := "1"
	if len(os.Args) > 1 {
		adabasModDBIDs = os.Args[1]
	}
	fmt.Println("Open connection to", adabasModDBIDs)

	connection, err := adabas.NewConnection(fmt.Sprintf("acj;target=%s", adabasModDBIDs))
	if err != nil {
		fmt.Println("Error connecting database:", err)
		return
	}
	defer connection.Close()

	readRequest, rerr := connection.CreateFileReadRequest(11)
	if rerr != nil {
		fmt.Println("Error creating read request:", rerr)
		return
	}
	err = readRequest.QueryFields("AA,AB,AS")
	if err != nil {
		fmt.Println("Error query field:", err)
		return
	}

	storeRequest, serr := connection.CreateStoreRequest(11)
	if serr != nil {
		fmt.Println("Error creating store request:", serr)
		return
	}
	serr = storeRequest.StoreFields("AS")
	if serr != nil {
		fmt.Println("Error define store fields:", serr)
		return
	}
	fmt.Println("Read logical search...", destMap)
	tc := &streamStruct{store: storeRequest}
	_, err = readRequest.ReadLogicalWithStream("AE='SMITH'", updateStream, tc)
	if err != nil {
		fmt.Println("Error updating records:", err)
		return
	}
	/*err = storeRequest.EndTransaction()
	if err != nil {
		fmt.Println("Error end of transaction:", err)
		return
	}*/
}
