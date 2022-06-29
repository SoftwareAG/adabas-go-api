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
	"errors"
	"sync"
)

// Stack stack creates structure
type Stack struct {
	lock sync.Mutex // you don't have to do this if you don't want thread safety
	s    []interface{}
	Size int
}

// NewStack creates a new stack instance
func NewStack() *Stack {
	return &Stack{sync.Mutex{}, make([]interface{}, 0), 0}
}

// Push push a new element into stack
func (s *Stack) Push(v interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.s = append(s.s, v)
	s.Size++
}

// Pop pop a new element out of stack. If empty a nil interface is returned. Error is indicating the case
func (s *Stack) Pop() (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.s)
	if l == 0 {
		return nil, errors.New("empty Stack")
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	s.Size--

	return res, nil
}

// Clear Clear the stack
func (s *Stack) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.s = s.s[:0]
}
