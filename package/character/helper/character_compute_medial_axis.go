package characterHelper

import (
	"github.com/bsthun/glyphcanvas/package/character"
	"math"
)

func CharacterComputeMedialAxis(char *character.Character) error {
	if char.IsEmpty() {
		return nil
	}

	// Clear previous medial axis data
	char.MedialAxis = []*character.Point{}
	char.SkeletonBranches = make(map[string][]*character.Point)

	// Step 1: Compute distance transform
	distanceField := computeDistanceTransform(char)

	// Step 2: Extract medial axis points using ridge detection
	medialPoints := extractMedialAxisPoints(char, distanceField)

	// Step 3: Order medial axis points into skeleton branches
	char.MedialAxis = medialPoints
	extractSkeletonBranches(char, distanceField)

	// Step 4: Prune short branches based on configuration
	pruneShortBranches(char)

	return nil
}

func computeDistanceTransform(char *character.Character) [][]float64 {
	sizeX := int(char.SizeX)
	sizeY := int(char.SizeY)

	// Initialize distance field
	distField := make([][]float64, sizeX)
	for x := 0; x < sizeX; x++ {
		distField[x] = make([]float64, sizeY)
		for y := 0; y < sizeY; y++ {
			if char.IsDrew(uint16(x), uint16(y)) {
				distField[x][y] = math.Inf(1) // Initialize to infinity for foreground
			} else {
				distField[x][y] = 0 // Background pixels have distance 0
			}
		}
	}

	// Forward pass
	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			if char.IsDrew(uint16(x), uint16(y)) {
				minDist := distField[x][y]

				// Check neighbors
				neighbors := [][]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}}
				for _, neighbor := range neighbors {
					nx, ny := x+neighbor[0], y+neighbor[1]
					if nx >= 0 && nx < sizeX && ny >= 0 && ny < sizeY {
						dist := distField[nx][ny]
						if neighbor[0] != 0 && neighbor[1] != 0 {
							dist += math.Sqrt2 // Diagonal distance
						} else {
							dist += 1.0 // Manhattan distance
						}
						if dist < minDist {
							minDist = dist
						}
					}
				}
				distField[x][y] = minDist
			}
		}
	}

	// Backward pass
	for x := sizeX - 1; x >= 0; x-- {
		for y := sizeY - 1; y >= 0; y-- {
			if char.IsDrew(uint16(x), uint16(y)) {
				minDist := distField[x][y]

				// Check neighbors
				neighbors := [][]int{{1, 1}, {1, 0}, {1, -1}, {0, 1}}
				for _, neighbor := range neighbors {
					nx, ny := x+neighbor[0], y+neighbor[1]
					if nx >= 0 && nx < sizeX && ny >= 0 && ny < sizeY {
						dist := distField[nx][ny]
						if neighbor[0] != 0 && neighbor[1] != 0 {
							dist += math.Sqrt2 // Diagonal distance
						} else {
							dist += 1.0 // Manhattan distance
						}
						if dist < minDist {
							minDist = dist
						}
					}
				}
				distField[x][y] = minDist
			}
		}
	}

	return distField
}

func extractMedialAxisPoints(char *character.Character, distField [][]float64) []*character.Point {
	var medialPoints []*character.Point
	threshold := char.Config.MedialAxisEpsilon

	sizeX := int(char.SizeX)
	sizeY := int(char.SizeY)

	for x := 1; x < sizeX-1; x++ {
		for y := 1; y < sizeY-1; y++ {
			if !char.IsDrew(uint16(x), uint16(y)) {
				continue
			}

			currentDist := distField[x][y]
			if currentDist < threshold {
				continue
			}

			// Check if this is a local maximum (ridge point)
			isLocalMax := true
			maxNeighborDist := 0.0

			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					neighborDist := distField[nx][ny]

					if neighborDist > maxNeighborDist {
						maxNeighborDist = neighborDist
					}

					if neighborDist > currentDist {
						isLocalMax = false
					}
				}
			}

			// Also check if this point is significant enough
			if isLocalMax && currentDist >= maxNeighborDist*0.9 {
				medialPoints = append(medialPoints, &character.Point{
					X: uint16(x),
					Y: uint16(y),
				})
			}
		}
	}

	return medialPoints
}

