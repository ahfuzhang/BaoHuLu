// Package utils provides low-level protobuf binary encoding/decoding primitives
package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"unsafe"
)

// ─── ConsumeVarint error codes ────────────────────────────────────────────────

// Pre-allocated sentinel errors for the two varint error cases.
// Using sentinel values avoids a heap allocation on every error path.
var (
	errVarintOverflow = errors.New("varint overflow")
	errVarintEOF      = errors.New("unexpected EOF reading varint")
)

// consumeVarintError converts a non-zero ConsumeVarint returnValue to an error.
// code 1 → varint overflow; code 2 (or any other) → unexpected EOF.
func consumeVarintError(code int64) error {
	if code == 1 {
		return errVarintOverflow
	}
	return errVarintEOF
}

// ─── Wire types ───────────────────────────────────────────────────────────────

// WireType enumerates the protobuf binary wire-type encodings.
type WireType int

const (
	WireTypeVarint     WireType = 0 // int32, int64, uint32, uint64, sint32, sint64, bool, enum
	WireType64bit      WireType = 1 // fixed64, sfixed64, double
	WireTypeLenDelim   WireType = 2 // string, bytes, embedded messages, packed repeated, map
	WireTypeStartGroup WireType = 3 // deprecated group start
	WireTypeEndGroup   WireType = 4 // deprecated group end
	WireType32bit      WireType = 5 // fixed32, sfixed32, float
)

// ─── Varint write ─────────────────────────────────────────────────────────────

// AppendVarint encodes v as a protobuf varint and appends it to b.
// VarintSize computes the byte count via a single LZCNT/BSR instruction (no comparisons).
// The switch dispatches on that integer value, which the compiler lowers to a jump table.
// todo: 这个函数的性能仍然有优化空间  => 使用 *[10]byte 来优化
func AppendVarint(b []byte, v uint64) []byte {
	// fast path
	if v < 0x80 {
		return append(b, byte(v))
	}
	if v < 0x4000 {
		return append(b,
			byte(v)|0x80,
			byte(v>>7))
	}
	n := (bits.Len64(v|1) + 6) / 7
	/*
		// 未发现明显的效果
		if cap(b)-len(b) >= n {
			// enought space
			b = b[:len(b)+n]
			_ = EncodeVarint(b, len(b), v)
			return b
		}
	*/
	switch n {
	case 0:
		return nil
	case 1:
		return append(b, byte(v))
	case 2:
		return append(b,
			byte(v)|0x80,
			byte(v>>7))
	case 3:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14))
	case 4:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21))
	case 5:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28))
	case 6:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28)|0x80,
			byte(v>>35))
	case 7:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28)|0x80,
			byte(v>>35)|0x80,
			byte(v>>42))
	case 8:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28)|0x80,
			byte(v>>35)|0x80,
			byte(v>>42)|0x80,
			byte(v>>49))
	case 9:
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28)|0x80,
			byte(v>>35)|0x80,
			byte(v>>42)|0x80,
			byte(v>>49)|0x80,
			byte(v>>56))
	case 10: // 10 bytes (full uint64)
		return append(b,
			byte(v)|0x80,
			byte(v>>7)|0x80,
			byte(v>>14)|0x80,
			byte(v>>21)|0x80,
			byte(v>>28)|0x80,
			byte(v>>35)|0x80,
			byte(v>>42)|0x80,
			byte(v>>49)|0x80,
			byte(v>>56)|0x80,
			byte(v>>63))
	default:
		return nil
	}
}

