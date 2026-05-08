package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileAndReadIniOptions(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	content := "k=v\n[options]\nq=88\nmask=*.jpg\n"
	if err := os.WriteFile("corner.ini", []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	parsed := parseFile("corner.ini")
	if parsed["options"]["q"] != "88" {
		t.Fatalf("unexpected parse result: %#v", parsed)
	}

	opts := readIniOptions(filepath.Join(tmp, "corner"))
	if opts["q"] != "88" {
		t.Fatalf("unexpected ini options: %#v", opts)
	}
}

func TestParseFileWithoutTrailingNewline(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	content := "[options]\nquality=77"
	if err := os.WriteFile("corner.ini", []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	parsed := parseFile("corner.ini")
	if parsed["options"]["quality"] != "77" {
		t.Fatalf("expected last line to be parsed, got %#v", parsed)
	}
}
