package pipeline

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corner/internal/config"
)

func makeJPEG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: uint8(50 + x), G: uint8(50 + y), B: 80, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
}

func makePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestProcessorRunKeepName(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	in := filepath.Join(tmp, "a.jpg")
	makeJPEG(t, in)

	p := NewProcessor(config.Config{
		Quality:    80,
		Width:      10,
		Radius:     2,
		Background: [3]uint8{255, 255, 255},
		Mask:       "*.jpg",
		OutDir:     "out",
		KeepName:   true,
	})

	stats, err := p.Run()
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if stats.Processed != 1 {
		t.Fatalf("expected 1 processed, got %d", stats.Processed)
	}
	if _, err := os.Stat(in); err != nil {
		t.Fatalf("input should still exist: %v", err)
	}
}

func TestProcessorRunGeneratedName(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	makeJPEG(t, filepath.Join(tmp, "a.jpg"))
	p := NewProcessor(config.Config{
		Quality:    75,
		Width:      0,
		Radius:     0,
		Background: [3]uint8{255, 255, 255},
		Mask:       "*.jpg",
		OutDir:     "out",
		KeepName:   false,
	})
	_, err := p.Run()
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	entries, err := os.ReadDir("out")
	if err != nil {
		t.Fatalf("readdir out: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one output file, got %d", len(entries))
	}
}

func TestProcessorRunPNGUsesTransparentCorners(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	makePNG(t, filepath.Join(tmp, "a.png"))
	p := NewProcessor(config.Config{
		Quality:    80,
		Width:      0,
		Radius:     4,
		Background: [3]uint8{255, 255, 255},
		Mask:       "*.png",
		OutDir:     "out",
		KeepName:   true,
	})
	if _, err := p.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	f, err := os.Open(filepath.Join("out", "a.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	out, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, cornerAlpha := out.At(0, 0).RGBA()
	if cornerAlpha == 0xffff {
		t.Fatalf("expected PNG corner to be transparent, alpha=%d", cornerAlpha)
	}
	_, _, _, centerAlpha := out.At(5, 5).RGBA()
	if centerAlpha != 0xffff {
		t.Fatalf("expected PNG center to stay opaque, alpha=%d", centerAlpha)
	}
}

func TestProcessorRunContinuesOnPerFileErrorAndAggregates(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	makeJPEG(t, filepath.Join(tmp, "ok.jpg"))
	makeJPEG(t, filepath.Join(tmp, "bad.jpg"))

	p := NewProcessor(config.Config{
		Quality:    80,
		Width:      0,
		Radius:     2,
		Background: [3]uint8{255, 255, 255},
		Mask:       "*.jpg",
		OutDir:     "out",
		KeepName:   true,
	})

	// Make one destination path invalid as a file to force processOne failure.
	if err := os.WriteFile(filepath.Join(tmp, "out"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	stats, err := p.Run()
	if err == nil {
		t.Fatal("expected aggregated error")
	}
	if stats.Processed != 0 {
		t.Fatalf("expected 0 processed with invalid out path, got %d", stats.Processed)
	}

	// Fix output path and break only one destination target.
	if err := os.Remove(filepath.Join(tmp, "out")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("out", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join("out", "bad.jpg"), 0o755); err != nil {
		t.Fatal(err)
	}

	stats, err = p.Run()
	if err == nil {
		t.Fatal("expected non-nil aggregated error")
	}
	if !strings.Contains(err.Error(), "bad.jpg") {
		t.Fatalf("expected failing file path in aggregated error, got %v", err)
	}
	if stats.Processed != 1 {
		t.Fatalf("expected one successful file despite one failure, got %d", stats.Processed)
	}
	if _, statErr := os.Stat(filepath.Join("out", "ok.jpg")); statErr != nil {
		t.Fatalf("expected output for successful file, got stat error: %v", statErr)
	}
}
