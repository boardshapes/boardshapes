package shared

import "strings"

// the byte is the chunk ID
type ErrUnknownChunkType byte

func (e ErrUnknownChunkType) Error() string {
	return "unknown chunk type encountered during deserialization: " + string(e)
}

func TrimNullByte(s string) string {
	return strings.TrimRight(s, "\x00")
}