func extractSkeletonBranches(char *character.Character, distField [][]float64) {
	if len(char.MedialAxis) == 0 {
		return
	}

	// Create a graph of medial axis points
	visited := make(map[string]bool)
	branchID := 0

	for _, point := range char.MedialAxis {
		pointKey := getPointKey(point)
		if visited[pointKey] {
			continue
		}

		// Start a new branch from this point
		branch := traceBranch(char, point, distField, visited)
		if len(branch) > 1 {
			branchKey := "branch_" + string(rune(branchID))
			char.SkeletonBranches[branchKey] = branch
			branchID++
		}
	}
}

func traceBranch(char *character.Character, startPoint *character.Point, distField [][]float64, visited map[string]bool) []*character.Point {
	var branch []*character.Point
	stack := []*character.Point{startPoint}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		pointKey := getPointKey(current)
		if visited[pointKey] {
			continue
		}

		visited[pointKey] = true
		branch = append(branch, &character.Point{X: current.X, Y: current.Y})

		// Find connected medial axis points
		neighbors := findMedialAxisNeighbors(char, current)
		for _, neighbor := range neighbors {
			neighborKey := getPointKey(neighbor)
			if !visited[neighborKey] {
				stack = append(stack, neighbor)
			}
		}
	}

	return branch
}

func findMedialAxisNeighbors(char *character.Character, point *character.Point) []*character.Point {
	var neighbors []*character.Point
	x, y := int16(point.X), int16(point.Y)

	// Check 8-connected neighborhood
	for dx := int16(-1); dx <= 1; dx++ {
		for dy := int16(-1); dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := x+dx, y+dy
			if nx >= 0 && ny >= 0 && nx < int16(char.SizeX) && ny < int16(char.SizeY) {
				neighborPoint := &character.Point{X: uint16(nx), Y: uint16(ny)}

				// Check if this neighbor is in the medial axis
				for _, medialPoint := range char.MedialAxis {
					if medialPoint.X == neighborPoint.X && medialPoint.Y == neighborPoint.Y {
						neighbors = append(neighbors, neighborPoint)
						break
					}
				}
			}
		}
	}

	return neighbors
}

func pruneShortBranches(char *character.Character) {
	threshold := char.Config.SkeletonPruningThreshold
	filteredBranches := make(map[string][]*character.Point)

	for branchKey, branch := range char.SkeletonBranches {
		branchLength := computeBranchLength(branch)

		if branchLength >= threshold {
			filteredBranches[branchKey] = branch
		}
	}

	char.SkeletonBranches = filteredBranches

	// Update medial axis to only include points from retained branches
	var filteredMedialAxis []*character.Point
	for _, branch := range char.SkeletonBranches {
		filteredMedialAxis = append(filteredMedialAxis, branch...)
	}
	char.MedialAxis = filteredMedialAxis
}

func computeBranchLength(branch []*character.Point) float64 {
	if len(branch) < 2 {
		return 0
	}

	totalLength := 0.0
	for i := 1; i < len(branch); i++ {
		dx := float64(int16(branch[i].X) - int16(branch[i-1].X))
		dy := float64(int16(branch[i].Y) - int16(branch[i-1].Y))
		totalLength += math.Sqrt(dx*dx + dy*dy)
	}

	return totalLength
}

func getPointKey(point *character.Point) string {
	return string(rune(point.X)) + "," + string(rune(point.Y))
}

func CharacterAnalyzeTopology(char *character.Character) error {
	if char.IsEmpty() {
		return nil
	}

	// Analyze topological properties of the character
	char.Topology["branchCount"] = len(char.SkeletonBranches)
	char.Topology["medialAxisLength"] = computeTotalMedialAxisLength(char)
	char.Topology["anchorPointCount"] = len(char.AnchorPoints)

	// Count different types of anchor points
	anchorTypeCounts := make(map[string]int)
	for _, anchor := range char.AnchorPoints {
		anchorTypeCounts[anchor.Type]++
	}
	char.Topology["anchorTypes"] = anchorTypeCounts

	// Compute connectivity measures
	char.Topology["connectivity"] = analyzeConnectivity(char)

	return nil
}

