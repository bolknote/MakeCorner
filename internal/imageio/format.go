package imageio

import (
	"fmt"
	"path/filepath"
	"strings"

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
)

const FormatJPEG = gd.FormatJPEG

func DetectFormat(path string) (gd.Format, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "jpg", "jpeg":
		return gd.FormatJPEG, nil
	case "png":
		return gd.FormatPNG, nil
	case "gif":
		return gd.FormatGIF, nil
	case "wbmp":
		return gd.FormatWBMP, nil
	case "webp":
		return gd.FormatWebP, nil
	case "bmp":
		return gd.FormatBMP, nil
	case "tga":
		return gd.FormatTGA, nil
	case "tif", "tiff":
		return gd.FormatTIFF, nil
	case "gd":
		return gd.FormatGD, nil
	case "gd2":
		return gd.FormatGD2, nil
	case "heif", "heic":
		return gd.FormatHEIF, nil
	case "avif":
		return gd.FormatAVIF, nil
	case "xpm":
		return gd.FormatXPM, nil
	case "xbm":
		return gd.FormatXBM, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %q", ext)
	}
}

// SupportsAlpha reports whether the encoder for format preserves per-pixel
// alpha. GIF/WBMP/TGA are intentionally excluded:
//   - GIF supports only one binary-transparent palette index;
//   - WBMP is monochrome;
//   - TGA encoding via libgd is fragile and not relied upon.
func SupportsAlpha(format gd.Format) bool {
	switch format {
	case gd.FormatPNG, gd.FormatWebP, gd.FormatHEIF, gd.FormatAVIF, gd.FormatTIFF:
		return true
	default:
		return false
	}
}
