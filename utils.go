package boardshapes

import (
	"image"
	"image/color"
	"math"
)

type SettableImage = interface {
	image.Image
	Set(x, y int, color color.Color)
}

func absDiff[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](a T, b T) T {
	if a > b {
		return a - b
	}
	return b - a
}

func DotProduct(x1, x2, y1, y2 float64) float64 {
	answer := (x1 * x2) + (y1 * y2)
	return answer
}

func GetNRGBA(c color.Color) color.NRGBA {
	var r, g, b, a uint32

	if nrgba, ok := c.(color.NRGBA); ok {
		// use non-alpha-premultiplied colors
		return nrgba
	}
	// use alpha-premultiplied colors
	r, g, b, a = c.RGBA()
	mult := 65535 / float64(a)
	// undo alpha-premultiplication
	r, g, b = uint32(float64(r)*mult), uint32(float64(g)*mult), uint32(float64(b)*mult)
	// reduce from 0-65535 to 0-255
	return color.NRGBA{uint8(r / 256), uint8(g / 256), uint8(b / 256), uint8(a / 256)}
}

func (v1 Vertex) DirectionTo(v2 Vertex) (x, y float64) {
	answerX := float64(v2.X - v1.X)
	answerY := float64(v2.Y - v1.Y)
	mag := math.Sqrt((answerX * answerX) + (answerY * answerY))
	return (answerX / mag), (answerY / mag)
}

func forNonDiagonalAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		function(x, y-1)
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		function(x, y+1)
	}
}

func forAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		if x > 0 {
			function(x-1, y-1)
		}
		function(x, y-1)
		if x < uint16(maxX)-1 {
			function(x+1, y-1)
		}
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		if x > 0 {
			function(x-1, y+1)
		}
		function(x, y+1)
		if x < uint16(maxX)-1 {
			function(x+1, y+1)
		}
	}
}
