package regionHelper

import (
	"fmt"
	"math"
	"sort"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionDetectLinesHough(reg *region.Region, edges []*region.EdgePoint) []*region.HoughAccumulator {
	if len(edges) < 2 {
		return []*region.HoughAccumulator{}
	}

	maxRho := math.Sqrt(float64(reg.GetSizeX()*reg.GetSizeX() + reg.GetSizeY()*reg.GetSizeY()))
	rhoStep := 1.0
	thetaStep := math.Pi / 180.0

	accumulator := make(map[string]int)

	for _, edge := range edges {
		for theta := 0.0; theta < math.Pi; theta += thetaStep {
			rho := float64(edge.X)*math.Cos(theta) + float64(edge.Y)*math.Sin(theta)

			rhoIdx := int((rho + maxRho) / rhoStep)
			thetaIdx := int(theta / thetaStep)

			key := fmt.Sprintf("%d_%d", rhoIdx, thetaIdx)
			accumulator[key]++
		}
	}

	threshold := len(edges) / 4
	lines := []*region.HoughAccumulator{}

	for key, votes := range accumulator {
		if votes > threshold {
			var rhoIdx, thetaIdx int
			fmt.Sscanf(key, "%d_%d", &rhoIdx, &thetaIdx)

			rho := float64(rhoIdx)*rhoStep - maxRho
			theta := float64(thetaIdx) * thetaStep

			lines = append(lines, &region.HoughAccumulator{
				Rho:   rho,
				Theta: theta,
				Votes: votes,
			})
		}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Votes > lines[j].Votes
	})

	if len(lines) > 5 {
		lines = lines[:5]
	}

	return lines
}