func computeTotalMedialAxisLength(char *character.Character) float64 {
	totalLength := 0.0
	for _, branch := range char.SkeletonBranches {
		totalLength += computeBranchLength(branch)
	}
	return totalLength
}

func analyzeConnectivity(char *character.Character) map[string]interface{} {
	connectivity := make(map[string]interface{})

	// Euler characteristic: V - E + F = 2 - 2g (for genus g)
	// For binary images: χ = C - H where C = connected components, H = holes

	connectedComponents := countConnectedComponents(char)
	holes := countHoles(char)

	connectivity["connectedComponents"] = connectedComponents
	connectivity["holes"] = holes
	connectivity["eulerCharacteristic"] = connectedComponents - holes

	return connectivity
}

func countConnectedComponents(char *character.Character) int {
	visited := make(map[string]bool)
	components := 0

	for _, point := range char.Draws {
		pointKey := getPointKey(point)
		if visited[pointKey] {
			continue
		}

		// Start a new connected component
		components++
		floodFillComponent(char, point, visited)
	}

	return components
}

func floodFillComponent(char *character.Character, startPoint *character.Point, visited map[string]bool) {
	stack := []*character.Point{startPoint}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		pointKey := getPointKey(current)
		if visited[pointKey] {
			continue
		}

		visited[pointKey] = true

		// Add 8-connected neighbors
		x, y := int16(current.X), int16(current.Y)
		for dx := int16(-1); dx <= 1; dx++ {
			for dy := int16(-1); dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy
				if nx >= 0 && ny >= 0 && nx < int16(char.SizeX) && ny < int16(char.SizeY) {
					if char.IsDrew(uint16(nx), uint16(ny)) {
						neighborKey := getPointKey(&character.Point{X: uint16(nx), Y: uint16(ny)})
						if !visited[neighborKey] {
							stack = append(stack, &character.Point{X: uint16(nx), Y: uint16(ny)})
						}
					}
				}
			}
		}
	}
}

func countHoles(char *character.Character) int {
	// Count holes using background connected components that are surrounded by foreground
	visited := make(map[string]bool)
	holes := 0

	for x := uint16(0); x < char.SizeX; x++ {
		for y := uint16(0); y < char.SizeY; y++ {
			if char.IsDrew(x, y) {
				continue // Skip foreground pixels
			}

			pointKey := getPointKey(&character.Point{X: x, Y: y})
			if visited[pointKey] {
				continue
			}

			// Check if this background component is a hole
			component := extractBackgroundComponent(char, &character.Point{X: x, Y: y}, visited)
			if isHole(char, component) {
				holes++
			}
		}
	}

	return holes
}

func extractBackgroundComponent(char *character.Character, startPoint *character.Point, visited map[string]bool) []*character.Point {
	var component []*character.Point
	stack := []*character.Point{startPoint}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		pointKey := getPointKey(current)
		if visited[pointKey] {
			continue
		}

		visited[pointKey] = true
		component = append(component, &character.Point{X: current.X, Y: current.Y})

		// Add 4-connected background neighbors
		x, y := int16(current.X), int16(current.Y)
		neighbors := [][]int16{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

		for _, neighbor := range neighbors {
			nx, ny := x+neighbor[0], y+neighbor[1]
			if nx >= 0 && ny >= 0 && nx < int16(char.SizeX) && ny < int16(char.SizeY) {
				if !char.IsDrew(uint16(nx), uint16(ny)) {
					neighborKey := getPointKey(&character.Point{X: uint16(nx), Y: uint16(ny)})
					if !visited[neighborKey] {
						stack = append(stack, &character.Point{X: uint16(nx), Y: uint16(ny)})
					}
				}
			}
		}
	}

	return component
}

func isHole(char *character.Character, component []*character.Point) bool {
	// A background component is a hole if it doesn't touch the image boundary
	for _, point := range component {
		if point.X == 0 || point.X == char.SizeX-1 || point.Y == 0 || point.Y == char.SizeY-1 {
			return false // Touches boundary, not a hole
		}
	}
	return true
}
