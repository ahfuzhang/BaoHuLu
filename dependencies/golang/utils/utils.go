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
	WireTypeVarint   WireType = 0 // int32, int64, uint32, uint64, sint32, sint64, bool, enum
	WireType64bit    WireType = 1 // fixed64, sfixed64, double
	WireTypeLenDelim WireType = 2 // string, bytes, embedded messages, packed repeated, map
	WireType32bit    WireType = 5 // fixed32, sfixed32, float
)

// ─── Varint write ─────────────────────────────────────────────────────────────

// AppendVarint encodes v as a protobuf varint and appends it to b.
// VarintSize computes the byte count via a single LZCNT/BSR instruction (no comparisons).
// The switch dispatches on that integer value, which the compiler lowers to a jump table.
func AppendVarint(b []byte, v uint64) []byte {
	switch VarintSize(v) {
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
	default: // 10 bytes (full uint64)
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

// ─── Varint read ──────────────────────────────────────────────────────────────

// ConsumeVarint reads a varint from b and returns the value and remaining bytes.
//
// Fast path (len(b) >= 9, little-endian platforms): loads 8 bytes as a single
// uint64, ANDs with 0x8080808080808080 to collect the MSB of every byte, then
// uses bits.TrailingZeros64 to locate the terminating byte in O(1).  A
// switch/case on that length lets the compiler emit a jump table — no per-byte
// loop in the common case.
//
// Slow path (len(b) < 9): falls back to the original byte-by-byte loop.
// 这个方法是负优化
func ConsumeVarintV2(b []byte) (uint64, []byte, error) {
	if len(b) < 9 {
		return consumeVarintSlow(b)
	}

	// ptr is a fixed-size array pointer covering the first 9 bytes.
	// The compiler knows ptr[0]…ptr[8] are always in bounds, eliminating
	// per-access slice bound checks throughout the fast path.
	var ptr *[9]byte = (*[9]byte)(unsafe.Pointer(&b[0]))

	// Load 8 bytes as a little-endian uint64 (safe on amd64/arm64).
	// Each byte's MSB sits at bit positions 7, 15, 23, … 63 in the word.
	// Masking with 0x8080808080808080 isolates all eight continuation bits.
	w := *(*uint64)(unsafe.Pointer(ptr))
	m := w & 0x8080808080808080

	if m != 0x8080808080808080 {
		// At least one of the first 8 bytes is a terminator (MSB == 0).
		// ^m & mask has bit (8k+7) set for each such byte k.
		// TrailingZeros64/8 converts that bit position to a byte index.
		n := bits.TrailingZeros64(^m&0x8080808080808080) / 8
		var x uint64
		switch n {
		case 0:
			x = uint64(ptr[0])
		case 1:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1])<<7
		case 2:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2])<<14
		case 3:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
				uint64(ptr[3])<<21
		case 4:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
				uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4])<<28
		case 5:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
				uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5])<<35
		case 6:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
				uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
				uint64(ptr[6])<<42
		case 7:
			x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
				uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
				uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7])<<49
		}
		return x, b[n+1:], nil
	}

	// All 8 bytes are continuation bytes; need to inspect ptr[8] (and maybe b[9]).
	// len(b) >= 9 is already guaranteed.
	if ptr[8] < 0x80 {
		x := uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
			uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7]&0x7f)<<49 | uint64(ptr[8])<<56
		return x, b[9:], nil
	}
	if len(b) < 10 {
		return 0, b, fmt.Errorf("unexpected EOF reading varint")
	}
	if b[9] >= 0x80 {
		return 0, b, fmt.Errorf("varint overflow")
	}
	x := uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
		uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
		uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7]&0x7f)<<49 | uint64(ptr[8]&0x7f)<<56 |
		uint64(b[9])<<63
	return x, b[10:], nil
}

