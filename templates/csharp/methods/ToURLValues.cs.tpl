{{- define "CsUrlValuesWriterFile"}}{{template "CsUrlValuesFileHeader" .}}
{{- with .Msg}}
{{- $goName := .GoName}}
// ============================================================================
// {{$goName}} — ToURLValues serialization (URL query-string format)
// ============================================================================

public partial struct {{$goName}}
{
    public readonly Error ToURLValues(ref RentedBuffer buf)
    {
        bool _first = true;
{{range .Fields}}
{{- if .IsMap}}
        if (this.{{.Name}} != null && this.{{.Name}}.Count > 0)
        {
            var _mapBuf{{.Name}} = new RentedBuffer(256);
            bool _mapFirst{{.Name}} = true;
            foreach (var _kv in this.{{.Name}})
            {
                if (!_mapFirst{{.Name}}) _mapBuf{{.Name}}.Append((byte)'&'); _mapFirst{{.Name}} = false;
{{- if eq .MapKeyCS "string"}}
                _mapBuf{{.Name}}.Append(_kv.Key ?? "");
{{- else if eq .MapKeyCS "bool"}}
                _mapBuf{{.Name}}.Append(_kv.Key ? "true"u8 : "false"u8);
{{- else if eq .MapKeyCS "long"}}
                _mapBuf{{.Name}}.Append(_kv.Key);
{{- else if eq .MapKeyCS "ulong"}}
                _mapBuf{{.Name}}.Append(_kv.Key);
{{- else}}
                _mapBuf{{.Name}}.Append((long)_kv.Key);
{{- end}}
                _mapBuf{{.Name}}.Append((byte)'=');
{{- if .UseMapValWrapper}}
                _kv.Value.Value.ToJSON(ref _mapBuf{{.Name}});
{{- else if .MapValIsMsg}}
                _kv.Value.ToJSON(ref _mapBuf{{.Name}});
{{- else if eq .MapValCS "string"}}
                _mapBuf{{.Name}}.Append(_kv.Value ?? "");
{{- else if eq .MapValCS "byte[]"}}
                _mapBuf{{.Name}}.Append(Convert.ToBase64String(_kv.Value));
{{- else if eq .MapValCS "bool"}}
                _mapBuf{{.Name}}.Append(_kv.Value);
{{- else if eq .MapValCS "double"}}
                _mapBuf{{.Name}}.Append(_kv.Value);
{{- else if eq .MapValCS "float"}}
                _mapBuf{{.Name}}.Append((double)_kv.Value);
{{- else if eq .MapValCS "long"}}
                _mapBuf{{.Name}}.Append(_kv.Value);
{{- else if eq .MapValCS "ulong"}}
                _mapBuf{{.Name}}.Append(_kv.Value);
{{- else if .MapValIsEnum}}
                _mapBuf{{.Name}}.Append((long)(int)_kv.Value);
{{- else}}
                _mapBuf{{.Name}}.Append((long)_kv.Value);
{{- end}}
            }
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(Uri.EscapeDataString(Encoding.UTF8.GetString(_mapBuf{{.Name}}.AsSpan())));
        }
{{- else if .IsRepeated}}
        if (this.{{.Name}} != null && this.{{.Name}}.Count > 0)
        {
{{- if .ElemIsMsg}}
            foreach (var _item{{.Name}} in this.{{.Name}})
            {
                var _subBuf{{.Name}} = new RentedBuffer(256);
                var _subErr{{.Name}} = _item{{.Name}}.ToURLValues(ref _subBuf{{.Name}});
                if (_subErr{{.Name}}.Err()) return _subErr{{.Name}};
                if (!_first) buf.Append((byte)'&'); _first = false;
                buf.Append({{$goName}}Tags.JsonKey{{.Name}});
                buf.Append((byte)'=');
                buf.Append(Uri.EscapeDataString(Encoding.UTF8.GetString(_subBuf{{.Name}}.AsSpan())));
            }
{{- else}}
            foreach (var _item{{.Name}} in this.{{.Name}})
            {
                if (!_first) buf.Append((byte)'&'); _first = false;
                buf.Append({{$goName}}Tags.JsonKey{{.Name}});
                buf.Append((byte)'=');
{{- if .IsString}}
                buf.Append(Uri.EscapeDataString(_item{{.Name}} ?? ""));
{{- else if .IsBytes}}
                buf.Append(Convert.ToBase64String(_item{{.Name}}));
{{- else if .IsBool}}
                buf.Append(_item{{.Name}});
{{- else if .IsEnum}}
                buf.Append((long)(int)_item{{.Name}});
{{- else if eq .ElemTypeCS "double"}}
                buf.Append(_item{{.Name}});
{{- else if eq .ElemTypeCS "float"}}
                buf.Append((double)_item{{.Name}});
{{- else if eq .ElemTypeCS "long"}}
                buf.Append(_item{{.Name}});
{{- else if eq .ElemTypeCS "ulong"}}
                buf.Append(_item{{.Name}});
{{- else}}
                buf.Append((long)_item{{.Name}});
{{- end}}
            }
{{- end}}
        }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if (this.{{.Name}} != null)
        {
            var _subBuf{{.Name}} = new RentedBuffer(256);
            var _subErr{{.Name}} = this.{{.Name}}.Value.ToURLValues(ref _subBuf{{.Name}});
            if (_subErr{{.Name}}.Err()) return _subErr{{.Name}};
            if (_subBuf{{.Name}}.AsSpan().Length > 0)
            {
                if (!_first) buf.Append((byte)'&'); _first = false;
                buf.Append({{$goName}}Tags.JsonKey{{.Name}});
                buf.Append((byte)'=');
                buf.Append(Uri.EscapeDataString(Encoding.UTF8.GetString(_subBuf{{.Name}}.AsSpan())));
            }
        }
{{- else}}
        {
            var _subBuf{{.Name}} = new RentedBuffer(256);
            var _subErr{{.Name}} = this.{{.Name}}.ToURLValues(ref _subBuf{{.Name}});
            if (_subErr{{.Name}}.Err()) return _subErr{{.Name}};
            if (_subBuf{{.Name}}.AsSpan().Length > 0)
            {
                if (!_first) buf.Append((byte)'&'); _first = false;
                buf.Append({{$goName}}Tags.JsonKey{{.Name}});
                buf.Append((byte)'=');
                buf.Append(Uri.EscapeDataString(Encoding.UTF8.GetString(_subBuf{{.Name}}.AsSpan())));
            }
        }
{{- end}}
{{- else if .IsBytes}}
        if (this.{{.Name}} != null && this.{{.Name}}.Length > 0)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(Convert.ToBase64String(this.{{.Name}}));
        }
{{- else if .IsBool}}
        if (this.{{.Name}})
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append("=true"u8);
        }
{{- else if .IsString}}
        if (!string.IsNullOrEmpty(this.{{.Name}}))
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(Uri.EscapeDataString(this.{{.Name}}));
        }
{{- else if .IsEnum}}
        if ((int)this.{{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append((long)(int)this.{{.Name}});
        }
{{- else if .IsDecimal}}
        if (this.{{.Name}} != 0m)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append((double)Math.Round(this.{{.Name}}, {{.DecimalRound}}, MidpointRounding.AwayFromZero));
        }
{{- else if eq .WriterType "double"}}
        if (this.{{.Name}} != 0.0)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(this.{{.Name}});
        }
{{- else if eq .WriterType "float"}}
        if (this.{{.Name}} != 0.0f)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append((double)this.{{.Name}});
        }
{{- else if .IsFixed64}}
        if (this.{{.Name}} != {{csDefault .WriterType}})
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(this.{{.Name}});
        }
{{- else if eq .WriterType "long"}}
        if (this.{{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(this.{{.Name}});
        }
{{- else if eq .WriterType "ulong"}}
        if (this.{{.Name}} != 0)
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append(this.{{.Name}});
        }
{{- else}}
        if (this.{{.Name}} != {{csDefault .WriterType}})
        {
            if (!_first) buf.Append((byte)'&'); _first = false;
            buf.Append({{$goName}}Tags.JsonKey{{.Name}});
            buf.Append((byte)'=');
            buf.Append((long)this.{{.Name}});
        }
{{- end}}
{{end}}
        return default;
    }
}
{{end}}{{end}}
{{with csModel .}}{{template "CsUrlValuesWriterFile" .}}{{end}}
