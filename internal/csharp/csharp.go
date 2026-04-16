package csharp

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

// ─── C# type helpers ──────────────────────────────────────────────────────────

var scalarProtoToCS = map[string]string{
	"double":   "double",
	"float":    "float",
	"int32":    "int",
	"int64":    "long",
	"uint32":   "uint",
	"uint64":   "ulong",
	"sint32":   "int",
	"sint64":   "long",
	"fixed32":  "uint",
	"fixed64":  "ulong",
	"sfixed32": "int",
	"sfixed64": "long",
	"bool":     "bool",
	"string":   "string",
	"bytes":    "byte[]",
}

// csWriterType returns the C# type for a field in the mutable writer struct.
func (g *Generator) csWriterType(fd protofile.FieldDef) string {
	if fd.Map {
		keyCS := g.csScalarType(fd.MapKey)
		valCS := g.csValType(fd.MapVal)
		return fmt.Sprintf("Dictionary<%s, %s>", keyCS, valCS)
	}
	if fd.Repeated {
		elem := g.csValType(fd.Type)
		return fmt.Sprintf("List<%s>", elem)
	}
	return g.csValType(fd.Type)
}

// csReadonlyType returns the C# type for a field in the readonly struct.
func (g *Generator) csReadonlyType(fd protofile.FieldDef) string {
	if fd.Map {
		keyCS := g.csScalarType(fd.MapKey)
		valCS := g.csReadonlyValType(fd.MapVal)
		return fmt.Sprintf("ImmutableDictionary<%s, %s>", keyCS, valCS)
	}
	if fd.Repeated {
		elem := g.csReadonlyValType(fd.Type)
		return fmt.Sprintf("List<%s>", elem)
	}
	return g.csReadonlyValType(fd.Type)
}

// csValType returns the C# element type for a proto type (mutable).
func (g *Generator) csValType(protoType string) string {
	if cs, ok := scalarProtoToCS[protoType]; ok {
		return cs
	}
	if _, ok := g.Enums[protoType]; ok {
		return protoType
	}
	return protofile.GoTypeName(protoType) // message: strip "Message" suffix
}

// csReadonlyValType returns the C# element type for readonly views.
func (g *Generator) csReadonlyValType(protoType string) string {
	if cs, ok := scalarProtoToCS[protoType]; ok {
		return cs
	}
	if _, ok := g.Enums[protoType]; ok {
		return protoType
	}
	return "Readonly" + protofile.GoTypeName(protoType)
}

// csScalarType returns the C# primitive for a proto scalar key type.
func (g *Generator) csScalarType(protoType string) string {
	if cs, ok := scalarProtoToCS[protoType]; ok {
		return cs
	}
	return "int"
}

// csDefaultValue returns the C# zero/default literal for a CS type.
func csDefaultValue(csType string) string {
	switch csType {
	case "bool":
		return "false"
	case "string":
		return "\"\""
	case "double":
		return "0.0"
	case "float":
		return "0.0f"
	}
	if strings.HasSuffix(csType, "[]") || strings.HasPrefix(csType, "Dictionary<") || strings.HasPrefix(csType, "ImmutableDictionary<") {
		return "null"
	}
	return "0"
}

// csReadLocalType returns the type of the local variable used during decoding.
func (g *Generator) csReadLocalType(fd protofile.FieldDef) string {
	if fd.Map {
		keyCS := g.csScalarType(fd.MapKey)
		valCS := g.csReadonlyValType(fd.MapVal)
		return fmt.Sprintf("Dictionary<%s, %s>", keyCS, valCS)
	}
	if fd.Repeated {
		elem := g.csReadonlyValType(fd.Type)
		return fmt.Sprintf("List<%s>", elem)
	}
	return g.csReadonlyType(fd)
}

// csProtoWireType returns the protobuf wire type integer for a proto field type.
func csProtoWireType(fd protofile.FieldDef) int {
	if fd.Map || fd.Repeated || fd.IsMsg {
		return 2 // LenDelim
	}
	switch fd.Type {
	case "double", "fixed64", "sfixed64":
		return 1
	case "float", "fixed32", "sfixed32":
		return 5
	case "string", "bytes":
		return 2
	}
	return 0 // varint
}

func csIsPackable(t string) bool {
	switch t {
	case "double", "float", "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool":
		return true
	}
	return false
}

// ─── template data types ──────────────────────────────────────────────────────

type CsRenderData struct {
	Namespace string
	Enums     []protofile.EnumDef
	Messages  []CsMsgTpl
}

