package recognize

import (
	"math"
	"sort"

	"github.com/bsthun/glyphcanvas/package/recognize/helper"
)

func RecognizeCharacter(features *CharacterFeature, database *FeatureDatabase) []RecognitionCandidate {
	var candidates []RecognitionCandidate

	for unicode, dbFeatures := range database.Characters {
		distance := computeFeatureDistance(features, dbFeatures)
		candidates = append(candidates, RecognitionCandidate{
			Unicode:  unicode,
			Distance: distance,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	// Add confidence scores
	for i := range candidates {
		candidates[i].Confidence = (1.0 - candidates[i].Distance) * 100
		if candidates[i].Confidence < 0 {
			candidates[i].Confidence = 0
		}
	}

	return candidates
}

func computeFeatureDistance(f1, f2 *CharacterFeature) float64 {
	distance := 0.0
	weight := 0.0

	// Grid signature distance (Hamming distance normalized)
	if len(f1.GridSignature) == len(f2.GridSignature) {
		hamming := 0.0
		for i := 0; i < len(f1.GridSignature); i++ {
			if f1.GridSignature[i] != f2.GridSignature[i] {
				hamming++
			}
		}
		distance += (hamming / float64(len(f1.GridSignature))) * 0.15
		weight += 0.15
	}

	// Direction histogram distance (Euclidean)
	dirDistance := 0.0
	for i := 0; i < 8; i++ {
		diff := f1.DirectionHist[i] - f2.DirectionHist[i]
		dirDistance += diff * diff
	}
	distance += math.Sqrt(dirDistance) * 0.12
	weight += 0.12

	// Zoning features distance
	zoneDistance := 0.0
	for i := 0; i < 16; i++ {
		diff := f1.ZoningFeatures[i] - f2.ZoningFeatures[i]
		zoneDistance += diff * diff
	}
	distance += math.Sqrt(zoneDistance) * 0.10
	weight += 0.10

	// Hu moments distance
	huDistance := 0.0
	for i := 0; i < 7; i++ {
		if math.Abs(f1.HuMoments[i]) > 1e-15 && math.Abs(f2.HuMoments[i]) > 1e-15 {
			logDiff := math.Log10(math.Abs(f1.HuMoments[i])) - math.Log10(math.Abs(f2.HuMoments[i]))
			huDistance += logDiff * logDiff
		}
	}
	distance += math.Sqrt(huDistance) * 0.15
	weight += 0.15

	// Aspect ratio distance
	aspectDiff := math.Abs(f1.AspectRatio - f2.AspectRatio)
	distance += aspectDiff * 0.08
	weight += 0.08

	// Density distance
	densityDiff := math.Abs(f1.Density - f2.Density)
	distance += densityDiff * 0.08
	weight += 0.08

	// Center of mass distance
	comDistance := math.Sqrt(math.Pow(f1.CenterOfMass[0]-f2.CenterOfMass[0], 2) +
		math.Pow(f1.CenterOfMass[1]-f2.CenterOfMass[1], 2))
	distance += comDistance * 0.05
	weight += 0.05

	// Topology distance (endpoints, junctions, regions)
	topologyDistance := 0.0
	if f1.EndPoints+f2.EndPoints > 0 {
		topologyDistance += math.Abs(float64(f1.EndPoints-f2.EndPoints)) / float64(f1.EndPoints+f2.EndPoints+1)
	}
	if f1.Junctions+f2.Junctions > 0 {
		topologyDistance += math.Abs(float64(f1.Junctions-f2.Junctions)) / float64(f1.Junctions+f2.Junctions+1)
	}
	if f1.RegionCount+f2.RegionCount > 0 {
		topologyDistance += math.Abs(float64(f1.RegionCount-f2.RegionCount)) / float64(f1.RegionCount+f2.RegionCount+1)
	}
	distance += topologyDistance * 0.12
	weight += 0.12

	// Region features distance
	regionDistance := computeRegionFeaturesDistance(f1.RegionFeatures, f2.RegionFeatures)
	distance += regionDistance * 0.10
	weight += 0.10

	// Chain code similarity (Levenshtein distance normalized)
	if len(f1.ChainCode) > 0 && len(f2.ChainCode) > 0 {
		chainDistance := float64(helper.LevenshteinDistance(f1.ChainCode, f2.ChainCode)) /
			float64(math.Max(float64(len(f1.ChainCode)), float64(len(f2.ChainCode))))
		distance += chainDistance * 0.05
		weight += 0.05
	}

	if weight > 0 {
		return distance / weight
	}
	return 1.0
}

func computeRegionFeaturesDistance(r1, r2 []RegionFeatureSet) float64 {
	if len(r1) == 0 && len(r2) == 0 {
		return 0.0
	}
	if len(r1) == 0 || len(r2) == 0 {
		return 1.0
	}

	// Use Hungarian algorithm approximation: match each region in r1 to closest in r2
	totalDistance := 0.0
	count := math.Min(float64(len(r1)), float64(len(r2)))

	for i := 0; i < int(count); i++ {
		minDist := math.Inf(1)
		for j := 0; j < len(r2); j++ {
			dist := computeSingleRegionDistance(r1[i], r2[j])
			if dist < minDist {
				minDist = dist
			}
		}
		totalDistance += minDist
	}

	// Penalty for different region counts
	countPenalty := math.Abs(float64(len(r1)-len(r2))) / float64(len(r1)+len(r2))

	return (totalDistance/count + countPenalty) / 2.0
}

func computeSingleRegionDistance(r1, r2 RegionFeatureSet) float64 {
	distance := 0.0

	// Arc type (categorical)
	if r1.ArcType != r2.ArcType {
		distance += 0.3
	}

	// Circularity
	distance += math.Abs(r1.Circularity-r2.Circularity) * 0.2

	// Linearity
	distance += math.Abs(r1.Linearity-r2.Linearity) * 0.2

	// Curve strength
	distance += math.Abs(r1.CurveStrength-r2.CurveStrength) * 0.1

	// Hu moments
	huDist := 0.0
	for i := 0; i < 7; i++ {
		diff := r1.HuMoments[i] - r2.HuMoments[i]
		huDist += diff * diff
	}
	distance += math.Sqrt(huDist) * 0.1

	// Relative size
	distance += math.Abs(r1.RelativeSize-r2.RelativeSize) * 0.05

	// Relative position
	posDistance := math.Sqrt(math.Pow(r1.RelativePos[0]-r2.RelativePos[0], 2) +
		math.Pow(r1.RelativePos[1]-r2.RelativePos[1], 2))
	distance += posDistance * 0.05

	return distance
}
