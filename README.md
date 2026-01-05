# Photon

A fast image format converter with interactive terminal UI and CLI mode.

## Features

- Interactive TUI with file browser and quality slider
- CLI mode for scripting and automation
- Supports 8 formats: PNG, JPEG, GIF, WebP, BMP, TIFF, AVIF, HEIC

## Installation

### Pre-built binaries

Download from [Releases](https://github.com/musem23/photon/releases/latest).

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/musem23/photon/releases/latest/download/photon-macos-arm64 -o photon
chmod +x photon
sudo mv photon /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/musem23/photon/releases/latest/download/photon-macos-amd64 -o photon
chmod +x photon
sudo mv photon /usr/local/bin/
```

**Linux**
```bash
curl -L https://github.com/musem23/photon/releases/latest/download/photon-linux-amd64 -o photon
chmod +x photon
sudo mv photon /usr/local/bin/
```

**Windows**

Download `photon-windows-amd64.exe` from [Releases](https://github.com/musem23/photon/releases/latest) and add to PATH.

### Build from source

Requires Go 1.24+ and libheif.

```bash
# macOS
brew install libheif

# Linux
sudo apt install libheif-dev

# Build
git clone https://github.com/musem23/photon.git
cd photon
CGO_ENABLED=1 go build -o photon ./cmd/photon
```

## Usage

### Interactive mode

```bash
photon
```

Navigate with arrow keys, adjust quality with `←/→`, press Enter to convert.

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate files |
| `←/→` | Adjust quality |
| `Enter` | Select/Convert |
| `Tab` | Show hidden files |
| `q` | Quit |

### CLI mode

```bash
# Single file
photon convert input.heic output.jpg
photon convert photo.png photo.webp -q 85

# Batch convert
photon batch ./photos --from heic --to jpg
photon batch ./images --from png --to avif -q 80
```

## Supported formats

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| PNG | Yes | Yes | Lossless with alpha |
| JPEG | Yes | Yes | Quality 1-100 |
| GIF | Yes | Yes | 256 colors |
| WebP | Yes | * | Modern format |
| BMP | Yes | Yes | Uncompressed |
| TIFF | Yes | Yes | Professional |
| AVIF | * | * | Best compression |
| HEIC | * | No | Apple format |

\* Pre-built binaries: WebP read-only, no HEIC/AVIF. Build from source with CGO for full support.

## Configuration

Preferences saved to `~/.config/photon/config.json`:

- `default_quality`: Output quality (default: 95)
- `default_format`: Preferred format (default: webp)
- `show_hidden_files`: Show dotfiles (default: false)

## License

MIT
