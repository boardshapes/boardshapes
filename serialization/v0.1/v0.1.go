package v0_1

import (
	main "boardshapes/boardshapes"
	"boardshapes/boardshapes/serialization/shared"
	"bytes"
	"encoding/binary"
	"image/color"
	"image/png"
	"io"
)

const (
	CHUNK_VERSION        = 0
	CHUNK_COLOR_TABLE    = 2
	CHUNK_SHAPE_GEOMETRY = 8
	CHUNK_SHAPE_COLOR    = 9
	CHUNK_SHAPE_IMAGE    = 10
	CHUNK_SHAPE_MASK     = 11
)

func BinaryDeserialize(r io.Reader) (*main.BoardshapesData, error) {
	data := &main.BoardshapesData{}

	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	colors := make(map[color.NRGBA]string, 0)
	shapes := make(map[int]main.ShapeData, 0)
	for {
		chunkId, err := buf.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch chunkId {
		case CHUNK_VERSION:
			version, err := buf.ReadString(0)
			if err != nil {
				return nil, err
			}
			version = shared.TrimNullByte(version)
			data.Version = version
		case CHUNK_COLOR_TABLE:
			nColors := new(uint32)
			binary.Read(&buf, binary.BigEndian, nColors)
			for range *nColors {
				channels := make([]byte, 4)
				_, err := buf.Read(channels)
				if err != nil {
					return nil, err
				}
				r, g, b, a := channels[0], channels[1], channels[2], channels[3]
				colorName, err := buf.ReadString(0)
				if err != nil {
					return nil, err
				}
				colorName = shared.TrimNullByte(colorName)
				colors[color.NRGBA{R: r, G: g, B: b, A: a}] = colorName
			}
		case CHUNK_SHAPE_GEOMETRY, CHUNK_SHAPE_COLOR, CHUNK_SHAPE_IMAGE, CHUNK_SHAPE_MASK: // shape chunks
			var shape main.ShapeData
			var inShapesMap bool
			shapeNumber := new(uint32)
			binary.Read(&buf, binary.BigEndian, shapeNumber)

			if shape, inShapesMap = shapes[int(*shapeNumber)]; !inShapesMap {
				shape = main.ShapeData{
					Number: int(*shapeNumber),
				}
			}

			switch chunkId {
			case CHUNK_SHAPE_GEOMETRY:
				d := make([]byte, 12)
				_, err := buf.Read(d)
				if err != nil {
					return nil, err
				}

				cornerX, cornerY, nVertices := binary.BigEndian.Uint32(d[0:4]), binary.BigEndian.Uint32(d[4:8]), binary.BigEndian.Uint32(d[8:12])
				shape.CornerX = int(cornerX)
				shape.CornerY = int(cornerY)

				path := make([]main.Vertex, nVertices)
				for i := range nVertices {
					bv := make([]byte, 8)
					_, err := buf.Read(bv)
					if err != nil {
						return nil, err
					}
					x, y := binary.BigEndian.Uint32(bv[0:4]), binary.BigEndian.Uint32(bv[4:8])
					path[i] = main.Vertex{X: uint16(x), Y: uint16(y)}
				}

				shape.Path = path
			case CHUNK_SHAPE_COLOR:
				d := make([]byte, 4)
				_, err := buf.Read(d)
				if err != nil {
					return nil, err
				}
				r, g, b, a := d[0], d[1], d[2], d[3]
				shape.Color = color.NRGBA{R: r, G: g, B: b, A: a}
			case CHUNK_SHAPE_IMAGE:
				l := new(uint32)
				binary.Read(&buf, binary.BigEndian, l)
				var pngBuf bytes.Buffer
				pngBuf.Grow(int(*l))
				_, err = io.CopyN(&pngBuf, &buf, int64(*l))
				if err != nil {
					return nil, err
				}
				img, err := png.Decode(&pngBuf)
				if err != nil {
					return nil, err
				}
				shape.Image = img
			case CHUNK_SHAPE_MASK:
				// WIP
			}

			shapes[int(*shapeNumber)] = shape
		default:
			return nil, shared.ErrUnknownChunkType(chunkId)
		}
	}
	// add color names to shapes

	return data, nil
}
