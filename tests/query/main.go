/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type caller struct {
	url         string
	file        uint32
	counter     int
	name        string
	search      string
	fields      string
	threadNr    uint32
	limit       uint64
	credentials string
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

		if c.credentials != "" {
			c := strings.Split(c.credentials, ":")
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
			if c.credentials != "" {
				c := strings.Split(c.credentials, ":")
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
		readRequest.Limit = c.limit
		err = readRequest.QueryFields(c.fields)
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
		switch {
		case c.search != "":
			result, err = readRequest.ReadLogicalWith(c.search)
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}
		case c.name != "":
			result, err = readRequest.ReadLogicalBy(c.name)
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}
		default:
			result, err = readRequest.ReadPhysicalSequence()
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}

		}
		if output {
			fmt.Printf("Result of query search=%s descriptor=%s and fields=%s", c.search, c.name, c.fields)
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
		panic(err)
	}
	cfg.Level.SetLevel(level)
	cfg.OutputPaths = []string{name}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	sugar.Infof("Start logging with level", level)
	adatypes.Central.Log = sugar

	return
}

func main() {
	ed := os.Getenv("ENABLE_DEBUG")
	if ed != "" {
		level := zapcore.ErrorLevel
		switch ed {
		case "1":
			level = zapcore.DebugLevel
			adatypes.Central.SetDebugLevel(true)
		case "2":
			level = zapcore.InfoLevel
		}

		err := initLogLevelWithFile("query.log", level)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
			return
		}
	}
	defer TimeTrack(time.Now(), "Done testsuite test")

	var countValue int
	var threadValue int
	var file int
	var limit int
	var name string
	var search string
	var fields string
	var credentials string
	var displayFdt bool
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	//flag.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	flag.IntVar(&countValue, "c", 1, "Number of loops")
	flag.IntVar(&threadValue, "t", 1, "Number of threads")
	flag.IntVar(&limit, "l", 10, "Number of records maximal read")
	flag.StringVar(&name, "n", "", "Read descriptor order")
	flag.StringVar(&search, "s", "", "Search request")
	flag.StringVar(&credentials, "p", "", "Define user and password credentials of type 'user:password'")
	flag.StringVar(&fields, "d", "", "Query field list")
	flag.IntVar(&file, "f", 11, "Adabas file used to read, should be Employees file")
	flag.BoolVar(&output, "o", false, "display output")
	flag.BoolVar(&displayFdt, "F", false, "display Field Definition table")
	flag.BoolVar(&close, "C", false, "Close Adabas connection in each loop")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic("could not create CPU profile: " + err.Error())
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic("could not start CPU profile: " + err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	names := strings.Split(name, ",")

	if displayFdt {
		ada, aerr := adabas.NewAdabas(args[0])
		if aerr != nil {
			panic("Error init Adabas call: " + aerr.Error())
		}
		defer ada.Close()
		fdt, err := ada.ReadFileDefinition(adabas.Fnr(file))
		if err != nil {
			panic("Error evaluate Adabas FDT: " + err.Error())
		}
		fmt.Printf("Display FDT of database %s file %d\n", args[0], file)
		fmt.Println(fdt.String())
	}

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		c := caller{url: args[0], counter: countValue, threadNr: i + 1, search: search,
			name: names[int(i)%len(names)], file: uint32(file), limit: uint64(limit),
			fields: fields, credentials: credentials}
		go callAdabas(c)

	}
	wg.Wait()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			panic("could not create memory profile: " + err.Error())
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic("could not write memory profile: " + err.Error())
		}
		defer f.Close()
		fmt.Println("Start testsuite test")
	}

}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
