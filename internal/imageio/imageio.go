package imageio

import (
	"fmt"

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
)

func Load(path string) (*gd.Image, error) {
	if _, err := DetectFormat(path); err != nil {
		return nil, err
	}
	im, err := gd.DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return im, nil
}

// ResizeIfNeeded returns the original image when no resize is necessary or a
// freshly scaled image otherwise. Ownership of the original image stays with
// the caller (it is never closed here).
func ResizeIfNeeded(img *gd.Image, width int) (*gd.Image, error) {
	if width <= 0 || width == img.Width() {
		return img, nil
	}
	h := max(int(float64(img.Height())*float64(width)/float64(img.Width())), 1)
	return img.Scale(uint(width), uint(h))
}

func Save(img *gd.Image, path string, quality int) error {
	if err := encodeFile(img, path, quality); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}

func encodeFile(img *gd.Image, path string, quality int) error {
	format, err := DetectFormat(path)
	if err != nil {
		return err
	}
	if !gd.SupportsFormat(format, true) {
		return fmt.Errorf("linked libgd does not support %s encoding", format)
	}

	q := clampQuality(quality)
	switch format {
	case gd.FormatJPEG:
		return img.EncodeJPEGFile(path, &gd.JPEGOptions{Quality: q})
	case gd.FormatPNG:
		return img.EncodePNGFile(path, nil)
	case gd.FormatGIF:
		return img.EncodeGIFFile(path)
	case gd.FormatWebP:
		return img.EncodeWebPFile(path, &gd.WebPOptions{Quality: q})
	case gd.FormatWBMP:
		return img.EncodeWBMPFile(path, gd.TrueColorAlpha(0, 0, 0, 0))
	case gd.FormatBMP:
		return img.EncodeBMPFile(path, nil)
	case gd.FormatTIFF:
		return img.EncodeTIFFFile(path)
	case gd.FormatGD:
		return img.EncodeGDFile(path)
	case gd.FormatGD2:
		return img.EncodeGD2File(path, nil)
	case gd.FormatHEIF:
		return img.EncodeHEIFFile(path, &gd.HEIFOptions{Quality: q})
	case gd.FormatAVIF:
		return img.EncodeAVIFFile(path, &gd.AVIFOptions{Quality: q})
	case gd.FormatXBM:
		return img.EncodeXBMFile(path, "corner", gd.TrueColorAlpha(0, 0, 0, 0))
	default:
		return img.EncodeFile(path)
	}
}

func clampQuality(quality int) int {
	if quality < 1 {
		return 1
	}
	if quality > 100 {
		return 100
	}
	return quality
}
