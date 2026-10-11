{{/* 项目文件使用相同根上下文的 .Project 字段。 */ -}}
{{- $version := .Project.Version -}}
{{- if or (eq $version "dev") (eq $version "(devel)") -}}
{{- $version = "v0.0.0-00010101000000-000000000000" -}}
{{- end -}}
module {{.Project.ModulePath}}

go 1.23

require (
	github.com/ahfuzhang/BaoHuLu {{$version}}
{{- if or .Project.WithTest .Project.WithBench}}
	github.com/bytedance/sonic v1.15.1
{{- end}}
{{- if .Project.WithDecimal}}
	github.com/govalues/decimal v0.1.3
{{- end}}
)
{{if or .Project.WithTest .Project.WithBench}}
require (
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic/loader v0.5.1 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/sys v0.22.0 // indirect
)
{{end}}
