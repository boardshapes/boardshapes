package geometry

import (
	"fmt"
	"math"
)

type Vertex struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

func (v1 Vertex) Add(v2 Vertex) Vertex {
	return Vertex{
		X: v1.X + v2.X,
		Y: v1.Y + v2.Y,
	}
}

func (v1 Vertex) Sub(v2 Vertex) Vertex {
	return Vertex{
		X: v1.X - v2.X,
		Y: v1.Y - v2.Y,
	}
}

func (v Vertex) ToVertexF() VertexF {
	return VertexF{
		X: float64(v.X),
		Y: float64(v.Y),
	}
}

func (v Vertex) String() string {
	return fmt.Sprintf("(%d, %d)", v.X, v.Y)
}

type VertexF struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (v1 VertexF) Add(v2 VertexF) VertexF {
	return VertexF{
		X: v1.X + v2.X,
		Y: v1.Y + v2.Y,
	}
}

func (v1 VertexF) Sub(v2 VertexF) VertexF {
	return VertexF{
		X: v1.X - v2.X,
		Y: v1.Y - v2.Y,
	}
}

func (v1 VertexF) Dot(v2 VertexF) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

func (v VertexF) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v VertexF) Normalized() VertexF {
	mag := v.Magnitude()
	return VertexF{
		X: v.X / mag,
		Y: v.Y / mag,
	}
}

func (v VertexF) Rotate90CCW() VertexF {
	return VertexF{
		X: -v.Y,
		Y: v.X,
	}
}

func (v1 VertexF) DistanceTo(v2 VertexF) float64 {
	return v1.Sub(v2).Magnitude()
}

func (v1 VertexF) DirectionTo(v2 VertexF) VertexF {
	return v2.Sub(v1).Normalized()
}

func (v VertexF) ToVertex() Vertex {
	return Vertex{
		X: uint16(math.Round(v.X)),
		Y: uint16(math.Round(v.Y)),
	}
}

func (v VertexF) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", v.X, v.Y)
}
