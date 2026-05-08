package render

import (
	"testing"

	gd "github.com/bolknote/go-gd/v2/pkg/gd"
)

func TestApplyBezierRoundedCorners_NoRadius_NoChange(t *testing.T) {
	img, _ := gd.NewTrueColor(8, 8)
	if err := img.SetPixel(0, 0, gd.TrueColorAlpha(10, 20, 30, 0)); err != nil {
		t.Fatalf("SetPixel: %v", err)
	}
	ApplyBezierRoundedCorners(img, 0, [3]uint8{255, 255, 255}, false)
	c, err := pixelRGBA(img, 0, 0)
	if err != nil {
		t.Fatalf("pixelRGBA: %v", err)
	}
	if c.R != 10 || c.G != 20 || c.B != 30 {
		t.Fatalf("pixel should not change when radius=0, got %#v", c)
	}
}

func TestApplyBezierRoundedCorners_ChangesCornerPixel(t *testing.T) {
	img, _ := gd.NewTrueColor(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if err := img.SetPixel(x, y, gd.TrueColorAlpha(10, 20, 30, 0)); err != nil {
				t.Fatalf("SetPixel: %v", err)
			}
		}
	}
	ApplyBezierRoundedCorners(img, 4, [3]uint8{255, 255, 255}, false)
	c, err := pixelRGBA(img, 0, 0)
	if err != nil {
		t.Fatalf("pixelRGBA: %v", err)
	}
	if c.R == 10 && c.G == 20 && c.B == 30 {
		t.Fatalf("corner pixel should be blended")
	}
}

func TestApplyBezierRoundedCorners_DoesNotCutInnerSideOfMirroredCorners(t *testing.T) {
	img, _ := gd.NewTrueColor(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if err := img.SetPixel(x, y, gd.TrueColorAlpha(10, 20, 30, 0)); err != nil {
				t.Fatalf("SetPixel: %v", err)
			}
		}
	}

	ApplyBezierRoundedCorners(img, 4, [3]uint8{255, 255, 255}, false)

	innerRightTop, err := pixelRGBA(img, 6, 3)
	if err != nil {
		t.Fatalf("pixelRGBA: %v", err)
	}
	if innerRightTop.R != 10 || innerRightTop.G != 20 || innerRightTop.B != 30 {
		t.Fatalf("top-right inner side should stay unchanged, got %#v", innerRightTop)
	}

	outerRightTop, err := pixelRGBA(img, 9, 0)
	if err != nil {
		t.Fatalf("pixelRGBA: %v", err)
	}
	if outerRightTop.R == 10 && outerRightTop.G == 20 && outerRightTop.B == 30 {
		t.Fatalf("top-right outer corner should be rounded, got %#v", outerRightTop)
	}
}

func TestOutsideCornerCoverage(t *testing.T) {
	pts := quarterBezierPoints(10, defaultSteps(10))
	if got := outsideCornerCoverage(0, 0, 10, pts); got != 1 {
		t.Fatalf("top-left pixel should be fully outside, got %v", got)
	}
	if got := outsideCornerCoverage(9, 9, 10, pts); got != 0 {
		t.Fatalf("inner tangent pixel should stay untouched, got %v", got)
	}
	foundEdge := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			got := outsideCornerCoverage(x, y, 10, pts)
			if got > 0 && got < 1 {
				foundEdge = true
			}
		}
	}
	if !foundEdge {
		t.Fatal("expected at least one anti-aliased edge pixel")
	}
}

func TestApplyBezierRoundedCorners_TransparentModeUsesAlpha(t *testing.T) {
	img, _ := gd.NewTrueColor(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if err := img.SetPixel(x, y, gd.TrueColorAlpha(10, 20, 30, 0)); err != nil {
				t.Fatalf("SetPixel: %v", err)
			}
		}
	}

	ApplyBezierRoundedCorners(img, 4, [3]uint8{255, 255, 255}, true)

	corner, err := pixelRGBA(img, 0, 0)
	if err != nil {
		t.Fatalf("pixelRGBA corner: %v", err)
	}
	if corner.R != 10 || corner.G != 20 || corner.B != 30 {
		t.Fatalf("transparent mode should not blend RGB with background, got %#v", corner)
	}

	center, err := pixelRGBA(img, 5, 5)
	if err != nil {
		t.Fatalf("pixelRGBA center: %v", err)
	}
	if center.R != 10 || center.G != 20 || center.B != 30 {
		t.Fatalf("center should stay unchanged, got %#v", center)
	}
}
