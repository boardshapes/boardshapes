package processing

import (
	"cmp"
	"errors"
	"fmt"
	"image"
	"slices"

	"github.com/boardshapes/boardshapes/geometry"
	"github.com/boardshapes/boardshapes/internal/regions"
	"github.com/boardshapes/boardshapes/utils/grid"
)

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

func CreateShapeFromRegion(region *regions.Region) (shape []geometry.Vertex, err error) {
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

	largestShapeVertices := slices.MaxFunc(possibleShapeVertices, func(a, b []geometry.Vertex) int {
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
		geometry.Vertex{
			X: largestShapeVertices[0].X - 1,
			Y: largestShapeVertices[0].Y - 1,
		},
		vertexMatrix,
		len(largestShapeVertices))
}

func createRegionPixelsMatrix(region *regions.Region, regionBounds image.Rectangle) (regionPixels [][]RegionPixel) {
	regionPixels = make([][]RegionPixel, regionBounds.Dx()+2)
	for i := range regionPixels {
		regionPixels[i] = make([]RegionPixel, regionBounds.Dy()+2)
	}

	for _, v := range *region {
		regionPixels[int(v.X+1)-regionBounds.Min.X][int(v.Y+1)-regionBounds.Min.Y].MarkInRegion()
	}

	verticesToVisit := []geometry.Vertex{{X: 0, Y: 0}}
	// visit outer pixels
	for len(verticesToVisit) > 0 {
		v := verticesToVisit[len(verticesToVisit)-1]
		verticesToVisit = verticesToVisit[:len(verticesToVisit)-1]
		if !regionPixels[v.X][v.Y].Visited() {
			regionPixels[v.X][v.Y].MarkVisited()
			grid.ForNonDiagonalAdjacents(
				v.X, v.Y, len(regionPixels), len(regionPixels[0]),
				func(x, y uint16) {
					if !regionPixels[x][y].Visited() && !regionPixels[x][y].IsOuter() {
						if regionPixels[x][y].InRegion() {
							regionPixels[x][y].MarkIsOuter()
						} else {
							verticesToVisit = append(verticesToVisit, geometry.Vertex{X: x, Y: y})
						}
					}
				})

		}
	}
	return regionPixels
}

func findShapes(regionPixels [][]RegionPixel) [][]geometry.Vertex {
	possibleShapeVertices := make([][]geometry.Vertex, 0, 1)
	for y := uint16(0); y < uint16(len(regionPixels[0])); y++ {
		for x := uint16(0); x < uint16(len(regionPixels)); x++ {
			rp := regionPixels[x][y]
			// check if inner pixel
			if !rp.Visited() && !rp.IsOuter() {
				verticesToVisit := []geometry.Vertex{{X: x, Y: y}}
				newInnerShape := make([]geometry.Vertex, 0)
				// visit inner pixels
				for len(verticesToVisit) > 0 {
					v := verticesToVisit[len(verticesToVisit)-1]
					verticesToVisit = verticesToVisit[:len(verticesToVisit)-1]
					if regionPixels[v.X][v.Y].Visited() {
						continue
					}
					regionPixels[v.X][v.Y].MarkVisited()
					grid.ForNonDiagonalAdjacents(
						v.X, v.Y, len(regionPixels), len(regionPixels[0]),
						func(x, y uint16) {
							if !regionPixels[x][y].Visited() && !regionPixels[x][y].IsInner() {
								if regionPixels[x][y].IsOuter() {
									regionPixels[x][y].MarkIsInner()
									newInnerShape = append(newInnerShape, geometry.Vertex{X: x, Y: y})
								} else {
									verticesToVisit = append(verticesToVisit, geometry.Vertex{X: x, Y: y})
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

func findSortedShapeVertices(startingVertex geometry.Vertex, vertexMatrix [][]bool, maxLength int) ([]geometry.Vertex, error) {
	var previousVertex geometry.Vertex
	var isPreviousVertexSet = false
	var currentVertex geometry.Vertex = startingVertex
	sortedShapeVertices := make([]geometry.Vertex, 0, maxLength)

	for {
		adjacentVertices := make([]geometry.Vertex, 0, 8)

		grid.ForAdjacents(currentVertex.X, currentVertex.Y, len(vertexMatrix), len(vertexMatrix[0]), func(x, y uint16) {
			if vertexMatrix[x][y] {
				adjacentVertices = append(adjacentVertices, geometry.Vertex{X: uint16(x), Y: uint16(y)})
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
