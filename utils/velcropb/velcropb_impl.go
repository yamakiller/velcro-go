package velcropb

import (
	"math"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"
)

var Impl defaultProtocol

var _ Protocol = defaultProtocol{}

const speculativeLength = 1

type defaultProtocol struct{}

// WriteMessage 实现 TLV(tag, length, value) and V(value)。
func (p defaultProtocol) WriteMessage(b []byte, number int32, writer Writer) (n int) {
	// 如果是要跳过的标记
	if number == SkipTagNumber {
		return writer.Write(b)
	}

	// TLV

	size := writer.Size()
	n += AppendTag(b, protowire.Number(number), protowire.BytesType)
	n += AppendVarint(b[n:], uint64(size))
	n += writer.Write(b[n:])
	return n
}

// WriteListPacked implements TLV(tag, length, value).
func (p defaultProtocol) WriteListPacked(buf []byte, number int32, length int, single Marshal) (n int) {
	n += AppendTag(buf, protowire.Number(number), protowire.BytesType)
	buf = buf[n:]

	prefix := speculativeLength
	offset := prefix
	for i := 0; i < length; i++ {
		offset += single(buf[offset:], SkipTagNumber, int32(i))
	}
	mlen := offset - prefix

	msiz := protowire.SizeVarint(uint64(mlen))
	if msiz != speculativeLength {
		copy(buf[msiz:], buf[prefix:prefix+mlen])
	}
	AppendVarint(buf[:msiz], uint64(mlen))

	n += msiz + mlen
	return n
}

// WriteMapEntry implements TLV(tag, length, value).
func (p defaultProtocol) WriteMapEntry(buf []byte, number int32, entry Marshal) (n int) {
	n += AppendTag(buf, protowire.Number(number), protowire.BytesType)
	buf = buf[n:]

	prefix := speculativeLength
	mlen := entry(buf[prefix:], MapEntryKeyFieldNumber, MapEntryValueFieldNumber)

	msiz := protowire.SizeVarint(uint64(mlen))
	if msiz != speculativeLength {
		copy(buf[msiz:], buf[prefix:prefix+mlen])
	}
	AppendVarint(buf[:msiz], uint64(mlen))

	n += msiz + mlen
	return n
}

// WriteBool implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteBool(b []byte, number int32, value bool) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], protowire.EncodeBool(value))
	return n
}

// WriteInt32 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteInt32(b []byte, number, value int32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], uint64(value))
	return n
}

// WriteInt64 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteInt64(b []byte, number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], uint64(value))
	return n
}

// WriteUint32 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteUint32(b []byte, number int32, value uint32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], uint64(value))
	return n
}

// WriteUint64 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteUint64(b []byte, number int32, value uint64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], value)
	return n
}

// WriteSint32 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteSint32(b []byte, number, value int32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], protowire.EncodeZigZag(int64(value)))
	return n
}

// WriteSint64 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteSint64(b []byte, number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.VarintType)
	}
	n += AppendVarint(b[n:], protowire.EncodeZigZag(value))
	return n
}

// WriteFloat implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteFloat(b []byte, number int32, value float32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed32Type)
	}
	n += AppendFixed32(b[n:], math.Float32bits(value))
	return n
}

// WriteDouble implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteDouble(b []byte, number int32, value float64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed64Type)
	}
	n += AppendFixed64(b[n:], math.Float64bits(value))
	return n
}

// WriteFixed32 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteFixed32(b []byte, number int32, value uint32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed32Type)
	}
	n += AppendFixed32(b[n:], value)
	return n
}

// WriteFixed64 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteFixed64(b []byte, number int32, value uint64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed64Type)
	}
	n += AppendFixed64(b[n:], value)
	return n
}

// WriteSfixed32 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteSfixed32(b []byte, number, value int32) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed32Type)
	}
	n += AppendFixed32(b[n:], uint32(value))
	return n
}

// WriteSfixed64 implements TV(tag, value) and V(value).
func (p defaultProtocol) WriteSfixed64(b []byte, number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.Fixed64Type)
	}
	n += AppendFixed64(b[n:], uint64(value))
	return n
}

// WriteString implements TLV(tag, length, value) and LV(length, value).
func (p defaultProtocol) WriteString(b []byte, number int32, value string) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.BytesType)
	}
	// only support proto3
	if EnforceUTF8() && !utf8.ValidString(value) {
		panic(errInvalidUTF8)
	}
	n += AppendString(b[n:], value)
	return n
}

// WriteBytes implements TLV(tag, length, value) and LV(length, value).
func (p defaultProtocol) WriteBytes(b []byte, number int32, value []byte) (n int) {
	if number != SkipTagNumber {
		n += AppendTag(b[n:], protowire.Number(number), protowire.BytesType)
	}
	n += AppendBytes(b[n:], value)
	return n
}

