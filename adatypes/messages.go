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
	"os"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
)

type errorCode int

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
	args    []interface{}
}

// Error error interface function, providing message error code and message. The Adabas error provides
// message code and message text
func (adaErr *Error) Error() string {
	return fmt.Sprintf("%v: %v", adaErr.Code, adaErr.Message)
}

// Language current message language
func Language() string {
	lang := os.Getenv("LANG")
	switch {
	case lang == "":
		lang = "en"
	default:
		if len(lang) < 2 {
			lang = "en"
		} else {
			lang = lang[0:2]
		}
	}
	Central.Log.Debugf("Current LANG: %s", lang)
	return lang
}

// NewGenericError create a genernic non Adabas response error
func NewGenericError(code errorCode, args ...interface{}) *Error {
	Central.Log.Debugf("Generate generic error for error code %d", code)
	msgCode := fmt.Sprintf("ADG%07d", code)
	// fmt.Printf("Generated out of %d -> %s\n", code, msgCode)
	msg := Translate(Language(), msgCode)
	if msg == "" {
		msg = "Unknown message for code: " + msgCode
	}
	for i, x := range args {
		m := fmt.Sprintf("%v", x)
		c := fmt.Sprintf("\\{%d\\}", i)
		re := regexp.MustCompile(c)
		msg = re.ReplaceAllString(msg, m)
	}
	if Central.IsDebugLevel() {
		Central.Log.Debugf("Generic error message created:[%s] %s", msgCode, msg)
		Central.Log.Debugf("Stack trace:\n%s", string(debug.Stack()))
	}
	return &Error{When: time.Now(), Code: msgCode, Message: msg, args: args}
}

// Translate translate to language
func (adaErr *Error) Translate(lang string) string {
	msg := Translate(lang, adaErr.Code)
	if msg == "" {
		msg = "Unknown message for code: " + adaErr.Code
	}
	for i, x := range adaErr.args {
		m := fmt.Sprintf("%v", x)
		c := fmt.Sprintf("\\{%d\\}", i)
		re := regexp.MustCompile(c)
		msg = re.ReplaceAllString(msg, m)
	}
	return msg
}

// initMessages loads messages from all message files on the given path.
func initMessages() *Error {
	once.Do(onceBody)
	return loadErr
}

// LoadMessages loads messages from all message files on the given path.
func loadMessagesOnce() *Error {
	Central.Log.Debugf("Load messages")
	locales = make(map[string]map[string]string)
	for _, m := range staticMessages {
		var messages map[string]string
		if messageMap, ok := locales[m.locale]; ok {
			messages = messageMap
		} else {
			messages = make(map[string]string)
			locales[m.locale] = messages
		}
		messages[m.code] = m.message
	}
	Central.Log.Debugf("Loaded messages: %d", len(staticMessages))

	return nil
}

// Translate translates content to target language.
func Translate(locale, message string, args ...interface{}) string {
	if err := initMessages(); err != nil {
		return "Error initialize Adabas messages"
	}
	if localeMap, ok := locales[locale]; ok {
		if message, ok := localeMap[message]; ok {
			Central.Log.Debugf("Found %s message: %s", locale, message)
			return message
		}
		Central.Log.Debugf("Message %s for locale %s not found", message, locale)
	}

	// If no message found, use the english message
	if locale != "en" {
		Central.Log.Debugf("Try locale en")
		if localeMap, ok := locales["en"]; ok {
			if message, ok := localeMap[message]; ok {
				return message
			}
		}
	}
	return ""
}