func EncodeVarint(dAtA []byte, offset int, v uint64) int {
	// hot path
	if v < 0x80 {
		offset--
		var p *byte = (*byte)(unsafe.Pointer(&dAtA[offset]))
		*p = uint8(v)
		return offset
	}
	if v < 0x4000 {
		offset -= 2
		var arr *[2]byte = (*[2]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v >> 7)
		return offset
	}
	n := (bits.Len64(v|1) + 6) / 7
	offset -= n
	switch n {
	case 0:
		return -10 // impossible branch. for cheat compiler to build jump table
	case 1:
		return -11 // impossible branch. for cheat compiler to build jump table
	case 2:
		return -12 // impossible branch. for cheat compiler to build jump table
	case 3:
		var arr *[3]byte = (*[3]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v >> 14)
		return offset
	case 4:
		var arr *[4]byte = (*[4]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v >> 21)
		return offset
	case 5:
		var arr *[5]byte = (*[5]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v >> 28)
		return offset
	case 6:
		var arr *[6]byte = (*[6]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v>>28) | 0x80
		arr[5] = uint8(v >> 35)
		return offset
	case 7:
		var arr *[7]byte = (*[7]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v>>28) | 0x80
		arr[5] = uint8(v>>35) | 0x80
		arr[6] = uint8(v >> 42)
		return offset
	case 8:
		var arr *[8]byte = (*[8]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v>>28) | 0x80
		arr[5] = uint8(v>>35) | 0x80
		arr[6] = uint8(v>>42) | 0x80
		arr[7] = uint8(v >> 49)
		return offset
	case 9:
		var arr *[9]byte = (*[9]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v>>28) | 0x80
		arr[5] = uint8(v>>35) | 0x80
		arr[6] = uint8(v>>42) | 0x80
		arr[7] = uint8(v>>49) | 0x80
		arr[8] = uint8(v >> 56)
		return offset
	case 10: // case 10: bit 63 set
		var arr *[10]byte = (*[10]byte)(unsafe.Pointer(&dAtA[offset]))
		arr[0] = uint8(v) | 0x80
		arr[1] = uint8(v>>7) | 0x80
		arr[2] = uint8(v>>14) | 0x80
		arr[3] = uint8(v>>21) | 0x80
		arr[4] = uint8(v>>28) | 0x80
		arr[5] = uint8(v>>35) | 0x80
		arr[6] = uint8(v>>42) | 0x80
		arr[7] = uint8(v>>49) | 0x80
		arr[8] = uint8(v>>56) | 0x80
		arr[9] = uint8(v >> 63)
		return offset
	default:
		return -13 // impossible branch. for cheat compiler to build jump table
	}
}

// AppendTag encodes a protobuf field tag (field number + wire type) and appends it to b.
func AppendTag(b []byte, fieldNum int, wt WireType) []byte {
	return AppendVarint(b, uint64(fieldNum)<<3|uint64(wt))
}

// VarintSize returns the number of bytes needed to encode v as a protobuf varint.
// Uses bits.Len64 (compiles to a single LZCNT/BSR instruction on amd64) to avoid
// sequential comparisons.
func VarintSize(v uint64) int {
	return (bits.Len64(v|1) + 6) / 7
}

// TagSize returns the number of bytes needed to encode a field tag.
func TagSize(fieldNum int, wt WireType) int {
	return VarintSize(uint64(fieldNum)<<3 | uint64(wt))
}

// AppendSint32 zigzag-encodes v and appends the varint to b.
func AppendSint32(b []byte, v int32) []byte {
	uv := (uint32(v) << 1) ^ uint32(v>>31)
	return AppendVarint(b, uint64(uv))
}

// AppendSint64 zigzag-encodes v and appends the varint to b.
func AppendSint64(b []byte, v int64) []byte {
	uv := (uint64(v) << 1) ^ uint64(v>>63)
	return AppendVarint(b, uv)
}

