package regionHelper

import (
	"testing"

	"github.com/bsthun/glyphcanvas/package/region"
)

func TestRegionComputeMoments(t *testing.T) {
	tests := []struct {
		name          string
		setupRegion   func() *region.Region
		expectedM00   float64
		expectedCentX float64
		expectedCentY float64
	}{
		{
			name: "3x3 square region",
			setupRegion: func() *region.Region {
				r := region.NewRegion(5, 5)
				for x := uint16(1); x <= 3; x++ {
					for y := uint16(1); y <= 3; y++ {
						r.Draw(x, y)
					}
				}
				return r
			},
			expectedM00:   9.0,
			expectedCentX: 2.0,
			expectedCentY: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRegion()
			moments := RegionComputeMoments(r)

			if moments["m00"] != tt.expectedM00 {
				t.Errorf("m00 = %v, want %v", moments["m00"], tt.expectedM00)
			}

			if moments["m00"] > 0 {
				centX := moments["m10"] / moments["m00"]
				centY := moments["m01"] / moments["m00"]

				if centX != tt.expectedCentX {
					t.Errorf("centroid X = %v, want %v", centX, tt.expectedCentX)
				}
				if centY != tt.expectedCentY {
					t.Errorf("centroid Y = %v, want %v", centY, tt.expectedCentY)
				}
			}
		})
	}
}

func BenchmarkRegionComputeMoments(b *testing.B) {
	r := region.NewRegion(100, 100)
	for x := uint16(20); x <= 80; x++ {
		for y := uint16(20); y <= 80; y++ {
			r.Draw(x, y)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RegionComputeMoments(r)
	}
}