// consumeVarintSlow is the fallback for len(b) < 9.
// 242.6 ns/op       626.49 MB/s  // 目前最快的版本
func consumeVarintSlow(b []byte) (uint64, []byte, error) {
	var x uint64
	var s uint
	for i, c := range b {
		if i == 10 {
			return 0, b, fmt.Errorf("varint overflow")
		}
		if c < 0x80 {
			x |= uint64(c) << s
			return x, b[i+1:], nil
		}
		x |= uint64(c&0x7f) << s
		s += 7
	}
	return 0, b, fmt.Errorf("unexpected EOF reading varint")
}

// 266.7 ns/op       570.03 MB/s  // 仍然是负优化
func ConsumeVarintNotAsm(b []byte) (uint64, []byte, error) {
	// Phase 1: scan-only loop — find the index of the first terminating byte
	// (MSB == 0).  No bit-ops here, just a single comparison per iteration.
	n := 0
	for n < len(b) && b[n] >= 0x80 {
		n++
	}
	// Check overflow before EOF so that a long all-continuation buffer
	// (e.g. 11×0x80) reports overflow rather than EOF, matching the
	// original behaviour.
	if n >= 10 {
		return 0, b, fmt.Errorf("varint overflow")
	}
	if n >= len(b) {
		return 0, b, fmt.Errorf("unexpected EOF reading varint")
	}

	// Phase 2: jump table — all bit-ops happen exactly once, no loop
	// overhead, no accumulated intermediate values.
	// Use a *[10]byte unsafe pointer to b[0] so every index access below is
	// against a compile-time-known fixed array, eliminating per-element
	// bounds checks that would otherwise be emitted for the slice.
	var x uint64
	ptr := (*[10]byte)(unsafe.Pointer(&b[0]))
	switch n {
	case 0:
		x = uint64(ptr[0])
	case 1:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1])<<7
	case 2:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2])<<14
	case 3:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3])<<21
	case 4:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4])<<28
	case 5:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5])<<35
	case 6:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
			uint64(ptr[6])<<42
	case 7:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
			uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7])<<49
	case 8:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
			uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7]&0x7f)<<49 | uint64(ptr[8])<<56
	case 9:
		x = uint64(ptr[0]&0x7f) | uint64(ptr[1]&0x7f)<<7 | uint64(ptr[2]&0x7f)<<14 |
			uint64(ptr[3]&0x7f)<<21 | uint64(ptr[4]&0x7f)<<28 | uint64(ptr[5]&0x7f)<<35 |
			uint64(ptr[6]&0x7f)<<42 | uint64(ptr[7]&0x7f)<<49 | uint64(ptr[8]&0x7f)<<56 |
			uint64(ptr[9])<<63
	}
	return x, b[n+1:], nil
}

// ConsumeTag reads a field tag (field number + wire type) from b.
func ConsumeTag(b []byte) (fieldNum int, wt WireType, rest []byte, err error) {
	var v uint64
	var code int64
	v, rest, code = ConsumeVarint(b) // rest is the named return; assigned directly
	if code != 0 {
		err = consumeVarintError(code)
		rest = b // restore original slice on error
		return
	}
	fieldNum = int(v >> 3)
	wt = WireType(v & 0x7)
	return
}

// ConsumeBytes reads a length-delimited byte slice from b.
func ConsumeBytes(b []byte) (data []byte, rest []byte, err error) {
	var l uint64
	var code int64
	l, rest, code = ConsumeVarint(b)
	if code != 0 {
		err = consumeVarintError(code)
		rest = b
		return
	}
	if uint64(len(rest)) < l {
		return nil, b, fmt.Errorf("not enough bytes: need %d have %d", l, len(rest))
	}
	return rest[:l], rest[l:], nil
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

func ReadInt32(b []byte) (int32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return int32(v), rest, nil
}

func ReadInt64(b []byte) (int64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return int64(v), rest, nil
}

func ReadUint32(b []byte) (uint32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return uint32(v), rest, nil
}

func ReadUint64(b []byte) (uint64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	return v, rest, nil
}

func ReadSint32(b []byte) (int32, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	n := int32((uint32(v) >> 1) ^ -(uint32(v) & 1))
	return n, rest, nil
}

func ReadSint64(b []byte) (int64, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return 0, b, consumeVarintError(code)
	}
	n := int64((v >> 1) ^ -(v & 1))
	return n, rest, nil
}

