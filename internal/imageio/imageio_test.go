package imageio

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io/fs"
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

func TestResizeIfNeededNoopForZeroOrSameWidth(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	writeFixture(t, in, 12, 6)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	got, err := ResizeIfNeeded(img, 0)
	if err != nil {
		t.Fatalf("ResizeIfNeeded zero width: %v", err)
	}
	if got != img {
		t.Fatal("expected same image pointer for width 0")
	}

	got, err = ResizeIfNeeded(img, img.Width())
	if err != nil {
		t.Fatalf("ResizeIfNeeded same width: %v", err)
	}
	if got != img {
		t.Fatal("expected same image pointer for same width")
	}
}

func TestSaveUnsupportedExtensionReturnsError(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	out := filepath.Join(tmp, "out.unknown")
	writeFixture(t, in, 12, 6)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	if err := Save(img, out, 80); err == nil {
		t.Fatal("expected save error for unsupported extension")
	}
}

func TestLoadRejectsUnsupportedExtension(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "in.txt")
	if err := os.WriteFile(p, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected extension error")
	}
}

func TestLoadReturnsDecodeErrorForBrokenJPEG(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "in.jpg")
	if err := os.WriteFile(p, []byte("not a jpeg payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected decode error for broken jpeg")
	}
}

func TestSaveReturnsFsErrExistWhenDestinationIsDirectory(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	outDir := filepath.Join(tmp, "out.png")
	writeFixture(t, in, 12, 6)
	if err := os.Mkdir(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	err = Save(img, outDir, 80)
	if err == nil {
		t.Fatal("expected save error")
	}
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("expected wrapped fs.ErrExist, got %v", err)
	}
}

func TestClampQualityBounds(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{in: -5, want: 1},
		{in: 0, want: 1},
		{in: 1, want: 1},
		{in: 50, want: 50},
		{in: 100, want: 100},
		{in: 101, want: 100},
	}
	for _, tc := range cases {
		if got := clampQuality(tc.in); got != tc.want {
			t.Fatalf("clampQuality(%d)=%d, want %d", tc.in, got, tc.want)
		}
	}
}
