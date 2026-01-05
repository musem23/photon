# Photon - Image Format Converter

```
 ⚛ PHOTON - Image Format Converter
```

## Overview

Photon is a powerful image format converter with an interactive terminal UI and CLI mode for scripting.

## Features

- Interactive TUI with file browser
- Format selection with visual feedback
- Quality slider with real-time preview
- User preferences saved automatically
- CLI mode for scripting/automation
- Support for 8 image formats

## Supported Formats

| Format | Extension(s) | Read | Write | Alpha | Compression | Notes |
|--------|--------------|------|-------|-------|-------------|-------|
| PNG | `.png` | Yes | Yes | Yes | Lossless | Native Go |
| JPEG | `.jpg`, `.jpeg` | Yes | Yes | No | Lossy | Quality 1-100 |
| GIF | `.gif` | Yes | Yes | Yes | Lossless | 256 colors max |
| WebP | `.webp` | Yes | Yes | Yes | Both | Modern format |
| BMP | `.bmp` | Yes | Yes | No | None | Uncompressed |
| TIFF | `.tiff`, `.tif` | Yes | Yes | Yes | Lossless | Professional |
| AVIF | `.avif` | Yes | Yes | Yes | Lossy | Best compression |
| HEIC | `.heic`, `.heif` | Yes | No | Yes | Lossy | Apple (read only) |

## Conversion Matrix

```
Source → Target  PNG  JPEG  GIF  WebP  BMP  TIFF  AVIF  HEIC
PNG               ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
JPEG              ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
GIF               ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
WebP              ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
BMP               ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
TIFF              ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
AVIF              ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
HEIC              ✓    ✓     ✓    ✓     ✓    ✓     ✓     ✗
```

## Usage

### Interactive Mode (TUI)

```bash
# Launch interactive interface
photon
```

**TUI Features:**
- File browser with image preview icons
- Format selector with descriptions
- Quality slider (1-100%)
- Settings panel for preferences
- Recent files list

**Keyboard Shortcuts:**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `←/→` or `h/l` | Adjust values |
| `Enter` | Select/Confirm |
| `Tab` | Toggle hidden files |
| `Esc` | Back/Cancel |
| `q` | Quit |

### CLI Mode

```bash
# Single conversion
photon convert <input> <output> [-q quality]

# Batch conversion
photon batch <directory> --from <ext> --to <ext> [-q quality]
```

**Examples:**
```bash
photon convert photo.heic photo.jpg
photon convert input.png output.webp -q 85
photon batch ./photos --from heic --to jpg
photon batch ./images --from png --to avif -q 80
```

## User Preferences

Preferences are saved to `~/.config/photon/config.json`:

| Setting | Default | Description |
|---------|---------|-------------|
| `default_quality` | 95 | Default output quality |
| `default_format` | webp | Preferred output format |
| `show_hidden_files` | false | Show hidden files in browser |
| `confirm_overwrite` | true | Confirm before overwriting |
| `last_input_dir` | ~ | Remember last directory |
| `recent_files` | [] | Last 10 converted files |
| `favorite_formats` | [webp, avif, jpg, png] | Pinned formats |

## Project Structure

```
photon/
├── cmd/
│   └── photon/
│       └── main.go              # Entry point (TUI + CLI)
├── internal/
│   ├── config/
│   │   └── config.go            # User preferences
│   ├── image/
│   │   ├── converter.go         # Convert, ConvertBatch
│   │   ├── converter_test.go    # Integration tests
│   │   ├── formats.go           # Encode, Decode
│   │   ├── formats_test.go      # Unit tests
│   │   └── init.go              # Decoder registration
│   └── tui/
│       ├── model.go             # Bubble Tea model
│       └── styles.go            # Lipgloss styles
├── go.mod
├── go.sum
└── SCOPE.md
```

## Dependencies

### Go Packages
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/spf13/cobra` - CLI framework
- `github.com/chai2010/webp` - WebP codec
- `golang.org/x/image` - BMP, TIFF
- `github.com/strukturag/libheif` - AVIF/HEIC (CGO)

### System Libraries
```bash
# macOS
brew install libheif

# Linux
apt install libheif-dev
```

## Test Coverage (91 tests)

### Unit Tests
- Format extension parsing (15 cases)
- Encode/Decode round-trip (49 combinations)
- HEIC write restriction
- Quality parameter effects
- Transparency preservation
- Error handling (empty, invalid, missing)

### Integration Tests
- Single file conversion
- Batch conversion
- Error cases (missing input, invalid output, unsupported format)
- Quality parameter effects
- All format outputs

## Building

```bash
# Build binary
CGO_ENABLED=1 go build -o photon ./cmd/photon

# Run tests
CGO_ENABLED=1 go test ./...

# Install globally
CGO_ENABLED=1 go install ./cmd/photon
```

## Limitations

1. **HEIC Writing**: Not supported (Apple license)
2. **Animated GIF/WebP**: First frame only
3. **EXIF/Metadata**: Not preserved
4. **Color Profiles**: Converted to sRGB
5. **Large Files**: Loaded into memory
