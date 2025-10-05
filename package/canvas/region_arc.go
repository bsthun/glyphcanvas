package canvas

import (
	"fmt"
	"math"
	"sort"
)

type ArcType int

const (
	ArcTypeCircle ArcType = iota
	ArcTypeStrengthLine
	ArcTypeCurveLine
	ArcTypeTriangle
	ArcTypeRectangle
)

type ArcFillType int

const (
	ArcFillTypeFill ArcFillType = iota
	ArcFillTypeStroke
)

type Arc struct {
	Type               ArcType
	Fill               ArcFillType
	CircleEllipseRatio float32
	LineDegree         float32
	ArcLineTheta       float32
}

type EdgePoint struct {
	X, Y  int
	Angle float64
}

type HoughAccumulator struct {
	rho   float64
	theta float64
	votes int
}

func (r *Region) Arc() *Arc {
	if len(r.Draws) < 3 {
		return nil
	}

	edges := r.extractEdges()
	if len(edges) < 3 {
		return nil
	}

	chainCode := r.computeChainCode(edges)
	curvatures := r.computeCurvatures(chainCode)

	moments := r.computeMoments()
	huInvariants := r.computeHuInvariants(moments)

	lines := r.detectLinesHough(edges)
	circles := r.detectCirclesHough(edges)

	arcType, fillType := r.classifyShape(huInvariants, curvatures, lines, circles)

	arc := &Arc{
		Type: arcType,
		Fill: fillType,
	}

	switch arcType {
	case ArcTypeCircle:
		arc.CircleEllipseRatio = r.computeEllipseRatio(moments)

	case ArcTypeStrengthLine:
		arc.LineDegree = r.computeLineDegree(lines)
		fmt.Printf("Line detected with degree: %.0f°\n", arc.LineDegree)

	case ArcTypeCurveLine:
		arc.ArcLineTheta = r.computeCurveStrength(curvatures, edges)
		fmt.Printf("Curve detected with strength: %.3f\n", arc.ArcLineTheta)

	case ArcTypeTriangle:
		corners := r.detectCorners(curvatures, edges)
		if len(corners) == 3 {
			fmt.Println("Triangle detected")
		}

	case ArcTypeRectangle:
		corners := r.detectCorners(curvatures, edges)
		if len(corners) == 4 {
			fmt.Println("Rectangle detected")
		}
	}

	r.printDetectedAngles(edges)

	return arc
}

func (r *Region) extractEdges() []EdgePoint {
	edges := []EdgePoint{}
	dx := []int{-1, 0, 1, -1, 1, -1, 0, 1}
	dy := []int{-1, -1, -1, 0, 0, 1, 1, 1}

	for x := uint16(1); x < r.SizeX-1; x++ {
		for y := uint16(1); y < r.SizeY-1; y++ {
			if !r.IsDrew(x, y) {
				continue
			}

			isEdge := false
			for i := 0; i < 8; i++ {
				nx := int(x) + dx[i]
				ny := int(y) + dy[i]
				if nx >= 0 && nx < int(r.SizeX) && ny >= 0 && ny < int(r.SizeY) {
					if !r.IsDrew(uint16(nx), uint16(ny)) {
						isEdge = true
						break
					}
				}
			}

			if isEdge {
				angle := r.computeGradientAngle(x, y)
				edges = append(edges, EdgePoint{int(x), int(y), angle})
			}
		}
	}

	return edges
}

func (r *Region) computeGradientAngle(x, y uint16) float64 {
	gx := 0.0
	gy := 0.0

	sobelX := [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	sobelY := [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			px := int(x) + i
			py := int(y) + j
			if px >= 0 && px < int(r.SizeX) && py >= 0 && py < int(r.SizeY) {
				val := 0.0
				if r.IsDrew(uint16(px), uint16(py)) {
					val = 1.0
				}
				gx += val * sobelX[j+1][i+1]
				gy += val * sobelY[j+1][i+1]
			}
		}
	}

	return math.Atan2(gy, gx)
}

