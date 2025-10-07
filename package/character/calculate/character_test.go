package characterCalculate

import (
	"fmt"
	"testing"

	"github.com/bsthun/glyphcanvas/package/character"
	"github.com/bsthun/glyphcanvas/package/character/helper"
)

func TestCharacterBasicFunctionality(t *testing.T) {
	// Create a simple test character (letter "L" shape)
	char := character.NewCharacter(10, 10, nil)

	// Draw vertical line
	for y := uint16(1); y <= 8; y++ {
		char.Draw(2, y)
	}

	// Draw horizontal line
	for x := uint16(2); x <= 6; x++ {
		char.Draw(x, 8)
	}

	if char.IsEmpty() {
		t.Error("Character should not be empty after drawing")
	}

	if char.GetPixelCount() == 0 {
		t.Error("Character should have pixels after drawing")
	}

	// Test bounding box
	if char.GetBoundingBoxWidth() == 0 || char.GetBoundingBoxHeight() == 0 {
		t.Error("Bounding box should be non-zero")
	}

	fmt.Printf("Character has %d pixels\n", char.GetPixelCount())
	fmt.Printf("Bounding box: %dx%d\n", char.GetBoundingBoxWidth(), char.GetBoundingBoxHeight())
}

func TestCharacterAnchorDetection(t *testing.T) {
	// Create a test character with clear anchor points
	char := createTestCharacterWithCorners()

	// Detect anchor points
	err := characterHelper.CharacterDetectAnchors(char)
	if err != nil {
		t.Errorf("Anchor detection failed: %v", err)
	}

	fmt.Printf("Detected %d anchor points\n", len(char.AnchorPoints))

	// Print anchor points for verification
	for i, anchor := range char.AnchorPoints {
		fmt.Printf("Anchor %d: (%d,%d) Type: %s Strength: %.2f\n",
			i, anchor.Point.X, anchor.Point.Y, anchor.Type, anchor.Strength)
	}

	if len(char.AnchorPoints) == 0 {
		t.Error("Should detect at least some anchor points in test character")
	}
}

func TestCharacterMedialAxis(t *testing.T) {
	// Create a test character
	char := createTestCharacterWithThickness()

	// Compute medial axis
	err := characterHelper.CharacterComputeMedialAxis(char)
	if err != nil {
		t.Errorf("Medial axis computation failed: %v", err)
	}

	fmt.Printf("Medial axis has %d points\n", len(char.MedialAxis))
	fmt.Printf("Skeleton has %d branches\n", len(char.SkeletonBranches))

	// Print skeleton branches
	for branchID, branch := range char.SkeletonBranches {
		fmt.Printf("Branch %s: %d points\n", branchID, len(branch))
	}

	if len(char.MedialAxis) == 0 {
		t.Error("Should compute some medial axis points for test character")
	}
}

func TestCharacterRegionBreakdown(t *testing.T) {
	// Create a test character with multiple parts
	char := createTestCharacterMultiRegion()

	// Break down into regions
	regions, err := CharacterBreakdownToRegions(char)
	if err != nil {
		t.Errorf("Region breakdown failed: %v", err)
	}

	fmt.Printf("Character broken down into %d regions\n", len(regions))

	// Print region information
	for i, region := range regions {
		fmt.Printf("Region %d: %d pixels, size %dx%d\n",
			i, len(region.Draws), region.GetSizeX(), region.GetSizeY())
	}

	if len(regions) == 0 {
		t.Error("Should produce at least one region")
	}
}

func TestCharacterComprehensiveAnalysis(t *testing.T) {
	// Create a complex test character
	char := createTestCharacterComplex()

	// Perform comprehensive analysis
	err := characterHelper.CharacterComprehensiveAnalysis(char)
	if err != nil {
		t.Errorf("Comprehensive analysis failed: %v", err)
	}

	// Get analysis summary
	summary := characterHelper.CharacterGetAnalysisSummary(char)

	fmt.Printf("=== Character Analysis Summary ===\n")
	fmt.Printf("Pixel count: %v\n", summary["pixelCount"])
	fmt.Printf("Anchor points: %v\n", summary["anchorPointCount"])
	fmt.Printf("Anchor types: %v\n", summary["anchorTypes"])
	fmt.Printf("Region count: %v\n", summary["regionCount"])

	if topology, ok := summary["topology"]; ok {
		fmt.Printf("Topology: %v\n", topology)
	}

	if classification, ok := summary["classification"]; ok {
		fmt.Printf("Classification: %v\n", classification)
	}

	if metrics, ok := summary["metrics"]; ok {
		fmt.Printf("Metrics: %v\n", metrics)
	}
}

