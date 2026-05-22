package utils

import (
	"fmt"
	"testing"
)

// Representative values that encode to exactly N bytes each.
var varintByteValues = [10]uint64{
	0,       // 1 byte:  v < 2^7
	1 << 7,  // 2 bytes: 2^7  <= v < 2^14
	1 << 14, // 3 bytes: 2^14 <= v < 2^21
	1 << 21, // 4 bytes: 2^21 <= v < 2^28
	1 << 28, // 5 bytes: 2^28 <= v < 2^35
	1 << 35, // 6 bytes: 2^35 <= v < 2^42
	1 << 42, // 7 bytes: 2^42 <= v < 2^49
	1 << 49, // 8 bytes: 2^49 <= v < 2^56
	1 << 56, // 9 bytes: 2^56 <= v < 2^63
	1 << 63, // 10 bytes: v >= 2^63
}

// ─── AppendVarint per-size benchmarks ────────────────────────────────────────

func BenchmarkAppendVarint1(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[0])
	}
	_ = buf
}

func BenchmarkAppendVarint2(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[1])
	}
	_ = buf
}

func BenchmarkAppendVarint3(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[2])
	}
	_ = buf
}

func BenchmarkAppendVarint4(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[3])
	}
	_ = buf
}

func BenchmarkAppendVarint5(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[4])
	}
	_ = buf
}

func BenchmarkAppendVarint6(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[5])
	}
	_ = buf
}

func BenchmarkAppendVarint7(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[6])
	}
	_ = buf
}

func BenchmarkAppendVarint8(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[7])
	}
	_ = buf
}

func BenchmarkAppendVarint9(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[8])
	}
	_ = buf
}

func BenchmarkAppendVarint10(b *testing.B) {
	buf := make([]byte, 0, 16)
	b.SetBytes(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = AppendVarint(buf[:0], varintByteValues[9])
	}
	_ = buf
}

// ─── EncodeVarintV1 per-size benchmarks ──────────────────────────────────────

func BenchmarkEncodeVarintV11(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 1, varintByteValues[0])
	}
	_ = buf
}

func BenchmarkEncodeVarintV12(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 2, varintByteValues[1])
	}
	_ = buf
}

func BenchmarkEncodeVarintV13(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 3, varintByteValues[2])
	}
	_ = buf
}

func BenchmarkEncodeVarintV14(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 4, varintByteValues[3])
	}
	_ = buf
}

func BenchmarkEncodeVarintV15(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 5, varintByteValues[4])
	}
	_ = buf
}

func BenchmarkEncodeVarintV16(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 6, varintByteValues[5])
	}
	_ = buf
}

func BenchmarkEncodeVarintV17(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 7, varintByteValues[6])
	}
	_ = buf
}

func BenchmarkEncodeVarintV18(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 8, varintByteValues[7])
	}
	_ = buf
}

func BenchmarkEncodeVarintV19(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 9, varintByteValues[8])
	}
	_ = buf
}

func BenchmarkEncodeVarintV110(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV1(buf, 10, varintByteValues[9])
	}
	_ = buf
}

// ─── EncodeVarint per-size benchmarks ────────────────────────────────────────

func BenchmarkEncodeVarint1(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 1, varintByteValues[0])
	}
	_ = buf
}

func BenchmarkEncodeVarint2(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 2, varintByteValues[1])
	}
	_ = buf
}

func BenchmarkEncodeVarint3(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 3, varintByteValues[2])
	}
	_ = buf
}

func BenchmarkEncodeVarint4(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 4, varintByteValues[3])
	}
	_ = buf
}

func BenchmarkEncodeVarint5(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 5, varintByteValues[4])
	}
	_ = buf
}

func BenchmarkEncodeVarint6(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 6, varintByteValues[5])
	}
	_ = buf
}

func BenchmarkEncodeVarint7(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 7, varintByteValues[6])
	}
	_ = buf
}

func BenchmarkEncodeVarint8(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 8, varintByteValues[7])
	}
	_ = buf
}

func BenchmarkEncodeVarint9(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 9, varintByteValues[8])
	}
	_ = buf
}

func BenchmarkEncodeVarint10(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarintV0(buf, 10, varintByteValues[9])
	}
	_ = buf
}

// ─── EncodeVarintV2 per-size benchmarks ──────────────────────────────────────

func BenchmarkEncodeVarintV21(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 1, varintByteValues[0])
	}
	_ = buf
}

