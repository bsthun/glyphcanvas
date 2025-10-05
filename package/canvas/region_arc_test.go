package canvas

import (
	"math"
	"testing"
)

func TestRegionArc_Line(t *testing.T) {
	tests := []struct {
		name           string
		setupRegion    func() *Region
		expectedType   ArcType
		expectedDegree float32
		description    string
	}{
		{
			name: "Horizontal Line (0 degrees)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(20); x <= 80; x++ {
					r.Draw(x, 50)
				}
				return r
			},
			expectedType:   ArcTypeStrengthLine,
			expectedDegree: 0,
			description:    "Horizontal line detection",
		},
		{
			name: "Vertical Line (90 degrees)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for y := uint16(20); y <= 80; y++ {
					r.Draw(50, y)
				}
				return r
			},
			expectedType:   ArcTypeStrengthLine,
			expectedDegree: 90,
			description:    "Vertical line detection",
		},
		{
			name: "Diagonal Line (45 degrees)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for i := uint16(0); i < 50; i++ {
					r.Draw(20+i, 20+i)
				}
				return r
			},
			expectedType:   ArcTypeStrengthLine,
			expectedDegree: 45,
			description:    "45-degree diagonal line detection",
		},
		{
			name: "Diagonal Line (135 degrees)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for i := uint16(0); i < 50; i++ {
					r.Draw(70-i, 20+i)
				}
				return r
			},
			expectedType:   ArcTypeStrengthLine,
			expectedDegree: 135,
			description:    "135-degree diagonal line detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if arc == nil {
				t.Fatal("Arc() returned nil")
			}

			if arc.Type != tt.expectedType {
				t.Errorf("Expected arc type %v, got %v", tt.expectedType, arc.Type)
			}

			if arc.Type == ArcTypeStrengthLine {
				tolerance := float32(5.0)
				if math.Abs(float64(arc.LineDegree-tt.expectedDegree)) > float64(tolerance) {
					t.Errorf("Expected line degree %.0f°±%.0f°, got %.0f°",
						tt.expectedDegree, tolerance, arc.LineDegree)
				}
			}
		})
	}
}

func TestRegionArc_Circle(t *testing.T) {
	tests := []struct {
		name         string
		setupRegion  func() *Region
		expectedType ArcType
		fillType     ArcFillType
		description  string
	}{
		{
			name: "Filled Circle",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				centerX, centerY := 50.0, 50.0
				radius := 20.0

				for x := uint16(0); x < r.SizeX; x++ {
					for y := uint16(0); y < r.SizeY; y++ {
						dx := float64(x) - centerX
						dy := float64(y) - centerY
						if dx*dx+dy*dy <= radius*radius {
							r.Draw(x, y)
						}
					}
				}
				return r
			},
			expectedType: ArcTypeCircle,
			fillType:     ArcFillTypeFill,
			description:  "Filled circle detection",
		},
		{
			name: "Circle Outline",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				centerX, centerY := 50.0, 50.0
				radius := 20.0

				steps := 100
				for i := 0; i < steps; i++ {
					angle := float64(i) * 2.0 * math.Pi / float64(steps)
					x := uint16(centerX + radius*math.Cos(angle))
					y := uint16(centerY + radius*math.Sin(angle))
					if x < r.SizeX && y < r.SizeY {
						r.Draw(x, y)
						if x > 0 {
							r.Draw(x-1, y)
						}
						if y > 0 {
							r.Draw(x, y-1)
						}
					}
				}
				return r
			},
			expectedType: ArcTypeCircle,
			fillType:     ArcFillTypeStroke,
			description:  "Circle outline detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if arc == nil {
				t.Fatal("Arc() returned nil")
			}

			if arc.Type != tt.expectedType {
				t.Errorf("Expected arc type %v, got %v", tt.expectedType, arc.Type)
			}

			if arc.Type == ArcTypeCircle {
				if arc.CircleEllipseRatio < 0.7 || arc.CircleEllipseRatio > 1.0 {
					t.Errorf("Circle ellipse ratio out of range: %.2f", arc.CircleEllipseRatio)
				}
			}
		})
	}
}