func (r *Region) computeChainCode(edges []EdgePoint) []int {
	if len(edges) < 2 {
		return []int{}
	}

	sortedEdges := r.sortEdgesForContour(edges)
	chainCode := []int{}

	for i := 1; i < len(sortedEdges); i++ {
		dx := sortedEdges[i].X - sortedEdges[i-1].X
		dy := sortedEdges[i].Y - sortedEdges[i-1].Y

		code := 0
		if dx == 1 && dy == 0 {
			code = 0
		} else if dx == 1 && dy == -1 {
			code = 1
		} else if dx == 0 && dy == -1 {
			code = 2
		} else if dx == -1 && dy == -1 {
			code = 3
		} else if dx == -1 && dy == 0 {
			code = 4
		} else if dx == -1 && dy == 1 {
			code = 5
		} else if dx == 0 && dy == 1 {
			code = 6
		} else if dx == 1 && dy == 1 {
			code = 7
		}

		chainCode = append(chainCode, code)
	}

	return chainCode
}

func (r *Region) sortEdgesForContour(edges []EdgePoint) []EdgePoint {
	if len(edges) == 0 {
		return edges
	}

	sorted := make([]EdgePoint, 0, len(edges))
	visited := make(map[int]bool)

	current := edges[0]
	sorted = append(sorted, current)
	visited[0] = true

	for len(sorted) < len(edges) {
		minDist := math.MaxFloat64
		minIdx := -1

		for i, edge := range edges {
			if visited[i] {
				continue
			}

			dist := math.Sqrt(float64((edge.X-current.X)*(edge.X-current.X) +
				(edge.Y-current.Y)*(edge.Y-current.Y)))

			if dist < minDist {
				minDist = dist
				minIdx = i
			}
		}

		if minIdx == -1 {
			break
		}

		current = edges[minIdx]
		sorted = append(sorted, current)
		visited[minIdx] = true
	}

	return sorted
}

func (r *Region) computeCurvatures(chainCode []int) []float64 {
	curvatures := make([]float64, len(chainCode))

	for i := 0; i < len(chainCode); i++ {
		prev := chainCode[(i-1+len(chainCode))%len(chainCode)]
		curr := chainCode[i]
		next := chainCode[(i+1)%len(chainCode)]

		angle1 := float64(curr-prev) * math.Pi / 4.0
		angle2 := float64(next-curr) * math.Pi / 4.0

		if angle1 > math.Pi {
			angle1 -= 2 * math.Pi
		} else if angle1 < -math.Pi {
			angle1 += 2 * math.Pi
		}

		if angle2 > math.Pi {
			angle2 -= 2 * math.Pi
		} else if angle2 < -math.Pi {
			angle2 += 2 * math.Pi
		}

		curvatures[i] = (angle1 + angle2) / 2.0
	}

	return curvatures
}

func (r *Region) computeMoments() map[string]float64 {
	moments := make(map[string]float64)

	m00, m10, m01, m11, m20, m02, m21, m12, m30, m03 := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0

	for x := uint16(0); x < r.SizeX; x++ {
		for y := uint16(0); y < r.SizeY; y++ {
			if r.IsDrew(x, y) {
				fx := float64(x)
				fy := float64(y)

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
		}
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

func (r *Region) computeHuInvariants(moments map[string]float64) []float64 {
	hu := make([]float64, 7)

	m00 := moments["m00"]
	if m00 == 0 {
		return hu
	}

	norm := math.Pow(m00, 2.5)

	eta20 := moments["mu20"] / norm
	eta02 := moments["mu02"] / norm
	eta11 := moments["mu11"] / norm
	eta30 := moments["mu30"] / math.Pow(m00, 3.5)
	eta21 := moments["mu21"] / math.Pow(m00, 3.5)
	eta12 := moments["mu12"] / math.Pow(m00, 3.5)
	eta03 := moments["mu03"] / math.Pow(m00, 3.5)

	hu[0] = eta20 + eta02
	hu[1] = math.Pow(eta20-eta02, 2) + 4*math.Pow(eta11, 2)
	hu[2] = math.Pow(eta30-3*eta12, 2) + math.Pow(3*eta21-eta03, 2)
	hu[3] = math.Pow(eta30+eta12, 2) + math.Pow(eta21+eta03, 2)
	hu[4] = (eta30-3*eta12)*(eta30+eta12)*(math.Pow(eta30+eta12, 2)-3*math.Pow(eta21+eta03, 2)) +
		(3*eta21-eta03)*(eta21+eta03)*(3*math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2))
	hu[5] = (eta20-eta02)*(math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2)) +
		4*eta11*(eta30+eta12)*(eta21+eta03)
	hu[6] = (3*eta21-eta03)*(eta30+eta12)*(math.Pow(eta30+eta12, 2)-3*math.Pow(eta21+eta03, 2)) -
		(eta30-3*eta12)*(eta21+eta03)*(3*math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2))

	return hu
}

