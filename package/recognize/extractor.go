package recognize

import (
	"os"

	"github.com/bsthun/glyphcanvas/package/character"
	characterCalculate "github.com/bsthun/glyphcanvas/package/character/calculate"
	characterHelper "github.com/bsthun/glyphcanvas/package/character/helper"
	"github.com/bsthun/glyphcanvas/package/recognize/helper"
	"github.com/bsthun/glyphcanvas/package/region"
	regionCalculate "github.com/bsthun/glyphcanvas/package/region/calculate"
	regionHelper "github.com/bsthun/glyphcanvas/package/region/helper"
	"gopkg.in/yaml.v3"
)

func ExtractFeatures(char *character.Character) (*CharacterFeature, error) {
	features := &CharacterFeature{}

	err := characterHelper.CharacterDetectAnchors(char)
	if err != nil {
		return nil, err
	}

	err = characterHelper.CharacterComputeMedialAxis(char)
	if err != nil {
		return nil, err
	}

	err = characterHelper.CharacterComprehensiveAnalysis(char)
	if err != nil {
		// Ignore error as it may not be critical
	}

	features.GridSignature = helper.ComputeGridSignature(char, 8)
	features.DirectionHist = helper.ComputeDirectionHistogram(char)
	features.ZoningFeatures = helper.ComputeZoningFeatures(char)
	features.ChainCode = helper.ComputeChainCodeFromBitmap(char)
	features.HuMoments = helper.ComputeHuMomentsFromChar(char)

	if char.GetBoundingBoxHeight() > 0 {
		features.AspectRatio = float64(char.GetBoundingBoxWidth()) / float64(char.GetBoundingBoxHeight())
	} else {
		features.AspectRatio = 1.0
	}

	totalArea := float64(char.GetBoundingBoxWidth() * char.GetBoundingBoxHeight())
	if totalArea > 0 {
		features.Density = float64(char.GetPixelCount()) / totalArea
	}

	cx, cy := helper.ComputeCenterOfMass(char)
	features.CenterOfMass = [2]float64{cx, cy}

	endpoints, junctions := helper.CountEndpointsAndJunctions(char)
	features.EndPoints = endpoints
	features.Junctions = junctions

	regions, _ := characterCalculate.CharacterBreakdownToRegions(char)
	features.RegionCount = len(regions)

	features.RegionFeatures = extractRegionFeatures(char, regions)

	features.TopologyHash = helper.ComputeTopologyHash(features.EndPoints, features.Junctions, features.RegionCount, features.ChainCode, features.GridSignature)

	return features, nil
}

func extractRegionFeatures(char *character.Character, regions []*region.Region) []RegionFeatureSet {
	var featureSets []RegionFeatureSet

	for _, reg := range regions {
		if reg == nil || len(reg.Draws) == 0 {
			continue
		}

		features := RegionFeatureSet{}

		arc := regionCalculate.RegionArc(reg)
		if arc != nil {
			features.ArcType = getArcTypeString(arc.Type)
			moments := regionHelper.RegionComputeMoments(reg)
			huMoments := regionHelper.RegionComputeHuInvariants(moments)
			features.Circularity = regionHelper.RegionComputeCircularity(huMoments)
			features.Linearity = regionHelper.RegionComputeLinearity(huMoments)

			// Compute curve strength
			edges := regionHelper.RegionExtractEdge(reg)
			chainCode := regionHelper.RegionComputeChainCode(edges)
			curvatures := regionHelper.RegionComputeCurvatures(chainCode)
			features.CurveStrength = float64(regionHelper.RegionComputeCurveStrength(curvatures, edges))
		}

		if arc == nil {
			moments := regionHelper.RegionComputeMoments(reg)
			huArray := regionHelper.RegionComputeHuInvariants(moments)
			copy(features.HuMoments[:], huArray)
		}

		edges := regionHelper.RegionExtractEdge(reg)
		chainCode := regionHelper.RegionComputeChainCode(edges)
		features.ChainCodeHash = helper.HashChainCode(chainCode)

		if char.GetPixelCount() > 0 {
			features.RelativeSize = float64(len(reg.Draws)) / float64(char.GetPixelCount())
		}

		if len(reg.Draws) > 0 {
			var sumX, sumY uint32
			for _, point := range reg.Draws {
				sumX += uint32(point.X)
				sumY += uint32(point.Y)
			}
			avgX := float64(sumX) / float64(len(reg.Draws))
			avgY := float64(sumY) / float64(len(reg.Draws))

			if char.SizeX > 0 && char.SizeY > 0 {
				features.RelativePos[0] = avgX / float64(char.SizeX)
				features.RelativePos[1] = avgY / float64(char.SizeY)
			}
		}

		featureSets = append(featureSets, features)

		if len(featureSets) >= 10 {
			break
		}
	}

	return featureSets
}

func getArcTypeString(arcType region.ArcType) string {
	switch arcType {
	case region.ArcTypeCircle:
		return "circle"
	case region.ArcTypeStrengthLine:
		return "strength_line"
	case region.ArcTypeCurveLine:
		return "curve_line"
	case region.ArcTypeTriangle:
		return "triangle"
	case region.ArcTypeRectangle:
		return "rectangle"
	default:
		return "unknown"
	}
}

func SaveDatabase(database *FeatureDatabase, path string) error {
	data, err := yaml.Marshal(database)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadDatabase(path string) (*FeatureDatabase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var database FeatureDatabase
	err = yaml.Unmarshal(data, &database)
	if err != nil {
		return nil, err
	}

	return &database, nil
}
