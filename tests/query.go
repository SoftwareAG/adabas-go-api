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
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
)

type caller struct {
	url         string
	file        uint32
	counter     int
	name        string
	fields      string
	threadNr    uint32
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
	tid := strconv.Itoa(int(c.threadNr))

	for i := 0; i < c.counter; i++ {
		l := adatypes.Central.Log.(*log.Logger)
		l.WithFields(log.Fields{
			"thread": tid,
		}).Debugf("Start counter")
		if close {
			connection, err = c.createConnection()
			if err != nil {
				fmt.Printf("Error create connection to thread %d\n", c.threadNr)
				return
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
		result, err = readRequest.ReadLogicalWith(c.name)
		if err != nil {
			fmt.Printf("Error reading thread %d with %d loops: %v\n", c.threadNr, i, err)
			return
		}
		if output {
			fmt.Println("Result of query " + c.name)
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

func initLogLevelWithFile(fileName string, level log.Level) (file *os.File, err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + fileName
	file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	log.SetLevel(level)

	log.SetOutput(file)
	myLog := log.New()
	myLog.SetLevel(level)
	myLog.Out = file

	// log.SetOutput(file)
	adatypes.Central.Log = myLog

	return
}

func main() {
	level := log.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	switch ed {
	case "1":
		level = log.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	case "2":
		level = log.InfoLevel
	default:
		level = log.ErrorLevel
	}

	f, err := initLogLevelWithFile("query.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer f.Close()
	defer TimeTrack(time.Now(), "Done testsuite test")

	var countValue int
	var threadValue int
	var file int
	var name string
	var fields string
	var credentials string
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	//flag.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	flag.IntVar(&countValue, "c", 1, "Number of loops")
	flag.IntVar(&threadValue, "t", 1, "Number of threads")
	flag.StringVar(&name, "n", "AE=SMITH", "Search request")
	flag.StringVar(&credentials, "p", "", "Define user and password credentials of type 'user:password'")
	flag.StringVar(&fields, "d", "AA", "Query field list")
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

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	names := strings.Split(name, ",")

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		c := caller{url: args[0], counter: countValue, threadNr: i + 1,
			name: names[int(i)%len(names)], file: uint32(file),
			fields: fields, credentials: credentials}
		go callAdabas(c)

	}
	wg.Wait()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
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