// AppendFixed32 appends v as 4 little-endian bytes.
func AppendFixed32(b []byte, v uint32) []byte {
	return append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// AppendFixed64 appends v as 8 little-endian bytes.
func AppendFixed64(b []byte, v uint64) []byte {
	return append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// AppendLenDelim appends a length-prefixed byte slice to b.
func AppendLenDelim(b []byte, data []byte) []byte {
	b = AppendVarint(b, uint64(len(data)))
	return append(b, data...)
}

// ConsumeTag reads a field tag (field number + wire type) from b.
func ConsumeTag(b []byte) (fieldNum int, wt WireType, rest []byte, err error) {
	c1 := b[0]
	wt = WireType(c1 & 0x7)
	if c1 < 0x80 {
		fieldNum = int(c1) >> 3
		rest = b[1:]
		return
	}
	if len(b) > 1 && b[1] < 0x80 {
		fieldNum = ((int(c1) >> 3) & 15) | (int(b[1]) << 4)
		rest = b[2:]
		return
	}
	// slow path
	var x uint64
	var s uint
	for i, c := range b {
		if i == 10 {
			err = fmt.Errorf("varint overflow")
			return
		}
		if c < 0x80 {
			x |= uint64(c) << s
			fieldNum = int(x >> 3)
			rest = b[i+1:]
			return
		}
		x |= uint64(c&0x7f) << s
		s += 7
	}
	err = fmt.Errorf("unexpected EOF reading varint")
	return
}

func ConsumeBytes(b []byte) (data []byte, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	if uint64(len(rest)) < v {
		return nil, b, fmt.Errorf("not enough bytes: need %d have %d", v, len(rest))
	}
	data = rest[:v]
	rest = rest[v:]
	return
}

// ConsumeBytes reads a length-delimited byte slice from b.
func ConsumeBytesV1(b []byte) ([]byte, []byte, error) {
	l := uint64(b[0])
	if l < 0x80 {
		// fast path
		rest := b[1:]
		if uint64(len(rest)) < l {
			return nil, b, fmt.Errorf("not enough bytes: need %d have %d", l, len(rest))
		}
		return rest[:l], rest[l:], nil
	}
	if len(b) > 1 && b[1] < 0x80 {
		l = (l & 0x7f) | (uint64(b[1]) << 7)
		rest := b[2:]
		if uint64(len(rest)) < l {
			return nil, b, fmt.Errorf("not enough bytes: need %d have %d", l, len(rest))
		}
		return rest[:l], rest[l:], nil
	}
	// slow path
	l = 0
	var s uint
	for i, c := range b {
		if i == 10 {
			return nil, nil, fmt.Errorf("varint overflow")
		}
		if c < 0x80 {
			l |= uint64(c) << s
			rest := b[i+1:]
			if uint64(len(rest)) < l {
				return nil, b, fmt.Errorf("not enough bytes: need %d have %d", l, len(rest))
			}
			return rest[:l], rest[l:], nil
		}
		l |= uint64(c&0x7f) << s
		s += 7
	}
	return nil, nil, fmt.Errorf("unexpected EOF reading varint")
}

// SkipField advances past a single field value of the given wire type.
func SkipField(wt WireType, b []byte) ([]byte, error) {
	switch wt {
	case WireTypeVarint:
		for len(b) > 0 {
			c := b[0]
			b = b[1:]
			if c < 0x80 {
				return b, nil
			}
		}
		return b, fmt.Errorf("EOF in varint skip")
	case WireType64bit:
		if len(b) < 8 {
			return b, fmt.Errorf("EOF in 64-bit skip")
		}
		return b[8:], nil
	case WireTypeLenDelim:
		_, rest, err := ConsumeBytes(b)
		return rest, err
	case WireType32bit:
		if len(b) < 4 {
			return b, fmt.Errorf("EOF in 32-bit skip")
		}
		return b[4:], nil
	}
	return b, fmt.Errorf("unknown wire type %d", wt)
}

// ─── Scalar read ──────────────────────────────────────────────────────────────

func ReadInt32V0(b []byte) (int32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return int32(v), rest, nil
}

func ReadInt32(b []byte) (n int32, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	n = int32(v)
	return
}

func ReadInt32V1(b []byte) (int32, []byte, error) {
	var v int32
	if b[0] < 0x80 {
		return int32(b[0]), b[1:], nil
	}
	if len(b) > 1 && b[1] < 0x80 {
		v = int32(b[0]&0x7F) | (int32(b[1]) << 7)
		return v, b[2:], nil
	}
	if len(b) > 2 && b[2] < 0x80 {
		v = int32(b[0]&0x7F) | (int32(b[1]&0x7F) << 7) | (int32(b[2]) << 14)
		return v, b[3:], nil
	}
	if len(b) > 3 && b[3] < 0x80 {
		v = int32(b[0]&0x7F) | (int32(b[1]&0x7F) << 7) | (int32(b[2]&0x7F) << 14) | (int32(b[3]) << 21)
		return v, b[4:], nil
	}
	if len(b) > 4 && b[4] < 0x80 {
		v = int32(b[0]&0x7F) | (int32(b[1]&0x7F) << 7) | (int32(b[2]&0x7F) << 14) | (int32(b[3]&0x7F) << 21) | (int32(b[4]) << 28)
		return v, b[5:], nil
	}
	fmt.Printf("%02X %02X %02X %02X %02X %02X %02X %02X %02X %02X \n", b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7], b[8], b[9])
	return 0, nil, errVarintOverflow
}

func ReadInt64V1(b []byte) (int64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return int64(v), rest, nil
}

func ReadInt64(b []byte) (n int64, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	n = int64(v)
	return
}

func ReadUint32V1(b []byte) (uint32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return uint32(v), rest, nil
}

func ReadUint32(b []byte) (n uint32, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	n = uint32(v)
	return
}

func ReadUint64V1(b []byte) (uint64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return v, rest, nil
}

func ReadUint64(b []byte) (n uint64, rest []byte, err error) {
	n, rest, err = ReadVarint(b)
	return
}

func ReadSint32V1(b []byte) (int32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	n := int32((uint32(v) >> 1) ^ -(uint32(v) & 1))
	return n, rest, nil
}

func ReadSint32(b []byte) (n int32, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	n = int32((uint32(v) >> 1) ^ -(uint32(v) & 1))
	return
}

func ReadSint64V1(b []byte) (int64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	n := int64((v >> 1) ^ -(v & 1))
	return n, rest, nil
}

func ReadSint64(b []byte) (n int64, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	if err != nil {
		return
	}
	n = int64((v >> 1) ^ -(v & 1))
	return
}

func ReadBoolV1(b []byte) (bool, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return false, b, consumeVarintError(code)
	}
	return v != 0, rest, nil
}

