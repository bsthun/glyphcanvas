package regionCalculate

import (
	"fmt"

	"github.com/bsthun/glyphcanvas/package/region"
	"github.com/bsthun/glyphcanvas/package/region/helper"
)

func RegionArc(r *region.Region) *region.Arc {
	if len(r.Draws) < 3 {
		return nil
	}

	edges := regionHelper.RegionExtractEdge(r)
	if len(edges) < 3 {
		return nil
	}

	chainCode := regionHelper.RegionComputeChainCode(edges)
	curvatures := regionHelper.RegionComputeCurvatures(chainCode)

	moments := regionHelper.RegionComputeMoments(r)
	huInvariants := regionHelper.RegionComputeHuInvariants(moments)

	lines := regionHelper.RegionDetectLinesHough(r, edges)
	circles := regionHelper.RegionDetectCirclesHough(r, edges)

	fillType := regionHelper.RegionDetermineFillType(r)
	arcType, fillType := regionHelper.RegionClassifyShape(fillType, len(r.Draws), huInvariants, curvatures, lines, circles)

	arc := &region.Arc{
		Type: arcType,
		Fill: fillType,
	}

	switch arcType {
	case region.ArcTypeCircle:
		arc.CircleEllipseRatio = regionHelper.RegionComputeEllipseRatio(moments)

	case region.ArcTypeStrengthLine:
		arc.LineDegree = regionHelper.RegionComputeLineDegree(lines)
		fmt.Printf("Line detected with degree: %.0f°\n", arc.LineDegree)

	case region.ArcTypeCurveLine:
		arc.ArcLineTheta = regionHelper.RegionComputeCurveStrength(curvatures, edges)
		fmt.Printf("Curve detected with strength: %.3f\n", arc.ArcLineTheta)

	case region.ArcTypeTriangle:
		corners := regionHelper.RegionDetectCorners(curvatures, edges)
		if len(corners) == 3 {
			fmt.Println("Triangle detected")
		}

	case region.ArcTypeRectangle:
		corners := regionHelper.RegionDetectCorners(curvatures, edges)
		if len(corners) == 4 {
			fmt.Println("Rectangle detected")
		}
	}

	regionHelper.RegionPrintDetectedAngles(edges)

	return arc
}
