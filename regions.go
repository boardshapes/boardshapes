package boardshapes

import (
	"image"
	"image/color"
	"slices"
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

func BuildRegionMap(img image.Image, options ShapeCreationOptions, regionFilter func(*Region) bool) *RegionMap {
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	regionMap := RegionMap{make([]*Region, 0, 20), make([][]*Region, dy), options}
	for i := range regionMap.pixels {
		regionMap.pixels[i] = make([]*Region, dx)
	}

	bd := img.Bounds()

	allowWhite := regionMap.options.AllowWhite
	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			pixel := Pixel{uint16(x), uint16(y)}
			if !regionMap.GetPixelHasRegion(pixel) {
				c := img.At(x, y)
				if c != Blank && (allowWhite || c != White) {
					regionMap.AddPixelToRegionMap(pixel, img)
				}
			}
		}
	}

	if regionFilter != nil {
		regionMap.FilterRegions(regionFilter)
	}

	return &regionMap
}

type Pixel struct {
	X, Y uint16
}

type Vertex struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

type Region []Pixel

type RegionMap struct {
	regions []*Region
	pixels  [][]*Region
	options ShapeCreationOptions
}

func (rm *RegionMap) NewRegion(pixel Pixel) (region *Region) {
	region = &Region{pixel}
	rm.regions = append(rm.regions, region)
	rm.pixels[pixel.Y][pixel.X] = region
	return
}

func (rm *RegionMap) AddPixelToRegion(pixel Pixel, region *Region) {
	*region = append(*region, pixel)
	rm.pixels[pixel.Y][pixel.X] = region
}

func (rm *RegionMap) AddPixelToRegionMap(pixel Pixel, img image.Image) {
	regionColor := img.At(int(pixel.X), int(pixel.Y))

	if !rm.GetPixelHasRegion(pixel) {
		region := rm.NewRegion(pixel)

		// iterative depth first traversal
		pixelsToVisit := make([]Pixel, 1, 8)
		pixelsToVisit[0] = pixel
		for len(pixelsToVisit) > 0 {
			cur := pixelsToVisit[len(pixelsToVisit)-1]
			pixelsToVisit = pixelsToVisit[:len(pixelsToVisit)-1]
			forNonDiagonalAdjacents(cur.X, cur.Y, len(rm.pixels[cur.Y]), len(rm.pixels), func(x, y uint16) {
				p := Pixel{x, y}
				if !rm.GetPixelHasRegion(p) && img.At(int(x), int(y)) == regionColor {
					rm.AddPixelToRegion(p, region)
					pixelsToVisit = append(pixelsToVisit, p)
				}
			})
		}
	}
}

func (rm *RegionMap) GetPixelHasRegion(pixel Pixel) (hasRegion bool) {
	return rm.pixels[pixel.Y][pixel.X] != nil
}

func (rm *RegionMap) GetRegionOfPixel(pixel Pixel) (region *Region) {
	return rm.pixels[pixel.Y][pixel.X]
}

func (rm *RegionMap) GetRegionByIndex(i int) *Region {
	return rm.regions[i]
}

func (rm *RegionMap) GetRegions() []*Region {
	return rm.regions
}

func (rm *RegionMap) FilterRegions(regionFilter func(*Region) bool) {
	for i, region := range rm.regions {
		if region != nil && !regionFilter(region) {
			rm.regions[i] = nil
		}
	}
	rm.cleanupRegions()
}

func (rm *RegionMap) cleanupRegions() {
	rm.regions = slices.DeleteFunc(rm.regions, func(r *Region) bool { return r == nil })
}

func (re *Region) GetBounds() (regionBounds image.Rectangle) {
	regionBounds = image.Rectangle{Min: image.Pt(65535, 65535), Max: image.Pt(0, 0)}
	for _, pixel := range *re {
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

func ColorRegionEquivalence(a color.Color, b color.Color) bool {
	return a == b
}

func FindRegionPosition(region *Region) (int, int) {
	corner := (*region)[0]

	for i := 0; i < len(*region); i++ {
		if (*region)[i].X < corner.X {
			corner.X = (*region)[i].X
		}
		if (*region)[i].Y < corner.Y {
			corner.Y = (*region)[i].Y
		}
	}

	return int(corner.X), int(corner.Y)
}

func GetColorOfRegion(region *Region, img image.Image, checkAll bool) color.Color {
	if checkAll {
		colorCounts := make(map[color.Color]uint, 1)
		for _, v := range *region {
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
		regionColor := img.At(int((*region)[0].X), int((*region)[0].Y))
		return regionColor
	}
}
