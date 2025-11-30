package boardshapes

import (
	"cmp"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"

	"golang.org/x/image/draw"
)

const VERSION = "0.1.1"

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
	regionPixels := createRegionPixelsMatrix(region, regionBounds)

	// find inner pixels to find shapes
	possibleShapeVertices := findShapes(regionPixels)

	if len(possibleShapeVertices) == 0 {
		return nil, errors.New("region-to-shape: region is too thin")
	}

	largestShapeVertices := slices.MaxFunc(possibleShapeVertices, func(a, b []Vertex) int {
		return cmp.Compare(len(a), len(b))
	})

	vertexMatrix := make([][]bool, regionBounds.Dx())
	for i := range vertexMatrix {
		vertexMatrix[i] = make([]bool, regionBounds.Dy())
	}

	// build matrix with all vertices translated by (-1, -1)
	// necessary because we added extra space for the region up above
	for _, v := range largestShapeVertices {
		vertexMatrix[v.X-1][v.Y-1] = true
	}

	return findSortedShapeVertices(
		Vertex{
			X: largestShapeVertices[0].X - 1,
			Y: largestShapeVertices[0].Y - 1,
		},
		vertexMatrix,
		len(largestShapeVertices))
}

func createRegionPixelsMatrix(region *Region, regionBounds image.Rectangle) (regionPixels [][]RegionPixel) {
	regionPixels = make([][]RegionPixel, regionBounds.Dx()+2)
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
			forNonDiagonalAdjacents(
				v.X, v.Y, len(regionPixels), len(regionPixels[0]),
				func(x, y uint16) {
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
	return regionPixels
}

func findShapes(regionPixels [][]RegionPixel) [][]Vertex {
	possibleShapeVertices := make([][]Vertex, 0, 1)
	for y := uint16(0); y < uint16(len(regionPixels[0])); y++ {
		for x := uint16(0); x < uint16(len(regionPixels)); x++ {
			rp := regionPixels[x][y]
			// check if inner pixel
			if !rp.Visited() && !rp.IsOuter() {
				verticesToVisit := []Vertex{{x, y}}
				newInnerShape := make([]Vertex, 0)
				// visit inner pixels
				for len(verticesToVisit) > 0 {
					v := verticesToVisit[len(verticesToVisit)-1]
					verticesToVisit = verticesToVisit[:len(verticesToVisit)-1]
					if regionPixels[v.X][v.Y].Visited() {
						continue
					}
					regionPixels[v.X][v.Y].MarkVisited()
					forNonDiagonalAdjacents(
						v.X, v.Y, len(regionPixels), len(regionPixels[0]),
						func(x, y uint16) {
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
				possibleShapeVertices = append(possibleShapeVertices, newInnerShape)
			}
		}
	}
	return possibleShapeVertices
}

func findSortedShapeVertices(startingVertex Vertex, vertexMatrix [][]bool, maxLength int) ([]Vertex, error) {
	var previousVertex Vertex
	var isPreviousVertexSet = false
	var currentVertex Vertex = startingVertex
	sortedShapeVertices := make([]Vertex, 0, maxLength)

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
			sortedShapeVertices = append(sortedShapeVertices, previousVertex)
		}

		sortedShapeVertices = append(sortedShapeVertices, currentVertex)

		if adjacentVertices[0] == previousVertex {
			previousVertex = currentVertex
			currentVertex = adjacentVertices[1]
		} else {
			previousVertex = currentVertex
			currentVertex = adjacentVertices[0]
		}

		if currentVertex == sortedShapeVertices[0] {
			return sortedShapeVertices, nil
		}

		if len(sortedShapeVertices) >= maxLength {
			return nil, errors.New("region-to-shape: could not close shape")
		}
	}
}

const DEFAULT_RDP_EPSILON = 10.0

func OptimizeShape(sortedVertexShape []Vertex) []Vertex {
	return OptimizeShapeWithEpsilon(sortedVertexShape, DEFAULT_RDP_EPSILON)
}

const MINIMUM_VERTICES_FOR_RDP = 15

func OptimizeShapeWithEpsilon(sortedVertexShape []Vertex, epsilon float64) []Vertex {
	//Try optimizing straight lines
	var optimizedShape []Vertex
	for i := 2; i < len(sortedVertexShape); i++ {
		x1, y1 := sortedVertexShape[i-2].DirectionTo(sortedVertexShape[i-1])
		x2, y2 := sortedVertexShape[i-1].DirectionTo(sortedVertexShape[i])
		if x1 == x2 && y1 == y2 {
			optimizedShape = append(sortedVertexShape[:i-1], sortedVertexShape[i:]...)
			i--
		}
	}

	//If epsilon is negative, skip RDP optimization
	if epsilon < 0 {
		return optimizedShape
	}

	//Check number of vertices after straight optimization to determine if RDP is needed
	if len(optimizedShape) > MINIMUM_VERTICES_FOR_RDP {
		//Split shape in half by finding furthest vertex from the startpoint
		distance := 0.0
		furthest := 0
		for i := range len(sortedVertexShape) {
			//Euclidean Distance
			dx, dy := sortedVertexShape[0].X-sortedVertexShape[i].X, sortedVertexShape[0].Y-sortedVertexShape[i].Y
			d := math.Sqrt(float64(dx*dx + dy*dy))
			if d >= distance {
				furthest = i
				distance = d
			}
		}
		//Cut in half
		half1 := sortedVertexShape[:furthest+1]
		half2 := sortedVertexShape[furthest:]
		//Add the starting point to the end of the second half
		half2 = append(half2, half1[0])

		//Perform RDP on the two halves
		half1 = RDPOptimizer(half1, epsilon)
		half2 = RDPOptimizer(half2, epsilon)

		//Combine halves back into one
		optimizedShape = append(half1[:len(half1)-1], half2[:len(half2)-1]...)
	}

	return optimizedShape
}

func RDPOptimizer(sortedVertexShape []Vertex, epsilon float64) []Vertex {
	//Check number of points
	if len(sortedVertexShape) < 2 {
		return sortedVertexShape
	}

	start := 0
	end := len(sortedVertexShape) - 1
	maxD := -1.0
	p1 := sortedVertexShape[0]
	p2 := sortedVertexShape[end]
	xDiff := p2.X - p1.X
	yDiff := p2.Y - p1.Y
	for i, p := range sortedVertexShape[1:end] {
		//Perpendicular Distance
		d := math.Abs(float64(yDiff*p.X - xDiff*p.Y + p2.X*p1.Y - p2.Y*p1.X))
		if d > maxD {
			start = i + 1
			maxD = d
		}
	}
	if maxD > epsilon {
		return append(
			RDPOptimizer(sortedVertexShape[:start+1], epsilon),
			RDPOptimizer(sortedVertexShape[start:], epsilon)[1:]...)
	}
	return []Vertex{sortedVertexShape[0], sortedVertexShape[end]}
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
				return false
			}
		}
	}
	return slices.Equal(sd.Path, other.Path)
}

