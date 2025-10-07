package regionHelper

import "math"

func RegionComputeEllipseRatio(moments map[string]float64) float32 {
	mu20 := moments["mu20"]
	mu02 := moments["mu02"]
	mu11 := moments["mu11"]

	if mu20+mu02 == 0 {
		return 1.0
	}

	lambda1 := (mu20 + mu02 + math.Sqrt(math.Pow(mu20-mu02, 2)+4*mu11*mu11)) / 2
	lambda2 := (mu20 + mu02 - math.Sqrt(math.Pow(mu20-mu02, 2)+4*mu11*mu11)) / 2

	if lambda1 > 0 && lambda2 > 0 {
		ratio := math.Min(lambda1, lambda2) / math.Max(lambda1, lambda2)
		return float32(ratio)
	}

	return 1.0
}
