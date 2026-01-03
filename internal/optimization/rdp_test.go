package optimization_test

import (
	"image/color"
	"math"
	"testing"

	"github.com/boardshapes/boardshapes"
	"github.com/boardshapes/boardshapes/geometry"
	"github.com/boardshapes/boardshapes/internal/optimization"
	"github.com/fogleman/gg"
)

func regularPolygon(n int, x, y, r float64) []geometry.VertexF {
	result := make([]geometry.VertexF, n)
	for i := 0; i < n; i++ {
		a := float64(i)*2*math.Pi/float64(n) - math.Pi/2
		result[i] = geometry.VertexF{X: x + r*math.Cos(a), Y: y + r*math.Sin(a)}
	}
	return result
}

type optimizeTestCase struct {
	name              string
	originalVertices  []geometry.Vertex
	generatedVertices []geometry.Vertex
	epsilon           float64
}

func generateTestShapes() []optimizeTestCase {
	const n = 5
	c := gg.NewContext(1024, 1024)
	points := regularPolygon(n, 512, 512, 400)
	c.SetColor(color.White)
	c.Clear()
	for i := range n + 1 {
		index := (i * 2) % n
		p := points[index]
		c.LineTo(p.X, p.Y)
	}
	c.SetColor(color.Black)
	c.Fill()

	dat := boardshapes.CreateShapes(c.Image(), boardshapes.ShapeCreationOptions{EpsilonRDP: -1})
	if len(dat.Shapes) == 0 {
		panic("No shapes generated")
	}

	pointsI := make([]geometry.Vertex, len(points))
	for i := range points {
		pointsI[i] = geometry.Vertex{
			X: uint16(math.Round(points[i].X)),
			Y: uint16(math.Round(points[i].Y)),
		}
	}

	return []optimizeTestCase{
		{
			name:              "star",
			originalVertices:  pointsI,
			generatedVertices: dat.Shapes[0].Path,
			epsilon:           2,
		},
	}
}

func TestOptimizeShapeWithEpsilon(t *testing.T) {
	c := generateTestShapes()
	for _, tc := range c {
		t.Run(tc.name, func(t *testing.T) {
			optimized := optimization.OptimizeShapeWithEpsilon(tc.generatedVertices, tc.epsilon)
			if len(optimized) > len(tc.originalVertices)*3 {
				t.Errorf(
					"Optimized shape has %d vertices, expected less than or equal to %d",
					len(optimized),
					len(tc.originalVertices)*3)
			}
		})
	}
}
