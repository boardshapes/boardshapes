package optimization

import (
	"math"

	"github.com/boardshapes/boardshapes/geometry"
)

const DEFAULT_RDP_EPSILON = 10.0

func OptimizeShape(sortedVertexShape []geometry.Vertex) []geometry.Vertex {
	return OptimizeShapeWithEpsilon(sortedVertexShape, DEFAULT_RDP_EPSILON)
}

const MINIMUM_VERTICES_FOR_RDP = 15

func OptimizeShapeWithEpsilon(sortedVertexShape []geometry.Vertex, epsilon float64) []geometry.Vertex {
	//Try optimizing straight lines
	var optimizedShape []geometry.Vertex
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
		//Split shape in half by finding furthest geometry.Vertex from the startpoint
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

func RDPOptimizer(sortedVertexShape []geometry.Vertex, epsilon float64) []geometry.Vertex {
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
	return []geometry.Vertex{sortedVertexShape[0], sortedVertexShape[end]}
}
