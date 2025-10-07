package regionHelper

import (
	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionDetermineFillType(reg *region.Region) region.ArcFillType {
	edgeCount := 0
	totalCount := 0

	for x := uint16(1); x < reg.GetSizeX()-1; x++ {
		for y := uint16(1); y < reg.GetSizeY()-1; y++ {
			if reg.IsDrew(x, y) {
				totalCount++

				hasEmpty := false
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx := int(x) + dx
						ny := int(y) + dy
						if nx >= 0 && nx < int(reg.GetSizeX()) && ny >= 0 && ny < int(reg.GetSizeY()) {
							if !reg.IsDrew(uint16(nx), uint16(ny)) {
								hasEmpty = true
								break
							}
						}
					}
					if hasEmpty {
						break
					}
				}

				if hasEmpty {
					edgeCount++
				}
			}
		}
	}

	if totalCount > 0 {
		ratio := float64(edgeCount) / float64(totalCount)
		if ratio > 0.3 {
			return region.ArcFillTypeStroke
		}
	}

	return region.ArcFillTypeFill
}
