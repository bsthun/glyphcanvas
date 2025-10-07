package regionHelper

import (
	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionComputeChainCode(edges []*region.EdgePoint) []int {
	if len(edges) < 2 {
		return []int{}
	}

	sortedEdges := RegionSortEdgesForContour(edges)
	chainCode := []int{}

	for i := 1; i < len(sortedEdges); i++ {
		dx := sortedEdges[i].X - sortedEdges[i-1].X
		dy := sortedEdges[i].Y - sortedEdges[i-1].Y

		code := 0
		if dx == 1 && dy == 0 {
			code = 0
		} else if dx == 1 && dy == -1 {
			code = 1
		} else if dx == 0 && dy == -1 {
			code = 2
		} else if dx == -1 && dy == -1 {
			code = 3
		} else if dx == -1 && dy == 0 {
			code = 4
		} else if dx == -1 && dy == 1 {
			code = 5
		} else if dx == 0 && dy == 1 {
			code = 6
		} else if dx == 1 && dy == 1 {
			code = 7
		}

		chainCode = append(chainCode, code)
	}

	return chainCode
}
