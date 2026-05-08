package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectFilesNoRecursive(t *testing.T) {
	files, err := collectFiles("*.missing", false, "out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %d", len(files))
	}
}

func TestGeneratedNameKeepsSourceExtension(t *testing.T) {
	cases := map[string]string{
		"photo.jpg":  ".jpg",
		"photo.png":  ".png",
		"photo.webp": ".webp",
		"photo.avif": ".avif",
	}
	for src, wantExt := range cases {
		got := generatedName("out", 1, 1, src)
		if filepath.Ext(got) != wantExt {
			t.Fatalf("generatedName(%q) ext = %q, want %q", src, filepath.Ext(got), wantExt)
		}
	}
}

func TestCollectFilesSkipsDirectoriesAndUnsupportedFiles(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(".git", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("README.md", []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeJPEG(t, filepath.Join(tmp, "image.jpg"))

	files, err := collectFiles("*", false, "out")
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "image.jpg" {
		t.Fatalf("expected only supported regular image paths, got %#v", files)
	}
}

func TestCollectFilesAcceptsBrokenImageForLaterReport(t *testing.T) {
	// Discovery is deliberately extension-only: files with a known image
	// extension reach Run, which reports per-file decode failures via
	// errors.Join instead of silently dropping them at discovery time.
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("broken.jpg", []byte("not decoded here"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectFiles("*", false, "out")
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "broken.jpg" {
		t.Fatalf("expected broken image to be discovered, got %#v", files)
	}
}

func TestCollectFilesRecursiveSkipsOutDir(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join("src", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join("out", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	makeJPEG(t, filepath.Join(tmp, "src", "nested", "in.jpg"))
	makeJPEG(t, filepath.Join(tmp, "out", "nested", "out.jpg"))

	files, err := collectFiles("*.jpg", true, "out")
	if err != nil {
		t.Fatalf("collectFiles recursive: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one recursive source image outside out dir, got %#v", files)
	}
	if filepath.Base(files[0]) != "in.jpg" {
		t.Fatalf("unexpected recursive match: %#v", files)
	}
}
