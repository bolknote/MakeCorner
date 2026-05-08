package render

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestQuarterBezierGolden(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "testdata", "bezier_quarter_r10.golden"))
	if err != nil {
		t.Fatalf("open golden: %v", err)
	}
	defer func() { _ = f.Close() }()

	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.Split(s.Text(), ",")
		if len(parts) != 2 {
			t.Fatalf("bad golden line: %q", s.Text())
		}
		x, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			t.Fatalf("bad x in golden line %q: %v", s.Text(), err)
		}
		want, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			t.Fatalf("bad y in golden line %q: %v", s.Text(), err)
		}
		got := yOnQuarterBezier(x, 10)
		if diff := abs(got - want); diff > 0.2 {
			t.Fatalf("x=%.2f got=%.4f want=%.4f diff=%.4f", x, got, want, diff)
		}
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan golden: %v", err)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func Example_quarterBezier() {
	fmt.Printf("%.2f\n", yOnQuarterBezier(5, 10))
	// Output: 1.34
}