func TestCharacterConfiguration(t *testing.T) {
	// Test custom configuration
	config := character.DefaultCharacterConfig()
	config.AnchorDetectionThreshold = 0.5
	config.MinRegionSize = 3

	err := config.Validate()
	if err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}

	char := character.NewCharacter(20, 20, config)
	if char.Config.AnchorDetectionThreshold != 0.5 {
		t.Error("Custom configuration not applied correctly")
	}

	fmt.Printf("Using custom config with anchor threshold: %.2f\n",
		char.Config.AnchorDetectionThreshold)
}

// Helper functions to create test characters

func createTestCharacterWithCorners() *character.Character {
	// Create a rectangular character with clear corners
	char := character.NewCharacter(15, 15, nil)

	// Draw rectangle outline
	for x := uint16(3); x <= 12; x++ {
		char.Draw(x, 3)  // Top
		char.Draw(x, 12) // Bottom
	}
	for y := uint16(3); y <= 12; y++ {
		char.Draw(3, y)  // Left
		char.Draw(12, y) // Right
	}

	return char
}

func createTestCharacterWithThickness() *character.Character {
	// Create a character with some thickness for medial axis computation
	char := character.NewCharacter(20, 20, nil)

	// Draw thick vertical line
	for y := uint16(2); y <= 18; y++ {
		for x := uint16(8); x <= 12; x++ {
			char.Draw(x, y)
		}
	}

	// Draw thick horizontal line
	for x := uint16(2); x <= 18; x++ {
		for y := uint16(8); y <= 12; y++ {
			char.Draw(x, y)
		}
	}

	return char
}

func createTestCharacterMultiRegion() *character.Character {
	// Create a character that should naturally break into multiple regions
	char := character.NewCharacter(25, 25, nil)

	// Draw first region (square)
	for x := uint16(2); x <= 8; x++ {
		for y := uint16(2); y <= 8; y++ {
			char.Draw(x, y)
		}
	}

	// Draw second region (circle-like)
	center := character.Point{X: 18, Y: 18}
	radius := 4.0
	for x := uint16(14); x <= 22; x++ {
		for y := uint16(14); y <= 22; y++ {
			dx := float64(int16(x) - int16(center.X))
			dy := float64(int16(y) - int16(center.Y))
			if dx*dx+dy*dy <= radius*radius {
				char.Draw(x, y)
			}
		}
	}

	// Draw connecting line
	for i := uint16(8); i <= 14; i++ {
		char.Draw(i, 10)
		char.Draw(10, i)
	}

	return char
}

func createTestCharacterComplex() *character.Character {
	// Create a complex character for comprehensive testing
	char := character.NewCharacter(30, 30, nil)

	// Main body (filled rectangle)
	for x := uint16(5); x <= 25; x++ {
		for y := uint16(10); y <= 20; y++ {
			char.Draw(x, y)
		}
	}

	// Add holes
	for x := uint16(8); x <= 12; x++ {
		for y := uint16(13); y <= 17; y++ {
			char.Erase(x, y)
		}
	}

	for x := uint16(18); x <= 22; x++ {
		for y := uint16(13); y <= 17; y++ {
			char.Erase(x, y)
		}
	}

	// Add protrusions
	for y := uint16(5); y <= 9; y++ {
		char.Draw(15, y)
		char.Draw(14, y)
		char.Draw(16, y)
	}

	for y := uint16(21); y <= 25; y++ {
		char.Draw(15, y)
		char.Draw(14, y)
		char.Draw(16, y)
	}

	// Add corners and details
	char.Draw(5, 8)
	char.Draw(25, 8)
	char.Draw(5, 22)
	char.Draw(25, 22)

	return char
}

func BenchmarkCharacterAnalysis(b *testing.B) {
	char := createTestCharacterComplex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear previous analysis
		char.ClearAnalysisResults()

		// Run analysis
		err := characterHelper.CharacterComprehensiveAnalysis(char)
		if err != nil {
			b.Errorf("Analysis failed: %v", err)
		}
	}
}
