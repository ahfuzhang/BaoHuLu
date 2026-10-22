{{with model .}}
{{/* URLValues 编码。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/base64"
	"net/url"
	"strconv"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = base64.StdEncoding
	_ = url.QueryEscape
	_ = strconv.IntSize
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
// ============================================================================
// {{$goName}} — ToURLValues serialization (URL query-string format)
// ============================================================================

func (m *{{$goName}}) ToURLValues(dst []byte) []byte {
{{- if .Fields}}
	_first := true
{{range .Fields}}
{{- if .Map}}
	if len(m.{{.Name}}) > 0 {
		_mapBuf{{.Name}} := make([]byte, 0, 256)
		_mapFirst{{.Name}} := true
		for _k, _v := range m.{{.Name}} {
			if !_mapFirst{{.Name}} {
				_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, '&')
			}
			_mapFirst{{.Name}} = false
{{- if eq .MapKey "string"}}
			_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, _k...)
{{- else if eq .MapKey "bool"}}
			if _k {
				_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, "true"...)
			} else {
				_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, "false"...)
			}
{{- else if or (eq .MapKey "uint64") (eq .MapKey "fixed64")}}
			_mapBuf{{.Name}} = strconv.AppendUint(_mapBuf{{.Name}}, uint64(_k), 10)
{{- else if or (eq .MapKey "uint32") (eq .MapKey "fixed32")}}
			_mapBuf{{.Name}} = strconv.AppendUint(_mapBuf{{.Name}}, uint64(_k), 10)
{{- else if or (eq .MapKey "int64") (eq .MapKey "sint64") (eq .MapKey "sfixed64")}}
			_mapBuf{{.Name}} = strconv.AppendInt(_mapBuf{{.Name}}, int64(_k), 10)
{{- else}}
			_mapBuf{{.Name}} = strconv.AppendInt(_mapBuf{{.Name}}, int64(_k), 10)
{{- end}}
			_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, '=')
{{- if .MapValIsMsg}}
			_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, _v.ToJSON(nil)...)
{{- else if eq .MapVal "string"}}
			_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, _v...)
{{- else if eq .MapVal "bytes"}}
			_mapBuf{{.Name}} = base64.StdEncoding.AppendEncode(_mapBuf{{.Name}}, _v)
{{- else if eq .MapVal "bool"}}
			if _v {
				_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, "true"...)
			} else {
				_mapBuf{{.Name}} = append(_mapBuf{{.Name}}, "false"...)
			}
{{- else if eq .MapVal "double"}}
			_mapBuf{{.Name}} = strconv.AppendFloat(_mapBuf{{.Name}}, _v, 'f', -1, 64)
{{- else if eq .MapVal "float"}}
			_mapBuf{{.Name}} = strconv.AppendFloat(_mapBuf{{.Name}}, float64(_v), 'f', -1, 32)
{{- else if or (eq .MapVal "uint64") (eq .MapVal "fixed64")}}
			_mapBuf{{.Name}} = strconv.AppendUint(_mapBuf{{.Name}}, uint64(_v), 10)
{{- else if or (eq .MapVal "uint32") (eq .MapVal "fixed32")}}
			_mapBuf{{.Name}} = strconv.AppendUint(_mapBuf{{.Name}}, uint64(_v), 10)
{{- else if or (eq .MapVal "int64") (eq .MapVal "sint64") (eq .MapVal "sfixed64")}}
			_mapBuf{{.Name}} = strconv.AppendInt(_mapBuf{{.Name}}, int64(_v), 10)
{{- else}}
			_mapBuf{{.Name}} = strconv.AppendInt(_mapBuf{{.Name}}, int64(_v), 10)
{{- end}}
		}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = append(dst, url.QueryEscape(string(_mapBuf{{.Name}}))...)
	}
{{- else if .Repeated}}
	for _, _v := range m.{{.Name}} {
{{- if .IsMsg}}
		{
			_sub{{.Name}} := _v.ToURLValues(nil)
			if !_first {
				dst = append(dst, '&')
			}
			_first = false
			dst = append(dst, "{{.JsonName}}="...)
			dst = append(dst, url.QueryEscape(string(_sub{{.Name}}))...)
		}
{{- else if eq .Type "string"}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = append(dst, url.QueryEscape(_v)...)
{{- else if eq .Type "bytes"}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = base64.StdEncoding.AppendEncode(dst, _v)
{{- else if eq .Type "bool"}}
		if _v {
			if !_first {
				dst = append(dst, '&')
			}
			_first = false
			dst = append(dst, "{{.JsonName}}=true"...)
		} else {
			if !_first {
				dst = append(dst, '&')
			}
			_first = false
			dst = append(dst, "{{.JsonName}}=false"...)
		}
{{- else if eq .Type "double"}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendFloat(dst, _v, 'f', -1, 64)
{{- else if eq .Type "float"}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendFloat(dst, float64(_v), 'f', -1, 32)
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendUint(dst, uint64(_v), 10)
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendInt(dst, int64(_v), 10)
{{- else}}
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendInt(dst, int64(_v), 10)
{{- end}}
	}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if m.{{.Name}} != nil {
		_sub{{.Name}} := m.{{.Name}}.ToURLValues(nil)
		if len(_sub{{.Name}}) > 0 {
			if !_first {
				dst = append(dst, '&')
			}
			_first = false
			dst = append(dst, "{{.JsonName}}="...)
			dst = append(dst, url.QueryEscape(string(_sub{{.Name}}))...)
		}
	}
{{- else}}
	{
		_sub{{.Name}} := m.{{.Name}}.ToURLValues(nil)
		if len(_sub{{.Name}}) > 0 {
			if !_first {
				dst = append(dst, '&')
			}
			_first = false
			dst = append(dst, "{{.JsonName}}="...)
			dst = append(dst, url.QueryEscape(string(_sub{{.Name}}))...)
		}
	}
{{- end}}
{{- else if .IsDecimal}}
	if !m.{{.Name}}.IsZero() {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		_df{{.Name}}, _ := m.{{.Name}}.Round({{decimalRound .}}).Float64()
		dst = strconv.AppendFloat(dst, _df{{.Name}}, 'f', {{decimalRound .}}, 64)
	}
{{- else if eq .Type "bytes"}}
	if len(m.{{.Name}}) > 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = base64.StdEncoding.AppendEncode(dst, m.{{.Name}})
	}
{{- else if eq .Type "bool"}}
	if m.{{.Name}} {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}=true"...)
	}
{{- else if eq .Type "string"}}
	if m.{{.Name}} != "" {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = append(dst, url.QueryEscape(m.{{.Name}})...)
	}
{{- else if .IsEnum}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendInt(dst, int64(m.{{.Name}}), 10)
	}
{{- else if eq .Type "double"}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendFloat(dst, m.{{.Name}}, 'f', -1, 64)
	}
{{- else if eq .Type "float"}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendFloat(dst, float64(m.{{.Name}}), 'f', -1, 32)
	}
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendUint(dst, m.{{.Name}}, 10)
	}
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendUint(dst, uint64(m.{{.Name}}), 10)
	}
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendInt(dst, m.{{.Name}}, 10)
	}
{{- else}}
	if m.{{.Name}} != 0 {
		if !_first {
			dst = append(dst, '&')
		}
		_first = false
		dst = append(dst, "{{.JsonName}}="...)
		dst = strconv.AppendInt(dst, int64(m.{{.Name}}), 10)
	}
{{- end}}
{{end}}
	_ = _first
{{- end}}
	return dst
}
{{end}}
{{end}}
