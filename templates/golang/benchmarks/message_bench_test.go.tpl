{{/* 普通 JSON / Protobuf benchmark，使用同包 benchBuild 辅助函数。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = json.Marshal
	_ = sonic.Marshal
)

{{with .Msg}}
{{- $goName := .GoName}}
{{- $roName := printf "Readonly%s" $goName}}
// ── ToJSON benchmarks ─────────────────────────────────────────────────────────

// BenchmarkToJSON_{{$goName}}_Generated measures ToJSON throughput of generated code.
// The output buffer is reused across iterations to demonstrate zero-allocation amortisation.
func BenchmarkToJSON_{{$goName}}_Generated(b *testing.B) {
	src := benchBuild{{$goName}}()
	j := src.ToJSON(nil)
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	buf := make([]byte, 0, len(j))
	for i := 0; i < b.N; i++ {
		buf = src.ToJSON(buf[:0])
	}
	_ = buf
}

// BenchmarkToJSON_{{$goName}}_EncodingJSON measures json.Marshal throughput.
func BenchmarkToJSON_{{$goName}}_EncodingJSON(b *testing.B) {
	src := benchBuild{{$goName}}()
	j, err := json.Marshal(src)
	if err != nil {
		b.Skipf("encoding/json.Marshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(src); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToJSON_{{$goName}}_Sonic measures sonic.Marshal throughput.
func BenchmarkToJSON_{{$goName}}_Sonic(b *testing.B) {
	src := benchBuild{{$goName}}()
	j, err := sonic.Marshal(src)
	if err != nil {
		b.Skipf("sonic.Marshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.Marshal(src); err != nil {
			b.Fatal(err)
		}
	}
}

// ── FromJSON benchmarks ───────────────────────────────────────────────────────

// BenchmarkFromJSON_{{$goName}}_Generated measures FromJSON throughput of generated code.
// The Readonly struct is reset between iterations to reuse its memory without extra allocs.
func BenchmarkFromJSON_{{$goName}}_Generated(b *testing.B) {
	src := benchBuild{{$goName}}()
	j := src.ToJSON(nil)
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	var r {{$roName}}
	for i := 0; i < b.N; i++ {
		r.Reset()
		if err := r.FromJSON(j); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFromJSON_{{$goName}}_EncodingJSON measures json.Unmarshal throughput.
func BenchmarkFromJSON_{{$goName}}_EncodingJSON(b *testing.B) {
	src := benchBuild{{$goName}}()
	j := src.ToJSON(nil)
	var w {{$goName}}
	if err := json.Unmarshal(j, &w); err != nil {
		b.Skipf("encoding/json.Unmarshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w = {{$goName}}{}
		if err := json.Unmarshal(j, &w); err != nil {
			b.Fatal(err)
		}
	}
	_ = w
}

// BenchmarkFromJSON_{{$goName}}_Sonic measures sonic.Unmarshal throughput.
func BenchmarkFromJSON_{{$goName}}_Sonic(b *testing.B) {
	src := benchBuild{{$goName}}()
	j := src.ToJSON(nil)
	var w {{$goName}}
	if err := sonic.Unmarshal(j, &w); err != nil {
		b.Skipf("sonic.Unmarshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w = {{$goName}}{}
		if err := sonic.Unmarshal(j, &w); err != nil {
			b.Fatal(err)
		}
	}
	_ = w
}

// ── Protobuf benchmarks ───────────────────────────────────────────────────────

// BenchmarkToProtobuf_{{$goName}} measures ToProtobuf throughput of generated code.
// The output buffer is reused across iterations.
func BenchmarkToProtobuf_{{$goName}}(b *testing.B) {
	src := benchBuild{{$goName}}()
	pb := src.ToProtobuf(nil)
	b.SetBytes(int64(len(pb)))
	b.ResetTimer()
	b.ReportAllocs()
	buf := make([]byte, 0, len(pb))
	for i := 0; i < b.N; i++ {
		buf = src.ToProtobuf(buf[:0])
	}
	_ = buf
}

// BenchmarkFromProtobuf_{{$goName}} measures FromProtobuf throughput of generated code.
// The Readonly struct is reset between iterations to reuse its memory without extra allocs.
func BenchmarkFromProtobuf_{{$goName}}(b *testing.B) {
	src := benchBuild{{$goName}}()
	pb := src.ToProtobuf(nil)
	b.SetBytes(int64(len(pb)))
	b.ResetTimer()
	b.ReportAllocs()
	var r {{$roName}}
	for i := 0; i < b.N; i++ {
		r.Reset()
		if err := r.FromProtobuf(pb); err != nil {
			b.Fatal(err)
		}
	}
}
{{end}}
