package v0_1

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/boardshapes/boardshapes/colors"
	"github.com/boardshapes/boardshapes/data"
	"github.com/boardshapes/boardshapes/geometry"
	"github.com/boardshapes/boardshapes/imageops"
	"github.com/boardshapes/boardshapes/serialization/shared"
)

const (
	CHUNK_VERSION        = 0
	CHUNK_COLOR_TABLE    = 2
	CHUNK_SHAPE_GEOMETRY = 8
	CHUNK_SHAPE_COLOR    = 9
	CHUNK_SHAPE_IMAGE    = 10
	CHUNK_SHAPE_MASK     = 11
)

func BinaryDeserialize(r io.Reader, options map[string]any) (*data.BoardshapesData, error) {
	boardshapesData := &data.BoardshapesData{}
	var baseImage image.Image
	if img, ok := options["baseImage"].(image.Image); ok {
		baseImage = img
	}

	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	colorNames := make(map[color.NRGBA]string, 0)
	shapes := make(map[int]data.ShapeData, 0)
	shapesUsingMasks := make([]int, 0)
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
			boardshapesData.Version = version
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
				colorNames[color.NRGBA{R: r, G: g, B: b, A: a}] = colorName
			}
		case CHUNK_SHAPE_GEOMETRY, CHUNK_SHAPE_COLOR, CHUNK_SHAPE_IMAGE, CHUNK_SHAPE_MASK: // shape chunks
			var shape data.ShapeData
			var inShapesMap bool
			shapeNumber := new(uint32)
			binary.Read(&buf, binary.BigEndian, shapeNumber)

			if shape, inShapesMap = shapes[int(*shapeNumber)]; !inShapesMap {
				shape = data.ShapeData{
					Number: int(*shapeNumber),
				}
			}

			switch chunkId {
			case CHUNK_SHAPE_GEOMETRY:
				d := make([]byte, 8)
				_, err := buf.Read(d)
				if err != nil {
					return nil, err
				}

				cornerX, cornerY, nVertices := binary.BigEndian.Uint16(d[0:2]), binary.BigEndian.Uint16(d[2:4]), binary.BigEndian.Uint32(d[4:8])
				shape.CornerX = int(cornerX)
				shape.CornerY = int(cornerY)

				path := make([]geometry.Vertex, nVertices)
				for i := range nVertices {
					bv := make([]byte, 4)
					_, err := buf.Read(bv)
					if err != nil {
						return nil, err
					}
					x, y := binary.BigEndian.Uint16(bv[0:2]), binary.BigEndian.Uint16(bv[2:4])
					path[i] = geometry.Vertex{X: uint16(x), Y: uint16(y)}
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
				shapesUsingMasks = append(shapesUsingMasks, shape.Number)
				width := new(uint16)
				binary.Read(&buf, binary.BigEndian, width)
				startsFilled, err := buf.ReadByte()
				if err != nil {
					return nil, err
				}

				filled := startsFilled > 0
				b, err := buf.ReadBytes(0x00)
				if err != nil {
					return nil, err
				}

				runLengths := make([]uint, 0)
				for len(b) > 0 {
					runLength, nBytes := binary.Uvarint(b)
					runLengths = append(runLengths, uint(runLength))
					b = b[nBytes:]
				}

				sum := uint(0)
				for _, rl := range runLengths {
					sum += rl
				}

				if sum%uint(*width) != 0 {
					return nil, errors.New("deserialization: mask width does not divide evenly into total number of pixels in mask")
				}

				height := sum / uint(*width)
				img := image.NewNRGBA(image.Rect(0, 0, int(*width), int(height)))
				i := 0
				for _, rl := range runLengths {
					for range rl {
						if filled {
							img.Set(i%int(*width), i/int(*width), colors.Black)
						} else {
							img.Set(i%int(*width), i/int(*width), colors.Blank)
						}
						i++
					}
					filled = !filled
				}
				shape.Image = img
			}

			shapes[int(*shapeNumber)] = shape
		default:
			return nil, shared.ErrUnknownChunkType(chunkId)
		}
	}

	// add color names to shapes
	for i, shape := range shapes {
		if shape.Color != nil {
			colorName, ok := colorNames[colors.GetNRGBA(shape.Color)]
			if ok {
				shape.ColorName = colorName
				shapes[i] = shape
			}
		}
	}

	var getPixelColor func(x, y int, shape data.ShapeData) color.Color
	if baseImage != nil {
		getPixelColor = func(x, y int, shape data.ShapeData) color.Color {
			return baseImage.At(shape.CornerX+x, shape.CornerY+y)
		}
	} else {
		getPixelColor = func(_, _ int, shape data.ShapeData) color.Color {
			return shape.Color
		}
	}

	// restore color to shapes using masks
	for _, shapeNumber := range shapesUsingMasks {
		shape := shapes[shapeNumber]
		img, ok := shape.Image.(imageops.SettableImage)
		if ok {
			for y := 0; y < img.Bounds().Dy(); y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {
					if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
						img.Set(x, y, getPixelColor(x, y, shape))
					}
				}
			}
		}
	}

	boardshapesData.Shapes = make([]data.ShapeData, 0, len(shapes))
	for _, shape := range shapes {
		boardshapesData.Shapes = append(boardshapesData.Shapes, shape)
	}

	return boardshapesData, nil
}

type JSONData struct {
	Version string          `json:"version"`
	Shapes  []JSONShapeData `json:"shapes"`
}

type JSONShapeData struct {
	Number      int         `json:"number"`
	CornerX     int         `json:"cornerX"`
	CornerY     int         `json:"cornerY"`
	Shape       []uint16    `json:"path"`
	Color       color.NRGBA `json:"color"`
	ColorString string      `json:"colorString"`
	Image       string      `json:"image"`
}

func JsonDeserialize(r io.Reader, options map[string]any) (*data.BoardshapesData, error) {
	var jsonData JSONData
	if err := json.NewDecoder(r).Decode(&jsonData); err != nil {
		return nil, err
	}

	boardshapesData := &data.BoardshapesData{
		Version: jsonData.Version,
		Shapes:  make([]data.ShapeData, len(jsonData.Shapes)),
	}

	for i, jsonShape := range jsonData.Shapes {
		path := make([]geometry.Vertex, len(jsonShape.Shape)/2)
		for j := range path {
			path[j] = geometry.Vertex{
				X: jsonShape.Shape[j*2],
				Y: jsonShape.Shape[j*2+1],
			}
		}

		var img image.Image
		if jsonShape.Image != "" {
			imgBytes, err := base64.StdEncoding.DecodeString(jsonShape.Image)
			if err != nil {
				return nil, err
			}
			img, err = png.Decode(bytes.NewReader(imgBytes))
			if err != nil {
				return nil, err
			}
		}

		boardshapesData.Shapes[i] = data.ShapeData{
			Number:    jsonShape.Number,
			CornerX:   jsonShape.CornerX,
			CornerY:   jsonShape.CornerY,
			Path:      path,
			Color:     jsonShape.Color,
			ColorName: jsonShape.ColorString,
			Image:     img,
		}
	}

	return boardshapesData, nil
}
