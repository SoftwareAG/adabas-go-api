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

// Package main Test application to read demo data in Adabas
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"

	log "github.com/sirupsen/logrus"
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
	helper := adatypes.NewHelper(buffer, int(received), binary.LittleEndian)
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
			if acbx.Acbxrsp != 0 {
				break
			}
			displayResult(acbx.Acbxisn, c.ada.AdabasBuffers[1].Bytes(), c.ada.AdabasBuffers[1].Received())
		}
		if acbx.Acbxrsp != 3 {
			fmt.Printf("Response code : %d\n", acbx.Acbxrsp)
		}

	}
	fmt.Printf("Finish thread %d with %d loops, quantity=%d\n", c.threadNr, c.counter, qunatity)

}

func initLogLevelWithFile(fileName string, level log.Level) (file *os.File, err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	name := p + "/" + fileName
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
	level := log.InfoLevel
	ed := os.Getenv("ENABLE_DEBUG")
	if ed == "1" {
		level = log.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	}

	f, err := initLogLevelWithFile("employee_client.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer f.Close()
	defer TimeTrack(time.Now(), "Done employee test")

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
		fmt.Printf("Usage: %s <url>\n", args[0])
		flag.PrintDefaults()
		return
	}

	fmt.Println("Start employee test")

	names := strings.Split(name, ",")

	wg.Add(threadValue)
	for i := uint32(0); i < uint32(threadValue); i++ {
		fmt.Printf("Start thread %d/%d\n", i+1, threadValue)
		adabas, err := adabas.NewAdabass(args[0])
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
