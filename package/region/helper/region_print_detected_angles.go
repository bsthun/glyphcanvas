package regionHelper

import (
	"fmt"
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionPrintDetectedAngles(edges []*region.EdgePoint) {
	if len(edges) == 0 {
		return
	}

	angleHistogram := make(map[int]int)
	targetAngles := []int{0, 45, 90, 135, 180}

	for _, edge := range edges {
		degree := edge.Angle * 180.0 / math.Pi

		if degree < 0 {
			degree += 180
		}

		for _, target := range targetAngles {
			if math.Abs(degree-float64(target)) < 22.5 ||
				math.Abs(degree-float64(target)+180) < 22.5 ||
				math.Abs(degree-float64(target)-180) < 22.5 {
				angleHistogram[target]++
				break
			}
		}
	}

	fmt.Println("Detected angle distribution in region:")
	for _, angle := range targetAngles {
		count := angleHistogram[angle]
		percentage := float64(count) * 100.0 / float64(len(edges))
		if count > 0 {
			fmt.Printf("  %3d°: %d edges (%.1f%%)\n", angle, count, percentage)
		}
	}
}
