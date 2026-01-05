//go:build !cgo || noheif

package image

import (
	"fmt"
	goimage "image"
	"io"
)

func init() {
	heifSupported = false
}

func decodeHEIF(data []byte) (goimage.Image, error) {
	return nil, fmt.Errorf("HEIC/AVIF support not available (build without CGO)")
}

func encodeAVIF(w io.Writer, img goimage.Image, quality int) error {
	return fmt.Errorf("AVIF encoding not available (build without CGO)")
}