func (r *Region) detectLinesHough(edges []EdgePoint) []HoughAccumulator {
	if len(edges) < 2 {
		return []HoughAccumulator{}
	}

	maxRho := math.Sqrt(float64(r.SizeX*r.SizeX + r.SizeY*r.SizeY))
	rhoStep := 1.0
	thetaStep := math.Pi / 180.0

	accumulator := make(map[string]int)

	for _, edge := range edges {
		for theta := 0.0; theta < math.Pi; theta += thetaStep {
			rho := float64(edge.X)*math.Cos(theta) + float64(edge.Y)*math.Sin(theta)

			rhoIdx := int((rho + maxRho) / rhoStep)
			thetaIdx := int(theta / thetaStep)

			key := fmt.Sprintf("%d_%d", rhoIdx, thetaIdx)
			accumulator[key]++
		}
	}

	threshold := len(edges) / 4
	lines := []HoughAccumulator{}

	for key, votes := range accumulator {
		if votes > threshold {
			var rhoIdx, thetaIdx int
			fmt.Sscanf(key, "%d_%d", &rhoIdx, &thetaIdx)

			rho := float64(rhoIdx)*rhoStep - maxRho
			theta := float64(thetaIdx) * thetaStep

			lines = append(lines, HoughAccumulator{
				rho:   rho,
				theta: theta,
				votes: votes,
			})
		}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].votes > lines[j].votes
	})

	if len(lines) > 5 {
		lines = lines[:5]
	}

	return lines
}

func (r *Region) detectCirclesHough(edges []EdgePoint) []HoughAccumulator {
	if len(edges) < 3 {
		return []HoughAccumulator{}
	}

	minRadius := 5.0
	maxRadius := math.Min(float64(r.SizeX), float64(r.SizeY)) / 2.0

	accumulator := make(map[string]int)

	for _, edge := range edges {
		for radius := minRadius; radius <= maxRadius; radius += 2.0 {
			for theta := 0.0; theta < 2*math.Pi; theta += math.Pi / 18 {
				a := float64(edge.X) - radius*math.Cos(theta)
				b := float64(edge.Y) - radius*math.Sin(theta)

				if a >= 0 && a < float64(r.SizeX) && b >= 0 && b < float64(r.SizeY) {
					key := fmt.Sprintf("%.0f_%.0f_%.0f", a, b, radius)
					accumulator[key]++
				}
			}
		}
	}

	threshold := len(edges) / 10
	circles := []HoughAccumulator{}

	for key, votes := range accumulator {
		if votes > threshold {
			var a, b, radius float64
			fmt.Sscanf(key, "%f_%f_%f", &a, &b, &radius)

			circles = append(circles, HoughAccumulator{
				rho:   radius,
				theta: math.Atan2(b, a),
				votes: votes,
			})
		}
	}

	sort.Slice(circles, func(i, j int) bool {
		return circles[i].votes > circles[j].votes
	})

	if len(circles) > 3 {
		circles = circles[:3]
	}

	return circles
}

