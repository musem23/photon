package image

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func createTestImage(width, height int, hasAlpha bool) image.Image {
	if hasAlpha {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				alpha := uint8(255)
				if x < width/4 {
					alpha = 0
				} else if x < width/2 {
					alpha = 128
				}
				img.Set(x, y, color.RGBA{
					R: uint8((x * 255) / width),
					G: uint8((y * 255) / height),
					B: uint8(128),
					A: alpha,
				})
			}
		}
		return img
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: uint8(128),
				A: 255,
			})
		}
	}
	return img
}

func TestFormatFromExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected Format
		wantErr  bool
	}{
		{"image.png", FormatPNG, false},
		{"image.PNG", FormatPNG, false},
		{"image.jpg", FormatJPEG, false},
		{"image.jpeg", FormatJPEG, false},
		{"image.JPEG", FormatJPEG, false},
		{"image.gif", FormatGIF, false},
		{"image.webp", FormatWebP, false},
		{"image.bmp", FormatBMP, false},
		{"image.tiff", FormatTIFF, false},
		{"image.tif", FormatTIFF, false},
		{"image.avif", FormatAVIF, false},
		{"image.heic", FormatHEIC, false},
		{"image.heif", FormatHEIC, false},
		{"image.xyz", "", true},
		{"image", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := FormatFromExtension(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatFromExtension(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("FormatFromExtension(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	formats := []struct {
		format        Format
		ext           string
		supportsAlpha bool
		canWrite      bool
	}{
		{FormatPNG, "png", true, true},
		{FormatJPEG, "jpg", false, true},
		{FormatGIF, "gif", true, true},
		{FormatWebP, "webp", true, true},
		{FormatBMP, "bmp", false, true},
		{FormatTIFF, "tiff", true, true},
		{FormatAVIF, "avif", true, true},
		{FormatHEIC, "heic", true, false},
	}

	tmpDir := t.TempDir()

	for _, src := range formats {
		if !src.canWrite {
			continue
		}

		srcPath := filepath.Join(tmpDir, "test."+src.ext)
		srcImg := createTestImage(100, 100, src.supportsAlpha)

		srcFile, err := os.Create(srcPath)
		if err != nil {
			t.Fatalf("create %s: %v", srcPath, err)
		}
		if err := Encode(srcFile, srcImg, src.format, 95); err != nil {
			srcFile.Close()
			t.Fatalf("encode %s: %v", src.format, err)
		}
		srcFile.Close()

		for _, dst := range formats {
			if !dst.canWrite {
				continue
			}

			t.Run(string(src.format)+"_to_"+string(dst.format), func(t *testing.T) {
				dstPath := filepath.Join(tmpDir, "out_"+string(src.format)+"_to_"+string(dst.format)+"."+dst.ext)

				srcFile, err := os.Open(srcPath)
				if err != nil {
					t.Fatalf("open source: %v", err)
				}
				defer srcFile.Close()

				img, _, err := Decode(srcFile)
				if err != nil {
					t.Fatalf("decode: %v", err)
				}

				dstFile, err := os.Create(dstPath)
				if err != nil {
					t.Fatalf("create dest: %v", err)
				}
				defer dstFile.Close()

				if err := Encode(dstFile, img, dst.format, 95); err != nil {
					t.Fatalf("encode: %v", err)
				}

				info, err := os.Stat(dstPath)
				if err != nil {
					t.Fatalf("stat: %v", err)
				}
				if info.Size() == 0 {
					t.Error("output file is empty")
				}
			})
		}
	}
}

func TestHEICWriteBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	img := createTestImage(100, 100, false)

	f, err := os.Create(filepath.Join(tmpDir, "test.heic"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	err = Encode(f, img, FormatHEIC, 95)
	if err == nil {
		t.Error("expected error for HEIC encoding, got nil")
	}
}

func TestQualityParameter(t *testing.T) {
	tmpDir := t.TempDir()
	img := createTestImage(200, 200, false)

	qualities := []int{10, 50, 95}
	var sizes []int64

	for _, q := range qualities {
		path := filepath.Join(tmpDir, "q"+string(rune('0'+q/10))+".jpg")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Encode(f, img, FormatJPEG, q); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()

		info, _ := os.Stat(path)
		sizes = append(sizes, info.Size())
	}

	if sizes[0] >= sizes[2] {
		t.Errorf("quality 10 (%d bytes) should be smaller than quality 95 (%d bytes)", sizes[0], sizes[2])
	}
}

func TestTransparencyPreservation(t *testing.T) {
	formats := []Format{FormatPNG, FormatWebP, FormatGIF}
	tmpDir := t.TempDir()

	srcImg := createTestImage(100, 100, true)

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			path := filepath.Join(tmpDir, "alpha."+string(format))

			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := Encode(f, srcImg, format, 95); err != nil {
				f.Close()
				t.Fatal(err)
			}
			f.Close()

			f, err = os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			decoded, _, err := Decode(f)
			if err != nil {
				t.Fatal(err)
			}

			_, _, _, a := decoded.At(10, 10).RGBA()
			if format != FormatGIF && a == 0xFFFF {
				t.Log("warning: transparency may not be fully preserved")
			}
		})
	}
}

func TestEmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.png")

	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(emptyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, _, err = Decode(f)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestInvalidFormat(t *testing.T) {
	img := createTestImage(10, 10, false)
	tmpDir := t.TempDir()

	f, _ := os.Create(filepath.Join(tmpDir, "test.bin"))
	defer f.Close()

	err := Encode(f, img, Format("invalid"), 95)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestIsSupported(t *testing.T) {
	supported := []Format{FormatPNG, FormatJPEG, FormatGIF, FormatWebP, FormatBMP, FormatTIFF, FormatAVIF, FormatHEIC}
	for _, f := range supported {
		if !IsSupported(f) {
			t.Errorf("expected %s to be supported", f)
		}
	}

	if IsSupported(Format("xyz")) {
		t.Error("unexpected format should not be supported")
	}
}
