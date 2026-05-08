package imageio

import (
	"testing"

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
)

func TestDetectFormat(t *testing.T) {
	cases := map[string]gd.Format{
		"a.jpg":  gd.FormatJPEG,
		"a.jpeg": gd.FormatJPEG,
		"a.png":  gd.FormatPNG,
		"a.gif":  gd.FormatGIF,
		"a.wbmp": gd.FormatWBMP,
		"a.webp": gd.FormatWebP,
		"a.bmp":  gd.FormatBMP,
		"a.tga":  gd.FormatTGA,
		"a.tiff": gd.FormatTIFF,
		"a.gd":   gd.FormatGD,
		"a.gd2":  gd.FormatGD2,
		"a.heic": gd.FormatHEIF,
		"a.avif": gd.FormatAVIF,
		"a.xpm":  gd.FormatXPM,
		"a.xbm":  gd.FormatXBM,
	}
	for in, want := range cases {
		got, err := DetectFormat(in)
		if err != nil {
			t.Fatalf("DetectFormat(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("DetectFormat(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := DetectFormat("a.txt"); err == nil {
		t.Fatal("expected unsupported extension error")
	}
}

func TestSupportsAlpha(t *testing.T) {
	if !SupportsAlpha(gd.FormatPNG) || !SupportsAlpha(gd.FormatWebP) || !SupportsAlpha(gd.FormatAVIF) {
		t.Fatal("expected alpha-capable formats to report true")
	}
	if SupportsAlpha(gd.FormatJPEG) || SupportsAlpha(gd.FormatBMP) {
		t.Fatal("expected non-alpha formats to report false")
	}
}
