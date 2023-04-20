package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestPNG(t *testing.T) {
	c, err := Encode("hello, world", L)
	if err != nil {
		t.Fatal(err)
	}

	pngdat := c.PNG()
	if true {
		os.WriteFile("x.png", pngdat, 0o666)
	}

	m, err := png.Decode(bytes.NewBuffer(pngdat))
	if err != nil {
		t.Fatal(err)
	}
	gm := m.(*image.Gray)

	scale := c.Scale
	siz := c.Size
	nbad := 0
	for y := 0; y < scale*(8+siz); y++ {
		for x := 0; x < scale*(8+siz); x++ {
			v := byte(255)
			if c.IsBlack(x/scale-4, y/scale-4) {
				v = 0
			}
			if gv := gm.At(x, y).(color.Gray).Y; gv != v {
				t.Errorf("%d,%d = %d, want %d", x, y, gv, v)
				nbad++
				if nbad >= 20 {
					t.Fatalf("too many bad pixels")
				}
			}
		}
	}
}

func BenchmarkPNG(b *testing.B) {
	c, err := Encode("0123456789012345678901234567890123456789", L)
	if err != nil {
		b.Fatal(err)
	}

	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = c.PNG()
	}
	b.SetBytes(int64(len(buf)))
}

func BenchmarkImagePNG(b *testing.B) {
	c, err := Encode("0123456789012345678901234567890123456789", L)
	if err != nil {
		b.Fatal(err)
	}

	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := png.Encode(&buf, c.Image()); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(buf.Len()))
}
