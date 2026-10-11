{{/* 单个 message：字段常量、写类型及 Readonly 类型声明。 */ -}}
{{template "partials/header.go.tpl" .}}

import (

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/fastjson"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
// {{$goName}} field tag IDs.
{{- $tagMaxLen := maxTagLen $goName .Fields}}
const (
{{- range .Fields}}
	{{padRight (printf "%s%sTag" $goName .Name) $tagMaxLen}} = {{.Number}}
{{- end}}
)

// {{$goName}} JSON field name string constants.
{{- $nameOfMaxLen := maxNameOfLen $goName .Fields}}
const (
{{- range .Fields}}
	{{padRight (printf "NameOf%s%s" $goName .Name) $nameOfMaxLen}} = "{{.JsonName}}"
{{- end}}
)
{{if hasYamlFields .Fields}}
// {{$goName}} YAML field name string constants.
const (
{{- range .Fields}}
{{- if .YamlName}}
	NameOfYamlField{{$goName}}{{.Name}} = "{{.YamlName}}"
{{- end}}
{{- end}}
)
{{end}}
{{msgCommentBlock .Comment}}// {{$goName}} writer struct.
// Fields are ordered by alignment (desc) then size (desc) for minimal memory padding.
type {{$goName}} struct {
{{- range writerAlignedFields .Fields}}
{{- if .StructTag}}
{{fieldCommentBlock .Comment}}	{{padRight .Name .NameW}} {{padRight .TypeStr .TypeW}} {{.StructTag}}
{{- else}}
{{fieldCommentBlock .Comment}}	{{padRight .Name .NameW}} {{.TypeStr}}
{{- end}}
{{- end}}
}

{{msgCommentBlock .Comment}}// Readonly{{$goName}} readonly struct.
// Fields are sorted with the fieldalignment strategy for
// minimal memory padding and minimal GC scan range.
type Readonly{{$goName}} struct {
	parser    fastjson.Parser
	strArena *fastjson.Arena
{{- range readerAlignedFields .ReaderFields}}
{{- if .StructTag}}
{{fieldCommentBlock .Comment}}	{{padRight .Name .NameW}} {{padRight .TypeStr .TypeW}} {{.StructTag}}
{{- else}}
{{fieldCommentBlock .Comment}}	{{padRight .Name .NameW}} {{.TypeStr}}
{{- end}}
{{- end}}
}
{{end}}
