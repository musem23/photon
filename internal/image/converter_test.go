package image

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createTestPNG(path string, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func TestConvert(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "input.png")
	dstPath := filepath.Join(tmpDir, "output.jpg")

	if err := createTestPNG(srcPath, 100, 100); err != nil {
		t.Fatalf("create test PNG: %v", err)
	}

	opts := DefaultOptions()
	if err := Convert(srcPath, dstPath, opts); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestConvertNonExistentInput(t *testing.T) {
	tmpDir := t.TempDir()
	err := Convert(
		filepath.Join(tmpDir, "nonexistent.png"),
		filepath.Join(tmpDir, "output.jpg"),
		DefaultOptions(),
	)
	if err == nil {
		t.Error("expected error for non-existent input")
	}
}

func TestConvertInvalidOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "input.png")

	if err := createTestPNG(srcPath, 10, 10); err != nil {
		t.Fatal(err)
	}

	err := Convert(srcPath, "/nonexistent/dir/output.jpg", DefaultOptions())
	if err == nil {
		t.Error("expected error for invalid output directory")
	}
}

func TestConvertUnsupportedOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "input.png")

	if err := createTestPNG(srcPath, 10, 10); err != nil {
		t.Fatal(err)
	}

	err := Convert(srcPath, filepath.Join(tmpDir, "output.xyz"), DefaultOptions())
	if err == nil {
		t.Error("expected error for unsupported output format")
	}
}

func TestConvertBatch(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "image"+string(rune('0'+i))+".png")
		if err := createTestPNG(path, 50, 50); err != nil {
			t.Fatal(err)
		}
	}

	opts := DefaultOptions()
	if err := ConvertBatch(tmpDir, "png", "jpg", opts); err != nil {
		t.Fatalf("ConvertBatch: %v", err)
	}

	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "image"+string(rune('0'+i))+".jpg")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected output file %s to exist", path)
		}
	}
}

func TestConvertBatchNoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	err := ConvertBatch(tmpDir, "png", "jpg", DefaultOptions())
	if err == nil {
		t.Error("expected error when no files match")
	}
}

func TestConvertBatchWithDotPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	if err := createTestPNG(filepath.Join(tmpDir, "test.png"), 10, 10); err != nil {
		t.Fatal(err)
	}

	if err := ConvertBatch(tmpDir, ".png", ".webp", DefaultOptions()); err != nil {
		t.Fatalf("ConvertBatch with dot prefix: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "test.webp")); err != nil {
		t.Error("expected output file to exist")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Quality != 95 {
		t.Errorf("expected default quality 95, got %d", opts.Quality)
	}
	if opts.Lossless {
		t.Error("expected default lossless to be false")
	}
}

func TestConvertWithCustomQuality(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "input.png")

	if err := createTestPNG(srcPath, 200, 200); err != nil {
		t.Fatal(err)
	}

	lowQPath := filepath.Join(tmpDir, "low.jpg")
	highQPath := filepath.Join(tmpDir, "high.jpg")

	lowOpts := Options{Quality: 10}
	highOpts := Options{Quality: 95}

	if err := Convert(srcPath, lowQPath, lowOpts); err != nil {
		t.Fatal(err)
	}
	if err := Convert(srcPath, highQPath, highOpts); err != nil {
		t.Fatal(err)
	}

	lowInfo, _ := os.Stat(lowQPath)
	highInfo, _ := os.Stat(highQPath)

	if lowInfo.Size() >= highInfo.Size() {
		t.Errorf("low quality (%d) should produce smaller file than high quality (%d)",
			lowInfo.Size(), highInfo.Size())
	}
}

func TestConvertAllFormats(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.png")

	if err := createTestPNG(srcPath, 100, 100); err != nil {
		t.Fatal(err)
	}

	formats := []string{"jpg", "gif", "webp", "bmp", "tiff", "avif"}

	for _, ext := range formats {
		t.Run("png_to_"+ext, func(t *testing.T) {
			dstPath := filepath.Join(tmpDir, "output."+ext)
			if err := Convert(srcPath, dstPath, DefaultOptions()); err != nil {
				t.Errorf("convert to %s: %v", ext, err)
			}
		})
	}
}
