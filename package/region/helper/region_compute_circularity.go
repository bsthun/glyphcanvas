package regionHelper

import "math"

func RegionComputeCircularity(hu []float64) float64 {
	if len(hu) < 2 {
		return 0
	}

	I1 := hu[0]
	I2 := hu[1]

	if I1 > 0 {
		circularity := 1.0 / (1.0 + math.Sqrt(I2)/I1)
		return circularity
	}

	return 0
}
