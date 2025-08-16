package shared

import (
	"strings"
)

// the byte is the chunk ID
type ErrUnknownChunkType byte

func (e ErrUnknownChunkType) Error() string {
	return "unknown chunk type encountered during deserialization: " + string(e)
}

func TrimNullByte(s string) string {
	return strings.TrimRight(s, "\x00")
}

// IDIOT
// // https://en.wikipedia.org/wiki/Run-length_encoding
// // this is a contender for the rest of the everything being lil endian
// func AppendRLE_Bit(b []byte, runLength uint32, value bool) []byte {
// 	if runLength == 0 {
// 		return b
// 	}

// 	var bitv byte = 0x00
// 	if value {
// 		bitv = 0x40
// 	}
// 	bc := bitv | byte(runLength&0xC0)
// 	if runLength&^0x3F > 0 {
// 		bc = 0x80 | bc
// 	}
// 	b = append(b, bc)

// 	runLength >>= 6
// 	for runLength > 0 {
// 		bc := byte(runLength & 0x7F)
// 		if runLength&^0x7F > 0 {
// 			bc |= 0x80
// 		}
// 		b = append(b, bc)
// 		runLength >>= 7
// 	}
// 	return b
// }

// // https://en.wikipedia.org/wiki/Run-length_encoding
// func GetRLE_Bit(b []byte) (runLength uint32, value bool) {
// 	bc := b[0]
// 	b = b[1:]
// 	value = bc&0x40 > 0
// 	runLength |=

// 	return runLength, value
// }
