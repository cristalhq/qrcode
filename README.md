# qrcode

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]
[![version-img]][version-url]

QR code for Go.

## Features

* Fast.
* Simple API.
* Dependency-free.
* Clean and tested code.
* Based on [Russ Cox qr](https://github.com/rsc/qr).

See [GUIDE.md](https://github.com/cristalhq/qrcode/blob/main/GUIDE.md) for more details.

## Install

Go version 1.17+

```
go get github.com/cristalhq/qrcode
```

## Example

```go
url := "otpauth://totp/Example:alice@bob.com?secret=JBSWY3DPEHPK3PXP&issuer=Example"

code, err := qrcode.Encode(url, qrcode.L)
checkErr(err)

f, err := os.Create("qr.jpg")
checkErr(err)
defer f.Close()

err = jpeg.Encode(f, code.Image(), nil)
checkErr(err)
```

Also see examples: [examples_test.go](https://github.com/cristalhq/qrcode/blob/main/example_test.go).

## Documentation

See [these docs][pkg-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristalhq/qrcode/workflows/build/badge.svg
[build-url]: https://github.com/cristalhq/qrcode/actions
[pkg-img]: https://pkg.go.dev/badge/cristalhq/qrcode
[pkg-url]: https://pkg.go.dev/github.com/cristalhq/qrcode
[reportcard-img]: https://goreportcard.com/badge/cristalhq/qrcode
[reportcard-url]: https://goreportcard.com/report/cristalhq/qrcode
[coverage-img]: https://codecov.io/gh/cristalhq/qrcode/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristalhq/qrcode
[version-img]: https://img.shields.io/github/v/release/cristalhq/qrcode
[version-url]: https://github.com/cristalhq/qrcode/releases
