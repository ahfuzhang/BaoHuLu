{{/* 写类型和 Readonly 类型的 Reset 放在同一个文件。 */ -}}
{{template "partials/header.go.tpl" .}}

{{if hasDecimalFields .Msg.Fields}}
import "github.com/govalues/decimal"
{{end}}

{{with .Msg}}
{{- $goName := .GoName}}
func (m *{{$goName}}) Reset() {
	clear(m.arena) // pointer might be cause not GC correctly, so must use memset(0) to dereference
	m.arena = m.arena[:0]
{{- range .Fields}}
{{- if .Map}}
	clear(m.{{.Name}})
{{- else if .Repeated}}
	m.{{.Name}} = m.{{.Name}}[:0]
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	m.{{.Name}} = nil
{{- else}}
	m.{{.Name}}.Reset()
{{- end}}
{{- else if isSliceType .GoType}}
	m.{{.Name}} = m.{{.Name}}[:0]
{{- else if .IsDecimal}}
	m.{{.Name}} = decimal.Decimal{}
{{- else}}
	m.{{.Name}} = {{zeroVal .GoType}}
{{- end}}
{{- end}}
	/*
	Risk of misuse: After calling `Reset()`, the data in objects referencing the value portion of
	a map, as well as members of array types, might be set to \0. If you want this to persist,
	be sure to use the `Clone()` method to copy them before calling `Reset()`.

	Reference to string and bytes types is safe, as these types are copied during decoding.

	误用风险：调用 Reset() 后，引用 map 的 value 部分的对象指针，以及引用数组类型中的成员，其内部的数据可能被设置为 \0。如果希望长期有效，一定要在 Reset() 之前使用 Clone() 方法进行复制。
	引用字符串类型以及 bytes 类型是安全的，这些类型会在 decode 时进行复制。
	*/
}

func (r *Readonly{{$goName}}) Reset() {
	r.rawBufferLen = 0
	r.parser.Reset()
{{- range .Fields}}
{{- if .Map}}
	clear(r.{{.Name}})
{{- if (mapValIsMsg .MapVal)}}
	clear(r._{{.Name}}Arr)
	r._{{.Name}}Arr = r._{{.Name}}Arr[:0]
{{- end}}
{{- else if .Repeated}}
	r.{{.Name}} = r.{{.Name}}[:0]
{{- else if .IsMsg}}
{{- if .IsRecursive}}
	r._has{{.Name}} = false
	if r.{{.Name}} != nil {
		r.{{.Name}}.Reset()
	}
{{- else}}
	r.{{.Name}}.Reset()
{{- end}}
{{- else if isSliceType .ReaderType}}
	r.{{.Name}} = r.{{.Name}}[:0]
{{- else if .IsDecimal}}
	r.{{.Name}} = decimal.Decimal{}
{{- else}}
	r.{{.Name}} = {{readerZero .ReaderType}}
{{- end}}
{{- end}}
	r.strArena = nil  // Make the object no longer referenced, and let the GC manage its lifecycle.
}
{{end}}
