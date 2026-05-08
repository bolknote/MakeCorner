# Corner

Corner is a command-line tool for resizing images and applying rounded corners.
The current renderer uses a Bezier-based corner mask and supports configurable output quality.

## Features

- Resize images to a target width while preserving aspect ratio
- Apply rounded corners with a configurable radius
- Configure processing with CLI flags or INI file
- Process files recursively or from a glob mask
- Save output with generated names or keep original filenames

## Project Structure

- `cmd/corner` - CLI entry point
- `internal/config` - flag and INI configuration parsing
- `internal/pipeline` - file discovery and processing workflow
- `internal/imageio` - thin `go-gd` v2 image loading, resizing, and saving helpers
- `internal/render` - Bezier-based rounded corner rendering over `gd.Image`
- `testdata` - golden fixtures for rendering tests

## Requirements

- Go 1.26+

## Build

```bash
go build -o corner ./cmd/corner
```

## Run

```bash
./corner -m "*" -o out -r 24 -q 85
```

## Common Flags

- `-m, --mask` input file mask (default: `*`)
- `-o, --out-dir` output directory
- `-w, --width` output width (0 keeps original width)
- `-r, --radius` corner radius
- `-q, --quality` lossy quality (used by JPEG/WebP backends)
- `-e, --save-exif` copy EXIF metadata for JPEG outputs (best-effort)

- `-R, --recursive` recursive scan
- `-k, --keep-name` keep original filenames

## Supported Formats

Corner uses `go-gd` v2 directly. Available codecs depend on the linked `libgd` build.

- Input/output: `.jpg`, `.jpeg`, `.png`, `.gif`, `.wbmp`, `.webp`, `.bmp`, `.tga`, `.tif`, `.tiff`, `.gd`, `.gd2`, `.heif`, `.heic`, `.avif`, `.xbm`
- Input only: `.xpm`

## Testing

```bash
go test ./...
```

## Static Analysis

`golangci-lint` already includes `staticcheck`, `errcheck`, `ineffassign`, and other standard Go linters.
Only external analyzer used separately is `govulncheck`.

Install tools:

```bash
make tools
```

Run lint checks:

```bash
make lint
```

Run vulnerability scanning:

```bash
make vulncheck
```

Run static checks (same flow style as `go-gd`):

```bash
make analyze
```

Run full local verification:

```bash
make check
```