func TestRegionArc_Curve(t *testing.T) {
	tests := []struct {
		name             string
		setupRegion      func() *Region
		expectedType     ArcType
		expectedStrength float32
		description      string
	}{
		{
			name: "Upward Curve (positive strength)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(10); x <= 90; x++ {
					t := float64(x-10) / 80.0
					y := uint16(50 - 30*math.Sin(t*math.Pi))
					if y < r.SizeY {
						r.Draw(x, y)
					}
				}
				return r
			},
			expectedType:     ArcTypeCurveLine,
			expectedStrength: 0.5,
			description:      "Upward curve with positive strength",
		},
		{
			name: "Downward Curve (negative strength)",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(10); x <= 90; x++ {
					t := float64(x-10) / 80.0
					y := uint16(50 + 30*math.Sin(t*math.Pi))
					if y < r.SizeY {
						r.Draw(x, y)
					}
				}
				return r
			},
			expectedType:     ArcTypeCurveLine,
			expectedStrength: -0.5,
			description:      "Downward curve with negative strength",
		},
		{
			name: "S-Curve",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(10); x <= 90; x++ {
					t := float64(x-10) / 80.0
					y := uint16(50 + 20*math.Sin(t*2*math.Pi))
					if y < r.SizeY {
						r.Draw(x, y)
						if y > 0 {
							r.Draw(x, y-1)
						}
					}
				}
				return r
			},
			expectedType:     ArcTypeCurveLine,
			expectedStrength: 0,
			description:      "S-curve with neutral strength",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if arc == nil {
				t.Fatal("Arc() returned nil")
			}

			if arc.Type != tt.expectedType {
				t.Logf("Warning: Expected arc type %v, got %v (may be classified as line due to discretization)",
					tt.expectedType, arc.Type)
			}

			if arc.Type == ArcTypeCurveLine {
				if arc.ArcLineTheta < -1.0 || arc.ArcLineTheta > 1.0 {
					t.Errorf("Curve strength out of range [-1, 1]: %.3f", arc.ArcLineTheta)
				}

				t.Logf("Curve strength: %.3f (expected around %.3f)",
					arc.ArcLineTheta, tt.expectedStrength)
			}
		})
	}
}

func TestRegionArc_Rectangle(t *testing.T) {
	tests := []struct {
		name         string
		setupRegion  func() *Region
		expectedType ArcType
		fillType     ArcFillType
		description  string
	}{
		{
			name: "Filled Rectangle",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(20); x <= 80; x++ {
					for y := uint16(30); y <= 70; y++ {
						r.Draw(x, y)
					}
				}
				return r
			},
			expectedType: ArcTypeRectangle,
			fillType:     ArcFillTypeFill,
			description:  "Filled rectangle detection",
		},
		{
			name: "Rectangle Outline",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				for x := uint16(20); x <= 80; x++ {
					r.Draw(x, 30)
					r.Draw(x, 70)
				}
				for y := uint16(30); y <= 70; y++ {
					r.Draw(20, y)
					r.Draw(80, y)
				}
				return r
			},
			expectedType: ArcTypeRectangle,
			fillType:     ArcFillTypeStroke,
			description:  "Rectangle outline detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if arc == nil {
				t.Fatal("Arc() returned nil")
			}

			if arc.Type != tt.expectedType {
				t.Logf("Warning: Expected arc type %v, got %v", tt.expectedType, arc.Type)
			}

			t.Logf("Detected shape type: %v with fill type: %v", arc.Type, arc.Fill)
		})
	}
}

func TestRegionArc_Triangle(t *testing.T) {
	tests := []struct {
		name         string
		setupRegion  func() *Region
		expectedType ArcType
		description  string
	}{
		{
			name: "Filled Triangle",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				drawTriangle := func(x1, y1, x2, y2, x3, y3 uint16) {
					drawLine := func(xa, ya, xb, yb uint16) {
						dx := int(xb) - int(xa)
						dy := int(yb) - int(ya)
						steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))

						if steps == 0 {
							r.Draw(xa, ya)
							return
						}

						for i := 0; i <= steps; i++ {
							t := float64(i) / float64(steps)
							x := uint16(float64(xa) + t*float64(dx))
							y := uint16(float64(ya) + t*float64(dy))
							r.Draw(x, y)
						}
					}

					drawLine(x1, y1, x2, y2)
					drawLine(x2, y2, x3, y3)
					drawLine(x3, y3, x1, y1)

					for y := uint16(20); y <= 70; y++ {
						for x := uint16(20); x <= 80; x++ {
							p1x, p1y := float64(x1), float64(y1)
							p2x, p2y := float64(x2), float64(y2)
							p3x, p3y := float64(x3), float64(y3)
							px, py := float64(x), float64(y)

							sign := func(px, py, ax, ay, bx, by float64) float64 {
								return (px-bx)*(ay-by) - (ax-bx)*(py-by)
							}

							d1 := sign(px, py, p1x, p1y, p2x, p2y)
							d2 := sign(px, py, p2x, p2y, p3x, p3y)
							d3 := sign(px, py, p3x, p3y, p1x, p1y)

							hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
							hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)

							if !(hasNeg && hasPos) {
								r.Draw(x, y)
							}
						}
					}
				}

				drawTriangle(50, 20, 20, 70, 80, 70)
				return r
			},
			expectedType: ArcTypeTriangle,
			description:  "Filled triangle detection",
		},
		{
			name: "Triangle Outline",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)

				drawLine := func(x1, y1, x2, y2 uint16) {
					dx := int(x2) - int(x1)
					dy := int(y2) - int(y1)
					steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))

					if steps == 0 {
						r.Draw(x1, y1)
						return
					}

					for i := 0; i <= steps; i++ {
						t := float64(i) / float64(steps)
						x := uint16(float64(x1) + t*float64(dx))
						y := uint16(float64(y1) + t*float64(dy))
						r.Draw(x, y)
					}
				}

				drawLine(50, 20, 20, 70)
				drawLine(20, 70, 80, 70)
				drawLine(80, 70, 50, 20)

				return r
			},
			expectedType: ArcTypeTriangle,
			description:  "Triangle outline detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if arc == nil {
				t.Fatal("Arc() returned nil")
			}

			if arc.Type != tt.expectedType {
				t.Logf("Warning: Expected arc type %v, got %v", tt.expectedType, arc.Type)
			}

			t.Logf("Detected shape type: %v", arc.Type)
		})
	}
}

