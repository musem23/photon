//go:build cgo && !nowebp

package image

import (
	goimage "image"
	"io"

	"github.com/chai2010/webp"
)

func init() {
	webpWriteSupported = true
}

func encodeWebP(w io.Writer, img goimage.Image, quality int) error {
	return webp.Encode(w, img, &webp.Options{Quality: float32(quality)})
}
