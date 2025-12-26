package regions

import (
	"image"
	"slices"

	"github.com/boardshapes/boardshapes/colors"
	"github.com/boardshapes/boardshapes/utils/grid"
)

type RegionMap struct {
	regions []*Region
	pixels  [][]*Region
	options RegionMappingOptions
}

type RegionMappingOptions struct {
	NoColorSeparation,
	AllowWhite,
	PreserveColor bool
}

func BuildRegionMap(img image.Image, options RegionMappingOptions, regionFilter func(*Region) bool) *RegionMap {
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
				if c != colors.Blank && (allowWhite || c != colors.White) {
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
			grid.ForNonDiagonalAdjacents(cur.X, cur.Y, len(rm.pixels[cur.Y]), len(rm.pixels), func(x, y uint16) {
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
