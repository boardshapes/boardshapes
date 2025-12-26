package imageops

import (
	"image"
	"image/color"
)

type SettableImage = interface {
	image.Image
	Set(x, y int, color color.Color)
}
