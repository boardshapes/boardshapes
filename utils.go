package boardshapes

import "math"

func absDiff[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](a T, b T) T {
	if a > b {
		return a - b
	}
	return b - a
}

func DotProduct(x1, x2, y1, y2 float64) float64 {
	answer := (x1 * x2) + (y1 * y2)
	return answer
}

func (v1 Vertex) DirectionTo(v2 Vertex) (x, y float64) {
	answerX := float64(v2.X - v1.X)
	answerY := float64(v2.Y - v1.Y)
	mag := math.Sqrt((answerX * answerX) + (answerY * answerY))
	return (answerX / mag), (answerY / mag)
}

func forNonDiagonalAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		function(x, y-1)
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		function(x, y+1)
	}
}

func forAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		if x > 0 {
			function(x-1, y-1)
		}
		function(x, y-1)
		if x < uint16(maxX)-1 {
			function(x+1, y-1)
		}
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		if x > 0 {
			function(x-1, y+1)
		}
		function(x, y+1)
		if x < uint16(maxX)-1 {
			function(x+1, y+1)
		}
	}
}
