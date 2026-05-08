package render

import "math"

const quarterCircleKappa = 0.5522847498307936

type point struct {
	x float64
	y float64
}

// quarterBezierPoints returns a sampled polyline approximating the quarter
// Bezier curve from (0,radius) to (radius,0) bowing toward the origin.
func quarterBezierPoints(radius float64, steps int) []point {
	if steps < 2 {
		steps = 2
	}
	p1y := radius - quarterCircleKappa*radius
	p2x := radius - quarterCircleKappa*radius

	pts := make([]point, steps+1)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		mt := 1 - t
		// p0=(0,radius), p1=(0,p1y), p2=(p2x,0), p3=(radius,0).
		pts[i] = point{
			x: 3*mt*t*t*p2x + t*t*t*radius,
			y: mt*mt*mt*radius + 3*mt*mt*t*p1y,
		}
	}
	return pts
}

// defaultSteps picks a sample count proportional to the radius.
func defaultSteps(radius float64) int {
	return int(math.Max(16, radius*4))
}

// yOnPoints returns the linearly interpolated y for x along the polyline pts.
// pts must come from quarterBezierPoints with the matching radius.
func yOnPoints(pts []point, x, radius float64) float64 {
	if x <= 0 {
		return radius
	}
	if x >= radius {
		return 0
	}
	for i := 0; i < len(pts)-1; i++ {
		left, right := pts[i], pts[i+1]
		if left.x <= x && x <= right.x {
			if right.x == left.x {
				return right.y
			}
			t := (x - left.x) / (right.x - left.x)
			return left.y + t*(right.y-left.y)
		}
	}
	return radius
}

func yOnQuarterBezier(x, radius float64) float64 {
	return yOnPoints(quarterBezierPoints(radius, defaultSteps(radius)), x, radius)
}
