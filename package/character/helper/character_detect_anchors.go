package characterHelper

import (
	"github.com/bsthun/glyphcanvas/package/character"
	"math"
	"sort"
)

func CharacterDetectAnchors(char *character.Character) error {
	if char.IsEmpty() {
		return nil
	}

	char.ClearAnalysisResults()

	// Step 1: Detect contour points (edge pixels)
	contourPoints := extractContourPoints(char)
	if len(contourPoints) < 3 {
		return nil // Not enough points for analysis
	}

	// Step 2: Compute curvature for each contour point
	curvatures := computeCurvatures(contourPoints, char.Config.MedialAxisEpsilon)

	// Step 3: Detect anchor points based on curvature and topology
	detectCurvatureAnchors(char, contourPoints, curvatures)

	// Step 4: Detect junction points
	if char.Config.EnableJunctionDetection {
		detectJunctionAnchors(char)
	}

	// Step 5: Detect extremum points (topmost, bottommost, leftmost, rightmost)
	detectExtremumAnchors(char)

	// Step 6: Filter and refine anchor points
	filterAnchors(char)

	return nil
}

func extractContourPoints(char *character.Character) []*character.Point {
	var contour []*character.Point

	// Find edge pixels using 8-connectivity
	for _, point := range char.Draws {
		x, y := point.X, point.Y
		isEdge := false

		// Check 8 neighbors
		for dx := int16(-1); dx <= 1; dx++ {
			for dy := int16(-1); dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx := uint16(int16(x) + dx)
				ny := uint16(int16(y) + dy)

				// If neighbor is outside bounds or not drawn, this is an edge pixel
				if nx >= char.SizeX || ny >= char.SizeY || !char.IsDrew(nx, ny) {
					isEdge = true
					break
				}
			}
			if isEdge {
				break
			}
		}

		if isEdge {
			contour = append(contour, &character.Point{X: x, Y: y})
		}
	}

	return contour
}

func computeCurvatures(contour []*character.Point, epsilon float64) []float64 {
	n := len(contour)
	curvatures := make([]float64, n)

	for i := 0; i < n; i++ {
		// Use a local window to compute curvature
		windowSize := int(math.Max(3, 1.0/epsilon))
		if windowSize > n/3 {
			windowSize = n / 3
		}

		prev := (i - windowSize + n) % n
		next := (i + windowSize) % n

		// Calculate vectors
		p1 := contour[prev]
		p2 := contour[i]
		p3 := contour[next]

		// Compute curvature using the angle between vectors
		v1x := float64(int16(p2.X) - int16(p1.X))
		v1y := float64(int16(p2.Y) - int16(p1.Y))
		v2x := float64(int16(p3.X) - int16(p2.X))
		v2y := float64(int16(p3.Y) - int16(p2.Y))

		// Normalize vectors
		len1 := math.Sqrt(v1x*v1x + v1y*v1y)
		len2 := math.Sqrt(v2x*v2x + v2y*v2y)

		if len1 < epsilon || len2 < epsilon {
			curvatures[i] = 0
			continue
		}

		v1x /= len1
		v1y /= len1
		v2x /= len2
		v2y /= len2

		// Compute angle between vectors
		dotProduct := v1x*v2x + v1y*v2y
		crossProduct := v1x*v2y - v1y*v2x

		// Clamp dot product to avoid numerical errors
		if dotProduct > 1.0 {
			dotProduct = 1.0
		}
		if dotProduct < -1.0 {
			dotProduct = -1.0
		}

		angle := math.Acos(dotProduct)
		if crossProduct < 0 {
			angle = -angle
		}

		curvatures[i] = math.Abs(angle)
	}

	return curvatures
}

func detectCurvatureAnchors(char *character.Character, contour []*character.Point, curvatures []float64) {
	threshold := char.Config.CurvatureThreshold

	for i, point := range contour {
		curvature := curvatures[i]

		if curvature > threshold {
			// Check if this is a local maximum
			isLocalMax := true
			windowSize := 3

			for j := -windowSize; j <= windowSize; j++ {
				if j == 0 {
					continue
				}
				idx := (i + j + len(curvatures)) % len(curvatures)
				if curvatures[idx] >= curvature {
					isLocalMax = false
					break
				}
			}

			if isLocalMax {
				// Determine anchor type based on curvature strength
				anchorType := "corner"
				if curvature > threshold*2 {
					anchorType = "sharp_corner"
				}

				strength := math.Min(curvature/math.Pi, 1.0)
				angle := computeDirectionAngle(char, point)

				char.AddAnchorPoint(point.X, point.Y, anchorType, strength, curvature, angle)
			}
		}
	}
}

func detectJunctionAnchors(char *character.Character) {
	// Detect junction points where multiple strokes meet
	// This uses a simplified approach - count connected components in local neighborhoods

	for _, point := range char.Draws {
		x, y := point.X, point.Y

		// Check local neighborhood for junction patterns
		junctionStrength := analyzeJunctionPattern(char, x, y)

		if junctionStrength > char.Config.AnchorDetectionThreshold {
			angle := computeDirectionAngle(char, point)
			char.AddAnchorPoint(x, y, "junction", junctionStrength, 0, angle)
		}
	}
}

