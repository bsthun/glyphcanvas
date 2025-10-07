package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionSortEdgesForContour(edges []*region.EdgePoint) []*region.EdgePoint {
	if len(edges) == 0 {
		return edges
	}

	sorted := make([]*region.EdgePoint, 0, len(edges))
	visited := make(map[int]bool)

	current := edges[0]
	sorted = append(sorted, current)
	visited[0] = true

	for len(sorted) < len(edges) {
		minDist := math.MaxFloat64
		minIdx := -1

		for i, edge := range edges {
			if visited[i] {
				continue
			}

			dist := math.Sqrt(float64((edge.X-current.X)*(edge.X-current.X) +
				(edge.Y-current.Y)*(edge.Y-current.Y)))

			if dist < minDist {
				minDist = dist
				minIdx = i
			}
		}

		if minIdx == -1 {
			break
		}

		current = edges[minIdx]
		sorted = append(sorted, current)
		visited[minIdx] = true
	}

	return sorted
}
