/*
* Copyright © 2018-2020 Software AG, Darmstadt, Germany and/or its licensors
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

//go:generate go run message_generate.go

// Package main Generate go files out of message content
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const messageFilePattern = `^\w+\.[a-zA-Z]{2}$`

var locales map[string]map[string]string

var headerTemplate = `/*
* Copyright © 2019-2020 Software AG, Darmstadt, Germany and/or its licensors
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

/*
* CODE GENERATED AUTOMATICALLY WITH adabas/generate
* THIS FILE MUST NOT BE EDITED BY HAND
*/

package {{.Name}}

var statisMessages = []struct {
		code    string
		locale string
		message  string
	}{
		{{range $localekey, $localevalue := .Messages}}{{range $key, $value := $localevalue}}
		  { "{{$key}}","en","{{$value}}" },{{end}} {{end}}
	}`

func generateByTemplate(file *os.File) error {

	buff := bytes.NewBuffer(nil)

	tmplHead, err := parseTemplates()
	if err != nil {
		return err
	}

	// Generate header
	if err := tmplHead.Execute(buff, struct {
		Name     string
		Messages map[string]map[string]string
	}{
		"adatypes",
		locales,
	}); err != nil {
		return err
	}

	// Formating code
	code, err := format.Source(buff.Bytes())
	if err != nil {
		fmt.Println("Error formatting ....", buff.String())
		return err
	}
	fmt.Fprintln(file, string(code))
	return nil
}

func parseTemplates() (*template.Template, error) {
	tmplHead, err := template.New("header").Parse(headerTemplate)
	if err != nil {
		return nil, err
	}
	return tmplHead, nil
}

func main() {
	fmt.Println("Generate message code")

	locales = make(map[string]map[string]string)

	curdir := os.Getenv("CURDIR")
	if curdir == "" {
		curdir = fmt.Sprintf("..%cadatypes%c", os.PathSeparator, os.PathSeparator)
	} else {
		curdir = fmt.Sprintf("%s%cadatypes%c", curdir, os.PathSeparator, os.PathSeparator)
	}
	destinationFile := fmt.Sprintf("%s%cstatic_messages.go", curdir, os.PathSeparator)

	if _, err := os.Stat(destinationFile); !os.IsNotExist(err) {
		fmt.Println("Remove old file: ", destinationFile)
		err = os.Remove(destinationFile)
		if err != nil {
			fmt.Println("Error deleting file", err)
			return
		}
	}

	locales = make(map[string]map[string]string)
	file, err := os.OpenFile(destinationFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}

	messageLocation := os.Getenv("GO_ADA_MESSAGES")
	if messageLocation == "" {
		messageLocation = fmt.Sprintf(".%cmessages", os.PathSeparator)
	}
	if _, err := os.Stat(messageLocation); os.IsNotExist(err) {
		fmt.Printf("Location not found: %s", messageLocation)
		fmt.Printf("Message location at %s not found, plesae set GO_ADA_MESSAGES", messageLocation)
		return
	}

	if error := filepath.Walk(messageLocation, loadMessageFile); error != nil && !os.IsNotExist(error) {
		fmt.Println("Error reading messages files:", error)
	}

	if err := generateByTemplate(file); err != nil {
		fmt.Println("Error generate with template", err)
		return
	}

}

// Load a single message file
func loadMessageFile(path string, info os.FileInfo, osError error) error {
	if osError != nil {
		return osError
	}
	if info.IsDir() {
		return nil
	}
	fmt.Printf("Check path %s\n", path)
	locale := parseLocaleFromFileName(info.Name())
	if matched, _ := regexp.MatchString(messageFilePattern, info.Name()); matched {
		fmt.Printf("Info Name %s locale %s\n", info.Name(), locale)
		// If already parsed a message file for this locale, merge both
		var messages map[string]string
		var ok bool
		if messages, ok = locales[locale]; !ok {
			messages = make(map[string]string)
			locales[locale] = messages
		}
		messageFile, err := os.OpenFile(path, os.O_RDONLY, 0666)
		if err != nil {
			fmt.Printf("Info Name error %#v\n", err)
			return err
		}
		scanner := bufio.NewScanner(messageFile)
		for scanner.Scan() {
			line := scanner.Text()
			msg := line[11:]
			msg = strings.Replace(msg, "\"", "'", -1)
			messages[line[:10]] = msg
		}
		fmt.Println("Scanned")
	} else {
		fmt.Printf("Ignoring file %s because it did not have a valid extension  \n", info.Name())
	}

	return nil
}

func parseLocaleFromFileName(file string) string {
	extension := filepath.Ext(file)[1:]
	return strings.ToLower(extension)
}
