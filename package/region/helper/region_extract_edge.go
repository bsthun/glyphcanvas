package regionHelper

import "github.com/bsthun/glyphcanvas/package/region"

func RegionExtractEdge(r *region.Region) []*region.EdgePoint {
	var edges []*region.EdgePoint
	dx := []int{-1, 0, 1, -1, 1, -1, 0, 1}
	dy := []int{-1, -1, -1, 0, 0, 1, 1, 1}

	for x := uint16(1); x < r.GetSizeX()-1; x++ {
		for y := uint16(1); y < r.GetSizeY()-1; y++ {
			if !r.IsDrew(x, y) {
				continue
			}

			isEdge := false
			for i := 0; i < 8; i++ {
				nx := int(x) + dx[i]
				ny := int(y) + dy[i]
				if nx >= 0 && nx < int(r.GetSizeX()) && ny >= 0 && ny < int(r.GetSizeY()) {
					if !r.IsDrew(uint16(nx), uint16(ny)) {
						isEdge = true
						break
					}
				}
			}

			if isEdge {
				angle := RegionComputeGradientAngle(r, x, y)
				edges = append(edges, &region.EdgePoint{
					X:     int(x),
					Y:     int(y),
					Angle: angle,
				})
			}
		}
	}

	return edges
}
