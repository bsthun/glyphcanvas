package characterHelper

import (
	"github.com/bsthun/glyphcanvas/package/character"
	"github.com/bsthun/glyphcanvas/package/region"
	regionHelper "github.com/bsthun/glyphcanvas/package/region/helper"
)

func CharacterComprehensiveAnalysis(char *character.Character) error {
	if char.IsEmpty() {
		return nil
	}

	// Step 1: Basic character analysis
	if err := performBasicCharacterAnalysis(char); err != nil {
		return err
	}

	// Step 2: Break down character into regions (basic implementation)
	regions := []*region.Region{createRegionFromCharacter(char)}
	char.Regions = regions

	// Step 3: Analyze each region using existing region analysis tools
	if err := analyzeCharacterRegions(char); err != nil {
		return err
	}

	// Step 4: Compute character-level metrics based on region analysis
	if err := computeCharacterLevelMetrics(char); err != nil {
		return err
	}

	// Step 5: Classify character geometric properties
	if err := classifyCharacterGeometry(char); err != nil {
		return err
	}

	return nil
}

func performBasicCharacterAnalysis(char *character.Character) error {
	// Detect anchor points
	if err := CharacterDetectAnchors(char); err != nil {
		return err
	}

	// Compute medial axis and topology
	if err := CharacterComputeMedialAxis(char); err != nil {
		return err
	}

	// Analyze topology
	if err := CharacterAnalyzeTopology(char); err != nil {
		return err
	}

	// Compute basic moments for the entire character
	char.Moments = computeCharacterMoments(char)

	return nil
}

func analyzeCharacterRegions(char *character.Character) error {
	for i, reg := range char.Regions {
		// Apply comprehensive region analysis to each region
		if err := analyzeIndividualRegion(reg, i, char); err != nil {
			return err
		}
	}

	return nil
}

func analyzeIndividualRegion(reg *region.Region, regionIndex int, char *character.Character) error {
	// Use existing region analysis tools with proper workflow

	// 1. Compute moments first (this is the foundation)
	moments := regionHelper.RegionComputeMoments(reg)
	storeRegionAnalysis(char, regionIndex, "moments", moments)

	// 2. Compute Hu invariants from moments
	huInvariants := regionHelper.RegionComputeHuInvariants(moments)
	storeRegionAnalysis(char, regionIndex, "huInvariants", huInvariants)

	// 3. Compute geometric properties using Hu invariants
	circularity := regionHelper.RegionComputeCircularity(huInvariants)
	storeRegionAnalysis(char, regionIndex, "circularity", circularity)

	linearity := regionHelper.RegionComputeLinearity(huInvariants)
	storeRegionAnalysis(char, regionIndex, "linearity", linearity)

	rectangularity := regionHelper.RegionComputeRectangularity(huInvariants)
	storeRegionAnalysis(char, regionIndex, "rectangularity", rectangularity)

	ellipseRatio := regionHelper.RegionComputeEllipseRatio(moments)
	storeRegionAnalysis(char, regionIndex, "ellipseRatio", ellipseRatio)

	// 4. Basic region properties
	storeRegionAnalysis(char, regionIndex, "pixelCount", len(reg.Draws))
	storeRegionAnalysis(char, regionIndex, "boundingArea", reg.GetSizeX()*reg.GetSizeY())

	return nil
}

func storeRegionAnalysis(char *character.Character, regionIndex int, analysisType string, result interface{}) {
	if char.Topology["regionAnalysis"] == nil {
		char.Topology["regionAnalysis"] = make(map[string]map[string]interface{})
	}

	regionAnalysis := char.Topology["regionAnalysis"].(map[string]map[string]interface{})
	regionKey := "region_" + string(rune(regionIndex))

	if regionAnalysis[regionKey] == nil {
		regionAnalysis[regionKey] = make(map[string]interface{})
	}

	regionAnalysis[regionKey][analysisType] = result
}

