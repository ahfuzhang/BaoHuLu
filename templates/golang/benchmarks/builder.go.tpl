{{with model .}}
{{/* 单个 message 的 benchmark 样本构造；由普通 benchmark、jsonv2 和 compare 共享。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"strconv"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = strconv.IntSize
	_ = utils.AppendVarint
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
{{if hasAnyRecursiveField .Fields -}}
// benchBuild{{$goName}}Base builds {{$goName}} with only non-recursive fields populated.
// It is an independent helper called by benchBuild{{$goName}} to provide safe benchmark
// data for recursive sub-member fields without causing infinite recursion.
func benchBuild{{$goName}}Base() {{$goName}} {
	w := {{$goName}}{}
{{- range .Fields}}
{{- if and .IsMsg (not .Repeated)}}
{{- if or .IsRecursive .ElemIsRecursive}}
	// Skip: recursive field; populated by benchBuild{{$goName}} using the base builder.
{{- else}}
	w.{{.Name}} = benchBuild{{.GoType}}()
{{- end}}
{{- else if .Map}}
{{- if eq .MapKey "bool"}}
	// Leave bool-keyed maps nil: encoding/json does not support map[bool] keys.
{{- else if .ElemIsRecursive}}
	// Skip: recursive map value; populated by benchBuild{{$goName}} using the base builder.
{{- else}}
	{
		m := make({{mapWriterGoType .}}, 101)
		{{benchMapFill .}}
		w.{{.Name}} = m
	}
{{- end}}
{{- else if .Repeated}}
{{- if .ElemIsRecursive}}
	// Skip: recursive repeated elem; populated by benchBuild{{$goName}} using the base builder.
{{- else}}
	{
		s := make({{.GoType}}, 101)
		{{benchSliceFill .}}
		w.{{.Name}} = s
	}
{{- end}}
{{- else if eq .Type "string"}}
	w.{{.Name}} = "benchmark payload with escape chars:\n newline \t tab \" double-quote \\ backslash; " +
		"padding to ensure length exceeds one \b hundred \f bytes: \r 0123456789abcdef0123456789abcdef 中文"
{{- else if eq .Type "bytes"}}
	w.{{.Name}} = []byte("benchmark bytes payload with special chars: \n \t \\ " +
		"padding to exceed one hundred bytes: 0123456789abcdef0123456789abcdef")
{{- else}}
	w.{{.Name}} = {{sampleLit .}}
{{- end}}
{{- end}}
	return w
}

// benchBuild{{$goName}} returns a large {{$goName}} for benchmarking.
// Non-bool-keyed maps have 101 entries, slices have 101 elements, string/bytes
// fields are ≥100 bytes and include escape sequences (\n \t \" \\).
// Bool-keyed maps are left nil so encoding/json benchmarks can run.
// Recursive sub-member fields are built using the independent benchBuild{{$goName}}Base
// to avoid infinite recursion.
func benchBuild{{$goName}}() {{$goName}} {
	w := benchBuild{{$goName}}Base()
{{- range .Fields}}
{{- if and .IsMsg .IsRecursive}}
	{ v := benchBuild{{.GoType}}Base(); w.{{.Name}} = &v }
{{- else if and .Map .ElemIsRecursive (ne .MapKey "bool")}}
	{
		m := make({{mapWriterGoType .}}, 1)
		{{benchMapFillRecursive .}}
		w.{{.Name}} = m
	}
{{- else if and .Repeated .ElemIsRecursive}}
	{
		s := make({{.GoType}}, 1)
		{{benchSliceFillRecursive .}}
		w.{{.Name}} = s
	}
{{- end}}
{{- end}}
	return w
}
{{- else -}}
// benchBuild{{$goName}} returns a large {{$goName}} for benchmarking.
// Non-bool-keyed maps have 101 entries, slices have 101 elements, string/bytes
// fields are ≥100 bytes and include escape sequences (\n \t \" \\).
// Bool-keyed maps are left nil so encoding/json benchmarks can run.
func benchBuild{{$goName}}() {{$goName}} {
	w := {{$goName}}{}
{{- range .Fields}}
{{- if and .IsMsg (not .Repeated)}}
	w.{{.Name}} = benchBuild{{.GoType}}()
{{- else if .Map}}
{{- if eq .MapKey "bool"}}
	// Leave bool-keyed maps nil: encoding/json does not support map[bool] keys.
{{- else}}
	{
		m := make({{mapWriterGoType .}}, 101)
		{{benchMapFill .}}
		w.{{.Name}} = m
	}
{{- end}}
{{- else if .Repeated}}
	{
		s := make({{.GoType}}, 101)
		{{benchSliceFill .}}
		w.{{.Name}} = s
	}
{{- else if eq .Type "string"}}
	w.{{.Name}} = "benchmark payload with escape chars:\n newline \t tab \" double-quote \\ backslash; " +
		"padding to ensure length exceeds one hundred bytes: 0123456789abcdef0123456789abcdef"
{{- else if eq .Type "bytes"}}
	w.{{.Name}} = []byte("benchmark bytes payload with special chars: \n \t \\ " +
		"padding to exceed one hundred bytes: 0123456789abcdef0123456789abcdef")
{{- else}}
	w.{{.Name}} = {{sampleLit .}}
{{- end}}
{{- end}}
	return w
}
{{- end}}
{{end}}
{{end}}
