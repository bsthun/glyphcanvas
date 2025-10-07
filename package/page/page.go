package page

import (
	"image"
	"image/color"
	"sort"

	"github.com/bsthun/glyphcanvas/package/character"
)

type Page struct {
	Width     int                `json:"width"`
	Height    int                `json:"height"`
	Image     image.Image        `json:"-"`
	TextAreas []*TextArea        `json:"text_areas"`
	Lines     []*TextLine        `json:"lines"`
	Words     []*Word            `json:"words"`
	Chars     []*CharacterBounds `json:"characters"`
}

type TextArea struct {
	X      int         `json:"x"`
	Y      int         `json:"y"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Lines  []*TextLine `json:"lines"`
}

type TextLine struct {
	X        int                `json:"x"`
	Y        int                `json:"y"`
	Width    int                `json:"width"`
	Height   int                `json:"height"`
	Words    []*Word            `json:"words"`
	Text     string             `json:"text"`
	Baseline int                `json:"baseline"`
	Chars    []*CharacterBounds `json:"characters"`
}

type Word struct {
	X          int                `json:"x"`
	Y          int                `json:"y"`
	Width      int                `json:"width"`
	Height     int                `json:"height"`
	Text       string             `json:"text"`
	Chars      []*CharacterBounds `json:"characters"`
	Confidence float64            `json:"confidence"`
}

type CharacterBounds struct {
	X          int                  `json:"x"`
	Y          int                  `json:"y"`
	Width      int                  `json:"width"`
	Height     int                  `json:"height"`
	Character  *character.Character `json:"-"`
	Unicode    string               `json:"unicode"`
	Text       string               `json:"text"`
	Confidence float64              `json:"confidence"`
}

func NewPage(img image.Image) *Page {
	bounds := img.Bounds()
	return &Page{
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Image:     img,
		TextAreas: []*TextArea{},
		Lines:     []*TextLine{},
		Words:     []*Word{},
		Chars:     []*CharacterBounds{},
	}
}

func (p *Page) DetectTextAreas() error {
	textAreas := findTextAreas(p.Image)
	p.TextAreas = textAreas
	return nil
}

func (p *Page) DetectLines() error {
	for _, area := range p.TextAreas {
		lines := findLinesInArea(p.Image, area)
		area.Lines = lines
		p.Lines = append(p.Lines, lines...)
	}

	sort.Slice(p.Lines, func(i, j int) bool {
		if p.Lines[i].Y != p.Lines[j].Y {
			return p.Lines[i].Y < p.Lines[j].Y
		}
		return p.Lines[i].X < p.Lines[j].X
	})

	return nil
}

func (p *Page) DetectWords() error {
	for _, line := range p.Lines {
		words := findWordsInLine(p.Image, line)
		line.Words = words
		p.Words = append(p.Words, words...)
	}
	return nil
}

func (p *Page) DetectCharacters() error {
	for _, word := range p.Words {
		chars := findCharactersInWord(p.Image, word)
		word.Chars = chars
		p.Chars = append(p.Chars, chars...)
	}

	for _, line := range p.Lines {
		for _, word := range line.Words {
			line.Chars = append(line.Chars, word.Chars...)
		}
	}

	return nil
}

func (p *Page) GetText() string {
	text := ""
	for i, line := range p.Lines {
		if i > 0 {
			text += "\n"
		}
		text += line.Text
	}
	return text
}

func (p *Page) GetPlainText() string {
	text := ""
	for i, line := range p.Lines {
		if i > 0 {
			text += "\n"
		}
		lineText := ""
		for j, word := range line.Words {
			if j > 0 {
				lineText += " "
			}
			lineText += word.Text
		}
		text += lineText
	}
	return text
}

func findTextAreas(img image.Image) []*TextArea {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to binary
	binary := make([][]bool, height)
	for y := 0; y < height; y++ {
		binary[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			c := color.GrayModel.Convert(img.At(x+bounds.Min.X, y+bounds.Min.Y)).(color.Gray)
			binary[y][x] = c.Y < 128
		}
	}

	// Find horizontal projections
	hProjection := make([]int, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if binary[y][x] {
				hProjection[y]++
			}
		}
	}

	// Find text blocks based on horizontal projection
	var areas []*TextArea
	inText := false
	startY := 0
	threshold := width / 50 // Minimum pixels per line to consider text

	for y := 0; y < height; y++ {
		if hProjection[y] > threshold && !inText {
			inText = true
			startY = y
		} else if hProjection[y] <= threshold && inText {
			inText = false
			if y-startY > 10 { // Minimum height for text area
				area := &TextArea{
					X:      0,
					Y:      startY,
					Width:  width,
					Height: y - startY,
					Lines:  []*TextLine{},
				}
				areas = append(areas, area)
			}
		}
	}

	// Handle case where text continues to end of image
	if inText && height-startY > 10 {
		area := &TextArea{
			X:      0,
			Y:      startY,
			Width:  width,
			Height: height - startY,
			Lines:  []*TextLine{},
		}
		areas = append(areas, area)
	}

	return areas
}

func findLinesInArea(img image.Image, area *TextArea) []*TextLine {
	bounds := img.Bounds()

	// Extract area image
	binary := make([][]bool, area.Height)
	for y := 0; y < area.Height; y++ {
		binary[y] = make([]bool, area.Width)
		for x := 0; x < area.Width; x++ {
			imgY := y + area.Y + bounds.Min.Y
			imgX := x + area.X + bounds.Min.X
			c := color.GrayModel.Convert(img.At(imgX, imgY)).(color.Gray)
			binary[y][x] = c.Y < 128
		}
	}

	// Find horizontal projection for lines
	hProjection := make([]int, area.Height)
	for y := 0; y < area.Height; y++ {
		for x := 0; x < area.Width; x++ {
			if binary[y][x] {
				hProjection[y]++
			}
		}
	}

	// Find individual lines
	var lines []*TextLine
	inLine := false
	startY := 0
	threshold := area.Width / 100 // Minimum pixels per line

	for y := 0; y < area.Height; y++ {
		if hProjection[y] > threshold && !inLine {
			inLine = true
			startY = y
		} else if hProjection[y] <= threshold && inLine {
			inLine = false
			if y-startY > 5 { // Minimum line height
				// Find actual text bounds in this line
				minX, maxX := findLineBounds(binary, startY, y)
				if maxX > minX {
					line := &TextLine{
						X:        area.X + minX,
						Y:        area.Y + startY,
						Width:    maxX - minX,
						Height:   y - startY,
						Words:    []*Word{},
						Text:     "",
						Baseline: area.Y + startY + (y-startY)*3/4, // Approximate baseline
						Chars:    []*CharacterBounds{},
					}
					lines = append(lines, line)
				}
			}
		}
	}

	// Handle case where line continues to end of area
	if inLine && area.Height-startY > 5 {
		minX, maxX := findLineBounds(binary, startY, area.Height)
		if maxX > minX {
			line := &TextLine{
				X:        area.X + minX,
				Y:        area.Y + startY,
				Width:    maxX - minX,
				Height:   area.Height - startY,
				Words:    []*Word{},
				Text:     "",
				Baseline: area.Y + startY + (area.Height-startY)*3/4,
				Chars:    []*CharacterBounds{},
			}
			lines = append(lines, line)
		}
	}

	return lines
}

func findLineBounds(binary [][]bool, startY, endY int) (int, int) {
	minX := len(binary[0])
	maxX := 0

	for y := startY; y < endY && y < len(binary); y++ {
		for x := 0; x < len(binary[y]); x++ {
			if binary[y][x] {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}
	}

	return minX, maxX + 1
}

func findWordsInLine(img image.Image, line *TextLine) []*Word {
	bounds := img.Bounds()

	// Extract line image
	binary := make([][]bool, line.Height)
	for y := 0; y < line.Height; y++ {
		binary[y] = make([]bool, line.Width)
		for x := 0; x < line.Width; x++ {
			imgY := y + line.Y + bounds.Min.Y
			imgX := x + line.X + bounds.Min.X
			c := color.GrayModel.Convert(img.At(imgX, imgY)).(color.Gray)
			binary[y][x] = c.Y < 128
		}
	}

	// Find vertical projection
	vProjection := make([]int, line.Width)
	for x := 0; x < line.Width; x++ {
		for y := 0; y < line.Height; y++ {
			if binary[y][x] {
				vProjection[x]++
			}
		}
	}

	// Find word boundaries
	var words []*Word
	inWord := false
	startX := 0
	threshold := 1 // Minimum pixels per column to be part of word

	for x := 0; x < line.Width; x++ {
		if vProjection[x] > threshold && !inWord {
			inWord = true
			startX = x
		} else if vProjection[x] <= threshold && inWord {
			inWord = false
			if x-startX > 3 { // Minimum word width
				word := &Word{
					X:          line.X + startX,
					Y:          line.Y,
					Width:      x - startX,
					Height:     line.Height,
					Text:       "",
					Chars:      []*CharacterBounds{},
					Confidence: 0.0,
				}
				words = append(words, word)
			}
		}
	}

	// Handle case where word continues to end of line
	if inWord && line.Width-startX > 3 {
		word := &Word{
			X:          line.X + startX,
			Y:          line.Y,
			Width:      line.Width - startX,
			Height:     line.Height,
			Text:       "",
			Chars:      []*CharacterBounds{},
			Confidence: 0.0,
		}
		words = append(words, word)
	}

	return words
}

func findCharactersInWord(img image.Image, word *Word) []*CharacterBounds {
	bounds := img.Bounds()

	// Extract word image
	binary := make([][]bool, word.Height)
	for y := 0; y < word.Height; y++ {
		binary[y] = make([]bool, word.Width)
		for x := 0; x < word.Width; x++ {
			imgY := y + word.Y + bounds.Min.Y
			imgX := x + word.X + bounds.Min.X
			c := color.GrayModel.Convert(img.At(imgX, imgY)).(color.Gray)
			binary[y][x] = c.Y < 128
		}
	}

	// Find character boundaries using connected components
	chars := findConnectedComponents(binary, word)

	// Sort characters left to right
	sort.Slice(chars, func(i, j int) bool {
		return chars[i].X < chars[j].X
	})

	return chars
}

func findConnectedComponents(binary [][]bool, word *Word) []*CharacterBounds {
	height := len(binary)
	width := len(binary[0])
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	var chars []*CharacterBounds

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if binary[y][x] && !visited[y][x] {
				minX, minY, maxX, maxY := floodFill(binary, visited, x, y)

				// Filter out noise (very small components)
				if maxX-minX >= 2 && maxY-minY >= 3 {
					charImg := extractCharacterImage(binary, minX, minY, maxX-minX+1, maxY-minY+1)

					char := &CharacterBounds{
						X:          word.X + minX,
						Y:          word.Y + minY,
						Width:      maxX - minX + 1,
						Height:     maxY - minY + 1,
						Character:  charImg,
						Unicode:    "",
						Text:       "",
						Confidence: 0.0,
					}
					chars = append(chars, char)
				}
			}
		}
	}

	return chars
}

func floodFill(binary, visited [][]bool, startX, startY int) (int, int, int, int) {
	height := len(binary)
	width := len(binary[0])

	minX, minY := startX, startY
	maxX, maxY := startX, startY

	stack := [][2]int{{startX, startY}}

	for len(stack) > 0 {
		x, y := stack[len(stack)-1][0], stack[len(stack)-1][1]
		stack = stack[:len(stack)-1]

		if x < 0 || x >= width || y < 0 || y >= height || visited[y][x] || !binary[y][x] {
			continue
		}

		visited[y][x] = true

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}

		// Add neighbors
		stack = append(stack, [2]int{x + 1, y})
		stack = append(stack, [2]int{x - 1, y})
		stack = append(stack, [2]int{x, y + 1})
		stack = append(stack, [2]int{x, y - 1})
	}

	return minX, minY, maxX, maxY
}

func extractCharacterImage(binary [][]bool, x, y, width, height int) *character.Character {
	char := character.NewCharacter(uint16(width), uint16(height), nil)

	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			sourceY := y + py
			sourceX := x + px

			if sourceY >= 0 && sourceY < len(binary) &&
				sourceX >= 0 && sourceX < len(binary[sourceY]) &&
				binary[sourceY][sourceX] {
				char.Draw(uint16(px), uint16(py))
			}
		}
	}

	return char
}
