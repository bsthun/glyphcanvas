package characterCalculate

import (
	"math"
	"sort"

	"github.com/bsthun/glyphcanvas/package/character"
	characterHelper "github.com/bsthun/glyphcanvas/package/character/helper"
	"github.com/bsthun/glyphcanvas/package/region"
)

func CharacterBreakdownToRegions(char *character.Character) ([]*region.Region, error) {
	if char.IsEmpty() {
		return []*region.Region{}, nil
	}

	// Step 1: Perform character analysis if not already done
	if len(char.AnchorPoints) == 0 {
		err := characterHelper.CharacterDetectAnchors(char)
		if err != nil {
			return nil, err
		}
	}

	if len(char.MedialAxis) == 0 {
		err := characterHelper.CharacterComputeMedialAxis(char)
		if err != nil {
			return nil, err
		}
	}

	// Step 2: Identify segmentation lines based on anchor points and medial axis
	segmentationLines := identifySegmentationLines(char)

	// Step 3: Segment the character using the identified lines
	regions := segmentCharacter(char, segmentationLines)

	// Step 4: Refine regions by merging small adjacent regions
	refinedRegions := refineRegions(char, regions)

	// Step 5: Analyze each region using existing region analysis tools
	analyzedRegions := analyzeRegions(refinedRegions)

	return analyzedRegions, nil
}

type SegmentationLine struct {
	StartPoint *character.Point
	EndPoint   *character.Point
	Type       string  // "anchor_based", "medial_based", "stroke_boundary"
	Strength   float64 // Importance of this segmentation line
}

func identifySegmentationLines(char *character.Character) []*SegmentationLine {
	var lines []*SegmentationLine

	// 1. Create segmentation lines based on anchor points
	lines = append(lines, createAnchorBasedLines(char)...)

	// 2. Create segmentation lines based on medial axis branching
	lines = append(lines, createMedialAxisBasedLines(char)...)

	// 3. Create segmentation lines based on stroke boundaries
	if char.Config.EnableStrokeAnalysis {
		lines = append(lines, createStrokeBoundaryLines(char)...)
	}

	// 4. Filter and prioritize segmentation lines
	return filterSegmentationLines(char, lines)
}

func createAnchorBasedLines(char *character.Character) []*SegmentationLine {
	var lines []*SegmentationLine

	// Group anchor points by type for different strategies
	junctionAnchors := char.GetAnchorPointsByType("junction")
	cornerAnchors := char.GetAnchorPointsByType("corner")
	extremumAnchors := getExtremumAnchors(char)

	// Strategy 1: Connect junction points to create primary segmentation
	for _, junction := range junctionAnchors {
		nearbyAnchors := findNearbyAnchors(char, junction, 20.0)
		for _, nearby := range nearbyAnchors {
			if nearby != junction {
				line := &SegmentationLine{
					StartPoint: junction.Point,
					EndPoint:   nearby.Point,
					Type:       "anchor_based",
					Strength:   (junction.Strength + nearby.Strength) / 2.0,
				}
				lines = append(lines, line)
			}
		}
	}

	// Strategy 2: Connect significant corner points
	for i, corner1 := range cornerAnchors {
		for j, corner2 := range cornerAnchors {
			if i >= j {
				continue
			}

			// Check if corners are on opposite sides or significant distance
			dist := computeDistance(corner1.Point, corner2.Point)
			if dist > char.Config.MinAnchorDistance*2 {
				line := &SegmentationLine{
					StartPoint: corner1.Point,
					EndPoint:   corner2.Point,
					Type:       "anchor_based",
					Strength:   (corner1.Strength + corner2.Strength) / 2.0 * 0.8, // Lower priority
				}
				lines = append(lines, line)
			}
		}
	}

	// Strategy 3: Create lines from extremum points to character center
	if len(extremumAnchors) > 0 {
		center := computeCharacterCenter(char)
		for _, extremum := range extremumAnchors {
			line := &SegmentationLine{
				StartPoint: extremum.Point,
				EndPoint:   center,
				Type:       "anchor_based",
				Strength:   extremum.Strength * 0.6, // Lower priority
			}
			lines = append(lines, line)
		}
	}

	return lines
}

