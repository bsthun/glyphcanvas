package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"sort"

	"github.com/bsthun/glyphcanvas/package/character"
	characterCalculate "github.com/bsthun/glyphcanvas/package/character/calculate"
	characterHelper "github.com/bsthun/glyphcanvas/package/character/helper"
	"github.com/bsthun/glyphcanvas/package/region"
	regionCalculate "github.com/bsthun/glyphcanvas/package/region/calculate"
	regionHelper "github.com/bsthun/glyphcanvas/package/region/helper"
	"gopkg.in/yaml.v3"
)

type CharacterFeature struct {
	Unicode        string             `yaml:"unicode"`
	GridSignature  string             `yaml:"grid_signature"`
	DirectionHist  [8]float64         `yaml:"direction_histogram"`
	ZoningFeatures [16]float64        `yaml:"zoning_features"`
	ChainCode      string             `yaml:"chain_code"`
	HuMoments      [7]float64         `yaml:"hu_moments"`
	AspectRatio    float64            `yaml:"aspect_ratio"`
	Density        float64            `yaml:"density"`
	CenterOfMass   [2]float64         `yaml:"center_of_mass"`
	EndPoints      int                `yaml:"end_points"`
	Junctions      int                `yaml:"junctions"`
	RegionCount    int                `yaml:"region_count"`
	RegionFeatures []RegionFeatureSet `yaml:"region_features"`
	TopologyHash   string             `yaml:"topology_hash"`
}

type RegionFeatureSet struct {
	ArcType       string     `yaml:"arc_type"`
	Circularity   float64    `yaml:"circularity"`
	Linearity     float64    `yaml:"linearity"`
	CurveStrength float64    `yaml:"curve_strength"`
	HuMoments     [7]float64 `yaml:"hu_moments"`
	ChainCodeHash string     `yaml:"chain_code_hash"`
	RelativeSize  float64    `yaml:"relative_size"`
	RelativePos   [2]float64 `yaml:"relative_position"`
}

type FeatureDatabase struct {
	Characters map[string]*CharacterFeature `yaml:"characters"`
}

type RecognitionCandidate struct {
	Unicode    string
	Confidence float64
	Distance   float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <image_file>\n", os.Args[0])
		os.Exit(1)
	}

	imagePath := os.Args[1]
	databasePath := "generate/extract/char.yml"

	database, err := loadDatabase(databasePath)
	if err != nil {
		log.Fatal("Failed to load database:", err)
	}

	char, err := loadCharacterFromImage(imagePath)
	if err != nil {
		log.Fatal("Failed to load image:", err)
	}

	features, err := extractFeatures(char)
	if err != nil {
		log.Fatal("Failed to extract features:", err)
	}

	candidates := recognizeCharacter(features, database)

	fmt.Printf("Recognition results for %s:\n", imagePath)
	for i, candidate := range candidates {
		if i >= 5 {
			break
		}
		confidence := (1.0 - candidate.Distance) * 100
		if confidence < 0 {
			confidence = 0
		}
		fmt.Printf("%d. Unicode: %s, Confidence: %.2f%%, Distance: %.4f\n",
			i+1, candidate.Unicode, confidence, candidate.Distance)
	}

	if len(candidates) > 0 {
		fmt.Printf("\nBest match: %s (%.2f%% confidence)\n",
			candidates[0].Unicode, (1.0-candidates[0].Distance)*100)
	}
}

func loadDatabase(path string) (*FeatureDatabase, error) {
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

func loadCharacterFromImage(filename string) (*character.Character, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	char := character.NewCharacter(uint16(bounds.Dx()), uint16(bounds.Dy()), nil)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y < 128 {
				char.Draw(uint16(x-bounds.Min.X), uint16(y-bounds.Min.Y))
			}
		}
	}

	return char, nil
}

