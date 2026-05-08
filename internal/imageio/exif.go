package imageio

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var ErrNotJPEG = errors.New("not a jpeg file")

// JPEG marker bytes (the byte that follows 0xFF).
const (
	markerSOI  = 0xD8 // start of image
	markerEOI  = 0xD9 // end of image
	markerSOS  = 0xDA // start of scan (entropy-coded data follows)
	markerAPP1 = 0xE1 // application-specific segment 1 (Exif lives here)
)

// markerStandalone reports whether the JPEG marker has no length/payload (TEM
// or any of the RST_n restart markers). Such markers are exactly two bytes on
// the wire (0xFF + marker).
func markerStandalone(marker byte) bool {
	const (
		markerTEM   = 0x01
		markerRST0  = 0xD0
		markerRST7  = 0xD7
	)
	return marker == markerTEM || (marker >= markerRST0 && marker <= markerRST7)
}

func ReadEXIFSegment(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return extractEXIFSegment(data)
}

// WriteEXIFSegment replaces (or inserts) the Exif APP1 segment in the JPEG at
// path. The write is atomic: the modified payload is staged in a sibling temp
// file and then renamed into place, preserving the original file mode.
func WriteEXIFSegment(path string, exifSegment []byte) error {
	if len(exifSegment) == 0 {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	cleaned, err := stripEXIFSegments(data)
	if err != nil {
		return err
	}
	out := make([]byte, 0, len(cleaned)+len(exifSegment))
	out = append(out, cleaned[:2]...)
	out = append(out, exifSegment...)
	out = append(out, cleaned[2:]...)
	return atomicWriteFile(path, out, info.Mode().Perm())
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// scanJPEG walks a JPEG byte stream after the SOI marker. For every recognized
// segment it calls visit(marker, start, end) where the slice [start:end]
// contains the marker, the optional 16-bit length and the payload. Iteration
// stops at SOS, EOI, or on malformed data; the returned tail index points at
// the start of the bytes that should be copied verbatim by the caller.
func scanJPEG(jpeg []byte, visit func(marker byte, start, end int)) (tail int, ok bool) {
	if len(jpeg) < 2 || jpeg[0] != 0xFF || jpeg[1] != markerSOI {
		return 0, false
	}
	i := 2
	for i+1 < len(jpeg) {
		if jpeg[i] != 0xFF {
			return i, true
		}
		marker := jpeg[i+1]
		switch {
		case marker == markerEOI, marker == markerSOS:
			return i, true
		case markerStandalone(marker):
			visit(marker, i, i+2)
			i += 2
			continue
		}
		if i+4 > len(jpeg) {
			return i, true
		}
		length := int(binary.BigEndian.Uint16(jpeg[i+2 : i+4]))
		if length < 2 || i+2+length > len(jpeg) {
			return i, true
		}
		end := i + 2 + length
		visit(marker, i, end)
		i = end
	}
	return i, true
}

// isExifAPP1 reports whether the segment at jpeg[start:end] is an APP1 segment
// carrying an Exif identifier.
func isExifAPP1(jpeg []byte, marker byte, start, end int) bool {
	if marker != markerAPP1 || end-start < 10 {
		return false
	}
	return string(jpeg[start+4:start+10]) == "Exif\x00\x00"
}

func extractEXIFSegment(jpeg []byte) ([]byte, error) {
	var found []byte
	_, ok := scanJPEG(jpeg, func(marker byte, start, end int) {
		if found == nil && isExifAPP1(jpeg, marker, start, end) {
			found = append([]byte(nil), jpeg[start:end]...)
		}
	})
	if !ok {
		return nil, ErrNotJPEG
	}
	return found, nil
}

func stripEXIFSegments(jpeg []byte) ([]byte, error) {
	if len(jpeg) < 2 {
		return nil, ErrNotJPEG
	}
	out := make([]byte, 0, len(jpeg))
	out = append(out, jpeg[:2]...)
	tail, ok := scanJPEG(jpeg, func(marker byte, start, end int) {
		if isExifAPP1(jpeg, marker, start, end) {
			return
		}
		out = append(out, jpeg[start:end]...)
	})
	if !ok {
		return nil, ErrNotJPEG
	}
	out = append(out, jpeg[tail:]...)
	return out, nil
}
