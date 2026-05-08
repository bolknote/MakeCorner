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

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
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

func TestResizeIfNeededKeepsMinimumHeightOne(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "thin.jpg")
	writeFixture(t, in, 3, 1)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	resized, err := ResizeIfNeeded(img, 1)
	if err != nil {
		t.Fatalf("ResizeIfNeeded: %v", err)
	}
	if resized != img {
		defer func() { _ = resized.Close() }()
	}
	if resized.Height() != 1 {
		t.Fatalf("expected min height=1 after resize, got %d", resized.Height())
	}
}

func TestSaveFailsWhenDestinationDirectoryMissing(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	writeFixture(t, in, 12, 6)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	out := filepath.Join(tmp, "missing", "out.png")
	if err := Save(img, out, 80); err == nil {
		t.Fatal("expected save error when output directory does not exist")
	}
}

func TestSaveWithDotfileNameUsesFallbackTempPrefix(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jpg")
	writeFixture(t, in, 12, 6)

	img, err := Load(in)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer func() { _ = img.Close() }()

	out := filepath.Join(tmp, ".png")
	if err := Save(img, out, 80); err != nil {
		t.Fatalf("Save dotfile path: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing output file: %v", err)
	}
}

func TestSyncDirReturnsErrorForMissingDir(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "nope")
	if err := syncDir(missing); err == nil {
		t.Fatal("expected syncDir error for missing directory")
	}
}

func TestEncodeFileReturnsDetectFormatError(t *testing.T) {
	prevDetect := detectFormatFn
	t.Cleanup(func() { detectFormatFn = prevDetect })
	detectFormatFn = func(string) (gd.Format, error) {
		return "", errors.New("boom")
	}
	if err := encodeFile(nil, "ignored", 80); err == nil {
		t.Fatal("expected detect format error")
	}
}

func TestEncodeFileReturnsUnsupportedFormatError(t *testing.T) {
	prevDetect := detectFormatFn
	prevSupports := supportsFormatFn
	t.Cleanup(func() {
		detectFormatFn = prevDetect
		supportsFormatFn = prevSupports
	})
	detectFormatFn = func(string) (gd.Format, error) { return gd.FormatPNG, nil }
	supportsFormatFn = func(gd.Format) bool { return false }
	if err := encodeFile(nil, "ignored", 80); err == nil {
		t.Fatal("expected unsupported-format error")
	}
}

func TestEncodeFileDispatchesByFormat(t *testing.T) {
	prevDetect := detectFormatFn
	prevSupports := supportsFormatFn
	prevJPEG := encodeJPEGFn
	prevPNG := encodePNGFn
	prevGIF := encodeGIFFn
	prevWebP := encodeWebPFn
	prevWBMP := encodeWBMPFn
	prevBMP := encodeBMPFn
	prevTIFF := encodeTIFFFn
	prevGD := encodeGDFn
	prevGD2 := encodeGD2Fn
	prevHEIF := encodeHEIFFn
	prevAVIF := encodeAVIFFn
	prevXBM := encodeXBMFn
	prevFallback := encodeFallbackFn
	t.Cleanup(func() {
		detectFormatFn = prevDetect
		supportsFormatFn = prevSupports
		encodeJPEGFn = prevJPEG
		encodePNGFn = prevPNG
		encodeGIFFn = prevGIF
		encodeWebPFn = prevWebP
		encodeWBMPFn = prevWBMP
		encodeBMPFn = prevBMP
		encodeTIFFFn = prevTIFF
		encodeGDFn = prevGD
		encodeGD2Fn = prevGD2
		encodeHEIFFn = prevHEIF
		encodeAVIFFn = prevAVIF
		encodeXBMFn = prevXBM
		encodeFallbackFn = prevFallback
	})

	called := ""
	gotQ := -1
	supportsFormatFn = func(gd.Format) bool { return true }
	encodeJPEGFn = func(_ *gd.Image, _ string, q int) error { called, gotQ = "jpeg", q; return nil }
	encodePNGFn = func(_ *gd.Image, _ string) error { called = "png"; return nil }
	encodeGIFFn = func(_ *gd.Image, _ string) error { called = "gif"; return nil }
	encodeWebPFn = func(_ *gd.Image, _ string, q int) error { called, gotQ = "webp", q; return nil }
	encodeWBMPFn = func(_ *gd.Image, _ string) error { called = "wbmp"; return nil }
	encodeBMPFn = func(_ *gd.Image, _ string) error { called = "bmp"; return nil }
	encodeTIFFFn = func(_ *gd.Image, _ string) error { called = "tiff"; return nil }
	encodeGDFn = func(_ *gd.Image, _ string) error { called = "gd"; return nil }
	encodeGD2Fn = func(_ *gd.Image, _ string) error { called = "gd2"; return nil }
	encodeHEIFFn = func(_ *gd.Image, _ string, q int) error { called, gotQ = "heif", q; return nil }
	encodeAVIFFn = func(_ *gd.Image, _ string, q int) error { called, gotQ = "avif", q; return nil }
	encodeXBMFn = func(_ *gd.Image, _ string) error { called = "xbm"; return nil }
	encodeFallbackFn = func(_ *gd.Image, _ string) error { called = "fallback"; return nil }

	cases := []struct {
		name      string
		format    gd.Format
		wantCall  string
		qualityIn int
		wantQ     int
	}{
		{name: "jpeg", format: gd.FormatJPEG, wantCall: "jpeg", qualityIn: 0, wantQ: 1},
		{name: "png", format: gd.FormatPNG, wantCall: "png"},
		{name: "gif", format: gd.FormatGIF, wantCall: "gif"},
		{name: "webp", format: gd.FormatWebP, wantCall: "webp", qualityIn: 101, wantQ: 100},
		{name: "wbmp", format: gd.FormatWBMP, wantCall: "wbmp"},
		{name: "bmp", format: gd.FormatBMP, wantCall: "bmp"},
		{name: "tiff", format: gd.FormatTIFF, wantCall: "tiff"},
		{name: "gd", format: gd.FormatGD, wantCall: "gd"},
		{name: "gd2", format: gd.FormatGD2, wantCall: "gd2"},
		{name: "heif", format: gd.FormatHEIF, wantCall: "heif", qualityIn: -1, wantQ: 1},
		{name: "avif", format: gd.FormatAVIF, wantCall: "avif", qualityIn: 80, wantQ: 80},
		{name: "xbm", format: gd.FormatXBM, wantCall: "xbm"},
		{name: "fallback", format: gd.FormatTGA, wantCall: "fallback"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			detectFormatFn = func(string) (gd.Format, error) { return tc.format, nil }
			called = ""
			gotQ = -1
			if err := encodeFile(nil, "ignored", tc.qualityIn); err != nil {
				t.Fatalf("encodeFile: %v", err)
			}
			if called != tc.wantCall {
				t.Fatalf("called %q, want %q", called, tc.wantCall)
			}
			if tc.wantQ != 0 && gotQ != tc.wantQ {
				t.Fatalf("quality %d, want %d", gotQ, tc.wantQ)
			}
		})
	}
}
