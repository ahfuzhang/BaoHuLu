package utils

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"
)

// ─── AppendVarint / VarintSize ────────────────────────────────────────────────

func TestAppendVarint(t *testing.T) {
	cases := []struct {
		name  string
		v     uint64
		want  []byte
		bytes int // expected encoded length
	}{
		// 1 byte: v < 2^7
		{"zero", 0, []byte{0x00}, 1},
		{"max1", (1 << 7) - 1, []byte{0x7f}, 1},
		// 2 bytes: 2^7 <= v < 2^14
		{"min2", 1 << 7, []byte{0x80, 0x01}, 2},
		{"max2", (1 << 14) - 1, []byte{0xff, 0x7f}, 2},
		// 3 bytes: 2^14 <= v < 2^21
		{"min3", 1 << 14, []byte{0x80, 0x80, 0x01}, 3},
		{"max3", (1 << 21) - 1, []byte{0xff, 0xff, 0x7f}, 3},
		// 4 bytes: 2^21 <= v < 2^28
		{"min4", 1 << 21, []byte{0x80, 0x80, 0x80, 0x01}, 4},
		{"max4", (1 << 28) - 1, []byte{0xff, 0xff, 0xff, 0x7f}, 4},
		// 5 bytes: 2^28 <= v < 2^35
		{"min5", 1 << 28, []byte{0x80, 0x80, 0x80, 0x80, 0x01}, 5},
		{"max5", (1 << 35) - 1, []byte{0xff, 0xff, 0xff, 0xff, 0x7f}, 5},
		// 6 bytes: 2^35 <= v < 2^42
		{"min6", 1 << 35, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 6},
		{"max6", (1 << 42) - 1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 6},
		// 7 bytes: 2^42 <= v < 2^49
		{"min7", 1 << 42, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 7},
		{"max7", (1 << 49) - 1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 7},
		// 8 bytes: 2^49 <= v < 2^56
		{"min8", 1 << 49, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 8},
		{"max8", (1 << 56) - 1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 8},
		// 9 bytes: 2^56 <= v < 2^63
		{"min9", 1 << 56, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 9},
		{"max9", (1 << 63) - 1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 9},
		// 10 bytes: v >= 2^63
		{"min10", 1 << 63, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 10},
		{"maxUint64", ^uint64(0), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AppendVarint(nil, tc.v)
			if len(got) != len(tc.want) {
				t.Fatalf("AppendVarint(%d): len=%d want %d", tc.v, len(got), len(tc.want))
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("AppendVarint(%d): byte[%d]=%#x want %#x", tc.v, i, got[i], tc.want[i])
				}
			}

			// Also verify VarintSize
			sz := VarintSize(tc.v)
			if sz != tc.bytes {
				t.Fatalf("VarintSize(%d)=%d want %d", tc.v, sz, tc.bytes)
			}

			// Round-trip via ConsumeVarint
			decoded, rest, code := ConsumeVarint(got)
			if code != 0 {
				t.Fatalf("ConsumeVarint round-trip: code=%d", code)
			}
			if decoded != tc.v {
				t.Fatalf("round-trip: got %d want %d", decoded, tc.v)
			}
			if len(rest) != 0 {
				t.Fatalf("round-trip: unexpected remaining bytes: %v", rest)
			}
		})
	}
}

// ─── ConsumeVarint error paths ────────────────────────────────────────────────

func TestConsumeVarint_Overflow(t *testing.T) {
	// 11 continuation bytes → overflow (code 1)
	b := make([]byte, 11)
	for i := range b {
		b[i] = 0x80
	}
	_, _, code := ConsumeVarint(b)
	if code != 1 {
		t.Fatalf("expected overflow (code=1), got code=%d", code)
	}
}

func TestConsumeVarint_EOF(t *testing.T) {
	// All bytes have MSB set → EOF (code 2)
	b := []byte{0x80, 0x80}
	_, _, code := ConsumeVarint(b)
	if code != 2 {
		t.Fatalf("expected EOF (code=2), got code=%d", code)
	}
}

