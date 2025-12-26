package geometry

import "math"

type Vertex struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

func (v1 Vertex) DirectionTo(v2 Vertex) (x, y float64) {
	answerX := float64(v2.X - v1.X)
	answerY := float64(v2.Y - v1.Y)
	mag := math.Sqrt((answerX * answerX) + (answerY * answerY))
	return (answerX / mag), (answerY / mag)
}
