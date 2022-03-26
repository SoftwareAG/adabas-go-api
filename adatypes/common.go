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

package adatypes

import (
	"fmt"
	"strings"
	"time"
)

type nilLogger struct {
}

func lognil() *nilLogger {
	return &nilLogger{}
}

func (*nilLogger) Debugf(format string, args ...interface{}) {
}

func (*nilLogger) Infof(format string, args ...interface{}) {
}

func (*nilLogger) Errorf(format string, args ...interface{}) {
}

func (*nilLogger) Fatal(args ...interface{}) {
}

// Log defines the log interface to manage other Log output frameworks
type Log interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
}

// centralOptions central structure containing the current log reference
type centralOptions struct {
	Log   Log
	debug bool
}

// Central central configuration
var Central = centralOptions{Log: lognil(), debug: false}

func (log *centralOptions) IsDebugLevel() bool {
	return log.debug
}

func (log *centralOptions) SetDebugLevel(debug bool) {
	log.debug = debug
	if debug {
		fmt.Println("Warning debug is enabled")
	}
}

// LogMultiLineString log multi line string to log. This prevent the \n display in log.
// Instead multiple lines are written to log
func LogMultiLineString(debug bool, logOutput string) {
	if debug && !Central.IsDebugLevel() {
		return
	}
	columns := strings.Split(logOutput, "\n")
	for _, c := range columns {
		if debug {
			Central.Log.Debugf("%s", c)
		} else {
			Central.Log.Errorf("%s", c)
		}
	}
}

// TimeTrack defer function measure the difference end log it to log management, like
//    defer TimeTrack(time.Now(), "CallAdabas "+string(adabas.Acbx.Acbxcmd[:]))
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Central.Log.Infof("%s took %s", name, elapsed)
}
