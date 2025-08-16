package serialization

import (
	main "boardshapes/boardshapes"
	"boardshapes/boardshapes/serialization/shared"
	v0_1 "boardshapes/boardshapes/serialization/v0.1"
	"bytes"
	"encoding/binary"
	"errors"
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

type DeserializeFunc func(r io.Reader) (*main.BoardshapesData, error)

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
	chunk = append(chunk, byte(len(colors)))

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
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.CornerX))
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.CornerY))
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(len(shape.Path)))

		for _, vert := range shape.Path {
			chunk = binary.BigEndian.AppendUint32(chunk, uint32(vert.X))
			chunk = binary.BigEndian.AppendUint32(chunk, uint32(vert.Y))
		}

		// shape color chunk
		nrgba := main.GetNRGBA(shape.Color)
		chunk = append(chunk, CHUNK_SHAPE_COLOR)
		chunk = binary.BigEndian.AppendUint32(chunk, uint32(shape.Number))
		chunk = append(chunk, nrgba.R, nrgba.G, nrgba.B, nrgba.A)

		// shape image chunk
		chunk = append(chunk, CHUNK_SHAPE_IMAGE)

		var pngBuf bytes.Buffer
		png.Encode(&pngBuf, shape.Image)

		chunk = binary.BigEndian.AppendUint32(chunk, uint32(pngBuf.Len()))
		chunk = append(chunk, pngBuf.Bytes()...)

		_, err = buf.Write(chunk)
		if err != nil {
			return err
		}
	}

	_, err = buf.WriteTo(w)

	return err
}

func BinaryDeserialize(r io.Reader) (*main.BoardshapesData, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	if buf.Bytes()[0] != 0 {
		return nil, ErrVersionNotFound
	}

	version, err := buf.ReadString(0)
	if err != nil {
		return nil, err
	}
	version = shared.TrimNullByte(version)

	vnums := strings.Split(version, ".")
	if len(vnums) < 2 {
		return nil, ErrVersionNotFound
	}

	deserializeFunc, ok := deserializers[vnums[0]+"."+vnums[1]]
	if !ok {
		return nil, ErrIncompatibleVersion
	}

	return deserializeFunc(&buf)
}
