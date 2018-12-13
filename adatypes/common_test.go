/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
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

package adatypes

import (
	"fmt"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	log "github.com/sirupsen/logrus"
)

func initTestLogWithFile(t *testing.T, fileName string) *os.File {
	file, err := initLogWithFile(fileName)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
		return nil
	}
	return file
}

func initLogWithFile(fileName string) (file *os.File, err error) {
	level := log.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	if ed == "1" {
		level = log.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	}
	return initLogLevelWithFile(fileName, level)
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
	myLog := log.New()
	myLog.SetLevel(level)
	myLog.Out = file

	// log.SetOutput(file)
	Central.Log = myLog
	return
}

func TestLog(t *testing.T) {
	f, err := initLogWithFile("log.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	hallo := "HELLO"
	Central.Log.Debugf("This is a test of data %s", hallo)
}
