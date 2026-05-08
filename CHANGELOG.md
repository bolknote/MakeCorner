# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Concurrent file processing via a `runtime.NumCPU()`-bounded worker pool
- `Processed N file(s)` summary printed to stdout after each run, with a
  failed-count suffix when any file fails
- `Stats.Failed` counter alongside `Stats.Processed`
- Multi-pattern brace expansion: `*.{jpg,png}` now produces multiple globs
  that are unioned with deduplication during file discovery
- Tests for multi-mask union/dedup, summary output, `Stats.Failed`
  accounting, and table-driven coverage for `expandMask`

### Fixed
- Brace expansion for multi-character alternatives. `*.{jpg,png}` previously
  collapsed to the character class `*.[jpgpn]` and silently matched the
  wrong files; it now expands to separate `*.jpg` / `*.png` patterns

### Changed
- `config.Config.Mask` (string) replaced by `config.Config.Masks` ([]string)
  to hold the expanded pattern set

## [0.1.0] - 2026-05-08

### Added
- Initial Go module and CLI entry point in `cmd/corner`
- Configuration parsing via flags and INI support in `internal/config`
- Image loading/saving and format helpers in `internal/imageio`
- Processing pipeline for discovery and run flow in `internal/pipeline`
- Bezier-based rounded corner renderer in `internal/render`
- Unit and golden tests for config, image IO, pipeline, and rendering
- Build/lint/analyze/vulnerability Makefile workflows and README documentation
