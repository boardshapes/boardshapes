package data

import (
	"image"
	"image/color"
	"slices"

	"github.com/boardshapes/boardshapes/geometry"
)

type ShapeData struct {
	Number    int
	Color     color.Color
	ColorName string
	CornerX   int
	CornerY   int
	Image     image.Image
	Path      []geometry.Vertex
}

func (sd ShapeData) Equal(other ShapeData) bool {
	if sd.Number != other.Number ||
		sd.Color != other.Color ||
		sd.ColorName != other.ColorName ||
		sd.CornerX != other.CornerX || sd.CornerY != other.CornerY {
		return false
	}
	aBds, bBds := sd.Image.Bounds(), other.Image.Bounds()
	width, height := aBds.Dx(), aBds.Dy()
	if width != bBds.Dx() {
		return false
	}
	if height != bBds.Dy() {
		return false
	}
	for y := range height {
		for x := range width {
			if sd.Image.At(aBds.Min.X+x, aBds.Min.Y+y) != other.Image.At(bBds.Min.X+x, bBds.Min.Y+y) {
				return false
			}
		}
	}
	return slices.Equal(sd.Path, other.Path)
}