func analyzeJunctionPattern(char *character.Character, x, y uint16) float64 {
	// Count connected components in 3x3 neighborhood
	components := 0
	visited := make(map[string]bool)

	for dx := int16(-1); dx <= 1; dx++ {
		for dy := int16(-1); dy <= 1; dy++ {
			nx := uint16(int16(x) + dx)
			ny := uint16(int16(y) + dy)

			key := string(rune(nx)) + "," + string(rune(ny))
			if visited[key] || nx >= char.SizeX || ny >= char.SizeY || !char.IsDrew(nx, ny) {
				continue
			}

			// Start a new component
			components++
			floodFillNeighborhood(char, nx, ny, x, y, visited)
		}
	}

	// Junction if 3 or more components meet
	if components >= 3 {
		return float64(components-2) / 3.0 // Normalize to 0-1 range
	}

	return 0
}

func floodFillNeighborhood(char *character.Character, startX, startY, centerX, centerY uint16, visited map[string]bool) {
	stack := []character.Point{{X: startX, Y: startY}}

	for len(stack) > 0 {
		point := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		key := string(rune(point.X)) + "," + string(rune(point.Y))
		if visited[key] {
			continue
		}
		visited[key] = true

		// Only explore within 3x3 neighborhood of center
		if math.Abs(float64(int16(point.X)-int16(centerX))) > 1 ||
			math.Abs(float64(int16(point.Y)-int16(centerY))) > 1 {
			continue
		}

		// Add neighbors to stack
		for dx := int16(-1); dx <= 1; dx++ {
			for dy := int16(-1); dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx := uint16(int16(point.X) + dx)
				ny := uint16(int16(point.Y) + dy)

				if nx < char.SizeX && ny < char.SizeY && char.IsDrew(nx, ny) {
					nkey := string(rune(nx)) + "," + string(rune(ny))
					if !visited[nkey] {
						stack = append(stack, character.Point{X: nx, Y: ny})
					}
				}
			}
		}
	}
}

func detectExtremumAnchors(char *character.Character) {
	if len(char.BoundingBox) == 0 {
		return
	}

	minX := char.BoundingBox["minX"]
	maxX := char.BoundingBox["maxX"]
	minY := char.BoundingBox["minY"]
	maxY := char.BoundingBox["maxY"]

	// Find extreme points
	for _, point := range char.Draws {
		x, y := point.X, point.Y

		if x == minX {
			angle := computeDirectionAngle(char, point)
			char.AddAnchorPoint(x, y, "extremum_left", 0.8, 0, angle)
		}
		if x == maxX {
			angle := computeDirectionAngle(char, point)
			char.AddAnchorPoint(x, y, "extremum_right", 0.8, 0, angle)
		}
		if y == minY {
			angle := computeDirectionAngle(char, point)
			char.AddAnchorPoint(x, y, "extremum_top", 0.8, 0, angle)
		}
		if y == maxY {
			angle := computeDirectionAngle(char, point)
			char.AddAnchorPoint(x, y, "extremum_bottom", 0.8, 0, angle)
		}
	}
}

func computeDirectionAngle(char *character.Character, point *character.Point) float64 {
	x, y := point.X, point.Y

	// Compute gradient direction using Sobel operator
	var gx, gy float64

	for dx := int16(-1); dx <= 1; dx++ {
		for dy := int16(-1); dy <= 1; dy++ {
			nx := uint16(int16(x) + dx)
			ny := uint16(int16(y) + dy)

			var value float64
			if nx < char.SizeX && ny < char.SizeY && char.IsDrew(nx, ny) {
				value = 1.0
			} else {
				value = 0.0
			}

			// Sobel kernels
			sobelX := [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
			sobelY := [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}

			gx += value * sobelX[dx+1][dy+1]
			gy += value * sobelY[dx+1][dy+1]
		}
	}

	return math.Atan2(gy, gx)
}

func filterAnchors(char *character.Character) {
	if len(char.AnchorPoints) == 0 {
		return
	}

	// Sort by strength (descending)
	sort.Slice(char.AnchorPoints, func(i, j int) bool {
		return char.AnchorPoints[i].Strength > char.AnchorPoints[j].Strength
	})

	// Remove anchors that are too close to each other
	filtered := []*character.AnchorPoint{}
	minDist := char.Config.MinAnchorDistance

	for _, anchor := range char.AnchorPoints {
		shouldAdd := true

		for _, existing := range filtered {
			dx := float64(int16(anchor.Point.X) - int16(existing.Point.X))
			dy := float64(int16(anchor.Point.Y) - int16(existing.Point.Y))
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < minDist {
				shouldAdd = false
				break
			}
		}

		if shouldAdd {
			filtered = append(filtered, anchor)
		}
	}

	char.AnchorPoints = filtered
}
