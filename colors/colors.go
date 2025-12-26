package colors

import (
	"image/color"
)

var Red color.NRGBA = color.NRGBA{uint8(255), uint8(0), uint8(0), uint8(255)}
var Green color.NRGBA = color.NRGBA{uint8(0), uint8(255), uint8(0), uint8(255)}
var Blue color.NRGBA = color.NRGBA{uint8(0), uint8(0), uint8(255), uint8(255)}
var White color.NRGBA = color.NRGBA{uint8(255), uint8(255), uint8(255), uint8(255)}
var Black color.NRGBA = color.NRGBA{uint8(0), uint8(0), uint8(0), uint8(255)}
var Blank color.NRGBA = color.NRGBA{uint8(0), uint8(0), uint8(0), uint8(0)}
var Cyan color.NRGBA = color.NRGBA{uint8(0), uint8(255), uint8(255), uint8(255)}
var Magenta color.NRGBA = color.NRGBA{uint8(255), uint8(0), uint8(255), uint8(255)}
var Yellow color.NRGBA = color.NRGBA{uint8(255), uint8(255), uint8(0), uint8(255)}

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
