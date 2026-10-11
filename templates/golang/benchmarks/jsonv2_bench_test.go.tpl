{{/* 仅在 jsonv2 实验启用时编译的两个 benchmark。 */ -}}
//go:build goexperiment.jsonv2

{{template "partials/header.go.tpl" .}}

import (
	jsonv2 "encoding/json/v2"
	"testing"
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = jsonv2.Marshal
)

{{with .Msg}}
{{- $goName := .GoName}}
// BenchmarkToJSON_{{$goName}}_EncodingJSONv2 measures encoding/json/v2 Marshal throughput.
func BenchmarkToJSON_{{$goName}}_EncodingJSONv2(b *testing.B) {
	src := benchBuild{{$goName}}()
	j, err := jsonv2.Marshal(src)
	if err != nil {
		b.Skipf("encoding/json/v2.Marshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jsonv2.Marshal(src); err != nil {
			b.Fatal(err)
		}
	}
}


// BenchmarkFromJSON_{{$goName}}_EncodingJSONv2 measures encoding/json/v2 Unmarshal throughput.
func BenchmarkFromJSON_{{$goName}}_EncodingJSONv2(b *testing.B) {
	src := benchBuild{{$goName}}()
	j := src.ToJSON(nil)
	var w {{$goName}}
	if err := jsonv2.Unmarshal(j, &w); err != nil {
		b.Skipf("encoding/json/v2.Unmarshal unsupported for {{$goName}}: %v", err)
	}
	b.SetBytes(int64(len(j)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w = {{$goName}}{}
		if err := jsonv2.Unmarshal(j, &w); err != nil {
			b.Fatal(err)
		}
	}
	_ = w
}
{{end}}
