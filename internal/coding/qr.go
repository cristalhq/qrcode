// Package coding implements low-level QR coding details.
package coding

import (
	"errors"
	"fmt"
)

func Encode(bitmap []byte, text string, level Level) (*Code, error) {
	var enc Encoding
	switch {
	case Num(text).Check():
		enc = Num(text)
	case Alpha(text).Check():
		enc = Alpha(text)
	default:
		enc = String(text)
	}

	version := MinVersion
	for {
		if version > MaxVersion {
			return nil, errors.New("text too long to encode as QR")
		}
		if enc.Bits(version) <= version.DataBytes(level)*8 {
			break
		}
		version++
	}
	return NewPlan(version, level, 0).EncodeInto(bitmap, enc)
}

// Encoding implements a QR data encoding scheme.
// The implementations--Numeric, Alphanumeric, and String--specify
// the character set and the mapping from UTF-8 to code bits.
// The more restrictive the mode, the fewer code bits are needed.
type Encoding interface {
	Check() bool
	Bits(v Version) int
	Encode(b *Bits, v Version)
}

// A Code is a square pixel grid.
type Code struct {
	Bitmap []byte // 1 is black, 0 is white
	Size   int    // number of pixels on a side
	Stride int    // number of bytes per row
}

// func (c *Code) Black(x, y int) bool {
// 	return 0 <= x && x < c.Size && 0 <= y && y < c.Size &&
// 		c.Bitmap[y*c.Stride+x/8]&(1<<uint(7-x&7)) != 0
// }

// A Plan describes how to construct a QR code
// with a specific version, level, and mask.
type Plan struct {
	Version Version
	Level   Level
	Mask    Mask

	DataBytes  int // number of data bytes
	CheckBytes int // number of error correcting (checksum) bytes
	Blocks     int // number of data blocks

	Pixel [][]Pixel // pixel map
}

// NewPlan returns a Plan for a QR code with the given version, level, and mask.
func NewPlan(version Version, level Level, mask Mask) *Plan {
	p := &Plan{
		Version: version,
		Level:   level,
		Mask:    mask,
	}
	p.vplan()
	p.fplan()
	p.lplan()
	p.mplan()
	return p
}

func (p *Plan) Encode(text Encoding) (*Code, error) {
	return p.EncodeInto(nil, text)
}

func (p *Plan) EncodeInto(bitmap []byte, text Encoding) (*Code, error) {
	if !text.Check() {
		return nil, errors.New("encoding not supported")
	}

	var b Bits
	text.Encode(&b, p.Version)

	if b.Bits() > p.DataBytes*8 {
		return nil, fmt.Errorf("cannot encode %d bits into %d-bit code", b.Bits(), p.DataBytes*8)
	}
	b.AddCheckBytes(p.Version, p.Level)
	bytes := b.Bytes()

	// Now we have the checksum bytes and the data bytes.
	// Construct the actual code.
	c := &Code{
		Size:   len(p.Pixel),
		Stride: (len(p.Pixel) + 7) &^ 7,
	}
	if bitmap == nil {
		bitmap = make([]byte, c.Stride*c.Size)
	}
	c.Bitmap = bitmap[:c.Stride*c.Size]

	crow := c.Bitmap
	for _, row := range p.Pixel {
		for x, pix := range row {
			switch pix.Role() {
			case Data, Check:
				o := pix.Offset()
				if bytes[o/8]&(1<<(7-o&7)) != 0 {
					pix ^= Black
				}
			}
			if pix&Black != 0 {
				crow[x/8] |= 1 << uint(7-x&7)
			}
		}
		crow = crow[c.Stride:]
	}
	return c, nil
}

func grid(siz int) [][]Pixel {
	m := make([][]Pixel, siz)
	pix := make([]Pixel, siz*siz)
	for i := range m {
		m[i], pix = pix[:siz], pix[siz:]
	}
	return m
}

