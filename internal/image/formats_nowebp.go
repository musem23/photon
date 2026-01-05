//go:build !cgo || nowebp

package image

import (
	"fmt"
	goimage "image"
	"io"
)

func init() {
	webpWriteSupported = false
}

func encodeWebP(w io.Writer, img goimage.Image, quality int) error {
	return fmt.Errorf("WebP encoding not available (build without CGO)")
}