// ReadMessage .
func (p defaultProtocol) ReadMessage(b []byte, _type int8, reader Reader) (n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.BytesType {
		return 0, errUnknown
	}
	// framed
	if wtyp == protowire.BytesType {
		b, n = ConsumeBytes(b)
		if n < 0 {
			return 0, errDecode
		}
	}
	offset := 0
	for offset < len(b) {
		// Parse the tag (field number and wire type).
		num, wtyp, l := ConsumeTag(b[offset:])
		offset += l
		if l < 0 {
			return offset, errDecode
		}
		if num > protowire.MaxValidNumber {
			return offset, errDecode
		}
		l, err = reader.Read(b[offset:], int8(wtyp), int32(num))
		if err != nil {
			return offset, err
		}
		offset += l
	}
	// check if framed
	if n == 0 {
		n = offset
	}
	return n, nil
}

// ReadBool implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadBool(b []byte, _type int8) (value bool, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}

	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return protowire.DecodeBool(v), n, nil
}

// ReadInt32 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadInt32(b []byte, _type int8) (value int32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return int32(v), n, nil
}

// ReadInt64 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadInt64(b []byte, _type int8) (value int64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return int64(v), n, nil
}

// ReadUint32 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadUint32(b []byte, _type int8) (value uint32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return uint32(v), n, nil
}

// ReadUint64 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadUint64(b []byte, _type int8) (value uint64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return v, n, nil
}

// ReadSint32 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadSint32(b []byte, _type int8) (value int32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return int32(protowire.DecodeZigZag(v & math.MaxUint32)), n, nil
}

// ReadSint64 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadSint64(b []byte, _type int8) (value int64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.VarintType {
		return value, 0, errUnknown
	}
	v, n := ConsumeVarint(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return protowire.DecodeZigZag(v), n, nil
}

// ReadFloat implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadFloat(b []byte, _type int8) (value float32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed32Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed32(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return math.Float32frombits(uint32(v)), n, nil
}

// ReadDouble implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadDouble(b []byte, _type int8) (value float64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed64Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed64(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return math.Float64frombits(v), n, nil
}

// ReadFixed32 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadFixed32(b []byte, _type int8) (value uint32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed32Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed32(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return uint32(v), n, nil
}

// ReadFixed64 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadFixed64(b []byte, _type int8) (value uint64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed64Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed64(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return v, n, nil
}

// ReadSfixed32 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadSfixed32(b []byte, _type int8) (value int32, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed32Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed32(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return int32(v), n, nil
}

// ReadSfixed64 implements TV(tag, value) and V(value).
func (p defaultProtocol) ReadSfixed64(b []byte, _type int8) (value int64, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != SkipTypeCheck && wtyp != protowire.Fixed64Type {
		return value, 0, errUnknown
	}
	v, n := protowire.ConsumeFixed64(b)
	if n < 0 {
		return value, 0, errDecode
	}
	return int64(v), n, nil
}

// ReadString implements TLV(tag, length, value) and V(length, value).
func (p defaultProtocol) ReadString(b []byte, _type int8) (value string, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != protowire.BytesType {
		return value, 0, errUnknown
	}
	v, n := ConsumeBytes(b)
	if n < 0 {
		return value, 0, errDecode
	}
	// only support proto3
	if EnforceUTF8() && !utf8.Valid(v) {
		return value, 0, errInvalidUTF8
	}
	return string(v), n, nil
}

// ReadBytes .
func (p defaultProtocol) ReadBytes(b []byte, _type int8) (value []byte, n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp != protowire.BytesType {
		return value, 0, errUnknown
	}
	v, n := ConsumeBytes(b)
	if n < 0 {
		return value, 0, errDecode
	}
	value = make([]byte, len(v))
	copy(value, v)
	return value, n, nil
}

// ReadList .
func (p defaultProtocol) ReadList(b []byte, _type int8, single Unmarshal) (n int, err error) {
	wtyp := protowire.Type(_type)
	if wtyp == protowire.BytesType {
		var framed []byte
		framed, n = ConsumeBytes(b)
		if n < 0 {
			return 0, errDecode
		}
		for len(framed) > 0 {
			off, err := single(framed, SkipTypeCheck)
			if err != nil {
				return 0, err
			}
			framed = framed[off:]
		}
		return n, nil
	}
	n, err = single(b, _type)
	return n, err
}