func ReadBool(b []byte) (ret bool, rest []byte, err error) {
	var v uint64
	v, rest, err = ReadVarint(b)
	// if err != nil {
	// 	return
	// }
	ret = v != 0
	return
}

func ReadFixed32(b []byte) (uint32, []byte, error) {
	if len(b) < 4 {
		return 0, b, fmt.Errorf("EOF reading fixed32")
	}
	return binary.LittleEndian.Uint32(b), b[4:], nil
}

func ReadFixed64(b []byte) (uint64, []byte, error) {
	if len(b) < 8 {
		return 0, b, fmt.Errorf("EOF reading fixed64")
	}
	// todo: 直接使用位运算会不会更快?
	return binary.LittleEndian.Uint64(b), b[8:], nil
}

func ReadSfixed32(b []byte) (int32, []byte, error) {
	v, rest, err := ReadFixed32(b)
	return int32(v), rest, err
}

func ReadSfixed64(b []byte) (int64, []byte, error) {
	v, rest, err := ReadFixed64(b)
	return int64(v), rest, err
}

func ReadFloat(b []byte) (float32, []byte, error) {
	v, rest, err := ReadFixed32(b)
	return math.Float32frombits(v), rest, err
}

func ReadDouble(b []byte) (float64, []byte, error) {
	v, rest, err := ReadFixed64(b)
	return math.Float64frombits(v), rest, err
}

// ReadString returns a string backed by the same memory as b (zero-copy).
func ReadString(b []byte) (string, []byte, error) {
	data, rest, err := ConsumeBytes(b)
	if err != nil { // todo: 内部的判断都可以去掉
		return "", b, err
	}
	if len(data) == 0 {
		return "", rest, nil
	}
	return unsafe.String(unsafe.SliceData(data), len(data)), rest, nil
}

