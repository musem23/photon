package main

import (
	"fmt"
	"os"

	"github.com/mahamedmuse/photon/internal/image"
	"github.com/mahamedmuse/photon/internal/tui"
	"github.com/spf13/cobra"
)

var (
	quality int
	fromExt string
	toExt   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "photon",
		Short: "Photon - Image format converter",
		Long: `
  ⚛ PHOTON - Image Format Converter

  A powerful image converter with support for PNG, JPEG, GIF,
  WebP, BMP, TIFF, AVIF, and HEIC formats.

  Run without arguments to launch the interactive TUI.
  Use subcommands for CLI/scripting mode.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	convertCmd := &cobra.Command{
		Use:     "convert <input> <output>",
		Aliases: []string{"c"},
		Short:   "Convert a single image (CLI mode)",
		Example: "  photon convert photo.heic photo.jpg\n  photon convert input.png output.webp -q 85",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := image.DefaultOptions()
			opts.Quality = quality
			return image.Convert(args[0], args[1], opts)
		},
	}
	convertCmd.Flags().IntVarP(&quality, "quality", "q", 95, "Output quality (1-100)")

	batchCmd := &cobra.Command{
		Use:     "batch <directory>",
		Aliases: []string{"b"},
		Short:   "Convert all images in a directory (CLI mode)",
		Example: "  photon batch ./photos --from heic --to jpg\n  photon batch ./images --from png --to webp -q 80",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := image.DefaultOptions()
			opts.Quality = quality
			return image.ConvertBatch(args[0], fromExt, toExt, opts)
		},
	}
	batchCmd.Flags().IntVarP(&quality, "quality", "q", 95, "Output quality (1-100)")
	batchCmd.Flags().StringVar(&fromExt, "from", "", "Source format (required)")
	batchCmd.Flags().StringVar(&toExt, "to", "", "Target format (required)")
	batchCmd.MarkFlagRequired("from")
	batchCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(convertCmd, batchCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