type ShapeCreationOptions struct {
	NoColorSeparation,
	AllowWhite,
	PreserveColor,
	KeepSmallRegions bool
	EpsilonRDP float64
}

func isRegionLargeEnough(region *Region) bool {
	const MINIMUM_NUMBER_OF_PIXELS_FOR_NON_SMALL_REGION = 50
	return len(*region) >= MINIMUM_NUMBER_OF_PIXELS_FOR_NON_SMALL_REGION
}

func CreateShapes(img image.Image, opts ShapeCreationOptions) (data *BoardshapesData) {
	data = &BoardshapesData{
		Version: VERSION,
	}

	img = ResizeImage(img)

	newImg := SimplifyImage(img, opts)

	var filter func(*Region) bool
	if opts.KeepSmallRegions {
		filter = nil
	} else {
		filter = isRegionLargeEnough
	}

	regionMap := BuildRegionMap(newImg, opts, filter)

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

		if opts.EpsilonRDP == 0 {
			shape = OptimizeShape(shape)
		} else {
			shape = OptimizeShapeWithEpsilon(shape, opts.EpsilonRDP)
		}

		shapeData := ShapeData{
			Number:    i,
			Color:     regionColor,
			ColorName: regionColorName,
			CornerX:   minX,
			CornerY:   minY,
			Image:     regionImage,
			Path:      shape,
		}

		data.Shapes = append(data.Shapes, shapeData)
	}

	return
}
