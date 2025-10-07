package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/bsthun/glyphcanvas/package/page"
	"github.com/bsthun/glyphcanvas/package/recognize"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <image_file>\n", os.Args[0])
		os.Exit(1)
	}

	imagePath := os.Args[1]
	databasePath := "generate/extract/char.yml"

	// Load character database
	fmt.Println("Loading character database...")
	database, err := recognize.LoadDatabase(databasePath)
	if err != nil {
		log.Fatal("Failed to load database:", err)
	}
	fmt.Printf("Loaded %d characters from database\n", len(database.Characters))

	// Load and process page image
	fmt.Printf("Processing page: %s\n", imagePath)
	pageData, err := processPage(imagePath, database)
	if err != nil {
		log.Fatal("Failed to process page:", err)
	}

	// Display results
	fmt.Printf("\n=== PAGE OCR RESULTS ===\n")
	fmt.Printf("Page dimensions: %dx%d\n", pageData.Width, pageData.Height)
	fmt.Printf("Found %d text areas\n", len(pageData.TextAreas))
	fmt.Printf("Found %d lines\n", len(pageData.Lines))
	fmt.Printf("Found %d words\n", len(pageData.Words))
	fmt.Printf("Found %d characters\n", len(pageData.Chars))

	fmt.Printf("\n=== EXTRACTED TEXT ===\n")
	text := pageData.GetPlainText()
	fmt.Println(text)

	fmt.Printf("\n=== DETAILED RESULTS ===\n")
	for i, area := range pageData.TextAreas {
		fmt.Printf("\nText Area %d: (%d,%d) %dx%d\n", i+1, area.X, area.Y, area.Width, area.Height)
		for j, line := range area.Lines {
			fmt.Printf("  Line %d: (%d,%d) %dx%d\n", j+1, line.X, line.Y, line.Width, line.Height)
			for k, word := range line.Words {
				avgConfidence := 0.0
				if len(word.Chars) > 0 {
					for _, char := range word.Chars {
						avgConfidence += char.Confidence
					}
					avgConfidence /= float64(len(word.Chars))
				}
				fmt.Printf("    Word %d: \"%s\" (%.1f%% confidence)\n", k+1, word.Text, avgConfidence)
			}
		}
	}
}

func processPage(imagePath string, database *recognize.FeatureDatabase) (*page.Page, error) {
	// Load image
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	// Create page
	pageData := page.NewPage(img)

	// Detect text structure
	fmt.Println("Detecting text areas...")
	err = pageData.DetectTextAreas()
	if err != nil {
		return nil, err
	}

	fmt.Println("Detecting text lines...")
	err = pageData.DetectLines()
	if err != nil {
		return nil, err
	}

	fmt.Println("Detecting words...")
	err = pageData.DetectWords()
	if err != nil {
		return nil, err
	}

	fmt.Println("Detecting characters...")
	err = pageData.DetectCharacters()
	if err != nil {
		return nil, err
	}

	// Recognize characters
	fmt.Println("Recognizing characters...")
	for i, char := range pageData.Chars {
		if i%50 == 0 {
			fmt.Printf("  Processed %d/%d characters\n", i, len(pageData.Chars))
		}

		if char.Character != nil {
			features, err := recognize.ExtractFeatures(char.Character)
			if err != nil {
				continue
			}

			candidates := recognize.RecognizeCharacter(features, database)
			if len(candidates) > 0 {
				best := candidates[0]
				char.Unicode = best.Unicode
				char.Text = unicodeToString(best.Unicode)
				char.Confidence = best.Confidence
			}
		}
	}

	// Build text from recognized characters
	for _, word := range pageData.Words {
		wordText := ""
		totalConfidence := 0.0
		validChars := 0

		for _, char := range word.Chars {
			if char.Text != "" {
				wordText += char.Text
				totalConfidence += char.Confidence
				validChars++
			}
		}

		word.Text = wordText
		if validChars > 0 {
			word.Confidence = totalConfidence / float64(validChars)
		}
	}

	// Build line text from words
	for _, line := range pageData.Lines {
		lineText := ""
		for i, word := range line.Words {
			if i > 0 && word.Text != "" {
				lineText += " "
			}
			lineText += word.Text
		}
		line.Text = lineText
	}

	fmt.Printf("Character recognition completed (%d characters processed)\n", len(pageData.Chars))

	return pageData, nil
}

func unicodeToString(unicode string) string {
	// Convert hex unicode to actual character
	if len(unicode) == 4 {
		if code, err := strconv.ParseInt(unicode, 16, 32); err == nil {
			return string(rune(code))
		}
	}
	return "?"
}
