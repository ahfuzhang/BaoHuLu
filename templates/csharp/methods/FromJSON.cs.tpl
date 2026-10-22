{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct Readonly{{$goName}}
{
    // ── FromJSON ─────────────────────────────────────────────────────────────

    public Error FromJSON(ReadOnlySpan<byte> text)
    {
        try
        {
            var _reader = new Utf8JsonReader(text, new JsonReaderOptions
            {
                AllowTrailingCommas = true,
                CommentHandling     = JsonCommentHandling.Skip
            });
            return FromJSONReader(ref _reader);
        }
        catch (System.Text.Json.JsonException ex)
        {
            return global::QiWa.Common.Error.WithLoc(1, ex.Message);
        }
    }

    internal Error FromJSONReader(ref Utf8JsonReader reader)
    {
{{- if .AsArray}}
{{- $f := index .Fields 0}}
        // @AsArray: the JSON value IS the array — read elements directly into the single repeated field.
        // A JSON null value (e.g. a null field) means "no data": leave defaults, consume nothing further.
        if (reader.TokenType == JsonTokenType.Null)
            return default;
        // If the reader is freshly created (None) or not yet at StartArray, advance.
        // When called from a nested context, caller already advanced to StartArray.
        if (reader.TokenType != JsonTokenType.StartArray)
        {
            if (!reader.Read() || reader.TokenType != JsonTokenType.StartArray)
                return global::QiWa.Common.Error.WithLoc(1, "expected [");
        }
        var _{{$f.Name}}List = this.{{$f.Name}} ?? new {{$f.LocalType}}();
        _{{$f.Name}}List.Clear();
        while (reader.Read() && reader.TokenType != JsonTokenType.EndArray)
        {
{{- if $f.ElemIsMsg}}
            {{$f.ReadonlyElemTypeCS}} _arrMsg{{$f.Name}} = default;
            var _arrErr{{$f.Name}} = _arrMsg{{$f.Name}}.FromJSONReader(ref reader);
            if (_arrErr{{$f.Name}}.Err()) return _arrErr{{$f.Name}};
            _{{$f.Name}}List.Add(_arrMsg{{$f.Name}});
{{- else if $f.IsString}}
            _{{$f.Name}}List.Add(reader.GetString() ?? string.Empty);
{{- else if $f.IsBytes}}
            _{{$f.Name}}List.Add(reader.GetBytesFromBase64());
{{- else if $f.IsBool}}
            _{{$f.Name}}List.Add(reader.ValueSpan.SequenceEqual("true"u8));
{{- else if eq $f.ElemTypeCS "double"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _arrdv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(_arrdv{{$f.Name}}); }
{{- else if eq $f.ElemTypeCS "float"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out float _arrfv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(_arrfv{{$f.Name}}); }
{{- else if $f.IsFixed64}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _arrfl{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(({{$f.ElemTypeCS}})_arrfl{{$f.Name}}); }
{{- else if eq $f.ElemTypeCS "long"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _arrl{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(_arrl{{$f.Name}}); }
{{- else if eq $f.ElemTypeCS "ulong"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _arrul{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(_arrul{{$f.Name}}); }
{{- else}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _arriv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{$f.Name}}"); _{{$f.Name}}List.Add(({{$f.ElemTypeCS}})_arriv{{$f.Name}}); }
{{- end}}
        }
        this.{{$f.Name}} = _{{$f.Name}}List;
        return default;
    }
{{- else}}
        // A JSON null value (e.g. a null message field) means "no data": leave defaults, consume nothing further.
        if (reader.TokenType == JsonTokenType.Null)
            return default;
        // If the reader is freshly created (None) or not yet at StartObject, advance.
        // When called from a nested context, caller already advanced to StartObject.
        if (reader.TokenType != JsonTokenType.StartObject)
        {
            if (!reader.Read() || reader.TokenType != JsonTokenType.StartObject)
                return global::QiWa.Common.Error.WithLoc(1, "expected {");
        }
{{- if .AsMap}}
{{- $f := index .Fields 0}}
        // @AsMap: the JSON object IS the map — iterate keys directly.
        var _{{$f.Name}}Dict = this.{{$f.Name}} ?? new {{$f.EffLocalType}}();
        while (reader.Read() && reader.TokenType != JsonTokenType.EndObject)
        {
            if (reader.TokenType != JsonTokenType.PropertyName)
                return global::QiWa.Common.Error.WithLoc(1, "expected property name");
{{- if eq $f.MapKeyCS "string"}}
            string _mke{{$f.Name}} = reader.GetString()!;
{{- else if eq $f.MapKeyCS "bool"}}
            bool _mke{{$f.Name}} = reader.ValueSpan.SequenceEqual("true"u8);
{{- else}}
            if (!Utf8Parser.TryParse(reader.ValueSpan, out {{$f.MapKeyCS}} _mke{{$f.Name}}, out _))
                return global::QiWa.Common.Error.WithLoc(1, "bad map key {{$f.Name}}");
{{- end}}
            reader.Read();
{{- if $f.MapValIsMsg}}
            {{$f.ReadonlyMapValCS}} _mvInner{{$f.Name}} = default;
            var _mvErr{{$f.Name}} = _mvInner{{$f.Name}}.FromJSONReader(ref reader);
            if (_mvErr{{$f.Name}}.Err()) return _mvErr{{$f.Name}};
{{- if $f.UseMapValWrapper}}
            _{{$f.Name}}Dict[_mke{{$f.Name}}] = new {{$f.WrapReadonlyMapValCS}}() { Value = _mvInner{{$f.Name}} };
{{- else}}
            _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvInner{{$f.Name}};
{{- end}}
{{- else if eq $f.MapVal "string"}}
            _{{$f.Name}}Dict[_mke{{$f.Name}}] = reader.GetString() ?? string.Empty;
{{- else if eq $f.MapVal "bytes"}}
            _{{$f.Name}}Dict[_mke{{$f.Name}}] = reader.GetBytesFromBase64();
{{- else if eq $f.MapVal "bool"}}
            _{{$f.Name}}Dict[_mke{{$f.Name}}] = reader.ValueSpan.SequenceEqual("true"u8);
{{- else if eq $f.MapVal "double"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _mvdv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvdv{{$f.Name}}; }
{{- else if eq $f.MapVal "float"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out float _mvfv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvfv{{$f.Name}}; }
{{- else if eq $f.MapVal "int64"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvl{{$f.Name}}; }
{{- else if eq $f.MapVal "uint64"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _mvul{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvul{{$f.Name}}; }
{{- else if eq $f.MapVal "sint64"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvl{{$f.Name}}; }
{{- else if eq $f.MapVal "sfixed64"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvl{{$f.Name}}; }
{{- else if eq $f.MapVal "fixed64"}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _mvul{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = _mvul{{$f.Name}}; }
{{- else}}
            { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _mviv{{$f.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{$f.Name}}"); _{{$f.Name}}Dict[_mke{{$f.Name}}] = ({{$f.MapValCS}})_mviv{{$f.Name}}; }
{{- end}}
        }
        this.{{$f.Name}} = _{{$f.Name}}Dict;
        return default;
    }
{{- else}}

        // local decode variables
{{- range .Fields}}
{{- if .IsMap}}
        var _{{.Name}}Dict = this.{{.Name}} ?? new {{.EffLocalType}}();
{{- else if .IsRepeated}}
        var _{{.Name}}List = this.{{.Name}} ?? new {{.LocalType}}();
        _{{.Name}}List.Clear();
{{- end}}
{{- end}}

        while (reader.Read() && reader.TokenType != JsonTokenType.EndObject)
        {
            if (reader.TokenType != JsonTokenType.PropertyName)
                return global::QiWa.Common.Error.WithLoc(1, "expected property name");
            var _propSpan = reader.ValueSpan;
            if (!reader.Read()) return global::QiWa.Common.Error.WithLoc(1, "unexpected end");
            // JSON null 视为字段未设置：跳过任意类型字段（标量/消息/map/数组）的 null 值。
            if (reader.TokenType == JsonTokenType.Null) { continue; }
{{- range $i, $f := .Fields}}
            {{if $i}}else {{end}}if (_propSpan.SequenceEqual({{$goName}}Tags.JsonKey{{.Name}}))
{{- if .IsMap}}
                {
                    if (reader.TokenType == JsonTokenType.StartObject)
                    {
                        while (reader.Read() && reader.TokenType != JsonTokenType.EndObject)
                        {
{{- if eq .MapKeyCS "string"}}
                            string _mke{{.Name}} = reader.GetString()!;
{{- else if eq .MapKeyCS "bool"}}
                            bool _mke{{.Name}} = reader.ValueSpan.SequenceEqual("true"u8);
{{- else}}
                            if (!Utf8Parser.TryParse(reader.ValueSpan, out {{.MapKeyCS}} _mke{{.Name}}, out _))
                                return global::QiWa.Common.Error.WithLoc(1, "bad map key {{.Name}}");
{{- end}}
                            reader.Read();
{{- if .MapValIsMsg}}
                            {{.ReadonlyMapValCS}} _mvInner{{.Name}} = default;
                            var _mvErr{{.Name}} = _mvInner{{.Name}}.FromJSONReader(ref reader);
                            if (_mvErr{{.Name}}.Err()) return _mvErr{{.Name}};
{{- if .UseMapValWrapper}}
                            _{{.Name}}Dict[_mke{{.Name}}] = new {{.WrapReadonlyMapValCS}}() { Value = _mvInner{{.Name}} };
{{- else}}
                            _{{.Name}}Dict[_mke{{.Name}}] = _mvInner{{.Name}};
{{- end}}
{{- else if eq .MapVal "string"}}
                            _{{.Name}}Dict[_mke{{.Name}}] = reader.GetString() ?? string.Empty;
{{- else if eq .MapVal "bytes"}}
                            _{{.Name}}Dict[_mke{{.Name}}] = reader.GetBytesFromBase64();
{{- else if eq .MapVal "bool"}}
                            _{{.Name}}Dict[_mke{{.Name}}] = reader.ValueSpan.SequenceEqual("true"u8);
{{- else if eq .MapVal "double"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _mvdv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvdv{{.Name}}; }
{{- else if eq .MapVal "float"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out float _mvfv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvfv{{.Name}}; }
{{- else if eq .MapVal "int64"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvl{{.Name}}; }
{{- else if eq .MapVal "uint64"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _mvul{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvul{{.Name}}; }
{{- else if eq .MapVal "sint64"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvl{{.Name}}; }
{{- else if eq .MapVal "sfixed64"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _mvl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvl{{.Name}}; }
{{- else if eq .MapVal "fixed64"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _mvul{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = _mvul{{.Name}}; }
{{- else}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _mviv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad map value {{.Name}}"); _{{.Name}}Dict[_mke{{.Name}}] = ({{.MapValCS}})_mviv{{.Name}}; }
{{- end}}
                        }
                    }
                }
{{- else if .IsRepeated}}
                {
                    _{{.Name}}List ??= new {{.LocalType}}();
                    if (reader.TokenType == JsonTokenType.StartArray)
                    {
                        while (reader.Read() && reader.TokenType != JsonTokenType.EndArray)
                        {
{{- if .ElemIsMsg}}
                            {{.ReadonlyElemTypeCS}} _arrMsg{{.Name}} = default;
                            var _arrErr{{.Name}} = _arrMsg{{.Name}}.FromJSONReader(ref reader);
                            if (_arrErr{{.Name}}.Err()) return _arrErr{{.Name}};
                            _{{.Name}}List.Add(_arrMsg{{.Name}});
{{- else if .IsString}}
                            _{{.Name}}List.Add(reader.GetString() ?? string.Empty);
{{- else if .IsBytes}}
                            _{{.Name}}List.Add(reader.GetBytesFromBase64());
{{- else if .IsBool}}
                            _{{.Name}}List.Add(reader.ValueSpan.SequenceEqual("true"u8));
{{- else if eq .ElemTypeCS "double"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _arrdv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(_arrdv{{.Name}}); }
{{- else if eq .ElemTypeCS "float"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out float _arrfv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(_arrfv{{.Name}}); }
{{- else if .IsFixed64}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _arrfl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(({{.ElemTypeCS}})_arrfl{{.Name}}); }
{{- else if eq .ElemTypeCS "long"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _arrl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(_arrl{{.Name}}); }
{{- else if eq .ElemTypeCS "ulong"}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _arrul{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(_arrul{{.Name}}); }
{{- else}}
                            { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _arriv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad list elem {{.Name}}"); _{{.Name}}List.Add(({{.ElemTypeCS}})_arriv{{.Name}}); }
{{- end}}
                        }
                    }
                }
{{- else if .IsMsg}}
                {
{{- if .UseDirectWrapper}}
                    if (this.{{.Name}} == null) this.{{.Name}} = new {{.EffReadonlyType}}();
                    var _jerr{{.Name}} = this.{{.Name}}.Value.FromJSONReader(ref reader);
{{- else}}
                    var _jerr{{.Name}} = this.{{.Name}}.FromJSONReader(ref reader);
{{- end}}
                    if (_jerr{{.Name}}.Err()) return _jerr{{.Name}};
                }
{{- else if .IsString}}
                    this.{{.Name}} = reader.GetString() ?? string.Empty;
{{- else if .IsBytes}}
                    this.{{.Name}} = reader.GetBytesFromBase64();
{{- else if .IsBool}}
                    this.{{.Name}} = reader.ValueSpan.SequenceEqual("true"u8);
{{- else if .IsEnum}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _ev{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = ({{.WriterType}})_ev{{.Name}}; }
{{- else if .IsDecimal}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _dv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = Math.Round((decimal)_dv{{.Name}}, {{.DecimalRound}}, MidpointRounding.AwayFromZero); }
{{- else if eq .WriterType "double"}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out double _dv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = _dv{{.Name}}; }
{{- else if eq .WriterType "float"}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out float _fv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = _fv{{.Name}}; }
{{- else if .IsFixed64}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _fl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = ({{.WriterType}})_fl{{.Name}}; }
{{- else if eq .WriterType "long"}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out long _fl{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = _fl{{.Name}}; }
{{- else if eq .WriterType "ulong"}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out ulong _ful{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = _ful{{.Name}}; }
{{- else}}
                    { if (!Utf8Parser.TryParse(reader.ValueSpan, out int _iv{{.Name}}, out _)) return global::QiWa.Common.Error.WithLoc(1, "bad {{.Name}}"); this.{{.Name}} = ({{.WriterType}})_iv{{.Name}}; }
{{- end}}
{{- end}}
            else { reader.Skip(); }
        }
{{- range .Fields}}
{{- if .IsMap}}
        this.{{.Name}} = _{{.Name}}Dict;
{{- else if .IsRepeated}}
        this.{{.Name}} = _{{.Name}}List;
{{- end}}
{{- end}}
        return default;
    }
{{- end}}
{{- end}}
}
{{end}}{{end}}
