package test

import (
	"image"
	"image/png"
	"os"
	"testing"
)

func LoadImage(t *testing.T, path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open test image: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode PNG image: %v", err)
	}

	return img
}
