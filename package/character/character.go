package character

import (
	"github.com/bsthun/glyphcanvas/package/region"
)

type Point struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

type AnchorPoint struct {
	Point     *Point  `json:"point"`
	Type      string  `json:"type"`      // "entry", "exit", "junction", "terminal", "extremum"
	Strength  float64 `json:"strength"`  // Significance of the anchor point (0-1)
	Curvature float64 `json:"curvature"` // Local curvature at this point
	Angle     float64 `json:"angle"`     // Direction angle in radians
}

type Character struct {
	SizeX  uint16                     `json:"sizeX"`
	SizeY  uint16                     `json:"sizeY"`
	Bitmap map[uint16]map[uint16]bool `json:"bitmap"`
	Draws  []*Point                   `json:"draws"`

	// Character-specific properties
	AnchorPoints     []*AnchorPoint      `json:"anchorPoints"`
	Regions          []*region.Region    `json:"regions"`
	MedialAxis       []*Point            `json:"medialAxis"`
	SkeletonBranches map[string][]*Point `json:"skeletonBranches"`

	// Analysis results
	Topology    map[string]interface{} `json:"topology"`
	Moments     map[string]float64     `json:"moments"`
	BoundingBox map[string]uint16      `json:"boundingBox"`

	// Configuration
	Config *CharacterConfig `json:"config"`
}

func NewCharacter(sizeX, sizeY uint16, config *CharacterConfig) *Character {
	if config == nil {
		config = DefaultCharacterConfig()
	}

	return &Character{
		SizeX:            sizeX,
		SizeY:            sizeY,
		Bitmap:           make(map[uint16]map[uint16]bool),
		Draws:            []*Point{},
		AnchorPoints:     []*AnchorPoint{},
		Regions:          []*region.Region{},
		MedialAxis:       []*Point{},
		SkeletonBranches: make(map[string][]*Point),
		Topology:         make(map[string]interface{}),
		Moments:          make(map[string]float64),
		BoundingBox:      make(map[string]uint16),
		Config:           config,
	}
}

func (c *Character) IsDrew(x, y uint16) bool {
	if _, ok := c.Bitmap[x]; !ok {
		return false
	}
	if _, ok := c.Bitmap[x][y]; !ok {
		return false
	}
	return c.Bitmap[x][y]
}

func (c *Character) Draw(x, y uint16) {
	if _, ok := c.Bitmap[x]; !ok {
		c.Bitmap[x] = make(map[uint16]bool)
	}
	c.Bitmap[x][y] = true
	c.Draws = append(c.Draws, &Point{X: x, Y: y})

	// Update bounding box
	c.updateBoundingBox(x, y)
}

func (c *Character) Erase(x, y uint16) {
	if _, ok := c.Bitmap[x]; !ok {
		return
	}
	c.Bitmap[x][y] = false

	// Remove from draws slice
	for i, point := range c.Draws {
		if point.X == x && point.Y == y {
			c.Draws = append(c.Draws[:i], c.Draws[i+1:]...)
			break
		}
	}

	// Recalculate bounding box if needed
	c.recalculateBoundingBox()
}

func (c *Character) GetSizeX() uint16 {
	return c.SizeX
}

func (c *Character) GetSizeY() uint16 {
	return c.SizeY
}

func (c *Character) GetPixelCount() int {
	return len(c.Draws)
}

func (c *Character) IsEmpty() bool {
	return len(c.Draws) == 0
}

func (c *Character) updateBoundingBox(x, y uint16) {
	if len(c.Draws) == 1 {
		// First pixel
		c.BoundingBox["minX"] = x
		c.BoundingBox["maxX"] = x
		c.BoundingBox["minY"] = y
		c.BoundingBox["maxY"] = y
		return
	}

	if x < c.BoundingBox["minX"] {
		c.BoundingBox["minX"] = x
	}
	if x > c.BoundingBox["maxX"] {
		c.BoundingBox["maxX"] = x
	}
	if y < c.BoundingBox["minY"] {
		c.BoundingBox["minY"] = y
	}
	if y > c.BoundingBox["maxY"] {
		c.BoundingBox["maxY"] = y
	}
}

func (c *Character) recalculateBoundingBox() {
	if len(c.Draws) == 0 {
		c.BoundingBox = make(map[string]uint16)
		return
	}

	minX, maxX := c.Draws[0].X, c.Draws[0].X
	minY, maxY := c.Draws[0].Y, c.Draws[0].Y

	for _, point := range c.Draws {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	c.BoundingBox["minX"] = minX
	c.BoundingBox["maxX"] = maxX
	c.BoundingBox["minY"] = minY
	c.BoundingBox["maxY"] = maxY
}

func (c *Character) GetBoundingBoxWidth() uint16 {
	if len(c.BoundingBox) == 0 {
		return 0
	}
	return c.BoundingBox["maxX"] - c.BoundingBox["minX"] + 1
}

func (c *Character) GetBoundingBoxHeight() uint16 {
	if len(c.BoundingBox) == 0 {
		return 0
	}
	return c.BoundingBox["maxY"] - c.BoundingBox["minY"] + 1
}

func (c *Character) AddAnchorPoint(x, y uint16, anchorType string, strength, curvature, angle float64) {
	anchor := &AnchorPoint{
		Point:     &Point{X: x, Y: y},
		Type:      anchorType,
		Strength:  strength,
		Curvature: curvature,
		Angle:     angle,
	}
	c.AnchorPoints = append(c.AnchorPoints, anchor)
}

func (c *Character) GetAnchorPointsByType(anchorType string) []*AnchorPoint {
	var result []*AnchorPoint
	for _, anchor := range c.AnchorPoints {
		if anchor.Type == anchorType {
			result = append(result, anchor)
		}
	}
	return result
}

func (c *Character) ClearAnalysisResults() {
	c.AnchorPoints = []*AnchorPoint{}
	c.Regions = []*region.Region{}
	c.MedialAxis = []*Point{}
	c.SkeletonBranches = make(map[string][]*Point)
	c.Topology = make(map[string]interface{})
	c.Moments = make(map[string]float64)
}
