{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{- $goName := .GoName}}
public partial struct {{$goName}}
{
    // ── ProtobufSize ────────────────────────────────────────────────────────

    public readonly int ProtobufSize()
    {
        int _size = 0;
{{- range .Fields}}
{{- if .IsMap}}
        if ({{.Name}} != null)
        {
            foreach (var _kv{{.Name}} in {{.Name}})
            {
{{- if eq .MapKey "string"}}
                if (_kv{{.Name}}.Key == null) continue;
{{- end}}
                int _entrySize{{.Name}} = 0;
{{- if eq .MapKey "string"}}
                int _klen{{.Name}} = (_kv{{.Name}}.Key.Length > 0) ? Encoding.UTF8.GetByteCount(_kv{{.Name}}.Key) : 0;
                _entrySize{{.Name}} += {{tagSize 1 2}} /* TagSize(1, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_klen{{.Name}}) | 1UL) + 6) / 7 + (_klen{{.Name}}));
{{- else if eq .MapKey "bool"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(_kv{{.Name}}.Key ? 1UL : 0UL) | 1UL) + 6) / 7;
{{- else if eq .MapKey "fixed32"}}
                _entrySize{{.Name}} += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4;
{{- else if eq .MapKey "sfixed32"}}
                _entrySize{{.Name}} += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4;
{{- else if eq .MapKey "fixed64"}}
                _entrySize{{.Name}} += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8;
{{- else if eq .MapKey "sfixed64"}}
                _entrySize{{.Name}} += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8;
{{- else if eq .MapKey "sint32"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32(_kv{{.Name}}.Key)) | 1UL) + 6) / 7;
{{- else if eq .MapKey "sint64"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64(_kv{{.Name}}.Key)) | 1UL) + 6) / 7;
{{- else}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_kv{{.Name}}.Key) | 1UL) + 6) / 7;
{{- end}}
{{- if .MapValIsMsg}}
{{- if .UseMapValWrapper}}
                int _vsize{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.Value.ProtobufSize() : 0;
{{- else}}
                int _vsize{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.ProtobufSize() : 0;
{{- end}}
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_vsize{{.Name}}) | 1UL) + 6) / 7 + (_vsize{{.Name}}));
{{- else if eq .MapVal "string"}}
                int _vlen{{.Name}} = _kv{{.Name}}.Value != null ? Encoding.UTF8.GetByteCount(_kv{{.Name}}.Value) : 0;
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_vlen{{.Name}}) | 1UL) + 6) / 7 + (_vlen{{.Name}}));
{{- else if eq .MapVal "bytes"}}
                int _vlen{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.Length : 0;
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_vlen{{.Name}}) | 1UL) + 6) / 7 + (_vlen{{.Name}}));
{{- else if eq .MapVal "double"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "float"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "fixed32"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "sfixed32"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "fixed64"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "sfixed64"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "sint32"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32(_kv{{.Name}}.Value)) | 1UL) + 6) / 7;
{{- else if eq .MapVal "sint64"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64(_kv{{.Name}}.Value)) | 1UL) + 6) / 7;
{{- else if eq .MapVal "bool"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(_kv{{.Name}}.Value ? 1UL : 0UL) | 1UL) + 6) / 7;
{{- else}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_kv{{.Name}}.Value) | 1UL) + 6) / 7;
{{- end}}
                _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_entrySize{{.Name}}) | 1UL) + 6) / 7 + (_entrySize{{.Name}}));
            }
        }
{{- else if .IsRepeated}}
        if ({{.Name}} != null && {{.Name}}.Count > 0)
        {
{{- if .IsPackable}}
{{- if .IsFixed32}}
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)({{.Name}}.Count * 4) | 1UL) + 6) / 7 + ({{.Name}}.Count * 4));
{{- else if .IsFixed64}}
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)({{.Name}}.Count * 8) | 1UL) + 6) / 7 + ({{.Name}}.Count * 8));
{{- else if .IsBool}}
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)({{.Name}}.Count) | 1UL) + 6) / 7 + ({{.Name}}.Count));
{{- else if .IsSint32}}
            int _packed{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            {
                _packed{{.Name}} += (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32(_pv{{.Name}})) | 1UL) + 6) / 7;
            }
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_packed{{.Name}}) | 1UL) + 6) / 7 + (_packed{{.Name}}));
{{- else if .IsSint64}}
            int _packed{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            {
                _packed{{.Name}} += (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64(_pv{{.Name}})) | 1UL) + 6) / 7;
            }
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_packed{{.Name}}) | 1UL) + 6) / 7 + (_packed{{.Name}}));
{{- else}}
            int _packed{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            {
                _packed{{.Name}} += (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_pv{{.Name}}) | 1UL) + 6) / 7;
            }
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_packed{{.Name}}) | 1UL) + 6) / 7 + (_packed{{.Name}}));
{{- end}}
{{- else if .IsString}}
            foreach (var _sv{{.Name}} in {{.Name}})
            {
                int _len{{.Name}} = (_sv{{.Name}} != null && _sv{{.Name}}.Length > 0) ? Encoding.UTF8.GetByteCount(_sv{{.Name}}) : 0;
                _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_len{{.Name}}) | 1UL) + 6) / 7 + (_len{{.Name}}));
            }
{{- else if .IsBytes}}
            foreach (var _bv{{.Name}} in {{.Name}})
            {
                int _len{{.Name}} = _bv{{.Name}} != null ? _bv{{.Name}}.Length : 0;
                _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_len{{.Name}}) | 1UL) + 6) / 7 + (_len{{.Name}}));
            }
{{- else if .ElemIsMsg}}
            foreach (var _mv{{.Name}} in {{.Name}})
            {
                int _subSize{{.Name}} = _mv{{.Name}}.ProtobufSize();
                if (_subSize{{.Name}} > 0)
                    _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_subSize{{.Name}}) | 1UL) + 6) / 7 + (_subSize{{.Name}}));
            }
{{- else}}
            foreach (var _ev{{.Name}} in {{.Name}})
            {
                _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_ev{{.Name}}) | 1UL) + 6) / 7;
            }
{{- end}}
        }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if ({{.Name}} != null)
        {
            int _subSize{{.Name}} = {{.Name}}.Value.ProtobufSize();
            if (_subSize{{.Name}} > 0)
                _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_subSize{{.Name}}) | 1UL) + 6) / 7 + (_subSize{{.Name}}));
        }
{{- else}}
        {
            int _subSize{{.Name}} = {{.Name}}.ProtobufSize();
            if (_subSize{{.Name}} > 0)
                _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_subSize{{.Name}}) | 1UL) + 6) / 7 + (_subSize{{.Name}}));
        }
{{- end}}
{{- else if .IsString}}
        if (!string.IsNullOrEmpty({{.Name}}))
        {
            int _len{{.Name}} = Encoding.UTF8.GetByteCount({{.Name}});
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_len{{.Name}}) | 1UL) + 6) / 7 + (_len{{.Name}}));
        }
{{- else if .IsBytes}}
        if ({{.Name}} != null && {{.Name}}.Length > 0)
        {
            _size += {{tagSize .Number 2}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)({{.Name}}.Length) | 1UL) + 6) / 7 + ({{.Name}}.Length));
        }
{{- else if .IsBool}}
        if ({{.Name}})
        {
            _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + 1;
        }
{{- else if .IsFixed32}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            _size += {{tagSize .Number 5}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, 32bit=5) */ + 4;
        }
{{- else if .IsDecimal}}
        if ({{.Name}} != 0m)
        {
            _size += {{tagSize .Number 1}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, 64bit=1) */ + 8;
        }
{{- else if .IsFixed64}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            _size += {{tagSize .Number 1}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, 64bit=1) */ + 8;
        }
{{- else if .IsSint32}}
        if ({{.Name}} != 0)
        {
            _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32({{.Name}})) | 1UL) + 6) / 7;
        }
{{- else if .IsSint64}}
        if ({{.Name}} != 0)
        {
            _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64({{.Name}})) | 1UL) + 6) / 7;
        }
{{- else if .IsEnum}}
        if ((int){{.Name}} != 0)
        {
            _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)(int){{.Name}}) | 1UL) + 6) / 7;
        }
{{- else}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            _size += {{tagSize .Number 0}} /* TagSize({{$goName}}Tags.Tag{{.Name}}, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong){{.Name}}) | 1UL) + 6) / 7;
        }
{{- end}}
{{- end}}
        return _size;
    }

    // ── ToProtobuf ───────────────────────────────────────────────────────────

    public readonly Error ToProtobuf(ref RentedBuffer buf)
    {
{{- range .Fields}}
{{- if .IsMap}}
        if ({{.Name}} != null)
        {
            foreach (var _kv{{.Name}} in {{.Name}})
            {
{{- if eq .MapKey "string"}}
                if (_kv{{.Name}}.Key == null) continue;
{{- end}}
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                int _entrySize{{.Name}} = 0;
{{- if eq .MapKey "string"}}
                int _klen{{.Name}} = (_kv{{.Name}}.Key.Length > 0) ? Encoding.UTF8.GetByteCount(_kv{{.Name}}.Key) : 0;
                _entrySize{{.Name}} += {{tagSize 1 2}} /* TagSize(1, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_klen{{.Name}}) | 1UL) + 6) / 7 + (_klen{{.Name}}));
{{- else if eq .MapKey "bool"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(_kv{{.Name}}.Key ? 1UL : 0UL) | 1UL) + 6) / 7;
{{- else if eq .MapKey "fixed32"}}
                _entrySize{{.Name}} += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4;
{{- else if eq .MapKey "sfixed32"}}
                _entrySize{{.Name}} += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4;
{{- else if eq .MapKey "fixed64"}}
                _entrySize{{.Name}} += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8;
{{- else if eq .MapKey "sfixed64"}}
                _entrySize{{.Name}} += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8;
{{- else if eq .MapKey "sint32"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32(_kv{{.Name}}.Key)) | 1UL) + 6) / 7;
{{- else if eq .MapKey "sint64"}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64(_kv{{.Name}}.Key)) | 1UL) + 6) / 7;
{{- else}}
                _entrySize{{.Name}} += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_kv{{.Name}}.Key) | 1UL) + 6) / 7;
{{- end}}
{{- if .MapValIsMsg}}
{{- if .UseMapValWrapper}}
                int _valMsgSize{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.Value.ProtobufSize() : 0;
{{- else}}
                int _valMsgSize{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.ProtobufSize() : 0;
{{- end}}
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_valMsgSize{{.Name}}) | 1UL) + 6) / 7 + (_valMsgSize{{.Name}}));
{{- else if eq .MapVal "string"}}
                int _vlen{{.Name}} = (_kv{{.Name}}.Value != null && _kv{{.Name}}.Value.Length > 0) ? Encoding.UTF8.GetByteCount(_kv{{.Name}}.Value) : 0;
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_vlen{{.Name}}) | 1UL) + 6) / 7 + (_vlen{{.Name}}));
{{- else if eq .MapVal "bytes"}}
                int _vlen{{.Name}} = _kv{{.Name}}.Value != null ? _kv{{.Name}}.Value.Length : 0;
                _entrySize{{.Name}} += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + ((64 - BitOperations.LeadingZeroCount((ulong)(_vlen{{.Name}}) | 1UL) + 6) / 7 + (_vlen{{.Name}}));
{{- else if eq .MapVal "double"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "float"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "fixed32"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "sfixed32"}}
                _entrySize{{.Name}} += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4;
{{- else if eq .MapVal "fixed64"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "sfixed64"}}
                _entrySize{{.Name}} += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8;
{{- else if eq .MapVal "sint32"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode32(_kv{{.Name}}.Value)) | 1UL) + 6) / 7;
{{- else if eq .MapVal "sint64"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(ZigZagEncode64(_kv{{.Name}}.Value)) | 1UL) + 6) / 7;
{{- else if eq .MapVal "bool"}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)(_kv{{.Name}}.Value ? 1UL : 0UL) | 1UL) + 6) / 7;
{{- else}}
                _entrySize{{.Name}} += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (64 - BitOperations.LeadingZeroCount((ulong)((ulong)_kv{{.Name}}.Value) | 1UL) + 6) / 7;
{{- end}}
                WriteVarint(ref buf, (ulong)_entrySize{{.Name}});
                // key (field 1)
{{- if eq .MapKey "string"}}
                WriteTag(ref buf, 1, WireTypeLenDelim);
                WriteString(ref buf, _kv{{.Name}}.Key ?? string.Empty);
{{- else if eq .MapKey "bool"}}
                WriteTag(ref buf, 1, WireTypeVarint);
                WriteVarint(ref buf, _kv{{.Name}}.Key ? 1UL : 0UL);
{{- else if eq .MapKey "fixed32"}}
                WriteTag(ref buf, 1, WireType32Bit);
                WriteFixed32(ref buf, _kv{{.Name}}.Key);
{{- else if eq .MapKey "sfixed32"}}
                WriteTag(ref buf, 1, WireType32Bit);
                WriteFixed32(ref buf, (uint)_kv{{.Name}}.Key);
{{- else if eq .MapKey "fixed64"}}
                WriteTag(ref buf, 1, WireType64Bit);
                WriteFixed64(ref buf, _kv{{.Name}}.Key);
{{- else if eq .MapKey "sfixed64"}}
                WriteTag(ref buf, 1, WireType64Bit);
                WriteFixed64(ref buf, (ulong)_kv{{.Name}}.Key);
{{- else if eq .MapKey "sint32"}}
                WriteTag(ref buf, 1, WireTypeVarint);
                WriteVarint(ref buf, ZigZagEncode32(_kv{{.Name}}.Key));
{{- else if eq .MapKey "sint64"}}
                WriteTag(ref buf, 1, WireTypeVarint);
                WriteVarint(ref buf, ZigZagEncode64(_kv{{.Name}}.Key));
{{- else}}
                WriteTag(ref buf, 1, WireTypeVarint);
                WriteVarint(ref buf, (ulong)_kv{{.Name}}.Key);
{{- end}}
                // value (field 2)
{{- if .MapValIsMsg}}
                WriteTag(ref buf, 2, WireTypeLenDelim);
                WriteVarint(ref buf, (ulong)_valMsgSize{{.Name}});
                if (_valMsgSize{{.Name}} > 0)
                {
{{- if .UseMapValWrapper}}
                    _kv{{.Name}}.Value.Value.ToProtobuf(ref buf);
{{- else}}
                    _kv{{.Name}}.Value.ToProtobuf(ref buf);
{{- end}}
                }
{{- else if eq .MapVal "string"}}
                WriteTag(ref buf, 2, WireTypeLenDelim);
                WriteString(ref buf, _kv{{.Name}}.Value ?? string.Empty);
{{- else if eq .MapVal "bytes"}}
                WriteTag(ref buf, 2, WireTypeLenDelim);
                WriteBytes(ref buf, _kv{{.Name}}.Value ?? Array.Empty<byte>());
{{- else if eq .MapVal "double"}}
                WriteTag(ref buf, 2, WireType64Bit);
                WriteFixed64(ref buf, BitConverter.DoubleToUInt64Bits(_kv{{.Name}}.Value));
{{- else if eq .MapVal "float"}}
                WriteTag(ref buf, 2, WireType32Bit);
                WriteFixed32(ref buf, BitConverter.SingleToUInt32Bits(_kv{{.Name}}.Value));
{{- else if eq .MapVal "fixed32"}}
                WriteTag(ref buf, 2, WireType32Bit);
                WriteFixed32(ref buf, _kv{{.Name}}.Value);
{{- else if eq .MapVal "sfixed32"}}
                WriteTag(ref buf, 2, WireType32Bit);
                WriteFixed32(ref buf, (uint)_kv{{.Name}}.Value);
{{- else if eq .MapVal "fixed64"}}
                WriteTag(ref buf, 2, WireType64Bit);
                WriteFixed64(ref buf, _kv{{.Name}}.Value);
{{- else if eq .MapVal "sfixed64"}}
                WriteTag(ref buf, 2, WireType64Bit);
                WriteFixed64(ref buf, (ulong)_kv{{.Name}}.Value);
{{- else if eq .MapVal "sint32"}}
                WriteTag(ref buf, 2, WireTypeVarint);
                WriteVarint(ref buf, ZigZagEncode32(_kv{{.Name}}.Value));
{{- else if eq .MapVal "sint64"}}
                WriteTag(ref buf, 2, WireTypeVarint);
                WriteVarint(ref buf, ZigZagEncode64(_kv{{.Name}}.Value));
{{- else if eq .MapVal "bool"}}
                WriteTag(ref buf, 2, WireTypeVarint);
                WriteVarint(ref buf, _kv{{.Name}}.Value ? 1UL : 0UL);
{{- else}}
                WriteTag(ref buf, 2, WireTypeVarint);
                WriteVarint(ref buf, (ulong)_kv{{.Name}}.Value);
{{- end}}
            }
        }
{{- else if .IsRepeated}}
        if ({{.Name}} != null && {{.Name}}.Count > 0)
        {
{{- if .IsPackable}}
{{- if .IsFixed32}}
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong)({{.Name}}.Count * 4));
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteFixed32(ref buf, {{if eq .Type "float"}}BitConverter.SingleToUInt32Bits(_pv{{.Name}}){{else if eq .Type "sfixed32"}}(uint)_pv{{.Name}}{{else}}_pv{{.Name}}{{end}});
{{- else if .IsFixed64}}
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong)({{.Name}}.Count * 8));
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteFixed64(ref buf, {{if eq .Type "double"}}BitConverter.DoubleToUInt64Bits(_pv{{.Name}}){{else if eq .Type "sfixed64"}}(ulong)_pv{{.Name}}{{else}}_pv{{.Name}}{{end}});
{{- else if .IsBool}}
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong){{.Name}}.Count);
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteVarint(ref buf, _pv{{.Name}} ? 1UL : 0UL);
{{- else if .IsSint32}}
            int _len{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            { ulong _zv{{.Name}} = ZigZagEncode32(_pv{{.Name}}); do { _len{{.Name}}++; _zv{{.Name}} >>= 7; } while (_zv{{.Name}} != 0); }
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong)_len{{.Name}});
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteVarint(ref buf, ZigZagEncode32(_pv{{.Name}}));
{{- else if .IsSint64}}
            int _len{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            { ulong _zv{{.Name}} = ZigZagEncode64(_pv{{.Name}}); do { _len{{.Name}}++; _zv{{.Name}} >>= 7; } while (_zv{{.Name}} != 0); }
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong)_len{{.Name}});
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteVarint(ref buf, ZigZagEncode64(_pv{{.Name}}));
{{- else}}
            int _len{{.Name}} = 0;
            foreach (var _pv{{.Name}} in {{.Name}})
            { ulong _vv{{.Name}} = (ulong)_pv{{.Name}}; do { _len{{.Name}}++; _vv{{.Name}} >>= 7; } while (_vv{{.Name}} != 0); }
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteVarint(ref buf, (ulong)_len{{.Name}});
            foreach (var _pv{{.Name}} in {{.Name}})
                WriteVarint(ref buf, (ulong)_pv{{.Name}});
{{- end}}
{{- else if .IsString}}
            foreach (var _sv{{.Name}} in {{.Name}})
            {
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                WriteString(ref buf, _sv{{.Name}} ?? string.Empty);
            }
{{- else if .IsBytes}}
            foreach (var _bv{{.Name}} in {{.Name}})
            {
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                WriteBytes(ref buf, _bv{{.Name}});
            }
{{- else if .ElemIsMsg}}
            foreach (var _mv{{.Name}} in {{.Name}})
            {
                int len = _mv{{.Name}}.ProtobufSize();
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                WriteVarint(ref buf, (ulong)len);
                if (len>0) _mv{{.Name}}.ToProtobuf(ref buf);
            }
{{- else}}
            foreach (var _ev{{.Name}} in {{.Name}})
            {
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
                WriteVarint(ref buf, (ulong)_ev{{.Name}});
            }
{{- end}}
        }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if ({{.Name}} != null)
        {
            int len = {{.Name}}.Value.ProtobufSize();
            if (len>0)
            {
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                WriteVarint(ref buf, (ulong)len);
                {{.Name}}.Value.ToProtobuf(ref buf);
            }
        }
{{- else}}
        {
            int len = {{.Name}}.ProtobufSize();
            if (len>0)
            {
                WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
                WriteVarint(ref buf, (ulong)len);
                {{.Name}}.ToProtobuf(ref buf);
            }
        }
{{- end}}
{{- else if .IsString}}
        if (!string.IsNullOrEmpty({{.Name}}))
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteString(ref buf, {{.Name}});
        }
{{- else if .IsBytes}}
        if ({{.Name}} != null && {{.Name}}.Length > 0)
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeLenDelim);
            WriteBytes(ref buf, {{.Name}});
        }
{{- else if .IsBool}}
        if ({{.Name}})
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
            WriteVarint(ref buf, 1UL);
        }
{{- else if .IsFixed32}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireType32Bit);
{{- if eq .Type "float"}}
            WriteFixed32(ref buf, BitConverter.SingleToUInt32Bits({{.Name}}));
{{- else if eq .Type "sfixed32"}}
            WriteFixed32(ref buf, (uint){{.Name}});
{{- else}}
            WriteFixed32(ref buf, {{.Name}});
{{- end}}
        }
{{- else if .IsDecimal}}
        if ({{.Name}} != 0m)
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireType64Bit);
            WriteFixed64(ref buf, BitConverter.DoubleToUInt64Bits((double)Math.Round({{.Name}}, {{.DecimalRound}}, MidpointRounding.AwayFromZero)));
        }
{{- else if .IsFixed64}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireType64Bit);
{{- if eq .Type "double"}}
            WriteFixed64(ref buf, BitConverter.DoubleToUInt64Bits({{.Name}}));
{{- else if eq .Type "sfixed64"}}
            WriteFixed64(ref buf, (ulong){{.Name}});
{{- else}}
            WriteFixed64(ref buf, {{.Name}});
{{- end}}
        }
{{- else if .IsSint32}}
        if ({{.Name}} != 0)
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
            WriteVarint(ref buf, ZigZagEncode32({{.Name}}));
        }
{{- else if .IsSint64}}
        if ({{.Name}} != 0)
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
            WriteVarint(ref buf, ZigZagEncode64({{.Name}}));
        }
{{- else if .IsEnum}}
        if ((int){{.Name}} != 0)
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
            WriteVarint(ref buf, (ulong)(int){{.Name}});
        }
{{- else}}
        if ({{.Name}} != {{csDefault .WriterType}})
        {
            WriteTag(ref buf, {{$goName}}Tags.Tag{{.Name}}, WireTypeVarint);
            WriteVarint(ref buf, (ulong){{.Name}});
        }
{{- end}}
{{- end}}
        return default;
    }
}
{{end}}{{end}}
