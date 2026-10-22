{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct Readonly{{$goName}}
{
    // ── FromProtobuf ─────────────────────────────────────────────────────────

    public Error FromProtobuf(ReadOnlySpan<byte> binary)
    {
        // local decode variables
{{- range .Fields}}
{{- if .IsMap}}
        var _{{.Name}}Dict = this.{{.Name}} ?? new {{.EffLocalType}}();
{{- else if .IsRepeated}}
        var _{{.Name}}List = this.{{.Name}} ?? new {{.LocalType}}();
        _{{.Name}}List.Clear();
{{- end}}
{{- end}}

        int _pos = 0;
        while (_pos < binary.Length)
        {
            if (!TryReadVarint(binary, _pos, out ulong _tagRaw, out int _tagBytes))
                return global::QiWa.Common.Error.WithLoc(1, "failed to read tag");
            _pos += _tagBytes;
            int _fieldNum = (int)(_tagRaw >> 3);
            int _wireType = (int)(_tagRaw & 7);
            switch (_fieldNum)
            {
{{- range .Fields}}
                case {{$goName}}Tags.Tag{{.Name}}: // {{.Name}}
{{- if .IsMap}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _elen{{.Name}}, out int _elb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad map {{.Name}}");
                        _pos += _elb{{.Name}};
                        var _entBin{{.Name}} = binary.Slice(_pos, (int)_elen{{.Name}});
{{- if eq .MapKeyCS "string"}}
                        {{.MapKeyCS}} _entKey{{.Name}} = string.Empty;
{{- else}}
                        {{.MapKeyCS}} _entKey{{.Name}} = default;
{{- end}}
{{- if .UseMapValWrapper}}
                        {{.WrapReadonlyMapValCS}} _entVal{{.Name}} = null;
{{- else if eq .ReadonlyMapValCS "string"}}
                        {{.ReadonlyMapValCS}} _entVal{{.Name}} = string.Empty;
{{- else}}
                        {{.ReadonlyMapValCS}} _entVal{{.Name}} = default;
{{- end}}
                        int _ep{{.Name}} = 0;
                        while (_ep{{.Name}} < _entBin{{.Name}}.Length)
                        {
                            if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _etag{{.Name}}, out int _etb{{.Name}})) break;
                            _ep{{.Name}} += _etb{{.Name}};
                            int _ef{{.Name}} = (int)(_etag{{.Name}} >> 3);
                            int _ew{{.Name}} = (int)(_etag{{.Name}} & 7);
                            if (_ef{{.Name}} == 1) // map key
                            {
{{- if eq .MapKey "string"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _ksl{{.Name}}, out int _ksb{{.Name}})) break;
                                _ep{{.Name}} += _ksb{{.Name}};
                                _entKey{{.Name}} = Encoding.UTF8.GetString(_entBin{{.Name}}.Slice(_ep{{.Name}}, (int)_ksl{{.Name}}));
                                _ep{{.Name}} += (int)_ksl{{.Name}};
{{- else if eq .MapKey "bool"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _kbv{{.Name}}, out int _kbb{{.Name}})) break;
                                _entKey{{.Name}} = _kbv{{.Name}} != 0;
                                _ep{{.Name}} += _kbb{{.Name}};
{{- else if eq .MapKey "fixed32"}}
                                TryReadFixed32(_entBin{{.Name}}, _ep{{.Name}}, out uint _kf32{{.Name}});
                                _entKey{{.Name}} = _kf32{{.Name}}; _ep{{.Name}} += 4;
{{- else if eq .MapKey "sfixed32"}}
                                TryReadFixed32(_entBin{{.Name}}, _ep{{.Name}}, out uint _ksf32{{.Name}});
                                _entKey{{.Name}} = (int)_ksf32{{.Name}}; _ep{{.Name}} += 4;
{{- else if eq .MapKey "fixed64"}}
                                TryReadFixed64(_entBin{{.Name}}, _ep{{.Name}}, out ulong _kf64{{.Name}});
                                _entKey{{.Name}} = _kf64{{.Name}}; _ep{{.Name}} += 8;
{{- else if eq .MapKey "sfixed64"}}
                                TryReadFixed64(_entBin{{.Name}}, _ep{{.Name}}, out ulong _ksf64{{.Name}});
                                _entKey{{.Name}} = (long)_ksf64{{.Name}}; _ep{{.Name}} += 8;
{{- else if eq .MapKey "sint32"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _ks32{{.Name}}, out int _kb32{{.Name}})) break;
                                _entKey{{.Name}} = ZigZagDecode32(_ks32{{.Name}}); _ep{{.Name}} += _kb32{{.Name}};
{{- else if eq .MapKey "sint64"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _ks64{{.Name}}, out int _kb64{{.Name}})) break;
                                _entKey{{.Name}} = ZigZagDecode64(_ks64{{.Name}}); _ep{{.Name}} += _kb64{{.Name}};
{{- else}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _kv{{.Name}}, out int _kvb{{.Name}})) break;
                                _entKey{{.Name}} = ({{.MapKeyCS}})_kv{{.Name}}; _ep{{.Name}} += _kvb{{.Name}};
{{- end}}
                            }
                            else if (_ef{{.Name}} == 2) // map value
                            {
{{- if .MapValIsMsg}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vmlen{{.Name}}, out int _vmlb{{.Name}})) break;
                                _ep{{.Name}} += _vmlb{{.Name}};
                                {{.ReadonlyMapValCS}} _vsubInner{{.Name}} = default;
                                var _vsubErr{{.Name}} = _vsubInner{{.Name}}.FromProtobuf(_entBin{{.Name}}.Slice(_ep{{.Name}}, (int)_vmlen{{.Name}}));
                                if (_vsubErr{{.Name}}.Err()) return _vsubErr{{.Name}};
{{- if .UseMapValWrapper}}
                                _entVal{{.Name}} = new {{.WrapReadonlyMapValCS}}() { Value = _vsubInner{{.Name}} }; _ep{{.Name}} += (int)_vmlen{{.Name}};
{{- else}}
                                _entVal{{.Name}} = _vsubInner{{.Name}}; _ep{{.Name}} += (int)_vmlen{{.Name}};
{{- end}}
{{- else if eq .MapVal "string"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vsl{{.Name}}, out int _vsb{{.Name}})) break;
                                _ep{{.Name}} += _vsb{{.Name}};
                                _entVal{{.Name}} = Encoding.UTF8.GetString(_entBin{{.Name}}.Slice(_ep{{.Name}}, (int)_vsl{{.Name}}));
                                _ep{{.Name}} += (int)_vsl{{.Name}};
{{- else if eq .MapVal "bytes"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vbl{{.Name}}, out int _vbb{{.Name}})) break;
                                _ep{{.Name}} += _vbb{{.Name}};
                                _entVal{{.Name}} = _entBin{{.Name}}.Slice(_ep{{.Name}}, (int)_vbl{{.Name}}).ToArray(); _ep{{.Name}} += (int)_vbl{{.Name}};
{{- else if eq .MapVal "double"}}
                                TryReadFixed64(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vd64{{.Name}});
                                _entVal{{.Name}} = BitConverter.UInt64BitsToDouble(_vd64{{.Name}}); _ep{{.Name}} += 8;
{{- else if eq .MapVal "float"}}
                                TryReadFixed32(_entBin{{.Name}}, _ep{{.Name}}, out uint _vf32{{.Name}});
                                _entVal{{.Name}} = BitConverter.UInt32BitsToSingle(_vf32{{.Name}}); _ep{{.Name}} += 4;
{{- else if eq .MapVal "fixed32"}}
                                TryReadFixed32(_entBin{{.Name}}, _ep{{.Name}}, out uint _vfx32{{.Name}});
                                _entVal{{.Name}} = _vfx32{{.Name}}; _ep{{.Name}} += 4;
{{- else if eq .MapVal "sfixed32"}}
                                TryReadFixed32(_entBin{{.Name}}, _ep{{.Name}}, out uint _vsfx32{{.Name}});
                                _entVal{{.Name}} = (int)_vsfx32{{.Name}}; _ep{{.Name}} += 4;
{{- else if eq .MapVal "fixed64"}}
                                TryReadFixed64(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vfx64{{.Name}});
                                _entVal{{.Name}} = _vfx64{{.Name}}; _ep{{.Name}} += 8;
{{- else if eq .MapVal "sfixed64"}}
                                TryReadFixed64(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vsfx64{{.Name}});
                                _entVal{{.Name}} = (long)_vsfx64{{.Name}}; _ep{{.Name}} += 8;
{{- else if eq .MapVal "sint32"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vs32{{.Name}}, out int _vsb32{{.Name}})) break;
                                _entVal{{.Name}} = ZigZagDecode32(_vs32{{.Name}}); _ep{{.Name}} += _vsb32{{.Name}};
{{- else if eq .MapVal "sint64"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vs64{{.Name}}, out int _vsb64{{.Name}})) break;
                                _entVal{{.Name}} = ZigZagDecode64(_vs64{{.Name}}); _ep{{.Name}} += _vsb64{{.Name}};
{{- else if eq .MapVal "bool"}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vbool{{.Name}}, out int _vboolb{{.Name}})) break;
                                _entVal{{.Name}} = _vbool{{.Name}} != 0; _ep{{.Name}} += _vboolb{{.Name}};
{{- else}}
                                if (!TryReadVarint(_entBin{{.Name}}, _ep{{.Name}}, out ulong _vvar{{.Name}}, out int _vvarb{{.Name}})) break;
                                _entVal{{.Name}} = ({{.MapValCS}})_vvar{{.Name}}; _ep{{.Name}} += _vvarb{{.Name}};
{{- end}}
                            }
                            else
                            {
                                if (!SkipField(_entBin{{.Name}}, _ep{{.Name}}, _ew{{.Name}}, out int _eskip{{.Name}})) break;
                                _ep{{.Name}} += _eskip{{.Name}};
                            }
                        }
                        _{{.Name}}Dict[_entKey{{.Name}}] = _entVal{{.Name}};
                        _pos += (int)_elen{{.Name}};
                    }
{{- else if .IsRepeated}}
                    {
{{- if .IsPackable}}
                        _{{.Name}}List ??= new {{.LocalType}}();
                        if (_wireType == WireTypeLenDelim)
                        { // packed
                            if (!TryReadVarint(binary, _pos, out ulong _plen{{.Name}}, out int _plb{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad packed {{.Name}}");
                            _pos += _plb{{.Name}};
                            int _packEnd{{.Name}} = _pos + (int)_plen{{.Name}};
                            while (_pos < _packEnd{{.Name}})
                            {
{{- if .IsFixed32}}
                                if (!TryReadFixed32(binary, _pos, out uint _pf32{{.Name}})) break;
{{- if eq .Type "float"}}
                                _{{.Name}}List.Add(BitConverter.UInt32BitsToSingle(_pf32{{.Name}}));
{{- else if eq .Type "sfixed32"}}
                                _{{.Name}}List.Add((int)_pf32{{.Name}});
{{- else}}
                                _{{.Name}}List.Add(_pf32{{.Name}});
{{- end}}
                                _pos += 4;
{{- else if .IsFixed64}}
                                if (!TryReadFixed64(binary, _pos, out ulong _pf64{{.Name}})) break;
{{- if eq .Type "double"}}
                                _{{.Name}}List.Add(BitConverter.UInt64BitsToDouble(_pf64{{.Name}}));
{{- else if eq .Type "sfixed64"}}
                                _{{.Name}}List.Add((long)_pf64{{.Name}});
{{- else}}
                                _{{.Name}}List.Add(_pf64{{.Name}});
{{- end}}
                                _pos += 8;
{{- else if .IsBool}}
                                if (!TryReadVarint(binary, _pos, out ulong _pbool{{.Name}}, out int _pboolb{{.Name}})) break;
                                _{{.Name}}List.Add(_pbool{{.Name}} != 0); _pos += _pboolb{{.Name}};
{{- else if .IsSint32}}
                                if (!TryReadVarint(binary, _pos, out ulong _psz{{.Name}}, out int _pszb{{.Name}})) break;
                                _{{.Name}}List.Add(ZigZagDecode32(_psz{{.Name}})); _pos += _pszb{{.Name}};
{{- else if .IsSint64}}
                                if (!TryReadVarint(binary, _pos, out ulong _psz64{{.Name}}, out int _psz64b{{.Name}})) break;
                                _{{.Name}}List.Add(ZigZagDecode64(_psz64{{.Name}})); _pos += _psz64b{{.Name}};
{{- else}}
                                if (!TryReadVarint(binary, _pos, out ulong _pv{{.Name}}, out int _pvb{{.Name}})) break;
                                _{{.Name}}List.Add(({{.ElemTypeCS}})_pv{{.Name}}); _pos += _pvb{{.Name}};
{{- end}}
                            }
                        }
                        else
                        { // non-packed single element
{{- if .IsFixed32}}
                            if (!TryReadFixed32(binary, _pos, out uint _npf32{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad fixed32 {{.Name}}");
{{- if eq .Type "float"}}
                            _{{.Name}}List.Add(BitConverter.UInt32BitsToSingle(_npf32{{.Name}})); _pos += 4;
{{- else if eq .Type "sfixed32"}}
                            _{{.Name}}List.Add((int)_npf32{{.Name}}); _pos += 4;
{{- else}}
                            _{{.Name}}List.Add(_npf32{{.Name}}); _pos += 4;
{{- end}}
{{- else if .IsFixed64}}
                            if (!TryReadFixed64(binary, _pos, out ulong _npf64{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad fixed64 {{.Name}}");
{{- if eq .Type "double"}}
                            _{{.Name}}List.Add(BitConverter.UInt64BitsToDouble(_npf64{{.Name}})); _pos += 8;
{{- else if eq .Type "sfixed64"}}
                            _{{.Name}}List.Add((long)_npf64{{.Name}}); _pos += 8;
{{- else}}
                            _{{.Name}}List.Add(_npf64{{.Name}}); _pos += 8;
{{- end}}
{{- else if .IsBool}}
                            if (!TryReadVarint(binary, _pos, out ulong _npbool{{.Name}}, out int _npboolb{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad bool {{.Name}}");
                            _{{.Name}}List.Add(_npbool{{.Name}} != 0); _pos += _npboolb{{.Name}};
{{- else if .IsSint32}}
                            if (!TryReadVarint(binary, _pos, out ulong _npsz{{.Name}}, out int _npszb{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad sint32 {{.Name}}");
                            _{{.Name}}List.Add(ZigZagDecode32(_npsz{{.Name}})); _pos += _npszb{{.Name}};
{{- else if .IsSint64}}
                            if (!TryReadVarint(binary, _pos, out ulong _npsz64{{.Name}}, out int _npsz64b{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad sint64 {{.Name}}");
                            _{{.Name}}List.Add(ZigZagDecode64(_npsz64{{.Name}})); _pos += _npsz64b{{.Name}};
{{- else}}
                            if (!TryReadVarint(binary, _pos, out ulong _npv{{.Name}}, out int _npvb{{.Name}}))
                                return global::QiWa.Common.Error.WithLoc(1, "bad varint {{.Name}}");
                            _{{.Name}}List.Add(({{.ElemTypeCS}})_npv{{.Name}}); _pos += _npvb{{.Name}};
{{- end}}
                        }
{{- else}}
                        // repeated non-packable
                        _{{.Name}}List ??= new {{.LocalType}}();
{{- if .IsString}}
                        if (!TryReadVarint(binary, _pos, out ulong _rslen{{.Name}}, out int _rslb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad string {{.Name}}");
                        _pos += _rslb{{.Name}};
                        _{{.Name}}List.Add(Encoding.UTF8.GetString(binary.Slice(_pos, (int)_rslen{{.Name}}))); _pos += (int)_rslen{{.Name}};
{{- else if .IsBytes}}
                        if (!TryReadVarint(binary, _pos, out ulong _rblen{{.Name}}, out int _rblb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad bytes {{.Name}}");
                        _pos += _rblb{{.Name}};
                        _{{.Name}}List.Add(binary.Slice(_pos, (int)_rblen{{.Name}}).ToArray()); _pos += (int)_rblen{{.Name}};
{{- else if .ElemIsMsg}}
                        if (!TryReadVarint(binary, _pos, out ulong _rmlen{{.Name}}, out int _rmlb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad msg {{.Name}}");
                        _pos += _rmlb{{.Name}};
                        {{.ReadonlyElemTypeCS}} _rsub{{.Name}} = default;
                        var _rsuberr{{.Name}} = _rsub{{.Name}}.FromProtobuf(binary.Slice(_pos, (int)_rmlen{{.Name}}));
                        if (_rsuberr{{.Name}}.Err()) return _rsuberr{{.Name}};
                        _{{.Name}}List.Add(_rsub{{.Name}}); _pos += (int)_rmlen{{.Name}};
{{- else}}
                        if (!TryReadVarint(binary, _pos, out ulong _rvar{{.Name}}, out int _rvarb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad elem {{.Name}}");
                        _{{.Name}}List.Add(({{.ElemTypeCS}})_rvar{{.Name}}); _pos += _rvarb{{.Name}};
{{- end}}
{{- end}}
                    }
{{- else if .IsMsg}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _mlen{{.Name}}, out int _mlb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad msg {{.Name}}");
                        _pos += _mlb{{.Name}};
{{- if .UseDirectWrapper}}
                        if (this.{{.Name}} == null) this.{{.Name}} = new {{.EffReadonlyType}}();
                        var _subErr{{.Name}} = this.{{.Name}}.Value.FromProtobuf(binary.Slice(_pos, (int)_mlen{{.Name}}));
{{- else}}
                        var _subErr{{.Name}} = this.{{.Name}}.FromProtobuf(binary.Slice(_pos, (int)_mlen{{.Name}}));
{{- end}}
                        if (_subErr{{.Name}}.Err()) return _subErr{{.Name}};
                        _pos += (int)_mlen{{.Name}};
                    }
{{- else if .IsString}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _slen{{.Name}}, out int _slb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad string {{.Name}}");
                        _pos += _slb{{.Name}};
                        this.{{.Name}} = Encoding.UTF8.GetString(binary.Slice(_pos, (int)_slen{{.Name}}));
                        _pos += (int)_slen{{.Name}};
                    }
{{- else if .IsBytes}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _blen{{.Name}}, out int _blb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad bytes {{.Name}}");
                        _pos += _blb{{.Name}};
                        this.{{.Name}} = binary.Slice(_pos, (int)_blen{{.Name}}).ToArray();
                        _pos += (int)_blen{{.Name}};
                    }
{{- else if .IsFixed32}}
                    {
                        if (!TryReadFixed32(binary, _pos, out uint _f32v{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad fixed32 {{.Name}}");
{{- if eq .Type "float"}}
                        this.{{.Name}} = BitConverter.UInt32BitsToSingle(_f32v{{.Name}});
{{- else if eq .Type "sfixed32"}}
                        this.{{.Name}} = (int)_f32v{{.Name}};
{{- else}}
                        this.{{.Name}} = _f32v{{.Name}};
{{- end}}
                        _pos += 4;
                    }
{{- else if .IsDecimal}}
                    {
                        if (!TryReadFixed64(binary, _pos, out ulong _f64v{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad fixed64 {{.Name}}");
                        this.{{.Name}} = Math.Round((decimal)BitConverter.UInt64BitsToDouble(_f64v{{.Name}}), {{.DecimalRound}}, MidpointRounding.AwayFromZero);
                        _pos += 8;
                    }
{{- else if .IsFixed64}}
                    {
                        if (!TryReadFixed64(binary, _pos, out ulong _f64v{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad fixed64 {{.Name}}");
{{- if eq .Type "double"}}
                        this.{{.Name}} = BitConverter.UInt64BitsToDouble(_f64v{{.Name}});
{{- else if eq .Type "sfixed64"}}
                        this.{{.Name}} = (long)_f64v{{.Name}};
{{- else}}
                        this.{{.Name}} = _f64v{{.Name}};
{{- end}}
                        _pos += 8;
                    }
{{- else if .IsBool}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _boolv{{.Name}}, out int _boolb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad bool {{.Name}}");
                        this.{{.Name}} = _boolv{{.Name}} != 0;
                        _pos += _boolb{{.Name}};
                    }
{{- else if .IsSint32}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _sz32v{{.Name}}, out int _sz32b{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad sint32 {{.Name}}");
                        this.{{.Name}} = ZigZagDecode32(_sz32v{{.Name}});
                        _pos += _sz32b{{.Name}};
                    }
{{- else if .IsSint64}}
                    {
                        if (!TryReadVarint(binary, _pos, out ulong _sz64v{{.Name}}, out int _sz64b{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad sint64 {{.Name}}");
                        this.{{.Name}} = ZigZagDecode64(_sz64v{{.Name}});
                        _pos += _sz64b{{.Name}};
                    }
{{- else}}
                    {
                        // varint / enum
                        if (!TryReadVarint(binary, _pos, out ulong _varv{{.Name}}, out int _varb{{.Name}}))
                            return global::QiWa.Common.Error.WithLoc(1, "bad varint {{.Name}}");
                        this.{{.Name}} = ({{.WriterType}})_varv{{.Name}};
                        _pos += _varb{{.Name}};
                    }
{{- end}}
                    break;
{{- end}}
                default:
                    if (!SkipField(binary, _pos, _wireType, out int _skip))
                        return global::QiWa.Common.Error.WithLoc(1, "truncated binary data");
                    _pos += _skip;
                    break;
            }
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
}
{{end}}{{end}}
