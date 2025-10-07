package regionHelper

import "math"

func RegionComputeCurvatures(chainCode []int) []float64 {
	curvatures := make([]float64, len(chainCode))

	for i := 0; i < len(chainCode); i++ {
		prev := chainCode[(i-1+len(chainCode))%len(chainCode)]
		curr := chainCode[i]
		next := chainCode[(i+1)%len(chainCode)]

		angle1 := float64(curr-prev) * math.Pi / 4.0
		angle2 := float64(next-curr) * math.Pi / 4.0

		if angle1 > math.Pi {
			angle1 -= 2 * math.Pi
		} else if angle1 < -math.Pi {
			angle1 += 2 * math.Pi
		}

		if angle2 > math.Pi {
			angle2 -= 2 * math.Pi
		} else if angle2 < -math.Pi {
			angle2 += 2 * math.Pi
		}

		curvatures[i] = (angle1 + angle2) / 2.0
	}

	return curvatures
}
