package ansi_test

import (
	"fmt"
	. "github.com/aletheia7/sd/ansi"
	"sort"
	"strings"
	"testing"
)

func TestPlain(t *testing.T) {
	DisableColors(true)
	PrintStyles()
}

func TestStyles(t *testing.T) {
	DisableColors(false)
	PrintStyles()
}

func TestDisableColors(t *testing.T) {
	fn := ColorFunc("red")

	buf := ColorCode("off")
	if buf != "" {
		t.Fail()
	}

	DisableColors(true)
	if Black != "" {
		t.Fail()
	}
	code := ColorCode("red")
	if code != "" {
		t.Fail()
	}
	s := fn("foo")
	if s != "foo" {
		t.Fail()
	}

	DisableColors(false)
	if Black == "" {
		t.Fail()
	}
	code = ColorCode("red")
	if code == "" {
		t.Fail()
	}
	// will have escape codes around it
	index := strings.Index(fn("foo"), "foo")
	if index <= 0 {
		t.Fail()
	}
}

// PrintStyles prints all style combinations to the terminal.
func PrintStyles() {
	bgColors := []string{
		"",
		":black",
		":red",
		":green",
		":yellow",
		":blue",
		":magenta",
		":cyan",
		":white",
	}

	keys := make([]string, 0, len(Colors))
	for k := range Colors {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, fg := range keys {
		for _, bg := range bgColors {
			fmt.Println(padColor(fg, []string{"" + bg, "+b" + bg, "+bh" + bg, "+u" + bg}))
			fmt.Println(padColor(fg, []string{"+s" + bg, "+i" + bg}))
			fmt.Println(padColor(fg, []string{"+uh" + bg, "+B" + bg, "+Bb" + bg /* backgrounds */, "" + bg + "+h"}))
			fmt.Println(padColor(fg, []string{"+b" + bg + "+h", "+bh" + bg + "+h", "+u" + bg + "+h", "+uh" + bg + "+h"}))
		}
	}
}

func pad(s string, length int) string {
	for len(s) < length {
		s += " "
	}
	return s
}

func padColor(color string, styles []string) string {
	buffer := ""
	for _, style := range styles {
		buffer += Color(pad(color+style, 20), color+style)
	}
	return buffer
}
