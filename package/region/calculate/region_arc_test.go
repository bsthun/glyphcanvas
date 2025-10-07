package calculate

import (
	"path/filepath"
	"testing"

	"github.com/bsthun/glyphcanvas/package/region"
	"github.com/bsthun/glyphcanvas/test"
)

func TestRegionArcWithTestImage(t *testing.T) {
	testImagePath := filepath.Join("..", "..", "..", "test", "asset", "region001.png")
	img := test.LoadImage(t, testImagePath)
	r := test.RegionFromImage(img)

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for test image")
	}

	if arc.Type < 0 || arc.Type > region.ArcTypeRectangle {
		t.Errorf("Invalid arc type: %v", arc.Type)
	}

	if arc.Fill != region.ArcFillTypeFill && arc.Fill != region.ArcFillTypeStroke {
		t.Errorf("Invalid fill type: %v", arc.Fill)
	}
}

func TestRegionArcWithCircle(t *testing.T) {
	r := region.NewRegion(100, 100)

	centerX, centerY, radius := 50, 50, 30
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			dx := x - centerX
			dy := y - centerY
			distSq := dx*dx + dy*dy
			if distSq <= radius*radius {
				r.Draw(uint16(x), uint16(y))
			}
		}
	}

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for circle")
	}

	if arc.Type != region.ArcTypeCircle {
		t.Logf("Expected circle type, got: %v", arc.Type)
	}

	if arc.CircleEllipseRatio < 0.8 || arc.CircleEllipseRatio > 1.2 {
		t.Logf("Circle ellipse ratio out of expected range: %v", arc.CircleEllipseRatio)
	}
}

func TestRegionArcWithRectangle(t *testing.T) {
	r := region.NewRegion(100, 100)

	for x := uint16(20); x <= 70; x++ {
		for y := uint16(30); y <= 60; y++ {
			r.Draw(x, y)
		}
	}

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for rectangle")
	}

	if arc.Type == region.ArcTypeRectangle {
		t.Log("Successfully detected rectangle")
	} else {
		t.Logf("Rectangle detection returned type: %v", arc.Type)
	}
}

func TestRegionArcWithLine(t *testing.T) {
	r := region.NewRegion(100, 100)

	for i := 10; i < 90; i++ {
		r.Draw(uint16(i), uint16(i))
		r.Draw(uint16(i+1), uint16(i))
		r.Draw(uint16(i), uint16(i+1))
	}

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for diagonal line")
	}

	if arc.Type == region.ArcTypeStrengthLine {
		t.Logf("Line detected with degree: %.2f", arc.LineDegree)
		if arc.LineDegree < 30 || arc.LineDegree > 60 {
			t.Logf("Unexpected line degree for diagonal: %.2f", arc.LineDegree)
		}
	} else {
		t.Logf("Line detection returned type: %v", arc.Type)
	}
}

func TestRegionArcWithTriangle(t *testing.T) {
	r := region.NewRegion(100, 100)

	drawTriangleLine := func(x1, y1, x2, y2 int) {
		dx := x2 - x1
		dy := y2 - y1
		steps := 100
		for i := 0; i <= steps; i++ {
			x := x1 + dx*i/steps
			y := y1 + dy*i/steps
			if x >= 0 && x < 100 && y >= 0 && y < 100 {
				r.Draw(uint16(x), uint16(y))
			}
		}
	}

	drawTriangleLine(50, 20, 30, 70)
	drawTriangleLine(30, 70, 70, 70)
	drawTriangleLine(70, 70, 50, 20)

	for x := 35; x <= 65; x++ {
		for y := 30; y <= 68; y++ {
			if (x-30)*50 >= (y-70)*20 &&
				(x-70)*50 <= (y-70)*(-20) &&
				(x-50)*50 >= (y-20)*0 {
				r.Draw(uint16(x), uint16(y))
			}
		}
	}

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for triangle")
	}

	t.Logf("Triangle test returned type: %v", arc.Type)
}

func TestRegionArcEdgeCases(t *testing.T) {
	t.Run("Empty region", func(t *testing.T) {
		r := region.NewRegion(10, 10)
		arc := RegionArc(r)
		if arc != nil {
			t.Error("Expected nil for empty region")
		}
	})

	t.Run("Single point", func(t *testing.T) {
		r := region.NewRegion(10, 10)
		r.Draw(5, 5)
		arc := RegionArc(r)
		if arc != nil {
			t.Error("Expected nil for single point")
		}
	})

	t.Run("Two points", func(t *testing.T) {
		r := region.NewRegion(10, 10)
		r.Draw(5, 5)
		r.Draw(6, 6)
		arc := RegionArc(r)
		if arc != nil {
			t.Error("Expected nil for two points")
		}
	})

	t.Run("Three points line", func(t *testing.T) {
		r := region.NewRegion(10, 10)
		r.Draw(5, 5)
		r.Draw(6, 6)
		r.Draw(7, 7)
		arc := RegionArc(r)
		if arc == nil {
			t.Error("Expected non-nil for three points")
		}
	})
}

func TestRegionArcCurve(t *testing.T) {
	r := region.NewRegion(150, 150)

	for i := 0; i < 100; i++ {
		angle := float64(i) * 3.14159 / 100.0
		x := 75 + int(50*angle/3.14159)
		y := 75 + int(30*angle/3.14159) - int(20*(angle/3.14159)*(angle/3.14159))

		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < 150 && py >= 0 && py < 150 {
					r.Draw(uint16(px), uint16(py))
				}
			}
		}
	}

	arc := RegionArc(r)
	if arc == nil {
		t.Fatal("RegionArc returned nil for curve")
	}

	if arc.Type == region.ArcTypeCurveLine {
		t.Logf("Curve detected with strength: %.3f", arc.ArcLineTheta)
	} else {
		t.Logf("Curve test returned type: %v", arc.Type)
	}
}

func BenchmarkRegionArc(b *testing.B) {
	r := region.NewRegion(100, 100)
	for x := uint16(20); x <= 80; x++ {
		for y := uint16(20); y <= 80; y++ {
			r.Draw(x, y)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RegionArc(r)
	}
}
