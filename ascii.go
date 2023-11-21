package qrcode

import (
	"bytes"
	"fmt"
)

func (c *Code) ASCII() string {
	var w asciiWriter
	return w.encode(c)
}

type asciiWriter struct {
	bytes.Buffer
}

func (wr *asciiWriter) encode(code *Code) string {
	wr.WriteString(reset)
	wr.border(code)

	for x := 0; x < code.Size; x++ {
		wr.WriteString(white + space)
		for y := 0; y < code.Size; y++ {
			c := white
			if code.IsBlack(x, y) {
				c = black
			}
			wr.WriteString(c + "  ")
		}
		wr.WriteString(white + space + reset + "\n")
	}

	wr.border(code)
	return wr.String()
}

func (wr *asciiWriter) border(code *Code) {
	for i := 0; i < 2; i++ {
		wr.WriteString(white + space)
		for x := 0; x < code.Size; x++ {
			fmt.Fprint(wr, "  ")
		}
		wr.WriteString(space + reset + "\n")
	}
}

const (
	space = "    "
	black = "\033[30;40m"
	white = "\033[30;47m"
	reset = "\033[0m"
)
