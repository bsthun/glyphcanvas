package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func main() {
	datasetPath := "generate/dataset/singlecharacter"
	outputPath := "generate/extract/char.yml"

	files, err := filepath.Glob(filepath.Join(datasetPath, "*.png"))
	if err != nil {
		log.Fatal("Failed to read dataset:", err)
	}

	database := &FeatureDatabase{
		Characters: make(map[string]*CharacterFeature),
	}

	for _, file := range files {
		unicode := extractUnicodeFromFilename(file)
		if unicode == "" {
			continue
		}

		fmt.Printf("Processing %s (Unicode: %s)...\n", filepath.Base(file), unicode)

		char, err := loadCharacterFromImage(file)
		if err != nil {
			log.Printf("Failed to load %s: %v\n", file, err)
			continue
		}

		features, err := extractFeatures(char)
		if err != nil {
			log.Printf("Failed to extract features from %s: %v\n", file, err)
			continue
		}

		features.Unicode = unicode
		database.Characters[unicode] = features
	}

	data, err := yaml.Marshal(database)
	if err != nil {
		log.Fatal("Failed to marshal YAML:", err)
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		log.Fatal("Failed to write output file:", err)
	}

	fmt.Printf("Feature extraction complete. Saved to %s\n", outputPath)
}

func extractUnicodeFromFilename(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	if strings.HasPrefix(base, "char_th_") {
		hex := strings.TrimPrefix(base, "char_th_")
		if code, err := strconv.ParseInt(hex, 16, 32); err == nil {
			return fmt.Sprintf("%04X", code)
		}
	} else if strings.HasPrefix(base, "char_en_upper_") {
		char := strings.TrimPrefix(base, "char_en_upper_")
		if len(char) == 1 {
			return fmt.Sprintf("%04X", rune(char[0]))
		}
	} else if strings.HasPrefix(base, "char_en_lower_") {
		char := strings.TrimPrefix(base, "char_en_lower_")
		if len(char) == 1 {
			return fmt.Sprintf("%04X", rune(char[0]))
		}
	} else if strings.HasPrefix(base, "char_") {
		digit := strings.TrimPrefix(base, "char_")
		if len(digit) == 1 && digit[0] >= '0' && digit[0] <= '9' {
			return fmt.Sprintf("%04X", rune(digit[0]))
		}
	}

	return ""
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

	features.TopologyHash = computeTopologyHash(features)

	return features, nil
}

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

		edges := regionHelper.RegionExtractEdge(reg)
		chainCode := regionHelper.RegionComputeChainCode(edges)
		features.ChainCodeHash = hashChainCode(chainCode)

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

func hashChainCode(chainCode []int) string {
	if len(chainCode) == 0 {
		return ""
	}

	str := ""
	for _, code := range chainCode {
		str += fmt.Sprintf("%d", code)
		if len(str) > 20 {
			break
		}
	}

	hash := 0
	for i, c := range str {
		hash = hash*31 + int(c) + i
	}

	return fmt.Sprintf("%08x", hash)
}

func computeTopologyHash(features *CharacterFeature) string {
	data := fmt.Sprintf("e%d_j%d_r%d_%s_%s",
		features.EndPoints,
		features.Junctions,
		features.RegionCount,
		features.ChainCode,
		features.GridSignature[:16])

	hash := 0
	for _, c := range data {
		hash = hash*31 + int(c)
	}

	return fmt.Sprintf("%016x", hash)
}
