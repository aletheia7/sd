// Copyright 2014 aletheia7.
//
// This file is part of gstack.
//
// gstack is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// gstack is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with gstack.  If not, see <http://www.gnu.org/licenses/>.

// Package gstack provides methods to easily obtain the function name, file path, and line number of go code.
//
// gstack answers the questions: What function am I in? What is the full path
// to the file? What is the line number?
package gstack 
import (
	"runtime"
	"strconv"
	"fmt"
)

type Stack struct {

	index		int
	func_name	string
	file_name	string
	line		int
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
