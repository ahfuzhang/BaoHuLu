{{/* Protobuf 编码：大小计算、append / VT 编码及其辅助方法。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"math"
	"math/bits"
	"unsafe"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = math.Float32bits
	_ = bits.Len64
	_ = unsafe.Pointer(nil)
	_ = utils.AppendVarint
)

{{with .Msg}}
{{- $goName := .GoName}}
func (m *{{$goName}}) ProtobufSize() int {
	size := 0
{{- range .Fields}}
{{- if .Map}}
	for k, v := range m.{{.Name}} {
		entrySize := 0
		{{- if eq .MapKey "string"}}
		{
			kn := len(k)
			entrySize += {{tagSize 1 2}} /* TagSize(1, LenDelim=2) */ + (bits.Len64((uint64(kn))|1) + 6) / 7 + kn
		}
		{{- else if eq .MapKey "bool"}}
		{
			var bk uint64
			if k {
				bk = 1
			}
			entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64((bk)|1) + 6) / 7
		}
		{{- else if eq .MapKey "fixed32"}}
		_ = k
		entrySize += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4
		{{- else if eq .MapKey "sfixed32"}}
		_ = k
		entrySize += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4
		{{- else if eq .MapKey "fixed64"}}
		_ = k
		entrySize += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8
		{{- else if eq .MapKey "sfixed64"}}
		_ = k
		entrySize += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8
		{{- else if eq .MapKey "sint32"}}
		entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64((uint64((uint32(k)<<1)^uint32(k>>31)))|1) + 6) / 7
		{{- else if eq .MapKey "sint64"}}
		entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64(((uint64(k)<<1)^uint64(k>>63))|1) + 6) / 7
		{{- else}}
		entrySize += {{tagSize 1 (protoWireTypeInt .MapKey)}} /* TagSize(1, {{protoWireType .MapKey}}) */ + (bits.Len64((uint64(k))|1) + 6) / 7
		{{- end}}
		{{- if eq .MapVal "string"}}
		{
			vn := len(v)
			entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(vn))|1) + 6) / 7 + vn
		}
		{{- else if eq .MapVal "bytes"}}
		{
			vn := len(v)
			entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(vn))|1) + 6) / 7 + vn
		}
		{{- else if mapValIsMsg .MapVal}}
		{
			var sub int
			if v != nil {
				sub = v.ProtobufSize()
			}
			entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(sub))|1) + 6) / 7 + sub
		}
		{{- else if eq .MapVal "double"}}
		_ = v
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else if eq .MapVal "float"}}
		_ = v
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "fixed32"}}
		_ = v
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "fixed64"}}
		_ = v
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else if eq .MapVal "sfixed32"}}
		_ = v
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "sfixed64"}}
		_ = v
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else}}
		entrySize += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (bits.Len64((uint64(v))|1) + 6) / 7
		{{- end}}
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(entrySize))|1) + 6) / 7 + entrySize
	}
{{- else if .Repeated}}
{{- if isPackable .Type}}
	if len(m.{{.Name}}) > 0 {
		packed := 0
		for _, v := range m.{{.Name}} {
			{{- if eq .Type "sint32"}}
			packed += (bits.Len64((uint64((uint32(v) << 1) ^ uint32(v>>31)))|1) + 6) / 7
			{{- else if eq .Type "sint64"}}
			packed += (bits.Len64(((uint64(v) << 1) ^ uint64(v>>63))|1) + 6) / 7
			{{- else if eq .Type "fixed32"}}
			_ = v; packed += 4
			{{- else if eq .Type "fixed64"}}
			_ = v; packed += 8
			{{- else if eq .Type "sfixed32"}}
			_ = v; packed += 4
			{{- else if eq .Type "sfixed64"}}
			_ = v; packed += 8
			{{- else if eq .Type "float"}}
			_ = v; packed += 4
			{{- else if eq .Type "double"}}
			_ = v; packed += 8
			{{- else if eq .Type "bool"}}
			_ = v; packed += 1
			{{- else}}
			packed += (bits.Len64((uint64(v))|1) + 6) / 7
			{{- end}}
		}
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(packed))|1) + 6) / 7 + packed
	}
{{- else if eq .Type "string"}}
	for _, v := range m.{{.Name}} {
		n := len(v)
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(n))|1) + 6) / 7 + n
	}
{{- else if eq .Type "bytes"}}
	for _, v := range m.{{.Name}} {
		n := len(v)
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(n))|1) + 6) / 7 + n
	}
{{- else}}
	for i := range m.{{.Name}} {
		sub := m.{{.Name}}[i].ProtobufSize()
		if sub > 0 {
			size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(sub))|1) + 6) / 7 + sub
		}
	}
{{- end}}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if m.{{.Name}} != nil {
		sub := m.{{.Name}}.ProtobufSize()
		if sub > 0 {
			size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(sub))|1) + 6) / 7 + sub
		}
	}
{{- else}}
	{
		sub := m.{{.Name}}.ProtobufSize()
		if sub > 0 {
			size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(sub))|1) + 6) / 7 + sub
		}
	}
{{- end}}
{{- else if .IsDecimal}}
	if !m.{{.Name}}.IsZero() {
		size += {{tagSize .Number 1}} /* TagSize({{$goName}}{{.Name}}Tag, 64bit=1) */ + 8
	}
{{- else if eq .Type "double"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 1}} /* TagSize({{$goName}}{{.Name}}Tag, 64bit=1) */ + 8
	}
{{- else if eq .Type "float"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 5}} /* TagSize({{$goName}}{{.Name}}Tag, 32bit=5) */ + 4
	}
{{- else if eq .Type "fixed32"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 5}} /* TagSize({{$goName}}{{.Name}}Tag, 32bit=5) */ + 4
	}
{{- else if eq .Type "fixed64"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 1}} /* TagSize({{$goName}}{{.Name}}Tag, 64bit=1) */ + 8
	}
{{- else if eq .Type "sfixed32"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 5}} /* TagSize({{$goName}}{{.Name}}Tag, 32bit=5) */ + 4
	}
{{- else if eq .Type "sfixed64"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 1}} /* TagSize({{$goName}}{{.Name}}Tag, 64bit=1) */ + 8
	}
{{- else if eq .Type "sint32"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 0}} /* TagSize({{$goName}}{{.Name}}Tag, Varint=0) */ + (bits.Len64((uint64((uint32(m.{{.Name}})<<1)^uint32(m.{{.Name}}>>31)))|1) + 6) / 7
	}
{{- else if eq .Type "sint64"}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 0}} /* TagSize({{$goName}}{{.Name}}Tag, Varint=0) */ + (bits.Len64(((uint64(m.{{.Name}})<<1)^uint64(m.{{.Name}}>>63))|1) + 6) / 7
	}
{{- else if eq .Type "bool"}}
	if m.{{.Name}} {
		size += {{tagSize .Number 0}} /* TagSize({{$goName}}{{.Name}}Tag, Varint=0) */ + 1
	}
{{- else if eq .Type "string"}}
	if m.{{.Name}} != "" {
		n := len(m.{{.Name}})
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(n))|1) + 6) / 7 + n
	}
{{- else if eq .Type "bytes"}}
	if len(m.{{.Name}}) > 0 {
		n := len(m.{{.Name}})
		size += {{tagSize .Number 2}} /* TagSize({{$goName}}{{.Name}}Tag, LenDelim=2) */ + (bits.Len64((uint64(n))|1) + 6) / 7 + n
	}
{{- else}}
	if m.{{.Name}} != 0 {
		size += {{tagSize .Number 0}} /* TagSize({{$goName}}{{.Name}}Tag, Varint=0) */ + (bits.Len64((uint64(m.{{.Name}}))|1) + 6) / 7
	}
{{- end}}
{{- end}}
	return size
}