func createMedialAxisBasedLines(char *character.Character) []*SegmentationLine {
	var lines []*SegmentationLine

	// Find branching points in the medial axis
	branchingPoints := findMedialAxisBranchingPoints(char)

	// Create segmentation lines from branching points to the boundary
	for _, branchPoint := range branchingPoints {
		boundaryPoints := findNearestBoundaryPoints(char, branchPoint)

		for _, boundaryPoint := range boundaryPoints {
			line := &SegmentationLine{
				StartPoint: branchPoint,
				EndPoint:   boundaryPoint,
				Type:       "medial_based",
				Strength:   0.7, // Medium priority
			}
			lines = append(lines, line)
		}
	}

	// Create lines between major skeleton branches
	branchConnections := findSkeletonBranchConnections(char)
	for _, connection := range branchConnections {
		lines = append(lines, connection)
	}

	return lines
}

func createStrokeBoundaryLines(char *character.Character) []*SegmentationLine {
	var lines []*SegmentationLine

	// Analyze stroke width variations to identify natural segmentation points
	strokeWidthMap := computeStrokeWidthMap(char)

	// Find points where stroke width changes significantly
	widthChangePoints := findStrokeWidthChangePoints(char, strokeWidthMap)

	// Create segmentation lines at width change points
	for _, changePoint := range widthChangePoints {
		// Find perpendicular line across the stroke at this point
		perpLine := computePerpendicularStrokeLine(char, changePoint, strokeWidthMap)
		if perpLine != nil {
			lines = append(lines, perpLine)
		}
	}

	return lines
}

func findNearbyAnchors(char *character.Character, anchor *character.AnchorPoint, maxDistance float64) []*character.AnchorPoint {
	var nearby []*character.AnchorPoint

	for _, other := range char.AnchorPoints {
		if other == anchor {
			continue
		}

		dist := computeDistance(anchor.Point, other.Point)
		if dist <= maxDistance {
			nearby = append(nearby, other)
		}
	}

	return nearby
}

func getExtremumAnchors(char *character.Character) []*character.AnchorPoint {
	var extremums []*character.AnchorPoint

	for _, anchor := range char.AnchorPoints {
		if anchor.Type == "extremum_left" || anchor.Type == "extremum_right" ||
			anchor.Type == "extremum_top" || anchor.Type == "extremum_bottom" {
			extremums = append(extremums, anchor)
		}
	}

	return extremums
}

func computeDistance(p1, p2 *character.Point) float64 {
	dx := float64(int16(p1.X) - int16(p2.X))
	dy := float64(int16(p1.Y) - int16(p2.Y))
	return math.Sqrt(dx*dx + dy*dy)
}

func computeCharacterCenter(char *character.Character) *character.Point {
	if len(char.BoundingBox) == 0 || len(char.Draws) == 0 {
		return &character.Point{X: char.SizeX / 2, Y: char.SizeY / 2}
	}

	// Compute centroid of character pixels
	var sumX, sumY uint32
	for _, point := range char.Draws {
		sumX += uint32(point.X)
		sumY += uint32(point.Y)
	}

	centerX := uint16(sumX / uint32(len(char.Draws)))
	centerY := uint16(sumY / uint32(len(char.Draws)))

	return &character.Point{X: centerX, Y: centerY}
}

func findMedialAxisBranchingPoints(char *character.Character) []*character.Point {
	var branchingPoints []*character.Point

	// Count connections for each medial axis point
	for _, point := range char.MedialAxis {
		connectionCount := 0

		// Check how many other medial axis points are connected to this one
		for _, other := range char.MedialAxis {
			if other == point {
				continue
			}

			dist := computeDistance(point, other)
			if dist <= math.Sqrt2+0.1 { // Adjacent points (including diagonal)
				connectionCount++
			}
		}

		// Points with 3 or more connections are branching points
		if connectionCount >= 3 {
			branchingPoints = append(branchingPoints, point)
		}
	}

	return branchingPoints
}

