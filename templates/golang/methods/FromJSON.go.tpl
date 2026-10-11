{{/* JSON 解码：FromJSON、fromJSONValue 和 fromJSONArray。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/base64"
	"math"
	"unsafe"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/fastfloat"
	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/fastjson"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = base64.StdEncoding
	_ = math.Float32bits
	_ = unsafe.Pointer(nil)
	_ = fastfloat.ParseBestEffort
	_ = fastjson.NewArena
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
{{- if .AsMap}}
{{- $f := index .Fields 0}}
func (r *Readonly{{$goName}}) fromJSONValue(obj *fastjson.Object, parser *fastjson.Parser, arena *fastjson.Arena) error {
	if r.strArena == nil && arena != nil {
		r.strArena = arena
	}
	var visitErr error
	if r.{{$f.Name}} == nil {
		{{- if mapValIsMsg $f.MapVal}}
		r.{{$f.Name}} = make(map[{{mapKeyGoType $f.MapKey}}]*{{readonlyTypeName $f.MapVal}}, obj.Len())
		{{- else}}
		r.{{$f.Name}} = make({{$f.ReaderType}}, obj.Len())
		{{- end}}
	}
{{- if mapValIsMsg $f.MapVal}}
	if r._{{$f.Name}}Arr == nil {
		r._{{$f.Name}}Arr = make([]{{readonlyTypeName $f.MapVal}}, 0, obj.Len())
	} else {
		if cap(r._{{$f.Name}}Arr) >= obj.Len() {
			r._{{$f.Name}}Arr = r._{{$f.Name}}Arr[:0]
		} else {
			r._{{$f.Name}}Arr = make([]{{readonlyTypeName $f.MapVal}}, 0, obj.Len())
		}
	}
{{- end}}
	obj.Visit(func(mk []byte, mv *fastjson.Value) (isContinue bool) {
		var mKey {{mapKeyGoType $f.MapKey}}
{{- $kc := jsonMapKeyClass $f.MapKey}}
{{- if eq $kc "string"}}
		mKey = unsafe.String(unsafe.SliceData(mk), len(mk))
		mKey = r.strArena.PutString(mKey)  // copy string
{{- else if eq $kc "bool"}}
		mKey = unsafe.String(unsafe.SliceData(mk), len(mk)) == "true"
{{- else if eq $kc "signed32"}}
		_mk64, _ek := fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(mk), len(mk)))
		if _ek != nil {
			visitErr = _ek
			return
		}
		mKey = {{mapKeyGoType $f.MapKey}}(_mk64)
{{- else if eq $kc "signed64"}}
		_mk64, _ek := fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(mk), len(mk)))
		if _ek != nil {
			visitErr = _ek
			return
		}
		mKey = _mk64
{{- else if eq $kc "unsigned32"}}
		_mku64, _ek := fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(mk), len(mk)))
		if _ek != nil {
			visitErr = _ek
			return
		}
		mKey = {{mapKeyGoType $f.MapKey}}(_mku64)
{{- else if eq $kc "unsigned64"}}
		_mku64, _ek := fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(mk), len(mk)))
		if _ek != nil {
			visitErr = _ek
			return
		}
		mKey = _mku64
{{- end}}
{{- if mapValIsMsg $f.MapVal}}
		r._{{$f.Name}}Arr = r._{{$f.Name}}Arr[:len(r._{{$f.Name}}Arr)+1]
		_mValIdx := len(r._{{$f.Name}}Arr) - 1
		sub := &r._{{$f.Name}}Arr[_mValIdx]
{{- if $f.MapValMsgIsAsArray}}
		_subArr, _eo := mv.Array()
		if _eo != nil {
			visitErr = _eo
			return
		}
		if _eo2 := sub.fromJSONArray(_subArr, parser, r.strArena); _eo2 != nil {
			visitErr = _eo2
			return
		}
{{- else}}
		_subObj, _eo := mv.Object()
		if _eo != nil {
			visitErr = _eo
			return
		}
		if _eo2 := sub.fromJSONValue(_subObj, parser, r.strArena); _eo2 != nil {
			visitErr = _eo2
			return
		}
{{- end}}
		r.{{$f.Name}}[mKey] = sub
{{- else}}
		var mVal {{mapValGoType $f.MapVal}}
{{- $vc := jsonScalarClass $f.MapVal}}
{{- if eq $vc "string"}}
		_s, _ev := mv.UnescapeString(parser)  // unescape and clone
		if _ev != nil {
			visitErr = _ev
			return
		}
		mVal = _s
{{- else if eq $vc "bytes"}}
		_b64, _ev := mv.StringBytes()
		if _ev != nil {
			visitErr = _ev
			return
		}
		var _ev2 error
		mVal = r.strArena.Reserve(len(_b64))
		mVal, _ev2 = base64.StdEncoding.AppendDecode(mVal[:0], _b64)
		if _ev2 != nil {
			visitErr = _ev2
			return
		}
{{- else if eq $vc "bool"}}
		_bv, _ev := mv.Bool()
		if _ev != nil {
			visitErr = _ev
			return
		}
		mVal = _bv
{{- else if eq $vc "float"}}
		var _fv float64
		if mv.Type(parser) == fastjson.TypeString {
			_sb, _ev := mv.StringBytes()
			if _ev != nil {
				visitErr = _ev
				return
			}
			var _ev2 error
			_fv, _ev2 = fastfloat.Parse(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ev2 != nil {
				visitErr = _ev2
				return
			}
		} else {
			var _ev error
			_fv, _ev = mv.Float64()
			if _ev != nil {
				visitErr = _ev
				return
			}
		}
		mVal = {{mapValGoType $f.MapVal}}(_fv)
{{- else if eq $vc "signed"}}
		_iv, _ev := mv.Int64()
		if _ev != nil {
			visitErr = _ev
			return
		}
		mVal = {{mapValGoType $f.MapVal}}(_iv)
{{- else if eq $vc "signed64"}}
		var _iv int64
		if mv.Type(parser) == fastjson.TypeString {
			_sb, _ev := mv.StringBytes()
			if _ev != nil {
				visitErr = _ev
				return
			}
			var _ev2 error
			_iv, _ev2 = fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ev2 != nil {
				visitErr = _ev2
				return
			}
		} else {
			var _ev error
			_iv, _ev = mv.Int64()
			if _ev != nil {
				visitErr = _ev
				return
			}
		}
		mVal = {{mapValGoType $f.MapVal}}(_iv)
{{- else if eq $vc "unsigned"}}
		_uv, _ev := mv.Uint64()
		if _ev != nil {
			visitErr = _ev
			return
		}
		mVal = {{mapValGoType $f.MapVal}}(_uv)
{{- else if eq $vc "unsigned64"}}
		var _uv uint64
		if mv.Type(parser) == fastjson.TypeString {
			_sb, _ev := mv.StringBytes()
			if _ev != nil {
				visitErr = _ev
				return
			}
			var _ev2 error
			_uv, _ev2 = fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ev2 != nil {
				visitErr = _ev2
				return
			}
		} else {
			var _ev error
			_uv, _ev = mv.Uint64()
			if _ev != nil {
				visitErr = _ev
				return
			}
		}
		mVal = {{mapValGoType $f.MapVal}}(_uv)
{{- end}}
		r.{{$f.Name}}[mKey] = mVal
{{- end}}
		isContinue = true
		return
	}, parser, {{if eq $f.MapKey "string"}}false{{else}}true{{end}})
	return visitErr
}
{{- else if .AsArray}}
{{- $f := index .Fields 0}}
func (r *Readonly{{$goName}}) fromJSONArray(arr []*fastjson.Value, parser *fastjson.Parser, arena *fastjson.Arena) error {
	if r.strArena == nil && arena != nil {
		r.strArena = arena
	}
	if cap(r.{{$f.Name}}) < len(arr) {
		r.{{$f.Name}} = make({{$f.ReaderType}}, 0, len(arr))
	} else {
		r.{{$f.Name}} = r.{{$f.Name}}[:0]
	}
{{- if $f.IsMsg}}
	for _, _item := range arr {
		var _elem {{elemType $f.ReaderType}}
{{- if $f.MsgIsAsArray}}
		_subArr, _eo := _item.Array()
		if _eo != nil {
			return _eo
		}
		if _eo2 := _elem.fromJSONArray(_subArr, parser, r.strArena); _eo2 != nil {
			return _eo2
		}
{{- else}}
		_subObj, _eo := _item.Object()
		if _eo != nil {
			return _eo
		}
		if _eo2 := _elem.fromJSONValue(_subObj, parser, r.strArena); _eo2 != nil {
			return _eo2
		}
{{- end}}
		r.{{$f.Name}} = append(r.{{$f.Name}}, _elem)
	}
{{- else if $f.IsEnum}}
	for _, _item := range arr {
		_iv, _ei := _item.Int64()
		if _ei != nil {
			return _ei
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_iv))
	}
{{- else}}
{{- $sc := jsonScalarClass $f.Type}}
	for _, _item := range arr {
{{- if eq $sc "string"}}
		_s, _ei := _item.UnescapeString(parser)  // Unescape and clone
		if _ei != nil {
			return _ei
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, _s)
{{- else if eq $sc "bytes"}}
		_b64, _ei := _item.StringBytes()
		if _ei != nil {
			return _ei
		}
		var _dec []byte
		var _ei2 error
		_dec = r.strArena.Reserve(len(_b64))
		_dec, _ei2 = base64.StdEncoding.AppendDecode(_dec[:0], _b64)
		if _ei2 != nil {
			return _ei2
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, _dec)
{{- else if eq $sc "bool"}}
		_bv, _ei := _item.Bool()
		if _ei != nil {
			return _ei
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, _bv)
{{- else if eq $sc "float"}}
		var _fv float64
		if _item.Type(parser) == fastjson.TypeString {
			_sb, _ei := _item.StringBytes()
			if _ei != nil {
				return _ei
			}
			var _ei2 error
			_fv, _ei2 = fastfloat.Parse(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ei2 != nil {
				return _ei2
			}
		} else {
			var _ei error
			_fv, _ei = _item.Float64()
			if _ei != nil {
				return _ei
			}
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_fv))
{{- else if eq $sc "signed"}}
		_iv, _ei := _item.Int64()
		if _ei != nil {
			return _ei
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_iv))
{{- else if eq $sc "signed64"}}
		var _iv int64
		if _item.Type(parser) == fastjson.TypeString {
			_sb, _ei := _item.StringBytes()
			if _ei != nil {
				return _ei
			}
			var _ei2 error
			_iv, _ei2 = fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ei2 != nil {
				return _ei2
			}
		} else {
			var _ei error
			_iv, _ei = _item.Int64()
			if _ei != nil {
				return _ei
			}
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_iv))
{{- else if eq $sc "unsigned"}}
		_uv, _ei := _item.Uint64()
		if _ei != nil {
			return _ei
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_uv))
{{- else if eq $sc "unsigned64"}}
		var _uv uint64
		if _item.Type(parser) == fastjson.TypeString {
			_sb, _ei := _item.StringBytes()
			if _ei != nil {
				return _ei
			}
			var _ei2 error
			_uv, _ei2 = fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
			if _ei2 != nil {
				return _ei2
			}
		} else {
			var _ei error
			_uv, _ei = _item.Uint64()
			if _ei != nil {
				return _ei
			}
		}
		r.{{$f.Name}} = append(r.{{$f.Name}}, {{elemType $f.ReaderType}}(_uv))
{{- end}}
	}
{{- end}}
	return nil
}
{{- else}}
func (r *Readonly{{$goName}}) fromJSONValue(obj *fastjson.Object, parser *fastjson.Parser, arena *fastjson.Arena) error {
	if r.strArena == nil && arena != nil {
		r.strArena = arena
	}
	var visitErr error
	obj.Visit(func(k []byte, v *fastjson.Value) (isContinue bool) {
		k1 := unsafe.String(unsafe.SliceData(k), len(k))
		switch k1 {
{{- range .Fields}}
		case NameOf{{$goName}}{{.Name}}:
{{- $nullSc := jsonScalarClass .Type}}
{{- $isPlainStrBytes := and (not .Map) (not .Repeated) (not .IsMsg) (not .IsEnum) (not .IsDecimal) (or (eq $nullSc "string") (eq $nullSc "bytes"))}}
{{- if not $isPlainStrBytes}}
			// a JSON null value means the field is absent; keep its zero value and skip.
			// (Type() is side-effect-free here since the value is object/array/number/bool/null, never a raw string.)
			if v.Type(parser) == fastjson.TypeNull {
				isContinue = true
				return
			}
{{- end}}
{{- if .Map}}
			_mapObj, _e := v.Object()
			if _e != nil {
				visitErr = _e
				return
			}
			if r.{{.Name}} == nil {
				{{- if mapValIsMsg .MapVal}}
				r.{{.Name}} = make(map[{{mapKeyGoType .MapKey}}]*{{readonlyTypeName .MapVal}}, _mapObj.Len())
				{{- else}}
				r.{{.Name}} = make({{.ReaderType}}, _mapObj.Len())
				{{- end}}
			}
{{- if mapValIsMsg .MapVal}}
			if r._{{.Name}}Arr == nil {
				r._{{.Name}}Arr = make([]{{readonlyTypeName .MapVal}}, 0, _mapObj.Len())
			} else {
				if cap(r._{{.Name}}Arr) >= _mapObj.Len() {
					r._{{.Name}}Arr = r._{{.Name}}Arr[:0]
				} else {
					r._{{.Name}}Arr = make([]{{readonlyTypeName .MapVal}}, 0, _mapObj.Len())
				}
			}
{{- end}}
			_mapObj.Visit(func(mk []byte, mv *fastjson.Value) (isContinue bool) {
				var mKey {{mapKeyGoType .MapKey}}
{{- $kc := jsonMapKeyClass .MapKey}}
{{- if eq $kc "string"}}
				mKey = unsafe.String(unsafe.SliceData(mk), len(mk))
				mKey = r.strArena.PutString(mKey)  // copy string  // todo: 如果 key 存在转义字符，这时候就会导致重复拷贝
{{- else if eq $kc "bool"}}
				mKey = unsafe.String(unsafe.SliceData(mk), len(mk)) == "true"
{{- else if eq $kc "signed32"}}
				_mk64, _ek := fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(mk), len(mk)))
				if _ek != nil {
					visitErr = _ek
					return
				}
				mKey = {{mapKeyGoType .MapKey}}(_mk64)
{{- else if eq $kc "signed64"}}
				_mk64, _ek := fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(mk), len(mk)))
				if _ek != nil {
					visitErr = _ek
					return
				}
				mKey = _mk64
{{- else if eq $kc "unsigned32"}}
				_mku64, _ek := fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(mk), len(mk)))
				if _ek != nil {
					visitErr = _ek
					return
				}
				mKey = {{mapKeyGoType .MapKey}}(_mku64)
{{- else if eq $kc "unsigned64"}}
				_mku64, _ek := fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(mk), len(mk)))
				if _ek != nil {
					visitErr = _ek
					return
				}
				mKey = _mku64
{{- end}}
{{- if mapValIsMsg .MapVal}}
				r._{{.Name}}Arr = r._{{.Name}}Arr[:len(r._{{.Name}}Arr)+1]
				_mValIdx := len(r._{{.Name}}Arr) - 1
				sub := &r._{{.Name}}Arr[_mValIdx]
{{- if .MapValMsgIsAsArray}}
				_subArr, _eo := mv.Array()
				if _eo != nil {
					visitErr = _eo
					return
				}
				if _eo2 := sub.fromJSONArray(_subArr, parser, r.strArena); _eo2 != nil {
					visitErr = _eo2
					return
				}
{{- else}}
				_subObj, _eo := mv.Object()
				if _eo != nil {
					visitErr = _eo
					return
				}
				if _eo2 := sub.fromJSONValue(_subObj, parser, r.strArena); _eo2 != nil {
					visitErr = _eo2
					return
				}
{{- end}}
				r.{{.Name}}[mKey] = sub
{{- else}}
				var mVal {{mapValGoType .MapVal}}
{{- $vc := jsonScalarClass .MapVal}}
{{- if eq $vc "string"}}
				_s, _ev := mv.UnescapeString(parser)  // unescape and clone
				if _ev != nil {
					visitErr = _ev
					return
				}
				mVal = _s
{{- else if eq $vc "bytes"}}
				_b64, _ev := mv.StringBytes()
				if _ev != nil {
					visitErr = _ev
					return
				}
				var _ev2 error
				mVal = r.strArena.Reserve(len(_b64))
				mVal, _ev2 = base64.StdEncoding.AppendDecode(mVal[:0], _b64)
				if _ev2 != nil {
					visitErr = _ev2
					return
				}
{{- else if eq $vc "bool"}}
				_bv, _ev := mv.Bool()
				if _ev != nil {
					visitErr = _ev
					return
				}
				mVal = _bv
{{- else if eq $vc "float"}}
				var _fv float64
				if mv.Type(parser) == fastjson.TypeString {
					_sb, _ev := mv.StringBytes()
					if _ev != nil {
						visitErr = _ev
						return
					}
					var _ev2 error
					_fv, _ev2 = fastfloat.Parse(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ev2 != nil {
						visitErr = _ev2
						return
					}
				} else {
					var _ev error
					_fv, _ev = mv.Float64()
					if _ev != nil {
						visitErr = _ev
						return
					}
				}
				mVal = {{mapValGoType .MapVal}}(_fv)
{{- else if eq $vc "signed"}}
				_iv, _ev := mv.Int64()
				if _ev != nil {
					visitErr = _ev
					return
				}
				mVal = {{mapValGoType .MapVal}}(_iv)
{{- else if eq $vc "signed64"}}
				var _iv int64
				if mv.Type(parser) == fastjson.TypeString {
					_sb, _ev := mv.StringBytes()
					if _ev != nil {
						visitErr = _ev
						return
					}
					var _ev2 error
					_iv, _ev2 = fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ev2 != nil {
						visitErr = _ev2
						return
					}
				} else {
					var _ev error
					_iv, _ev = mv.Int64()
					if _ev != nil {
						visitErr = _ev
						return
					}
				}
				mVal = {{mapValGoType .MapVal}}(_iv)
{{- else if eq $vc "unsigned"}}
				_uv, _ev := mv.Uint64()
				if _ev != nil {
					visitErr = _ev
					return
				}
				mVal = {{mapValGoType .MapVal}}(_uv)
{{- else if eq $vc "unsigned64"}}
				var _uv uint64
				if mv.Type(parser) == fastjson.TypeString {
					_sb, _ev := mv.StringBytes()
					if _ev != nil {
						visitErr = _ev
						return
					}
					var _ev2 error
					_uv, _ev2 = fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ev2 != nil {
						visitErr = _ev2
						return
					}
				} else {
					var _ev error
					_uv, _ev = mv.Uint64()
					if _ev != nil {
						visitErr = _ev
						return
					}
				}
				mVal = {{mapValGoType .MapVal}}(_uv)
{{- end}}
				r.{{.Name}}[mKey] = mVal
{{- end}}
				isContinue = true
				return
			}, parser, {{if eq .MapKey "string"}}false{{else}}true{{end}})
{{- else if .Repeated}}
			_arr, _e := v.Array()
			if _e != nil {
				visitErr = _e
				return
			}
			if cap(r.{{.Name}}) < len(_arr) {
				r.{{.Name}} = make({{.ReaderType}}, 0, len(_arr))
			} else {
				r.{{.Name}} = r.{{.Name}}[:0]
			}
{{- if .IsMsg}}
			for _, _item := range _arr {
				var _elem {{elemType .ReaderType}}
{{- if .MsgIsAsArray}}
				_subArr, _eo := _item.Array()
				if _eo != nil {
					visitErr = _eo
					return
				}
				if _eo2 := _elem.fromJSONArray(_subArr, parser, r.strArena); _eo2 != nil {
					visitErr = _eo2
					return
				}
{{- else}}
				_subObj, _eo := _item.Object()
				if _eo != nil {
					visitErr = _eo
					return
				}
				if _eo2 := _elem.fromJSONValue(_subObj, parser, r.strArena); _eo2 != nil {
					visitErr = _eo2
					return
				}
{{- end}}
				r.{{.Name}} = append(r.{{.Name}}, _elem)
			}
{{- else if .IsEnum}}
			for _, _item := range _arr {
				_iv, _ei := _item.Int64()
				if _ei != nil {
					visitErr = _ei
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_iv))
			}
{{- else}}
{{- $sc := jsonScalarClass .Type}}
			for _, _item := range _arr {
{{- if eq $sc "string"}}
				_s, _ei := _item.UnescapeString(parser)  // unescape and clone
				if _ei != nil {
					visitErr = _ei
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, _s)
{{- else if eq $sc "bytes"}}
				_b64, _ei := _item.StringBytes()
				if _ei != nil {
					visitErr = _ei
					return
				}
				var _dec []byte
				var _ei2 error
				_dec = r.strArena.Reserve(len(_b64))
				_dec, _ei2 = base64.StdEncoding.AppendDecode(_dec[:0], _b64)
				if _ei2 != nil {
					visitErr = _ei2
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, _dec)
{{- else if eq $sc "bool"}}
				_bv, _ei := _item.Bool()
				if _ei != nil {
					visitErr = _ei
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, _bv)
{{- else if eq $sc "float"}}
				var _fv float64
				if _item.Type(parser) == fastjson.TypeString {
					_sb, _ei := _item.StringBytes()
					if _ei != nil {
						visitErr = _ei
						return
					}
					var _ei2 error
					_fv, _ei2 = fastfloat.Parse(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ei2 != nil {
						visitErr = _ei2
						return
					}
				} else {
					var _ei error
					_fv, _ei = _item.Float64()
					if _ei != nil {
						visitErr = _ei
						return
					}
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_fv))
{{- else if eq $sc "signed"}}
				_iv, _ei := _item.Int64()
				if _ei != nil {
					visitErr = _ei
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_iv))
{{- else if eq $sc "signed64"}}
				var _iv int64
				if _item.Type(parser) == fastjson.TypeString {
					_sb, _ei := _item.StringBytes()
					if _ei != nil {
						visitErr = _ei
						return
					}
					var _ei2 error
					_iv, _ei2 = fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ei2 != nil {
						visitErr = _ei2
						return
					}
				} else {
					var _ei error
					_iv, _ei = _item.Int64()
					if _ei != nil {
						visitErr = _ei
						return
					}
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_iv))
{{- else if eq $sc "unsigned"}}
				_uv, _ei := _item.Uint64()
				if _ei != nil {
					visitErr = _ei
					return
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_uv))
{{- else if eq $sc "unsigned64"}}
				var _uv uint64
				if _item.Type(parser) == fastjson.TypeString {
					_sb, _ei := _item.StringBytes()
					if _ei != nil {
						visitErr = _ei
						return
					}
					var _ei2 error
					_uv, _ei2 = fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
					if _ei2 != nil {
						visitErr = _ei2
						return
					}
				} else {
					var _ei error
					_uv, _ei = _item.Uint64()
					if _ei != nil {
						visitErr = _ei
						return
					}
				}
				r.{{.Name}} = append(r.{{.Name}}, {{elemType .ReaderType}}(_uv))
{{- end}}
			}
{{- end}}
{{- else if .IsMsg}}
{{- if .MsgIsAsArray}}
			_arr, _e := v.Array()
			if _e != nil {
				visitErr = _e
				return
			}
{{- if .IsRecursive}}
			if r.{{.Name}} == nil {
				r.{{.Name}} = &{{.ReaderType}}{}
			}
			r._has{{.Name}} = true
{{- end}}
			if _e2 := r.{{.Name}}.fromJSONArray(_arr, parser, r.strArena); _e2 != nil {
				visitErr = _e2
				return
			}
{{- else}}
			_subObj, _e := v.Object()
			if _e != nil {
				visitErr = _e
				return
			}
{{- if .IsRecursive}}
			if r.{{.Name}} == nil {
				r.{{.Name}} = &{{.ReaderType}}{}
			}
			r._has{{.Name}} = true
{{- end}}
			if _e2 := r.{{.Name}}.fromJSONValue(_subObj, parser, r.strArena); _e2 != nil {
				visitErr = _e2
				return
			}
{{- end}}
{{- else if .IsEnum}}
			_iv, _e := v.Int64()
			if _e != nil {
				visitErr = _e
				return
			}
			r.{{.Name}} = {{.ReaderType}}(_iv)
{{- else if .IsDecimal}}
			_fv{{.Name}}, _e{{.Name}} := v.Float64()
			if _e{{.Name}} != nil {
				visitErr = _e{{.Name}}
				return
			}
			r.{{.Name}} = decimal.MustNew(int64(math.Round(_fv{{.Name}} * math.Pow10({{.DecimalRound}}))), {{.DecimalRound}})
{{- else}}
{{- $sc := jsonScalarClass .Type}}
{{- if eq $sc "string"}}
			_s, _e := v.UnescapeString(parser)
			if _e != nil {
				if v.Type(parser) == fastjson.TypeNull {
					// a JSON null value means the field is absent; keep its zero value and skip.
					// (checked on the error path so the non-null fast path never triggers Type()'s unescape side-effect.)
					isContinue = true
					return
				}
				visitErr = _e
				return
			}
			r.{{.Name}} = _s
{{- else if eq $sc "bytes"}}
			_b64, _e := v.StringBytes()
			if _e != nil {
				if v.Type(parser) == fastjson.TypeNull {
					// a JSON null value means the field is absent; keep its zero value and skip.
					// (checked on the error path so the non-null fast path never triggers Type()'s unescape side-effect.)
					isContinue = true
					return
				}
				visitErr = _e
				return
			}
			var _e2 error
			r.{{.Name}}, _e2 = base64.StdEncoding.AppendDecode(r.{{.Name}}[:0], _b64)
			if _e2 != nil {
				visitErr = _e2
				return
			}
{{- else if eq $sc "bool"}}
			_bv, _e := v.Bool()
			if _e != nil {
				visitErr = _e
				return
			}
			r.{{.Name}} = _bv
{{- else if eq $sc "float"}}
			var _fv float64
			if v.Type(parser) == fastjson.TypeString {
				_sb, _e := v.StringBytes()
				if _e != nil {
					visitErr = _e
					return
				}
				var _e2 error
				_fv, _e2 = fastfloat.Parse(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
				if _e2 != nil {
					visitErr = _e2
					return
				}
			} else {
				var _e error
				_fv, _e = v.Float64()
				if _e != nil {
					visitErr = _e
					return
				}
			}
			r.{{.Name}} = {{.ReaderType}}(_fv)
{{- else if eq $sc "signed"}}
			_iv, _e := v.Int64()
			if _e != nil {
				visitErr = _e
				return
			}
			r.{{.Name}} = {{.ReaderType}}(_iv)
{{- else if eq $sc "signed64"}}
			var _iv int64
			if v.Type(parser) == fastjson.TypeString {
				_sb, _e := v.StringBytes()
				if _e != nil {
					visitErr = _e
					return
				}
				var _e2 error
				_iv, _e2 = fastfloat.ParseInt64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
				if _e2 != nil {
					visitErr = _e2
					return
				}
			} else {
				var _e error
				_iv, _e = v.Int64()
				if _e != nil {
					visitErr = _e
					return
				}
			}
			r.{{.Name}} = {{.ReaderType}}(_iv)
{{- else if eq $sc "unsigned"}}
			_uv, _e := v.Uint64()
			if _e != nil {
				visitErr = _e
				return
			}
			r.{{.Name}} = {{.ReaderType}}(_uv)
{{- else if eq $sc "unsigned64"}}
			var _uv uint64
			if v.Type(parser) == fastjson.TypeString {
				_sb, _e := v.StringBytes()
				if _e != nil {
					visitErr = _e
					return
				}
				var _e2 error
				_uv, _e2 = fastfloat.ParseUint64(unsafe.String(unsafe.SliceData(_sb), len(_sb)))
				if _e2 != nil {
					visitErr = _e2
					return
				}
			} else {
				var _e error
				_uv, _e = v.Uint64()
				if _e != nil {
					visitErr = _e
					return
				}
			}
			r.{{.Name}} = {{.ReaderType}}(_uv)
{{- end}}
{{- end}}
{{- end}}
		}
		isContinue = true
		return
	}, parser, true/*skip unescape keys, because all key must by proto field name*/)
	return visitErr
}
{{- end}}

func (r *Readonly{{$goName}}) FromJSON(src []byte) error {
	r.rawBufferLen = len(src)
	parser := &r.parser
{{- if hasStringOrBytesFields .Fields}}
	if r.strArena == nil {
		r.strArena = fastjson.NewArena(r.rawBufferLen)
	}
{{- end}}
	v, err := parser.Parse(unsafe.String(unsafe.SliceData(src), len(src)), r.strArena)
	if err != nil {
		return err
	}
{{- if .AsArray}}
	arr, err := v.Array()
	if err != nil {
		return err
	}
	return r.fromJSONArray(arr, parser, r.strArena)
{{- else}}
	obj, err := v.Object()
	if err != nil {
		return err
	}
	return r.fromJSONValue(obj, parser, r.strArena)
{{- end}}
}
{{end}}
