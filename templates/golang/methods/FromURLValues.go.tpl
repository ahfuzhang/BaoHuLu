{{with model .}}
{{/* URLValues 解码。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/url"
	"strconv"
{{- if hasDecimalFields .Msg.Fields}}
	"github.com/govalues/decimal"
{{- end}}
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = base64.StdEncoding
	_ = fmt.Errorf
	_ = math.Float32bits
	_ = url.QueryEscape
	_ = strconv.IntSize
{{- if hasDecimalFields .Msg.Fields}}
	_ = decimal.Decimal{}
{{- end}}
)

{{with .Msg}}
{{- $goName := .GoName}}
// ============================================================================
// Readonly{{$goName}} — FromURLValues deserialization (URL query-string format)
// ============================================================================

func (r *Readonly{{$goName}}) FromURLValues(src []byte) error {
	_q, _qe := url.ParseQuery(string(src))
	if _qe != nil {
		return fmt.Errorf("parse url values for {{$goName}}: %w", _qe)
	}
	for _k, _vs := range _q {
		if len(_vs) == 0 {
			continue
		}
		switch _k {
{{range .Fields}}
		case "{{.JsonName}}":
{{- if .Map}}
			{
				_subQ, _e := url.ParseQuery(_vs[0])
				if _e != nil {
					return fmt.Errorf("bad map value for {{.JsonName}}: %w", _e)
				}
{{- if .MapValIsMsg}}
				_dict{{.Name}} := make(map[{{mapKeyGoType .MapKey}}]*{{readonlyTypeName .MapVal}})
{{- else}}
				_dict{{.Name}} := make({{.ReaderType}})
{{- end}}
				for _mk, _mvs := range _subQ {
					if len(_mvs) == 0 {
						continue
					}
{{- if eq .MapKey "string"}}
					_key := _mk
{{- else if eq .MapKey "bool"}}
					_key := _mk == "true" || _mk == "1"
{{- else if or (eq .MapKey "uint64") (eq .MapKey "fixed64")}}
					_keyV, _ke := strconv.ParseUint(_mk, 10, 64)
					if _ke != nil {
						return fmt.Errorf("bad key for {{.JsonName}}: %w", _ke)
					}
					_key := {{mapKeyGoType .MapKey}}(_keyV)
{{- else if or (eq .MapKey "uint32") (eq .MapKey "fixed32")}}
					_keyV, _ke := strconv.ParseUint(_mk, 10, 32)
					if _ke != nil {
						return fmt.Errorf("bad key for {{.JsonName}}: %w", _ke)
					}
					_key := {{mapKeyGoType .MapKey}}(_keyV)
{{- else if or (eq .MapKey "int64") (eq .MapKey "sint64") (eq .MapKey "sfixed64")}}
					_keyV, _ke := strconv.ParseInt(_mk, 10, 64)
					if _ke != nil {
						return fmt.Errorf("bad key for {{.JsonName}}: %w", _ke)
					}
					_key := {{mapKeyGoType .MapKey}}(_keyV)
{{- else}}
					_keyV, _ke := strconv.ParseInt(_mk, 10, 32)
					if _ke != nil {
						return fmt.Errorf("bad key for {{.JsonName}}: %w", _ke)
					}
					_key := {{mapKeyGoType .MapKey}}(_keyV)
{{- end}}
{{- if .MapValIsMsg}}
					_val := new({{readonlyTypeName .MapVal}})
					if _ve := _val.FromJSON([]byte(_mvs[0])); _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = _val
{{- else if eq .MapVal "string"}}
					_dict{{.Name}}[_key] = _mvs[0]
{{- else if eq .MapVal "bytes"}}
					_valB, _ve := base64.StdEncoding.DecodeString(_mvs[0])
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = _valB
{{- else if eq .MapVal "bool"}}
					_dict{{.Name}}[_key] = _mvs[0] == "true" || _mvs[0] == "1"
{{- else if eq .MapVal "double"}}
					_valF, _ve := strconv.ParseFloat(_mvs[0], 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = _valF
{{- else if eq .MapVal "float"}}
					_valF, _ve := strconv.ParseFloat(_mvs[0], 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = float32(_valF)
{{- else if or (eq .MapVal "uint64") (eq .MapVal "fixed64")}}
					_valU, _ve := strconv.ParseUint(_mvs[0], 10, 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = {{mapValGoType .MapVal}}(_valU)
{{- else if or (eq .MapVal "uint32") (eq .MapVal "fixed32")}}
					_valU, _ve := strconv.ParseUint(_mvs[0], 10, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = {{mapValGoType .MapVal}}(_valU)
{{- else if or (eq .MapVal "int64") (eq .MapVal "sint64") (eq .MapVal "sfixed64")}}
					_valI, _ve := strconv.ParseInt(_mvs[0], 10, 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = {{mapValGoType .MapVal}}(_valI)
{{- else}}
					_valI, _ve := strconv.ParseInt(_mvs[0], 10, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_dict{{.Name}}[_key] = {{mapValGoType .MapVal}}(_valI)
{{- end}}
				}
				r.{{.Name}} = _dict{{.Name}}
			}
{{- else if .Repeated}}
			{
				_list{{.Name}} := make({{.ReaderType}}, 0, len(_vs))
				for _, _sv := range _vs {
{{- if .IsMsg}}
					var _elem{{.Name}} {{readonlyTypeName .Type}}
					if _se := _elem{{.Name}}.FromURLValues([]byte(_sv)); _se != nil {
						return _se
					}
					_list{{.Name}} = append(_list{{.Name}}, _elem{{.Name}})
{{- else if eq .Type "bytes"}}
					_b, _be := base64.StdEncoding.DecodeString(_sv)
					if _be != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _be)
					}
					_list{{.Name}} = append(_list{{.Name}}, _b)
{{- else if eq .Type "bool"}}
					_list{{.Name}} = append(_list{{.Name}}, _sv == "true" || _sv == "1")
{{- else if eq .Type "string"}}
					_list{{.Name}} = append(_list{{.Name}}, _sv)
{{- else if .IsEnum}}
					_v, _ve := strconv.ParseInt(_sv, 10, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, {{elemType .ReaderType}}(_v))
{{- else if eq .Type "double"}}
					_v, _ve := strconv.ParseFloat(_sv, 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, _v)
{{- else if eq .Type "float"}}
					_v, _ve := strconv.ParseFloat(_sv, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, float32(_v))
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
					_v, _ve := strconv.ParseUint(_sv, 10, 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, {{elemType .ReaderType}}(_v))
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
					_v, _ve := strconv.ParseUint(_sv, 10, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, {{elemType .ReaderType}}(_v))
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
					_v, _ve := strconv.ParseInt(_sv, 10, 64)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, {{elemType .ReaderType}}(_v))
{{- else}}
					_v, _ve := strconv.ParseInt(_sv, 10, 32)
					if _ve != nil {
						return fmt.Errorf("bad value for {{.JsonName}}: %w", _ve)
					}
					_list{{.Name}} = append(_list{{.Name}}, {{elemType .ReaderType}}(_v))
{{- end}}
				}
				r.{{.Name}} = _list{{.Name}}
			}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
			if r.{{.Name}} == nil {
				r.{{.Name}} = new({{.ReaderType}})
			}
			r._has{{.Name}} = true
			if _e := r.{{.Name}}.FromURLValues([]byte(_vs[0])); _e != nil {
				return _e
			}
{{- else}}
			if _e := r.{{.Name}}.FromURLValues([]byte(_vs[0])); _e != nil {
				return _e
			}
{{- end}}
{{- else if .IsDecimal}}
			{
				_v, _e := strconv.ParseFloat(_vs[0], 64)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = decimal.MustNew(int64(math.Round(_v*math.Pow10({{decimalRound .}}))), {{decimalRound .}})
			}
{{- else if eq .Type "bytes"}}
			{
				_v, _e := base64.StdEncoding.DecodeString(_vs[0])
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = _v
			}
{{- else if eq .Type "bool"}}
			r.{{.Name}} = _vs[0] == "true" || _vs[0] == "1"
{{- else if eq .Type "string"}}
			r.{{.Name}} = _vs[0]
{{- else if .IsEnum}}
			{
				_v, _e := strconv.ParseInt(_vs[0], 10, 32)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = {{.ReaderType}}(_v)
			}
{{- else if eq .Type "double"}}
			{
				_v, _e := strconv.ParseFloat(_vs[0], 64)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = _v
			}
{{- else if eq .Type "float"}}
			{
				_v, _e := strconv.ParseFloat(_vs[0], 32)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = float32(_v)
			}
{{- else if or (eq .Type "uint64") (eq .Type "fixed64")}}
			{
				_v, _e := strconv.ParseUint(_vs[0], 10, 64)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = {{.GoType}}(_v)
			}
{{- else if or (eq .Type "uint32") (eq .Type "fixed32")}}
			{
				_v, _e := strconv.ParseUint(_vs[0], 10, 32)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = {{.GoType}}(_v)
			}
{{- else if or (eq .Type "int64") (eq .Type "sint64") (eq .Type "sfixed64")}}
			{
				_v, _e := strconv.ParseInt(_vs[0], 10, 64)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = {{.GoType}}(_v)
			}
{{- else}}
			{
				_v, _e := strconv.ParseInt(_vs[0], 10, 32)
				if _e != nil {
					return fmt.Errorf("bad value for {{.JsonName}}: %w", _e)
				}
				r.{{.Name}} = {{.GoType}}(_v)
			}
{{- end}}
{{end}}
		}
	}
	return nil
}
{{end}}
{{end}}