func TestConsumeVarint_EmptySlice(t *testing.T) {
	_, _, code := ConsumeVarint(nil)
	if code != 2 {
		t.Fatalf("expected EOF (code=2) on empty input, got code=%d", code)
	}
}

// ─── AppendTag / TagSize ──────────────────────────────────────────────────────

func TestAppendTagAndTagSize(t *testing.T) {
	cases := []struct {
		fieldNum int
		wt       WireType
	}{
		{1, WireTypeVarint},
		{1, WireType64bit},
		{1, WireTypeLenDelim},
		{1, WireType32bit},
		{16, WireTypeVarint},   // field 16 pushes tag into 2-byte range
		{2048, WireTypeVarint}, // larger field number
	}
	for _, tc := range cases {
		b := AppendTag(nil, tc.fieldNum, tc.wt)
		sz := TagSize(tc.fieldNum, tc.wt)
		if len(b) != sz {
			t.Errorf("AppendTag/TagSize mismatch for field=%d wt=%d: len=%d sz=%d",
				tc.fieldNum, tc.wt, len(b), sz)
		}

		// Round-trip
		fn, wt, rest, err := ConsumeTag(b)
		if err != nil {
			t.Errorf("ConsumeTag: %v", err)
			continue
		}
		if fn != tc.fieldNum || wt != tc.wt {
			t.Errorf("ConsumeTag round-trip: got field=%d wt=%d, want field=%d wt=%d",
				fn, wt, tc.fieldNum, tc.wt)
		}
		if len(rest) != 0 {
			t.Errorf("ConsumeTag: unexpected rest bytes")
		}
	}
}

// func TestConsumeTag_Error(t *testing.T) {
// 	_, _, _, err := ConsumeTag(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// ─── Sint32 / Sint64 round-trips ──────────────────────────────────────────────

