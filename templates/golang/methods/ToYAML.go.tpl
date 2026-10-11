{{/* YAML 编码。共享 YAML key 和注释常量只在本文件定义。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/base64"
	"strconv"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = base64.StdEncoding
	_ = strconv.IntSize
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
{{- $hasYamlNames := hasYamlFields .Fields}}
{{- if $hasYamlNames}}

// yaml field key constants for {{$goName}}.
const (
{{- range .Fields}}
{{- if .IsRawBuf}}{{continue}}{{end}}
	_{{$goName}}_{{.Name}}_yamlKey = "{{yamlKey .}}"
{{- end}}
)
{{end}}
{{- $msgCmtNums := .CommentLineNums}}
{{- if hasAnyYamlComments .}}

// yaml comment constants for {{$goName}}.
const (
{{- range $i, $line := .Comment}}
{{- if gt (index $msgCmtNums $i) 0}}
	_yamlCommentLine{{index $msgCmtNums $i}} = {{printf "%q" (printf "%s\n" $line)}}
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsRawBuf}}{{continue}}{{end}}
{{- $f := .}}
{{- range $i, $line := .Comment}}
{{- if gt (index $f.CommentLineNums $i) 0}}
	_yamlCommentLine{{index $f.CommentLineNums $i}} = {{printf "%q" (printf "%s\n" $line)}}
{{- end}}
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
	_yamlInlineCommentLine{{.InlineCommentLineNum}} = {{printf "%q" (printf "%s\n" .InlineComment)}}
{{- end}}
{{- end}}
)
{{end}}
// ============================================================================
// {{$goName}} — ToYAML serialization
// ============================================================================

// ToYAML appends the YAML encoding of m to dst using indent as the base
// indentation width (in spaces). Returns the extended slice.
func (m *{{$goName}}) ToYAML(dst []byte, indent int) []byte {
	sp := make([]byte, indent)
	for _i := range sp {
		sp[_i] = ' '
	}
	sp2 := make([]byte, indent+2)
	for _i := range sp2 {
		sp2[_i] = ' '
	}
	_ = sp
	_ = sp2
	// _strVal appends a YAML-safe string scalar value, quoting when necessary.
	_strVal := func(dst []byte, s string) []byte {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < 0x20 || c == ':' || c == '#' || c == '"' || c == '\'' || c == '\\' ||
				c == '{' || c == '}' || c == '[' || c == ']' {
				return strconv.AppendQuote(dst, s)
			}
		}
		return append(dst, s...)
	}
	_ = _strVal
{{- if .AsMap}}
{{- $f0 := index .Fields 0}}
{{- if .Comment}}
	if len(m.{{$f0.Name}}) > 0 {
{{- range $i, $line := .Comment}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $msgCmtNums $i $line}}...)
{{- end}}
	}
{{- end}}
	for _k, _v := range m.{{$f0.Name}} {
		dst = append(dst, sp...)
{{- if eq $f0.MapKey "string"}}
		dst = _strVal(dst, _k)
{{- else if eq $f0.MapKey "bool"}}
		if _k {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
{{- else if or (eq $f0.MapKey "uint64") (eq $f0.MapKey "fixed64")}}
		dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else if or (eq $f0.MapKey "uint32") (eq $f0.MapKey "fixed32")}}
		dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else if or (eq $f0.MapKey "int64") (eq $f0.MapKey "sint64") (eq $f0.MapKey "sfixed64")}}
		dst = strconv.AppendInt(dst, int64(_k), 10)
{{- else}}
		dst = strconv.AppendInt(dst, int64(_k), 10)
{{- end}}
		dst = append(dst, ": "...)
{{- if $f0.MapValIsMsg}}
		if _v != nil {
			dst = append(dst, '\n')
			dst = _v.ToYAML(dst, indent+2)
		} else {
			dst = append(dst, '\n')
		}
{{- else if eq $f0.MapVal "string"}}
		dst = _strVal(dst, _v)
		dst = append(dst, '\n')
{{- else if eq $f0.MapVal "bytes"}}
		dst = append(dst, base64.StdEncoding.EncodeToString(_v)...)
		dst = append(dst, '\n')
{{- else if eq $f0.MapVal "bool"}}
		if _v {
			dst = append(dst, "true\n"...)
		} else {
			dst = append(dst, "false\n"...)
		}
{{- else if eq $f0.MapVal "double"}}
		dst = strconv.AppendFloat(dst, _v, 'f', -1, 64)
		dst = append(dst, '\n')
{{- else if eq $f0.MapVal "float"}}
		dst = strconv.AppendFloat(dst, float64(_v), 'f', -1, 32)
		dst = append(dst, '\n')
{{- else if or (eq $f0.MapVal "uint64") (eq $f0.MapVal "fixed64")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
		dst = append(dst, '\n')
{{- else if or (eq $f0.MapVal "uint32") (eq $f0.MapVal "fixed32")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
		dst = append(dst, '\n')
{{- else if or (eq $f0.MapVal "int64") (eq $f0.MapVal "sint64") (eq $f0.MapVal "sfixed64")}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
		dst = append(dst, '\n')
{{- else}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if .AsArray}}
{{- $fa := index .Fields 0}}
{{- if .Comment}}
	if len(m.{{$fa.Name}}) > 0 {
{{- range $i, $line := .Comment}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $msgCmtNums $i $line}}...)
{{- end}}
	}
{{- end}}
	for _, _v := range m.{{$fa.Name}} {
		dst = append(dst, sp...)
		dst = append(dst, "- "...)
{{- if $fa.IsMsg}}
		dst = append(dst, '\n')
		dst = _v.ToYAML(dst, indent+2)
{{- else if eq $fa.Type "string"}}
		dst = _strVal(dst, _v)
		dst = append(dst, '\n')
{{- else if eq $fa.Type "bytes"}}
		dst = append(dst, base64.StdEncoding.EncodeToString(_v)...)
		dst = append(dst, '\n')
{{- else if eq $fa.Type "bool"}}
		if _v {
			dst = append(dst, "true\n"...)
		} else {
			dst = append(dst, "false\n"...)
		}
{{- else if eq $fa.Type "double"}}
		dst = strconv.AppendFloat(dst, _v, 'f', -1, 64)
		dst = append(dst, '\n')
{{- else if eq $fa.Type "float"}}
		dst = strconv.AppendFloat(dst, float64(_v), 'f', -1, 32)
		dst = append(dst, '\n')
{{- else if or (eq $fa.Type "uint64") (eq $fa.Type "fixed64")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
		dst = append(dst, '\n')
{{- else if or (eq $fa.Type "uint32") (eq $fa.Type "fixed32")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
		dst = append(dst, '\n')
{{- else if or (eq $fa.Type "int64") (eq $fa.Type "sint64") (eq $fa.Type "sfixed64")}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
		dst = append(dst, '\n')
{{- else}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
		dst = append(dst, '\n')
{{- end}}
	}
{{- else}}
{{- range $i, $line := .Comment}}
	dst = append(dst, sp...)
	dst = append(dst, '#')
	dst = append(dst, {{yamlCommentExpr $msgCmtNums $i $line}}...)
{{- end}}
{{range .Fields}}
{{- if .IsRawBuf}}{{continue}}{{end}}
{{- $cmt := .Comment}}
{{- $cmtNums := .CommentLineNums}}
{{- if .Map}}
	if len(m.{{.Name}}) > 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":  "}}...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":\n"}}...)
{{- end}}
		for _k, _v := range m.{{.Name}} {
			dst = append(dst, sp2...)
{{- if eq .MapKey "string"}}
			dst = _strVal(dst, _k)
{{- else if eq .MapKey "bool"}}
			if _k {
				dst = append(dst, "true"...)
			} else {
				dst = append(dst, "false"...)
			}
{{- else if or (eq .MapKey "uint64") (eq .MapKey "fixed64")}}
			dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else if or (eq .MapKey "uint32") (eq .MapKey "fixed32")}}
			dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else if or (eq .MapKey "int64") (eq .MapKey "sint64") (eq .MapKey "sfixed64")}}
			dst = strconv.AppendInt(dst, int64(_k), 10)
{{- else}}
			dst = strconv.AppendInt(dst, int64(_k), 10)
{{- end}}
			dst = append(dst, ": "...)
{{- if .MapValIsMsg}}
			dst = append(dst, '\n')
			dst = _v.ToYAML(dst, indent+4)
{{- else if eq .MapVal "string"}}
			dst = _strVal(dst, _v)
			dst = append(dst, '\n')
{{- else if eq .MapVal "bytes"}}
			dst = append(dst, base64.StdEncoding.EncodeToString(_v)...)
			dst = append(dst, '\n')
{{- else if eq .MapVal "bool"}}
			if _v {
				dst = append(dst, "true\n"...)
			} else {
				dst = append(dst, "false\n"...)
			}
{{- else if eq .MapVal "double"}}
			dst = strconv.AppendFloat(dst, _v, 'f', -1, 64)
			dst = append(dst, '\n')
{{- else if eq .MapVal "float"}}
			dst = strconv.AppendFloat(dst, float64(_v), 'f', -1, 32)
			dst = append(dst, '\n')
{{- else if or (eq .MapVal "uint64") (eq .MapVal "fixed64")}}
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '\n')
{{- else if or (eq .MapVal "uint32") (eq .MapVal "fixed32")}}
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '\n')
{{- else if or (eq .MapVal "int64") (eq .MapVal "sint64") (eq .MapVal "sfixed64")}}
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '\n')
{{- else}}
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '\n')
{{- end}}
		}
	}
{{- else if .Repeated}}
	if len(m.{{.Name}}) > 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":  "}}...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":\n"}}...)
{{- end}}
		for _, _v := range m.{{.Name}} {
			dst = append(dst, sp2...)
			dst = append(dst, "- "...)
{{- if .IsMsg}}
			dst = append(dst, '\n')
			dst = _v.ToYAML(dst, indent+4)
{{- else if eq .Type "string"}}
			dst = _strVal(dst, _v)
			dst = append(dst, '\n')
{{- else if eq .Type "bytes"}}
			dst = append(dst, base64.StdEncoding.EncodeToString(_v)...)
			dst = append(dst, '\n')
{{- else if eq .Type "bool"}}
			if _v {
				dst = append(dst, "true\n"...)
			} else {
				dst = append(dst, "false\n"...)
			}
{{- else if eq .Type "double"}}
			dst = strconv.AppendFloat(dst, _v, 'f', -1, 64)
			dst = append(dst, '\n')
{{- else if eq .Type "float"}}
			dst = strconv.AppendFloat(dst, float64(_v), 'f', -1, 32)
			dst = append(dst, '\n')
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '\n')
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '\n')
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '\n')
{{- else}}
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '\n')
{{- end}}
		}
	}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if m.{{.Name}} != nil {
		_sub{{.Name}} := m.{{.Name}}.ToYAML(nil, indent+2)
		if len(_sub{{.Name}}) > 0 {
{{- range $i, $line := $cmt}}
			dst = append(dst, sp...)
			dst = append(dst, '#')
			dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
			dst = append(dst, sp...)
{{- if gt .InlineCommentLineNum 0}}
			dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":  "}}...)
			dst = append(dst, '#')
			dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
			dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":\n"}}...)
{{- end}}
			dst = append(dst, _sub{{.Name}}...)
		}
	}
{{- else}}
	{
		_sub{{.Name}} := m.{{.Name}}.ToYAML(nil, indent+2)
		if len(_sub{{.Name}}) > 0 {
{{- range $i, $line := $cmt}}
			dst = append(dst, sp...)
			dst = append(dst, '#')
			dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
			dst = append(dst, sp...)
{{- if gt .InlineCommentLineNum 0}}
			dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":  "}}...)
			dst = append(dst, '#')
			dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
			dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ":\n"}}...)
{{- end}}
			dst = append(dst, _sub{{.Name}}...)
		}
	}
{{- end}}
{{- else if .IsDecimal}}
	if !m.{{.Name}}.IsZero() {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		_df{{.Name}}, _ := m.{{.Name}}.Round({{decimalRound .}}).Float64()
		dst = strconv.AppendFloat(dst, _df{{.Name}}, 'f', {{decimalRound .}}, 64)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if eq .Type "bool"}}
	if m.{{.Name}} {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": true  "}}...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": true\n"}}...)
{{- end}}
	}
{{- else if eq .Type "string"}}
	if m.{{.Name}} != "" {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = _strVal(dst, m.{{.Name}})
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if eq .Type "bytes"}}
	if len(m.{{.Name}}) > 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = append(dst, base64.StdEncoding.EncodeToString(m.{{.Name}})...)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if eq .Type "double"}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendFloat(dst, m.{{.Name}}, 'f', -1, 64)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if eq .Type "float"}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendFloat(dst, float64(m.{{.Name}}), 'f', -1, 32)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendUint(dst, m.{{.Name}}, 10)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendUint(dst, uint64(m.{{.Name}}), 10)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendInt(dst, m.{{.Name}}, 10)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- else}}
	if m.{{.Name}} != 0 {
{{- range $i, $line := $cmt}}
		dst = append(dst, sp...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlCommentExpr $cmtNums $i $line}}...)
{{- end}}
		dst = append(dst, sp...)
		dst = append(dst, {{yamlKeyAppendExpr $hasYamlNames $goName . ": "}}...)
		dst = strconv.AppendInt(dst, int64(m.{{.Name}}), 10)
{{- if gt .InlineCommentLineNum 0}}
		dst = append(dst, "  "...)
		dst = append(dst, '#')
		dst = append(dst, {{yamlInlineCommentExpr .}}...)
{{- else}}
		dst = append(dst, '\n')
{{- end}}
	}
{{- end}}
{{end}}
{{- end}}
	return dst
}
{{end}}
