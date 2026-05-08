package render

import (
	"math"

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
)

const cornerSamples = 8

func ApplyBezierRoundedCorners(img *gd.Image, radius int, bg [3]uint8, transparent bool) {
	if transparent {
		_ = img.AlphaBlending(false)
		_ = img.SaveAlpha(true)
	}
	applyRoundedCorners(img, radius, func(x, y int, coverage float64) {
		if transparent {
			setTransparent(img, x, y, coverage)
			return
		}
		setBlended(img, x, y, bg, coverage)
	})
}

func applyRoundedCorners(img *gd.Image, radius int, apply func(x, y int, coverage float64)) {
	if radius <= 0 {
		return
	}
	w, h := img.Width(), img.Height()
	if w == 0 || h == 0 {
		return
	}
	if maxR := min(w, h) / 2; radius > maxR {
		radius = maxR
	}
	size := radius + 1
	pts := quarterBezierPoints(float64(radius), defaultSteps(float64(radius)))

	apply4 := func(maskX, maskY, dstX, dstY int) {
		if c := outsideCornerCoverage(maskX, maskY, radius, pts); c > 0 {
			apply(dstX, dstY, c)
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			apply4(x, y, x, y)
			apply4(size-1-x, y, w-size+x, y)
			apply4(x, size-1-y, x, h-size+y)
			apply4(size-1-x, size-1-y, w-size+x, h-size+y)
		}
	}
}

func outsideCornerCoverage(x, y, radius int, pts []point) float64 {
	if radius <= 0 {
		return 0
	}
	const total = cornerSamples * cornerSamples
	r := float64(radius)
	outside := 0
	for sy := 0; sy < cornerSamples; sy++ {
		py := float64(y) + (float64(sy)+0.5)/cornerSamples
		for sx := 0; sx < cornerSamples; sx++ {
			px := float64(x) + (float64(sx)+0.5)/cornerSamples
			if py < yOnPoints(pts, px, r) {
				outside++
			}
		}
	}
	return float64(outside) / float64(total)
}

// setBlended replaces (x,y) with the linear blend between the original pixel
// and bg, weighted by the coverage t in [0, 1] (clamped). Channel arithmetic
// uses math.Round to avoid the systematic rounding-toward-zero bias that
// int(...) would introduce on every blended pixel.
//
// Note: gd treats alpha=0 as fully opaque (gd alpha range is 0..127).
func setBlended(img *gd.Image, x, y int, bg [3]uint8, t float64) {
	t = clamp01(t)
	orig, err := pixelRGBA(img, x, y)
	if err != nil {
		return
	}
	inv := 1.0 - t
	_ = img.SetPixel(x, y, gd.TrueColorAlpha(
		blend8(inv, float64(orig.R), t, float64(bg[0])),
		blend8(inv, float64(orig.G), t, float64(bg[1])),
		blend8(inv, float64(orig.B), t, float64(bg[2])),
		0,
	))
}

// setTransparent marks (x,y) with a per-pixel alpha proportional to the
// coverage t. The original RGB stays intact so transparent corners remain
// faithful to the source image when previewed against another background.
func setTransparent(img *gd.Image, x, y int, t float64) {
	t = clamp01(t)
	if t == 0 {
		return
	}
	orig, err := pixelRGBA(img, x, y)
	if err != nil {
		return
	}
	alpha := int(math.Round(t * 127))
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 127 {
		alpha = 127
	}
	_ = img.SetPixel(x, y, gd.TrueColorAlpha(int(orig.R), int(orig.G), int(orig.B), alpha))
}

func pixelRGBA(img *gd.Image, x, y int) (gd.RGBA, error) {
	c, err := img.Pixel(x, y)
	if err != nil {
		return gd.RGBA{}, err
	}
	return img.ColorRGBA(c)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func blend8(invT, src, t, dst float64) int {
	v := int(math.Round(invT*src + t*dst))
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
