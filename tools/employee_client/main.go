/*
* Copyright © 2018-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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

// Test application to read data in Adabas using the Adabas example file 11,
// containing EMPLOYEES example data.
// It can start a read using multiple threads and loops.
// A search string can restrict the query to a specific case.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type caller struct {
	ada      *adabas.Adabas
	counter  int
	name     string
	threadNr uint32
}

var wg sync.WaitGroup
var output bool

func displayResult(isn adatypes.Isn, buffer []byte, received uint64) {
	if !output {
		return
	}
	helper := adatypes.NewHelper(buffer, int(received), adabas.Endian())
	aa, err := helper.ReceiveString(8)
	if err != nil {
		fmt.Println("Error receiving AA", err)
		return
	}
	ac, err := helper.ReceiveString(20)
	if err != nil {
		fmt.Println("Error receiving AC", err)
		return
	}
	ad, err := helper.ReceiveString(20)
	if err != nil {
		fmt.Println("Error receiving AD", err)
		return
	}
	ae, err := helper.ReceiveString(20)
	if err != nil {
		fmt.Println("Error receiving AE", err)
		return
	}
	as, err := helper.ReceiveInt32()
	if err != nil {
		fmt.Println("Error receiving AS", err)
		return
	}

	fmt.Printf("ISN=%-4d AA=%s AC=%s AD=%s AE=%s AS=%d\n", isn, aa, ac, ad, ae, as)
}

func callAdabas(c caller) {
	defer wg.Done()
	defer c.ada.Close()

	steps := c.counter / 10
	if c.counter < 50 || steps == 0 {
		steps = 50
	}
	var qunatity uint64
	for i := 0; i < c.counter; i++ {
		if i > 0 && i%steps == 0 {
			fmt.Printf("Call thread %d counter %d query for %s quantity=%d\n", c.threadNr,
				i, c.name, qunatity)
		}
		err := c.ada.Open()
		if err != nil {
			fmt.Println("Error opening database:", err)
			return
		}

		l := len(c.name)
		len := strconv.Itoa(l)

		acbx := c.ada.Acbx
		acbx.Acbxcmd = [2]byte{'S', '1'}
		acbx.Acbxfnr = 11
		acbx.Acbxcid = [4]uint8{0xff, 0xff, 0xff, 0xff}
		acbx.Acbxisl = 0
		acbx.Acbxisq = 0
		c.ada.AdabasBuffers = make([]*adabas.Buffer, 0)
		c.ada.AdabasBuffers = append(c.ada.AdabasBuffers, adabas.NewBuffer(adabas.AbdAQFb))
		c.ada.AdabasBuffers[0].WriteString("AA,AB,ASN,4,B.")
		c.ada.AdabasBuffers = append(c.ada.AdabasBuffers, adabas.NewBuffer(adabas.AbdAQRb))
		c.ada.AdabasBuffers[1].Allocate(1024)
		c.ada.AdabasBuffers = append(c.ada.AdabasBuffers, adabas.NewBuffer(adabas.AbdAQSb))
		c.ada.AdabasBuffers[2].WriteString("AE," + len + ".")
		c.ada.AdabasBuffers = append(c.ada.AdabasBuffers, adabas.NewBuffer(adabas.AbdAQVb))
		c.ada.AdabasBuffers[3].WriteString(c.name)
		err = c.ada.CallAdabas()
		if err != nil {
			fmt.Printf("Error calling adabas : %v rsp=%d\n", err, acbx.Acbxrsp)
			return
		}
		qunatity = acbx.Acbxisq
		//	fmt.Println(acbx.String())
		// displayResult(acbx.Acbxisn, c.ada.AdabasBuffers[1].Bytes(), c.ada.AdabasBuffers[1].Received())
		//	fmt.Print(c.ada.AdabasBuffers[1].String())

		acbx.Acbxcmd = [2]byte{'L', '1'}
		acbx.Acbxfnr = 11
		acbx.Acbxcop[1] = 'N'
		for {
			err = c.ada.CallAdabas()
			if err != nil {
				fmt.Printf("Error calling adabas : %v rsp=%d\n", err, acbx.Acbxrsp)
				return
			}
			if acbx.Acbxrsp != adabas.AdaNormal {
				break
			}
			displayResult(acbx.Acbxisn, c.ada.AdabasBuffers[1].Bytes(), c.ada.AdabasBuffers[1].Received())
		}
		if acbx.Acbxrsp != adabas.AdaEOF {
			fmt.Printf("Response code : %d\n", acbx.Acbxrsp)
		}

	}
	fmt.Printf("Finish thread %d with %d loops, quantity=%d\n", c.threadNr, c.counter, qunatity)

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

	err := initLogLevelWithFile("employee_client.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer TimeTrack(time.Now(), "Finished employee test and ")

	var countValue int
	var threadValue int
	var name string

	//flag.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	flag.IntVar(&countValue, "c", 1, "Number of loops")
	flag.IntVar(&threadValue, "t", 1, "Number of threads")
	flag.StringVar(&name, "n", "SMITH", "Test search for employee names separated by ','")
	flag.BoolVar(&output, "o", false, "display output")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Start employee test, connection to %s...\n", args[0])

	names := strings.Split(name, ",")

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		id := adabas.NewAdabasID()
		adabas, err := adabas.NewAdabas(args[0], id)
		if err != nil {
			fmt.Println("Error createing Adabas link:", err)
			return
		}
		c := caller{ada: adabas, counter: countValue, threadNr: i + 1,
			name: names[int(i)%len(names)]}
		go callAdabas(c)

	}
	wg.Wait()
}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
