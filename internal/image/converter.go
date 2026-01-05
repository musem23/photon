package image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Quality  int
	Lossless bool
}

func DefaultOptions() Options {
	return Options{
		Quality:  95,
		Lossless: false,
	}
}

func Convert(inputPath, outputPath string, opts Options) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer inputFile.Close()

	img, srcFormat, err := Decode(inputFile)
	if err != nil {
		return err
	}

	dstFormat, err := FormatFromExtension(outputPath)
	if err != nil {
		return err
	}

	if !IsSupported(dstFormat) {
		return fmt.Errorf("output format not supported: %s", dstFormat)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outputFile.Close()

	if err := Encode(outputFile, img, dstFormat, opts.Quality); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("encode image: %w", err)
	}

	fmt.Printf("Converted %s (%s) -> %s (%s)\n", inputPath, srcFormat, outputPath, dstFormat)
	return nil
}

func ConvertBatch(dir string, fromExt, toExt string, opts Options) error {
	fromExt = strings.TrimPrefix(strings.ToLower(fromExt), ".")
	toExt = strings.TrimPrefix(strings.ToLower(toExt), ".")

	pattern := filepath.Join(dir, "*."+fromExt)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob pattern: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .%s files found in %s", fromExt, dir)
	}

	var errors []string
	for _, inputPath := range files {
		base := strings.TrimSuffix(filepath.Base(inputPath), "."+fromExt)
		outputPath := filepath.Join(dir, base+"."+toExt)

		if err := Convert(inputPath, outputPath, opts); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", inputPath, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to convert %d files:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	fmt.Printf("Batch complete: %d files converted\n", len(files))
	return nil
}
