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

// Test suite application to read data in Adabas. It provides the possibility
// to define the Adabas file to be used.
// It can start a read using multiple threads and loops.
// A search string can restrict the query to a specific case.
// It can track CPU and memory profiler informations.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/pkg/profile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type caller struct {
	url        string
	file       uint32
	counter    int
	name       string
	threadNr   uint32
	credential string
}

var wg sync.WaitGroup
var output = false
var close = false

func (c caller) createConnection() (*adabas.Connection, error) {
	connStr := fmt.Sprintf("acj;target=%s;auth=NONE,id=%d,user=user%04d", c.url, c.threadNr, c.threadNr)
	connection, err := adabas.NewConnection(connStr)
	if err != nil {
		fmt.Println("Open connection error", err)
		return nil, err
	}
	return connection, nil
}

func callAdabas(c caller) {
	defer wg.Done()

	var connection *adabas.Connection
	var err error
	if !close {
		connection, err = c.createConnection()
		if err != nil {
			fmt.Printf("Error create connection to thread %d\n", c.threadNr)
			return
		}
		defer connection.Close()

		if c.credential != "" {
			c := strings.Split(c.credential, ":")
			if len(c) != 2 {
				fmt.Printf("User credentials invalid format")
				return
			}
			connection.AddCredential(c[0], c[1])
		}

		err = connection.Open()
		if err != nil {
			fmt.Printf("Error opening connection to thread %d: %v\n", c.threadNr, err)
			return
		}
	}

	steps := c.counter / 10
	if c.counter < 50 || steps == 0 {
		steps = 50
	}
	maxTime := 1.0

	last := time.Now()

	for i := 0; i < c.counter; i++ {
		if close {
			connection, err = c.createConnection()
			if err != nil {
				fmt.Printf("Error create connection to thread %d\n", c.threadNr)
				return
			}
			if c.credential != "" {
				c := strings.Split(c.credential, ":")
				if len(c) != 2 {
					fmt.Printf("User credentials invalid format")
					return
				}
				connection.AddCredential(c[0], c[1])
			}

			err = connection.Open()
			if err != nil {
				fmt.Printf("Error opening connection to thread %d\n", c.threadNr)
				return
			}
		}
		readRequest, rerr := connection.CreateFileReadRequest(adabas.Fnr(c.file))
		if rerr != nil {
			fmt.Println("Error creating read reference of database:", rerr)
			return
		}
		err = readRequest.QueryFields("AA,AB,AS[N]")
		if err != nil {
			fmt.Println("Error query fields of database file:", err)
			return
		}

		newTime := time.Now()
		diff := newTime.Sub(last)
		//fmt.Println(diff.Minutes())
		if (i > 0 && i%steps == 0) || (diff.Minutes() > maxTime) {
			if diff.Minutes() > maxTime {
				maxTime += 1.0
			}
			fmt.Printf("Call thread %d counter %d query for %s used %v\n", c.threadNr,
				i, c.name, diff)
			//			last = newTime
		}
		var result *adabas.Response
		result, err = readRequest.ReadLogicalWith("AE=" + c.name)
		if err != nil {
			fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
			return
		}
		if output {
			result.DumpValues()
		}
		if close {
			connection.Close()
		} else {
			connection.Release()
		}
	}
	fmt.Printf("Finish thread %d with %d loops\n", c.threadNr, c.counter)

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

	sugar.Infof("Start logging with level %v", level)
	adatypes.Central.Log = sugar

	return
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

	err := initLogLevelWithFile("testsuite.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer TimeTrack(time.Now(), "Done testsuite test")

	var countValue int
	var threadValue int
	var file int
	var name string
	var credential string
	var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile")
	var memprofile = flag.Bool("memprofile", false, "write memory profile")

	//flag.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	flag.IntVar(&countValue, "c", 1, "Number of loops")
	flag.IntVar(&threadValue, "t", 1, "Number of threads")
	flag.StringVar(&name, "n", "SMITH", "Test search for employee names separated by ','")
	flag.StringVar(&credential, "p", "", "Define user and password credentials of type 'user:password'")
	flag.IntVar(&file, "f", 11, "Adabas file used to read, should be Employees file")
	flag.BoolVar(&output, "o", false, "display output")
	flag.BoolVar(&close, "C", false, "Close Adabas connection in each loop")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if *cpuprofile {
		defer profile.Start().Stop()
	}
	if *memprofile {
		defer profile.Start(profile.MemProfile).Stop()
		//defer writeMemProfile(*memprofile)
	}

	names := strings.Split(name, ",")

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		c := caller{url: args[0], counter: countValue, threadNr: i + 1,
			name: names[int(i)%len(names)], file: uint32(file),
			credential: credential}
		go callAdabas(c)

	}
	wg.Wait()

}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
