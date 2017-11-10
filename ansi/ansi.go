/*
Package ansi is a small, fast library to create ANSI colored strings and codes.

Installation

    # this installs the color viewer and the package
    go get -u github.com/mgutz/ansi/cmd/ansi-mgutz

Example

	// colorize a string, SLOW
	msg := ansi.Color("foo", "red+b:white")

	// create a closure to avoid recalculating ANSI code compilation
	phosphorize := ansi.ColorFunc("green+h:black")
	msg = phosphorize("Bring back the 80s!")
	msg2 := phospohorize("Look, I'm a CRT!")

	// cache escape codes and build strings manually
	lime := ansi.ColorCode("green+h:black")
	reset := ansi.ColorCode("reset")

	fmt.Println(lime, "Bring back the 80s!", reset)

Other examples

	Color(s, "red")            // red
	Color(s, "red+b")          // red bold
	Color(s, "red+B")          // red blinking
	Color(s, "red+u")          // red underline
	Color(s, "red+bh")         // red bold bright
	Color(s, "red:white")      // red on white
	Color(s, "red+b:white+h")  // red bold on white bright
	Color(s, "red+B:white+h")  // red blink on white bright

To view color combinations, from terminal

	ansi-mgutz

Style format

	"foregroundColor+attributes:backgroundColor+attributes"

Colors

	black
	red
	green
	yellow
	blue
	magenta
	cyan
	white

Attributes

	b = bold foreground
	B = Blink foreground
	u = underline foreground
	h = high intensity (bright) foreground, background
	i = inverse

Wikipedia ANSI escape codes [Colors](http://en.wikipedia.org/wiki/ANSI_escape_code#Colors)
*/
package ansi

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	black = iota
	red
	green
	yellow
	blue
	magenta
	cyan
	white
	defaultt = 9

	normalIntensityFG = 30
	highIntensityFG   = 90
	normalIntensityBG = 40
	highIntensityBG   = 100

	start         = "\033["
	bold          = "1;"
	blink         = "5;"
	underline     = "4;"
	inverse       = "7;"
	strikethrough = "9;"

	// Reset is the ANSI reset escape sequence
	Reset = "\033[0m"
	// DefaultBG is the default background
	DefaultBG = "\033[49m"
	// DefaultFG is the default foreground
	DefaultFG = "\033[39m"
)

var (
	plain = false
	// Colors maps common color names to their ANSI color code.
	Colors = map[string]int{
		"black":   black,
		"red":     red,
		"green":   green,
		"yellow":  yellow,
		"blue":    blue,
		"magenta": magenta,
		"cyan":    cyan,
		"white":   white,
		"default": defaultt,
		"0":       0, "1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "10": 10, "11": 11, "12": 12, "13": 13, "14": 14, "15": 15, "16": 16, "17": 17, "18": 18, "19": 19, "20": 20, "21": 21, "22": 22, "23": 23, "24": 24, "25": 25, "26": 26, "27": 27, "28": 28, "29": 29, "30": 30, "31": 31, "32": 32, "33": 33, "34": 34, "35": 35, "36": 36, "37": 37, "38": 38, "39": 39, "40": 40, "41": 41, "42": 42, "43": 43, "44": 44, "45": 45, "46": 46, "47": 47, "48": 48, "49": 49, "50": 50, "51": 51, "52": 52, "53": 53, "54": 54, "55": 55, "56": 56, "57": 57, "58": 58, "59": 59, "60": 60, "61": 61, "62": 62, "63": 63, "64": 64, "65": 65, "66": 66, "67": 67, "68": 68, "69": 69, "70": 70, "71": 71, "72": 72, "73": 73, "74": 74, "75": 75, "76": 76, "77": 77, "78": 78, "79": 79, "80": 80, "81": 81, "82": 82, "83": 83, "84": 84, "85": 85, "86": 86, "87": 87, "88": 88, "89": 89, "90": 90, "91": 91, "92": 92, "93": 93, "94": 94, "95": 95, "96": 96, "97": 97, "98": 98, "99": 99, "100": 100, "101": 101, "102": 102, "103": 103, "104": 104, "105": 105, "106": 106, "107": 107, "108": 108, "109": 109, "110": 110, "111": 111, "112": 112, "113": 113, "114": 114, "115": 115, "116": 116, "117": 117, "118": 118, "119": 119, "120": 120, "121": 121, "122": 122, "123": 123, "124": 124, "125": 125, "126": 126, "127": 127, "128": 128, "129": 129, "130": 130, "131": 131, "132": 132, "133": 133, "134": 134, "135": 135, "136": 136, "137": 137, "138": 138, "139": 139, "140": 140, "141": 141, "142": 142, "143": 143, "144": 144, "145": 145, "146": 146, "147": 147, "148": 148, "149": 149, "150": 150, "151": 151, "152": 152, "153": 153, "154": 154, "155": 155, "156": 156, "157": 157, "158": 158, "159": 159, "160": 160, "161": 161, "162": 162, "163": 163, "164": 164, "165": 165, "166": 166, "167": 167, "168": 168, "169": 169, "170": 170, "171": 171, "172": 172, "173": 173, "174": 174, "175": 175, "176": 176, "177": 177, "178": 178, "179": 179, "180": 180, "181": 181, "182": 182, "183": 183, "184": 184, "185": 185, "186": 186, "187": 187, "188": 188, "189": 189, "190": 190, "191": 191, "192": 192, "193": 193, "194": 194, "195": 195, "196": 196, "197": 197, "198": 198, "199": 199, "200": 200, "201": 201, "202": 202, "203": 203, "204": 204, "205": 205, "206": 206, "207": 207, "208": 208, "209": 209, "210": 210, "211": 211, "212": 212, "213": 213, "214": 214, "215": 215, "216": 216, "217": 217, "218": 218, "219": 219, "220": 220, "221": 221, "222": 222, "223": 223, "224": 224, "225": 225, "226": 226, "227": 227, "228": 228, "229": 229, "230": 230, "231": 231, "232": 232, "233": 233, "234": 234, "235": 235, "236": 236, "237": 237, "238": 238, "239": 239, "240": 240, "241": 241, "242": 242, "243": 243, "244": 244, "245": 245, "246": 246, "247": 247, "248": 248, "249": 249, "250": 250, "251": 251, "252": 252, "253": 253, "254": 254, "255": 255,
	}
	Black        = ColorCode("black")
	Red          = ColorCode("red")
	Green        = ColorCode("green")
	Yellow       = ColorCode("yellow")
	Blue         = ColorCode("blue")
	Magenta      = ColorCode("magenta")
	Cyan         = ColorCode("cyan")
	White        = ColorCode("white")
	LightBlack   = ColorCode("black+h")
	LightRed     = ColorCode("red+h")
	LightGreen   = ColorCode("green+h")
	LightYellow  = ColorCode("yellow+h")
	LightBlue    = ColorCode("blue+h")
	LightMagenta = ColorCode("magenta+h")
	LightCyan    = ColorCode("cyan+h")
	LightWhite   = ColorCode("white+h")
)

