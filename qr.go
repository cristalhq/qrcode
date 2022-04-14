package qrcode

import (
	"image"
	"image/color"

	"github.com/cristalhq/qrcode/internal/coding"
)

// A Level denotes a QR error correction level.
// From least to most tolerant of errors, they are L, M, Q, H.
type Level int

const (
	L Level = 0 // 20% redundant
	M Level = 1 // 38% redundant
	Q Level = 2 // 55% redundant
	H Level = 3 // 65% redundant
)

// Encode returns an encoding of text at the given error correction level.
func Encode(text string, level Level) (*Code, error) {
	return EncodeInto(nil, text, level)
}

func EncodeInto(bitmap []byte, text string, level Level) (*Code, error) {
	cc, err := coding.Encode(bitmap, text, coding.Level(level))
	if err != nil {
		return nil, err
	}

	code := &Code{
		Bitmap: cc.Bitmap,
		Size:   cc.Size,
		Stride: cc.Stride,
		Scale:  8,
	}
	return code, nil
}

// A Code is a square pixel grid.
// It implements image.Image and direct PNG encoding.
type Code struct {
	Bitmap []byte // 1 is black, 0 is white
	Size   int    // number of pixels on a side
	Stride int    // number of bytes per row
	Scale  int    // number of image pixels per QR pixel
}

// IsBlack returns true if the pixel at (x,y) is black.
func (c *Code) IsBlack(x, y int) bool {
	return 0 <= x && x < c.Size &&
		0 <= y && y < c.Size &&
		c.Bitmap[y*c.Stride+x/8]&(1<<uint(7-x&7)) != 0
}

// Image returns an Image displaying the code.
func (c *Code) Image() image.Image {
	return &codeImage{c}
}

// codeImage implements image.Image
type codeImage struct{ *Code }

func (c *codeImage) Bounds() image.Rectangle {
	d := (c.Size + 8) * c.Scale
	return image.Rect(0, 0, d, d)
}

func (c *codeImage) At(x, y int) color.Color {
	if c.IsBlack(x/c.Scale-4, y/c.Scale-4) {
		return blackColor
	}
	return whiteColor
}

func (c *codeImage) ColorModel() color.Model {
	return color.GrayModel
}

var whiteColor, blackColor color.Color = color.Gray{0xFF}, color.Gray{0x00}