func findNearestBoundaryPoints(char *character.Character, point *character.Point) []*character.Point {
	var boundaryPoints []*character.Point

	// Cast rays in multiple directions to find boundary intersections
	directions := []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi, 5 * math.Pi / 4, 3 * math.Pi / 2, 7 * math.Pi / 4}

	for _, angle := range directions {
		boundaryPoint := castRayToBoundary(char, point, angle)
		if boundaryPoint != nil {
			boundaryPoints = append(boundaryPoints, boundaryPoint)
		}
	}

	return boundaryPoints
}

func castRayToBoundary(char *character.Character, start *character.Point, angle float64) *character.Point {
	dx := math.Cos(angle)
	dy := math.Sin(angle)

	x := float64(start.X)
	y := float64(start.Y)

	for step := 0; step < int(math.Max(float64(char.SizeX), float64(char.SizeY))); step++ {
		x += dx
		y += dy

		nx := uint16(math.Round(x))
		ny := uint16(math.Round(y))

		// Check bounds
		if nx >= char.SizeX || ny >= char.SizeY {
			break
		}

		// Check if we've hit the boundary (transition from foreground to background)
		if !char.IsDrew(nx, ny) {
			// Go back one step to find the last foreground pixel
			x -= dx
			y -= dy
			return &character.Point{X: uint16(math.Round(x)), Y: uint16(math.Round(y))}
		}
	}

	return nil
}

func findSkeletonBranchConnections(char *character.Character) []*SegmentationLine {
	var lines []*SegmentationLine

	// Find connections between different skeleton branches
	branchEndpoints := make(map[string][]*character.Point)

	for branchID, branch := range char.SkeletonBranches {
		if len(branch) > 0 {
			// Get endpoints of each branch
			endpoints := []*character.Point{branch[0], branch[len(branch)-1]}
			branchEndpoints[branchID] = endpoints
		}
	}

	// Connect endpoints of different branches if they're close
	branches := make([]string, 0, len(branchEndpoints))
	for branchID := range branchEndpoints {
		branches = append(branches, branchID)
	}

	for i, branch1 := range branches {
		for j, branch2 := range branches {
			if i >= j {
				continue
			}

			endpoints1 := branchEndpoints[branch1]
			endpoints2 := branchEndpoints[branch2]

			for _, ep1 := range endpoints1 {
				for _, ep2 := range endpoints2 {
					dist := computeDistance(ep1, ep2)
					if dist < char.Config.MinAnchorDistance*1.5 {
						line := &SegmentationLine{
							StartPoint: ep1,
							EndPoint:   ep2,
							Type:       "medial_based",
							Strength:   0.8,
						}
						lines = append(lines, line)
					}
				}
			}
		}
	}

	return lines
}

func computeStrokeWidthMap(char *character.Character) map[string]float64 {
	strokeWidths := make(map[string]float64)

	// For each medial axis point, compute the stroke width
	for _, point := range char.MedialAxis {
		width := computeLocalStrokeWidth(char, point)
		key := getPointKey(point)
		strokeWidths[key] = width
	}

	return strokeWidths
}

func computeLocalStrokeWidth(char *character.Character, point *character.Point) float64 {
	// Cast rays in perpendicular directions to find stroke boundaries
	maxWidth := 0.0

	directions := []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4}

	for _, angle := range directions {
		width1 := castRayToBackground(char, point, angle)
		width2 := castRayToBackground(char, point, angle+math.Pi)
		totalWidth := width1 + width2

		if totalWidth > maxWidth {
			maxWidth = totalWidth
		}
	}

	return maxWidth
}