// ColorCode returns the ANSI color color code for style.
func ColorCode(style string) string {
	return colorCode(style).String()
}

// Gets the ANSI color code for a style.
func colorCode(style string) *bytes.Buffer {
	buf := &bytes.Buffer{}
	if plain || style == "" {
		return buf
	}
	if style == "reset" {
		buf.WriteString(Reset)
		return buf
	} else if style == "off" {
		return buf
	}

	foregroundBackground := strings.Split(style, ":")
	foreground := strings.Split(foregroundBackground[0], "+")
	fgKey := foreground[0]
	fg := Colors[fgKey]
	fgStyle := ""
	if len(foreground) > 1 {
		fgStyle = foreground[1]
	}

	bg, bgStyle := "", ""

	if len(foregroundBackground) > 1 {
		background := strings.Split(foregroundBackground[1], "+")
		bg = background[0]
		if len(background) > 1 {
			bgStyle = background[1]
		}
	}

	buf.WriteString(start)
	base := normalIntensityFG
	if len(fgStyle) > 0 {
		if strings.Contains(fgStyle, "b") {
			buf.WriteString(bold)
		}
		if strings.Contains(fgStyle, "B") {
			buf.WriteString(blink)
		}
		if strings.Contains(fgStyle, "u") {
			buf.WriteString(underline)
		}
		if strings.Contains(fgStyle, "i") {
			buf.WriteString(inverse)
		}
		if strings.Contains(fgStyle, "s") {
			buf.WriteString(strikethrough)
		}
		if strings.Contains(fgStyle, "h") {
			base = highIntensityFG
		}
	}

	// if 256-color
	n, err := strconv.Atoi(fgKey)
	if err == nil {
		fmt.Fprintf(buf, "38;5;%d;", n)
	} else {
		fmt.Fprintf(buf, "%d;", base+fg)
	}

	base = normalIntensityBG
	if len(bg) > 0 {
		if strings.Contains(bgStyle, "h") {
			base = highIntensityBG
		}
		// if 256-color
		n, err := strconv.Atoi(bg)
		if err == nil {
			fmt.Fprintf(buf, "48;5;%d;", n)
		} else {
			fmt.Fprintf(buf, "%d;", base+Colors[bg])
		}
	}

	// remove last ";"
	buf.Truncate(buf.Len() - 1)
	buf.WriteRune('m')
	return buf
}

// Color colors a string based on the ANSI color code for style.
func Color(s, style string) string {
	if plain || len(style) < 1 {
		return s
	}
	buf := colorCode(style)
	buf.WriteString(s)
	buf.WriteString(Reset)
	return buf.String()
}

// ColorFunc creates a closure to avoid computation ANSI color code.
func ColorFunc(style string) func(string) string {
	if style == "" {
		return func(s string) string {
			return s
		}
	}
	color := ColorCode(style)
	return func(s string) string {
		if plain || s == "" {
			return s
		}
		buf := bytes.NewBufferString(color)
		buf.WriteString(s)
		buf.WriteString(Reset)
		result := buf.String()
		return result
	}
}

// DisableColors disables ANSI color codes. The default is false (colors are on).
func DisableColors(disable bool) {
	plain = disable
	if plain {
		Black = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Magenta = ""
		Cyan = ""
		White = ""
		LightBlack = ""
		LightRed = ""
		LightGreen = ""
		LightYellow = ""
		LightBlue = ""
		LightMagenta = ""
		LightCyan = ""
		LightWhite = ""
	} else {
		Black = ColorCode("black")
		Red = ColorCode("red")
		Green = ColorCode("green")
		Yellow = ColorCode("yellow")
		Blue = ColorCode("blue")
		Magenta = ColorCode("magenta")
		Cyan = ColorCode("cyan")
		White = ColorCode("white")
		LightBlack = ColorCode("black+h")
		LightRed = ColorCode("red+h")
		LightGreen = ColorCode("green+h")
		LightYellow = ColorCode("yellow+h")
		LightBlue = ColorCode("blue+h")
		LightMagenta = ColorCode("magenta+h")
		LightCyan = ColorCode("cyan+h")
		LightWhite = ColorCode("white+h")
	}
}
