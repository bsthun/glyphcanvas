package test

import (
	"image"
	"image/color"

	"github.com/bsthun/glyphcanvas/package/region"
)

func RegionFromImage(img image.Image) *region.Region {
	bounds := img.Bounds()
	width := uint16(bounds.Max.X - bounds.Min.X)
	height := uint16(bounds.Max.Y - bounds.Min.Y)

	r := region.NewRegion(width, height)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			gray, _, _, _ := color.GrayModel.Convert(c).RGBA()
			if gray > 32768 {
				r.Draw(uint16(x-bounds.Min.X), uint16(y-bounds.Min.Y))
			}
		}
	}

	return r
}
