package serialization

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"image/color"
	"image/png"
	"io"
	"strings"

	main "github.com/boardshapes/boardshapes"
	"github.com/boardshapes/boardshapes/colors"
	"github.com/boardshapes/boardshapes/data"
	v0_1 "github.com/boardshapes/boardshapes/serialization/v0.1"
)

const (
	CHUNK_VERSION        = 0
	CHUNK_COLOR_TABLE    = 2
	CHUNK_SHAPE_GEOMETRY = 8
	CHUNK_SHAPE_COLOR    = 9
	CHUNK_SHAPE_IMAGE    = 10
	CHUNK_SHAPE_MASK     = 11
)

type BinaryDeserializeFunc func(r io.Reader, options map[string]any) (*data.BoardshapesData, error)
type JsonDeserializeFunc func(r io.Reader, options map[string]any) (*data.BoardshapesData, error)

// the byte is the chunk ID
type ErrUnknownChunkType byte

func (e ErrUnknownChunkType) Error() string {
	return "unknown chunk type encountered during deserialization: " + string(e)
}

var ErrVersionNotFound = errors.New("version of the data could not be found, cannot deserialize with backwards-compatible deserializer")
var ErrInvalidVersion = errors.New("version of the data is invalid, cannot deserialize with backwards-compatible deserializer")
var ErrIncompatibleVersion = errors.New("version of the data is incompatible with the backwards-compatible deserializer, cannot deserialize")

var binaryDeserializers = map[string]BinaryDeserializeFunc{
	"0.1": v0_1.BinaryDeserialize,
}

var jsonDeserializers = map[string]JsonDeserializeFunc{
	"0.1": v0_1.JsonDeserialize,
}

type SerializationOptions struct {
	UseMasks bool
}

var DefaultOptions = SerializationOptions{
	UseMasks: true,
}

func BinarySerialize(w io.Writer, data *data.BoardshapesData, options *SerializationOptions) error {
	if options == nil {
		options = &DefaultOptions
	}

	var buf bytes.Buffer

	// write version chunk
	chunk := append([]byte{CHUNK_VERSION}, append([]byte(main.Version), 0)...)
	_, err := buf.Write(chunk)
	if err != nil {
		return err
	}

	// get all colorNames
	colorNames := make(map[color.Color]string)
	for _, shape := range data.Shapes {
		if shape.ColorName != "" {
			colorNames[shape.Color] = shape.ColorName
		}
	}

	// write color table chunk
	chunk = []byte{CHUNK_COLOR_TABLE}
	chunk = binary.BigEndian.AppendUint32(chunk, uint32(len(colorNames)))

	for color, name := range colorNames {
		nrgba := colors.GetNRGBA(color)
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
		nrgba := colors.GetNRGBA(shape.Color)
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

func BinaryDeserialize(r io.Reader, options map[string]any) (*data.BoardshapesData, error) {
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

	deserializeFunc, ok := binaryDeserializers[vnums[0]+"."+vnums[1]]
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
	CornerX     int         `json:"cornerX"`
	CornerY     int         `json:"cornerY"`
	Shape       []uint16    `json:"path"`
	Color       color.NRGBA `json:"color"`
	ColorString string      `json:"colorString"`
	Image       string      `json:"image"`
}

func JsonSerialize(w io.Writer, data *data.BoardshapesData) error {
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
				return err
			}
			imgBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
		}

		jsonData.Shapes[i] = JSONShapeData{
			Number:      shape.Number,
			CornerX:     shape.CornerX,
			CornerY:     shape.CornerY,
			Shape:       points,
			Color:       colors.GetNRGBA(shape.Color),
			ColorString: shape.ColorName,
			Image:       imgBase64,
		}
	}

	if err := json.NewEncoder(w).Encode(jsonData); err != nil {
		return err
	}

	return nil
}

type JSONWithVersion struct {
	Version string `json:"version"`
}

func JsonDeserialize(r io.Reader, options map[string]any) (*data.BoardshapesData, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	var jsonWithVersion JSONWithVersion
	if err := json.Unmarshal(buf.Bytes(), &jsonWithVersion); err != nil {
		return nil, err
	}

	vnums := strings.Split(jsonWithVersion.Version, ".")
	if len(vnums) < 2 {
		return nil, ErrVersionNotFound
	}

	deserializeFunc, ok := jsonDeserializers[vnums[0]+"."+vnums[1]]
	if !ok {
		return nil, ErrIncompatibleVersion
	}

	return deserializeFunc(&buf, options)
}
