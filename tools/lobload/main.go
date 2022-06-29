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

// Large object load example application to read file data in Adabas a predefined
// Adabas Map "LOBEXAMPLE".
// It loads a file content and stores it into a Adabas field defined by some
// different file specific parameters like mimetype etc.
// In advance a SHA checksum is generated and stored in a Adabas field. The
// checksum can be used to validate storage late on.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/nfnt/resize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
	level := zapcore.ErrorLevel
	ed := os.Getenv("ENABLE_DEBUG")
	switch ed {
	case "1":
		level = zapcore.DebugLevel
		adatypes.Central.SetDebugLevel(true)
	case "2":
		level = zapcore.InfoLevel
	}

	err := initLogLevelWithFile("query.log", level)
	if err != nil {
		fmt.Printf("Initial logging error: %v\n", err)
		os.Exit(1)
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

func loadFile(fileName string, ada *adabas.Adabas) error {
	fmt.Println("Load file", fileName)
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	fmt.Printf("Number of bytes read: %d/%d -> %v\n", n, len(data), err)
	if err != nil {
		return err
	}
	h := sha256.New()
	h.Write(data)
	fmt.Printf("SHA ALL: %x\n", h.Sum(nil))
	var buffer bytes.Buffer
	buffer.Write(data)
	srcImage, _, _ := image.Decode(&buffer)
	//dstImageFill := imaging.Fill(srcImage, 100, 100, imaging.Center, imaging.Lanczos)
	newImage := resize.Resize(200, 0, srcImage, resize.Lanczos3)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, newImage, nil)
	if err != nil {
		fmt.Println("Error generating thumbnail", err)
		return err
	}
	thumbnail := buf.Bytes()
	fmt.Println("Thumbnail data size", len(thumbnail))

	adabasMap, _, serr := adabas.SearchMapRepository(ada.ID, "LOBEXAMPLE")
	if serr != nil {
		fmt.Println("Search map repository", serr)
		return serr
	}
	storeRequest, err := adabas.NewAdabasMapNameStoreRequest(ada, adabasMap)
	if err != nil {
		fmt.Println("Error create store request", err)
		return err
	}
	defer storeRequest.Close()

	adatypes.Central.Log.Debugf("Store fields prepare Picture")
	err = storeRequest.StoreFields("Host,Date,Picture,Thumbnail,Location,Size,Checksum")
	if err != nil {
		fmt.Println("Store fields error", err)
		return err
	}
	storeRecord, rErr := storeRequest.CreateRecord()
	if rErr != nil {
		fmt.Println("Create record error", rErr)
		return rErr
	}
	adatypes.Central.Log.Debugf("Set value to Picture")
	storeRecord.SetValue("Host", hostname)
	err = storeRecord.SetValue("Date", time.Now().Unix())
	if err != nil {
		fmt.Println("Error setting data", err)
	}
	storeRecord.SetValue("Directory", filepath.Dir(fileName))
	storeRecord.SetValue("Filename", filepath.Base(fileName))
	storeRecord.SetValue("absoluteFilename", fileName)
	storeRecord.SetValue("Picture", data)
	storeRecord.SetValue("Thumbnail", thumbnail)
	storeRecord.SetValue("PictureSHAchecksum", fmt.Sprintf("%X", h.Sum(nil)))
	adatypes.Central.Log.Debugf("Done set value to Picture, searching ...")

	err = storeRequest.Store(storeRecord)
	if err != nil {
		fmt.Printf("Store request error %v\n", err)
		return err
	}
	fmt.Println("Store record into ISN=", storeRecord.Isn)
	storeRequest.EndTransaction()
	validateUsingMap(ada, storeRecord.Isn)
	return nil
}

func validateUsingMap(a *adabas.Adabas, isn adatypes.Isn) {
	fmt.Println("Validate using Map and ISN=", isn)
	mapRepository := adabas.NewMapRepository(a.URL, 4)
	request, err := adabas.NewReadRequest("LOBEXAMPLE", a, mapRepository)
	if err != nil {
		fmt.Printf("New map request error %v\n", err)
		return
	}
	defer request.Close()
	_, openErr := request.Open()
	if openErr == nil {
		err := request.QueryFields("Picture")
		if err != nil {
			return
		}
		fmt.Println("Query defined, read record ...")
		result, rerr := request.ReadISN(isn)
		if rerr == nil {
			picValue := result.Values[0].HashFields["Picture"]
			if picValue == nil {
				return
			}
		}
	}
	fmt.Println("Data validated with map methods")
}

func main() {
	var fileName string
	var dbidParameter string
	var mapFnrParameter int
	flag.StringVar(&fileName, "p", "", "File name of picture to be imported")
	flag.StringVar(&dbidParameter, "d", "23", "Map repository Database id")
	flag.IntVar(&mapFnrParameter, "f", 4, "Map repository file number")
	flag.Parse()

	if fileName == "" {
		fmt.Println("File name option is required")
		return
	}

	id := adabas.NewAdabasID()
	a, err := adabas.NewAdabas(dbidParameter, id)
	if err != nil {
		fmt.Println("Adabas target generation error", err)
		return
	}
	adabas.AddGlobalMapRepository(a.URL, adabas.Fnr(mapFnrParameter))
	defer adabas.DelGlobalMapRepository(a.URL, adabas.Fnr(mapFnrParameter))
	adabas.DumpGlobalMapRepositories()

	err = filepath.Walk(fileName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fmt.Println("Check", path)
		if strings.HasSuffix(strings.ToLower(path), ".jpg") {
			fmt.Println("Load", path)
			return loadFile(path, a)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking path", err)
	}
	fmt.Println("End of lob load")

}
