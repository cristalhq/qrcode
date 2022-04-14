package coding

import (
	"fmt"
	"strings"
)

// Num is the encoding for numeric data.
// The only valid characters are the decimal digits 0 through 9.
type Num string

func (s Num) String() string {
	return fmt.Sprintf("Num(%#q)", string(s))
}

func (s Num) Check() bool {
	for _, c := range s {
		if c < '0' || '9' < c {
			return false
		}
	}
	return true
}

var numLen = [3]int{10, 12, 14}

func (s Num) Bits(v Version) int {
	return 4 + numLen[v.sizeClass()] + (10*len(s)+2)/3
}

func (s Num) Encode(b *Bits, v Version) {
	b.Write(1, 4)
	b.Write(uint(len(s)), numLen[v.sizeClass()])
	var i int
	for i = 0; i+3 <= len(s); i += 3 {
		w := uint(s[i]-'0')*100 + uint(s[i+1]-'0')*10 + uint(s[i+2]-'0')
		b.Write(w, 10)
	}
	switch len(s) - i {
	case 1:
		w := uint(s[i] - '0')
		b.Write(w, 4)
	case 2:
		w := uint(s[i]-'0')*10 + uint(s[i+1]-'0')
		b.Write(w, 7)
	}
}

// Alpha is the encoding for alphanumeric data.
// The valid characters are 0-9A-Z$%*+-./: and space.
type Alpha string

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

func (s Alpha) String() string {
	return fmt.Sprintf("Alpha(%#q)", string(s))
}

func (s Alpha) Check() bool {
	for _, c := range s {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}
	return true
}

var alphaLen = [3]int{9, 11, 13}

func (s Alpha) Bits(v Version) int {
	return 4 + alphaLen[v.sizeClass()] + (11*len(s)+1)/2
}

func (s Alpha) Encode(b *Bits, v Version) {
	b.Write(2, 4)
	b.Write(uint(len(s)), alphaLen[v.sizeClass()])
	var i int
	for i = 0; i+2 <= len(s); i += 2 {
		w := uint(strings.IndexRune(alphabet, rune(s[i])))*45 +
			uint(strings.IndexRune(alphabet, rune(s[i+1])))
		b.Write(w, 11)
	}

	if i < len(s) {
		w := uint(strings.IndexRune(alphabet, rune(s[i])))
		b.Write(w, 6)
	}
}

// String is the encoding for 8-bit data. All bytes are valid.
type String string

func (s String) String() string {
	return fmt.Sprintf("String(%#q)", string(s))
}

func (s String) Check() bool { return true }

var stringLen = [3]int{8, 16, 16}

func (s String) Bits(v Version) int {
	return 4 + stringLen[v.sizeClass()] + 8*len(s)
}

func (s String) Encode(b *Bits, v Version) {
	b.Write(4, 4)
	b.Write(uint(len(s)), stringLen[v.sizeClass()])
	for i := 0; i < len(s); i++ {
		b.Write(uint(s[i]), 8)
	}
}
