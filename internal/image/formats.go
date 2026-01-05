package image

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/strukturag/libheif/go/heif"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

type Format string

const (
	FormatPNG  Format = "png"
	FormatJPEG Format = "jpeg"
	FormatGIF  Format = "gif"
	FormatWebP Format = "webp"
	FormatBMP  Format = "bmp"
	FormatTIFF Format = "tiff"
	FormatAVIF Format = "avif"
	FormatHEIC Format = "heic"
)

var supportedFormats = map[Format]bool{
	FormatPNG:  true,
	FormatJPEG: true,
	FormatGIF:  true,
	FormatWebP: true,
	FormatBMP:  true,
	FormatTIFF: true,
	FormatAVIF: true,
	FormatHEIC: true,
}

var writeOnlyFormats = map[Format]bool{
	FormatHEIC: true,
}

func FormatFromExtension(path string) (Format, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	switch ext {
	case "png":
		return FormatPNG, nil
	case "jpg", "jpeg":
		return FormatJPEG, nil
	case "gif":
		return FormatGIF, nil
	case "webp":
		return FormatWebP, nil
	case "bmp":
		return FormatBMP, nil
	case "tiff", "tif":
		return FormatTIFF, nil
	case "avif":
		return FormatAVIF, nil
	case "heic", "heif":
		return FormatHEIC, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", ext)
	}
}

func Decode(r io.Reader) (image.Image, Format, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("read image data: %w", err)
	}

	if isHEIF(data) {
		img, err := decodeHEIF(data)
		if err != nil {
			return nil, "", err
		}
		format := FormatHEIC
		if isAVIF(data) {
			format = FormatAVIF
		}
		return img, format, nil
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image: %w", err)
	}
	return img, Format(format), nil
}

func isHEIF(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	// Check for ftyp box
	if string(data[4:8]) != "ftyp" {
		return false
	}
	brand := string(data[8:12])
	return brand == "heic" || brand == "heix" || brand == "hevc" ||
		brand == "avif" || brand == "avis" || brand == "mif1"
}

func isAVIF(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if string(data[4:8]) != "ftyp" {
		return false
	}
	brand := string(data[8:12])
	return brand == "avif" || brand == "avis"
}

func decodeHEIF(data []byte) (image.Image, error) {
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

func Encode(w io.Writer, img image.Image, format Format, quality int) error {
	switch format {
	case FormatPNG:
		return png.Encode(w, img)
	case FormatJPEG:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case FormatGIF:
		return gif.Encode(w, img, nil)
	case FormatWebP:
		return webp.Encode(w, img, &webp.Options{Quality: float32(quality)})
	case FormatBMP:
		return bmp.Encode(w, img)
	case FormatTIFF:
		return tiff.Encode(w, img, nil)
	case FormatAVIF:
		return encodeAVIF(w, img, quality)
	case FormatHEIC:
		return fmt.Errorf("HEIC encoding not supported (Apple license restriction)")
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

func encodeAVIF(w io.Writer, img image.Image, quality int) error {
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

func IsSupported(format Format) bool {
	return supportedFormats[format]
}

func CanWrite(format Format) bool {
	return supportedFormats[format] && !writeOnlyFormats[format]
}
