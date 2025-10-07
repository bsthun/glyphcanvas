package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionComputeGradientAngle(reg *region.Region, x, y uint16) float64 {
	gx := 0.0
	gy := 0.0

	sobelX := [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	sobelY := [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			px := int(x) + i
			py := int(y) + j
			if px >= 0 && px < int(reg.GetSizeX()) && py >= 0 && py < int(reg.GetSizeY()) {
				val := 0.0
				if reg.IsDrew(uint16(px), uint16(py)) {
					val = 1.0
				}
				gx += val * sobelX[j+1][i+1]
				gy += val * sobelY[j+1][i+1]
			}
		}
	}

	return math.Atan2(gy, gx)
}
