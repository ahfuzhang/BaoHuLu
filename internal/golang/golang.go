package golang

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

// ─── wire-type enum ──────────────────────────────────────────────────────────

// WireTypeVal mirrors the protobuf wire-type encoding.
type WireTypeVal int

const (
	WireTypeVarint   WireTypeVal = 0 // int32, int64, uint32, uint64, sint32, sint64, bool, enum
	WireType64bit    WireTypeVal = 1 // fixed64, sfixed64, double
	WireTypeLenDelim WireTypeVal = 2 // string, bytes, embedded messages, packed repeated, map
	WireType32bit    WireTypeVal = 5 // fixed32, sfixed32, float
)

func (w WireTypeVal) String() string {
	switch w {
	case WireTypeVarint:
		return "utils.WireTypeVarint"
	case WireType64bit:
		return "utils.WireType64bit"
	case WireTypeLenDelim:
		return "utils.WireTypeLenDelim"
	case WireType32bit:
		return "utils.WireType32bit"
	}
	return "utils.WireType(0)"
}

func WireType(protoType string, isMsg bool) WireTypeVal {
	switch protoType {
	case "double", "fixed64", "sfixed64":
		return WireType64bit
	case "float", "fixed32", "sfixed32":
		return WireType32bit
	case "bytes", "string":
		return WireTypeLenDelim
	default:
		if isMsg {
			return WireTypeLenDelim
		}
		return WireTypeVarint
	}
}

// ─── template helpers ─────────────────────────────────────────────────────────

func IsPackable(t string) bool {
	switch t {
	case "double", "float", "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool":
		return true
	}
	return false
}

// Is8ByteNumeric returns true for packable proto types whose Go representation is 8 bytes wide.
func Is8ByteNumeric(t string) bool {
	switch t {
	case "double", "int64", "uint64", "sint64", "fixed64", "sfixed64":
		return true
	}
	return false
}

func ZeroVal(goType string) string {
	switch goType {
	case "bool":
		return "false"
	case "string":
		return `""`
	case "[]byte":
		return "nil"
	default:
		if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[") || strings.HasPrefix(goType, "*") {
			return "nil"
		}
		return "0"
	}
}

func ReaderZero(rt string) string {
	if strings.HasPrefix(rt, "[]") || strings.HasPrefix(rt, "map[") {
		return "nil"
	}
	switch rt {
	case "bool":
		return "false"
	case "string":
		return `""`
	case "[]byte":
		return "nil"
	default:
		if strings.HasPrefix(rt, "Readonly") {
			return rt + "{}"
		}
		return "0"
	}
}

func ReadFuncForType(protoType string) string {
	switch protoType {
	case "double":
		return "utils.ReadDouble"
	case "float":
		return "utils.ReadFloat"
	case "int32":
		return "utils.ReadInt32"
	case "int64":
		return "utils.ReadInt64"
	case "uint32":
		return "utils.ReadUint32"
	case "uint64":
		return "utils.ReadUint64"
	case "sint32":
		return "utils.ReadSint32"
	case "sint64":
		return "utils.ReadSint64"
	case "fixed32":
		return "utils.ReadFixed32"
	case "fixed64":
		return "utils.ReadFixed64"
	case "sfixed32":
		return "utils.ReadSfixed32"
	case "sfixed64":
		return "utils.ReadSfixed64"
	case "bool":
		return "utils.ReadBool"
	case "string":
		return "utils.ReadString"
	case "bytes":
		return "utils.ReadBytes"
	}
	return "utils.ReadInt32" // enum
}

func ProtoWireType(pt string) WireTypeVal {
	switch pt {
	case "double", "fixed64", "sfixed64":
		return WireType64bit
	case "float", "fixed32", "sfixed32":
		return WireType32bit
	case "bytes", "string":
		return WireTypeLenDelim
	}
	return WireTypeVarint
}

// ─── JSON decode helpers ──────────────────────────────────────────────────────

