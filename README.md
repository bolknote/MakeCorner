# Corner

Corner is a command-line tool for resizing images and applying rounded corners.
The current renderer uses a Bezier-based corner mask and supports configurable output quality.

## Features

- Resize images to a target width while preserving aspect ratio
- Apply rounded corners with a configurable radius and background color
- Configure processing with CLI flags or an INI file
- Process files recursively or from a glob mask (with brace expansion)
- Process files in parallel across all available CPU cores
- Save output with generated names or keep original filenames
- Optionally copy EXIF metadata for JPEG outputs (atomic write)

## Project Structure

- `cmd/corner` - CLI entry point and easter egg
- `internal/config` - flag and INI configuration parsing
- `internal/pipeline` - file discovery and processing workflow
- `internal/imageio` - thin `go-gd` v2 image loading, atomic saving, format helpers, EXIF copy
- `internal/render` - Bezier-based rounded corner rendering over `gd.Image`
- `testdata` - golden fixtures for rendering tests

## Requirements

- Go 1.26+
- `libgd` (with the codecs you plan to use) and `pkg-config`
  - Debian/Ubuntu: `sudo apt-get install -y libgd-dev pkg-config`
  - macOS (Homebrew): `brew install gd pkg-config`

## Build

```bash
go build -o corner ./cmd/corner
```

## Run

```bash
./corner -m "*" -o out -r 24 -q 85 -b "#ffffff"
```

## Flags

All flags accept both a long form and a short alias.

| Long          | Short | Default     | Description                                       |
| ------------- | ----- | ----------- | ------------------------------------------------- |
| `--quality`   | `-q`  | `85`        | JPEG/WebP/HEIF/AVIF quality (1..100)              |
| `--width`     | `-w`  | `660`       | Output width; `0` keeps the source width          |
| `--radius`    | `-r`  | `10`        | Corner radius in pixels                           |
| `--background`| `-b`  | `#ffffff`   | Corner blend color for non-alpha outputs (`#RRGGBB`) |
| `--mask`      | `-m`  | `*`         | Input file mask (glob)                            |
| `--out-dir`   | `-o`  | `out`       | Output directory                                  |
| `--save-exif` | `-e`  | `false`     | Copy EXIF metadata for JPEG outputs (best-effort) |
| `--recursive` | `-R`  | `false`     | Recursive file discovery                          |
| `--keep-name` | `-k`  | `false`     | Keep source file names                            |

In recursive mode the mask is matched against each file's *base name*
(e.g. `*.jpg`), not the full relative path. The configured output directory is
skipped automatically while walking, so generated outputs do not feed back
into the next run.

Brace expansion is supported in masks:

- `*.{jpg,png}` matches files of either extension (results are unioned and
  deduplicated).
- `*.{j,J}{p,P}{g,G}` collapses to the character class `*.[jJ][pP][gG]`,
  giving you case-insensitive single-character alternatives.

### INI file

If `corner.ini` (or the legacy `makecorner.ini`) is present in the current
directory, its `[options]` section is parsed before flags. Both long and
short flag names are recognized as keys; CLI flags always win over INI.

```ini
[options]
quality=90
background=#202024
recursive=1
```

## Supported Formats

Corner uses `go-gd` v2 directly. Available codecs depend on the linked `libgd` build.

- Input/output: `.jpg`, `.jpeg`, `.png`, `.gif`, `.wbmp`, `.webp`, `.bmp`, `.tga`, `.tif`, `.tiff`, `.gd`, `.gd2`, `.heif`, `.heic`, `.avif`, `.xbm`
- Input only: `.xpm`

Per-pixel alpha is only used for formats that genuinely support it
(PNG, WebP, HEIF, AVIF, TIFF). For JPEG, GIF, BMP, WBMP, TGA the corner is
blended against `--background` instead.

## Atomic writes

Both image saves and EXIF copy go via a sibling temp file followed by
`rename(2)`, so partially written outputs are never visible to other tools or
to a second run of Corner.

## Concurrency

Files are processed in parallel by a worker pool sized to `runtime.NumCPU()`.
Per-file logs may interleave; the final aggregated error (if any) lists every
failed file. The output directory is created once before any worker starts, and
all writes go through the atomic temp+rename path described above, so
concurrent runs do not race on the destination.

On completion the binary prints a one-line summary to stdout:

```
Processed 12 file(s)
```

When some files fail, the count of failures is appended:

```
Processed 10 file(s), 2 failed
```

The process still exits non-zero in that case and the per-file errors are
emitted via `slog`.

## Testing

```bash
go test ./...
```

## Static Analysis

`golangci-lint` already includes `staticcheck`, `errcheck`, `ineffassign`, and other standard Go linters.
The only external analyzer used separately is `govulncheck`.

Install tools (pinned versions, see `Makefile`):

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

Run full local verification (build, vet, lint, tests, vulncheck):

```bash
make check
```
