{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct Readonly{{$goName}}
{
    // ── Clone ─────────────────────────────────────────────────────────────────

    public void Clone(ref {{$goName}} dst)
    {
{{- range .Fields}}
{{- if .IsMap}}
{{- if .MapValIsMsg}}
{{- if .UseMapValWrapper}}
        if (dst.{{.Name}} == null)
            dst.{{.Name}} = new Dictionary<{{.MapKeyCS}}, {{.WrapMapValCS}}>({{.Name}}?.Count ?? 0);
        else
            dst.{{.Name}}.Clear();
        if ({{.Name}} != null)
            foreach (var _kv{{.Name}} in {{.Name}})
            {
                {{.MapValCS}} _v{{.Name}} = default;
                if (_kv{{.Name}}.Value != null) _kv{{.Name}}.Value.Value.Clone(ref _v{{.Name}});
                dst.{{.Name}}[_kv{{.Name}}.Key] = new {{.WrapMapValCS}}() { Value = _v{{.Name}} };
            }
{{- else}}
        if (dst.{{.Name}} == null)
            dst.{{.Name}} = new Dictionary<{{.MapKeyCS}}, {{.MapValCS}}>({{.Name}}?.Count ?? 0);
        else
            dst.{{.Name}}.Clear();
        if ({{.Name}} != null)
            foreach (var _kv{{.Name}} in {{.Name}})
            {
                {{.MapValCS}} _v{{.Name}} = default;
                if (_kv{{.Name}}.Value != null) _kv{{.Name}}.Value.Clone(ref _v{{.Name}});
                dst.{{.Name}}[_kv{{.Name}}.Key] = _v{{.Name}};
            }
{{- end}}
{{- else}}
        if (dst.{{.Name}} == null)
            dst.{{.Name}} = new Dictionary<{{.MapKeyCS}}, {{.MapValCS}}>({{.Name}}?.Count ?? 0);
        else
            dst.{{.Name}}.Clear();
        if ({{.Name}} != null)
            foreach (var _kv{{.Name}} in {{.Name}})
                dst.{{.Name}}[_kv{{.Name}}.Key] = _kv{{.Name}}.Value;
{{- end}}
{{- else if .IsRepeated}}
{{- if .ElemIsMsg}}
        if (dst.{{.Name}} == null)
            dst.{{.Name}} = new List<{{.ElemTypeCS}}>({{.Name}}?.Count ?? 0);
        else
            dst.{{.Name}}.Clear();
        if ({{.Name}} != null)
            foreach (var _item{{.Name}} in {{.Name}})
            {
                {{.ElemTypeCS}} _e{{.Name}} = default;
                _item{{.Name}}.Clone(ref _e{{.Name}});
                dst.{{.Name}}.Add(_e{{.Name}});
            }
{{- else}}
        if (dst.{{.Name}} == null)
            dst.{{.Name}} = new List<{{.ElemTypeCS}}>({{.Name}}?.Count ?? 0);
        else
            dst.{{.Name}}.Clear();
        if ({{.Name}} != null)
            foreach (var _item{{.Name}} in {{.Name}})
                dst.{{.Name}}.Add(_item{{.Name}});
{{- end}}
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if ({{.Name}} != null)
        {
            if (dst.{{.Name}} == null) dst.{{.Name}} = new {{.EffWriterType}}();
            {{.Name}}.Value.Clone(ref dst.{{.Name}}.Value);
        }
        else
            dst.{{.Name}} = null;
{{- else}}
        {{.Name}}.Clone(ref dst.{{.Name}});
{{- end}}
{{- else}}
        dst.{{.Name}} = {{.Name}};
{{- end}}
{{- end}}
    }
}
{{end}}{{end}}
