// Package utils provides low-level protobuf binary encoding/decoding primitives
package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
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

// EncodeVarintBackward encodes v as a varint ending at dAtA[offset] (exclusive)
// and returns the starting offset. Equivalent to protohelpers.EncodeVarint:
// the varint bytes are placed at dAtA[base:offset] where base is returned.
func EncodeVarintBackward(dAtA []byte, offset int, v uint64) int {
	offset -= VarintSize(v)
	base := offset
	for v >= 0x80 {
		dAtA[offset] = byte(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = byte(v)
	return base
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

// AppendJSONKey appends `"key":` to dst using a single capacity check.
// It pre-grows dst by len(key)+3 bytes, then writes the bytes from tail to head
// to reduce bounds checks.
func AppendJSONKey(dst []byte, key string) []byte {
	nameLen := len(key)
	n := nameLen + 3
	base := len(dst)
	if cap(dst)-base < n {
		tmp := make([]byte, base, base+n+256)
		copy(tmp, dst)
		dst = tmp
	}
	dst = dst[:base+n]
	pos := base + n
	dst[pos-1] = ':'
	pos--
	dst[pos-1] = '"'
	pos -= nameLen
	copy(dst[pos-1:], key)
	dst[pos-2] = '"'
	return dst
}

// EncodeJSONString appends a JSON-safe escaped version of s into dst.
// No heap allocations are performed.
func EncodeJSONString(s string, dst []byte) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	for i := 0; i < len(b); i++ {
		c := b[i]
		if (escapeTable[c/8] & (1 << (c % 8))) == 0 {
			// fast path
			dst = append(dst, c)
			continue
		}
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
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\b':
			dst = append(dst, '\\', 'b')
		default:
			// control characters must be escaped as \uXXXX per JSON spec
			dst = append(dst,
				'\\', 'u', '0', '0',
				hex[c>>4],
				hex[c&0xf],
			)
		}
	}
	return dst
}

func EncodeJSONStringV0(s string, dst []byte) []byte {
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

const hex = "0123456789abcdef"

func EncodeJSONStringV2(s string, dst []byte) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	for i := 0; i < len(b); i++ {
		switch b[i] {
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
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\b':
			dst = append(dst, '\\', 'b')
		default:
			if b[i] >= 0x20 {
				dst = append(dst, b[i])
				continue
			}
			// control characters must be escaped as \uXXXX per JSON spec
			dst = append(dst,
				'\\', 'u', '0', '0',
				hex[b[i]>>4],
				hex[b[i]&0xf],
			)
		}
	}
	return dst
}

var escapeTable [256 / 8]byte = func() [256 / 8]byte {
	var table [256 / 8]byte
	for i := 0; i < 256; i++ {
		if i < 0x20 || i == '"' || i == '\\' {
			table[i/8] |= 1 << (i % 8)
		}
	}
	return table
}()

func EncodeJSONStringV3(s string, dst []byte) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	for i := 0; i < len(b); i++ {
		c := b[i]
		if (escapeTable[c/8] & (1 << (c % 8))) == 0 {
			// fast path
			dst = append(dst, c)
			continue
		}
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
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\b':
			dst = append(dst, '\\', 'b')
		default:
			// control characters must be escaped as \uXXXX per JSON spec
			dst = append(dst,
				'\\', 'u', '0', '0',
				hex[c>>4],
				hex[c&0xf],
			)
		}
	}
	return dst
}

var escapeTable2 [256]bool = func() [256]bool {
	var table [256]bool
	for i := 0; i < 256; i++ {
		if i < 0x20 || i == '"' || i == '\\' {
			table[i] = true
		}
	}
	return table
}()

func EncodeJSONStringV4(s string, dst []byte) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	for i := 0; i < len(b); i++ {
		c := b[i]
		if !escapeTable2[c] {
			// fast path
			dst = append(dst, c)
			continue
		}
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
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\b':
			dst = append(dst, '\\', 'b')
		default:
			// control characters must be escaped as \uXXXX per JSON spec
			dst = append(dst,
				'\\', 'u', '0', '0',
				hex[c>>4],
				hex[c&0xf],
			)
		}
	}
	return dst
}

// copy from fastjson
func EncodeJSONStringSlow(s string, dst []byte) []byte {
	if !hasSpecialChars(s) {
		// Fast path - nothing to escape.
		dst = append(dst, '"')
		dst = append(dst, s...)
		dst = append(dst, '"')
		return dst
	}

	// Slow path.
	return strconv.AppendQuote(dst, s)
}

// todo: 值得使用 avx2 加速
func hasSpecialChars(s string) bool {
	if strings.IndexByte(s, '"') >= 0 || strings.IndexByte(s, '\\') >= 0 {
		return true
	}
	for i := range len(s) {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
}

func ConsumeVarint(b []byte) (uint64, []byte, int64) {
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
