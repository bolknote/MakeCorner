# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-08

### Added
- Initial Go module and CLI entry point in `cmd/corner`
- Configuration parsing via flags and INI support in `internal/config`
- Image loading/saving and format helpers in `internal/imageio`
- Processing pipeline for discovery and run flow in `internal/pipeline`
- Bezier-based rounded corner renderer in `internal/render`
- Unit and golden tests for config, image IO, pipeline, and rendering
- Build/lint/analyze/vulnerability Makefile workflows and README documentation
