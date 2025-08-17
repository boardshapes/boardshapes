package shared

import (
	"image"
	"image/color"
	"strings"
)

// the byte is the chunk ID
type ErrUnknownChunkType byte

func (e ErrUnknownChunkType) Error() string {
	return "deserialization: unknown chunk type " + string(e)
}

func TrimNullByte(s string) string {
	return strings.TrimRight(s, "\x00")
}

type SettableImage = interface {
	image.Image
	Set(x, y int, color color.Color)
}