// ReadMapEntry .
func (p defaultProtocol) ReadMapEntry(b []byte, _type int8, umk, umv Unmarshal) (int, error) {
	offset := 0
	wtyp := protowire.Type(_type)
	if wtyp != protowire.BytesType {
		return 0, errUnknown
	}
	bs, n := ConsumeBytes(b)
	if n < 0 {
		return 0, errDecode
	}
	offset = n

	// Map entries are represented as a two-element message with fields
	// containing the key and value.
	for len(bs) > 0 {
		num, wtyp, n := ConsumeTag(bs)
		if n < 0 {
			return 0, errDecode
		}
		if num > protowire.MaxValidNumber {
			return 0, errDecode
		}
		bs = bs[n:]
		err := errUnknown
		switch num {
		case MapEntryKeyFieldNumber:
			n, err = umk(bs, int8(wtyp))
		case MapEntryValueFieldNumber:
			n, err = umv(bs, int8(wtyp))
		}
		if err == errUnknown {
			n, err = p.Skip(b, int8(wtyp), int32(num))
		}
		if err != nil && err != errUnknown {
			return 0, err
		}
		bs = bs[n:]
	}
	return offset, nil
}

// Skip .
func (p defaultProtocol) Skip(b []byte, _type int8, number int32) (n int, err error) {
	n = protowire.ConsumeFieldValue(protowire.Number(number), protowire.Type(_type), b)
	if n < 0 {
		return 0, errDecode
	}
	return n, nil
}

// SizeBool .
func (p defaultProtocol) SizeBool(number int32, value bool) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(protowire.EncodeBool(value))
	return n
}

// SizeInt32 .
func (p defaultProtocol) SizeInt32(number, value int32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(uint64(value))
	return n
}

// SizeInt64 .
func (p defaultProtocol) SizeInt64(number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(uint64(value))
	return n
}

// SizeUint32 .
func (p defaultProtocol) SizeUint32(number int32, value uint32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(uint64(value))
	return n
}

// SizeUint64 .
func (p defaultProtocol) SizeUint64(number int32, value uint64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(value)
	return n
}

// SizeSint32 .
func (p defaultProtocol) SizeSint32(number, value int32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(protowire.EncodeZigZag(int64(value)))
	return n
}

// SizeSint64 .
func (p defaultProtocol) SizeSint64(number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.VarintType))
	}
	n += protowire.SizeVarint(protowire.EncodeZigZag(value))
	return n
}

// SizeFloat .
func (p defaultProtocol) SizeFloat(number int32, value float32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed32Type))
	}
	n += protowire.SizeFixed32()
	return n
}

// SizeDouble .
func (p defaultProtocol) SizeDouble(number int32, value float64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed64Type))
	}
	n += protowire.SizeFixed64()
	return n
}

// SizeFixed32 .
func (p defaultProtocol) SizeFixed32(number int32, value uint32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed32Type))
	}
	n += protowire.SizeFixed32()
	return n
}

// SizeFixed64 .
func (p defaultProtocol) SizeFixed64(number int32, value uint64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed64Type))
	}
	n += protowire.SizeFixed64()
	return n
}

// SizeSfixed32 .
func (p defaultProtocol) SizeSfixed32(number, value int32) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed32Type))
	}
	n += protowire.SizeFixed32()
	return n
}

// SizeSfixed64 .
func (p defaultProtocol) SizeSfixed64(number int32, value int64) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.Fixed64Type))
	}
	n += protowire.SizeFixed64()
	return n
}

// SizeString .
func (p defaultProtocol) SizeString(number int32, value string) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.BytesType))
	}
	n += protowire.SizeBytes(len(value))
	return n
}

// SizeBytes .
func (p defaultProtocol) SizeBytes(number int32, value []byte) (n int) {
	if number != SkipTagNumber {
		n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.BytesType))
	}
	n += protowire.SizeBytes(len(value))
	return n
}

// SizeMessage .
func (p defaultProtocol) SizeMessage(number int32, sizer Sizer) (n int) {
	if number == SkipTagNumber {
		return sizer.Size()
	}
	n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.BytesType))
	n += protowire.SizeBytes(sizer.Size())
	return n
}

// SizeListPacked .
func (p defaultProtocol) SizeListPacked(number int32, length int, single EntrySize) (n int) {
	n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.BytesType))

	mlen := 0
	for i := 0; i < length; i++ {
		// append mlen
		mlen += single(SkipTagNumber, int32(i))
	}
	n += protowire.SizeBytes(mlen)
	return n
}

// SizeMapEntry .
func (p defaultProtocol) SizeMapEntry(number int32, entry EntrySize) (n int) {
	n += protowire.SizeVarint(protowire.EncodeTag(protowire.Number(number), protowire.BytesType))
	mlen := entry(MapEntryKeyFieldNumber, MapEntryValueFieldNumber)
	n += protowire.SizeBytes(mlen)
	return n
}
