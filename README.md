#### sd 
sd is a go (golang) package. sd.Journal provides methods to write to the systemd journal

#### Install 
```bash
go get github.com/aletheia7/sd
cd <sd location>
go test -v
```

#### Documentation
See: [godoc sd](http://godoc.org/github.com/aletheia7/sd) 

New_journal() and New_journal_m() create a Journal struct. Journal.Emerg(), Journal.Alert(), Journal.Crit(), Journal.Err(), Journal.Warning(), Journal.Notice(), Journal.Info(), Journal.Debug() write to the systemd journal. Each method contains a *_m (map variation) method that allows sending your own fields. The map suppports string and []byte (binary). Each method also contains a _m_f (map & format variation) method that supports fmt.Printf style arguments. Each of the methods will add journal fields GO_FILE, GO_LINE, and GO_FUNC fields to the journal to indicate where the methods were called. The *_m_f methods can take nil map in order to only use the format functionality.
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