func ReadBytes(b []byte) ([]byte, []byte, error) {
	return ConsumeBytes(b)
}

func ReadVarint(b []byte) (v uint64, rest []byte, err error) {
	if b[0] < 0x80 {
		v = uint64(b[0])
		rest = b[1:]
		return
	}
	if len(b) >= 10 {
		if b[1] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]) << 7)
			rest = b[2:]
			return
		}
		if b[2] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]) << 14)
			rest = b[3:]
			return
		}
		if b[3] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]) << 21)
			rest = b[4:]
			return
		}
		if b[4] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]) << 28)
			rest = b[5:]
			return
		}
		if b[5] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]&0x7F) << 28) |
				(uint64(b[5]) << 35)
			rest = b[6:]
			return
		}
		if b[6] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]&0x7F) << 28) |
				(uint64(b[5]&0x7F) << 35) |
				(uint64(b[6]) << 42)
			rest = b[7:]
			return
		}
		if b[7] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]&0x7F) << 28) |
				(uint64(b[5]&0x7F) << 35) |
				(uint64(b[6]&0x7F) << 42) |
				(uint64(b[7]) << 49)
			rest = b[8:]
			return
		}
		if b[8] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]&0x7F) << 28) |
				(uint64(b[5]&0x7F) << 35) |
				(uint64(b[6]&0x7F) << 42) |
				(uint64(b[7]&0x7F) << 49) |
				(uint64(b[8]) << 56)
			rest = b[9:]
			return
		}
		if b[9] < 0x80 {
			v = uint64(b[0]&0x7F) |
				(uint64(b[1]&0x7F) << 7) |
				(uint64(b[2]&0x7F) << 14) |
				(uint64(b[3]&0x7F) << 21) |
				(uint64(b[4]&0x7F) << 28) |
				(uint64(b[5]&0x7F) << 35) |
				(uint64(b[6]&0x7F) << 42) |
				(uint64(b[7]&0x7F) << 49) |
				(uint64(b[8]&0x7F) << 56) |
				(uint64(b[9]) << 63)
			rest = b[10:]
			return
		}
		err = errVarintEOF
		return
	}
	if len(b) > 1 && b[1] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]) << 7)
		rest = b[2:]
		return
	}
	if len(b) > 2 && b[2] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]) << 14)
		rest = b[3:]
		return
	}
	if len(b) > 3 && b[3] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]) << 21)
		rest = b[4:]
		return
	}
	if len(b) > 4 && b[4] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]) << 28)
		rest = b[5:]
		return
	}
	if len(b) > 5 && b[5] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]&0x7F) << 28) |
			(uint64(b[5]) << 35)
		rest = b[6:]
		return
	}
	if len(b) > 6 && b[6] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]&0x7F) << 28) |
			(uint64(b[5]&0x7F) << 35) |
			(uint64(b[6]) << 42)
		rest = b[7:]
		return
	}
	if len(b) > 7 && b[7] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]&0x7F) << 28) |
			(uint64(b[5]&0x7F) << 35) |
			(uint64(b[6]&0x7F) << 42) |
			(uint64(b[7]) << 49)
		rest = b[8:]
		return
	}
	if len(b) > 8 && b[8] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]&0x7F) << 28) |
			(uint64(b[5]&0x7F) << 35) |
			(uint64(b[6]&0x7F) << 42) |
			(uint64(b[7]&0x7F) << 49) |
			(uint64(b[8]) << 56)
		rest = b[9:]
		return
	}
	if len(b) > 9 && b[9] < 0x80 {
		v = uint64(b[0]&0x7F) |
			(uint64(b[1]&0x7F) << 7) |
			(uint64(b[2]&0x7F) << 14) |
			(uint64(b[3]&0x7F) << 21) |
			(uint64(b[4]&0x7F) << 28) |
			(uint64(b[5]&0x7F) << 35) |
			(uint64(b[6]&0x7F) << 42) |
			(uint64(b[7]&0x7F) << 49) |
			(uint64(b[8]&0x7F) << 56) |
			(uint64(b[9]) << 63)
		rest = b[10:]
		return
	}
	err = errVarintEOF
	return
}

