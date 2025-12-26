package data

import "fmt"

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
