{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Enum}}
public enum {{.Name}} : int
{
{{- range .Values}}
    {{.Name}} = {{.Number}},
{{- end}}
}
{{end}}{{end}}
