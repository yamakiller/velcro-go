package velcropb

import (
	"errors"

	"google.golang.org/protobuf/encoding/protowire"
)

// Generic field names and numbers for synthetic map entry messages.
const (
	MapEntryKeyFieldNumber   = 1  // protoreflect.FieldNumber
	MapEntryValueFieldNumber = 2  // protoreflect.FieldNumber
	SkipTagNumber            = -1 // protowire.Number
	SkipTypeCheck            = -1 // protowire.Type
)

var errUnknown = errors.New("BUG: internal error (unknown)")

var errDecode = errors.New("cannot parse invalid wire-format data")

var errInvalidUTF8 = errors.New("field contains invalid UTF-8")

func ConsumeTag(b []byte) (protowire.Number, protowire.Type, int) {
	v, n := ConsumeVarint(b)
	if n < 0 {
		return 0, 0, n // forward error code
	}
	num, typ := protowire.DecodeTag(v)
	if num < protowire.MinValidNumber {
		return 0, 0, -2
	}
	return num, typ, n
}

// ConsumeVarint parses b as a varint-encoded uint64, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeVarint(b []byte) (v uint64, n int) {
	var y uint64
	if len(b) <= 0 {
		return 0, -1
	}
	v = uint64(b[0])
	if v < 0x80 {
		return v, 1
	}
	v -= 0x80

	if len(b) <= 1 {
		return 0, -1
	}
	y = uint64(b[1])
	v += y << 7
	if y < 0x80 {
		return v, 2
	}
	v -= 0x80 << 7

	if len(b) <= 2 {
		return 0, -1
	}
	y = uint64(b[2])
	v += y << 14
	if y < 0x80 {
		return v, 3
	}
	v -= 0x80 << 14

	if len(b) <= 3 {
		return 0, -1
	}
	y = uint64(b[3])
	v += y << 21
	if y < 0x80 {
		return v, 4
	}
	v -= 0x80 << 21

	if len(b) <= 4 {
		return 0, -1
	}
	y = uint64(b[4])
	v += y << 28
	if y < 0x80 {
		return v, 5
	}
	v -= 0x80 << 28

	if len(b) <= 5 {
		return 0, -1
	}
	y = uint64(b[5])
	v += y << 35
	if y < 0x80 {
		return v, 6
	}
	v -= 0x80 << 35

	if len(b) <= 6 {
		return 0, -1
	}
	y = uint64(b[6])
	v += y << 42
	if y < 0x80 {
		return v, 7
	}
	v -= 0x80 << 42

	if len(b) <= 7 {
		return 0, -1
	}
	y = uint64(b[7])
	v += y << 49
	if y < 0x80 {
		return v, 8
	}
	v -= 0x80 << 49

	if len(b) <= 8 {
		return 0, -1
	}
	y = uint64(b[8])
	v += y << 56
	if y < 0x80 {
		return v, 9
	}
	v -= 0x80 << 56

	if len(b) <= 9 {
		return 0, -1
	}
	y = uint64(b[9])
	v += y << 63
	if y < 2 {
		return v, 10
	}
	return 0, -3
}

// ConsumeBytes parses b as a length-prefixed bytes value, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeBytes(b []byte) (v []byte, total int) {
	m, n := ConsumeVarint(b)
	total = int(m) + n
	if n < 0 || total > len(b) {
		return nil, -1 // forward error code
	}
	return b[n:total], total
}

func AppendTag(b []byte, num protowire.Number, typ protowire.Type) int {
	return AppendVarint(b, protowire.EncodeTag(num, typ))
}

func AppendVarint(b []byte, v uint64) int {
	switch {
	case v < 1<<7:
		b[0] = uint8(v)
		return 1
	case v < 1<<14:
		b[0] = uint8(v&0x7f | 0x80)
		b[1] = uint8(v >> 7)
		return 2
	case v < 1<<21:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8(v >> 14)
		return 3
	case v < 1<<28:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8(v >> 21)
		return 4
	case v < 1<<35:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8(v >> 28)
		return 5
	case v < 1<<42:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8((v>>28)&0x7f | 0x80)
		b[5] = uint8(v >> 35)
		return 6
	case v < 1<<49:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8((v>>28)&0x7f | 0x80)
		b[5] = uint8((v>>35)&0x7f | 0x80)
		b[6] = uint8(v >> 42)
		return 7
	case v < 1<<56:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8((v>>28)&0x7f | 0x80)
		b[5] = uint8((v>>35)&0x7f | 0x80)
		b[6] = uint8((v>>42)&0x7f | 0x80)
		b[7] = uint8(v >> 49)
		return 8
	case v < 1<<63:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8((v>>28)&0x7f | 0x80)
		b[5] = uint8((v>>35)&0x7f | 0x80)
		b[6] = uint8((v>>42)&0x7f | 0x80)
		b[7] = uint8((v>>49)&0x7f | 0x80)
		b[8] = uint8(v >> 56)
		return 9
	default:
		b[0] = uint8((v>>0)&0x7f | 0x80)
		b[1] = uint8((v>>7)&0x7f | 0x80)
		b[2] = uint8((v>>14)&0x7f | 0x80)
		b[3] = uint8((v>>21)&0x7f | 0x80)
		b[4] = uint8((v>>28)&0x7f | 0x80)
		b[5] = uint8((v>>35)&0x7f | 0x80)
		b[6] = uint8((v>>42)&0x7f | 0x80)
		b[7] = uint8((v>>49)&0x7f | 0x80)
		b[8] = uint8((v>>56)&0x7f | 0x80)
		b[9] = 1
		return 10
	}
}

// AppendFixed32 appends v to b as a little-endian uint32.
func AppendFixed32(b []byte, v uint32) int {
	_ = b[3]
	b[0] = byte(v >> 0)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	return 4
}

// AppendFixed64 appends v to b as a little-endian uint64.
func AppendFixed64(b []byte, v uint64) int {
	_ = b[7]
	b[0] = byte(v >> 0)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return 8
}

// AppendBytes appends v to b as a length-prefixed bytes value.
func AppendBytes(b, v []byte) (n int) {
	n += AppendVarint(b, uint64(len(v)))
	n += copy(b[n:], v)
	return n
}

// AppendString appends v to b as a length-prefixed bytes value.
func AppendString(b []byte, v string) (n int) {
	n += AppendVarint(b, uint64(len(v)))
	n += copy(b[n:], v)
	return n
}

// EnforceUTF8 todo: use as a switch.
func EnforceUTF8() bool {
	return false
}
