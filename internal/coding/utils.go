package coding

import (
	"strconv"

	"github.com/cristalhq/qrcode/internal/gf256"
)

// A Level represents a QR error correction level.
// From least to most tolerant of errors, they are L, M, Q, H.
type Level int

const (
	L Level = 0
	M Level = 1
	Q Level = 2
	H Level = 3
)

func (l Level) String() string {
	switch l {
	case L:
		return "L"
	case M:
		return "M"
	case Q:
		return "Q"
	case H:
		return "H"
	default:
		return strconv.Itoa(int(l))
	}
}

// A Version represents a QR version.
// The version specifies the size of the QR code:
// a QR code with version v has 4v+17 pixels on a side.
// Versions number from 1 to 40: the larger the version,
// the more information the code can store.
type Version int

const (
	MinVersion Version = 1
	MaxVersion Version = 40
)

func (v Version) String() string {
	return strconv.Itoa(int(v))
}

// DataBytes returns the number of data bytes that can be
// stored in a QR code with the given version and level.
func (v Version) DataBytes(l Level) int {
	vt := &vtab[v]
	lev := &vt.level[l]
	return vt.bytes - lev.nblock*lev.check
}

func (v Version) sizeClass() int {
	switch {
	case v <= 9:
		return 0
	case v <= 26:
		return 1
	default:
		return 2
	}
}

type Bits struct {
	b    []byte
	nbit int
}

func (b *Bits) Reset() {
	b.b = b.b[:0]
	b.nbit = 0
}

func (b *Bits) Bits() int {
	return b.nbit
}

func (b *Bits) Bytes() []byte {
	if b.nbit%8 != 0 {
		panic("qr: fractional byte")
	}
	return b.b
}

func (b *Bits) Append(p []byte) {
	if b.nbit%8 != 0 {
		panic("qr: fractional byte")
	}
	b.b = append(b.b, p...)
	b.nbit += 8 * len(p)
}

func (b *Bits) Pad(n int) {
	if n < 0 {
		panic("qr: invalid pad size")
	}
	if n <= 4 {
		b.Write(0, n)
	} else {
		b.Write(0, 4)
		n -= 4
		n -= -b.Bits() & 7
		b.Write(0, -b.Bits()&7)
		pad := n / 8
		for i := 0; i < pad; i += 2 {
			b.Write(0xec, 8)
			if i+1 >= pad {
				break
			}
			b.Write(0x11, 8)
		}
	}
}

// qrField is the field for QR error correction.
var qrField = gf256.NewField(0x11d, 2)

func (b *Bits) AddCheckBytes(v Version, l Level) {
	nd := v.DataBytes(l)
	if b.nbit < nd*8 {
		b.Pad(nd*8 - b.nbit)
	}
	if b.nbit != nd*8 {
		panic("qr: too much data")
	}

	dat := b.Bytes()
	vt := &vtab[v]
	lev := &vt.level[l]
	db := nd / lev.nblock
	extra := nd % lev.nblock
	chk := make([]byte, lev.check)
	rs := gf256.NewRSEncoder(qrField, lev.check)
	for i := 0; i < lev.nblock; i++ {
		if i == lev.nblock-extra {
			db++
		}
		rs.ECC(dat[:db], chk)
		b.Append(chk)
		dat = dat[db:]
	}

	if len(b.Bytes()) != vt.bytes {
		panic("qr: internal error")
	}
}

func (b *Bits) Write(v uint, nbit int) {
	for nbit > 0 {
		n := nbit
		if n > 8 {
			n = 8
		}
		if b.nbit%8 == 0 {
			b.b = append(b.b, 0)
		} else {
			m := -b.nbit & 7
			if n > m {
				n = m
			}
		}
		b.nbit += n
		sh := uint(nbit - n)
		b.b[len(b.b)-1] |= uint8(v >> sh << uint(-b.nbit&7))
		v -= v >> sh << sh
		nbit -= n
	}
}

const (
	Black Pixel = 1 << iota
	Invert
)

// A Pixel describes a single pixel in a QR code.
type Pixel uint32

func OffsetPixel(o uint) Pixel {
	return Pixel(o << 6)
}

func (p Pixel) Offset() uint {
	return uint(p >> 6)
}

func (p Pixel) Role() PixelRole {
	return PixelRole(p>>2) & 15
}

func (p Pixel) String() string {
	s := p.Role().String()
	if p&Black != 0 {
		s += "+black"
	}
	if p&Invert != 0 {
		s += "+invert"
	}
	s += "+" + strconv.FormatUint(uint64(p.Offset()), 10)
	return s
}

// A PixelRole describes the role of a QR pixel.
type PixelRole uint32

const (
	_         PixelRole = iota
	Position            // position squares (large)
	Alignment           // alignment squares (small)
	Timing              // timing strip between position squares
	Format              // format metadata
	PVersion            // version pattern
	Unused              // unused pixel
	Data                // data bit
	Check               // error correction check bit
	Extra
)

func (r PixelRole) Pixel() Pixel {
	return Pixel(r << 2)
}

var roles = []string{
	"",
	"position",
	"alignment",
	"timing",
	"format",
	"pversion",
	"unused",
	"data",
	"check",
	"extra",
}

func (r PixelRole) String() string {
	if Position <= r && r <= Check {
		return roles[r]
	}
	return strconv.Itoa(int(r))
}

// A Mask describes a mask that is applied to the QR
// code to avoid QR artifacts being interpreted as
// alignment and timing patterns (such as the squares
// in the corners). Valid masks are integers from 0 to 7.
type Mask int

func (m Mask) Invert(y, x int) bool {
	if m < 0 {
		return false
	}
	return mfunc[m](y, x)
}

// http://www.swetake.com/qr/qr5_en.html
var mfunc = [...]func(int, int) bool{
	func(i, j int) bool { return (i+j)%2 == 0 },
	func(i, j int) bool { return i%2 == 0 },
	func(i, j int) bool { return j%3 == 0 },
	func(i, j int) bool { return (i+j)%3 == 0 },
	func(i, j int) bool { return (i/2+j/3)%2 == 0 },
	func(i, j int) bool { return i*j%2+i*j%3 == 0 },
	func(i, j int) bool { return (i*j%2+i*j%3)%2 == 0 },
	func(i, j int) bool { return (i*j%3+(i+j)%2)%2 == 0 },
}