func castRayToBackground(char *character.Character, start *character.Point, angle float64) float64 {
	dx := math.Cos(angle)
	dy := math.Sin(angle)

	x := float64(start.X)
	y := float64(start.Y)
	distance := 0.0

	for step := 0; step < 50; step++ { // Limit search distance
		x += dx
		y += dy
		distance += 1.0

		nx := uint16(math.Round(x))
		ny := uint16(math.Round(y))

		// Check bounds or background
		if nx >= char.SizeX || ny >= char.SizeY || !char.IsDrew(nx, ny) {
			return distance
		}
	}

	return distance
}

func findStrokeWidthChangePoints(char *character.Character, strokeWidths map[string]float64) []*character.Point {
	var changePoints []*character.Point
	threshold := 2.0 // Significant change threshold

	for _, point := range char.MedialAxis {
		key := getPointKey(point)
		currentWidth := strokeWidths[key]

		// Check neighboring medial axis points for width changes
		neighbors := findMedialAxisNeighbors(char, point)
		for _, neighbor := range neighbors {
			neighborKey := getPointKey(neighbor)
			neighborWidth := strokeWidths[neighborKey]

			if math.Abs(currentWidth-neighborWidth) > threshold {
				changePoints = append(changePoints, point)
				break
			}
		}
	}

	return changePoints
}

func computePerpendicularStrokeLine(char *character.Character, point *character.Point, strokeWidths map[string]float64) *SegmentationLine {
	// Find the direction of the stroke at this point
	strokeDirection := computeLocalStrokeDirection(char, point)

	// Compute perpendicular direction
	perpAngle := strokeDirection + math.Pi/2

	// Find boundary points in perpendicular direction
	boundary1 := castRayToBoundary(char, point, perpAngle)
	boundary2 := castRayToBoundary(char, point, perpAngle+math.Pi)

	if boundary1 != nil && boundary2 != nil {
		return &SegmentationLine{
			StartPoint: boundary1,
			EndPoint:   boundary2,
			Type:       "stroke_boundary",
			Strength:   0.6,
		}
	}

	return nil
}

func computeLocalStrokeDirection(char *character.Character, point *character.Point) float64 {
	// Find nearby medial axis points to estimate direction
	neighbors := findMedialAxisNeighbors(char, point)

	if len(neighbors) == 0 {
		return 0 // Default direction
	}

	// Compute average direction vector
	var sumDx, sumDy float64
	for _, neighbor := range neighbors {
		dx := float64(int16(neighbor.X) - int16(point.X))
		dy := float64(int16(neighbor.Y) - int16(point.Y))
		sumDx += dx
		sumDy += dy
	}

	return math.Atan2(sumDy, sumDx)
}

func findMedialAxisNeighbors(char *character.Character, point *character.Point) []*character.Point {
	var neighbors []*character.Point

	for _, other := range char.MedialAxis {
		if other == point {
			continue
		}

		dist := computeDistance(point, other)
		if dist <= math.Sqrt2+0.1 { // Adjacent points
			neighbors = append(neighbors, other)
		}
	}

	return neighbors
}

func filterSegmentationLines(char *character.Character, lines []*SegmentationLine) []*SegmentationLine {
	// Sort by strength (descending)
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Strength > lines[j].Strength
	})

	// Remove overlapping or redundant lines
	var filtered []*SegmentationLine

	for _, line := range lines {
		shouldAdd := true

		for _, existing := range filtered {
			if linesOverlap(line, existing, 3.0) {
				shouldAdd = false
				break
			}
		}

		if shouldAdd {
			filtered = append(filtered, line)
		}

		// Limit the number of segmentation lines
		if len(filtered) >= char.Config.MaxRegions-1 {
			break
		}
	}

	return filtered
}

func linesOverlap(line1, line2 *SegmentationLine, threshold float64) bool {
	// Check if lines are too similar or overlapping
	dist1 := computeDistance(line1.StartPoint, line2.StartPoint) + computeDistance(line1.EndPoint, line2.EndPoint)
	dist2 := computeDistance(line1.StartPoint, line2.EndPoint) + computeDistance(line1.EndPoint, line2.StartPoint)

	return math.Min(dist1, dist2) < threshold
}

