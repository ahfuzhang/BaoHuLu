package utils

import (
	"math"
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

func TestConsumeTag_Error(t *testing.T) {
	_, _, _, err := ConsumeTag(nil)
	if err == nil {
		t.Fatal("expected error on empty input")
	}
}

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

func TestConsumeBytes_VarintError(t *testing.T) {
	_, _, err := ConsumeBytes(nil)
	if err == nil {
		t.Fatal("expected error on empty input")
	}
}

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

func TestReadString_Error(t *testing.T) {
	_, _, err := ReadString(nil)
	if err == nil {
		t.Fatal("expected error on empty input")
	}
}

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
