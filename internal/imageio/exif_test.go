package imageio

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func makeJPEGWithEXIF() []byte {
	// SOI + APP1(Exif) + SOS + payload + EOI
	return []byte{
		0xFF, 0xD8,
		0xFF, 0xE1, 0x00, 0x0A, 'E', 'x', 'i', 'f', 0x00, 0x00, 0x11, 0x22,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x02, 0x03, 0x00, 0x3F, 0x00,
		0x12, 0x34, 0x56,
		0xFF, 0xD9,
	}
}

func makeJPEGNoEXIF() []byte {
	return []byte{
		0xFF, 0xD8,
		0xFF, 0xDB, 0x00, 0x04, 0x00, 0x00,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x02, 0x03, 0x00, 0x3F, 0x00,
		0x01, 0x02, 0x03,
		0xFF, 0xD9,
	}
}

func TestExtractEXIFSegment(t *testing.T) {
	seg, err := extractEXIFSegment(makeJPEGWithEXIF())
	if err != nil {
		t.Fatalf("extract exif: %v", err)
	}
	if len(seg) == 0 {
		t.Fatal("expected EXIF segment")
	}
	if !bytes.Contains(seg, []byte("Exif\x00\x00")) {
		t.Fatalf("unexpected segment payload: %x", seg)
	}
}

func TestWriteEXIFSegment(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "out.jpg")
	if err := os.WriteFile(dst, makeJPEGNoEXIF(), 0o600); err != nil {
		t.Fatal(err)
	}

	seg, err := extractEXIFSegment(makeJPEGWithEXIF())
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteEXIFSegment(dst, seg); err != nil {
		t.Fatalf("write exif: %v", err)
	}

	out, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	got, err := extractEXIFSegment(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected EXIF segment after write")
	}
}

func TestWriteEXIFSegmentReplacesExistingExif(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "out.jpg")
	if err := os.WriteFile(dst, makeJPEGWithEXIF(), 0o600); err != nil {
		t.Fatal(err)
	}

	seg, err := extractEXIFSegment(makeJPEGWithEXIF())
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteEXIFSegment(dst, seg); err != nil {
		t.Fatalf("write exif: %v", err)
	}

	out, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if got := bytes.Count(out, []byte("Exif\x00\x00")); got != 1 {
		t.Fatalf("expected one EXIF segment, got %d", got)
	}
}

func TestWriteEXIFSegmentRejectsInvalidJPEG(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "out.jpg")
	if err := os.WriteFile(dst, []byte{0xFF}, 0o600); err != nil {
		t.Fatal(err)
	}

	seg, err := extractEXIFSegment(makeJPEGWithEXIF())
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteEXIFSegment(dst, seg); !errors.Is(err, ErrNotJPEG) {
		t.Fatalf("expected ErrNotJPEG, got %v", err)
	}
}

func TestReadEXIFSegmentFromFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "in.jpg")
	if err := os.WriteFile(p, makeJPEGWithEXIF(), 0o600); err != nil {
		t.Fatal(err)
	}

	seg, err := ReadEXIFSegment(p)
	if err != nil {
		t.Fatalf("read exif: %v", err)
	}
	if !bytes.Contains(seg, []byte("Exif\x00\x00")) {
		t.Fatalf("expected exif payload, got %x", seg)
	}
}

func TestReadEXIFSegmentNotJPEG(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "in.jpg")
	if err := os.WriteFile(p, []byte("plain text"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ReadEXIFSegment(p)
	if !errors.Is(err, ErrNotJPEG) {
		t.Fatalf("expected ErrNotJPEG, got %v", err)
	}
}

func TestReadEXIFSegmentWithoutExifReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "in.jpg")
	if err := os.WriteFile(p, makeJPEGNoEXIF(), 0o600); err != nil {
		t.Fatal(err)
	}

	seg, err := ReadEXIFSegment(p)
	if err != nil {
		t.Fatalf("read exif: %v", err)
	}
	if len(seg) != 0 {
		t.Fatalf("expected no exif segment, got %x", seg)
	}
}

func TestWriteEXIFSegmentEmptySegmentIsNoop(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "out.jpg")
	original := makeJPEGNoEXIF()
	if err := os.WriteFile(dst, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WriteEXIFSegment(dst, nil); err != nil {
		t.Fatalf("empty exif segment should be ignored, got %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("file should stay unchanged on empty segment")
	}
}

func TestWriteEXIFSegmentMissingFile(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "missing.jpg")
	seg, err := extractEXIFSegment(makeJPEGWithEXIF())
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteEXIFSegment(missing, seg); err == nil {
		t.Fatal("expected open error for missing destination")
	}
}
