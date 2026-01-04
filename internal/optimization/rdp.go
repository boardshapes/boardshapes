package optimization

import (
	"math"

	"github.com/boardshapes/boardshapes/geometry"
)

const defaultRdpEpsilon = 2.0

func OptimizeShape(sortedVertexShape []geometry.Vertex) []geometry.Vertex {
	return OptimizeShapeWithEpsilon(sortedVertexShape, defaultRdpEpsilon)
}

const minimumVerticesForRdp = 6

func OptimizeShapeWithEpsilon(sortedVertexShape []geometry.Vertex, epsilon float64) []geometry.Vertex {
	//Try optimizing straight lines
	optimizedShape := make([]geometry.Vertex, len(sortedVertexShape))
	copy(optimizedShape, sortedVertexShape)
	for i := 2; i < len(optimizedShape); i++ {
		v1 := optimizedShape[i-2].ToVertexF().DirectionTo(optimizedShape[i-1].ToVertexF())
		v2 := optimizedShape[i-1].ToVertexF().DirectionTo(optimizedShape[i].ToVertexF())
		if v1 == v2 {
			optimizedShape = append(optimizedShape[:i-1], optimizedShape[i:]...)
			i--
		}
	}

	//If epsilon is negative, skip RDP optimization
	if epsilon < 0 {
		return optimizedShape
	}

	//Check number of vertices after straight optimization to determine if RDP is needed
	if len(optimizedShape) > minimumVerticesForRdp {
		//Split shape in half by finding furthest geometry.Vertex from the startpoint
		startP := sortedVertexShape[0].ToVertexF()
		distance := 0.0
		furthest := 0
		for i := range sortedVertexShape {
			d := startP.DistanceTo(sortedVertexShape[i].ToVertexF())
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
	perpDir := p1.ToVertexF().DirectionTo(p2.ToVertexF()).Rotate90CCW()
	for i, p := range sortedVertexShape[1:end] {
		//Perpendicular Distance
		d := math.Abs(p.ToVertexF().Sub(p1.ToVertexF()).Dot(perpDir))
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
