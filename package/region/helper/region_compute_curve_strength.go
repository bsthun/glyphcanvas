package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionComputeCurveStrength(curvatures []float64, edges []*region.EdgePoint) float32 {
	if len(curvatures) == 0 {
		return 0
	}

	totalCurvature := 0.0
	positiveCurvature := 0.0

	for _, c := range curvatures {
		totalCurvature += math.Abs(c)
		if c > 0 {
			positiveCurvature += c
		}
	}

	avgCurvature := totalCurvature / float64(len(curvatures))

	direction := 0.0
	if len(edges) >= 2 {
		start := edges[0]
		end := edges[len(edges)-1]

		dx := float64(end.X - start.X)
		dy := float64(end.Y - start.Y)

		if dx > 0 || dy < 0 {
			direction = 1.0
		} else {
			direction = -1.0
		}
	}

	strength := math.Tanh(avgCurvature * 2)

	if totalCurvature > 0 {
		bias := (positiveCurvature / totalCurvature) - 0.5
		strength += bias * 0.3
	}

	strength *= direction

	strength = math.Max(-1.0, math.Min(1.0, strength))

	return float32(strength)
}
