package boardshapes

import (
	"cmp"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"slices"

	"golang.org/x/image/draw"
)

const VERSION = "0.1.0"

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

// func manhattanDistance(a Vertex, b Vertex) int {
// 	return absDiff(int(a.X), int(b.X)) + absDiff(int(a.Y), int(b.Y))
// }

type RegionPixel byte

const (
	REGION_PIXEL_IN_REGION = 0b00000001
	REGION_PIXEL_VISITED   = 0b00000010
	REGION_PIXEL_IS_OUTER  = 0b00000100
	REGION_PIXEL_IS_INNER  = 0b00001000
)

func (r *RegionPixel) MarkInRegion() {
	*r = *r | REGION_PIXEL_IN_REGION
}

func (r *RegionPixel) MarkVisited() {
	*r = *r | REGION_PIXEL_VISITED
}

func (r *RegionPixel) MarkIsOuter() {
	*r = *r | REGION_PIXEL_IS_OUTER
}

func (r *RegionPixel) MarkIsInner() {
	*r = *r | REGION_PIXEL_IS_INNER
}

func (r RegionPixel) InRegion() bool {
	return r&REGION_PIXEL_IN_REGION > 0
}

func (r RegionPixel) Visited() bool {
	return r&REGION_PIXEL_VISITED > 0
}

func (r RegionPixel) IsOuter() bool {
	return r&REGION_PIXEL_IS_OUTER > 0
}

func (r RegionPixel) IsInner() bool {
	return r&REGION_PIXEL_IS_INNER > 0
}

func (r RegionPixel) String() string {
	return fmt.Sprintf("in region: %t; visited: %t; in shape: %t", r.InRegion(), r.Visited(), r.IsOuter())
}