// CsFieldTpl carries all info needed by the C# template for one field.
type CsFieldTpl struct {
	// identity
	Name     string // PascalCase name
	JsonName string // original proto name (JSON key)
	Number   int    // proto field number
	// type classification
	IsMap      bool
	IsRepeated bool
	IsMsg      bool
	IsEnum     bool
	IsString   bool
	IsBytes    bool
	IsBool     bool
	IsSint32   bool // sint32
	IsSint64   bool // sint64
	IsFixed32  bool // float, fixed32, sfixed32
	IsFixed64  bool // double, fixed64, sfixed64
	IsPackable bool // repeated packable numeric
	// C# type strings
	WriterType         string // C# type for mutable struct field
	ReadonlyType       string // C# type for readonly struct field
	LocalType          string // C# type for local decode variable
	ElemTypeCS         string // element C# type (for repeated / map values)
	ReadonlyElemTypeCS string // readonly element C# type
	MapKeyCS           string // C# map key type
	MapValCS           string // C# map value type
	ReadonlyMapValCS   string // readonly C# map value type
	MapValIsMsg        bool
	ElemIsMsg          bool // repeated element is a message
	// proto meta
	MapKey string // proto key type (for map entry decode)
	MapVal string // proto val type
	Type   string // proto field type
}

type CsMsgTpl struct {
	Name   string // proto name
	GoName string // stripped name (used as C# type name)
	Fields []CsFieldTpl
}

// ─── generator ────────────────────────────────────────────────────────────────

type Generator struct {
	*protofile.Generator
}

func NewGenerator(pg *protofile.Generator) *Generator {
	return &Generator{Generator: pg}
}

func (g *Generator) buildCSField(fd protofile.FieldDef) CsFieldTpl {
	f := CsFieldTpl{
		Name:               fd.Name,
		JsonName:           fd.JsonName,
		Number:             fd.Number,
		IsMap:              fd.Map,
		IsRepeated:         fd.Repeated,
		IsMsg:              fd.IsMsg,
		IsEnum:             fd.IsEnum,
		IsString:           fd.Type == "string",
		IsBytes:            fd.Type == "bytes",
		IsBool:             fd.Type == "bool",
		IsSint32:           fd.Type == "sint32",
		IsSint64:           fd.Type == "sint64",
		IsFixed32:          fd.Type == "float" || fd.Type == "fixed32" || fd.Type == "sfixed32",
		IsFixed64:          fd.Type == "double" || fd.Type == "fixed64" || fd.Type == "sfixed64",
		IsPackable:         fd.Repeated && csIsPackable(fd.Type),
		WriterType:         g.csWriterType(fd),
		ReadonlyType:       g.csReadonlyType(fd),
		LocalType:          g.csReadLocalType(fd),
		ElemTypeCS:         g.csValType(fd.Type),
		ReadonlyElemTypeCS: g.csReadonlyValType(fd.Type),
		MapKey:             fd.MapKey,
		MapVal:             fd.MapVal,
		Type:               fd.Type,
	}
	if fd.Map {
		f.MapKeyCS = g.csScalarType(fd.MapKey)
		f.MapValCS = g.csValType(fd.MapVal)
		f.ReadonlyMapValCS = g.csReadonlyValType(fd.MapVal)
		_, f.MapValIsMsg, _ = g.ProtoTypeToGo(fd.MapVal, false)
	}
	if fd.Repeated {
		f.ElemIsMsg = fd.IsMsg
	}
	return f
}

func (g *Generator) RenderCS(out *os.File, namespace string) error {
	var enums []protofile.EnumDef
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, *ed)
	}

	// Keep C# field order aligned with the Go writer struct layout.
	writerLayouts := make(map[string]protofile.MsgLayoutInfo)

	var msgs []CsMsgTpl
	for _, name := range g.Order {
		md := g.Messages[name]

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

		sortedFields := protofile.SortFieldsWithCallbacks(md.Fields, writerSizeOf, writerPtrdataOf)
		writerLayouts[name] = protofile.ComputeStructLayout(sortedFields, writerSizeOf, writerPtrdataOf)

		mt := CsMsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name)}
		for _, fd := range sortedFields {
			mt.Fields = append(mt.Fields, g.buildCSField(fd))
		}
		msgs = append(msgs, mt)
	}

	data := CsRenderData{
		Namespace: namespace,
		Enums:     enums,
		Messages:  msgs,
	}

	fnMap := template.FuncMap{
		"csDefault":  csDefaultValue,
		"upperFirst": protofile.UpperFirst,
		"goTypeName": protofile.GoTypeName,
	}

	tmpl, err := template.New("cs").Funcs(fnMap).Parse(csCodeTemplate)
	if err != nil {
		return fmt.Errorf("parse cs template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── C# code template ─────────────────────────────────────────────────────────

//go:embed cs.tpl
var csCodeTemplate string
