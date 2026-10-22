{{- define "CsUrlValuesReadonlyFile"}}{{template "CsUrlValuesFileHeader" .}}
{{- with .Msg}}
{{- $goName := .GoName}}
// ============================================================================
// Readonly{{$goName}} — FromURLValues deserialization (URL query-string format)
// ============================================================================

public partial struct Readonly{{$goName}}
{
    public Error FromURLValues(ReadOnlySpan<byte> binary)
    {
        // Step 1: decode the URL query string into a collection using the standard library
        var _q = HttpUtility.ParseQueryString(Encoding.UTF8.GetString(binary));

        // Step 2: dispatch each key to its struct member
        foreach (string? _k in _q.Keys)
        {
            if (_k == null) continue;
            var _vals = _q.GetValues(_k);
            if (_vals == null || _vals.Length == 0) continue;
            switch (_k)
            {
{{range .Fields}}
                case "{{.JsonName}}":
{{- if .IsMap}}
                {
                    var _dict{{.Name}} = new {{.EffReadonlyType}}();
                    foreach (var _part in _vals[0].Split('&', StringSplitOptions.RemoveEmptyEntries))
                    {
                        var _meq = _part.IndexOf('=');
                        var _mk = _meq < 0 ? _part : _part.Substring(0, _meq);
                        var _mv = _meq < 0 ? "" : _part.Substring(_meq + 1);
                        {{csUrlKeyParse .MapKeyCS .JsonName}}
{{- if .UseMapValWrapper}}
                        var _mval{{.Name}} = new {{.ReadonlyMapValCS}}();
                        var _mvalErr{{.Name}} = _mval{{.Name}}.FromJSON(Encoding.UTF8.GetBytes(_mv));
                        if (_mvalErr{{.Name}}.Err()) return _mvalErr{{.Name}};
                        _dict{{.Name}}[_key] = new {{.WrapReadonlyMapValCS}} { Value = _mval{{.Name}} };
{{- else if eq .MapValCS "string"}}
                        _dict{{.Name}}[_key] = _mv;
{{- else if eq .MapValCS "byte[]"}}
                        _dict{{.Name}}[_key] = Convert.FromBase64String(_mv);
{{- else if eq .MapValCS "bool"}}
                        _dict{{.Name}}[_key] = _mv == "true" || _mv == "1";
{{- else if .MapValIsMsg}}
                        var _mval{{.Name}} = new {{.ReadonlyMapValCS}}();
                        var _mvalErr{{.Name}} = _mval{{.Name}}.FromJSON(Encoding.UTF8.GetBytes(_mv));
                        if (_mvalErr{{.Name}}.Err()) return _mvalErr{{.Name}};
                        _dict{{.Name}}[_key] = _mval{{.Name}};
{{- else if .MapValIsEnum}}
                        if (!int.TryParse(_mv, out int _mvi{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _dict{{.Name}}[_key] = ({{.ReadonlyMapValCS}})_mvi{{.Name}};
{{- else}}
                        if (!{{.MapValCS}}.TryParse(_mv, out {{.MapValCS}} _mvi{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _dict{{.Name}}[_key] = _mvi{{.Name}};
{{- end}}
                    }
                    this.{{.Name}} = _dict{{.Name}};
                    break;
                }
{{- else if .IsRepeated}}
                {
                    var _list{{.Name}} = new {{.LocalType}}(_vals.Length);
                    foreach (var _sv in _vals)
                    {
{{- if .ElemIsMsg}}
                        var _sub{{.Name}} = new {{.ReadonlyElemTypeCS}}();
                        var _err{{.Name}} = _sub{{.Name}}.FromURLValues(Encoding.UTF8.GetBytes(_sv));
                        if (_err{{.Name}}.Err()) return _err{{.Name}};
                        _list{{.Name}}.Add(_sub{{.Name}});
{{- else if .IsBytes}}
                        _list{{.Name}}.Add(Convert.FromBase64String(_sv));
{{- else if .IsBool}}
                        _list{{.Name}}.Add(_sv == "true" || _sv == "1");
{{- else if .IsString}}
                        _list{{.Name}}.Add(_sv);
{{- else if .IsEnum}}
                        if (!int.TryParse(_sv, out int _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(({{.ReadonlyElemTypeCS}})_el{{.Name}});
{{- else if eq .ElemTypeCS "double"}}
                        if (!double.TryParse(_sv, out double _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- else if eq .ElemTypeCS "float"}}
                        if (!float.TryParse(_sv, out float _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- else if eq .ElemTypeCS "long"}}
                        if (!long.TryParse(_sv, out long _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- else if eq .ElemTypeCS "ulong"}}
                        if (!ulong.TryParse(_sv, out ulong _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- else if eq .ElemTypeCS "uint"}}
                        if (!uint.TryParse(_sv, out uint _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- else}}
                        if (!int.TryParse(_sv, out int _el{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                        _list{{.Name}}.Add(_el{{.Name}});
{{- end}}
                    }
                    this.{{.Name}} = _list{{.Name}};
                    break;
                }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
                {
                    var _subInner{{.Name}} = new {{.ReadonlyType}}();
                    var _err{{.Name}} = _subInner{{.Name}}.FromURLValues(Encoding.UTF8.GetBytes(_vals[0]));
                    if (_err{{.Name}}.Err()) return _err{{.Name}};
                    this.{{.Name}} = new {{.EffReadonlyType}} { Value = _subInner{{.Name}} };
                    break;
                }
{{- else}}
                {
                    var _sub{{.Name}} = new {{.ReadonlyType}}();
                    var _err{{.Name}} = _sub{{.Name}}.FromURLValues(Encoding.UTF8.GetBytes(_vals[0]));
                    if (_err{{.Name}}.Err()) return _err{{.Name}};
                    this.{{.Name}} = _sub{{.Name}};
                    break;
                }
{{- end}}
{{- else if .IsBytes}}
                    this.{{.Name}} = Convert.FromBase64String(_vals[0]);
                    break;
{{- else if .IsBool}}
                    this.{{.Name}} = _vals[0] == "true" || _vals[0] == "1";
                    break;
{{- else if .IsString}}
                    this.{{.Name}} = _vals[0];
                    break;
{{- else if .IsEnum}}
                    if (!int.TryParse(_vals[0], out int _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = ({{.ReadonlyType}})_p{{.Name}};
                    break;
{{- else if .IsDecimal}}
                    if (!double.TryParse(_vals[0], out double _pd{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = (decimal)_pd{{.Name}};
                    break;
{{- else if eq .WriterType "double"}}
                    if (!double.TryParse(_vals[0], out double _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- else if eq .WriterType "float"}}
                    if (!float.TryParse(_vals[0], out float _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- else if eq .WriterType "long"}}
                    if (!long.TryParse(_vals[0], out long _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- else if eq .WriterType "ulong"}}
                    if (!ulong.TryParse(_vals[0], out ulong _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- else if eq .WriterType "uint"}}
                    if (!uint.TryParse(_vals[0], out uint _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- else}}
                    if (!int.TryParse(_vals[0], out int _p{{.Name}})) return global::QiWa.Common.Error.WithLoc(1, "bad value {{.JsonName}}");
                    this.{{.Name}} = _p{{.Name}};
                    break;
{{- end}}
{{end}}
            }
        }
        return default;
    }
}
{{end}}{{end}}

{{with csModel .}}{{template "CsUrlValuesReadonlyFile" .}}{{end}}
