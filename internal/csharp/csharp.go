package csharp

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// CsOneTypeData is passed to the per-type named templates (CsTagsWriterFile,
// CsReadonlyFile) so they can render a single message into its own file.
type CsOneTypeData struct {
	Namespace string
	Msg       CsMsgTpl
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

// buildCSTmpl compiles the C# code template with standard FuncMap.
func buildCSTmpl() (*template.Template, error) {
	fnMap := template.FuncMap{
		"csDefault":  csDefaultValue,
		"upperFirst": protofile.UpperFirst,
		"goTypeName": protofile.GoTypeName,
	}
	tmpl, err := template.New("cs").Funcs(fnMap).Parse(csCodeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse cs template: %w", err)
	}
	return tmpl, nil
}

// renderCSData executes the C# template with the given data into w.
func renderCSData(w io.Writer, data CsRenderData) error {
	tmpl, err := buildCSTmpl()
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

// renderCSNamedTmpl executes the named sub-template (e.g. "CsTagsWriterFile"
// or "CsReadonlyFile") with data into w.
func renderCSNamedTmpl(w io.Writer, name string, data CsOneTypeData) error {
	tmpl, err := buildCSTmpl()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// buildMsgTpls returns the per-message template data in proto order,
// honouring struct-layout constraints for field ordering.
func (g *Generator) buildMsgTpls() ([]CsMsgTpl, map[string]protofile.MsgLayoutInfo) {
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
	return msgs, writerLayouts
}

// buildEnumTpls returns all enum definitions in declaration order.
func (g *Generator) buildEnumTpls() []protofile.EnumDef {
	var enums []protofile.EnumDef
	for _, name := range g.EnumOrder() {
		enums = append(enums, *g.Enums[name])
	}
	return enums
}

// RenderCS renders all enums and messages into the single file out.
func (g *Generator) RenderCS(out *os.File, namespace string) error {
	msgs, _ := g.buildMsgTpls()
	return renderCSData(out, CsRenderData{
		Namespace: namespace,
		Enums:     g.buildEnumTpls(),
		Messages:  msgs,
	})
}

// RenderCSFiles generates per-type .cs files into outDir:
//   - "{base}.Enums.cs"          — shared enums (when present)
//   - "{base}.{GoName}.cs"       — XxTags + Xx (mutable writer)
//   - "{base}.Readonly{GoName}.cs" — ReadonlyXx (immutable reader)
//
// Each file contains exactly one logical type group, so ReportGenerator shows
// focused, per-type coverage pages.
func (g *Generator) RenderCSFiles(outDir, baseFileName, namespace string) error {
	enums := g.buildEnumTpls()
	msgs, _ := g.buildMsgTpls()

	// Enums file — only when the proto defines enums.
	if len(enums) > 0 {
		p := filepath.Join(outDir, baseFileName+".Enums.cs")
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("create %s: %w", p, err)
		}
		err = renderCSData(f, CsRenderData{Namespace: namespace, Enums: enums})
		f.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", p, err)
		}
	}

	for _, mt := range msgs {
		data := CsOneTypeData{Namespace: namespace, Msg: mt}

		// XxTags + Xx (mutable writer)
		writerPath := filepath.Join(outDir, baseFileName+"."+mt.GoName+".cs")
		wf, err := os.Create(writerPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", writerPath, err)
		}
		err = renderCSNamedTmpl(wf, "CsTagsWriterFile", data)
		wf.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", writerPath, err)
		}

		// ReadonlyXx
		readonlyPath := filepath.Join(outDir, baseFileName+".Readonly"+mt.GoName+".cs")
		rf, err := os.Create(readonlyPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", readonlyPath, err)
		}
		err = renderCSNamedTmpl(rf, "CsReadonlyFile", data)
		rf.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", readonlyPath, err)
		}
	}
	return nil
}

// ─── C# test helpers ──────────────────────────────────────────────────────────

// primitiveCSLit returns a non-zero C# literal for a primitive CS type,
// or an empty string when the type is not a known primitive.
func primitiveCSLit(csType string) string {
	switch csType {
	case "bool":
		return "true"
	case "string":
		return `"hello"`
	case "byte[]":
		return "new byte[] { 0x01, 0x02, 0x03 }"
	case "float":
		return "1.5f"
	case "double":
		return "1.5"
	case "long":
		return "42L"
	case "ulong":
		return "42UL"
	case "uint":
		return "42U"
	case "int":
		return "42"
	}
	return ""
}

// csSampleLit returns a non-zero C# literal suitable for populating a field in
// the MakeSample helper function of the test file.
func csSampleLit(f CsFieldTpl) string {
	if f.IsMap {
		keySample := primitiveCSLit(f.MapKeyCS)
		if keySample == "" {
			keySample = "1"
		}
		var valSample string
		if f.MapValIsMsg {
			valSample = fmt.Sprintf("%sTests.MakeSample%s()", f.MapValCS, f.MapValCS)
		} else {
			valSample = primitiveCSLit(f.MapValCS)
			if valSample == "" {
				valSample = fmt.Sprintf("(%s)1", f.MapValCS)
			}
		}
		return fmt.Sprintf("new %s { { %s, %s } }", f.WriterType, keySample, valSample)
	}
	if f.IsRepeated {
		var elemSample string
		if f.ElemIsMsg {
			elemSample = fmt.Sprintf("%sTests.MakeSample%s()", f.ElemTypeCS, f.ElemTypeCS)
		} else if f.IsEnum {
			elemSample = fmt.Sprintf("(%s)1", f.ElemTypeCS)
		} else {
			elemSample = primitiveCSLit(f.ElemTypeCS)
			if elemSample == "" {
				elemSample = fmt.Sprintf("(%s)1", f.ElemTypeCS)
			}
		}
		return fmt.Sprintf("new %s { %s }", f.WriterType, elemSample)
	}
	if f.IsMsg {
		return fmt.Sprintf("%sTests.MakeSample%s()", f.WriterType, f.WriterType)
	}
	if f.IsEnum {
		return fmt.Sprintf("(%s)1", f.WriterType)
	}
	sample := primitiveCSLit(f.WriterType)
	if sample == "" {
		return fmt.Sprintf("(%s)1", f.WriterType)
	}
	return sample
}

// firstCsStringField returns the first scalar (non-map, non-repeated) string
// field, or nil when no such field exists.
func firstCsStringField(fields []CsFieldTpl) *CsFieldTpl {
	for i := range fields {
		f := &fields[i]
		if !f.IsMap && !f.IsRepeated && f.IsString {
			return f
		}
	}
	return nil
}

// RenderCSTest renders the unit-test source file for the parsed proto into out.
// namespace is the proto namespace (same as used for the main C# file).
func (g *Generator) RenderCSTest(out *os.File, namespace string) error {
	var enums []protofile.EnumDef
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, *ed)
	}

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
		"csDefault":          csDefaultValue,
		"upperFirst":         protofile.UpperFirst,
		"goTypeName":         protofile.GoTypeName,
		"csSampleLit":        csSampleLit,
		"firstCsStringField": firstCsStringField,
	}

	tmpl, err := template.New("cs_test").Funcs(fnMap).Parse(csTestCodeTemplate)
	if err != nil {
		return fmt.Errorf("parse cs_test template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── C# code template ─────────────────────────────────────────────────────────

//go:embed cs.tpl
var csCodeTemplate string

//go:embed cs_test.tpl
var csTestCodeTemplate string
