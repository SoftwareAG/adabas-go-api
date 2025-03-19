/*
* Copyright © 2021-2025 Software GmbH, Darmstadt, Germany and/or its licensors
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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/pkg/profile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type caller struct {
	url         string
	shortOutput bool
	credential  string
}

func (c caller) createConnection() (*adabas.Connection, error) {
	connStr := fmt.Sprintf("acj;target=%s;auth=NONE", c.url)
	connection, err := adabas.NewConnection(connStr)
	if err != nil {
		fmt.Println("Open connection error", err)
		return nil, err
	}
	return connection, nil
}

func callAdabas(call caller) {

	var connection *adabas.Connection
	var err error
	connection, err = call.createConnection()
	if err != nil {
		fmt.Printf("Error create connection to database %v\n", err)
		return
	}
	defer connection.Close()

	if call.credential != "" {
		c := strings.Split(call.credential, ":")
		if len(c) != 2 {
			fmt.Printf("User credentials invalid format")
			return
		}
		connection.AddCredential(c[0], c[1])
	}

	err = connection.Open()
	if err != nil {
		fmt.Printf("Error opening connection: %v\n", err)
		return
	}

	if connection.IsCluster() {
		nodes := connection.GetClusterNodes()
		if !call.shortOutput {
			fmt.Println("\nCluster node list:")
		}
		buffer := &bytes.Buffer{}
		for i, n := range nodes {
			if call.shortOutput {
				if buffer.Len() > 0 {
					buffer.WriteRune(';')
				}
				buffer.WriteString(n.String())
			} else {
				prefix := "  "
				if i == 0 {
					prefix = "* "
				}
				fmt.Println(prefix + n.String())
			}
		}
		if call.shortOutput {
			fmt.Println(buffer.String())
		} else {
			fmt.Printf("Finish cluster list\n\n")
		}
	}

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

	err := initLogLevelWithFile("betacluster.log", level)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}

	var credential string
	var shortOutput bool
	var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile")
	var memprofile = flag.Bool("memprofile", false, "write memory profile")

	flag.StringVar(&credential, "p", "", "Define user and password credentials of type 'user:password'")
	flag.BoolVar(&shortOutput, "s", false, "display short version of output")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	if !shortOutput {
		defer TimeTrack(time.Now(), "Done cluster evaluation")
	}

	if *cpuprofile {
		defer profile.Start().Stop()
	}
	if *memprofile {
		defer profile.Start(profile.MemProfile).Stop()
		//defer writeMemProfile(*memprofile)
	}
	if !shortOutput {
		fmt.Printf("Start cluster evaluation\n")
	}
	c := caller{url: args[0], shortOutput: shortOutput, credential: credential}
	callAdabas(c)

}

// TimeTrack logger
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