// JsonMapKeyClass classifies a proto map-key type for JSON decoding.
func JsonMapKeyClass(mapKey string) string {
	switch mapKey {
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "int32", "sint32", "sfixed32":
		return "signed32"
	case "int64", "sint64", "sfixed64":
		return "signed64"
	case "uint32", "fixed32":
		return "unsigned32"
	case "uint64", "fixed64":
		return "unsigned64"
	default:
		return "string"
	}
}

// JsonScalarClass classifies a proto scalar type for reading from a *fastjson.Value.
func JsonScalarClass(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "bytes":
		return "bytes"
	case "bool":
		return "bool"
	case "float", "double":
		return "float"
	case "int32", "sint32", "sfixed32":
		return "signed"
	case "int64", "sint64", "sfixed64":
		return "signed64"
	case "uint32", "fixed32":
		return "unsigned"
	case "uint64", "fixed64":
		return "unsigned64"
	default:
		return "signed" // enum fallback
	}
}

// ─── template data types ──────────────────────────────────────────────────────

type FieldTpl struct {
	protofile.FieldDef
	WireType   WireTypeVal
	ReaderType string
	IsRawBuf   bool // synthetic rawBuffer []byte field for readonly structs
}

type MsgTpl struct {
	Name         string     // proto message name (used as map/lookup key)
	GoName       string     // Go type name: proto name with "Message" suffix stripped
	Comment      []string   // proto comment lines (without leading //)
	Fields       []FieldTpl // writer fields, sorted for optimal layout
	ReaderFields []FieldTpl // readonly fields = Fields + rawBuffer, all sorted
}

type EnumTpl struct {
	Name   string
	Values []protofile.EnumValue
}

type RenderData struct {
	Package  string
	Enums    []EnumTpl
	Messages []MsgTpl
}

// ─── generator ────────────────────────────────────────────────────────────────

type Generator struct {
	*protofile.Generator
}

func NewGenerator(pg *protofile.Generator) *Generator {
	return &Generator{Generator: pg}
}

func (g *Generator) readerGoType(fd protofile.FieldDef) string {
	if fd.Map {
		keyGo, _, _ := g.ProtoTypeToGo(fd.MapKey, false)
		valGo, isMsg, _ := g.ProtoTypeToGo(fd.MapVal, false)
		if isMsg {
			valGo = protofile.ReadonlyGoTypeName(fd.MapVal)
		}
		return fmt.Sprintf("map[%s]%s", keyGo, valGo)
	}
	if fd.Repeated {
		base, isMsg, _ := g.ProtoTypeToGo(fd.Type, false)
		if isMsg {
			return "[]" + protofile.ReadonlyGoTypeName(fd.Type)
		}
		return "[]" + base
	}
	if fd.IsMsg {
		return protofile.ReadonlyGoTypeName(fd.Type)
	}
	return fd.GoType
}

func (g *Generator) makeFieldTpl(fd protofile.FieldDef) FieldTpl {
	var wt WireTypeVal
	if fd.Map || fd.Repeated {
		wt = WireTypeLenDelim
	} else {
		wt = WireType(fd.Type, fd.IsMsg)
	}
	return FieldTpl{
		FieldDef:   fd,
		WireType:   wt,
		ReaderType: g.readerGoType(fd),
	}
}

