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
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initTestLogWithFile(t *testing.T, fileName string) {
	err := initLogWithFile(fileName)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}
}

func initLogWithFile(fileName string) (err error) {
	level := zap.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	if ed == "1" {
		level = zap.DebugLevel
		Central.SetDebugLevel(true)
	}
	return initLogLevelWithFile(fileName, level)
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Remove leading slash left by url.Parse()
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func initLogLevelWithFile(fileName string, level zapcore.Level) (err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	var name string
	if runtime.GOOS == "windows" {
		zap.RegisterSink("winfile", newWinFileSink)
		//		OutputPaths: []string{"stdout", "winfile:///" + filepath.Join(GlobalConfigDir.Path, "info.log.json")},
		name = "winfile:///" + p + string(os.PathSeparator) + fileName
	} else {
		name = "file://" + filepath.ToSlash(p+string(os.PathSeparator)+fileName)
	}

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
		return err
	}
	cfg.Level.SetLevel(level)
	cfg.OutputPaths = []string{name}
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	Central.Log = sugar

	sugar.Infof("AdabasGoApi logger initialization succeeded")
	return nil
}

func TestLog(t *testing.T) {
	err := initLogWithFile("adatypes.Central.Log.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer TimeTrack(time.Now(), "Time Track Unit test ")

	hallo := "HELLO"
	Central.Log.Debugf("This is a test of data %s", hallo)

	LogMultiLineString(true, "ABC\nXXXX\n")
	d := Central.IsDebugLevel()
	Central.SetDebugLevel(true)
	Central.SetDebugLevel(false)
	Central.SetDebugLevel(d)
}