func (r *Region) classifyShape(hu []float64, curvatures []float64, lines, circles []HoughAccumulator) (ArcType, ArcFillType) {
	fillType := r.determineFillType()

	if len(circles) > 0 && circles[0].votes > len(r.Draws)/3 {
		circularity := r.computeCircularity(hu)
		if circularity > 0.7 {
			return ArcTypeCircle, fillType
		}
	}

	if len(lines) > 0 && lines[0].votes > len(r.Draws)/2 {
		linearity := r.computeLinearity(hu)
		if linearity > 0.8 {
			return ArcTypeStrengthLine, fillType
		}
	}

	corners := r.detectCorners(curvatures, nil)
	if len(corners) == 3 {
		return ArcTypeTriangle, fillType
	} else if len(corners) == 4 {
		rectangularity := r.computeRectangularity(hu)
		if rectangularity > 0.7 {
			return ArcTypeRectangle, fillType
		}
	}

	avgCurvature := 0.0
	for _, c := range curvatures {
		avgCurvature += math.Abs(c)
	}
	if len(curvatures) > 0 {
		avgCurvature /= float64(len(curvatures))
	}

	if avgCurvature > 0.1 && avgCurvature < 0.8 {
		return ArcTypeCurveLine, fillType
	}

	return ArcTypeStrengthLine, fillType
}

func (r *Region) determineFillType() ArcFillType {
	edgeCount := 0
	totalCount := 0

	for x := uint16(1); x < r.SizeX-1; x++ {
		for y := uint16(1); y < r.SizeY-1; y++ {
			if r.IsDrew(x, y) {
				totalCount++

				hasEmpty := false
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx := int(x) + dx
						ny := int(y) + dy
						if nx >= 0 && nx < int(r.SizeX) && ny >= 0 && ny < int(r.SizeY) {
							if !r.IsDrew(uint16(nx), uint16(ny)) {
								hasEmpty = true
								break
							}
						}
					}
					if hasEmpty {
						break
					}
				}

				if hasEmpty {
					edgeCount++
				}
			}
		}
	}

	if totalCount > 0 {
		ratio := float64(edgeCount) / float64(totalCount)
		if ratio > 0.3 {
			return ArcFillTypeStroke
		}
	}

	return ArcFillTypeFill
}

func (r *Region) computeCircularity(hu []float64) float64 {
	if len(hu) < 2 {
		return 0
	}

	I1 := hu[0]
	I2 := hu[1]

	if I1 > 0 {
		circularity := 1.0 / (1.0 + math.Sqrt(I2)/I1)
		return circularity
	}

	return 0
}

func (r *Region) computeLinearity(hu []float64) float64 {
	if len(hu) < 3 {
		return 0
	}

	sum := 0.0
	for i := 2; i < len(hu); i++ {
		sum += math.Abs(hu[i])
	}

	if sum < 0.01 {
		return 1.0
	}

	return 1.0 / (1.0 + sum*100)
}

func (r *Region) computeRectangularity(hu []float64) float64 {
	if len(hu) < 7 {
		return 0
	}

	rectangleHu := []float64{0.16, 0.0013, 0, 0, 0, 0, 0}

	diff := 0.0
	for i := 0; i < 7; i++ {
		diff += math.Abs(hu[i] - rectangleHu[i])
	}

	return math.Exp(-diff * 10)
}

func (r *Region) detectCorners(curvatures []float64, edges []EdgePoint) []int {
	corners := []int{}
	threshold := math.Pi / 6

	for i := 0; i < len(curvatures); i++ {
		if math.Abs(curvatures[i]) > threshold {
			isLocalMax := true
			for j := i - 2; j <= i+2; j++ {
				if j < 0 || j >= len(curvatures) || j == i {
					continue
				}
				if math.Abs(curvatures[j]) > math.Abs(curvatures[i]) {
					isLocalMax = false
					break
				}
			}

			if isLocalMax {
				corners = append(corners, i)
			}
		}
	}

	return corners
}

