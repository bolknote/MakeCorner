package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromFlags(t *testing.T) {
	cfg, err := Load("corner", []string{"--quality", "90", "--width", "320", "--radius", "12", "--background", "#112233", "--mask", "*", "--out-dir", "res", "--save-exif", "--recursive", "--keep-name"})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Quality != 90 || cfg.Width != 320 || cfg.Radius != 12 {
		t.Fatalf("unexpected numeric config: %+v", cfg)
	}
	if cfg.Background != [3]uint8{0x11, 0x22, 0x33} {
		t.Fatalf("unexpected background: %#v", cfg.Background)
	}
	if cfg.Mask != "*" {
		t.Fatalf("unexpected translated mask: %q", cfg.Mask)
	}
	if cfg.OutDir != "res" || !cfg.SaveExif || !cfg.Recursive || !cfg.KeepName {
		t.Fatalf("unexpected bool/string fields: %+v", cfg)
	}
}

func TestLoadFromIniFallbackAndValidation(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	ini := "[options]\nquality=77\nbackground=#abcdef\nrecursive=1\n"
	if err := os.WriteFile("makecorner.ini", []byte(ini), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(filepath.Join(tmp, "corner"), nil)
	if err != nil {
		t.Fatalf("Load from ini error: %v", err)
	}
	if cfg.Quality != 77 || cfg.Background != [3]uint8{0xab, 0xcd, 0xef} || !cfg.Recursive {
		t.Fatalf("unexpected ini config: %+v", cfg)
	}

	_, err = Load("corner", []string{"--background", "oops"})
	if err == nil || !strings.Contains(err.Error(), "invalid background color") {
		t.Fatalf("expected invalid background error, got: %v", err)
	}
}

func TestMooASCIIContainsCow(t *testing.T) {
	moo := MooASCII()
	if !strings.Contains(moo, "(oo)") {
		t.Fatalf("moo ascii changed unexpectedly")
	}
}

func TestLoadTranslatesLegacyMask(t *testing.T) {
	cfg, err := Load("corner", []string{"--mask", "*.{j,J}{p,P}{g,G}"})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Mask != "*.[jJ][pP][gG]" {
		t.Fatalf("unexpected translated mask: %q", cfg.Mask)
	}
}
