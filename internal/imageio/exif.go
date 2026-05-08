package imageio

import (
	"encoding/binary"
	"errors"
	"os"
)

var ErrNotJPEG = errors.New("not a jpeg file")

const (
	markerSOI = 0xD8
	markerEOI = 0xD9
	markerSOS = 0xDA
	markerAPP1 = 0xE1
)

func ReadEXIFSegment(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return extractEXIFSegment(data)
}

func WriteEXIFSegment(path string, exifSegment []byte) error {
	if len(exifSegment) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cleaned, err := stripEXIFSegments(data)
	if err != nil {
		return err
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	out := make([]byte, 0, len(cleaned)+len(exifSegment))
	out = append(out, cleaned[:2]...)
	out = append(out, exifSegment...)
	out = append(out, cleaned[2:]...)
	return os.WriteFile(path, out, st.Mode().Perm())
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
		case marker == 0x01, marker >= 0xD0 && marker <= 0xD7:
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
