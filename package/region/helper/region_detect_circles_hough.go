package regionHelper

import (
	"fmt"
	"math"
	"sort"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionDetectCirclesHough(reg *region.Region, edges []*region.EdgePoint) []*region.HoughAccumulator {
	if len(edges) < 3 {
		return []*region.HoughAccumulator{}
	}

	minRadius := 5.0
	maxRadius := math.Min(float64(reg.GetSizeX()), float64(reg.GetSizeY())) / 2.0

	accumulator := make(map[string]int)

	for _, edge := range edges {
		for radius := minRadius; radius <= maxRadius; radius += 2.0 {
			for theta := 0.0; theta < 2*math.Pi; theta += math.Pi / 18 {
				a := float64(edge.X) - radius*math.Cos(theta)
				b := float64(edge.Y) - radius*math.Sin(theta)

				if a >= 0 && a < float64(reg.GetSizeX()) && b >= 0 && b < float64(reg.GetSizeY()) {
					key := fmt.Sprintf("%.0f_%.0f_%.0f", a, b, radius)
					accumulator[key]++
				}
			}
		}
	}

	threshold := len(edges) / 10
	circles := []*region.HoughAccumulator{}

	for key, votes := range accumulator {
		if votes > threshold {
			var a, b, radius float64
			fmt.Sscanf(key, "%f_%f_%f", &a, &b, &radius)

			circles = append(circles, &region.HoughAccumulator{
				Rho:   radius,
				Theta: math.Atan2(b, a),
				Votes: votes,
			})
		}
	}

	sort.Slice(circles, func(i, j int) bool {
		return circles[i].Votes > circles[j].Votes
	})

	if len(circles) > 3 {
		circles = circles[:3]
	}

	return circles
}
