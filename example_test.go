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

package sd_test

import (
	"sd"
)

func ExampleJournal() {

	j := sd.New_journal()
	j.Alert("Alert example")

	// COMMENT_2_BINARY = abcNULLabc
	m := map[string]interface{}{"COMMENT_1": "This function ran successfully",
			"COMMENT_2_BINARY": []byte{0x61, 0x62, 0x63, 0x00, 0x61, 0x62, 0x63},
	}

	// Use: "journal -f --output verbose" to see fields
	j.Alert_m(m, "Alert_m exmaple")

	j.Alert_m_f(m, "Alert_m_f example: Salary: %v, Year: %v", 0.00, 2014)
}
