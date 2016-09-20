// Copyright 2016 aletheia7. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause
// license that can be found in the LICENSE file.

// Package gstack provides methods to easily obtain the function name, file path, and line number of go code.
//
// gstack answers the questions: What function am I in? What is the full path
// to the file? What is the line number?
package gstack

import (
	"fmt"
	"runtime"
	"strconv"
)

type Stack struct {
	index     int
	func_name string
	file_name string
	line      int
}

// New() returns a Stack based on the immediate function scope.
//
// Implicit index = 2. Index = 1 is not very useful. It will be New().
func New() *Stack {
	return get_stack(2)
}

// New_index returns a Stack.
//
// When index = 2, Stack is the parent function caller scope; i.e. the function that
// called gstack.New()
//
// When index = 3, Stack is the next level up.
func New_index(index int) *Stack {
	return get_stack(index)
}

func get_stack(index int) *Stack {

	if pc, _, _, ok := runtime.Caller(index); ok {
		pc = pc - 1
		f := runtime.FuncForPC(pc)
		name := f.Name()
		file, line := f.FileLine(pc)
		return &Stack{index: index, func_name: name, file_name: file, line: line}
	}
	return &Stack{index: index}
}

// Return the function name of the function call.
func (s *Stack) Func() string {
	return s.func_name
}

// Return the file name of the function call.
func (s *Stack) File() string {
	return s.file_name
}

// Return the line number of the function call.
func (s *Stack) Line() int {
	return s.line
}

// Return the line number of the function call as a string
func (s *Stack) Line_s() string {
	return strconv.FormatInt(int64(s.line), 10)
}

// Stringer Interface
func (s *Stack) String() string {
	return fmt.Sprintf("Index: %d, Function: %s, File: %s, Line: %d", s.index, s.func_name, s.file_name, s.line)
}
