//go:build cgo && !noheif

package image

import (
	"fmt"
	goimage "image"
	"io"
	"os"

	"github.com/strukturag/libheif/go/heif"
)

func init() {
	heifSupported = true
}

func decodeHEIF(data []byte) (goimage.Image, error) {
	ctx, err := heif.NewContext()
	if err != nil {
		return nil, fmt.Errorf("create heif context: %w", err)
	}

	if err := ctx.ReadFromMemory(data); err != nil {
		return nil, fmt.Errorf("read heif data: %w", err)
	}

	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		return nil, fmt.Errorf("get primary image: %w", err)
	}

	img, err := handle.DecodeImage(heif.ColorspaceUndefined, heif.ChromaUndefined, nil)
	if err != nil {
		return nil, fmt.Errorf("decode heif image: %w", err)
	}

	return img.GetImage()
}

func encodeAVIF(w io.Writer, img goimage.Image, quality int) error {
	rgba := toRGBA(img)
	ctx, err := heif.EncodeFromImage(rgba, heif.CompressionAV1, quality, heif.LosslessModeDisabled, heif.LoggingLevelNone)
	if err != nil {
		return fmt.Errorf("encode avif: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "avif-*.avif")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := ctx.WriteToFile(tmpPath); err != nil {
		return fmt.Errorf("write avif: %w", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read avif: %w", err)
	}

	_, err = w.Write(data)
	return err
}