func TestSint32RoundTrip(t *testing.T) {
	cases := []int32{0, 1, -1, 127, -128, math.MaxInt32, math.MinInt32}
	for _, v := range cases {
		b := AppendSint32(nil, v)
		got, rest, err := ReadSint32(b)
		if err != nil {
			t.Errorf("ReadSint32(%d): %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("Sint32 round-trip: got %d want %d", got, v)
		}
		if len(rest) != 0 {
			t.Errorf("Sint32 round-trip: unexpected rest")
		}
	}
}

func TestSint64RoundTrip(t *testing.T) {
	cases := []int64{0, 1, -1, 127, -128, math.MaxInt64, math.MinInt64}
	for _, v := range cases {
		b := AppendSint64(nil, v)
		got, rest, err := ReadSint64(b)
		if err != nil {
			t.Errorf("ReadSint64(%d): %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("Sint64 round-trip: got %d want %d", got, v)
		}
		if len(rest) != 0 {
			t.Errorf("Sint64 round-trip: unexpected rest")
		}
	}
}

// ─── Fixed32 / Fixed64 ────────────────────────────────────────────────────────

func TestFixed32RoundTrip(t *testing.T) {
	cases := []uint32{0, 1, 0xdeadbeef, math.MaxUint32}
	for _, v := range cases {
		b := AppendFixed32(nil, v)
		if len(b) != 4 {
			t.Errorf("AppendFixed32: len=%d want 4", len(b))
		}
		got, rest, err := ReadFixed32(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Fixed32 round-trip failed for %d: got=%d err=%v", v, got, err)
		}
	}
}

func TestFixed32_EOF(t *testing.T) {
	_, _, err := ReadFixed32([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestFixed64RoundTrip(t *testing.T) {
	cases := []uint64{0, 1, 0xdeadbeefcafebabe, math.MaxUint64}
	for _, v := range cases {
		b := AppendFixed64(nil, v)
		if len(b) != 8 {
			t.Errorf("AppendFixed64: len=%d want 8", len(b))
		}
		got, rest, err := ReadFixed64(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Fixed64 round-trip failed for %d: got=%d err=%v", v, got, err)
		}
	}
}

func TestFixed64_EOF(t *testing.T) {
	_, _, err := ReadFixed64([]byte{0x01, 0x02, 0x03})
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

// ─── Sfixed32 / Sfixed64 ──────────────────────────────────────────────────────

func TestSfixed32RoundTrip(t *testing.T) {
	cases := []int32{0, 1, -1, math.MaxInt32, math.MinInt32}
	for _, v := range cases {
		b := AppendFixed32(nil, uint32(v))
		got, rest, err := ReadSfixed32(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Sfixed32 round-trip failed for %d: got=%d err=%v", v, got, err)
		}
	}
}

func TestSfixed64RoundTrip(t *testing.T) {
	cases := []int64{0, 1, -1, math.MaxInt64, math.MinInt64}
	for _, v := range cases {
		b := AppendFixed64(nil, uint64(v))
		got, rest, err := ReadSfixed64(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Sfixed64 round-trip failed for %d: got=%d err=%v", v, got, err)
		}
	}
}

// ─── Float / Double ───────────────────────────────────────────────────────────

func TestFloatRoundTrip(t *testing.T) {
	cases := []float32{0, 1.5, -1.5, float32(math.Pi), float32(math.MaxFloat32)}
	for _, v := range cases {
		b := AppendFixed32(nil, math.Float32bits(v))
		got, rest, err := ReadFloat(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Float round-trip failed for %v: got=%v err=%v", v, got, err)
		}
	}
}

func TestDoubleRoundTrip(t *testing.T) {
	cases := []float64{0, 1.5, -1.5, math.Pi, math.MaxFloat64}
	for _, v := range cases {
		b := AppendFixed64(nil, math.Float64bits(v))
		got, rest, err := ReadDouble(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("Double round-trip failed for %v: got=%v err=%v", v, got, err)
		}
	}
}

// ─── Int32 / Int64 / Uint32 / Uint64 / Bool ───────────────────────────────────

func TestReadInt32(t *testing.T) {
	cases := []int32{0, 1, -1, math.MaxInt32}
	for _, v := range cases {
		b := AppendVarint(nil, uint64(v))
		got, rest, err := ReadInt32(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("ReadInt32(%d): got=%d err=%v", v, got, err)
		}
	}
}

func TestReadInt64(t *testing.T) {
	cases := []int64{0, 1, -1, math.MaxInt64}
	for _, v := range cases {
		b := AppendVarint(nil, uint64(v))
		got, rest, err := ReadInt64(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("ReadInt64(%d): got=%d err=%v", v, got, err)
		}
	}
}

func TestReadUint32(t *testing.T) {
	cases := []uint32{0, 1, math.MaxUint32}
	for _, v := range cases {
		b := AppendVarint(nil, uint64(v))
		got, rest, err := ReadUint32(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("ReadUint32(%d): got=%d err=%v", v, got, err)
		}
	}
}

func TestReadUint64(t *testing.T) {
	cases := []uint64{0, 1, math.MaxUint64}
	for _, v := range cases {
		b := AppendVarint(nil, v)
		got, rest, err := ReadUint64(b)
		if err != nil || got != v || len(rest) != 0 {
			t.Errorf("ReadUint64(%d): got=%d err=%v", v, got, err)
		}
	}
}

func TestReadBool(t *testing.T) {
	bTrue := AppendVarint(nil, 1)
	bFalse := AppendVarint(nil, 0)

	v, _, err := ReadBool(bTrue)
	if err != nil || !v {
		t.Errorf("ReadBool(true): got=%v err=%v", v, err)
	}
	v, _, err = ReadBool(bFalse)
	if err != nil || v {
		t.Errorf("ReadBool(false): got=%v err=%v", v, err)
	}
}

// ─── LenDelim / String / Bytes ────────────────────────────────────────────────

func TestAppendLenDelim(t *testing.T) {
	data := UnsafeBytesFromString("hello")
	b := AppendLenDelim(nil, data)
	// First byte should be the length (5)
	l, rest, code := ConsumeVarint(b)
	if code != 0 || l != 5 {
		t.Fatalf("LenDelim length: got=%d code=%d", l, code)
	}
	if string(rest) != "hello" {
		t.Fatalf("LenDelim data: got=%q", rest)
	}
}

func TestConsumeBytes(t *testing.T) {
	data := UnsafeBytesFromString("world")
	b := AppendLenDelim(nil, data)
	got, rest, err := ConsumeBytes(b)
	if err != nil || string(got) != "world" || len(rest) != 0 {
		t.Errorf("ConsumeBytes: got=%q rest=%v err=%v", got, rest, err)
	}
}

func TestConsumeBytes_NotEnough(t *testing.T) {
	// Encode length=10 but only provide 3 bytes of data
	b := AppendVarint(nil, 10)
	b = append(b, []byte{1, 2, 3}...)
	_, _, err := ConsumeBytes(b)
	if err == nil {
		t.Fatal("expected not-enough-bytes error")
	}
}

// func TestConsumeBytes_VarintError(t *testing.T) {
// 	_, _, err := ConsumeBytes(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

func TestReadString(t *testing.T) {
	s := "hello, protobuf"
	b := AppendLenDelim(nil, UnsafeBytesFromString(s))
	got, rest, err := ReadString(b)
	if err != nil || got != s || len(rest) != 0 {
		t.Errorf("ReadString: got=%q rest=%v err=%v", got, rest, err)
	}
}

func TestUnsafeBytesFromString(t *testing.T) {
	s := "hello"
	b := UnsafeBytesFromString(s)
	if string(b) != s {
		t.Fatalf("UnsafeBytesFromString: got=%q want %q", b, s)
	}
	if unsafe.SliceData(b) != unsafe.StringData(s) {
		t.Fatal("UnsafeBytesFromString copied the input string")
	}
}

func TestReadString_Empty(t *testing.T) {
	b := AppendLenDelim(nil, []byte{})
	got, rest, err := ReadString(b)
	if err != nil || got != "" || len(rest) != 0 {
		t.Errorf("ReadString(empty): got=%q rest=%v err=%v", got, rest, err)
	}
}

// func TestReadString_Error(t *testing.T) {
// 	_, _, err := ReadString(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

func TestReadBytes(t *testing.T) {
	data := []byte{0xde, 0xad, 0xbe, 0xef}
	b := AppendLenDelim(nil, data)
	got, rest, err := ReadBytes(b)
	if err != nil || len(rest) != 0 {
		t.Fatalf("ReadBytes: err=%v", err)
	}
	for i, v := range data {
		if got[i] != v {
			t.Errorf("ReadBytes[%d]: got=%#x want=%#x", i, got[i], v)
		}
	}
}

// ─── SkipField ────────────────────────────────────────────────────────────────

func TestSkipField_Varint(t *testing.T) {
	b := AppendVarint(nil, 12345)
	b = append(b, 0xff) // sentinel
	rest, err := SkipField(WireTypeVarint, b)
	if err != nil {
		t.Fatalf("SkipField varint: %v", err)
	}
	if len(rest) != 1 || rest[0] != 0xff {
		t.Errorf("SkipField varint: unexpected rest %v", rest)
	}
}

func TestSkipField_Varint_EOF(t *testing.T) {
	// All bytes are continuation bytes → EOF
	_, err := SkipField(WireTypeVarint, []byte{0x80, 0x80})
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestSkipField_64bit(t *testing.T) {
	b := AppendFixed64(nil, 0x0102030405060708)
	b = append(b, 0xaa)
	rest, err := SkipField(WireType64bit, b)
	if err != nil || len(rest) != 1 || rest[0] != 0xaa {
		t.Errorf("SkipField 64bit: rest=%v err=%v", rest, err)
	}
}

func TestSkipField_64bit_EOF(t *testing.T) {
	_, err := SkipField(WireType64bit, []byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestSkipField_LenDelim(t *testing.T) {
	b := AppendLenDelim(nil, UnsafeBytesFromString("skip me"))
	b = append(b, 0xbb)
	rest, err := SkipField(WireTypeLenDelim, b)
	if err != nil || len(rest) != 1 || rest[0] != 0xbb {
		t.Errorf("SkipField LenDelim: rest=%v err=%v", rest, err)
	}
}

func TestSkipField_32bit(t *testing.T) {
	b := AppendFixed32(nil, 0xdeadbeef)
	b = append(b, 0xcc)
	rest, err := SkipField(WireType32bit, b)
	if err != nil || len(rest) != 1 || rest[0] != 0xcc {
		t.Errorf("SkipField 32bit: rest=%v err=%v", rest, err)
	}
}

func TestSkipField_32bit_EOF(t *testing.T) {
	_, err := SkipField(WireType32bit, []byte{1, 2})
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestSkipField_UnknownWireType(t *testing.T) {
	_, err := SkipField(WireType(99), []byte{0x01})
	if err == nil {
		t.Fatal("expected unknown wire type error")
	}
}

// ─── Multi-byte trailing data (ensure rest slicing is correct) ────────────────

func TestAppendVarint_ExistingBuffer(t *testing.T) {
	prefix := []byte{0xAA, 0xBB}
	b := AppendVarint(prefix, 300)
	if b[0] != 0xAA || b[1] != 0xBB {
		t.Errorf("AppendVarint should preserve existing prefix")
	}
	if len(b) != 4 { // 2 prefix + 2 varint bytes for 300
		t.Errorf("unexpected length %d", len(b))
	}
}

// ─── EncodeVarint ─────────────────────────────────────────────────────────────

func TestEncodeVarint(t *testing.T) {
	cases := []struct {
		name string
		v    uint64
		size int
	}{
		// 1 byte: fast path (v < 0x80)
		{"zero", 0, 1},
		{"max1", (1 << 7) - 1, 1},
		// 2 bytes: fast path (0x80 <= v < 0x4000)
		{"min2", 1 << 7, 2},
		{"max2", (1 << 14) - 1, 2},
		// 3-10 bytes: switch path
		{"min3", 1 << 14, 3},
		{"max3", (1 << 21) - 1, 3},
		{"min4", 1 << 21, 4},
		{"max4", (1 << 28) - 1, 4},
		{"min5", 1 << 28, 5},
		{"max5", (1 << 35) - 1, 5},
		{"min6", 1 << 35, 6},
		{"max6", (1 << 42) - 1, 6},
		{"min7", 1 << 42, 7},
		{"max7", (1 << 49) - 1, 7},
		{"min8", 1 << 49, 8},
		{"max8", (1 << 56) - 1, 8},
		{"min9", 1 << 56, 9},
		{"max9", (1 << 63) - 1, 9},
		{"min10", uint64(1) << 63, 10},
		{"maxUint64", ^uint64(0), 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]byte, tc.size)
			gotOffset := EncodeVarint(buf, tc.size, tc.v)
			if gotOffset != 0 {
				t.Fatalf("EncodeVarint(%d): expected offset 0, got %d", tc.v, gotOffset)
			}
			want := AppendVarint(nil, tc.v)
			if len(buf) != len(want) {
				t.Fatalf("EncodeVarint(%d): len=%d want %d", tc.v, len(buf), len(want))
			}
			for i := range want {
				if buf[i] != want[i] {
					t.Fatalf("EncodeVarint(%d): byte[%d]=%#x want %#x", tc.v, i, buf[i], want[i])
				}
			}
			// Round-trip via ConsumeVarint
			decoded, rest, code := ConsumeVarint(buf)
			if code != 0 {
				t.Fatalf("EncodeVarint(%d): ConsumeVarint code=%d", tc.v, code)
			}
			if decoded != tc.v {
				t.Fatalf("EncodeVarint(%d): round-trip got %d", tc.v, decoded)
			}
			if len(rest) != 0 {
				t.Fatalf("EncodeVarint(%d): unexpected rest bytes", tc.v)
			}
		})
	}
}

// ─── consumeVarintError overflow path ─────────────────────────────────────────

func TestConsumeTag_Overflow(t *testing.T) {
	b := make([]byte, 11)
	for i := range b {
		b[i] = 0x80
	}
	_, _, _, err := ConsumeTag(b)
	if err == nil {
		t.Fatal("expected overflow error from ConsumeTag")
	}
}

func TestConsumeBytes_Overflow(t *testing.T) {
	b := make([]byte, 11)
	for i := range b {
		b[i] = 0x80
	}
	_, _, err := ConsumeBytes(b)
	if err == nil {
		t.Fatal("expected overflow error from ConsumeBytes")
	}
}

// ─── Read* error paths ────────────────────────────────────────────────────────

// func TestReadInt32_Error(t *testing.T) {
// 	_, _, err := ReadInt32(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadInt64_Error(t *testing.T) {
// 	_, _, err := ReadInt64(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadUint32_Error(t *testing.T) {
// 	_, _, err := ReadUint32(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadUint64_Error(t *testing.T) {
// 	_, _, err := ReadUint64(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadBool_Error(t *testing.T) {
// 	_, _, err := ReadBool(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadSint32_Error(t *testing.T) {
// 	_, _, err := ReadSint32(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// func TestReadSint64_Error(t *testing.T) {
// 	_, _, err := ReadSint64(nil)
// 	if err == nil {
// 		t.Fatal("expected error on empty input")
// 	}
// }

// ─── EncodeJSONString ─────────────────────────────────────────────────────────

func TestEncodeJSONString(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain", "hello world", "hello world"},
		{"double_quote", `say "hi"`, `say \"hi\"`},
		{"newline", "line\nnewline", `line\nnewline`},
		{"tab", "tab\there", `tab\there`},
		{"carriage_return", "cr\rreturn", `cr\rreturn`},
		{"backslash", `back\slash`, `back\\slash`},
		{"form_feed", "\fpage", `\fpage`},
		{"backspace", "\bhello", `\bhello`},
		{"ctrl_soh", "\x01", "\\u0001"},
		{"ctrl_us", "\x1f", "\\u001f"},
		{"ctrl_null", "\x00", "\\u0000"},
		{"ctrl_vtab", "\x0b", "\\u000b"},
		{"mixed", "a\"b\nc\\d", `a\"b\nc\\d`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(EncodeJSONString(tc.input, nil))
			if got != tc.want {
				t.Fatalf("EncodeJSONString(%q): got %q, want %q", tc.input, got, tc.want)
			}
		})
	}

	// Test with pre-allocated dst (skips the capacity-expansion branch).
	t.Run("preallocated_dst", func(t *testing.T) {
		dst := make([]byte, 0, 512)
		got := string(EncodeJSONString("hello", dst))
		if got != "hello" {
			t.Fatalf("EncodeJSONString preallocated: got %q, want %q", got, "hello")
		}
	})

	// Test that existing content in dst is preserved.
	t.Run("appends_to_existing", func(t *testing.T) {
		dst := make([]byte, 0, 512)
		dst = append(dst, []byte("prefix:")...)
		got := string(EncodeJSONString("hi", dst))
		if got != "prefix:hi" {
			t.Fatalf("EncodeJSONString append: got %q, want %q", got, "prefix:hi")
		}
	})
}

// ─── ReadVarint ───────────────────────────────────────────────────────────────

func TestReadVarint(t *testing.T) {
	cases := []struct {
		name string
		v    uint64
	}{
		{"1byte_min", 0},
		{"1byte_max", (1 << 7) - 1},
		{"2byte_min", 1 << 7},
		{"2byte_max", (1 << 14) - 1},
		{"3byte_min", 1 << 14},
		{"3byte_max", (1 << 21) - 1},
		{"4byte_min", 1 << 21},
		{"4byte_max", (1 << 28) - 1},
		{"5byte_min", 1 << 28},
		{"5byte_max", (1 << 35) - 1},
		{"6byte_min", 1 << 35},
		{"6byte_max", (1 << 42) - 1},
		{"7byte_min", 1 << 42},
		{"7byte_max", (1 << 49) - 1},
		{"8byte_min", 1 << 49},
		{"8byte_max", (1 << 56) - 1},
		{"9byte_min", 1 << 56},
		{"9byte_max", (1 << 63) - 1},
		{"10byte_min", uint64(1) << 63},
		{"10byte_max", ^uint64(0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := AppendVarint(nil, tc.v)
			b = append(b, 0xAA) // sentinel to verify rest slicing
			got, rest, err := ReadVarint(b)
			if err != nil {
				t.Fatalf("ReadVarint(%d): unexpected error: %v", tc.v, err)
			}
			if got != tc.v {
				t.Fatalf("ReadVarint(%d): got %d want %d", tc.v, got, tc.v)
			}
			if len(rest) != 1 || rest[0] != 0xAA {
				t.Fatalf("ReadVarint(%d): unexpected rest %v", tc.v, rest)
			}
		})
	}
}

func TestReadVarint_EOF(t *testing.T) {
	// 10 continuation bytes: all 10 conditions false → errVarintEOF
	b := make([]byte, 10)
	for i := range b {
		b[i] = 0x80
	}
	_, _, err := ReadVarint(b)
	if err == nil {
		t.Fatal("expected EOF error from ReadVarint")
	}
}

var benchmarkInput = "benchmark payload with escape chars:\n newline \t tab \" double-quote \\ backslash; " +
	"padding to ensure length exceeds one hundred bytes: 0123456789abcdef0123456789abcdef"

// 449.89 MB/s
// 优化后： 964.47 MB/s
// v5 版本优化后: 1285.23 MB/s
func BenchmarkEncodeJSONString(b *testing.B) {
	dst := make([]byte, 0, 256)
	b.SetBytes(int64(len(benchmarkInput)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = EncodeJSONString(benchmarkInput, dst[:0])
	}
	_ = dst
}

// BenchmarkDecodeVarint 在同一 Benchmark 内对比 ReadVarint 与 ConsumeVarintV1 的性能。
//
// 运行方式：
//
//	go test -bench=BenchmarkDecodeVarint -v -benchmem ./...
//
// go test -test.fullpath=true -benchmem -run=^$ -bench ^BenchmarkDecodeVarint$ github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils
// 每个 case 覆盖一种编码长度（1~10 字节），两个实现紧邻，便于直接比较 ns/op。
func BenchmarkDecodeVarint(b *testing.B) {
	// 构造各种编码长度的测试向量：先编码，再在末尾补哨兵字节 0xAA
	cases := []struct {
		name string
		v    uint64
	}{
		{"1byte", 0},
		{"1byte_max", (1 << 7) - 1},
		{"2byte", 1 << 7},
		{"2byte_max", (1 << 14) - 1},
		{"3byte", 1 << 14},
		{"4byte", 1 << 21},
		{"5byte", 1 << 28},
		{"6byte", 1 << 35},
		{"7byte", 1 << 42},
		{"8byte", 1 << 49},
		{"9byte", 1 << 56},
		{"10byte", ^uint64(0)},
	}

	type decodeFn func([]byte) (uint64, []byte, error)
	impls := []struct {
		name string
		fn   decodeFn
	}{
		{"ReadVarint", ReadVarint},
		{"ConsumeVarintV1", ConsumeVarintV1},
	}

	for _, tc := range cases {
		encoded := AppendVarint(nil, tc.v)
		encoded = append(encoded, 0xAA) // 哨兵：确保 rest 切片正确

		for _, impl := range impls {
			b.Run(tc.name+"/"+impl.name, func(b *testing.B) {
				b.SetBytes(int64(len(encoded)))
				b.ResetTimer()
				var (
					sink  uint64
					dummy []byte
				)
				for i := 0; i < b.N; i++ {
					sink, dummy, _ = impl.fn(encoded)
				}
				_ = sink
				_ = dummy
			})
		}
	}
}

// TestCompareReadVarintVsConsumeVarintV1 使用 testing.Benchmark 对比两种 varint 解码实现的性能，
// 覆盖 1~10 字节编码长度，输出 ns/op、MB/s 以及 ConsumeVarintV1 相对于 ReadVarint 的差值百分比。
// 正值（+xx%）表示 ConsumeVarintV1 更快，负值（-xx%）表示更慢。
//
// 运行方式：
//
//	go test -v -run=TestCompareReadVarintVsConsumeVarintV1 ./...
//
// go test -test.fullpath=true -run ^TestCompareReadVarintVsConsumeVarintV1$ github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils -v
func TestCompareReadVarintVsConsumeVarintV1(t *testing.T) {
	cases := []struct {
		name string
		v    uint64
	}{
		{"1byte", 0},
		{"2byte", 1 << 7},
		{"3byte", 1 << 14},
		{"4byte", 1 << 21},
		{"5byte", 1 << 28},
		{"6byte", 1 << 35},
		{"7byte", 1 << 42},
		{"8byte", 1 << 49},
		{"9byte", 1 << 56},
		{"10byte", ^uint64(0)},
	}

	type result struct {
		nsPerOp  float64
		mbPerSec float64
	}

	bench := func(fn func([]byte) (uint64, []byte, error), encoded []byte) result {
		r := testing.Benchmark(func(b *testing.B) {
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()
			var (
				sink  uint64
				dummy []byte
				err   error
			)
			for i := 0; i < b.N; i++ {
				sink, dummy, err = fn(encoded)
				if err != nil {
					panic(err)
				}
			}
			_ = sink
			_ = dummy
		})
		var mbPerSec float64
		if r.Bytes > 0 && r.T > 0 && r.N > 0 {
			mbPerSec = float64(r.Bytes) * float64(r.N) / 1e6 / r.T.Seconds()
		}
		return result{
			nsPerOp:  float64(r.NsPerOp()),
			mbPerSec: mbPerSec,
		}
	}

	const hdr = "%-8s  %12s  %10s  %12s  %10s  %10s\n"
	const sep = "%-8s  %12s  %10s  %12s  %10s  %10s\n"
	const row = "%-8s  %12.2f  %10.1f  %12.2f  %10.1f  %+9.1f%%\n"

	fmt.Printf(hdr, "Size", "RV ns/op", "RV MB/s", "V1 ns/op", "V1 MB/s", "Delta")
	fmt.Printf(sep,
		strings.Repeat("-", 8),
		strings.Repeat("-", 12),
		strings.Repeat("-", 10),
		strings.Repeat("-", 12),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
	)

	for _, tc := range cases {
		encoded := AppendVarint(nil, tc.v)
		encoded = append(encoded, 0xAA) // 哨兵：确保 rest 切片正确
		encoded = append(encoded, "0123456789"...)

		rv := bench(ReadVarint, encoded)
		cv := bench(ConsumeVarintV1, encoded)

		// 正值 = ConsumeVarintV1 更快（ns 更低），负值 = 更慢
		delta := (rv.mbPerSec - cv.mbPerSec) / rv.mbPerSec * 100

		fmt.Printf(row,
			tc.name,
			rv.nsPerOp, rv.mbPerSec,
			cv.nsPerOp, cv.mbPerSec,
			delta,
		)
	}
}