func (m *{{$goName}}) ToProtobufByAppend(in []byte) []byte {
{{- range .ReverseFields}}
{{- if .Map}}
	for k, v := range m.{{.Name}} {
		entrySize := 0
		{{- if eq .MapKey "string"}}
		{
			kn := len(k)
			entrySize += {{tagSize 1 2}} /* TagSize(1, LenDelim=2) */ + (bits.Len64((uint64(kn))|1) + 6) / 7 + kn
		}
		{{- else if eq .MapKey "bool"}}
		{
			var bk uint64
			if k {
				bk = 1
			}
			entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64((bk)|1) + 6) / 7
		}
		{{- else if eq .MapKey "fixed32"}}
		entrySize += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4
		{{- else if eq .MapKey "sfixed32"}}
		entrySize += {{tagSize 1 5}} /* TagSize(1, 32bit=5) */ + 4
		{{- else if eq .MapKey "fixed64"}}
		entrySize += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8
		{{- else if eq .MapKey "sfixed64"}}
		entrySize += {{tagSize 1 1}} /* TagSize(1, 64bit=1) */ + 8
		{{- else if eq .MapKey "sint32"}}
		entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64((uint64((uint32(k)<<1)^uint32(k>>31)))|1) + 6) / 7
		{{- else if eq .MapKey "sint64"}}
		entrySize += {{tagSize 1 0}} /* TagSize(1, Varint=0) */ + (bits.Len64(((uint64(k)<<1)^uint64(k>>63))|1) + 6) / 7
		{{- else}}
		entrySize += {{tagSize 1 (protoWireTypeInt .MapKey)}} /* TagSize(1, {{protoWireType .MapKey}}) */ + (bits.Len64((uint64(k))|1) + 6) / 7
		{{- end}}
		{{- if eq .MapVal "string"}}
		{
			vn := len(v)
			entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(vn))|1) + 6) / 7 + vn
		}
		{{- else if eq .MapVal "bytes"}}
		{
			vn := len(v)
			entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(vn))|1) + 6) / 7 + vn
		}
		{{- else if (mapValIsMsg .MapVal)}}
		var valMsgSize int
		if v != nil {
			valMsgSize = v.ProtobufSize()
		}
		entrySize += {{tagSize 2 2}} /* TagSize(2, LenDelim=2) */ + (bits.Len64((uint64(valMsgSize))|1) + 6) / 7 + valMsgSize
		{{- else if eq .MapVal "double"}}
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else if eq .MapVal "float"}}
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "fixed32"}}
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "fixed64"}}
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else if eq .MapVal "sfixed32"}}
		entrySize += {{tagSize 2 5}} /* TagSize(2, 32bit=5) */ + 4
		{{- else if eq .MapVal "sfixed64"}}
		entrySize += {{tagSize 2 1}} /* TagSize(2, 64bit=1) */ + 8
		{{- else}}
		entrySize += {{tagSize 2 0}} /* TagSize(2, Varint=0) */ + (bits.Len64((uint64(v))|1) + 6) / 7
		{{- end}}
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendVarint(in, uint64(entrySize))
		in = utils.AppendTag(in, 1, {{protoWireType .MapKey}})
		{{- if eq .MapKey "string"}}
		in = utils.AppendVarint(in, uint64(len(k)))
		in = append(in, k...)
		{{- else if eq .MapKey "bool"}}
		if k {
			in = utils.AppendVarint(in, 1)
		} else {
			in = utils.AppendVarint(in, 0)
		}
		{{- else if eq .MapKey "fixed32"}}
		in = utils.AppendFixed32(in, k)
		{{- else if eq .MapKey "sfixed32"}}
		in = utils.AppendFixed32(in, uint32(k))
		{{- else if eq .MapKey "fixed64"}}
		in = utils.AppendFixed64(in, k)
		{{- else if eq .MapKey "sfixed64"}}
		in = utils.AppendFixed64(in, uint64(k))
		{{- else if eq .MapKey "sint32"}}
		in = utils.AppendSint32(in, k)
		{{- else if eq .MapKey "sint64"}}
		in = utils.AppendSint64(in, k)
		{{- else}}
		in = utils.AppendVarint(in, uint64(k))
		{{- end}}
		{{- if (mapValIsMsg .MapVal)}}
		in = utils.AppendTag(in, 2, utils.WireTypeLenDelim)
		{{- else}}
		in = utils.AppendTag(in, 2, {{protoWireType .MapVal}})
		{{- end}}
		{{- if eq .MapVal "string"}}
		in = utils.AppendVarint(in, uint64(len(v)))
		in = append(in, v...)
		{{- else if eq .MapVal "bytes"}}
		in = utils.AppendLenDelim(in, v)
		{{- else if (mapValIsMsg .MapVal)}}
		in = utils.AppendVarint(in, uint64(valMsgSize))
		if v != nil {
			in = v.ToProtobuf(in)
		}
		{{- else if eq .MapVal "double"}}
		in = utils.AppendFixed64(in, math.Float64bits(v))
		{{- else if eq .MapVal "float"}}
		in = utils.AppendFixed32(in, math.Float32bits(v))
		{{- else if eq .MapVal "fixed32"}}
		in = utils.AppendFixed32(in, uint32(v))
		{{- else if eq .MapVal "fixed64"}}
		in = utils.AppendFixed64(in, uint64(v))
		{{- else if eq .MapVal "sfixed32"}}
		in = utils.AppendFixed32(in, uint32(v))
		{{- else if eq .MapVal "sfixed64"}}
		in = utils.AppendFixed64(in, uint64(v))
		{{- else}}
		in = utils.AppendVarint(in, uint64(v))
		{{- end}}
	}
{{- else if .Repeated}}
{{- if isPackable .Type}}
	if len(m.{{.Name}}) > 0 {
		packedSize := 0
		for _, v := range m.{{.Name}} {
			{{- if eq .Type "sint32"}}
			packedSize += (bits.Len64((uint64((uint32(v) << 1) ^ uint32(v>>31)))|1) + 6) / 7
			{{- else if eq .Type "sint64"}}
			packedSize += (bits.Len64(((uint64(v) << 1) ^ uint64(v>>63))|1) + 6) / 7
			{{- else if eq .Type "fixed32"}}
			_ = v; packedSize += 4
			{{- else if eq .Type "fixed64"}}
			_ = v; packedSize += 8
			{{- else if eq .Type "sfixed32"}}
			_ = v; packedSize += 4
			{{- else if eq .Type "sfixed64"}}
			_ = v; packedSize += 8
			{{- else if eq .Type "float"}}
			_ = v; packedSize += 4
			{{- else if eq .Type "double"}}
			_ = v; packedSize += 8
			{{- else if eq .Type "bool"}}
			_ = v; packedSize += 1
			{{- else}}
			packedSize += (bits.Len64((uint64(v))|1) + 6) / 7
			{{- end}}
		}
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendVarint(in, uint64(packedSize))
		for _, v := range m.{{.Name}} {
			{{- if eq .Type "sint32"}}
			in = utils.AppendSint32(in, v)
			{{- else if eq .Type "sint64"}}
			in = utils.AppendSint64(in, v)
			{{- else if eq .Type "fixed32"}}
			in = utils.AppendFixed32(in, uint32(v))
			{{- else if eq .Type "fixed64"}}
			in = utils.AppendFixed64(in, uint64(v))
			{{- else if eq .Type "sfixed32"}}
			in = utils.AppendFixed32(in, uint32(v))
			{{- else if eq .Type "sfixed64"}}
			in = utils.AppendFixed64(in, uint64(v))
			{{- else if eq .Type "float"}}
			in = utils.AppendFixed32(in, math.Float32bits(v))
			{{- else if eq .Type "double"}}
			in = utils.AppendFixed64(in, math.Float64bits(v))
			{{- else if eq .Type "bool"}}
			if v {
				in = append(in, 1)
			} else {
				in = append(in, 0)
			}
			{{- else}}
			in = utils.AppendVarint(in, uint64(v))
			{{- end}}
		}
	}
{{- else if eq .Type "string"}}
	for _, v := range m.{{.Name}} {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendLenDelim(in, unsafe.Slice(unsafe.StringData(v), len(v)))
	}
{{- else if eq .Type "bytes"}}
	for _, v := range m.{{.Name}} {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendLenDelim(in, v)
	}
{{- else}}
	for i := range m.{{.Name}} {
		sub := m.{{.Name}}[i].ProtobufSize()
		if sub > 0 {
			in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
			in = utils.AppendVarint(in, uint64(sub))
			in = m.{{.Name}}[i].ToProtobuf(in)
		}
	}
{{- end}}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if m.{{.Name}} != nil {
		sub := m.{{.Name}}.ProtobufSize()
		if sub > 0 {
			in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
			in = utils.AppendVarint(in, uint64(sub))
			in = m.{{.Name}}.ToProtobuf(in)
		}
	}
{{- else}}
	{
		sub := m.{{.Name}}.ProtobufSize()
		if sub > 0 {
			in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
			in = utils.AppendVarint(in, uint64(sub))
			in = m.{{.Name}}.ToProtobuf(in)
		}
	}
{{- end}}
{{- else if .IsDecimal}}
	if !m.{{.Name}}.IsZero() {
		_df{{.Name}}, _ := m.{{.Name}}.Round({{.DecimalRound}}).Float64()
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType64bit)
		in = utils.AppendFixed64(in, math.Float64bits(_df{{.Name}}))
	}
{{- else if eq .Type "double"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType64bit)
		in = utils.AppendFixed64(in, math.Float64bits(m.{{.Name}}))
	}
{{- else if eq .Type "float"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType32bit)
		in = utils.AppendFixed32(in, math.Float32bits(m.{{.Name}}))
	}
{{- else if eq .Type "fixed32"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType32bit)
		in = utils.AppendFixed32(in, m.{{.Name}})
	}
{{- else if eq .Type "fixed64"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType64bit)
		in = utils.AppendFixed64(in, m.{{.Name}})
	}
{{- else if eq .Type "sfixed32"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType32bit)
		in = utils.AppendFixed32(in, uint32(m.{{.Name}}))
	}
{{- else if eq .Type "sfixed64"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireType64bit)
		in = utils.AppendFixed64(in, uint64(m.{{.Name}}))
	}
{{- else if eq .Type "sint32"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeVarint)
		in = utils.AppendSint32(in, m.{{.Name}})
	}
{{- else if eq .Type "sint64"}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeVarint)
		in = utils.AppendSint64(in, m.{{.Name}})
	}
{{- else if eq .Type "bool"}}
	if m.{{.Name}} {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeVarint)
		in = append(in, 1)
	}
{{- else if eq .Type "string"}}
	if m.{{.Name}} != "" {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendLenDelim(in, unsafe.Slice(unsafe.StringData(m.{{.Name}}), len(m.{{.Name}})))
	}
{{- else if eq .Type "bytes"}}
	if len(m.{{.Name}}) > 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeLenDelim)
		in = utils.AppendLenDelim(in, m.{{.Name}})
	}
{{- else}}
	if m.{{.Name}} != 0 {
		in = utils.AppendTag(in, {{$goName}}{{.Name}}Tag, utils.WireTypeVarint)
		in = utils.AppendVarint(in, uint64(m.{{.Name}}))
	}
{{- end}}
{{- end}}
	return in
}