func (region *Region) CreateShape() (shape []Vertex, err error) {
	if len(*region) == 0 {
		return nil, errors.New("region-to-shape: region is empty")
	}
	regionBounds := region.GetBounds()

	// will sastisfy my requirements.
	regionPixels := make([][]RegionPixel, regionBounds.Dx()+2)
	for i := range regionPixels {
		regionPixels[i] = make([]RegionPixel, regionBounds.Dy()+2)
	}

	for _, v := range *region {
		regionPixels[int(v.X+1)-regionBounds.Min.X][int(v.Y+1)-regionBounds.Min.Y].MarkInRegion()
	}

	verticesToVisit := []Vertex{{0, 0}}
	// visit outer pixels
	for len(verticesToVisit) > 0 {
		v := verticesToVisit[len(verticesToVisit)-1]
		verticesToVisit = verticesToVisit[:len(verticesToVisit)-1]
		if !regionPixels[v.X][v.Y].Visited() {
			regionPixels[v.X][v.Y].MarkVisited()
			forNonDiagonalAdjacents(v.X, v.Y, len(regionPixels), len(regionPixels[0]), func(x, y uint16) {
				if !regionPixels[x][y].Visited() && !regionPixels[x][y].IsOuter() {
					if regionPixels[x][y].InRegion() {
						regionPixels[x][y].MarkIsOuter()
					} else {
						verticesToVisit = append(verticesToVisit, Vertex{x, y})
					}
				}
			})

		}
	}

	vertexShapes := make([][]Vertex, 0, 1)

	// find inner pixels
	for y := uint16(0); y < uint16(len(regionPixels[0])); y++ {
		for x := uint16(0); x < uint16(len(regionPixels)); x++ {
			rp := regionPixels[x][y]
			// check if inner pixel
			if !rp.Visited() && !rp.IsOuter() {
				verticesToVisit := []Vertex{{x, y}}
				newInnerShape := make([]Vertex, 0, regionBounds.Dx()+regionBounds.Dy())
				// visit inner pixels
				for len(verticesToVisit) > 0 {
					v := verticesToVisit[len(verticesToVisit)-1]
					verticesToVisit = verticesToVisit[:len(verticesToVisit)-1]
					if !regionPixels[v.X][v.Y].Visited() {
						regionPixels[v.X][v.Y].MarkVisited()
						forNonDiagonalAdjacents(v.X, v.Y, len(regionPixels), len(regionPixels[0]), func(x, y uint16) {
							if !regionPixels[x][y].Visited() && !regionPixels[x][y].IsInner() {
								if regionPixels[x][y].IsOuter() {
									regionPixels[x][y].MarkIsInner()
									newInnerShape = append(newInnerShape, Vertex{x, y})
								} else {
									verticesToVisit = append(verticesToVisit, Vertex{x, y})
								}
							}
						})

					}
				}
				vertexShapes = append(vertexShapes, newInnerShape)
			}
		}
	}

	vertexMatrix := make([][]bool, regionBounds.Dx())
	for i := range vertexMatrix {
		vertexMatrix[i] = make([]bool, regionBounds.Dy())
	}

	if len(vertexShapes) == 0 {
		return nil, errors.New("region-to-shape: region is too thin")
	}

	vertexShape := slices.MaxFunc(vertexShapes, func(a, b []Vertex) int {
		return cmp.Compare(len(a), len(b))
	})

	// translate all vertices by (-1, -1)
	// necessary because we added extra space for the region up above
	for i, v := range vertexShape {
		vertexShape[i].X--
		vertexShape[i].Y--
		vertexMatrix[v.X-1][v.Y-1] = true
	}

	var previousVertex Vertex
	var isPreviousVertexSet = false
	var currentVertex Vertex = vertexShape[0]
	sortedOuterVertexShape := make([]Vertex, 0, len(vertexShape))

	for {
		adjacentVertices := make([]Vertex, 0, 8)

		forAdjacents(currentVertex.X, currentVertex.Y, len(vertexMatrix), len(vertexMatrix[0]), func(x, y uint16) {
			if vertexMatrix[x][y] {
				adjacentVertices = append(adjacentVertices, Vertex{uint16(x), uint16(y)})
			}
		})

		if len(adjacentVertices) != 2 {
			return nil, errors.New("region-to-shape: shape generation failed")
		}

		if !isPreviousVertexSet {
			isPreviousVertexSet = true
			previousVertex = adjacentVertices[0]
			sortedOuterVertexShape = append(sortedOuterVertexShape, previousVertex)
		}

		sortedOuterVertexShape = append(sortedOuterVertexShape, currentVertex)

		if adjacentVertices[0] == previousVertex {
			previousVertex = currentVertex
			currentVertex = adjacentVertices[1]
		} else {
			previousVertex = currentVertex
			currentVertex = adjacentVertices[0]
		}

		if currentVertex == sortedOuterVertexShape[0] {
			return sortedOuterVertexShape, nil
		}

		if len(sortedOuterVertexShape) >= len(vertexShape) {
			return nil, errors.New("region-to-shape: could not close shape")
		}
	}
}

func StraightOpt(sortedVertexShape []Vertex) []Vertex {
	for i := 2; i < len(sortedVertexShape); i++ {
		x1, y1 := sortedVertexShape[i-2].DirectionTo(sortedVertexShape[i-1])
		x2, y2 := sortedVertexShape[i-1].DirectionTo(sortedVertexShape[i])
		if x1 == x2 && y1 == y2 {
			sortedVertexShape = append(sortedVertexShape[:i-1], sortedVertexShape[i:]...)
			i--
		}
	}
	return sortedVertexShape
}

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

func SimplifyImage(img image.Image, options ShapeCreationOptions) (result image.Image) {
	bd := img.Bounds()
	var newImg *image.Paletted
	if options.AllowWhite {
		newImg = image.NewPaletted(bd, color.Palette{Blank, White, Black, Red, Green, Blue})
	} else {
		newImg = image.NewPaletted(bd, color.Palette{White, Black, Red, Green, Blue})
	}

	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			c := GetNRGBA(img.At(x, y))
			r, g, b, a := int(c.R), int(c.G), int(c.B), int(c.A)
			var newPixelColor color.NRGBA
			avg := (r + g + b) / 3
			if a < 10 {
				if options.AllowWhite {
					newPixelColor = Blank
				} else {
					newPixelColor = White
				}
			} else if max(absDiff(avg, r), absDiff(avg, g), absDiff(avg, b)) < 10 {
				// todo: better way to detect black maybe
				if max(r, g, b) > 115 {
					newPixelColor = White
				} else {
					newPixelColor = Black
				}
			} else if r > g && r > b {
				newPixelColor = Red
			} else if g > r && (g > b || b-g < 10) {
				newPixelColor = Green
			} else if b > r && b > g {
				newPixelColor = Blue
			} else {
				newPixelColor = White
			}
			newImg.Set(x, y, newPixelColor)
		}
	}

	return newImg
}

