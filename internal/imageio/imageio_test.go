package imageio

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: uint8(10 + x), B: uint8(10 + y), A: 255})
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

func TestLoadResizeSave(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	out := filepath.Join(tmp, "out.png")
	writeFixture(t, in, 12, 6)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	resized, err := ResizeIfNeeded(img, 6)
	if err != nil {
		t.Fatalf("ResizeIfNeeded: %v", err)
	}
	if resized != img {
		defer func() { _ = resized.Close() }()
	}
	if resized.Width() != 6 {
		t.Fatalf("unexpected width: %d", resized.Width())
	}
	if err := Save(resized, out, 80); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing output file: %v", err)
	}
}