func computeCharacterLevelMetrics(char *character.Character) error {
	if len(char.Regions) == 0 {
		return nil
	}

	// Aggregate metrics from all regions
	aggregatedMetrics := make(map[string]interface{})

	// Count different shape types
	shapeTypeCounts := make(map[string]int)
	totalCircularity := 0.0
	totalLinearity := 0.0
	totalRectangularity := 0.0
	totalArea := 0.0

	regionAnalysis := char.Topology["regionAnalysis"]
	if regionAnalysis != nil {
		regions := regionAnalysis.(map[string]map[string]interface{})

		for _, analysis := range regions {
			// Count shape types
			if shapeClass, ok := analysis["shapeClass"].(string); ok {
				shapeTypeCounts[shapeClass]++
			}

			// Sum geometric metrics
			if circularity, ok := analysis["circularity"].(float64); ok {
				totalCircularity += circularity
			}
			if linearity, ok := analysis["linearity"].(float64); ok {
				totalLinearity += linearity
			}
			if rectangularity, ok := analysis["rectangularity"].(float64); ok {
				totalRectangularity += rectangularity
			}
			if moments, ok := analysis["moments"].(map[string]float64); ok {
				if m00, exists := moments["m00"]; exists {
					totalArea += m00
				}
			}
		}
	}

	regionCount := len(char.Regions)
	if regionCount > 0 {
		aggregatedMetrics["averageCircularity"] = totalCircularity / float64(regionCount)
		aggregatedMetrics["averageLinearity"] = totalLinearity / float64(regionCount)
		aggregatedMetrics["averageRectangularity"] = totalRectangularity / float64(regionCount)
		aggregatedMetrics["totalArea"] = totalArea
		aggregatedMetrics["regionCount"] = regionCount
		aggregatedMetrics["shapeDistribution"] = shapeTypeCounts
	}

	// Character complexity metrics
	aggregatedMetrics["anchorPointDensity"] = float64(len(char.AnchorPoints)) / totalArea
	aggregatedMetrics["medialAxisComplexity"] = computeMedialAxisComplexity(char)
	aggregatedMetrics["skeletonBranchCount"] = len(char.SkeletonBranches)

	char.Topology["characterMetrics"] = aggregatedMetrics

	return nil
}

func computeMedialAxisComplexity(char *character.Character) float64 {
	if len(char.MedialAxis) == 0 {
		return 0.0
	}

	// Complexity based on:
	// 1. Number of medial axis points
	// 2. Number of skeleton branches
	// 3. Total medial axis length
	// 4. Branching factor

	pointCount := float64(len(char.MedialAxis))
	branchCount := float64(len(char.SkeletonBranches))
	totalLength := 0.0

	if medialAxisLength, ok := char.Topology["medialAxisLength"].(float64); ok {
		totalLength = medialAxisLength
	}

	// Normalized complexity score
	boundingArea := float64(char.GetBoundingBoxWidth() * char.GetBoundingBoxHeight())
	if boundingArea == 0 {
		return 0.0
	}

	complexity := (pointCount + branchCount*2 + totalLength) / boundingArea
	return complexity
}

func classifyCharacterGeometry(char *character.Character) error {
	if char.Topology["characterMetrics"] == nil {
		return nil
	}

	metrics := char.Topology["characterMetrics"].(map[string]interface{})
	classification := make(map[string]interface{})

	// Classify based on aggregated metrics
	if avgCircularity, ok := metrics["averageCircularity"].(float64); ok {
		if avgCircularity > char.Config.CircularityThreshold {
			classification["circularCharacter"] = true
		} else {
			classification["circularCharacter"] = false
		}
	}

	if avgLinearity, ok := metrics["averageLinearity"].(float64); ok {
		if avgLinearity > char.Config.LinearityThreshold {
			classification["linearCharacter"] = true
		} else {
			classification["linearCharacter"] = false
		}
	}

	if avgRectangularity, ok := metrics["averageRectangularity"].(float64); ok {
		if avgRectangularity > char.Config.RectangularityThreshold {
			classification["rectangularCharacter"] = true
		} else {
			classification["rectangularCharacter"] = false
		}
	}

	// Classify character complexity
	if complexity, ok := metrics["medialAxisComplexity"].(float64); ok {
		if complexity < 0.1 {
			classification["complexityLevel"] = "simple"
		} else if complexity < 0.3 {
			classification["complexityLevel"] = "moderate"
		} else {
			classification["complexityLevel"] = "complex"
		}
	}

	// Classify based on topology
	if connectivity, ok := char.Topology["connectivity"].(map[string]interface{}); ok {
		if holes, holesTok := connectivity["holes"].(int); holesTok {
			if holes == 0 {
				classification["topologyType"] = "solid"
			} else if holes == 1 {
				classification["topologyType"] = "single_hole"
			} else {
				classification["topologyType"] = "multiple_holes"
			}
		}

		if components, compOk := connectivity["connectedComponents"].(int); compOk {
			if components > 1 {
				classification["topologyType"] = "disconnected"
			}
		}
	}

	// Classify based on anchor points
	anchorTypes := make(map[string]int)
	for _, anchor := range char.AnchorPoints {
		anchorTypes[anchor.Type]++
	}

	if anchorTypes["junction"] > 2 {
		classification["hasMultipleJunctions"] = true
	}
	if anchorTypes["corner"] > 4 {
		classification["hasManyCornerscharacter"] = true
	}

	char.Topology["characterClassification"] = classification

	return nil
}

