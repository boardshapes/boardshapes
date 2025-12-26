package regions

import (
	"image"
	"image/color"
)

type Pixel struct {
	X, Y uint16
}

type Region []Pixel

func (r *Region) GetBounds() (regionBounds image.Rectangle) {
	regionBounds = image.Rectangle{Min: image.Pt(65535, 65535), Max: image.Pt(0, 0)}
	for _, pixel := range *r {
		if pixel.X < uint16(regionBounds.Min.X) {
			regionBounds.Min.X = int(pixel.X)
		}
		if pixel.Y < uint16(regionBounds.Min.Y) {
			regionBounds.Min.Y = int(pixel.Y)
		}
		if pixel.X+1 > uint16(regionBounds.Max.X) {
			regionBounds.Max.X = int(pixel.X) + 1
		}
		if pixel.Y+1 > uint16(regionBounds.Max.Y) {
			regionBounds.Max.Y = int(pixel.Y) + 1
		}
	}
	return
}

func (r *Region) FindRegionPosition() (int, int) {
	corner := (*r)[0]

	for i := 0; i < len(*r); i++ {
		if (*r)[i].X < corner.X {
			corner.X = (*r)[i].X
		}
		if (*r)[i].Y < corner.Y {
			corner.Y = (*r)[i].Y
		}
	}

	return int(corner.X), int(corner.Y)
}

func (r *Region) GetColorOfRegion(img image.Image, checkAll bool) color.Color {
	if checkAll {
		colorCounts := make(map[color.Color]uint, 1)
		for _, v := range *r {
			colorCounts[img.At(int(v.X), int(v.Y))]++
		}
		var mostCommonColor color.Color
		var mostCommonColorCount uint = 0
		for k, v := range colorCounts {
			if v > mostCommonColorCount {
				mostCommonColorCount = v
				mostCommonColor = k
			}
		}
		return mostCommonColor
	} else {
		regionColor := img.At(int((*r)[0].X), int((*r)[0].Y))
		return regionColor
	}
}
