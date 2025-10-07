package regionHelper

import (
	"math"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionComputeLineDegree(lines []*region.HoughAccumulator) float32 {
	if len(lines) == 0 {
		return 0
	}

	theta := lines[0].Theta
	degree := theta * 180.0 / math.Pi

	// Normalize angle to be in terms of line direction (0-180 degrees)
	// Hough theta represents perpendicular to line, so adjust by 90 degrees
	lineDegree := degree - 90
	if lineDegree < 0 {
		lineDegree += 180
	}
	if lineDegree >= 180 {
		lineDegree -= 180
	}

	targetAngles := []float64{0, 45, 90, 135}
	minDiff := math.MaxFloat64
	bestAngle := 0.0

	for _, target := range targetAngles {
		diff := math.Abs(lineDegree - target)
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
		// Check wraparound at 180
		diff = math.Abs(lineDegree - (target + 180))
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
		diff = math.Abs(lineDegree - (target - 180))
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
	}

	// For display purposes, we can use 180 instead of 0 for vertical lines
	if bestAngle == 0 && math.Abs(lineDegree-180) < math.Abs(lineDegree) {
		bestAngle = 180
	}

	return float32(bestAngle)
}
