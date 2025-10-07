package helper

import (
	"fmt"

	"github.com/bsthun/glyphcanvas/package/character"
	regionHelper "github.com/bsthun/glyphcanvas/package/region/helper"
)

func ComputeGridSignature(char *character.Character, gridSize int) string {
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

func ComputeDirectionHistogram(char *character.Character) [8]float64 {
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

func ComputeZoningFeatures(char *character.Character) [16]float64 {
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

func ComputeChainCodeFromBitmap(char *character.Character) string {
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

func ComputeHuMomentsFromChar(char *character.Character) [7]float64 {
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

func ComputeCenterOfMass(char *character.Character) (float64, float64) {
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

func CountEndpointsAndJunctions(char *character.Character) (int, int) {
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

func HashChainCode(chainCode []int) string {
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

func ComputeTopologyHash(endpoints, junctions, regionCount int, chainCode, gridSignature string) string {
	data := fmt.Sprintf("e%d_j%d_r%d_%s_%s",
		endpoints,
		junctions,
		regionCount,
		chainCode,
		gridSignature[:min(16, len(gridSignature))])

	hash := 0
	for _, c := range data {
		hash = hash*31 + int(c)
	}

	return fmt.Sprintf("%016x", hash)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
