package region

type Region struct {
	SizeX  uint16                     `json:"sizeX"`
	SizeY  uint16                     `json:"sizeY"`
	Bitmap map[uint16]map[uint16]bool `json:"bitmap"`
	Draws  []*Point                   `json:"draws"`
}

func NewRegion(sizeX, sizeY uint16) *Region {
	return &Region{
		SizeX:  sizeX,
		SizeY:  sizeY,
		Bitmap: make(map[uint16]map[uint16]bool),
		Draws:  []*Point{},
	}
}

func (r *Region) IsDrew(x, y uint16) bool {
	if _, ok := r.Bitmap[x]; !ok {
		return false
	}
	if _, ok := r.Bitmap[x][y]; !ok {
		return false
	}
	return r.Bitmap[x][y]
}

func (r *Region) Draw(x, y uint16) {
	if _, ok := r.Bitmap[x]; !ok {
		r.Bitmap[x] = make(map[uint16]bool)
	}
	r.Bitmap[x][y] = true
	r.Draws = append(r.Draws, &Point{X: x, Y: y})
}

func (r *Region) Erase(x, y uint16) {
	if _, ok := r.Bitmap[x]; !ok {
		return
	}
	r.Bitmap[x][y] = false
}

func (r *Region) GetSizeX() uint16 {
	return r.SizeX
}

func (r *Region) GetSizeY() uint16 {
	return r.SizeY
}
