package imageops

import (
	"image"
	"math"

	"golang.org/x/image/draw"
)

// Resizes the image to the default 1920x1080. Uses [ResizeImageTo].
func ResizeImage(img image.Image) image.Image {
	const MAX_HEIGHT = 1080
	const MAX_WIDTH = 1920

	return ResizeImageTo(img, MAX_WIDTH, MAX_HEIGHT)
}

// Constrains the image to the given dimensions, preserving aspect ratio.
// If either dimension is set to 0 or less, it will be ignored (effectively like if you set it to infinity).
func ResizeImageTo(img image.Image, width, height int) image.Image {
	bd := img.Bounds()
	if (width <= 0 && height <= 0) || (width >= bd.Dx() && height >= bd.Dy()) {
		width, height = bd.Dx(), bd.Dy()
	} else if width <= 0 {
		wScalar := float64(height) / float64(bd.Dy())
		width = int(math.Round(float64(bd.Dx()) * wScalar))
	} else if height <= 0 {
		hScalar := float64(width) / float64(bd.Dx())
		height = int(math.Round(float64(bd.Dx()) * hScalar))
	} else {
		wScalar := float64(height) / float64(bd.Dy())
		hScalar := float64(width) / float64(bd.Dx())
		scalar := math.Min(wScalar, hScalar)
		width = int(math.Round(float64(bd.Dx()) * scalar))
		height = int(math.Round(float64(bd.Dy()) * scalar))
	}

	scaledImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.NearestNeighbor.Scale(scaledImg, scaledImg.Rect, img, img.Bounds(), draw.Over, nil)
	return scaledImg
}