func BenchmarkEncodeVarintV22(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 2, varintByteValues[1])
	}
	_ = buf
}

func BenchmarkEncodeVarintV23(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 3, varintByteValues[2])
	}
	_ = buf
}

func BenchmarkEncodeVarintV24(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 4, varintByteValues[3])
	}
	_ = buf
}

func BenchmarkEncodeVarintV25(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 5, varintByteValues[4])
	}
	_ = buf
}

func BenchmarkEncodeVarintV26(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 6, varintByteValues[5])
	}
	_ = buf
}

func BenchmarkEncodeVarintV27(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 7, varintByteValues[6])
	}
	_ = buf
}

func BenchmarkEncodeVarintV28(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 8, varintByteValues[7])
	}
	_ = buf
}

func BenchmarkEncodeVarintV29(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 9, varintByteValues[8])
	}
	_ = buf
}

func BenchmarkEncodeVarintV210(b *testing.B) {
	buf := make([]byte, 16)
	b.SetBytes(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeVarint(buf, 10, varintByteValues[9])
	}
	_ = buf
}

// ─── Throughput comparison table ─────────────────────────────────────────────

// TestAppendVarintThroughput runs all three varint encoding functions for each
// encoded byte length (1-10) and prints a markdown comparison table (MB/s).
// go test -test.fullpath=true -v -run ^TestAppendVarintThroughput$ github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils
func TestAppendVarintThroughput(t *testing.T) {
	appendFuncs := [10]func(*testing.B){
		BenchmarkAppendVarint1, BenchmarkAppendVarint2, BenchmarkAppendVarint3,
		BenchmarkAppendVarint4, BenchmarkAppendVarint5, BenchmarkAppendVarint6,
		BenchmarkAppendVarint7, BenchmarkAppendVarint8, BenchmarkAppendVarint9,
		BenchmarkAppendVarint10,
	}
	v1Funcs := [10]func(*testing.B){
		BenchmarkEncodeVarintV11, BenchmarkEncodeVarintV12, BenchmarkEncodeVarintV13,
		BenchmarkEncodeVarintV14, BenchmarkEncodeVarintV15, BenchmarkEncodeVarintV16,
		BenchmarkEncodeVarintV17, BenchmarkEncodeVarintV18, BenchmarkEncodeVarintV19,
		BenchmarkEncodeVarintV110,
	}
	v2Funcs := [10]func(*testing.B){
		BenchmarkEncodeVarint1, BenchmarkEncodeVarint2, BenchmarkEncodeVarint3,
		BenchmarkEncodeVarint4, BenchmarkEncodeVarint5, BenchmarkEncodeVarint6,
		BenchmarkEncodeVarint7, BenchmarkEncodeVarint8, BenchmarkEncodeVarint9,
		BenchmarkEncodeVarint10,
	}
	v3Funcs := [10]func(*testing.B){
		BenchmarkEncodeVarintV21, BenchmarkEncodeVarintV22, BenchmarkEncodeVarintV23,
		BenchmarkEncodeVarintV24, BenchmarkEncodeVarintV25, BenchmarkEncodeVarintV26,
		BenchmarkEncodeVarintV27, BenchmarkEncodeVarintV28, BenchmarkEncodeVarintV29,
		BenchmarkEncodeVarintV210,
	}

	mbPerSec := func(sz int, fn func(*testing.B)) float64 {
		r := testing.Benchmark(fn)
		if r.N == 0 {
			return 0
		}
		// float64 division for sub-nanosecond precision (NsPerOp truncates to int)
		nsPerOpF := float64(r.T.Nanoseconds()) / float64(r.N)
		return float64(sz) * 1e9 / nsPerOpF / 1e6
	}

	fmt.Println()
	fmt.Println("| Encoded bytes | AppendVarint (MB/s) | EncodeVarintV1 (MB/s) | EncodeVarint (MB/s) | EncodeVarintV2 (MB/s) |")
	fmt.Println("|:---:|---:|---:|---:|---:|")

	for i := range 10 {
		sz := i + 1
		fmt.Printf("| %d | %.2f | %.2f | %.2f | %.2f |\n",
			sz,
			mbPerSec(sz, appendFuncs[i]),
			mbPerSec(sz, v1Funcs[i]),
			mbPerSec(sz, v2Funcs[i]),
			mbPerSec(sz, v3Funcs[i]),
		)
	}
}