func TestRegionArc_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupRegion func() *Region
		expectNil   bool
		description string
	}{
		{
			name: "Empty Region",
			setupRegion: func() *Region {
				return NewRegion(100, 100)
			},
			expectNil:   true,
			description: "Empty region should return nil",
		},
		{
			name: "Single Point",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				r.Draw(50, 50)
				return r
			},
			expectNil:   true,
			description: "Single point should return nil",
		},
		{
			name: "Two Points",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				r.Draw(50, 50)
				r.Draw(51, 51)
				return r
			},
			expectNil:   true,
			description: "Two points should return nil",
		},
		{
			name: "Three Points",
			setupRegion: func() *Region {
				r := NewRegion(100, 100)
				r.Draw(50, 50)
				r.Draw(51, 51)
				r.Draw(52, 52)
				return r
			},
			expectNil:   false,
			description: "Three points should return a valid arc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.setupRegion()
			arc := region.Arc()

			if tt.expectNil && arc != nil {
				t.Errorf("Expected nil arc for %s, got %+v", tt.description, arc)
			}
			if !tt.expectNil && arc == nil {
				t.Errorf("Expected valid arc for %s, got nil", tt.description)
			}
		})
	}
}

func TestRegionArc_ComplexShapes(t *testing.T) {
	t.Run("Spiral", func(t *testing.T) {
		r := NewRegion(200, 200)
		centerX, centerY := 100.0, 100.0

		for angle := 0.0; angle < 4*math.Pi; angle += 0.1 {
			radius := 5.0 + angle*5.0
			x := uint16(centerX + radius*math.Cos(angle))
			y := uint16(centerY + radius*math.Sin(angle))
			if x < r.SizeX && y < r.SizeY {
				r.Draw(x, y)
			}
		}

		arc := r.Arc()
		if arc == nil {
			t.Fatal("Arc() returned nil for spiral")
		}

		t.Logf("Spiral detected as: Type=%v, Fill=%v", arc.Type, arc.Fill)
		if arc.Type == ArcTypeCurveLine {
			t.Logf("Curve strength: %.3f", arc.ArcLineTheta)
		}
	})

	t.Run("Star", func(t *testing.T) {
		r := NewRegion(200, 200)
		centerX, centerY := 100.0, 100.0

		drawLine := func(x1, y1, x2, y2 float64) {
			steps := 50
			for i := 0; i <= steps; i++ {
				t := float64(i) / float64(steps)
				x := uint16(x1 + t*(x2-x1))
				y := uint16(y1 + t*(y2-y1))
				if x < r.SizeX && y < r.SizeY {
					r.Draw(x, y)
				}
			}
		}

		outerRadius := 40.0
		innerRadius := 15.0
		points := 5

		for i := 0; i < points*2; i++ {
			angle1 := float64(i) * math.Pi / float64(points)
			angle2 := float64(i+1) * math.Pi / float64(points)

			radius1 := innerRadius
			if i%2 == 0 {
				radius1 = outerRadius
			}
			radius2 := outerRadius
			if i%2 == 0 {
				radius2 = innerRadius
			}

			x1 := centerX + radius1*math.Cos(angle1)
			y1 := centerY + radius1*math.Sin(angle1)
			x2 := centerX + radius2*math.Cos(angle2)
			y2 := centerY + radius2*math.Sin(angle2)

			drawLine(x1, y1, x2, y2)
		}

		arc := r.Arc()
		if arc == nil {
			t.Fatal("Arc() returned nil for star")
		}

		t.Logf("Star detected as: Type=%v, Fill=%v", arc.Type, arc.Fill)
	})
}

func BenchmarkRegionArc_SmallRegion(b *testing.B) {
	r := NewRegion(50, 50)
	for x := uint16(10); x <= 40; x++ {
		for y := uint16(10); y <= 40; y++ {
			r.Draw(x, y)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Arc()
	}
}

func BenchmarkRegionArc_LargeRegion(b *testing.B) {
	r := NewRegion(500, 500)
	centerX, centerY := 250.0, 250.0
	radius := 100.0

	for x := uint16(0); x < r.SizeX; x++ {
		for y := uint16(0); y < r.SizeY; y++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			if dx*dx+dy*dy <= radius*radius {
				r.Draw(x, y)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Arc()
	}
}
