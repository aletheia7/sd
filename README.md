#### sd 
Go package that provides methods to write to the systemd journal

#### Install 
```bash
go get github.com/aletheia7/sd
cd <sd location>
go test -v
```

#### Documentation
[godoc sd](http://godoc.org/github.com/aletheia7/sd) 

New_journal() and New_journal_m() create a Journal struct. Journal.Emerg(), 
Journal.Alert(), Journal.Crit(), Journal.Err(), Journal.Warning(),
Journal.Notice(), Journal.Info(), Journal.Debug() write to the systemd journal.E

Each method contains a *_m (map variation) method that allows sending your own
fields. The map suppports string and []byte (binary).

Each method also contains a *_m_f (map & format variation) method that supports
[fmt.Printf](http://godoc.org/fmt#Printf) style arguments.

Each method also contains a *_a (array variation) method that allows sending your
own fields as an array of SOMEFILE=value strings. An *_a_f variation supports
fmt.Printf style arguments.

Each of the methods will add journal fields GO_FILE, GO_LINE, and GO_FUNC fieldsi
 to the journal to indicate where the methods were called. The *_m_f methods
 can take nil map in order to only use the format functionality.

#### Helpful Hints
+ You may need to increase RateLimitInterval and/or RateLimitBurst settings in
journald.conf when sending large amounts of data to the journal. Data will
not appear in the log when settings are too low. 
+ This package is gorouttine safe, however problems have occurred when 
[runtime.GOMAXPROCS](http://godoc.org/runtime#GOMAXPROCS) is used.
#### Example

```go
package main
import (
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
}
```
##### Ouput
```bash
# tail the systemd journal 
journal -f --output verbose
```

![LGPL](http://www.gnu.org/graphics/lgplv3-147x51.png)
