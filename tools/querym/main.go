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

// Query application to read data in Adabas using Adabas Map references.
// It can start a read using multiple threads and loops.
// A search string can restrict the query to a specific case.
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
	counter     int
	mapName     string
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
	connStr := fmt.Sprintf("acj;map;auth=NONE,id=%d,user=user%04d", c.threadNr, c.threadNr)
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
	}

	steps := c.counter / 10
	if c.counter < 50 || steps == 0 {
		steps = 50
	}
	maxTime := 1.0

	last := time.Now()
	dumpMessage := true
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
		readRequest, rerr := connection.CreateMapReadRequest(c.mapName)
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
			if dumpMessage {
				fmt.Println("Search for ", c.search)
				dumpMessage = false
			}
			result, err = readRequest.ReadLogicalWith(c.search)
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}
		case c.name != "":
			if dumpMessage {
				fmt.Println("Order by ", c.name)
				dumpMessage = false
			}
			result, err = readRequest.ReadLogicalBy(c.name)
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}
		default:
			if dumpMessage {
				fmt.Println("Physical read")
				dumpMessage = false
			}
			result, err = readRequest.ReadPhysicalSequence()
			if err != nil {
				fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
				return
			}

		}
		if output {
			fmt.Printf("Result of query search=%s descriptor=%s and fields=%s\n", c.search, c.name, c.fields)
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

	err := initLogLevelWithFile("querym.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer TimeTrack(time.Now(), "Done testsuite test")

	var countValue int
	var threadValue int
	var repository string
	var limit int
	var name string
	var search string
	var fields string
	var showMaps bool
	var showMap bool
	var credentials string
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	//flag.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	flag.IntVar(&countValue, "c", 1, "Number of loops [default 1]")
	flag.IntVar(&threadValue, "t", 1, "Number of threads [default 1]")
	flag.IntVar(&limit, "l", 10, "Number of records maximal read  [default 10]")
	flag.StringVar(&name, "n", "", "Read descriptor order")
	flag.StringVar(&search, "s", "", "Search request like 'name=SMITH'")
	flag.StringVar(&credentials, "p", "", "Define user and password credentials of type 'user:password'")
	flag.StringVar(&fields, "d", "", "Query field list like 'name,personnel-id', '*' queries all fields, [default is '']")
	flag.StringVar(&repository, "r", "", "Adabas map repository used for search. Need to be defined using '<db target>,fnr'")
	flag.BoolVar(&output, "o", false, "display output")
	flag.BoolVar(&showMaps, "M", false, "List all maps available")
	flag.BoolVar(&showMap, "m", false, "List all map fields of given map")
	flag.BoolVar(&close, "C", false, "Close Adabas connection in each loop")
	flag.Parse()
	args := flag.Args()
	if !showMaps && len(args) < 1 {
		fmt.Printf("Usage: %s <map name>\n", os.Args[0])
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

	err = adabas.AddGlobalMapRepositoryReference(repository)
	if err != nil {
		fmt.Printf("Error repository: %v\n", err)
		os.Exit(1)
	}

	if showMaps {
		adabas.DumpGlobalMapRepositories()
		if len(args) < 1 {
			return
		}
	}
	if showMap {
		id := adabas.NewAdabasID()
		m, _, merr := adabas.SearchMapRepository(id, args[0])
		if merr != nil {
			fmt.Println("Searched map not found", merr)
			return
		}
		fmt.Println(m.String())
	}

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		c := caller{mapName: args[0], counter: countValue, threadNr: i + 1,
			name: names[int(i)%len(names)], limit: uint64(limit), search: search,
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
		fmt.Println("Wrote memory profiler data")
	}

}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
