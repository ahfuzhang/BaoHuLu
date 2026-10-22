{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct Readonly{{$goName}}
{
    // ── Reset ─────────────────────────────────────────────────────────────────

    public void Reset()
    {
{{- range .Fields}}
{{- if .IsMap}}
        {{.Name}}?.Clear();
{{- else if .IsRepeated}}
        {{.Name}}?.Clear();
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        {{.Name}} = null;
{{- else}}
        {{.Name}}.Reset();
{{- end}}
{{- else if .IsString}}
        {{.Name}} = string.Empty;
{{- else if .IsBytes}}
        {{.Name}} = Array.Empty<byte>();
{{- else if .IsBool}}
        {{.Name}} = false;
{{- else}}
        {{.Name}} = {{csDefault .ReadonlyType}};
{{- end}}
{{- end}}
    }
}
{{- $goName := .GoName}}
public partial struct {{$goName}}
{
    // ── Reset ─────────────────────────────────────────────────────────────────

    public void Reset()
    {
{{- range .Fields}}
{{- if .IsMap}}
        {{.Name}}?.Clear();
{{- else if .IsRepeated}}
        {{.Name}}?.Clear();
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        {{.Name}} = null;
{{- else}}
        {{.Name}}.Reset();
{{- end}}
{{- else if .IsString}}
        {{.Name}} = string.Empty;
{{- else}}
        {{.Name}} = {{csDefault .WriterType}};
{{- end}}
{{- end}}
    }
}
{{end}}{{end}}
