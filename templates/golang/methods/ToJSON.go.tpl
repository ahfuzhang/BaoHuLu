{{with model .}}
{{/* JSON 编码：包含普通对象、AsMap 和 AsArray 分支。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/base64"
	"math"
	"strconv"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = base64.StdEncoding
	_ = math.Float32bits
	_ = strconv.IntSize
	_ = utils.AppendVarint
)

{{with .Msg}}
{{- $goName := .GoName}}
{{- if .AsMap}}
{{- $f := index .Fields 0}}
func (m *{{$goName}}) ToJSON(dst []byte) []byte {
	dst = append(dst, '{')
	_jsonFirst := true
	for _k, _v := range m.{{$f.Name}} {
		if !_jsonFirst {
			dst = append(dst, ',')
		}
		_jsonFirst = false
{{- if eq $f.MapKey "string"}}
		dst = append(dst, '"')
		dst = utils.EncodeJSONString(_k, dst)
		dst = append(dst, '"')
{{- else}}
		dst = append(dst, '"')
{{- if eq $f.MapKey "bool"}}
		if _k {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
{{- else if or (eq $f.MapKey "uint32") (eq $f.MapKey "uint64") (eq $f.MapKey "fixed32") (eq $f.MapKey "fixed64")}}
		dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else}}
		dst = strconv.AppendInt(dst, int64(_k), 10)
{{- end}}
		dst = append(dst, '"')
{{- end}}
		dst = append(dst, ':')
{{- if eq $f.MapVal "string"}}
		dst = append(dst, '"')
		dst = utils.EncodeJSONString(_v, dst)
		dst = append(dst, '"')
{{- else if eq $f.MapVal "bool"}}
		if _v {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
{{- else if eq $f.MapVal "double"}}
		{
			_fv := _v
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 64)
			}
		}
{{- else if eq $f.MapVal "float"}}
		{
			_fv := float64(_v)
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 32)
			}
		}
{{- else if or (eq $f.MapVal "uint32") (eq $f.MapVal "fixed32")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq $f.MapVal "uint64") (eq $f.MapVal "fixed64")}}
		if uint64(_v) > 9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendUint(dst, uint64(_v), 10)
		}
{{- else if eq $f.MapVal "bytes"}}
		dst = append(dst, '"')
		dst = base64.StdEncoding.AppendEncode(dst, _v)
		dst = append(dst, '"')
{{- else if mapValIsMsg $f.MapVal}}
		if _v != nil {
			dst = _v.ToJSON(dst)
		} else {
			dst = append(dst, "null"...)
		}
{{- else if or (eq $f.MapVal "int64") (eq $f.MapVal "sint64") (eq $f.MapVal "sfixed64")}}
		if int64(_v) > 9007199254740991 || int64(_v) < -9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendInt(dst, int64(_v), 10)
		}
{{- else}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
{{- end}}
	}
	dst = append(dst, '}')
	return dst
}
{{- else if .AsArray}}
{{- $f := index .Fields 0}}
func (m *{{$goName}}) ToJSON(dst []byte) []byte {
	dst = append(dst, '[')
	for _i, _v := range m.{{$f.Name}} {
		if _i > 0 {
			dst = append(dst, ',')
		}
{{- if $f.IsMsg}}
		dst = _v.ToJSON(dst)
{{- else if eq $f.Type "string"}}
		dst = append(dst, '"')
		dst = utils.EncodeJSONString(_v, dst)
		dst = append(dst, '"')
{{- else if eq $f.Type "bool"}}
		if _v {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
{{- else if eq $f.Type "double"}}
		{
			_fv := _v
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 64)
			}
		}
{{- else if eq $f.Type "float"}}
		{
			_fv := float64(_v)
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 32)
			}
		}
{{- else if or (eq $f.Type "uint32") (eq $f.Type "fixed32")}}
		dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq $f.Type "uint64") (eq $f.Type "fixed64")}}
		if uint64(_v) > 9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendUint(dst, uint64(_v), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendUint(dst, uint64(_v), 10)
		}
{{- else if eq $f.Type "bytes"}}
		dst = append(dst, '"')
		dst = base64.StdEncoding.AppendEncode(dst, _v)
		dst = append(dst, '"')
{{- else if or (eq $f.Type "int64") (eq $f.Type "sint64") (eq $f.Type "sfixed64")}}
		if int64(_v) > 9007199254740991 || int64(_v) < -9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendInt(dst, int64(_v), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendInt(dst, int64(_v), 10)
		}
{{- else}}
		dst = strconv.AppendInt(dst, int64(_v), 10)
{{- end}}
	}
	dst = append(dst, ']')
	return dst
}
{{- else}}
func (m *{{$goName}}) ToJSON(dst []byte) []byte {
	dst = append(dst, '{')
{{- if .Fields}}
	_jsonFirstField := true
{{- range $i, $f := .Fields}}
{{- if $f.Map}}
	if len(m.{{$f.Name}}) > 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		{
			dst = append(dst, '{')
			_jsonFirst := true
			for _k, _v := range m.{{$f.Name}} {
				if !_jsonFirst {
					dst = append(dst, ',')
				}
				_jsonFirst = false
{{- if eq $f.MapKey "string"}}
				dst = append(dst, '"')
				dst = utils.EncodeJSONString(_k, dst)
				dst = append(dst, '"')
{{- else}}
				dst = append(dst, '"')
{{- if eq $f.MapKey "bool"}}
				if _k {
					dst = append(dst, "true"...)
				} else {
					dst = append(dst, "false"...)
				}
{{- else if or (eq $f.MapKey "uint32") (eq $f.MapKey "uint64") (eq $f.MapKey "fixed32") (eq $f.MapKey "fixed64")}}
				dst = strconv.AppendUint(dst, uint64(_k), 10)
{{- else}}
				dst = strconv.AppendInt(dst, int64(_k), 10)
{{- end}}
				dst = append(dst, '"')
{{- end}}
				dst = append(dst, ':')
{{- if eq $f.MapVal "string"}}
				dst = append(dst, '"')
				dst = utils.EncodeJSONString(_v, dst)
				dst = append(dst, '"')
{{- else if eq $f.MapVal "bool"}}
				if _v {
					dst = append(dst, "true"...)
				} else {
					dst = append(dst, "false"...)
				}
{{- else if eq $f.MapVal "double"}}
				{
					_fv := _v
					_iv := int64(_fv)
					if math.Round(_fv) == _fv && float64(_iv) == _fv {
						dst = strconv.AppendInt(dst, _iv, 10)
					} else {
						dst = strconv.AppendFloat(dst, _fv, 'f', -1, 64)
					}
				}
{{- else if eq $f.MapVal "float"}}
				{
					_fv := float64(_v)
					_iv := int64(_fv)
					if math.Round(_fv) == _fv && float64(_iv) == _fv {
						dst = strconv.AppendInt(dst, _iv, 10)
					} else {
						dst = strconv.AppendFloat(dst, _fv, 'f', -1, 32)
					}
				}
{{- else if or (eq $f.MapVal "uint32") (eq $f.MapVal "fixed32")}}
				dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq $f.MapVal "uint64") (eq $f.MapVal "fixed64")}}
				if uint64(_v) > 9007199254740991 {
					dst = append(dst, '"')
					dst = strconv.AppendUint(dst, uint64(_v), 10)
					dst = append(dst, '"')
				} else {
					dst = strconv.AppendUint(dst, uint64(_v), 10)
				}
{{- else if eq $f.MapVal "bytes"}}
				dst = append(dst, '"')
				dst = base64.StdEncoding.AppendEncode(dst, _v)
				dst = append(dst, '"')
{{- else if mapValIsMsg $f.MapVal}}
				if _v != nil {
					dst = _v.ToJSON(dst)
				} else {
					dst = append(dst, "null"...)
				}
{{- else if or (eq $f.MapVal "int64") (eq $f.MapVal "sint64") (eq $f.MapVal "sfixed64")}}
				if int64(_v) > 9007199254740991 || int64(_v) < -9007199254740991 {
					dst = append(dst, '"')
					dst = strconv.AppendInt(dst, int64(_v), 10)
					dst = append(dst, '"')
				} else {
					dst = strconv.AppendInt(dst, int64(_v), 10)
				}
{{- else}}
				dst = strconv.AppendInt(dst, int64(_v), 10)
{{- end}}
			}
			dst = append(dst, '}')
		}
	}
{{- else if $f.Repeated}}
	if len(m.{{$f.Name}}) > 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		{
			dst = append(dst, '[')
			for _i, _v := range m.{{$f.Name}} {
				if _i > 0 {
					dst = append(dst, ',')
				}
{{- if $f.IsMsg}}
				dst = _v.ToJSON(dst)
{{- else if eq $f.Type "string"}}
				dst = append(dst, '"')
				dst = utils.EncodeJSONString(_v, dst)
				dst = append(dst, '"')
{{- else if eq $f.Type "bool"}}
				if _v {
					dst = append(dst, "true"...)
				} else {
					dst = append(dst, "false"...)
				}
{{- else if eq $f.Type "double"}}
				{
					_fv := _v
					_iv := int64(_fv)
					if math.Round(_fv) == _fv && float64(_iv) == _fv {
						dst = strconv.AppendInt(dst, _iv, 10)
					} else {
						dst = strconv.AppendFloat(dst, _fv, 'f', -1, 64)
					}
				}
{{- else if eq $f.Type "float"}}
				{
					_fv := float64(_v)
					_iv := int64(_fv)
					if math.Round(_fv) == _fv && float64(_iv) == _fv {
						dst = strconv.AppendInt(dst, _iv, 10)
					} else {
						dst = strconv.AppendFloat(dst, _fv, 'f', -1, 32)
					}
				}
{{- else if or (eq $f.Type "uint32") (eq $f.Type "fixed32")}}
				dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq $f.Type "uint64") (eq $f.Type "fixed64")}}
				if uint64(_v) > 9007199254740991 {
					dst = append(dst, '"')
					dst = strconv.AppendUint(dst, uint64(_v), 10)
					dst = append(dst, '"')
				} else {
					dst = strconv.AppendUint(dst, uint64(_v), 10)
				}
{{- else if eq $f.Type "bytes"}}
				dst = append(dst, '"')
				dst = base64.StdEncoding.AppendEncode(dst, _v)
				dst = append(dst, '"')
{{- else if or (eq $f.Type "int64") (eq $f.Type "sint64") (eq $f.Type "sfixed64")}}
				if int64(_v) > 9007199254740991 || int64(_v) < -9007199254740991 {
					dst = append(dst, '"')
					dst = strconv.AppendInt(dst, int64(_v), 10)
					dst = append(dst, '"')
				} else {
					dst = strconv.AppendInt(dst, int64(_v), 10)
				}
{{- else}}
				dst = strconv.AppendInt(dst, int64(_v), 10)
{{- end}}
			}
			dst = append(dst, ']')
		}
	}
{{- else if $f.IsMsg}}
{{- if $f.IsRecursive}}
	if m.{{$f.Name}} != nil {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = m.{{$f.Name}}.ToJSON(dst)
	}
{{- else}}
	if !_jsonFirstField {
		dst = append(dst, ',')
	}
	_jsonFirstField = false
	dst = append(dst, '"')
	dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
	dst = append(dst, '"', ':')
	dst = m.{{$f.Name}}.ToJSON(dst)
{{- end}}
{{- else if eq $f.Type "string"}}
	if len(m.{{$f.Name}}) > 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = append(dst, '"')
		dst = utils.EncodeJSONString(m.{{$f.Name}}, dst)
		dst = append(dst, '"')
	}
{{- else if eq $f.Type "bool"}}
	if m.{{$f.Name}} {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = append(dst, "true"...)
	}
{{- else if $f.IsDecimal}}
	if !m.{{$f.Name}}.IsZero() {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		{
			_df, _ := m.{{$f.Name}}.Round({{decimalRound $f}}).Float64()
			dst = strconv.AppendFloat(dst, _df, 'f', {{decimalRound $f}}, 64)
		}
	}
{{- else if eq $f.Type "double"}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		{
			_fv := m.{{$f.Name}}
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 64)
			}
		}
	}
{{- else if eq $f.Type "float"}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		{
			_fv := float64(m.{{$f.Name}})
			_iv := int64(_fv)
			if math.Round(_fv) == _fv && float64(_iv) == _fv {
				dst = strconv.AppendInt(dst, _iv, 10)
			} else {
				dst = strconv.AppendFloat(dst, _fv, 'f', -1, 32)
			}
		}
	}
{{- else if or (eq $f.Type "uint32") (eq $f.Type "fixed32")}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = strconv.AppendUint(dst, uint64(m.{{$f.Name}}), 10)
	}
{{- else if or (eq $f.Type "uint64") (eq $f.Type "fixed64")}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		if uint64(m.{{$f.Name}}) > 9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendUint(dst, uint64(m.{{$f.Name}}), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendUint(dst, uint64(m.{{$f.Name}}), 10)
		}
	}
{{- else if eq $f.Type "bytes"}}
	if len(m.{{$f.Name}}) > 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = append(dst, '"')
		dst = base64.StdEncoding.AppendEncode(dst, m.{{$f.Name}})
		dst = append(dst, '"')
	}
{{- else if or (eq $f.Type "int64") (eq $f.Type "sint64") (eq $f.Type "sfixed64")}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		if int64(m.{{$f.Name}}) > 9007199254740991 || int64(m.{{$f.Name}}) < -9007199254740991 {
			dst = append(dst, '"')
			dst = strconv.AppendInt(dst, int64(m.{{$f.Name}}), 10)
			dst = append(dst, '"')
		} else {
			dst = strconv.AppendInt(dst, int64(m.{{$f.Name}}), 10)
		}
	}
{{- else}}
	if m.{{$f.Name}} != 0 {
		if !_jsonFirstField {
			dst = append(dst, ',')
		}
		_jsonFirstField = false
		dst = append(dst, '"')
		dst = append(dst, NameOf{{$goName}}{{$f.Name}}...)
		dst = append(dst, '"', ':')
		dst = strconv.AppendInt(dst, int64(m.{{$f.Name}}), 10)
	}
{{- end}}
{{- end}}
{{- end}}
	dst = append(dst, '}')
	return dst
}
{{- end}}
{{end}}
{{end}}