func computeCharacterMoments(char *character.Character) map[string]float64 {
	moments := make(map[string]float64)

	// Convert character to region-like structure for moment computation
	m00, m10, m01, m11, m20, m02, m21, m12, m30, m03 := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0

	for _, point := range char.Draws {
		fx := float64(point.X)
		fy := float64(point.Y)

		m00 += 1
		m10 += fx
		m01 += fy
		m11 += fx * fy
		m20 += fx * fx
		m02 += fy * fy
		m21 += fx * fx * fy
		m12 += fx * fy * fy
		m30 += fx * fx * fx
		m03 += fy * fy * fy
	}

	moments["m00"] = m00
	moments["m10"] = m10
	moments["m01"] = m01
	moments["m11"] = m11
	moments["m20"] = m20
	moments["m02"] = m02
	moments["m21"] = m21
	moments["m12"] = m12
	moments["m30"] = m30
	moments["m03"] = m03

	// Central moments
	if m00 > 0 {
		cx := m10 / m00
		cy := m01 / m00
		moments["cx"] = cx
		moments["cy"] = cy

		mu20 := m20 - cx*m10
		mu02 := m02 - cy*m01
		mu11 := m11 - cx*m01
		mu30 := m30 - 3*cx*m20 + 2*cx*cx*m10
		mu21 := m21 - 2*cx*m11 - cy*m20 + 2*cx*cx*m01
		mu12 := m12 - 2*cy*m11 - cx*m02 + 2*cy*cy*m10
		mu03 := m03 - 3*cy*m02 + 2*cy*cy*m01

		moments["mu20"] = mu20
		moments["mu02"] = mu02
		moments["mu11"] = mu11
		moments["mu30"] = mu30
		moments["mu21"] = mu21
		moments["mu12"] = mu12
		moments["mu03"] = mu03
	}

	return moments
}

func CharacterGetAnalysisSummary(char *character.Character) map[string]interface{} {
	summary := make(map[string]interface{})

	// Basic character info
	summary["characterSize"] = map[string]uint16{
		"width":  char.SizeX,
		"height": char.SizeY,
	}
	summary["pixelCount"] = len(char.Draws)
	summary["boundingBox"] = char.BoundingBox

	// Anchor points summary
	summary["anchorPointCount"] = len(char.AnchorPoints)
	anchorTypeCounts := make(map[string]int)
	for _, anchor := range char.AnchorPoints {
		anchorTypeCounts[anchor.Type]++
	}
	summary["anchorTypes"] = anchorTypeCounts

	// Region analysis summary
	summary["regionCount"] = len(char.Regions)

	// Topology summary
	if char.Topology["connectivity"] != nil {
		summary["topology"] = char.Topology["connectivity"]
	}

	// Classification summary
	if char.Topology["characterClassification"] != nil {
		summary["classification"] = char.Topology["characterClassification"]
	}

	// Metrics summary
	if char.Topology["characterMetrics"] != nil {
		summary["metrics"] = char.Topology["characterMetrics"]
	}

	return summary
}

func createRegionFromCharacter(char *character.Character) *region.Region {
	reg := region.NewRegion(char.SizeX, char.SizeY)

	for _, point := range char.Draws {
		reg.Draw(point.X, point.Y)
	}

	return reg
}
