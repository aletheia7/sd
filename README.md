#### gstack 
sd is a go (golang) package. sd.Journal provides methods to write to the systemd journal

#### Install 
```bash
go get github.com/aletheia7/sd
cd <sd location>
go test -v
```

#### Documentation
```bash
godoc sd 
```
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
