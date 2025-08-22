package serialization

import (
	main "boardshapes/boardshapes"
	v0_1 "boardshapes/boardshapes/serialization/v0.1"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
)

const (
	CHUNK_VERSION        = 0
	CHUNK_COLOR_TABLE    = 2
	CHUNK_SHAPE_GEOMETRY = 8
	CHUNK_SHAPE_COLOR    = 9
	CHUNK_SHAPE_IMAGE    = 10
	CHUNK_SHAPE_MASK     = 11
)

type DeserializeFunc func(r io.Reader, options map[string]any) (*main.BoardshapesData, error)

// the byte is the chunk ID
type ErrUnknownChunkType byte

func (e ErrUnknownChunkType) Error() string {
	return "unknown chunk type encountered during deserialization: " + string(e)
}

var ErrVersionNotFound = errors.New("version of the data could not be found, cannot deserialize with backwards-compatible deserializer")
var ErrInvalidVersion = errors.New("version of the data is invalid, cannot deserialize with backwards-compatible deserializer")
var ErrIncompatibleVersion = errors.New("version of the data is incompatible with the backwards-compatible deserializer, cannot deserialize")

var deserializers = map[string]DeserializeFunc{
	"0.1": v0_1.BinaryDeserialize,
}

type SerializationOptions struct {
	UseMasks bool
}

var DefaultOptions = SerializationOptions{
	UseMasks: true,
}

func BinarySerialize(w io.Writer, data main.BoardshapesData, options *SerializationOptions) error {
	if options == nil {
		options = &DefaultOptions
	}

	var buf bytes.Buffer

	// write version chunk
	chunk := append([]byte{CHUNK_VERSION}, append([]byte(main.VERSION), 0)...)
	_, err := buf.Write(chunk)
	if err != nil {
		return err
	}

	// get all colors
	colors := make(map[color.Color]string)
	for _, shape := range data.Shapes {
		if shape.ColorName != "" {
			colors[shape.Color] = shape.ColorName
		}
	}

	// write color table chunk
	chunk = []byte{CHUNK_COLOR_TABLE}
	chunk = binary.BigEndian.AppendUint32(chunk, uint32(len(colors)))

	for color, name := range colors {
		nrgba := main.GetNRGBA(color)
		chunk = append(chunk, nrgba.R, nrgba.G, nrgba.B, nrgba.A)
		chunk = append(chunk, name...)
		chunk = append(chunk, 0)
	}
	_, err = buf.Write(chunk)
	if err != nil {
		return err
	}

	// write shapes chunks
	for _, shape := range data.Shapes {
		// shape geometry chunk
		chunk := []byte{CHUNK_SHAPE_GEOMETRY}

		chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.Number))
		chunk = binary.BigEndian.AppendUint16(chunk, uint16(shape.CornerX))
		chunk = binary.BigEndian.AppendUint16(chunk, uint16(shape.CornerY))
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(len(shape.Path)))

		for _, vert := range shape.Path {
			chunk = binary.BigEndian.AppendUint16(chunk, uint16(vert.X))
			chunk = binary.BigEndian.AppendUint16(chunk, uint16(vert.Y))
		}

		// shape color chunk
		nrgba := main.GetNRGBA(shape.Color)
		chunk = append(chunk, CHUNK_SHAPE_COLOR)
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.Number))
		chunk = append(chunk, nrgba.R, nrgba.G, nrgba.B, nrgba.A)

		if shape.Image != nil && shape.Image.Bounds().Dx() > 0 && shape.Image.Bounds().Dy() > 0 {
			if options.UseMasks {
				// shape mask chunk
				chunk = append(chunk, CHUNK_SHAPE_MASK)
				chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.Number))

				img := shape.Image
				bds := img.Bounds()

				chunk = binary.BigEndian.AppendUint16(chunk, uint16(bds.Dx()))

				_, _, _, a := img.At(bds.Min.X, bds.Min.Y).RGBA()
				prevFilled := a > 0
				if prevFilled {
					chunk = append(chunk, 1)
				} else {
					chunk = append(chunk, 0)
				}

				runLength := 0

				for y := bds.Min.Y; y < bds.Max.Y; y++ {
					for x := bds.Min.X; x < bds.Max.X; x++ {
						_, _, _, a := img.At(x, y).RGBA()
						filled := a > 0
						if prevFilled == filled {
							runLength++
						} else {
							chunk = binary.AppendUvarint(chunk, uint64(runLength))
							runLength = 1
							prevFilled = filled
						}
					}
				}

				chunk = binary.AppendUvarint(chunk, uint64(runLength))
				chunk = append(chunk, 0)
			} else {
				// shape image chunk
				chunk = append(chunk, CHUNK_SHAPE_IMAGE)
				chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.Number))

				var pngBuf bytes.Buffer
				png.Encode(&pngBuf, shape.Image)

				chunk = binary.BigEndian.AppendUint32(chunk, uint32(pngBuf.Len()))
				chunk = append(chunk, pngBuf.Bytes()...)
			}
		}

		_, err = buf.Write(chunk)
		if err != nil {
			return err
		}
	}

	_, err = buf.WriteTo(w)

	return err
}

