package qrcode_test

import (
	"image/png"
	"os"

	"github.com/cristalhq/qrcode"
)

func Example() {
	url := "otpauth://totp/Example:alice@bob.com?secret=JBSWY3DPEHPK3PXP&issuer=Example"

	code, err := qrcode.Encode(url, qrcode.H)
	checkErr(err)

	f, err := os.Create("qr.png")
	checkErr(err)
	defer f.Close()

	err = png.Encode(f, code.Image())
	checkErr(err)

	// Output:
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
