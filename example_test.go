// Copyright 2016 aletheia7. All rights reserved. Use of this source code is
// governed by a BSD-2-Clause license that can be found in the LICENSE file.

// Package sd_test provides an example of package sd
package sd_test

import (
	"github.com/aletheia7/sd"
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
