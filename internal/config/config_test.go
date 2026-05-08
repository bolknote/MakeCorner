package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromFlags(t *testing.T) {
	cfg, err := Load("corner", []string{
		"--quality", "90", "--width", "320", "--radius", "12",
		"--background", "#112233", "--mask", "*", "--out-dir", "res",
		"--save-exif", "--recursive", "--keep-name",
	}, io.Discard)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Quality != 90 || cfg.Width != 320 || cfg.Radius != 12 {
		t.Fatalf("unexpected numeric config: %+v", cfg)
	}
	if cfg.Background != [3]uint8{0x11, 0x22, 0x33} {
		t.Fatalf("unexpected background: %#v", cfg.Background)
	}
	if len(cfg.Masks) != 1 || cfg.Masks[0] != "*" {
		t.Fatalf("unexpected translated masks: %v", cfg.Masks)
	}
	if cfg.OutDir != "res" || !cfg.SaveExif || !cfg.Recursive || !cfg.KeepName {
		t.Fatalf("unexpected bool/string fields: %+v", cfg)
	}
}

func TestLoadFromIniFallbackAndValidation(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	ini := "[options]\nquality=77\nbackground=#abcdef\nrecursive=1\n"
	if err := os.WriteFile("makecorner.ini", []byte(ini), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(filepath.Join(tmp, "corner"), nil, io.Discard)
	if err != nil {
		t.Fatalf("Load from ini error: %v", err)
	}
	if cfg.Quality != 77 || cfg.Background != [3]uint8{0xab, 0xcd, 0xef} || !cfg.Recursive {
		t.Fatalf("unexpected ini config: %+v", cfg)
	}

	_, err = Load("corner", []string{"--background", "oops"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "invalid background color") {
		t.Fatalf("expected invalid background error, got: %v", err)
	}
}

func TestLoadTranslatesLegacyMask(t *testing.T) {
	cfg, err := Load("corner", []string{"--mask", "*.{j,J}{p,P}{g,G}"}, io.Discard)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(cfg.Masks) != 1 || cfg.Masks[0] != "*.[jJ][pP][gG]" {
		t.Fatalf("unexpected translated masks: %v", cfg.Masks)
	}
}

func TestExpandMaskMultiCharBraces(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"*.jpg", []string{"*.jpg"}},
		{"*.{jpg,png}", []string{"*.jpg", "*.png"}},
		{"*.{jpg,png,webp}", []string{"*.jpg", "*.png", "*.webp"}},
		{"*.{j,J}{p,P}{g,G}", []string{"*.[jJ][pP][gG]"}},
		{"img.{jpg,png}.bak", []string{"img.jpg.bak", "img.png.bak"}},
		{"*.{,jpg}", []string{"*.", "*.jpg"}},
		{"*.{a,-}", []string{"*.a", "*.-"}},
		{"*.{a,]}", []string{"*.a", "*.]"}},
		{"*.{jpg,png", []string{"*.{jpg,png"}},
		{"*", []string{"*"}},
	}
	for _, tc := range cases {
		got := expandMask(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("expandMask(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("expandMask(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestLoadReturnsSentinelOnFlagParseError(t *testing.T) {
	var diag bytes.Buffer
	_, err := Load("corner", []string{"--no-such-flag"}, &diag)
	if !errors.Is(err, ErrFlagParse) {
		t.Fatalf("expected ErrFlagParse, got %v", err)
	}
	if diag.Len() == 0 {
		t.Fatal("flag package should have already written a diagnostic to stderr")
	}
}

func TestLoadReturnsSentinelOnHelp(t *testing.T) {
	var diag bytes.Buffer
	_, err := Load("corner", []string{"--help"}, &diag)
	if !errors.Is(err, ErrUsageRequested) {
		t.Fatalf("expected ErrUsageRequested, got %v", err)
	}
	if !strings.Contains(diag.String(), "Flags:") {
		t.Fatalf("usage should be written to provided writer, got %q", diag.String())
	}
}

func TestLoadValidationErrorIsNotASentinel(t *testing.T) {
	_, err := Load("corner", []string{"--quality", "999"}, io.Discard)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errors.Is(err, ErrFlagParse) || errors.Is(err, ErrUsageRequested) {
		t.Fatalf("validation errors must not surface as flag-parse sentinels: %v", err)
	}
}