func extractFeatures(char *character.Character) (*CharacterFeature, error) {
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

	features.GridSignature = computeGridSignature(char, 8)
	features.DirectionHist = computeDirectionHistogram(char)
	features.ZoningFeatures = computeZoningFeatures(char)
	features.ChainCode = computeChainCodeFromBitmap(char)
	features.HuMoments = computeHuMomentsFromChar(char)

	if char.GetBoundingBoxHeight() > 0 {
		features.AspectRatio = float64(char.GetBoundingBoxWidth()) / float64(char.GetBoundingBoxHeight())
	} else {
		features.AspectRatio = 1.0
	}

	totalArea := float64(char.GetBoundingBoxWidth() * char.GetBoundingBoxHeight())
	if totalArea > 0 {
		features.Density = float64(char.GetPixelCount()) / totalArea
	}

	cx, cy := computeCenterOfMass(char)
	features.CenterOfMass = [2]float64{cx, cy}

	endpoints, junctions := countEndpointsAndJunctions(char)
	features.EndPoints = endpoints
	features.Junctions = junctions

	regions, _ := characterCalculate.CharacterBreakdownToRegions(char)
	features.RegionCount = len(regions)

	features.RegionFeatures = extractRegionFeatures(char, regions)

	return features, nil
}

func recognizeCharacter(features *CharacterFeature, database *FeatureDatabase) []RecognitionCandidate {
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
		chainDistance := float64(levenshteinDistance(f1.ChainCode, f2.ChainCode)) /
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

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// Helper functions from extract command
func computeGridSignature(char *character.Character, gridSize int) string {
	grid := make([][]bool, gridSize)
	for i := range grid {
		grid[i] = make([]bool, gridSize)
	}

	cellWidth := float64(char.SizeX) / float64(gridSize)
	cellHeight := float64(char.SizeY) / float64(gridSize)

	for x, col := range char.Bitmap {
		for y, val := range col {
			if val {
				gridX := int(float64(x) / cellWidth)
				gridY := int(float64(y) / cellHeight)

				if gridX >= gridSize {
					gridX = gridSize - 1
				}
				if gridY >= gridSize {
					gridY = gridSize - 1
				}

				grid[gridY][gridX] = true
			}
		}
	}

	signature := ""
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if grid[y][x] {
				signature += "1"
			} else {
				signature += "0"
			}
		}
	}

	return signature
}

