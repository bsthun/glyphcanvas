package regionHelper

import "math"

func RegionComputeLinearity(hu []float64) float64 {
	if len(hu) < 3 {
		return 0
	}

	sum := 0.0
	for i := 2; i < len(hu); i++ {
		sum += math.Abs(hu[i])
	}

	if sum < 0.01 {
		return 1.0
	}

	return 1.0 / (1.0 + sum*100)
}
