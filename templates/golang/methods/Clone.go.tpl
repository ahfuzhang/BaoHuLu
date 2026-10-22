{{with model .}}
{{/* Readonly 类型到写类型的深拷贝。 */ -}}
{{template "partials/header.go.tpl" .}}

import (
	"unsafe"
)

// Keep field-dependent imports valid for every message shape.
var (
	_ = unsafe.Pointer(nil)
)

{{with .Msg}}
{{- $goName := .GoName}}
func (r *Readonly{{$goName}}) Clone(dst *{{$goName}}) *{{$goName}} {
	// _alignZeroPad provides zero bytes for alignment padding without heap allocation.
	// append(slice, _alignZeroPad[:n]...) appends a string — no temporary []byte is allocated.
	const _alignZeroPad = "\x00\x00\x00\x00\x00\x00\x00" // 7 bytes covers up to 8-byte alignment
	if dst == nil {
		dst = &{{$goName}}{}
	}
	if dst.arena == nil {
		dst.arena = make([]byte, 0, r.rawBufferLen)  // todo: 做一个预估大小的函数
	}
{{- range .Fields}}
{{- if .Map}}
	if len(r.{{.Name}}) > 0 {
		if dst.{{.Name}} == nil {
			{{- if mapValIsMsg .MapVal}}
			dst.{{.Name}} = make(map[{{mapKeyGoType .MapKey}}]*{{mapValGoType .MapVal}}, len(r.{{.Name}}))
			{{- else}}
			dst.{{.Name}} = make({{.GoType}}, len(r.{{.Name}}))
			{{- end}}
		}
		for _rk, _rv := range r.{{.Name}} {
		{{- if eq .MapKey "string"}}
			_kloc := len(dst.arena)
			dst.arena = append(dst.arena, _rk...)
			_nk := unsafe.String(&dst.arena[_kloc], len(_rk))
			{{- if eq .MapVal "string"}}
			if len(_rv) > 0 {
				_vloc := len(dst.arena)
				dst.arena = append(dst.arena, _rv...)
				dst.{{.Name}}[_nk] = unsafe.String(&dst.arena[_vloc], len(_rv))
			} else {
				dst.{{.Name}}[_nk] = ""
			}
			{{- else if eq .MapVal "bytes"}}
			if len(_rv) > 0 {
				_vloc := len(dst.arena)
				dst.arena = append(dst.arena, _rv...)
				_vend := _vloc + len(_rv)
				dst.{{.Name}}[_nk] = dst.arena[_vloc:_vend:_vend]
			} else {
				dst.{{.Name}}[_nk] = nil
			}
			{{- else if mapValIsMsg .MapVal}}
			_nv := new({{mapValGoType .MapVal}})  // todo: 在 arena 中分配，减少对象
			_rv.Clone(_nv)
			dst.{{.Name}}[_nk] = _nv
			{{- else}}
			dst.{{.Name}}[_nk] = _rv
			{{- end}}
		{{- else}}
			{{- if eq .MapVal "string"}}
			if len(_rv) > 0 {
				_vloc := len(dst.arena)
				dst.arena = append(dst.arena, _rv...)
				dst.{{.Name}}[_rk] = unsafe.String(&dst.arena[_vloc], len(_rv))
			} else {
				dst.{{.Name}}[_rk] = ""
			}
			{{- else if eq .MapVal "bytes"}}
			if len(_rv) > 0 {
				_vloc := len(dst.arena)
				dst.arena = append(dst.arena, _rv...)
				_vend := _vloc + len(_rv)
				dst.{{.Name}}[_rk] = dst.arena[_vloc:_vend:_vend]
			} else {
				dst.{{.Name}}[_rk] = nil
			}
			{{- else if mapValIsMsg .MapVal}}
			_nv := new({{mapValGoType .MapVal}})  // todo: 这里可以使用数组，减少对象的数量
			_rv.Clone(_nv)
			dst.{{.Name}}[_rk] = _nv
			{{- else}}
			dst.{{.Name}}[_rk] = _rv
			{{- end}}
		{{- end}}
		}
	}
{{- else if .Repeated}}
{{- if isPackable .Type}}
	if len(r.{{.Name}}) > 0 {
		{{- if is8ByteNumeric .Type}}
		if _rem := len(dst.arena) & 7; _rem != 0 {
			dst.arena = append(dst.arena, _alignZeroPad[:8-_rem]...)
		}
		_nb := len(r.{{.Name}}) * 8
		{{- else if eq .Type "bool"}}
		_nb := len(r.{{.Name}}) * 1
		{{- else}}
		if _rem := len(dst.arena) & 3; _rem != 0 {
			dst.arena = append(dst.arena, _alignZeroPad[:4-_rem]...)
		}
		_nb := len(r.{{.Name}}) * 4
		{{- end}}
		if cap(dst.arena)-len(dst.arena) < _nb {
			_nc := cap(dst.arena)*2 + _nb
			_na := make([]byte, len(dst.arena), _nc)
			copy(_na, dst.arena)
			dst.arena = _na
		}
		_loc := len(dst.arena)
		dst.arena = dst.arena[:_loc+_nb]
		copy(dst.arena[_loc:], unsafe.Slice((*byte)(unsafe.Pointer(&r.{{.Name}}[0])), _nb))
		_n := len(r.{{.Name}})
		dst.{{.Name}} = unsafe.Slice((*{{elemType .GoType}})(unsafe.Pointer(&dst.arena[_loc])), _n)
	} else {
		dst.{{.Name}} = dst.{{.Name}}[:0]
	}
{{- else if eq .Type "string"}}
	if len(r.{{.Name}}) > 0 {
		_total := 0
		for _, _s := range r.{{.Name}} {
			_total += len(_s)
		}
		if cap(dst.arena)-len(dst.arena) < _total {
			_nc := cap(dst.arena)*2 + _total
			_na := make([]byte, len(dst.arena), _nc)
			copy(_na, dst.arena)
			dst.arena = _na
		}
		if cap(dst.{{.Name}}) >= len(r.{{.Name}}) {
			dst.{{.Name}} = dst.{{.Name}}[:len(r.{{.Name}})]
		} else {
			dst.{{.Name}} = make([]string, len(r.{{.Name}}))
		}
		for _i, _s := range r.{{.Name}} {
			_loc := len(dst.arena)
			dst.arena = append(dst.arena, _s...)
			dst.{{.Name}}[_i] = unsafe.String(&dst.arena[_loc], len(_s))
		}
	} else {
		dst.{{.Name}} = dst.{{.Name}}[:0]
	}
{{- else if eq .Type "bytes"}}
	if len(r.{{.Name}}) > 0 {
		_total := 0
		for _, _b := range r.{{.Name}} {
			_total += len(_b)
		}
		if cap(dst.arena)-len(dst.arena) < _total {
			_nc := cap(dst.arena)*2 + _total
			_na := make([]byte, len(dst.arena), _nc)
			copy(_na, dst.arena)
			dst.arena = _na
		}
		if cap(dst.{{.Name}}) >= len(r.{{.Name}}) {
			dst.{{.Name}} = dst.{{.Name}}[:len(r.{{.Name}})]
		} else {
			dst.{{.Name}} = make([][]byte, len(r.{{.Name}}))
		}
		for _i, _b := range r.{{.Name}} {
			_loc := len(dst.arena)
			dst.arena = append(dst.arena, _b...)
			_end := _loc + len(_b)
			dst.{{.Name}}[_i] = dst.arena[_loc:_end:_end]
		}
	} else {
		dst.{{.Name}} = dst.{{.Name}}[:0]
	}
{{- else if .IsMsg}}
	if len(r.{{.Name}}) > 0 {
		if cap(dst.{{.Name}}) >= len(r.{{.Name}}) {
			dst.{{.Name}} = dst.{{.Name}}[:len(r.{{.Name}})]
		} else {
			dst.{{.Name}} = make([]{{elemType .GoType}}, len(r.{{.Name}}))
		}
		for _i := range r.{{.Name}} {
			r.{{.Name}}[_i].Clone(&dst.{{.Name}}[_i])
		}
	} else {
		dst.{{.Name}} = dst.{{.Name}}[:0]
	}
{{- else}}
	// repeated enum: direct copy
	if len(r.{{.Name}}) > 0 {
		if cap(dst.{{.Name}}) >= len(r.{{.Name}}) {
			dst.{{.Name}} = dst.{{.Name}}[:len(r.{{.Name}})]
			copy(dst.{{.Name}}, r.{{.Name}})
		} else {
			dst.{{.Name}} = make([]{{elemType .GoType}}, len(r.{{.Name}}))
			copy(dst.{{.Name}}, r.{{.Name}})
		}
	} else {
		dst.{{.Name}} = dst.{{.Name}}[:0]
	}
{{- end}}
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	if r._has{{.Name}} && r.{{.Name}} != nil {
		dst.{{.Name}} = r.{{.Name}}.Clone(dst.{{.Name}})
	}
{{- else}}
	r.{{.Name}}.Clone(&dst.{{.Name}})
{{- end}}
{{- else if eq .Type "string"}}
	if len(r.{{.Name}}) > 0 {
		_loc := len(dst.arena)
		dst.arena = append(dst.arena, r.{{.Name}}...)
		dst.{{.Name}} = unsafe.String(&dst.arena[_loc], len(r.{{.Name}}))
	} else {
		dst.{{.Name}} = ""
	}
{{- else if eq .Type "bytes"}}
	if len(r.{{.Name}}) > 0 {
		_loc := len(dst.arena)
		dst.arena = append(dst.arena, r.{{.Name}}...)
		_end := _loc + len(r.{{.Name}})
		dst.{{.Name}} = dst.arena[_loc:_end:_end]
	} else {
		dst.{{.Name}} = nil
	}
{{- else}}
	dst.{{.Name}} = r.{{.Name}}
{{- end}}
{{- end}}
	return dst
}
{{end}}
{{end}}