// UnsafeBytesFromString returns a []byte view of s without copying.
// The returned slice aliases s's backing storage and must be treated as read-only.
func UnsafeBytesFromString(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// EncodeJSONString appends a JSON-safe escaped version of s into dst.
// No heap allocations are performed.
func EncodeJSONString(s string, dst []byte) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	needed := len(s) * 6
	if cap(dst)-len(dst) < needed {
		tmp := make([]byte, len(dst), len(dst)+needed+256)
		copy(tmp, dst)
		dst = tmp
	}
	offset := len(dst)
	d := dst[offset : offset+needed]
	ii := 0
	for i := 0; i < len(b); i++ {
		c := b[i]
		if (escapeTable[c/8] & (1 << (c % 8))) == 0 {
			// fast path
			d[ii] = c
			ii++
			continue
		}
		switch c {
		case '"':
			d[ii+1] = '"'
			d[ii] = '\\'
			ii += 2
		case '\n':
			d[ii+1] = 'n'
			d[ii] = '\\'
			ii += 2
		case '\t':
			d[ii+1] = 't'
			d[ii] = '\\'
			ii += 2
		case '\r':
			d[ii+1] = 'r'
			d[ii] = '\\'
			ii += 2
		case '\\':
			d[ii+1] = '\\'
			d[ii] = '\\'
			ii += 2
		case '\f':
			d[ii+1] = 'f'
			d[ii] = '\\'
			ii += 2
		case '\b':
			d[ii+1] = 'b'
			d[ii] = '\\'
			ii += 2
		default:
			// control characters must be escaped as \uXXXX per JSON spec
			d[ii+5] = hex[c&0xf]
			d[ii+4] = hex[(c>>4)&0xf]
			d[ii+3] = '0'
			d[ii+2] = '0'
			d[ii+1] = 'u'
			d[ii] = '\\'
			ii += 6
		}
	}
	return dst[:offset+ii]
}

const hex = "0123456789abcdef"

var escapeTable [256 / 8]byte = func() [256 / 8]byte {
	var table [256 / 8]byte
	for i := 0; i < 256; i++ {
		if i < 0x20 || i == '"' || i == '\\' {
			table[i/8] |= 1 << (i % 8)
		}
	}
	return table
}()

func ConsumeVarint(b []byte) (uint64, []byte, int64) {
	// todo: 加速 1-2 字节的解码性能
	/*
		// 看起来是负优化
		if b[0] < 0x80 {
			// fast path
			return uint64(b[0]), b[1:], 0
		}
		if len(b) > 1 && b[1] < 0x80 {
			return (uint64(b[0]) & 0x7F) | (uint64(b[1]) << 7), b[2:], 0
		}
	*/
	var x uint64
	var s uint
	for i, c := range b {
		if i == 10 {
			return 0, b, 1
		}
		if c < 0x80 {
			x |= uint64(c) << s
			return x, b[i+1:], 0
		}
		x |= uint64(c&0x7f) << s
		s += 7
	}
	return 0, b, 2
}

func ConsumeVarintV1(b []byte) (v uint64, rest []byte, err error) {
	// todo: 加速 1-2 字节的解码性能
	/*
		// 看起来是负优化
		if b[0] < 0x80 {
			// fast path
			return uint64(b[0]), b[1:], 0
		}
		if len(b) > 1 && b[1] < 0x80 {
			return (uint64(b[0]) & 0x7F) | (uint64(b[1]) << 7), b[2:], 0
		}
	*/
	var x uint64
	var s uint
	for i, c := range b {
		if i == 10 {
			return 0, b, errVarintEOF
		}
		if c < 0x80 {
			x |= uint64(c) << s
			return x, b[i+1:], nil
		}
		x |= uint64(c&0x7f) << s
		s += 7
	}
	return 0, b, errVarintOverflow
}
