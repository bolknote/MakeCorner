package render

import gd "github.com/bolknote/go-gd/v2/pkg/gd"

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

func setBlended(img *gd.Image, x, y int, bg [3]uint8, t float64) {
	orig, err := pixelRGBA(img, x, y)
	if err != nil {
		return
	}
	inv := 1.0 - t
	_ = img.SetPixel(x, y, gd.TrueColorAlpha(
		int(inv*float64(orig.R)+t*float64(bg[0])),
		int(inv*float64(orig.G)+t*float64(bg[1])),
		int(inv*float64(orig.B)+t*float64(bg[2])),
		0,
	))
}

func setTransparent(img *gd.Image, x, y int, t float64) {
	if t <= 0 {
		return
	}
	if t > 1 {
		t = 1
	}
	orig, err := pixelRGBA(img, x, y)
	if err != nil {
		return
	}
	alpha := int(t * 127)
	_ = img.SetPixel(x, y, gd.TrueColorAlpha(int(orig.R), int(orig.G), int(orig.B), alpha))
}

func pixelRGBA(img *gd.Image, x, y int) (gd.RGBA, error) {
	c, err := img.Pixel(x, y)
	if err != nil {
		return gd.RGBA{}, err
	}
	return img.ColorRGBA(c)
}
