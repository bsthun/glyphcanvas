package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/bsthun/glyphcanvas/package/page"
	"github.com/bsthun/gut"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontManager handles loading and managing fonts
type FontManager struct {
	ThaiFont    font.Face
	EnglishFont font.Face
}

// NewFontManager creates a new font manager with loaded fonts
func NewFontManager() (*FontManager, error) {
	fm := &FontManager{}

	// Load Thai font (Noto Sans Thai)
	thaiFontPath := "generate/font/NotoSansThaiLooped-Regular.ttf"
	thaiFont, err := loadFont(thaiFontPath, 12)
	if err != nil {
		fmt.Printf("Warning: Failed to load Thai font %s: %v\n", thaiFontPath, err)
		fm.ThaiFont = basicfont.Face7x13
	} else {
		fm.ThaiFont = thaiFont
	}

	// Load English font (Roboto)
	englishFontPath := "generate/font/Roboto-Regular.ttf"
	englishFont, err := loadFont(englishFontPath, 12)
	if err != nil {
		fmt.Printf("Warning: Failed to load English font %s: %v\n", englishFontPath, err)
		fm.EnglishFont = basicfont.Face7x13
	} else {
		fm.EnglishFont = englishFont
	}

	return fm, nil
}

// GetFont returns appropriate font for the given text
func (fm *FontManager) GetFont(text string) font.Face {
	// Simple heuristic: if text contains Thai characters, use Thai font
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F { // Thai Unicode block
			return fm.ThaiFont
		}
	}
	return fm.EnglishFont
}

// RenderTextAreasOverlay renders text areas with colored bounding boxes
func RenderTextAreasOverlay(pageData *page.Page, fontManager *FontManager) error {
	if pageData.Image == nil {
		return fmt.Errorf("no image in page data")
	}

	// Create a copy of the original image
	bounds := pageData.Image.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, pageData.Image, bounds.Min, draw.Src)

	// Draw text area bounding boxes
	for i, area := range pageData.TextAreas {
		areaColor := getAreaColor(i)
		drawRectangle(img, area.X, area.Y, area.Width, area.Height, areaColor, 2)

		// Draw area label
		label := fmt.Sprintf("Area %d", i+1)
		drawText(img, label, area.X, area.Y-2, fontManager.EnglishFont, areaColor)
	}

	// Generate random filename
	randomID := *gut.Random("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	filename := fmt.Sprintf("generate/recognize/output_areas_%s.png", randomID)

	return saveImage(img, filename)
}

// RenderLinesOverlay renders text lines with colored bounding boxes
func RenderLinesOverlay(pageData *page.Page, fontManager *FontManager) error {
	if pageData.Image == nil {
		return fmt.Errorf("no image in page data")
	}

	// Create a copy of the original image
	bounds := pageData.Image.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, pageData.Image, bounds.Min, draw.Src)

	// Draw line bounding boxes
	for i, line := range pageData.Lines {
		lineColor := getLineColor(i)
		drawRectangle(img, line.X, line.Y, line.Width, line.Height, lineColor, 1)

		// Draw line number
		label := fmt.Sprintf("L%d", i+1)
		drawText(img, label, line.X, line.Y-2, fontManager.EnglishFont, lineColor)
	}

	// Generate random filename
	randomID := *gut.Random("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	filename := fmt.Sprintf("generate/recognize/output_lines_%s.png", randomID)

	return saveImage(img, filename)
}

// RenderWordsOverlay renders words with colored bounding boxes
func RenderWordsOverlay(pageData *page.Page, fontManager *FontManager) error {
	if pageData.Image == nil {
		return fmt.Errorf("no image in page data")
	}

	// Create a copy of the original image
	bounds := pageData.Image.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, pageData.Image, bounds.Min, draw.Src)

	// Draw word bounding boxes
	for i, word := range pageData.Words {
		wordColor := getWordColor(i)
		drawRectangle(img, word.X, word.Y, word.Width, word.Height, wordColor, 1)

		// Draw word text above the box if recognized
		if word.Text != "" {
			textFont := fontManager.GetFont(word.Text)
			drawText(img, word.Text, word.X, word.Y-2, textFont, wordColor)
		}
	}

	// Generate random filename
	randomID := *gut.Random("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	filename := fmt.Sprintf("generate/recognize/output_words_%s.png", randomID)

	return saveImage(img, filename)
}

// RenderCharactersOverlay renders individual characters with bounding boxes and recognized text
func RenderCharactersOverlay(pageData *page.Page, fontManager *FontManager) error {
	if pageData.Image == nil {
		return fmt.Errorf("no image in page data")
	}

	// Create a copy of the original image
	bounds := pageData.Image.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, pageData.Image, bounds.Min, draw.Src)

	// Draw character bounding boxes
	for idx, char := range pageData.Chars {
		charColor := getCharColor(idx)
		drawRectangle(img, char.X, char.Y, char.Width, char.Height, charColor, 1)

		// Draw recognized character above the box
		if char.Text != "" {
			textFont := fontManager.GetFont(char.Text)
			// Draw with background for better visibility
			bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}
			drawTextWithBackground(img, char.Text, char.X, char.Y-2, textFont, charColor, bgColor)
		}

		// Draw confidence score below if available
		if char.Confidence > 0 {
			confidence := fmt.Sprintf("%.0f%%", char.Confidence)
			drawText(img, confidence, char.X, char.Y+char.Height+10, fontManager.EnglishFont, charColor)
		}
	}

	// Generate random filename
	randomID := *gut.Random("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	filename := fmt.Sprintf("generate/recognize/output_chars_%s.png", randomID)

	return saveImage(img, filename)
}

