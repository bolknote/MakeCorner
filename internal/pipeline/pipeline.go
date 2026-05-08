package pipeline

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"corner/internal/config"
	"corner/internal/imageio"
	"corner/internal/render"
)

type Processor struct {
	cfg config.Config
}

type Stats struct {
	Processed int
}

func NewProcessor(cfg config.Config) *Processor {
	return &Processor{cfg: cfg}
}

func (p *Processor) Run() (Stats, error) {
	files, err := collectFiles(p.cfg.Mask, p.cfg.Recursive, p.cfg.OutDir)
	if err != nil {
		return Stats{}, err
	}
	if len(files) == 0 {
		return Stats{}, fmt.Errorf("no files found")
	}

	if err := os.MkdirAll(p.cfg.OutDir, 0o755); err != nil {
		return Stats{}, err
	}

	var stats Stats
	var runErr error
	for i, src := range files {
		started := time.Now()
		if err := p.processOne(src, i+1, len(files)); err != nil {
			fileErr := fmt.Errorf("processing %s: %w", src, err)
			slog.Error("file processing failed",
				"file", src,
				"index", i+1,
				"total", len(files),
				"duration", time.Since(started),
				"error", err,
			)
			runErr = errors.Join(runErr, fileErr)
			continue
		}
		slog.Info("file processed",
			"file", src,
			"index", i+1,
			"total", len(files),
			"duration", time.Since(started),
		)
		stats.Processed++
	}
	if runErr != nil {
		return stats, fmt.Errorf("completed with errors: %w", runErr)
	}
	return stats, nil
}

func (p *Processor) processOne(src string, idx, total int) error {
	var exifSegment []byte
	if p.cfg.SaveExif {
		seg, err := imageio.ReadEXIFSegment(src)
		switch {
		case err == nil:
			exifSegment = seg
		case !errors.Is(err, imageio.ErrNotJPEG):
			slog.Warn("failed to read EXIF segment",
				"file", src,
				"error", err,
			)
		}
	}

	img, err := imageio.Load(src)
	if err != nil {
		return err
	}
	defer func() { _ = img.Close() }()

	if resized, err := imageio.ResizeIfNeeded(img, p.cfg.Width); err != nil {
		return err
	} else if resized != img {
		_ = img.Close()
		img = resized
	}

	dst := filepath.Join(p.cfg.OutDir, filepath.Base(src))
	if !p.cfg.KeepName {
		dst = generatedName(p.cfg.OutDir, idx, total, src)
	}

	format, err := imageio.DetectFormat(dst)
	if err != nil {
		return err
	}
	render.ApplyBezierRoundedCorners(img, p.cfg.Radius, p.cfg.Background, imageio.SupportsAlpha(format))

	if err := imageio.Save(img, dst, p.cfg.Quality); err != nil {
		return err
	}
	if len(exifSegment) > 0 && format == imageio.FormatJPEG {
		if err := imageio.WriteEXIFSegment(dst, exifSegment); err != nil {
			slog.Warn("failed to write EXIF segment",
				"file", dst,
				"error", err,
			)
		}
	}
	return nil
}

func generatedName(outDir string, idx, total int, src string) string {
	now := time.Now().Format("2006.01.02")
	ext := filepath.Ext(src)
	if ext == "" {
		ext = ".jpg"
	}
	if total <= 1 {
		return filepath.Join(outDir, now+ext)
	}
	prec := len(fmt.Sprintf("%d", total))
	return filepath.Join(outDir, fmt.Sprintf("%s.%0*d%s", now, prec, idx, ext))
}

func collectFiles(mask string, recursive bool, outDir string) ([]string, error) {
	if !recursive {
		matches, err := filepath.Glob(mask)
		if err != nil {
			return nil, err
		}
		var files []string
		for _, path := range matches {
			if processableFile(path) {
				files = append(files, path)
			}
		}
		sort.Strings(files)
		return files, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ignore := filepath.Clean(filepath.Join(wd, outDir))
	var files []string
	err = filepath.WalkDir(wd, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if filepath.Clean(path) == ignore {
				return filepath.SkipDir
			}
			return nil
		}
		if matched, _ := filepath.Match(mask, filepath.Base(path)); matched && processableFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func processableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	if _, err = imageio.DetectFormat(path); err != nil {
		return false
	}
	img, err := imageio.Load(path)
	if err != nil {
		return false
	}
	_ = img.Close()
	return true
}