func (g *Generator) Render(out *os.File) error {
	var enums []EnumTpl
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, EnumTpl{Name: ed.Name, Values: ed.Values})
	}

	// writerLayouts and readerLayouts store the computed sizeof/ptrdata for each
	// message after its fields have been sorted. Messages are processed in
	// definition order (g.Order), which is always dependency-first (inner messages
	// before outer), so outer messages can look up inner values immediately.
	writerLayouts := make(map[string]protofile.MsgLayoutInfo)
	readerLayouts := make(map[string]protofile.MsgLayoutInfo)

	var msgs []MsgTpl
	for _, name := range g.Order {
		md := g.Messages[name]
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment}

		// --- Writer struct: sort using precomputed writer layouts for IsMsg fields.
		writerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg {
				if li, ok := writerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		writerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg {
				if li, ok := writerLayouts[fd.Type]; ok {
					return li.Ptrdata
				}
			}
			return protofile.FieldPtrdata(fd)
		}

		sortedWriterDefs := protofile.SortFieldsWithCallbacks(md.Fields, writerSizeOf, writerPtrdataOf)
		writerLayouts[name] = protofile.ComputeStructLayout(sortedWriterDefs, writerSizeOf, writerPtrdataOf)

		for _, fd := range sortedWriterDefs {
			mt.Fields = append(mt.Fields, g.makeFieldTpl(fd))
		}

		// --- Readonly struct: include rawBuffer in the sort, and use precomputed
		// readonly layouts for IsMsg fields (readonly types are larger due to rawBuffer).
		rawBufDef := protofile.FieldDef{Name: "rawBuffer", Type: "bytes", GoType: "[]byte"}
		readerDefs := make([]protofile.FieldDef, 0, len(md.Fields)+1)
		readerDefs = append(readerDefs, rawBufDef)
		readerDefs = append(readerDefs, md.Fields...)

		readerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg {
				if li, ok := readerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		readerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg {
				if li, ok := readerLayouts[fd.Type]; ok {
					return li.Ptrdata
				}
			}
			return protofile.FieldPtrdata(fd)
		}

		sortedReaderDefs := protofile.SortFieldsWithCallbacks(readerDefs, readerSizeOf, readerPtrdataOf)
		readerLayouts[name] = protofile.ComputeStructLayout(sortedReaderDefs, readerSizeOf, readerPtrdataOf)

		for _, fd := range sortedReaderDefs {
			if fd.Name == rawBufDef.Name && fd.Number == 0 {
				mt.ReaderFields = append(mt.ReaderFields, FieldTpl{
					FieldDef:   fd,
					ReaderType: "[]byte",
					IsRawBuf:   true,
				})
			} else {
				mt.ReaderFields = append(mt.ReaderFields, g.makeFieldTpl(fd))
			}
		}

		msgs = append(msgs, mt)
	}

	data := RenderData{
		Package:  g.Pkg,
		Enums:    enums,
		Messages: msgs,
	}

	fnMap := template.FuncMap{
		"fieldCommentBlock": func(lines []string) string {
			if len(lines) == 0 {
				return ""
			}
			var sb strings.Builder
			for _, line := range lines {
				sb.WriteString("\t//")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			return sb.String()
		},
		"msgCommentBlock": func(lines []string) string {
			if len(lines) == 0 {
				return ""
			}
			var sb strings.Builder
			for _, line := range lines {
				sb.WriteString("//")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			return sb.String()
		},
		"zeroVal":        ZeroVal,
		"readerZero":     ReaderZero,
		"isPackable":     IsPackable,
		"is8ByteNumeric": Is8ByteNumeric,
		"isSliceType":    func(s string) bool { return strings.HasPrefix(s, "[]") },
		"readFunc":       ReadFuncForType,
		"protoWireType":  ProtoWireType,
		"trimPtr":        func(s string) string { return strings.TrimPrefix(s, "*") },
		"mapKeyGoType":   func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValGoType":   func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValIsMsg":    func(s string) bool { _, isMsg, _ := g.ProtoTypeToGo(s, false); return isMsg },
		"upperFirst":     protofile.UpperFirst,
		"readerElemType": func(fd protofile.FieldDef) string {
			return protofile.ReadonlyGoTypeName(fd.Type)
		},
		"readonlyTypeName": protofile.ReadonlyGoTypeName,
		// JSON decoding helpers.
		"jsonMapKeyClass": JsonMapKeyClass,
		"jsonScalarClass": JsonScalarClass,
		"elemType":        func(s string) string { return strings.TrimPrefix(s, "[]") },
	}

	tmpl, err := template.New("pb").Funcs(fnMap).Parse(codeTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── code template ────────────────────────────────────────────────────────────

//go:embed go.tpl
var codeTemplate string
