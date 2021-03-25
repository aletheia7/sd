// Copyright 2016 aletheia7. All rights reserved. Use of this source code is
// governed by a BSD-2-Clause license that can be found in the LICENSE file.

// Package sd_test tests the package sd
package sd_test

import (
	. "github.com/aletheia7/sd"
	"testing"
)

func Test_Info(t *testing.T) {
	j := New_journal()
	if err := j.Info("Info test"); err != nil {
		t.Error(err)
	}
}

func Test_Info_m(t *testing.T) {
	j := New_journal()
	if err := j.Info_m(nil, "Info test"); err != nil {
		t.Error(err)
	}
}

func Test_Info_m_f(t *testing.T) {
	j := New_journal()
	m := map[string]interface{}{"USER_DATA": `yikes, what happened`, "USER_BYTES": string([]byte{0x4a, 0x65, 0x73, 0x75, 0x73, 0x20, 0x64, 0x69, 0x65, 0x64, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x79, 0x6f, 0x75, 0x72, 0x20, 0x73, 0x69, 0x6e, 0x2c, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x79, 0x6f, 0x75, 0x21, 0x20, 0x47, 0x6f, 0x64, 0x20, 0x42, 0x6c, 0x65, 0x73, 0x73, 0x2e})}
	if err := j.Info_m_f(m, "Info test with args: %s %d", "more data", 123); err != nil {
		t.Error(err)
	}
}