// vplan creates a Plan for the given version.
func (p *Plan) vplan() {
	v := p.Version
	if v < MinVersion || v > MaxVersion {
		panic(fmt.Sprintf("qr: invalid QR version %d", int(v)))
	}
	siz := 17 + int(v)*4
	m := grid(siz)
	p.Pixel = m

	// Timing markers (overwritten by boxes).
	const ti = 6 // timing is in row/column 6 (counting from 0)
	for i := range m {
		p := Timing.Pixel()
		if i&1 == 0 {
			p |= Black
		}
		m[i][ti] = p
		m[ti][i] = p
	}

	// Position boxes.
	posBox(m, 0, 0)
	posBox(m, siz-7, 0)
	posBox(m, 0, siz-7)

	// Alignment boxes.
	info := &vtab[v]
	for x := 4; x+5 < siz; {
		for y := 4; y+5 < siz; {
			// don't overwrite timing markers
			if (x < 7 && y < 7) || (x < 7 && y+5 >= siz-7) || (x+5 >= siz-7 && y < 7) {
				// skip
			} else {
				alignBox(m, x, y)
			}
			if y == 4 {
				y = info.apos
			} else {
				y += info.astride
			}
		}
		if x == 4 {
			x = info.apos
		} else {
			x += info.astride
		}
	}

	// Version pattern.
	pat := vtab[v].pattern
	if pat != 0 {
		v := pat
		for x := 0; x < 6; x++ {
			for y := 0; y < 3; y++ {
				p := PVersion.Pixel()
				if v&1 != 0 {
					p |= Black
				}
				m[siz-11+y][x] = p
				m[x][siz-11+y] = p
				v >>= 1
			}
		}
	}

	// One lonely black pixel
	m[siz-8][8] = Unused.Pixel() | Black
}

// fplan adds the format pixels.
func (p *Plan) fplan() {
	// Format pixels.
	fb := uint32(p.Level^1) << 13 // level: L=01, M=00, Q=11, H=10
	fb |= uint32(p.Mask) << 10    // mask
	const formatPoly = 0x537
	rem := fb
	for i := 14; i >= 10; i-- {
		if rem&(1<<uint(i)) != 0 {
			rem ^= formatPoly << uint(i-10)
		}
	}
	fb |= rem
	invert := uint32(0x5412)
	siz := len(p.Pixel)
	for i := uint(0); i < 15; i++ {
		pix := Format.Pixel() + OffsetPixel(i)
		if (fb>>i)&1 == 1 {
			pix |= Black
		}
		if (invert>>i)&1 == 1 {
			pix ^= Invert | Black
		}
		// top left
		switch {
		case i < 6:
			p.Pixel[i][8] = pix
		case i < 8:
			p.Pixel[i+1][8] = pix
		case i < 9:
			p.Pixel[8][7] = pix
		default:
			p.Pixel[8][14-i] = pix
		}
		// bottom right
		switch {
		case i < 8:
			p.Pixel[8][siz-1-int(i)] = pix
		default:
			p.Pixel[siz-1-int(14-i)][8] = pix
		}
	}
}

