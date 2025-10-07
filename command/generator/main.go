package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Character sets
var (
	englishLowercase = "abcdefghijklmnopqrstuvwxyz"
	englishUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	englishDigits    = "0123456789"
	thaiConsonants   = "กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮ"
	thaiVowels       = "ะาำิีึืุูเแโใไ็ฺ่้๊๋"
	thaiNumbers      = "๐๑๒๓๔๕๖๗๘๙"
	thaiSpecial      = "ฯๆ"
)

// CharacterInfo holds information about a character
type CharacterInfo struct {
	Character string
	Name      string
	Unicode   rune
	Category  string
}

func main() {
	fmt.Println("Starting character dataset generation...")

	outputDir := "generate/dataset/singlecharacter"

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Load fonts for different character sets
	thaiFontPath := "generate/font/NotoSansThaiLooped-Regular.ttf"
	thaiFontFace, err := loadFont(thaiFontPath, 32)
	if err != nil {
		fmt.Printf("Warning: Failed to load Thai font %s: %v\n", thaiFontPath, err)
		fmt.Println("Will use basic font for Thai characters...")
		thaiFontFace = basicfont.Face7x13
	} else {
		fmt.Println("Loaded Noto Sans Thai for Thai characters")
	}

	englishFontPath := "generate/font/Roboto-Regular.ttf"
	englishFontFace, err := loadFont(englishFontPath, 32)
	if err != nil {
		fmt.Printf("Warning: Failed to load English font %s: %v\n", englishFontPath, err)
		fmt.Println("Will use basic font for English characters...")
		englishFontFace = basicfont.Face7x13
	} else {
		fmt.Println("Loaded Roboto for English characters")
	}

	// Create font map for different character categories
	fontMap := map[string]font.Face{
		"english": englishFontFace,
		"thai":    thaiFontFace,
	}

	var characters []CharacterInfo

	// Add English lowercase
	for _, ch := range englishLowercase {
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      string(ch),
			Unicode:   ch,
			Category:  "english_lowercase",
		})
	}

	// Add English uppercase
	for _, ch := range englishUppercase {
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      string(ch),
			Unicode:   ch,
			Category:  "english_uppercase",
		})
	}

	// Add English digits
	for _, ch := range englishDigits {
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      fmt.Sprintf("%s", string(ch)),
			Unicode:   ch,
			Category:  "english_digits",
		})
	}

	// Add Thai consonants
	for _, ch := range thaiConsonants {
		name := fmt.Sprintf("%04x", ch)
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      name,
			Unicode:   ch,
			Category:  "thai_consonants",
		})
	}

	// Add Thai vowels
	for _, ch := range thaiVowels {
		name := fmt.Sprintf("%04x", ch)
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      name,
			Unicode:   ch,
			Category:  "thai_vowels",
		})
	}

	// Add Thai numbers
	for _, ch := range thaiNumbers {
		name := fmt.Sprintf("%04x", ch)
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      name,
			Unicode:   ch,
			Category:  "thai_numbers",
		})
	}

	// Add Thai special characters
	for _, ch := range thaiSpecial {
		name := fmt.Sprintf("%04x", ch)
		characters = append(characters, CharacterInfo{
			Character: string(ch),
			Name:      name,
			Unicode:   ch,
			Category:  "thai_special",
		})
	}

	fmt.Printf("Generating %d character images...\n", len(characters))

	// Generate images for each character
	generated := 0
	failed := 0

	for i, charInfo := range characters {
		var filename string
		if charInfo.Category == "english_lowercase" {
			filename = fmt.Sprintf("char_en_lower_%s.png", charInfo.Name)
		} else if charInfo.Category == "english_uppercase" {
			filename = fmt.Sprintf("char_en_upper_%s.png", charInfo.Name)
		} else if charInfo.Category == "english_digits" {
			filename = fmt.Sprintf("char_%s.png", charInfo.Name)
		} else if strings.HasPrefix(charInfo.Category, "thai") {
			filename = fmt.Sprintf("char_th_%s.png", charInfo.Name)
		} else {
			filename = fmt.Sprintf("char_%s.png", charInfo.Name)
		}
		outputPath := filepath.Join(outputDir, filename)

		err := generateCharacterImage(charInfo, outputPath, fontMap)
		if err != nil {
			fmt.Printf("Failed to generate %s (%s): %v\n", charInfo.Character, charInfo.Name, err)
			failed++
		} else {
			generated++
		}

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == len(characters) {
			fmt.Printf("Progress: %d/%d (generated: %d, failed: %d)\n",
				i+1, len(characters), generated, failed)
		}
	}

	fmt.Printf("\nCharacter dataset generation complete!\n")
	fmt.Printf("Generated: %d images\n", generated)
	fmt.Printf("Failed: %d images\n", failed)
	fmt.Printf("Output directory: %s\n", outputDir)
}

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

func generateCharacterImage(charInfo CharacterInfo, outputPath string, fontMap map[string]font.Face) error {
	const (
		maxSize = 64
		padding = 8
	)

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, maxSize, maxSize))

	// Fill with white background
	for y := 0; y < maxSize; y++ {
		for x := 0; x < maxSize; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// Select appropriate font based on character type
	var face font.Face
	if strings.HasPrefix(charInfo.Category, "english") {
		// Use Roboto font for English characters
		face = fontMap["english"]
	} else if strings.HasPrefix(charInfo.Category, "thai") {
		// Use Noto font for Thai characters, fallback to basic font if character not found
		if characterExistsInFont(fontMap["thai"], charInfo.Character) {
			face = fontMap["thai"]
		} else {
			face = basicfont.Face7x13
		}
	} else {
		// Default to English font for other characters
		face = fontMap["english"]
	}

	// Calculate text position to center it
	bounds, _ := font.BoundString(face, charInfo.Character)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	// Center the text
	x := (maxSize - textWidth) / 2
	y := (maxSize + textHeight) / 2

	// Ensure the character fits within bounds
	if textWidth > maxSize-2*padding || textHeight > maxSize-2*padding {
		// Scale down if needed
		scale := float64(maxSize-2*padding) / float64(maxInt(textWidth, textHeight))
		if scale < 1.0 {
			// For now, just center it as best as we can
			x = padding
			y = maxSize - padding
		}
	}

	// Draw the character
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}), // Black text
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	drawer.DrawString(charInfo.Character)

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Encode as PNG
	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %v", err)
	}

	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func characterExistsInFont(face font.Face, char string) bool {
	// Check if character exists by trying to get glyph bounds
	bounds, advance := font.BoundString(face, char)

	// If both bounds and advance are zero, character doesn't exist
	if bounds.Max.X == 0 && bounds.Max.Y == 0 && advance == 0 {
		return false
	}

	// Check if the character has reasonable dimensions
	width := (bounds.Max.X - bounds.Min.X).Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()

	// Character should have some width or height, or at least some advance
	if width < 1 && height < 1 && advance == 0 {
		return false
	}

	return true
}
