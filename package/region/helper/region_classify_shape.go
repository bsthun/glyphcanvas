package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionClassifyShape(fillType region.ArcFillType, drawsCount int, hu []float64, curvatures []float64, lines, circles []*region.HoughAccumulator) (region.ArcType, region.ArcFillType) {
	if len(circles) > 0 && circles[0].Votes > drawsCount/3 {
		circularity := RegionComputeCircularity(hu)
		if circularity > 0.7 {
			return region.ArcTypeCircle, fillType
		}
	}

	if len(lines) > 0 && lines[0].Votes > drawsCount/2 {
		linearity := RegionComputeLinearity(hu)
		if linearity > 0.8 {
			return region.ArcTypeStrengthLine, fillType
		}
	}

	corners := RegionDetectCorners(curvatures, nil)
	if len(corners) == 3 {
		return region.ArcTypeTriangle, fillType
	} else if len(corners) == 4 {
		rectangularity := RegionComputeRectangularity(hu)
		if rectangularity > 0.7 {
			return region.ArcTypeRectangle, fillType
		}
	}

	avgCurvature := 0.0
	for _, c := range curvatures {
		avgCurvature += math.Abs(c)
	}
	if len(curvatures) > 0 {
		avgCurvature /= float64(len(curvatures))
	}

	if avgCurvature > 0.1 && avgCurvature < 0.8 {
		return region.ArcTypeCurveLine, fillType
	}

	return region.ArcTypeStrengthLine, fillType
}
