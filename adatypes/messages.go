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
	"regexp"
	"sync"
	"time"
)

const messagePrefix = "ADAGE"

type errorCode int

const messageFilePattern = `^\w+\.[a-zA-Z]{2}$`

var locales map[string]map[string]string
var once sync.Once
var loadErr *Error
var onceBody = func() {
	Central.Log.Debugf("Only once load messages")
	loadErr = loadMessagesOnce()
}

// Error error message with code and time
type Error struct {
	When    time.Time
	Code    string
	Message string
}

// Error error interface function, providing message error code and message. The Adabas error provides
// message code and message text
func (e Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

// NewGenericError create a genernic non Adabas response error
func NewGenericError(code errorCode, args ...interface{}) *Error {
	Central.Log.Debugf("Generate generice error for error code %d", code)
	msgCode := fmt.Sprintf("ADG%07d", code)
	// fmt.Printf("Generated out of %d -> %s\n", code, msgCode)
	msg := Translate("en", msgCode)
	for i, x := range args {
		m := fmt.Sprintf("%v", x)
		c := fmt.Sprintf("\\{%d\\}", i)
		re := regexp.MustCompile(c)
		msg = re.ReplaceAllString(msg, m)
	}
	return &Error{When: time.Now(), Code: msgCode, Message: msg}
}

// initMessages loads messages from all message files on the given path.
func initMessages() *Error {
	once.Do(onceBody)
	return loadErr
}

// LoadMessages loads messages from all message files on the given path.
func loadMessagesOnce() *Error {
	Central.Log.Debugf("Load messages")
	fmt.Println("Load messages")
	locales = make(map[string]map[string]string)
	for _, m := range statisMessages {
		var messages map[string]string
		if messageMap, ok := locales[m.locale]; ok {
			messages = messageMap
		} else {
			messages = make(map[string]string)
			locales[m.locale] = messages
		}
		messages[m.code] = m.message
	}
	Central.Log.Debugf("Loaded messages: %d", len(statisMessages))

	return nil
}

// Translate translates content to target language.
func Translate(locale, message string, args ...interface{}) string {
	if err := initMessages(); err != nil {
		return "Error initialize Adabas messages"
	}
	if localeMap, ok := locales[locale]; ok {
		if message, ok := localeMap[message]; ok {
			return message
		}
	}

	return "Unknown message for code: " + message
}
