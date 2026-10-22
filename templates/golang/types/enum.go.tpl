{{with model .}}
{{/* 单个 enum：类型声明及全部枚举值。 */ -}}
{{template "partials/header.go.tpl" .}}

{{with .Enum}}
type {{.Name}} int32
{{$enumName := .Name}}
{{- $enumMaxLen := maxEnumValLen .Values}}
const (
{{- range .Values}}
	{{padRight (enumValueGoName .Name) $enumMaxLen}} {{$enumName}} = {{.Number}}
{{- end}}
)
{{end}}
{{end}}
