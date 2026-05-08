package render

import "testing"

func TestQuarterBezierEndpoints(t *testing.T) {
	pts := quarterBezierPoints(10, 32)
	if pts[0].x != 0 || pts[0].y != 10 {
		t.Fatalf("unexpected start point: %#v", pts[0])
	}
	last := pts[len(pts)-1]
	if last.x != 10 || last.y != 0 {
		t.Fatalf("unexpected end point: %#v", last)
	}
}

func TestYOnQuarterBezierMonotonic(t *testing.T) {
	prev := 0.0
	for x := 10.0; x >= 0; x -= 0.25 {
		y := yOnQuarterBezier(x, 10)
		if y < prev {
			t.Fatalf("curve should be monotonic by x, x=%f y=%f prev=%f", x, y, prev)
		}
		prev = y
	}
}

func TestYOnQuarterBezierBowsOutwardFromCorner(t *testing.T) {
	if got := yOnQuarterBezier(5, 10); got > 2 {
		t.Fatalf("middle of corner curve should stay near the outer edge, got %f", got)
	}
}
