/*
* Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */

// Package main Deprecated test application
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var output = false
var url string

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

	sugar.Infof("Start logging with level %v", level)
	adatypes.Central.Log = sugar

	return
}

func dumpStream(record *adabas.Record, x interface{}) error {
	storeRequest := x.(*adabas.StoreRequest)
	a, _ := record.SearchValue("RN")
	b, _ := record.SearchValue("RD")
	fmt.Printf("Read %d -> %s = %s\n", record.Isn, a, b.String())
	if b.String() != "" {
		if strings.HasPrefix(b.String(), url) {
			record.SetValue("RD", "")
			storeRequest.Update(record)
		}
	}
	return nil
}

func main() {
	level := zapcore.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	switch ed {
	case "1":
		level = zapcore.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	case "2":
		level = zapcore.InfoLevel
	}

	err := initLogLevelWithFile("clear_map_reference.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer TimeTrack(time.Now(), "Done testsuite test")

	var file int

	flag.IntVar(&file, "f", 11, "Adabas file used to read, should be Employees file")
	flag.BoolVar(&output, "o", false, "display output")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	url = args[0]

	connection, err := adabas.NewConnection("acj;target=" + args[0])
	if err != nil {
		fmt.Println("Error creating the connection", err)
		return
	}
	defer connection.Close()
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(adabas.Fnr(file))
	if rErr != nil {
		fmt.Println("Error creating read request", rErr)
		return
	}
	readRequest.Limit = 0
	readRequest.Multifetch = 1

	storeRequest, rErr := connection.CreateStoreRequest(adabas.Fnr(file))
	if rErr != nil {
		fmt.Println("Error creating read request", rErr)
		return
	}

	err = readRequest.QueryFields("RN,RD")
	if err != nil {
		fmt.Println("Error wrong fields in file", err)
		return
	}
	err = storeRequest.StoreFields("RN,RD")
	if err != nil {
		fmt.Println("Error wrong fields in file", err)
		return
	}
	_, err = readRequest.ReadLogicalByStream("RN", dumpStream, storeRequest)
	if err != nil {
		fmt.Println("Error reading data", err)
		return
	}
	connection.EndTransaction()
}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
