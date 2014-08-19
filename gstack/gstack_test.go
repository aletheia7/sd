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

// Package gstack_test
package gstack_test

import (
	"testing"
	"fmt"
	. "gstack"
)

func Test_All(t *testing.T) {

	s := New()
	if s.Func() == "" {
		t.Fail()
	}
	t.Log(s)
	// Get Caller; i.e. Parent
	s = New_index(3)
	if s.Func() == "" {
		t.Fail()
	}
	t.Log(s)
}

func Example() {

	stack1 := New()			// Index: 2
	fmt.Println(stack1)
	stack2 := New_index(3)	// Index: 3 
	fmt.Println(stack2)

	// Index: 2, Function: gstack_test.Example, File: /<ommitted path ...>/go/src/gstack/gstack_test.go, Line: 27
	// Index: 3, Function: testing.runExample, File: /<ommitted path ...>/go/src/pkg/testing/example.go, Line: 98
}
