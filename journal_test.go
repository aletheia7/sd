// Copyright 2014 aletheia7.
//
// This file is part of sd.
//
// sd is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// sd is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with sd.  If not, see <http://www.gnu.org/licenses/>.

// Package sd_test tests the package sd
package sd_test

import (
	"testing"
	. "sd"
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