func (r *Region) computeEllipseRatio(moments map[string]float64) float32 {
	mu20 := moments["mu20"]
	mu02 := moments["mu02"]
	mu11 := moments["mu11"]

	if mu20+mu02 == 0 {
		return 1.0
	}

	lambda1 := (mu20 + mu02 + math.Sqrt(math.Pow(mu20-mu02, 2)+4*mu11*mu11)) / 2
	lambda2 := (mu20 + mu02 - math.Sqrt(math.Pow(mu20-mu02, 2)+4*mu11*mu11)) / 2

	if lambda1 > 0 && lambda2 > 0 {
		ratio := math.Min(lambda1, lambda2) / math.Max(lambda1, lambda2)
		return float32(ratio)
	}

	return 1.0
}

func (r *Region) computeLineDegree(lines []HoughAccumulator) float32 {
	if len(lines) == 0 {
		return 0
	}

	theta := lines[0].theta
	degree := theta * 180.0 / math.Pi

	// Normalize angle to be in terms of line direction (0-180 degrees)
	// Hough theta represents perpendicular to line, so adjust by 90 degrees
	lineDegree := degree - 90
	if lineDegree < 0 {
		lineDegree += 180
	}
	if lineDegree >= 180 {
		lineDegree -= 180
	}

	targetAngles := []float64{0, 45, 90, 135}
	minDiff := math.MaxFloat64
	bestAngle := 0.0

	for _, target := range targetAngles {
		diff := math.Abs(lineDegree - target)
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
		// Check wraparound at 180
		diff = math.Abs(lineDegree - (target + 180))
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
		diff = math.Abs(lineDegree - (target - 180))
		if diff < minDiff {
			minDiff = diff
			bestAngle = target
		}
	}

	// For display purposes, we can use 180 instead of 0 for vertical lines
	if bestAngle == 0 && math.Abs(lineDegree-180) < math.Abs(lineDegree) {
		bestAngle = 180
	}

	return float32(bestAngle)
}

func (r *Region) computeCurveStrength(curvatures []float64, edges []EdgePoint) float32 {
	if len(curvatures) == 0 {
		return 0
	}

	totalCurvature := 0.0
	positiveCurvature := 0.0

	for _, c := range curvatures {
		totalCurvature += math.Abs(c)
		if c > 0 {
			positiveCurvature += c
		}
	}

	avgCurvature := totalCurvature / float64(len(curvatures))

	direction := 0.0
	if len(edges) >= 2 {
		start := edges[0]
		end := edges[len(edges)-1]

		dx := float64(end.X - start.X)
		dy := float64(end.Y - start.Y)

		if dx > 0 || dy < 0 {
			direction = 1.0
		} else {
			direction = -1.0
		}
	}

	strength := math.Tanh(avgCurvature * 2)

	if totalCurvature > 0 {
		bias := (positiveCurvature / totalCurvature) - 0.5
		strength += bias * 0.3
	}

	strength *= direction

	strength = math.Max(-1.0, math.Min(1.0, strength))

	return float32(strength)
}

func (r *Region) printDetectedAngles(edges []EdgePoint) {
	if len(edges) == 0 {
		return
	}

	angleHistogram := make(map[int]int)
	targetAngles := []int{0, 45, 90, 135, 180}

	for _, edge := range edges {
		degree := edge.Angle * 180.0 / math.Pi

		if degree < 0 {
			degree += 180
		}

		for _, target := range targetAngles {
			if math.Abs(degree-float64(target)) < 22.5 ||
				math.Abs(degree-float64(target)+180) < 22.5 ||
				math.Abs(degree-float64(target)-180) < 22.5 {
				angleHistogram[target]++
				break
			}
		}
	}

	fmt.Println("Detected angle distribution in region:")
	for _, angle := range targetAngles {
		count := angleHistogram[angle]
		percentage := float64(count) * 100.0 / float64(len(edges))
		if count > 0 {
			fmt.Printf("  %3d°: %d edges (%.1f%%)\n", angle, count, percentage)
		}
	}
}
