[![](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/aletheia7/sd) 

#### Install 
```bash
go get github.com/aletheia7/sd
cd <sd location>
go test -v
```
Older systemd versions used libsystemd-journal. Change the following line if
you have libsystemd-journal:

```go
// #cgo pkg-config: --cflags --libs libsystemd
```
to
```go
// #cgo pkg-config: --cflags --libs libsystemd-journal
``` 

#### Documentation

New_journal() and New_journal_m() create a Journal struct. Journal.Emerg(), 
Journal.Alert(), Journal.Crit(), Journal.Err(), Journal.Warning(),
Journal.Notice(), Journal.Info(), Journal.Debug() write to the systemd journal.

Each method contains a format *f method; i.e. Infof, Errf.

Each method also contains a *_a (array variation) method that allows sending your
own fields as an array of SOMEFILE=value strings. An *_a_f variation supports
fmt.Printf style arguments.

Each method contains a *_m (map variation) method that allows sending your own
fields. The map suppports string and []byte (binary).

Each method also contains a *_m_f (map & format variation) method that supports
[fmt.Printf](http://godoc.org/fmt#Printf) style arguments.

Each of the methods will add journal fields GO_FILE, and GO_FUNC fields to the
journal to indicate where the methods were called. The *_m_f methods can take
nil map in order to only use the format functionality.

#### Helpful Hints
+ You may need to increase RateLimitInterval and/or RateLimitBurst settings in
journald.conf when sending large amounts of data to the journal. Data will
not appear in the log when settings are too low. 

* journalctl will truncate output by default. journalctl uses `less` as it's
pager and sets it's own defaults via SYSTEMD_LESS. To restore the output, set
the SYSTEMD_LESS environment variable; i.e. `export SYSTEMD_LESS=FRXMK`. See
`man journalctl`.

#### Example

```go
package main
import (
	"io"
	"log"
	"sd"
)

func main() {

	j := sd.New_journal()
	j.Alert("Alert example")

	// COMMENT_2_BINARY = abcNULLabc
	m := map[string]interface{}{"COMMENT_1": "This function ran successfully",
		"COMMENT_2_BINARY": []byte{0x61, 0x62, 0x63, 0x00, 0x61, 0x62, 0x63},
	}

	// Use: "journal -f --output verbose" to see fields
	j.Alert_m(m, "Alert_m exmaple")

	j.Alert_m_f(m, "Alert_m_f example: Salary: %v, Year: %v", 0.00, 2014)

	// Use log package
	// Remove ANSI escape sequences
	// systemd will convert messages to binary with ANSI escapes sequences
	sd.Set_default_remove_ansi_escape(true)
	j := sd.New_journal()
	// systemd will output red text
	j.Set_writer_priority(sd.Log_err)
	// Hello World
	// Hello is green
	// World is yellow
	s := "\x1b[32mHello\x1b[0m \x1b[93mWorld\x1b[0m"
	log.SetFlags(0)
	// Send to stderr and systemd-journald
	log.SetOutput(io.MultiWriter(os.Stderr, j))
	log.Println(s)
}
```
##### Ouput
```bash
# tail the systemd journal 
journal -f --output verbose
```

#### License 

Use of this source code is governed by a BSD-2-Clause license that can be found
in the LICENSE file.

[![BSD-2-Clause License](img/osi_logo_100X133_90ppi_0.png)](https://opensource.org/)
