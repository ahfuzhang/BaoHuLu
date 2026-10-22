{{with model .}}
{{/* Protobuf 解码：FromProtobuf 和 FromProtobufWithArena。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"fmt"
	"math"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/fastjson"
	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = fmt.Errorf
	_ = math.Float32bits
	_ = fastjson.NewArena
	_ = utils.AppendVarint
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
func (r *Readonly{{$goName}}) FromProtobuf(in []byte) error {
	r.rawBufferLen = len(in)
{{- if hasStringOrBytesFields .Fields}}
	if r.strArena == nil {
		r.strArena = fastjson.NewArena(r.rawBufferLen)
	}
{{- end}}
	var err error
	for len(in) > 0 {
		var fieldNum int
		var wt utils.WireType
		fieldNum, wt, in, err = utils.ConsumeTag(in)
		if err != nil {
			return err
		}
		switch fieldNum {
{{- range .Fields}}
		case {{$goName}}{{.Name}}Tag: // {{.Name}}
			{{- if not (or (eq .Type "int32") (eq .Type "uint32") (eq .Type "sint32") (eq .Type "int64") (eq .Type "uint64") (eq .Type "sint64"))}}
			if {{if or .Map .Repeated .IsMsg (eq .Type "string") (eq .Type "bytes")}}wt != utils.WireTypeLenDelim{{else if or (eq .Type "double") (eq .Type "fixed64") (eq .Type "sfixed64")}}wt != utils.WireType64bit{{else if or (eq .Type "float") (eq .Type "fixed32") (eq .Type "sfixed32")}}wt != utils.WireType32bit{{else}}wt != utils.WireTypeVarint{{end}} {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			{{- end}}
{{- if .Map}}
			var entryData []byte
			entryData, in, err = utils.ConsumeBytes(in)
			if err != nil {
				return err
			}
			// todo: 提前 count 元素个数
			var mKey {{mapKeyGoType .MapKey}}
			{{- if mapValIsMsg .MapVal}}
			if cap(r._{{.Name}}Arr) < 16 {
				r._{{.Name}}Arr = append(make([]{{readonlyTypeName .MapVal}}, 0, 16), r._{{.Name}}Arr...)
			}
			if cap(r._{{.Name}}Arr) == len(r._{{.Name}}Arr) {
				r._{{.Name}}Arr = append(r._{{.Name}}Arr, {{readonlyTypeName .MapVal}}{})
			} else {
				r._{{.Name}}Arr = r._{{.Name}}Arr[:len(r._{{.Name}}Arr)+1]
			}
			_mValIdx := len(r._{{.Name}}Arr) - 1
			{{- else}}
			var mVal {{mapValGoType .MapVal}}
			{{- end}}
			for len(entryData) > 0 {
				var efn int
				var ewt utils.WireType
				efn, ewt, entryData, err = utils.ConsumeTag(entryData)
				if err != nil {
					return err
				}
				switch efn {
				case 1:
					mKey, entryData, err = {{readFunc .MapKey}}(entryData)
					{{- if eq .MapKey "string"}}
					mKey = r.strArena.PutString(mKey)  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
					{{- end}}
				case 2:
					{{- if mapValIsMsg .MapVal}}
					var _subBytes []byte
					_subBytes, entryData, err = utils.ConsumeBytes(entryData)
					if err != nil {
						break
					}
					err = r._{{.Name}}Arr[_mValIdx].FromProtobufWithArena(_subBytes, r.strArena)  // reuse parent message's arena
					{{- else if mapValIsEnum .MapVal}}
					{
						var _ev int32
						_ev, entryData, err = utils.ReadInt32(entryData)
						mVal = {{mapValGoType .MapVal}}(_ev)
					}
					{{- else}}
					mVal, entryData, err = {{readFunc .MapVal}}(entryData)
					{{- if eq .MapVal "string"}}
					mVal = r.strArena.PutString(mVal)  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
					{{- else if eq .MapVal "bytes"}}
					mVal = r.strArena.PutBytes(mVal)  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
					{{- end}}
					{{- end}}
				default:
					entryData, err = utils.SkipField(ewt, entryData)
				}
				if err != nil {
					return err
				}
			}
			if r.{{.Name}} == nil {
				{{- if mapValIsMsg .MapVal}}
				r.{{.Name}} = make(map[{{mapKeyGoType .MapKey}}]*{{readonlyTypeName .MapVal}}, 64)
				{{- else}}
				r.{{.Name}} = make({{.ReaderType}}, 64)
				{{- end}}
			}
			{{- if mapValIsMsg .MapVal}}
			r.{{.Name}}[mKey] = &r._{{.Name}}Arr[_mValIdx]
			{{- else}}
			r.{{.Name}}[mKey] = mVal
			{{- end}}
{{- else if .Repeated}}
            if cap(r.{{.Name}}) < 64 {
				temp := make({{.ReaderType}}, 0, 64)  // 一开始就分配一个大一点内存，避免频繁的对 slice 扩容
				r.{{.Name}} = append(temp, r.{{.Name}}...)
			}
{{- if isPackable .Type}}
			var packedData []byte
			packedData, in, err = utils.ConsumeBytes(in)
			if err != nil {
				return err
			}
			for len(packedData) > 0 {
				var pv {{mapKeyGoType .Type}}
				pv, packedData, err = {{readFunc .Type}}(packedData)
				if err != nil {
					return err
				}
				r.{{.Name}} = append(r.{{.Name}}, pv)
			}
{{- else if eq .Type "string"}}
			var sv string
			sv, in, err = utils.ReadString(in)
			if err != nil {
				return err
			}
			sv = r.strArena.PutString(sv)  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
			r.{{.Name}} = append(r.{{.Name}}, sv)
{{- else if eq .Type "bytes"}}
			var bv []byte
			bv, in, err = utils.ReadBytes(in)
			if err != nil {
				return err
			}
			bv = r.strArena.PutBytes(bv)  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
			r.{{.Name}} = append(r.{{.Name}}, bv)
{{- else}}
			var subData []byte
			subData, in, err = utils.ConsumeBytes(in)
			if err != nil {
				return err
			}
			var elem {{readerElemType .}}
			if err = elem.FromProtobufWithArena(subData, r.strArena); err != nil {
				return err
			}
			r.{{.Name}} = append(r.{{.Name}}, elem)
{{- end}}
{{- else if .IsMsg}}
			var subData []byte
			subData, in, err = utils.ConsumeBytes(in)
			if err != nil {
				return err
			}
{{- if .IsRecursive}}
			if r.{{.Name}} == nil {
				r.{{.Name}} = &{{.ReaderType}}{}
			}
			r._has{{.Name}} = true
{{- end}}
			if err = r.{{.Name}}.FromProtobufWithArena(subData, r.strArena); err != nil {
				return err
			}
{{- else if .IsDecimal}}
			var _rawDouble{{.Name}} float64
			_rawDouble{{.Name}}, in, err = utils.ReadDouble(in)
			if err != nil {
				return err
			}
			r.{{.Name}} = decimal.MustNew(int64(math.Round(_rawDouble{{.Name}} * math.Pow10({{.DecimalRound}}))), {{.DecimalRound}})
{{- else if eq .Type "double"}}
			r.{{.Name}}, in, err = utils.ReadDouble(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "float"}}
			r.{{.Name}}, in, err = utils.ReadFloat(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "fixed32"}}
			r.{{.Name}}, in, err = utils.ReadFixed32(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "fixed64"}}
			r.{{.Name}}, in, err = utils.ReadFixed64(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "sfixed32"}}
			r.{{.Name}}, in, err = utils.ReadSfixed32(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "sfixed64"}}
			r.{{.Name}}, in, err = utils.ReadSfixed64(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "sint32"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadSint32(in)
			} else if wt == utils.WireType32bit {
				// 此处只是兼容性的解码逻辑，没有风险。
				// 注意： RewriteDatatype 功能，可能导致二进制格式无法兼容。别的工具生成的代码，无法被解码
				var _fv uint32
				_fv, in, err = utils.ReadFixed32(in)
				r.{{.Name}} = int32(_fv)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "sint64"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadSint64(in)
			} else if wt == utils.WireType64bit {
				var _fv uint64
				_fv, in, err = utils.ReadFixed64(in)
				r.{{.Name}} = int64(_fv)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "int32"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadInt32(in)
			} else if wt == utils.WireType32bit {
				var _fv uint32
				_fv, in, err = utils.ReadFixed32(in)
				r.{{.Name}} = int32(_fv)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "int64"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadInt64(in)
			} else if wt == utils.WireType64bit {
				var _fv uint64
				_fv, in, err = utils.ReadFixed64(in)
				r.{{.Name}} = int64(_fv)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "uint32"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadUint32(in)
			} else if wt == utils.WireType32bit {
				r.{{.Name}}, in, err = utils.ReadFixed32(in)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "uint64"}}
			if wt == utils.WireTypeVarint {
				r.{{.Name}}, in, err = utils.ReadUint64(in)
			} else if wt == utils.WireType64bit {
				r.{{.Name}}, in, err = utils.ReadFixed64(in)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field {{.Name}}", wt)
			}
			if err != nil {
				return err
			}
{{- else if eq .Type "bool"}}
			r.{{.Name}}, in, err = utils.ReadBool(in)
			if err != nil {
				return err
			}
{{- else if eq .Type "string"}}
			r.{{.Name}}, in, err = utils.ReadString(in)
			if err != nil {
				return err
			}
			r.{{.Name}} = r.strArena.PutString(r.{{.Name}})  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
{{- else if eq .Type "bytes"}}
			r.{{.Name}}, in, err = utils.ReadBytes(in)
			if err != nil {
				return err
			}
			r.{{.Name}} = r.strArena.PutBytes(r.{{.Name}})  // The string must be copied; otherwise, Reset() may cause the string content to be set to \0.
{{- else}}
			var ev int32
			ev, in, err = utils.ReadInt32(in)
			if err != nil {
				return err
			}
			r.{{.Name}} = {{.ReaderType}}(ev)
{{- end}}
{{- end}}
		default:
			var err2 error
			in, err2 = utils.SkipField(wt, in)
			if err2 != nil {
				return err2
			}
		}
	}
	return nil
}
func (r *Readonly{{$goName}}) FromProtobufWithArena(in []byte, arena *fastjson.Arena) error {
	r.strArena = arena
	return r.FromProtobuf(in)
}
{{end}}
{{end}}
