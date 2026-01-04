package optimization_test

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"testing"

	"github.com/boardshapes/boardshapes"
	"github.com/boardshapes/boardshapes/geometry"
	"github.com/boardshapes/boardshapes/internal/optimization"
	"github.com/fogleman/gg"
)

func regularPolygon(n int, x, y, r float64) []geometry.VertexF {
	result := make([]geometry.VertexF, n)
	for i := range n {
		a := float64(i)*2*math.Pi/float64(n) - math.Pi/2
		result[i] = geometry.VertexF{X: x + r*math.Cos(a), Y: y + r*math.Sin(a)}
	}
	return result
}

type optimizeTestCase struct {
	name              string
	keyVertices       []geometry.Vertex
	generatedVertices []geometry.Vertex
	epsilon           float64
}

func generateStarTestCase() optimizeTestCase {
	const n = 5
	c := gg.NewContext(1024, 1024)
	pointsF := regularPolygon(n, 512, 512, 400)
	c.SetColor(color.White)
	c.Clear()
	for i := range n + 1 {
		index := (i * 2) % n
		p := pointsF[index]
		c.LineTo(p.X, p.Y)
	}
	c.SetColor(color.Black)
	c.Fill()

	dat := boardshapes.CreateShapes(c.Image(), boardshapes.ShapeCreationOptions{EpsilonRDP: -1})
	if len(dat.Shapes) != 1 {
		log.Fatal("Could not generate star test case")
	}

	points := make([]geometry.Vertex, len(pointsF))
	for i := range pointsF {
		points[i] = geometry.Vertex{
			X: uint16(math.Round(pointsF[i].X)),
			Y: uint16(math.Round(pointsF[i].Y)),
		}
	}

	return optimizeTestCase{
		name:              "star",
		keyVertices:       points,
		generatedVertices: dat.Shapes[0].Path,
		epsilon:           2,
	}
}

func generateJaggedPolygonTestCase() optimizeTestCase {
	rng := rand.New(rand.NewSource(8675309))
	const n = 100

	c := gg.NewContext(1024, 1024)
	c.SetColor(color.White)
	c.Clear()

	pointsF := make([]geometry.VertexF, n)
	for i := range n {
		a := float64(i)*2*math.Pi/float64(n) - math.Pi/2
		r := rng.Float64()*400 + 100
		pointsF[i] = geometry.VertexF{X: 512 + r*math.Cos(a), Y: 512 + r*math.Sin(a)}
	}

	for _, v := range pointsF {
		c.LineTo(v.X, v.Y)
	}
	c.ClosePath()
	c.SetColor(color.Black)
	c.Fill()

	dat := boardshapes.CreateShapes(c.Image(), boardshapes.ShapeCreationOptions{EpsilonRDP: -1})
	if len(dat.Shapes) != 1 {
		log.Fatal("Could not generate star test case")
	}

	points := make([]geometry.Vertex, len(pointsF))
	for i := range pointsF {
		points[i] = geometry.Vertex{
			X: uint16(math.Round(pointsF[i].X)),
			Y: uint16(math.Round(pointsF[i].Y)),
		}
	}

	return optimizeTestCase{
		name:              "jagged-polygon",
		keyVertices:       points,
		generatedVertices: dat.Shapes[0].Path,
		epsilon:           2,
	}
}

func generateTestCases() []optimizeTestCase {
	return []optimizeTestCase{
		generateStarTestCase(),
		generateJaggedPolygonTestCase(),
	}
}

func TestOptimizeShapeWithEpsilon(t *testing.T) {
	c := generateTestCases()
	for _, tc := range c {
		t.Run(tc.name, func(t *testing.T) {
			optimized := optimization.OptimizeShapeWithEpsilon(tc.generatedVertices, tc.epsilon)

			t.Logf("Optimized shape of %d vertices to %d vertices", len(tc.generatedVertices), len(optimized))

			if len(optimized) > len(tc.keyVertices)*3 {
				t.Errorf(
					"Optimized shape has %d vertices, expected less than or equal to %d",
					len(optimized),
					len(tc.keyVertices)*3)

			outer:
				for _, v := range tc.keyVertices {
					for _, ov := range optimized {
						if v.ToVertexF().DistanceTo(ov.ToVertexF()) <= tc.epsilon {
							continue outer
						}
						t.Errorf("Key vertex %v not found in optimized vertices within epsilon %f", v, tc.epsilon)
					}
				}
			}
		})
	}
}
