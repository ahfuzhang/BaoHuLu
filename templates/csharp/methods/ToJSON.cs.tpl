{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct {{$goName}}
{
    // ── ToJSON ───────────────────────────────────────────────────────────────

    public readonly void ToJSON(ref RentedBuffer buf)
    {
{{- if .AsArray}}
{{- $f := index .Fields 0}}
        // @AsArray: serialize the single repeated field directly as a bare JSON array.
        buf.Append((byte)'[');
        if ({{$f.Name}} != null)
            for (int _i{{$f.Name}} = 0; _i{{$f.Name}} < {{$f.Name}}.Count; _i{{$f.Name}}++)
            {
                if (_i{{$f.Name}} > 0) buf.Append((byte)',');
                var _re{{$f.Name}} = {{$f.Name}}[_i{{$f.Name}}];
{{- if $f.ElemIsMsg}}
                _re{{$f.Name}}.ToJSON(ref buf);
{{- else if $f.IsString}}
                buf.Append((byte)'"'); buf.AppendAsJsonEscapedString(_re{{$f.Name}}); buf.Append((byte)'"');
{{- else if $f.IsBytes}}
                buf.Append((byte)'"'); buf.Append(Convert.ToBase64String(_re{{$f.Name}})); buf.Append((byte)'"');
{{- else if $f.IsBool}}
                buf.Append(_re{{$f.Name}});
{{- else if $f.IsFixed64}}
                buf.Append((byte)'"'); buf.Append((long)_re{{$f.Name}}); buf.Append((byte)'"');
{{- else if eq $f.ElemTypeCS "double"}}
                buf.Append(_re{{$f.Name}});
{{- else if eq $f.ElemTypeCS "float"}}
                buf.Append((double)_re{{$f.Name}});
{{- else if eq $f.ElemTypeCS "long"}}
                buf.Append((byte)'"'); buf.Append(_re{{$f.Name}}); buf.Append((byte)'"');
{{- else if eq $f.ElemTypeCS "ulong"}}
                buf.Append((byte)'"'); buf.Append(_re{{$f.Name}}); buf.Append((byte)'"');
{{- else}}
                buf.Append((long)_re{{$f.Name}});
{{- end}}
            }
        buf.Append((byte)']');
{{- else}}
        buf.Append((byte)'{');
{{- if .AsMap}}
{{- $f := index .Fields 0}}
        // @AsMap: serialize the map directly without a field-name wrapper.
        bool _mfirst{{$f.Name}} = true;
        if ({{$f.Name}} != null)
            foreach (var _kv{{$f.Name}} in {{$f.Name}})
            {
{{- if eq $f.MapKeyCS "string"}}
                if (_kv{{$f.Name}}.Key == null) continue;
{{- end}}
                if (!_mfirst{{$f.Name}}) buf.Append((byte)','); _mfirst{{$f.Name}} = false;
                buf.Append((byte)'"');
{{- if eq $f.MapKeyCS "string"}}
                buf.Append(_kv{{$f.Name}}.Key);  // no need to JSON-escape map keys since protobuf maps only allow simple scalar types as keys
{{- else if eq $f.MapKeyCS "bool"}}
                buf.Append(_kv{{$f.Name}}.Key ? "true"u8 : "false"u8);
{{- else}}
                buf.Append((long)_kv{{$f.Name}}.Key);
{{- end}}
                buf.Append((byte)'"'); buf.Append((byte)':');
{{- if $f.MapValIsMsg}}
{{- if $f.UseMapValWrapper}}
                if (_kv{{$f.Name}}.Value == null) buf.Append("null"u8); else _kv{{$f.Name}}.Value.Value.ToJSON(ref buf);
{{- else}}
                if (_kv{{$f.Name}}.Value == null) buf.Append("null"u8); else _kv{{$f.Name}}.Value.ToJSON(ref buf);
{{- end}}
{{- else if eq $f.MapVal "string"}}
                if (_kv{{$f.Name}}.Value == null) buf.Append("null"u8); else { buf.Append((byte)'"'); buf.AppendAsJsonEscapedString(_kv{{$f.Name}}.Value); buf.Append((byte)'"'); }
{{- else if eq $f.MapVal "bytes"}}
                if (_kv{{$f.Name}}.Value == null) buf.Append("null"u8); else { buf.Append((byte)'"'); buf.Append(Convert.ToBase64String(_kv{{$f.Name}}.Value)); buf.Append((byte)'"'); }
{{- else if eq $f.MapVal "bool"}}
                buf.Append(_kv{{$f.Name}}.Value);
{{- else if eq $f.MapVal "double"}}
                buf.Append(_kv{{$f.Name}}.Value);
{{- else if eq $f.MapVal "float"}}
                buf.Append((double)_kv{{$f.Name}}.Value);
{{- else if eq $f.MapVal "int64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{$f.Name}}.Value); buf.Append((byte)'"');
{{- else if eq $f.MapVal "uint64"}}
                buf.Append((byte)'"'); buf.Append(_kv{{$f.Name}}.Value); buf.Append((byte)'"');
{{- else if eq $f.MapVal "sint64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{$f.Name}}.Value); buf.Append((byte)'"');
{{- else if eq $f.MapVal "sfixed64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{$f.Name}}.Value); buf.Append((byte)'"');
{{- else if eq $f.MapVal "fixed64"}}
                buf.Append((byte)'"'); buf.Append(_kv{{$f.Name}}.Value); buf.Append((byte)'"');
{{- else}}
                buf.Append((long)_kv{{$f.Name}}.Value);
{{- end}}
            }
{{- else}}
        bool _first = true;
{{- range .Fields}}
{{- if .IsMap}}
        if ({{.Name}} != null && {{.Name}}.Count > 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":{"u8);
            bool _mfirst{{.Name}} = true;
            foreach (var _kv{{.Name}} in {{.Name}})
            {
{{- if eq .MapKeyCS "string"}}
                if (_kv{{.Name}}.Key == null) continue;
{{- end}}
                if (!_mfirst{{.Name}}) buf.Append((byte)','); _mfirst{{.Name}} = false;
                buf.Append((byte)'"');
{{- if eq .MapKeyCS "string"}}
                buf.AppendAsJsonEscapedString(_kv{{.Name}}.Key);
{{- else if eq .MapKeyCS "bool"}}
                buf.Append(_kv{{.Name}}.Key ? "true"u8 : "false"u8);
{{- else}}
                buf.Append((long)_kv{{.Name}}.Key);
{{- end}}
                buf.Append((byte)'"'); buf.Append((byte)':');
{{- if .MapValIsMsg}}
{{- if .UseMapValWrapper}}
                if (_kv{{.Name}}.Value == null) buf.Append("null"u8); else _kv{{.Name}}.Value.Value.ToJSON(ref buf);
{{- else}}
                if (_kv{{.Name}}.Value == null) buf.Append("null"u8); else _kv{{.Name}}.Value.ToJSON(ref buf);
{{- end}}
{{- else if eq .MapVal "string"}}
                if (_kv{{.Name}}.Value == null) buf.Append("null"u8); else { buf.Append((byte)'"'); buf.AppendAsJsonEscapedString(_kv{{.Name}}.Value); buf.Append((byte)'"'); }
{{- else if eq .MapVal "bytes"}}
                if (_kv{{.Name}}.Value == null) buf.Append("null"u8); else { buf.Append((byte)'"'); buf.Append(Convert.ToBase64String(_kv{{.Name}}.Value)); buf.Append((byte)'"'); }
{{- else if eq .MapVal "bool"}}
                buf.Append(_kv{{.Name}}.Value);
{{- else if eq .MapVal "double"}}
                buf.Append(_kv{{.Name}}.Value);
{{- else if eq .MapVal "float"}}
                buf.Append((double)_kv{{.Name}}.Value);
{{- else if eq .MapVal "int64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{.Name}}.Value); buf.Append((byte)'"');
{{- else if eq .MapVal "uint64"}}
                buf.Append((byte)'"'); buf.Append(_kv{{.Name}}.Value); buf.Append((byte)'"');
{{- else if eq .MapVal "sint64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{.Name}}.Value); buf.Append((byte)'"');
{{- else if eq .MapVal "sfixed64"}}
                buf.Append((byte)'"'); buf.Append((long)_kv{{.Name}}.Value); buf.Append((byte)'"');
{{- else if eq .MapVal "fixed64"}}
                buf.Append((byte)'"'); buf.Append(_kv{{.Name}}.Value); buf.Append((byte)'"');
{{- else}}
                buf.Append((long)_kv{{.Name}}.Value);
{{- end}}
            }
            buf.Append((byte)'}');
        }
{{- else if .IsRepeated}}
        if ({{.Name}} != null && {{.Name}}.Count > 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":["u8);
            for (int _i{{.Name}} = 0; _i{{.Name}} < {{.Name}}.Count; _i{{.Name}}++)
            {
                if (_i{{.Name}} > 0) buf.Append((byte)',');
                var _re{{.Name}} = {{.Name}}[_i{{.Name}}];
{{- if .ElemIsMsg}}
                _re{{.Name}}.ToJSON(ref buf);
{{- else if .IsString}}
                buf.Append((byte)'"'); buf.AppendAsJsonEscapedString(_re{{.Name}}); buf.Append((byte)'"');
{{- else if .IsBytes}}
                buf.Append((byte)'"'); buf.Append(Convert.ToBase64String(_re{{.Name}})); buf.Append((byte)'"');
{{- else if .IsBool}}
                buf.Append(_re{{.Name}});
{{- else if .IsFixed64}}
                buf.Append((byte)'"'); buf.Append((long)_re{{.Name}}); buf.Append((byte)'"');
{{- else if eq .ElemTypeCS "double"}}
                buf.Append(_re{{.Name}});
{{- else if eq .ElemTypeCS "float"}}
                buf.Append((double)_re{{.Name}});
{{- else if eq .ElemTypeCS "long"}}
                buf.Append((byte)'"'); buf.Append(_re{{.Name}}); buf.Append((byte)'"');
{{- else if eq .ElemTypeCS "ulong"}}
                buf.Append((byte)'"'); buf.Append(_re{{.Name}}); buf.Append((byte)'"');
{{- else}}
                buf.Append((long)_re{{.Name}});
{{- end}}
            }
            buf.Append((byte)']');
        }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if ({{.Name}} != null)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            {{.Name}}.Value.ToJSON(ref buf);
        }
{{- else}}
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            {{.Name}}.ToJSON(ref buf);
        }
{{- end}}
{{- else if .IsString}}
        if (!string.IsNullOrEmpty({{.Name}}))
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":\""u8);
            buf.AppendAsJsonEscapedString({{.Name}});
            buf.Append((byte)'"');
        }
{{- else if .IsBytes}}
        if ({{.Name}} != null && {{.Name}}.Length > 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":\""u8);
            buf.Append(Convert.ToBase64String({{.Name}}));
            buf.Append((byte)'"');
        }
{{- else if .IsBool}}
        if ({{.Name}})
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":true"u8);
        }
{{- else if .IsEnum}}
        if ((int){{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            buf.Append((long)(int){{.Name}});
        }
{{- else if .IsDecimal}}
        if ({{.Name}} != 0m)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            buf.Append((double)Math.Round({{.Name}}, {{.DecimalRound}}, MidpointRounding.AwayFromZero));
        }
{{- else if eq .WriterType "double"}}
        if ({{.Name}} != 0.0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            buf.Append({{.Name}});
        }
{{- else if eq .WriterType "float"}}
        if ({{.Name}} != 0.0f)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            buf.Append((double){{.Name}});
        }
{{- else if .IsFixed64}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":\""u8);
            buf.Append((long){{.Name}});
            buf.Append((byte)'"');
        }
{{- else if eq .WriterType "long"}}
        if ({{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":\""u8);
            buf.Append({{.Name}});
            buf.Append((byte)'"');
        }
{{- else if eq .WriterType "ulong"}}
        if ({{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":\""u8);
            buf.Append({{.Name}});
            buf.Append((byte)'"');
        }
{{- else}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            if (!_first) buf.Append((byte)','); _first = false;
            buf.Append((byte)'"'); buf.Append({{$goName}}Tags.JsonKey{{.Name}}); buf.Append("\":"u8);
            buf.Append((long){{.Name}});
        }
{{- end}}
{{- end}}
{{- end}}
        buf.Append((byte)'}');
{{- end}}
    }
}
{{end}}{{end}}
