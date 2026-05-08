package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReturnsErrorOnEmptyArgs(t *testing.T) {
	err := run(nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
	if !strings.Contains(err.Error(), "argv[0]") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunMooPrintsASCII(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"corner", "--moo"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("run moo: %v", err)
	}
	if !strings.Contains(out.String(), "(__)") {
		t.Fatalf("expected moo ascii output, got %q", out.String())
	}
}

func TestRunReturnsConfigError(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := run([]string{"corner", "--quality", "999"}, &out, &errOut)
	if err == nil {
		t.Fatal("expected config validation error")
	}
	if !strings.Contains(err.Error(), "quality") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPrintsSummaryOnSuccess(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Write a minimal valid JPEG so the pipeline has something to process.
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := range 20 {
		for x := range 20 {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	f, err := os.Create(filepath.Join(tmp, "sample.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 80}); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	var out bytes.Buffer
	err = run([]string{"corner", "--mask", "*.jpg", "--out-dir", "out", "--radius", "0", "--width", "0"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !strings.Contains(out.String(), "Processed 1 file(s)") {
		t.Fatalf("expected summary on stdout, got %q", out.String())
	}
}

func TestRunReturnsPipelineError(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err = run([]string{"corner", "--mask", "*.jpg", "--out-dir", "out"}, &out, io.Discard)
	if err == nil {
		t.Fatal("expected pipeline error for empty input")
	}
	if !strings.Contains(err.Error(), "no files found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