func ReadBool(b []byte) (bool, []byte, error) {
	v, rest, code := ConsumeVarint(b)
	if code != 0 {
		return false, b, consumeVarintError(code)
	}
	return v != 0, rest, nil
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
	if err != nil {
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

// UnsafeBytesFromString returns a []byte view of s without copying.
// The returned slice aliases s's backing storage and must be treated as read-only.
func UnsafeBytesFromString(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// EncodeJSONString appends a JSON-safe escaped version of s into dst.
// No heap allocations are performed.
func EncodeJSONString(s string, dst []byte) []byte {
	ptr := unsafe.StringData(s)
	for i := 0; i < len(s); i++ {
		c := *(*byte)(unsafe.Add(unsafe.Pointer(ptr), i))
		switch c {
		case '"':
			dst = append(dst, '\\', '"')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\t':
			dst = append(dst, '\\', 't')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\\':
			dst = append(dst, '\\', '\\')
		default:
			if c < 0x20 {
				// control characters must be escaped as \uXXXX per JSON spec
				dst = append(dst,
					'\\', 'u', '0', '0',
					"0123456789abcdef"[c>>4],
					"0123456789abcdef"[c&0xf],
				)
			} else {
				dst = append(dst, c)
			}
		}
	}
	return dst
}

// ConsumeVarint decodes a Protocol Buffers base-128 variable-length integer
// from the head of b.
//
// Return values:
//
//	value       — the decoded uint64 value (0 on error)
//	left        — remaining bytes after the consumed varint (== b on error)
//	returnValue — 0 success / 1 varint overflow / 2 unexpected EOF
func ConsumeVarint(b []byte) (value uint64, left []byte, returnValue int64) {
	// Phase 1: scan for the terminating byte (first byte with MSB == 0).
	n := 0
	for n < len(b) && b[n] >= 0x80 {
		n++
	}
	// Check overflow before EOF (matches Go protobuf error ordering).
	if n >= 10 {
		return 0, b, 1 // varint overflow
	}
	if n >= len(b) {
		return 0, b, 2 // unexpected EOF
	}

	// Phase 2: jump table — all bit-ops in one shot, no loop.
	var x uint64
	switch n {
	case 0:
		x = uint64(b[0])
	case 1:
		x = uint64(b[0]&0x7f) | uint64(b[1])<<7
	case 2:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2])<<14
	case 3:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3])<<21
	case 4:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4])<<28
	case 5:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4]&0x7f)<<28 | uint64(b[5])<<35
	case 6:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4]&0x7f)<<28 | uint64(b[5]&0x7f)<<35 |
			uint64(b[6])<<42
	case 7:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4]&0x7f)<<28 | uint64(b[5]&0x7f)<<35 |
			uint64(b[6]&0x7f)<<42 | uint64(b[7])<<49
	case 8:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4]&0x7f)<<28 | uint64(b[5]&0x7f)<<35 |
			uint64(b[6]&0x7f)<<42 | uint64(b[7]&0x7f)<<49 | uint64(b[8])<<56
	case 9:
		x = uint64(b[0]&0x7f) | uint64(b[1]&0x7f)<<7 | uint64(b[2]&0x7f)<<14 |
			uint64(b[3]&0x7f)<<21 | uint64(b[4]&0x7f)<<28 | uint64(b[5]&0x7f)<<35 |
			uint64(b[6]&0x7f)<<42 | uint64(b[7]&0x7f)<<49 | uint64(b[8]&0x7f)<<56 |
			uint64(b[9])<<63
	}
	return x, b[n+1:], 0
}
