// Copyright 2016 aletheia7. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause
// license that can be found in the LICENSE file.

// Package gstack_test
package gstack_test

import (
	"fmt"
	. "gstack"
	"testing"
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

	stack1 := New() // Index: 2
	fmt.Println(stack1)
	stack2 := New_index(3) // Index: 3
	fmt.Println(stack2)

	// Index: 2, Function: gstack_test.Example, File: /<ommitted path ...>/go/src/gstack/gstack_test.go, Line: 27
	// Index: 3, Function: testing.runExample, File: /<ommitted path ...>/go/src/pkg/testing/example.go, Line: 98
}