// lplan edits a version-only Plan to add information
// about the error correction levels.
func (p *Plan) lplan() {
	v := p.Version
	l := p.Level

	nblock := vtab[v].level[l].nblock
	ne := vtab[v].level[l].check
	nde := (vtab[v].bytes - ne*nblock) / nblock
	extra := (vtab[v].bytes - ne*nblock) % nblock
	dataBits := (nde*nblock + extra) * 8
	checkBits := ne * nblock * 8

	p.DataBytes = vtab[v].bytes - ne*nblock
	p.CheckBytes = ne * nblock
	p.Blocks = nblock

	// Make data + checksum pixels.
	data := make([]Pixel, dataBits)
	for i := range data {
		data[i] = Data.Pixel() | OffsetPixel(uint(i))
	}
	check := make([]Pixel, checkBits)
	for i := range check {
		check[i] = Check.Pixel() | OffsetPixel(uint(i+dataBits))
	}

	// Split into blocks.
	dataList := make([][]Pixel, nblock)
	checkList := make([][]Pixel, nblock)
	for i := 0; i < nblock; i++ {
		// The last few blocks have an extra data byte (8 pixels).
		nd := nde
		if i >= nblock-extra {
			nd++
		}
		dataList[i], data = data[0:nd*8], data[nd*8:]
		checkList[i], check = check[0:ne*8], check[ne*8:]
	}
	if len(data) != 0 || len(check) != 0 {
		panic("qr: data/check math")
	}

	// Build up bit sequence, taking first byte of each block,
	// then second byte, and so on. Then checksums.
	bits := make([]Pixel, dataBits+checkBits)
	dst := bits
	for i := 0; i < nde+1; i++ {
		for _, b := range dataList {
			if i*8 < len(b) {
				copy(dst, b[i*8:(i+1)*8])
				dst = dst[8:]
			}
		}
	}
	for i := 0; i < ne; i++ {
		for _, b := range checkList {
			if i*8 < len(b) {
				copy(dst, b[i*8:(i+1)*8])
				dst = dst[8:]
			}
		}
	}
	if len(dst) != 0 {
		panic("qr: dst math")
	}

	// Sweep up pair of columns,
	// then down, assigning to right then left pixel.
	// Repeat.
	// See Figure 2 of http://www.pclviewer.com/rs2/qrtopology.htm
	siz := len(p.Pixel)
	rem := make([]Pixel, 7)
	for i := range rem {
		rem[i] = Extra.Pixel()
	}
	src := append(bits, rem...)
	for x := siz; x > 0; {
		for y := siz - 1; y >= 0; y-- {
			if p.Pixel[y][x-1].Role() == 0 {
				p.Pixel[y][x-1], src = src[0], src[1:]
			}
			if p.Pixel[y][x-2].Role() == 0 {
				p.Pixel[y][x-2], src = src[0], src[1:]
			}
		}
		x -= 2
		if x == 7 { // vertical timing strip
			x--
		}
		for y := 0; y < siz; y++ {
			if p.Pixel[y][x-1].Role() == 0 {
				p.Pixel[y][x-1], src = src[0], src[1:]
			}
			if p.Pixel[y][x-2].Role() == 0 {
				p.Pixel[y][x-2], src = src[0], src[1:]
			}
		}
		x -= 2
	}
}

// mplan edits a version+level-only Plan to add the mask.
func (p *Plan) mplan() {
	for y, row := range p.Pixel {
		for x, pix := range row {
			r := pix.Role()
			if (r == Data || r == Check || r == Extra) && p.Mask.Invert(y, x) {
				row[x] ^= Black | Invert
			}
		}
	}
}

// posBox draws a position (large) box at upper left x, y.
func posBox(m [][]Pixel, x, y int) {
	pos := Position.Pixel()
	// box
	for dy := 0; dy < 7; dy++ {
		for dx := 0; dx < 7; dx++ {
			p := pos
			if dx == 0 || dx == 6 || dy == 0 || dy == 6 || 2 <= dx && dx <= 4 && 2 <= dy && dy <= 4 {
				p |= Black
			}
			m[y+dy][x+dx] = p
		}
	}
	// white border
	for dy := -1; dy < 8; dy++ {
		if 0 <= y+dy && y+dy < len(m) {
			if x > 0 {
				m[y+dy][x-1] = pos
			}
			if x+7 < len(m) {
				m[y+dy][x+7] = pos
			}
		}
	}
	for dx := -1; dx < 8; dx++ {
		if 0 <= x+dx && x+dx < len(m) {
			if y > 0 {
				m[y-1][x+dx] = pos
			}
			if y+7 < len(m) {
				m[y+7][x+dx] = pos
			}
		}
	}
}

// alignBox draw an alignment (small) box at upper left x, y.
func alignBox(m [][]Pixel, x, y int) {
	// box
	align := Alignment.Pixel()
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 5; dx++ {
			p := align
			if dx == 0 || dx == 4 || dy == 0 || dy == 4 || dx == 2 && dy == 2 {
				p |= Black
			}
			m[y+dy][x+dx] = p
		}
	}
}