func computeDirectionHistogram(char *character.Character) [8]float64 {
	var hist [8]float64
	directions := [][2]int{
		{1, 0}, {1, 1}, {0, 1}, {-1, 1},
		{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
	}

	for x, col := range char.Bitmap {
		for y, val := range col {
			if val {
				for i, dir := range directions {
					nx := int(x) + dir[0]
					ny := int(y) + dir[1]

					if nx >= 0 && ny >= 0 && uint16(nx) < char.SizeX && uint16(ny) < char.SizeY {
						if char.IsDrew(uint16(nx), uint16(ny)) {
							hist[i]++
						}
					}
				}
			}
		}
	}

	total := 0.0
	for _, v := range hist {
		total += v
	}
	if total > 0 {
		for i := range hist {
			hist[i] /= total
		}
	}

	return hist
}

func computeZoningFeatures(char *character.Character) [16]float64 {
	var features [16]float64
	zoneWidth := float64(char.SizeX) / 4.0
	zoneHeight := float64(char.SizeY) / 4.0

	for _, point := range char.Draws {
		zoneX := int(float64(point.X) / zoneWidth)
		zoneY := int(float64(point.Y) / zoneHeight)

		if zoneX >= 4 {
			zoneX = 3
		}
		if zoneY >= 4 {
			zoneY = 3
		}

		zoneIdx := zoneY*4 + zoneX
		features[zoneIdx]++
	}

	total := 0.0
	for _, v := range features {
		total += v
	}
	if total > 0 {
		for i := range features {
			features[i] /= total
		}
	}

	return features
}

func computeChainCodeFromBitmap(char *character.Character) string {
	if len(char.Draws) == 0 {
		return ""
	}

	visited := make(map[string]bool)
	startX, startY := char.Draws[0].X, char.Draws[0].Y
	currentX, currentY := startX, startY

	chainCode := ""
	directions := [][2]int{
		{1, 0}, {1, 1}, {0, 1}, {-1, 1},
		{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
	}

	maxSteps := 100
	for step := 0; step < maxSteps; step++ {
		key := fmt.Sprintf("%d,%d", currentX, currentY)
		if visited[key] {
			break
		}
		visited[key] = true

		found := false
		for i, dir := range directions {
			nx := int(currentX) + dir[0]
			ny := int(currentY) + dir[1]

			if nx >= 0 && ny >= 0 && uint16(nx) < char.SizeX && uint16(ny) < char.SizeY {
				nextKey := fmt.Sprintf("%d,%d", nx, ny)
				if !visited[nextKey] && char.IsDrew(uint16(nx), uint16(ny)) {
					chainCode += fmt.Sprintf("%d", i)
					currentX, currentY = uint16(nx), uint16(ny)
					found = true
					break
				}
			}
		}

		if !found {
			break
		}

		if len(chainCode) > 50 {
			break
		}
	}

	return chainCode
}

func computeHuMomentsFromChar(char *character.Character) [7]float64 {
	moments := make(map[string]float64)

	m00, m10, m01 := 0.0, 0.0, 0.0
	m20, m02, m11 := 0.0, 0.0, 0.0
	m30, m03, m21, m12 := 0.0, 0.0, 0.0, 0.0

	for _, point := range char.Draws {
		x := float64(point.X)
		y := float64(point.Y)

		m00++
		m10 += x
		m01 += y
		m11 += x * y
		m20 += x * x
		m02 += y * y
		m30 += x * x * x
		m03 += y * y * y
		m21 += x * x * y
		m12 += x * y * y
	}

	moments["m00"] = m00
	moments["m10"] = m10
	moments["m01"] = m01

	if m00 > 0 {
		xc := m10 / m00
		yc := m01 / m00

		moments["mu20"] = m20 - xc*m10
		moments["mu02"] = m02 - yc*m01
		moments["mu11"] = m11 - xc*m01
		moments["mu30"] = m30 - 3*xc*m20 + 2*xc*xc*m10
		moments["mu03"] = m03 - 3*yc*m02 + 2*yc*yc*m01
		moments["mu21"] = m21 - 2*xc*m11 - yc*m20 + 2*xc*xc*m01
		moments["mu12"] = m12 - 2*yc*m11 - xc*m02 + 2*yc*yc*m10
	}

	huArray := regionHelper.RegionComputeHuInvariants(moments)
	var result [7]float64
	copy(result[:], huArray)
	return result
}

func computeCenterOfMass(char *character.Character) (float64, float64) {
	if len(char.Draws) == 0 {
		return 0, 0
	}

	var sumX, sumY uint32
	for _, point := range char.Draws {
		sumX += uint32(point.X)
		sumY += uint32(point.Y)
	}

	cx := float64(sumX) / float64(len(char.Draws))
	cy := float64(sumY) / float64(len(char.Draws))

	if char.SizeX > 0 && char.SizeY > 0 {
		cx /= float64(char.SizeX)
		cy /= float64(char.SizeY)
	}

	return cx, cy
}

func countEndpointsAndJunctions(char *character.Character) (int, int) {
	endpoints := 0
	junctions := 0

	for x, col := range char.Bitmap {
		for y, val := range col {
			if val {
				neighbors := 0
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx := int(x) + dx
						ny := int(y) + dy
						if nx >= 0 && ny >= 0 && uint16(nx) < char.SizeX && uint16(ny) < char.SizeY {
							if char.IsDrew(uint16(nx), uint16(ny)) {
								neighbors++
							}
						}
					}
				}

				if neighbors == 1 {
					endpoints++
				} else if neighbors > 2 {
					junctions++
				}
			}
		}
	}

	return endpoints, junctions
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