type BoardshapesData struct {
	Version string
	Shapes  []ShapeData
}

func (bd BoardshapesData) Equal(other BoardshapesData) (equal bool, reason string) {
	if bd.Version != other.Version {
		return false, "version mismatch"
	}
	if len(bd.Shapes) != len(other.Shapes) {
		return false, "shape count mismatch"
	}
outer:
	for _, outShape := range other.Shapes {
		for _, inShape := range bd.Shapes {
			if inShape.Equal(outShape) {
				continue outer
			}
		}
		return false, fmt.Sprintf("shape has no matching shape: %d", outShape.Number)
	}
	return true, ""
}

type ShapeData struct {
	Number    int
	Color     color.Color
	ColorName string
	CornerX   int
	CornerY   int
	Image     image.Image
	Path      []Vertex
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
				img1, err := os.Create("bad1.png")
				if err != nil {
					panic(err)
				}
				defer img1.Close()
				png.Encode(img1, sd.Image)
				img2, err := os.Create("bad2.png")
				if err != nil {
					panic(err)
				}
				defer img2.Close()
				png.Encode(img2, other.Image)
				return false
			}
		}
	}
	return slices.Equal(sd.Path, other.Path)
}

type ShapeCreationOptions struct {
	NoColorSeparation, AllowWhite, PreserveColor bool
}

func CreateShapes(img image.Image, opts ShapeCreationOptions) (data *BoardshapesData) {
	data = &BoardshapesData{
		Version: VERSION,
	}

	img = ResizeImage(img)

	newImg := SimplifyImage(img, opts)

	regionMap := buildRegionMapWithoutSmallRegions(newImg, opts)

	regions := regionMap.GetRegions()
	numRegions := len(regions)

	data.Shapes = make([]ShapeData, 0, numRegions)

	for i := range numRegions {
		region := regionMap.GetRegionByIndex(i)

		minX, minY := FindRegionPosition(region)
		regionColor := GetColorOfRegion(region, newImg, opts.NoColorSeparation)
		var regionColorName string

		switch regionColor {
		case Red:
			regionColorName = "Red"
		case Green:
			regionColorName = "Green"
		case Blue:
			regionColorName = "Blue"
		case Black:
			regionColorName = "Black"
		case White:
			regionColorName = "White"
		}

		regionImage := image.NewNRGBA(region.GetBounds())

		if opts.PreserveColor {
			for j := 0; j < len(*region); j++ {
				regionImage.Set(int((*region)[j].X), int((*region)[j].Y), img.At(int((*region)[j].X), int((*region)[j].Y)))
			}
		} else {
			for j := 0; j < len(*region); j++ {
				regionImage.Set(int((*region)[j].X), int((*region)[j].Y), regionColor)
			}
		}

		shape, err := region.CreateShape()
		if err != nil {
			continue
		}
		optimizedShape := StraightOpt(shape)

		shapeData := ShapeData{
			Number:    i,
			Color:     regionColor,
			ColorName: regionColorName,
			CornerX:   minX,
			CornerY:   minY,
			Image:     regionImage,
			Path:      optimizedShape,
		}

		data.Shapes = append(data.Shapes, shapeData)
	}

	return
}

type SettableImage = interface {
	image.Image
	Set(x, y int, color color.Color)
}

const MINIMUM_NUMBER_OF_PIXELS_FOR_NON_SMALL_REGION = 50

// this is awful
func buildRegionMapWithoutSmallRegions(img image.Image, options ShapeCreationOptions) (regionMap *RegionMap) {
	var removedColor color.Color
	if options.AllowWhite {
		removedColor = Blank
	} else {
		removedColor = White
	}

	regionMap = BuildRegionMap(img, options, func(r *Region) bool {
		if len(*r) >= MINIMUM_NUMBER_OF_PIXELS_FOR_NON_SMALL_REGION {
			return true
		}
		if i, ok := img.(SettableImage); ok {
			for _, pixel := range *r {
				i.Set(int(pixel.X), int(pixel.Y), removedColor)
			}
		}
		return false
	})

	return regionMap
}
