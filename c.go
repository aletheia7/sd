// Copyright 2016 aletheia7. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause
// license that can be found in the LICENSE file.

// +build linux,cgo

package sd

/*
#cgo pkg-config: --cflags --libs libsystemd
#include <stdlib.h>
#include <systemd/sd-journal.h>
#include <unistd.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"unsafe"
)

func init() {
	max_fields = uint64(C.sysconf(C._SC_IOV_MAX))
	valid_field, _ = regexp.Compile(sd_valid_field_regexp)
}

var (
	valid_field         *regexp.Regexp
	max_fields          uint64
	sd_field_name_sep_s = string(sd_field_name_sep_b)
	sd_field_name_sep_b = []byte{61}
)

// Send writes to the systemd-journal. The keys must be uppercase strings
// without a leading _. The other send methods are easier to use. See Info(),
// Infom(), Info_m_f(), etc. A MESSAGE key in field is the only required
// field.
//
func (j *Journal) Send(fields map[string]interface{}) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	if max_fields < uint64(len(fields)) {
		return errors.New(fmt.Sprintf("Field count cannot exceed %v: %v given", max_fields, len(fields)))
	}
	if j.add_go_code_fields {
		st := new_index(4)
		fields[sd_go_func] = st.Func()
		fields[sd_go_file] = st.File() + `:` + st.Line_s()
	}
	iov := C.malloc(C.size_t(C.sizeof_struct_iovec * len(fields)))
	i := 0
	defer func() {
		for j := 0; j < i; j++ {
			C.free(((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(j)*C.sizeof_struct_iovec))).iov_base)
		}
		C.free(iov)
	}()
	for k, v := range fields {
		if valid_field.FindString(k) == "" {
			return fmt.Errorf("field violates regexp %v : %v", valid_field, k)
		}
		switch t := v.(type) {
		case string:
			s := k + sd_field_name_sep_s + t
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = unsafe.Pointer(C.CString(s))
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(s))
		case Priority:
			s := k + sd_field_name_sep_s + string(t)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = unsafe.Pointer(C.CString(s))
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(s))
		case []byte:
			b := bytes.Join([][]byte{[]byte(k), t}, sd_field_name_sep_b)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = C.CBytes(b)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(b))
		default:
			return fmt.Errorf("Error: Unsupported field value: key = %v", k)
		}
		i++
	}
	switch {
	case j.send_stderr != Sd_send_stderr_allow_override:
		if j.send_stderr == Sd_send_stderr_true {
			fmt.Fprintf(os.Stderr, "%v", fields[Sd_message])
		}
	case default_send_stderr == Sd_send_stderr_true:
		fmt.Fprintf(os.Stderr, "%v", fields[Sd_message])
	}
	n, _ := C.sd_journal_sendv((*C.struct_iovec)(iov), C.int(len(fields)))
	if n != 0 {
		return errors.New("Error with sd_journal_sendv arguments")
	}
	return nil
}

type stack struct {
	index     int
	func_name string
	file_name string
	line      int
}

// New() returns a Stack based on the immediate function scope.
//
// Implicit index = 2. Index = 1 is not very useful. It will be New().
func new() *stack {
	return get_stack(2)
}

// New_index returns a Stack.
//
// When index = 2, Stack is the parent function caller scope; i.e. the function that
// called gstack.New()
//
// When index = 3, Stack is the next level up.
func new_index(index int) *stack {
	return get_stack(index)
}

func get_stack(index int) *stack {

	if pc, _, _, ok := runtime.Caller(index); ok {
		pc = pc - 1
		f := runtime.FuncForPC(pc)
		name := f.Name()
		file, line := f.FileLine(pc)
		return &stack{index: index, func_name: name, file_name: file, line: line}
	}
	return &stack{index: index}
}

// Return the function name of the function call.
func (s *stack) Func() string {
	return s.func_name
}

// Return the file name of the function call.
func (s *stack) File() string {
	return s.file_name
}

// Return the line number of the function call.
func (s *stack) Line() int {
	return s.line
}

// Return the line number of the function call as a string
func (s *stack) Line_s() string {
	return strconv.Itoa(s.line)
}

// Stringer Interface
func (s *stack) String() string {
	return fmt.Sprintf("Index: %d, Function: %s, File: %s, Line: %d", s.index, s.func_name, s.file_name, s.line)
}
