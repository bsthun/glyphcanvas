package regionHelper

import "math"

func RegionComputeRectangularity(hu []float64) float64 {
	if len(hu) < 7 {
		return 0
	}

	rectangleHu := []float64{0.16, 0.0013, 0, 0, 0, 0, 0}

	diff := 0.0
	for i := 0; i < 7; i++ {
		diff += math.Abs(hu[i] - rectangleHu[i])
	}

	return math.Exp(-diff * 10)
}