func BinaryDeserialize(r io.Reader, options map[string]any) (*main.BoardshapesData, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	bufBytes := buf.Bytes()
	if bufBytes[0] != 0 {
		return nil, ErrVersionNotFound
	}

	nullIndex := bytes.IndexByte(bufBytes[1:], 0)
	if nullIndex == -1 {
		return nil, ErrVersionNotFound
	}
	version := string(bufBytes[1 : nullIndex+1])

	vnums := strings.Split(version, ".")
	if len(vnums) < 2 {
		return nil, ErrVersionNotFound
	}

	deserializeFunc, ok := deserializers[vnums[0]+"."+vnums[1]]
	if !ok {
		return nil, ErrIncompatibleVersion
	}

	return deserializeFunc(&buf, options)
}

type JSONData struct {
	Version string          `json:"version"`
	Shapes  []JSONShapeData `json:"shapes"`
}

type JSONShapeData struct {
	Number      int         `json:"number"`
	CornerX     int         `json:"corner_x"`
	CornerY     int         `json:"corner_y"`
	Shape       []uint16    `json:"path"`
	Color       color.NRGBA `json:"color"`
	ColorString string      `json:"color_string"`
	Image       string      `json:"image"`
}

func JsonSerialize(data *main.BoardshapesData) ([]byte, error) {
	jsonData := JSONData{
		Version: data.Version,
		Shapes:  make([]JSONShapeData, len(data.Shapes)),
	}

	for i, shape := range data.Shapes {
		points := make([]uint16, len(shape.Path)*2)
		for j, v := range shape.Path {
			points[j*2] = v.X
			points[j*2+1] = v.Y
		}

		var imgBase64 string
		if shape.Image != nil {
			buf := new(bytes.Buffer)
			if err := png.Encode(buf, shape.Image); err != nil {
				return nil, err
			}
			imgBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
		}

		jsonData.Shapes[i] = JSONShapeData{
			Number:      shape.Number,
			CornerX:     shape.CornerX,
			CornerY:     shape.CornerY,
			Shape:       points,
			Color:       shape.Color.(color.NRGBA),
			ColorString: shape.ColorName,
			Image:       imgBase64,
		}
	}

	return json.Marshal(jsonData)
}

func JsonDeserialize(d []byte) (*main.BoardshapesData, error) {
	var jsonData JSONData
	if err := json.Unmarshal(d, &jsonData); err != nil {
		return nil, err
	}

	data := &main.BoardshapesData{
		Version: jsonData.Version,
		Shapes:  make([]main.ShapeData, len(jsonData.Shapes)),
	}

	for i, jsonShape := range jsonData.Shapes {
		path := make([]main.Vertex, len(jsonShape.Shape)/2)
		for j := range path {
			path[j] = main.Vertex{
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

		data.Shapes[i] = main.ShapeData{
			Number:    jsonShape.Number,
			CornerX:   jsonShape.CornerX,
			CornerY:   jsonShape.CornerY,
			Path:      path,
			Color:     jsonShape.Color,
			ColorName: jsonShape.ColorString,
			Image:     img,
		}
	}

	return data, nil
}