// RenderFullOverlay renders a comprehensive overlay with all elements
func RenderFullOverlay(pageData *page.Page, fontManager *FontManager) error {
	if pageData.Image == nil {
		return fmt.Errorf("no image in page data")
	}

	// Create a copy of the original image
	bounds := pageData.Image.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, pageData.Image, bounds.Min, draw.Src)

	// Draw text areas (thick blue boxes)
	for i, area := range pageData.TextAreas {
		drawRectangle(img, area.X, area.Y, area.Width, area.Height, color.RGBA{0, 100, 255, 255}, 3)
		label := fmt.Sprintf("Area %d", i+1)
		drawTextWithBackground(img, label, area.X, area.Y-15, fontManager.EnglishFont,
			color.RGBA{0, 100, 255, 255}, color.RGBA{255, 255, 255, 200})
	}

	// Draw lines (medium green boxes)
	for _, line := range pageData.Lines {
		drawRectangle(img, line.X, line.Y, line.Width, line.Height, color.RGBA{0, 200, 100, 255}, 2)
	}

	// Draw words (thin red boxes)
	for _, word := range pageData.Words {
		if word.Text != "" {
			drawRectangle(img, word.X, word.Y, word.Width, word.Height, color.RGBA{255, 100, 0, 255}, 1)
		}
	}

	// Draw recognized text
	for _, line := range pageData.Lines {
		if line.Text != "" {
			textFont := fontManager.GetFont(line.Text)
			// Draw recognized text above the line
			drawTextWithBackground(img, line.Text, line.X, line.Y-5, textFont,
				color.RGBA{255, 0, 0, 255}, color.RGBA{255, 255, 255, 180})
		}
	}

	// Generate random filename
	randomID := *gut.Random("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	filename := fmt.Sprintf("generate/recognize/output_full_%s.png", randomID)

	return saveImage(img, filename)
}

// Helper functions

func loadFont(fontPath string, size float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %v", err)
	}

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}

	return face, nil
}

func drawRectangle(img *image.RGBA, x, y, width, height int, col color.RGBA, thickness int) {
	// Draw top and bottom lines
	for t := 0; t < thickness; t++ {
		for i := x; i < x+width; i++ {
			if y+t >= 0 && y+t < img.Bounds().Dy() && i >= 0 && i < img.Bounds().Dx() {
				img.Set(i, y+t, col)
			}
			if y+height-t-1 >= 0 && y+height-t-1 < img.Bounds().Dy() && i >= 0 && i < img.Bounds().Dx() {
				img.Set(i, y+height-t-1, col)
			}
		}
	}

	// Draw left and right lines
	for t := 0; t < thickness; t++ {
		for j := y; j < y+height; j++ {
			if x+t >= 0 && x+t < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
				img.Set(x+t, j, col)
			}
			if x+width-t-1 >= 0 && x+width-t-1 < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
				img.Set(x+width-t-1, j, col)
			}
		}
	}
}

func drawText(img *image.RGBA, text string, x, y int, face font.Face, col color.Color) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(text)
}

func drawTextWithBackground(img *image.RGBA, text string, x, y int, face font.Face, textCol color.Color, bgCol color.RGBA) {
	// Get text bounds
	bounds, _ := font.BoundString(face, text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	// Draw background rectangle
	for dy := 0; dy < textHeight+4; dy++ {
		for dx := 0; dx < textWidth+4; dx++ {
			px := x + dx - 2
			py := y + dy - textHeight - 2
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				// Blend with existing color
				existing := img.RGBAAt(px, py)
				blended := blendColors(existing, bgCol)
				img.Set(px, py, blended)
			}
		}
	}

	// Draw text
	drawText(img, text, x, y, face, textCol)
}

func blendColors(base, overlay color.RGBA) color.RGBA {
	alpha := float64(overlay.A) / 255.0
	invAlpha := 1.0 - alpha

	return color.RGBA{
		R: uint8(float64(base.R)*invAlpha + float64(overlay.R)*alpha),
		G: uint8(float64(base.G)*invAlpha + float64(overlay.G)*alpha),
		B: uint8(float64(base.B)*invAlpha + float64(overlay.B)*alpha),
		A: 255,
	}
}

func saveImage(img image.Image, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %v", err)
	}

	fmt.Printf("Saved overlay image: %s\n", filename)
	return nil
}

// Color generation functions
func getAreaColor(index int) color.RGBA {
	colors := []color.RGBA{
		{0, 100, 255, 255}, // Blue
		{255, 100, 0, 255}, // Orange
		{0, 200, 100, 255}, // Green
		{200, 0, 200, 255}, // Magenta
		{255, 200, 0, 255}, // Yellow
	}
	return colors[index%len(colors)]
}

func getLineColor(index int) color.RGBA {
	colors := []color.RGBA{
		{0, 200, 100, 255},   // Green
		{100, 150, 255, 255}, // Light Blue
		{255, 150, 100, 255}, // Light Orange
		{150, 100, 255, 255}, // Purple
		{200, 200, 0, 255},   // Yellow-Green
	}
	return colors[index%len(colors)]
}

func getWordColor(index int) color.RGBA {
	colors := []color.RGBA{
		{255, 100, 0, 255},   // Orange
		{255, 0, 100, 255},   // Pink
		{100, 255, 0, 255},   // Lime
		{0, 255, 200, 255},   // Cyan
		{200, 100, 255, 255}, // Light Purple
	}
	return colors[index%len(colors)]
}

func getCharColor(index int) color.RGBA {
	colors := []color.RGBA{
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 255, 0, 255},   // Yellow
		{255, 0, 255, 255},   // Magenta
		{0, 255, 255, 255},   // Cyan
		{128, 128, 128, 255}, // Gray
		{255, 128, 0, 255},   // Orange
	}
	return colors[index%len(colors)]
}