// marshalToSizedBufferVT writes m into dAtA using vtprotobuf's backward-fill
// approach (starts at the end, works toward the front).
// Returns number of bytes written = len(dAtA) - final_i.
func (m *{{$goName}}) marshalToSizedBufferVT(dAtA []byte) int {
	if m == nil {
		return 0
	}
	i := len(dAtA)
	_ = i
{{- range .Fields}}
{{- if .Map}}
	for k := range m.{{.Name}} {
		v := m.{{.Name}}[k]
		baseI := i
		{{- if eq .MapVal "string"}}
		i -= len(v)
		copy(dAtA[i:], v)
		i = utils.EncodeVarint(dAtA, i, uint64(len(v)))
		i--
		dAtA[i] = 18 /*(2<<3)|2=18, field=2, wireType=LenDelim*/
		{{- else if eq .MapVal "bytes"}}
		i -= len(v)
		copy(dAtA[i:], v)
		i = utils.EncodeVarint(dAtA, i, uint64(len(v)))
		i--
		dAtA[i] = 18 /*(2<<3)|2=18, field=2, wireType=LenDelim*/
		{{- else if mapValIsMsg .MapVal}}
		size := v.marshalToSizedBufferVT(dAtA[:i])
		i -= size
		i = utils.EncodeVarint(dAtA, i, uint64(size))
		i--
		dAtA[i] = 18 /*(2<<3)|2=18, field=2, wireType=LenDelim*/
		{{- else if eq .MapVal "double"}}
		i -= 8
		_dv := math.Float64bits(v)
		dAtA[i+7] = uint8(_dv >> 56)
		dAtA[i+6] = uint8(_dv >> 48)
		dAtA[i+5] = uint8(_dv >> 40)
		dAtA[i+4] = uint8(_dv >> 32)
		dAtA[i+3] = uint8(_dv >> 24)
		dAtA[i+2] = uint8(_dv >> 16)
		dAtA[i+1] = uint8(_dv >> 8)
		dAtA[i+0] = uint8(_dv)
		i--
		dAtA[i] = 17 /*(2<<3)|1=17, field=2, wireType=64bit*/
		{{- else if eq .MapVal "float"}}
		i -= 4
		_fv := math.Float32bits(v)
		dAtA[i+3] = uint8(_fv >> 24)
		dAtA[i+2] = uint8(_fv >> 16)
		dAtA[i+1] = uint8(_fv >> 8)
		dAtA[i+0] = uint8(_fv)
		i--
		dAtA[i] = 21 /*(2<<3)|5=21, field=2, wireType=32bit*/
		{{- else if eq .MapVal "fixed32"}}
		i -= 4
		dAtA[i+3] = uint8(uint32(v) >> 24)
		dAtA[i+2] = uint8(uint32(v) >> 16)
		dAtA[i+1] = uint8(uint32(v) >> 8)
		dAtA[i+0] = uint8(uint32(v))
		i--
		dAtA[i] = 21 /*(2<<3)|5=21, field=2, wireType=32bit*/
		{{- else if eq .MapVal "sfixed32"}}
		i -= 4
		dAtA[i+3] = uint8(uint32(v) >> 24)
		dAtA[i+2] = uint8(uint32(v) >> 16)
		dAtA[i+1] = uint8(uint32(v) >> 8)
		dAtA[i+0] = uint8(uint32(v))
		i--
		dAtA[i] = 21 /*(2<<3)|5=21, field=2, wireType=32bit*/
		{{- else if eq .MapVal "fixed64"}}
		i -= 8
		dAtA[i+7] = uint8(uint64(v) >> 56)
		dAtA[i+6] = uint8(uint64(v) >> 48)
		dAtA[i+5] = uint8(uint64(v) >> 40)
		dAtA[i+4] = uint8(uint64(v) >> 32)
		dAtA[i+3] = uint8(uint64(v) >> 24)
		dAtA[i+2] = uint8(uint64(v) >> 16)
		dAtA[i+1] = uint8(uint64(v) >> 8)
		dAtA[i+0] = uint8(uint64(v))
		i--
		dAtA[i] = 17 /*(2<<3)|1=17, field=2, wireType=64bit*/
		{{- else if eq .MapVal "sfixed64"}}
		i -= 8
		dAtA[i+7] = uint8(uint64(v) >> 56)
		dAtA[i+6] = uint8(uint64(v) >> 48)
		dAtA[i+5] = uint8(uint64(v) >> 40)
		dAtA[i+4] = uint8(uint64(v) >> 32)
		dAtA[i+3] = uint8(uint64(v) >> 24)
		dAtA[i+2] = uint8(uint64(v) >> 16)
		dAtA[i+1] = uint8(uint64(v) >> 8)
		dAtA[i+0] = uint8(uint64(v))
		i--
		dAtA[i] = 17 /*(2<<3)|1=17, field=2, wireType=64bit*/
		{{- else if eq .MapVal "sint32"}}
		i = utils.EncodeVarint(dAtA, i, uint64((uint32(v)<<1)^uint32(v>>31)))
		i--
		dAtA[i] = 16 /*(2<<3)|0=16, field=2, wireType=Varint*/
		{{- else if eq .MapVal "sint64"}}
		i = utils.EncodeVarint(dAtA, i, (uint64(v)<<1)^uint64(v>>63))
		i--
		dAtA[i] = 16 /*(2<<3)|0=16, field=2, wireType=Varint*/
		{{- else if eq .MapVal "bool"}}
		i--
		if v { dAtA[i] = 1 } else { dAtA[i] = 0 }
		i--
		dAtA[i] = 16 /*(2<<3)|0=16, field=2, wireType=Varint*/
		{{- else}}
		i = utils.EncodeVarint(dAtA, i, uint64(v))
		i--
		dAtA[i] = 16 /*(2<<3)|0=16, field=2, wireType=Varint*/
		{{- end}}
		{{- if eq .MapKey "string"}}
		i -= len(k)
		copy(dAtA[i:], k)
		i = utils.EncodeVarint(dAtA, i, uint64(len(k)))
		i--
		dAtA[i] = 10 /*(1<<3)|2=10, field=1, wireType=LenDelim*/
		{{- else if eq .MapKey "bool"}}
		i--
		if k { dAtA[i] = 1 } else { dAtA[i] = 0 }
		i--
		dAtA[i] = 8 /*(1<<3)|0=8, field=1, wireType=Varint*/
		{{- else if eq .MapKey "fixed32"}}
		i -= 4
		dAtA[i+3] = uint8(uint32(k) >> 24)
		dAtA[i+2] = uint8(uint32(k) >> 16)
		dAtA[i+1] = uint8(uint32(k) >> 8)
		dAtA[i+0] = uint8(uint32(k))
		i--
		dAtA[i] = 13 /*(1<<3)|5=13, field=1, wireType=32bit*/
		{{- else if eq .MapKey "sfixed32"}}
		i -= 4
		dAtA[i+3] = uint8(uint32(k) >> 24)
		dAtA[i+2] = uint8(uint32(k) >> 16)
		dAtA[i+1] = uint8(uint32(k) >> 8)
		dAtA[i+0] = uint8(uint32(k))
		i--
		dAtA[i] = 13 /*(1<<3)|5=13, field=1, wireType=32bit*/
		{{- else if eq .MapKey "fixed64"}}
		i -= 8
		dAtA[i+7] = uint8(uint64(k) >> 56)
		dAtA[i+6] = uint8(uint64(k) >> 48)
		dAtA[i+5] = uint8(uint64(k) >> 40)
		dAtA[i+4] = uint8(uint64(k) >> 32)
		dAtA[i+3] = uint8(uint64(k) >> 24)
		dAtA[i+2] = uint8(uint64(k) >> 16)
		dAtA[i+1] = uint8(uint64(k) >> 8)
		dAtA[i+0] = uint8(uint64(k))
		i--
		dAtA[i] = 9 /*(1<<3)|1=9, field=1, wireType=64bit*/
		{{- else if eq .MapKey "sfixed64"}}
		i -= 8
		dAtA[i+7] = uint8(uint64(k) >> 56)
		dAtA[i+6] = uint8(uint64(k) >> 48)
		dAtA[i+5] = uint8(uint64(k) >> 40)
		dAtA[i+4] = uint8(uint64(k) >> 32)
		dAtA[i+3] = uint8(uint64(k) >> 24)
		dAtA[i+2] = uint8(uint64(k) >> 16)
		dAtA[i+1] = uint8(uint64(k) >> 8)
		dAtA[i+0] = uint8(uint64(k))
		i--
		dAtA[i] = 9 /*(1<<3)|1=9, field=1, wireType=64bit*/
		{{- else if eq .MapKey "sint32"}}
		i = utils.EncodeVarint(dAtA, i, uint64((uint32(k)<<1)^uint32(k>>31)))
		i--
		dAtA[i] = 8 /*(1<<3)|0=8, field=1, wireType=Varint*/
		{{- else if eq .MapKey "sint64"}}
		i = utils.EncodeVarint(dAtA, i, (uint64(k)<<1)^uint64(k>>63))
		i--
		dAtA[i] = 8 /*(1<<3)|0=8, field=1, wireType=Varint*/
		{{- else}}
		i = utils.EncodeVarint(dAtA, i, uint64(k))
		i--
		dAtA[i] = 8 /*(1<<3)|0=8, field=1, wireType=Varint*/
		{{- end}}
		// write outer map entry length + field tag
		i = utils.EncodeVarint(dAtA, i, uint64(baseI-i))
		{{writeTagBytes .Number 2}}
	}
{{- else if .Repeated}}
{{- if isPackable .Type}}
	if len(m.{{.Name}}) > 0 {
		pksize := 0
		for _, v := range m.{{.Name}} {
			{{- if eq .Type "sint32"}}
			pksize += ((bits.Len64(uint64((uint32(v)<<1)^uint32(v>>31))|1) + 6) / 7)
			{{- else if eq .Type "sint64"}}
			pksize += ((bits.Len64((uint64(v)<<1)^uint64(v>>63)|1) + 6) / 7)
			{{- else if eq .Type "fixed32"}}
			_ = v; pksize += 4
			{{- else if eq .Type "fixed64"}}
			_ = v; pksize += 8
			{{- else if eq .Type "sfixed32"}}
			_ = v; pksize += 4
			{{- else if eq .Type "sfixed64"}}
			_ = v; pksize += 8
			{{- else if eq .Type "float"}}
			_ = v; pksize += 4
			{{- else if eq .Type "double"}}
			_ = v; pksize += 8
			{{- else if eq .Type "bool"}}
			_ = v; pksize += 1
			{{- else}}
			pksize += ((bits.Len64(uint64(v)|1) + 6) / 7)
			{{- end}}
		}
		i -= pksize
		j := i
		for _, v := range m.{{.Name}} {
			{{- if eq .Type "sint32"}}
			uv := uint64((uint32(v)<<1)^uint32(v>>31))
			for uv >= 0x80 { dAtA[j] = byte(uv)|0x80; uv >>= 7; j++ }
			dAtA[j] = byte(uv); j++
			{{- else if eq .Type "sint64"}}
			uv := (uint64(v)<<1)^uint64(v>>63)
			for uv >= 0x80 { dAtA[j] = byte(uv)|0x80; uv >>= 7; j++ }
			dAtA[j] = byte(uv); j++
			{{- else if eq .Type "fixed32"}}
			dAtA[j+3] = uint8(uint32(v) >> 24)
			dAtA[j+2] = uint8(uint32(v) >> 16)
			dAtA[j+1] = uint8(uint32(v) >> 8)
			dAtA[j+0] = uint8(uint32(v))
			j += 4
			{{- else if eq .Type "fixed64"}}
			dAtA[j+7] = uint8(uint64(v) >> 56)
			dAtA[j+6] = uint8(uint64(v) >> 48)
			dAtA[j+5] = uint8(uint64(v) >> 40)
			dAtA[j+4] = uint8(uint64(v) >> 32)
			dAtA[j+3] = uint8(uint64(v) >> 24)
			dAtA[j+2] = uint8(uint64(v) >> 16)
			dAtA[j+1] = uint8(uint64(v) >> 8)
			dAtA[j+0] = uint8(uint64(v))
			j += 8
			{{- else if eq .Type "sfixed32"}}
			dAtA[j+3] = uint8(uint32(v) >> 24)
			dAtA[j+2] = uint8(uint32(v) >> 16)
			dAtA[j+1] = uint8(uint32(v) >> 8)
			dAtA[j+0] = uint8(uint32(v))
			j += 4
			{{- else if eq .Type "sfixed64"}}
			dAtA[j+7] = uint8(uint64(v) >> 56)
			dAtA[j+6] = uint8(uint64(v) >> 48)
			dAtA[j+5] = uint8(uint64(v) >> 40)
			dAtA[j+4] = uint8(uint64(v) >> 32)
			dAtA[j+3] = uint8(uint64(v) >> 24)
			dAtA[j+2] = uint8(uint64(v) >> 16)
			dAtA[j+1] = uint8(uint64(v) >> 8)
			dAtA[j+0] = uint8(uint64(v))
			j += 8
			{{- else if eq .Type "float"}}
			_fv := math.Float32bits(v)
			dAtA[j+3] = uint8(_fv >> 24)
			dAtA[j+2] = uint8(_fv >> 16)
			dAtA[j+1] = uint8(_fv >> 8)
			dAtA[j+0] = uint8(_fv)
			j += 4
			{{- else if eq .Type "double"}}
			_dv := math.Float64bits(v)
			dAtA[j+7] = uint8(_dv >> 56)
			dAtA[j+6] = uint8(_dv >> 48)
			dAtA[j+5] = uint8(_dv >> 40)
			dAtA[j+4] = uint8(_dv >> 32)
			dAtA[j+3] = uint8(_dv >> 24)
			dAtA[j+2] = uint8(_dv >> 16)
			dAtA[j+1] = uint8(_dv >> 8)
			dAtA[j+0] = uint8(_dv)
			j += 8
			{{- else if eq .Type "bool"}}
			if v { dAtA[j] = 1 } else { dAtA[j] = 0 }; j++
			{{- else}}
			uv := uint64(v)
			for uv >= 0x80 { dAtA[j] = byte(uv)|0x80; uv >>= 7; j++ }
			dAtA[j] = byte(uv); j++
			{{- end}}
		}
		i = utils.EncodeVarint(dAtA, i, uint64(pksize))
		{{writeTagBytes .Number 2}}
	}
{{- else if eq .Type "string"}}
	for iNdEx := len(m.{{.Name}}) - 1; iNdEx >= 0; iNdEx-- {
		s := m.{{.Name}}[iNdEx]
		i -= len(s)
		copy(dAtA[i:], s)
		i = utils.EncodeVarint(dAtA, i, uint64(len(s)))
		{{writeTagBytes .Number 2}}
	}
{{- else if eq .Type "bytes"}}
	for iNdEx := len(m.{{.Name}}) - 1; iNdEx >= 0; iNdEx-- {
		b := m.{{.Name}}[iNdEx]
		i -= len(b)
		copy(dAtA[i:], b)
		i = utils.EncodeVarint(dAtA, i, uint64(len(b)))
		{{writeTagBytes .Number 2}}
	}
{{- else}}
	for iNdEx := len(m.{{.Name}}) - 1; iNdEx >= 0; iNdEx-- {
		size := m.{{.Name}}[iNdEx].marshalToSizedBufferVT(dAtA[:i])
		i -= size
		i = utils.EncodeVarint(dAtA, i, uint64(size))
		{{writeTagBytes .Number 2}}
	}
{{- end}}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if m.{{.Name}} != nil {
		size := m.{{.Name}}.marshalToSizedBufferVT(dAtA[:i])
		i -= size
		i = utils.EncodeVarint(dAtA, i, uint64(size))
		{{writeTagBytes .Number 2}}
	}
{{- else}}
	{
		size := m.{{.Name}}.marshalToSizedBufferVT(dAtA[:i])
		if size > 0 {
			i -= size
			i = utils.EncodeVarint(dAtA, i, uint64(size))
			{{writeTagBytes .Number 2}}
		}
	}
{{- end}}
{{- else if .IsDecimal}}
	if !m.{{.Name}}.IsZero() {
		_dvdec{{.Name}}, _ := m.{{.Name}}.Round({{.DecimalRound}}).Float64()
		i -= 8
		_dv := math.Float64bits(_dvdec{{.Name}})
		dAtA[i+7] = uint8(_dv >> 56)
		dAtA[i+6] = uint8(_dv >> 48)
		dAtA[i+5] = uint8(_dv >> 40)
		dAtA[i+4] = uint8(_dv >> 32)
		dAtA[i+3] = uint8(_dv >> 24)
		dAtA[i+2] = uint8(_dv >> 16)
		dAtA[i+1] = uint8(_dv >> 8)
		dAtA[i+0] = uint8(_dv)
		{{writeTagBytes .Number 1}}
	}
{{- else if eq .Type "double"}}
	if m.{{.Name}} != 0 {
		i -= 8
		_dv := math.Float64bits(m.{{.Name}})
		dAtA[i+7] = uint8(_dv >> 56)
		dAtA[i+6] = uint8(_dv >> 48)
		dAtA[i+5] = uint8(_dv >> 40)
		dAtA[i+4] = uint8(_dv >> 32)
		dAtA[i+3] = uint8(_dv >> 24)
		dAtA[i+2] = uint8(_dv >> 16)
		dAtA[i+1] = uint8(_dv >> 8)
		dAtA[i+0] = uint8(_dv)
		{{writeTagBytes .Number 1}}
	}
{{- else if eq .Type "float"}}
	if m.{{.Name}} != 0 {
		i -= 4
		_fv := math.Float32bits(m.{{.Name}})
		dAtA[i+3] = uint8(_fv >> 24)
		dAtA[i+2] = uint8(_fv >> 16)
		dAtA[i+1] = uint8(_fv >> 8)
		dAtA[i+0] = uint8(_fv)
		{{writeTagBytes .Number 5}}
	}
{{- else if eq .Type "fixed32"}}
	if m.{{.Name}} != 0 {
		i -= 4
		dAtA[i+3] = uint8(m.{{.Name}} >> 24)
		dAtA[i+2] = uint8(m.{{.Name}} >> 16)
		dAtA[i+1] = uint8(m.{{.Name}} >> 8)
		dAtA[i+0] = uint8(m.{{.Name}})
		{{writeTagBytes .Number 5}}
	}
{{- else if eq .Type "fixed64"}}
	if m.{{.Name}} != 0 {
		i -= 8
		dAtA[i+7] = uint8(m.{{.Name}} >> 56)
		dAtA[i+6] = uint8(m.{{.Name}} >> 48)
		dAtA[i+5] = uint8(m.{{.Name}} >> 40)
		dAtA[i+4] = uint8(m.{{.Name}} >> 32)
		dAtA[i+3] = uint8(m.{{.Name}} >> 24)
		dAtA[i+2] = uint8(m.{{.Name}} >> 16)
		dAtA[i+1] = uint8(m.{{.Name}} >> 8)
		dAtA[i+0] = uint8(m.{{.Name}})
		{{writeTagBytes .Number 1}}
	}
{{- else if eq .Type "sfixed32"}}
	if m.{{.Name}} != 0 {
		i -= 4
		dAtA[i+3] = uint8(uint32(m.{{.Name}}) >> 24)
		dAtA[i+2] = uint8(uint32(m.{{.Name}}) >> 16)
		dAtA[i+1] = uint8(uint32(m.{{.Name}}) >> 8)
		dAtA[i+0] = uint8(uint32(m.{{.Name}}))
		{{writeTagBytes .Number 5}}
	}
{{- else if eq .Type "sfixed64"}}
	if m.{{.Name}} != 0 {
		i -= 8
		dAtA[i+7] = uint8(uint64(m.{{.Name}}) >> 56)
		dAtA[i+6] = uint8(uint64(m.{{.Name}}) >> 48)
		dAtA[i+5] = uint8(uint64(m.{{.Name}}) >> 40)
		dAtA[i+4] = uint8(uint64(m.{{.Name}}) >> 32)
		dAtA[i+3] = uint8(uint64(m.{{.Name}}) >> 24)
		dAtA[i+2] = uint8(uint64(m.{{.Name}}) >> 16)
		dAtA[i+1] = uint8(uint64(m.{{.Name}}) >> 8)
		dAtA[i+0] = uint8(uint64(m.{{.Name}}))
		{{writeTagBytes .Number 1}}
	}
{{- else if eq .Type "sint32"}}
	if m.{{.Name}} != 0 {
		i = utils.EncodeVarint(dAtA, i, uint64((uint32(m.{{.Name}})<<1)^uint32(m.{{.Name}}>>31)))
		{{writeTagBytes .Number 0}}
	}
{{- else if eq .Type "sint64"}}
	if m.{{.Name}} != 0 {
		i = utils.EncodeVarint(dAtA, i, (uint64(m.{{.Name}})<<1)^uint64(m.{{.Name}}>>63))
		{{writeTagBytes .Number 0}}
	}
{{- else if eq .Type "bool"}}
	if m.{{.Name}} {
		i--
		if m.{{.Name}} { dAtA[i] = 1 } else { dAtA[i] = 0 }
		{{writeTagBytes .Number 0}}
	}
{{- else if eq .Type "string"}}
	if len(m.{{.Name}}) > 0 {
		i -= len(m.{{.Name}})
		copy(dAtA[i:], m.{{.Name}})
		i = utils.EncodeVarint(dAtA, i, uint64(len(m.{{.Name}})))
		{{writeTagBytes .Number 2}}
	}
{{- else if eq .Type "bytes"}}
	if len(m.{{.Name}}) > 0 {
		i -= len(m.{{.Name}})
		copy(dAtA[i:], m.{{.Name}})
		i = utils.EncodeVarint(dAtA, i, uint64(len(m.{{.Name}})))
		{{writeTagBytes .Number 2}}
	}
{{- else}}
	if m.{{.Name}} != 0 {
		i = utils.EncodeVarint(dAtA, i, uint64(m.{{.Name}}))
		{{writeTagBytes .Number 0}}
	}
{{- end}}
{{- end}}
	return len(dAtA) - i
}

// ToProtobufVT serializes m into dst, reusing its backing array when cap is sufficient.
// Pass nil to allocate a fresh buffer. Equivalent to vtprotobuf MarshalVT().
func (m *{{$goName}}) ToProtobufVT(dst []byte) []byte {
	if m == nil {
		return dst
	}
	// todo: 当存在 map 类型时。这个方法的调用导致了 ToProtobuf() 两次遍历 map，最多导致 33.1% 的性能下降
	size := m.ProtobufSize()
	oldLen := len(dst)
	if cap(dst)-oldLen < size {
		next := make([]byte, oldLen+size)
		copy(next, dst)
		dst = next
	} else {
		dst = dst[:oldLen+size]
	}
	n := m.marshalToSizedBufferVT(dst[oldLen:])
	return dst[:oldLen+n]
}

func (m *{{$goName}}) ToProtobuf(dst []byte) []byte {
{{- if hasMapsOrMsgs .Fields}}
	return m.ToProtobufByAppend(dst)
{{- else}}
	return m.ToProtobufVT(dst)
{{- end}}
}
{{end}}