func segmentCharacter(char *character.Character, segmentationLines []*SegmentationLine) []*region.Region {
	// Create initial region containing all character pixels
	regions := []*region.Region{createRegionFromCharacter(char)}

	// Apply each segmentation line to split regions
	for _, line := range segmentationLines {
		regions = applySemgentationLine(char, regions, line)
	}

	return regions
}

func createRegionFromCharacter(char *character.Character) *region.Region {
	reg := region.NewRegion(char.SizeX, char.SizeY)

	for _, point := range char.Draws {
		reg.Draw(point.X, point.Y)
	}

	return reg
}

func applySemgentationLine(char *character.Character, regions []*region.Region, line *SegmentationLine) []*region.Region {
	var newRegions []*region.Region

	for _, reg := range regions {
		splitRegions := splitRegionByLine(reg, line)
		newRegions = append(newRegions, splitRegions...)
	}

	return newRegions
}

func splitRegionByLine(reg *region.Region, line *SegmentationLine) []*region.Region {
	// Simple implementation: split based on which side of the line pixels are on
	region1 := region.NewRegion(reg.GetSizeX(), reg.GetSizeY())
	region2 := region.NewRegion(reg.GetSizeX(), reg.GetSizeY())

	for _, point := range reg.Draws {
		side := getPointSideOfLine(point, line)
		if side >= 0 {
			region1.Draw(point.X, point.Y)
		} else {
			region2.Draw(point.X, point.Y)
		}
	}

	// Return non-empty regions
	var result []*region.Region
	if len(region1.Draws) > 0 {
		result = append(result, region1)
	}
	if len(region2.Draws) > 0 {
		result = append(result, region2)
	}

	// If no split occurred, return original region
	if len(result) == 0 {
		return []*region.Region{reg}
	}

	return result
}

func getPointSideOfLine(point *region.Point, line *SegmentationLine) float64 {
	// Use cross product to determine which side of the line the point is on
	x1, y1 := float64(line.StartPoint.X), float64(line.StartPoint.Y)
	x2, y2 := float64(line.EndPoint.X), float64(line.EndPoint.Y)
	x, y := float64(point.X), float64(point.Y)

	return (x2-x1)*(y-y1) - (y2-y1)*(x-x1)
}

func refineRegions(char *character.Character, regions []*region.Region) []*region.Region {
	// Merge small adjacent regions
	minSize := char.Config.MinRegionSize

	var refined []*region.Region
	for _, reg := range regions {
		if uint16(len(reg.Draws)) >= minSize {
			refined = append(refined, reg)
		} else {
			// Try to merge with adjacent larger region
			merged := false
			for _, other := range refined {
				if regionsAreAdjacent(reg, other) {
					mergeRegions(other, reg)
					merged = true
					break
				}
			}
			if !merged {
				// Keep small region if it can't be merged
				refined = append(refined, reg)
			}
		}
	}

	return refined
}

func regionsAreAdjacent(reg1, reg2 *region.Region) bool {
	// Check if regions share any adjacent pixels
	for _, point1 := range reg1.Draws {
		for dx := int16(-1); dx <= 1; dx++ {
			for dy := int16(-1); dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx := uint16(int16(point1.X) + dx)
				ny := uint16(int16(point1.Y) + dy)

				if reg2.IsDrew(nx, ny) {
					return true
				}
			}
		}
	}
	return false
}

func mergeRegions(target, source *region.Region) {
	for _, point := range source.Draws {
		target.Draw(point.X, point.Y)
	}
}

func analyzeRegions(regions []*region.Region) []*region.Region {
	// Apply existing region analysis to each region
	// This would use the existing region helper functions
	// For now, just return the regions as-is since the existing
	// analysis functions can be applied separately
	return regions
}

func getPointKey(point *character.Point) string {
	return string(rune(point.X)) + "," + string(rune(point.Y))
}
