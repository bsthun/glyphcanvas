package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionDetectCorners(curvatures []float64, edges []*region.EdgePoint) []int {
	corners := []int{}
	threshold := math.Pi / 6

	for i := 0; i < len(curvatures); i++ {
		if math.Abs(curvatures[i]) > threshold {
			isLocalMax := true
			for j := i - 2; j <= i+2; j++ {
				if j < 0 || j >= len(curvatures) || j == i {
					continue
				}
				if math.Abs(curvatures[j]) > math.Abs(curvatures[i]) {
					isLocalMax = false
					break
				}
			}

			if isLocalMax {
				corners = append(corners, i)
			}
		}
	}

	return corners
}
