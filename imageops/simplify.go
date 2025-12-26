package imageops

import (
	"image"
	"image/color"

	"github.com/boardshapes/boardshapes/colors"
	"github.com/boardshapes/boardshapes/utils"
)

type ImageSimplificationOptions struct {
	AllowWhite bool
}

func SimplifyImage(img image.Image, options ImageSimplificationOptions) (result image.Image) {
	bd := img.Bounds()
	var newImg *image.Paletted
	if options.AllowWhite {
		newImg = image.NewPaletted(bd, color.Palette{colors.Blank, colors.White, colors.Black, colors.Red, colors.Green, colors.Blue})
	} else {
		newImg = image.NewPaletted(bd, color.Palette{colors.White, colors.Black, colors.Red, colors.Green, colors.Blue})
	}

	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			c := colors.GetNRGBA(img.At(x, y))
			r, g, b, a := int(c.R), int(c.G), int(c.B), int(c.A)
			var newPixelColor color.NRGBA
			avg := (r + g + b) / 3
			if a < 10 {
				if options.AllowWhite {
					newPixelColor = colors.Blank
				} else {
					newPixelColor = colors.White
				}
			} else if max(utils.AbsDiff(avg, r), utils.AbsDiff(avg, g), utils.AbsDiff(avg, b)) < 10 {
				// todo: better way to detect black maybe
				if max(r, g, b) > 115 {
					newPixelColor = colors.White
				} else {
					newPixelColor = colors.Black
				}
			} else if r > g && r > b {
				newPixelColor = colors.Red
			} else if g > r && (g > b || b-g < 10) {
				newPixelColor = colors.Green
			} else if b > r && b > g {
				newPixelColor = colors.Blue
			} else {
				newPixelColor = colors.White
			}
			newImg.Set(x, y, newPixelColor)
		}
	}

	return newImg
}
