package csharp

import (
	_ "embed"
	"fmt"
	"io"
	"math/bits"
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
	if fd.DecimalRound > 0 {
		return "decimal"
	}
	return g.csValType(fd.Type)
}

// csReadonlyType returns the C# type for a field in the readonly struct.
func (g *Generator) csReadonlyType(fd protofile.FieldDef) string {
	if fd.Map {
		keyCS := g.csScalarType(fd.MapKey)
		valCS := g.csReadonlyValType(fd.MapVal)
		return fmt.Sprintf("Dictionary<%s, %s>", keyCS, valCS)
	}
	if fd.Repeated {
		elem := g.csReadonlyValType(fd.Type)
		return fmt.Sprintf("List<%s>", elem)
	}
	if fd.DecimalRound > 0 {
		return "decimal"
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
	case "decimal":
		return "0m"
	}
	if strings.HasSuffix(csType, "[]") || strings.HasPrefix(csType, "Dictionary<") {
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
	Namespace    string
	BaseFileName string
	Enums        []protofile.EnumDef
	Messages     []CsMsgTpl
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
	Name                 string   // PascalCase name
	JsonName             string   // original proto name (JSON key)
	Number               int      // proto field number
	Comment              []string // proto comment lines (without leading //), extension lines stripped
	CommentLineNums      []int    // proto file line numbers for each Comment entry (1-based)
	InlineComment        string   // trailing // comment on the same line as the field (text only, without //)
	InlineCommentLineNum int      // proto file line number of the field itself (1-based), 0 if absent
	// type classification
	IsMap        bool
	IsRepeated   bool
	IsMsg        bool
	IsEnum       bool
	IsString     bool
	IsBytes      bool
	IsBool       bool
	IsSint32     bool // sint32
	IsSint64     bool // sint64
	IsFixed32    bool // float, fixed32, sfixed32
	IsFixed64    bool // double, fixed64, sfixed64
	IsPackable   bool // repeated packable numeric
	IsDecimal    bool // @decimal=round:N annotation: double → System.Decimal
	DecimalRound int  // rounding precision for @decimal fields
	// yaml
	YamlName string // @yamlName override; non-empty overrides yaml key
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
	MapValIsEnum       bool // map value type is an enum
	ElemIsMsg          bool // repeated element is a message
	// proto meta
	MapKey string // proto key type (for map entry decode)
	MapVal string // proto val type
	Type   string // proto field type
	// Wrapper support (set in buildMsgTpls second pass)
	IsRecursive          bool   // from protofile.FieldDef.IsRecursive
	UseMapValWrapper     bool   // map value is a message: use Wrapper class for map values
	UseDirectWrapper     bool   // plain msg field that is recursive: use Wrapper class
	EffWriterType        string // effective C# type for writer field (with wrapper if applicable)
	EffReadonlyType      string // effective C# type for readonly field (with wrapper if applicable)
	EffLocalType         string // effective C# type for local decode variable (with wrapper if applicable)
	WrapMapValCS         string // e.g. "ValueTypesWrapper" — writer wrapper type for map value
	WrapReadonlyMapValCS string // e.g. "ReadonlyValueTypesWrapper" — readonly wrapper type for map value
	MapValIsSelfRef      bool   // map value type is the same as the enclosing message (self-referential)
}

type CsMsgTpl struct {
	Name            string   // proto name
	GoName          string   // stripped name (used as C# type name)
	Comment         []string // proto message comment lines (without leading //), extension lines stripped
	CommentLineNums []int    // proto file line numbers for each Comment entry (1-based)
	Fields          []CsFieldTpl
	NeedsWrapper bool // true when this message type needs Wrapper classes generated
	AsMap        bool // true when @AsMap annotation present: single map field, JSON is flat map
	AsArray      bool // true when @AsArray annotation present: single repeated field, JSON is a bare array
	UrlValues    bool // true when @UrlValues annotation present: generate ToURLValues/FromURLValues
	Yaml         bool // true when @yaml annotation present: generate ToYAML/FromYAML methods
}

// ─── generator ────────────────────────────────────────────────────────────────

type Generator struct {
	*protofile.Generator
}

func NewGenerator(pg *protofile.Generator) *Generator {
	return &Generator{Generator: pg}
}

func (g *Generator) buildCSField(fd protofile.FieldDef) CsFieldTpl {
	isDecimal := fd.DecimalRound > 0
	f := CsFieldTpl{
		Name:                 fd.Name,
		JsonName:             fd.JsonName,
		Number:               fd.Number,
		Comment:              fd.Comment,
		CommentLineNums:      fd.CommentLineNums,
		InlineComment:        fd.InlineComment,
		InlineCommentLineNum: fd.InlineCommentLineNum,
		IsMap:                fd.Map,
		IsRepeated:         fd.Repeated,
		IsMsg:              fd.IsMsg,
		IsEnum:             fd.IsEnum,
		IsString:           fd.Type == "string",
		IsBytes:            fd.Type == "bytes",
		IsBool:             fd.Type == "bool",
		IsSint32:           fd.Type == "sint32",
		IsSint64:           fd.Type == "sint64",
		IsFixed32:          fd.Type == "float" || fd.Type == "fixed32" || fd.Type == "sfixed32",
		IsFixed64:          !isDecimal && (fd.Type == "double" || fd.Type == "fixed64" || fd.Type == "sfixed64"),
		IsPackable:         fd.Repeated && csIsPackable(fd.Type),
		WriterType:         g.csWriterType(fd),
		ReadonlyType:       g.csReadonlyType(fd),
		LocalType:          g.csReadLocalType(fd),
		ElemTypeCS:         g.csValType(fd.Type),
		ReadonlyElemTypeCS: g.csReadonlyValType(fd.Type),
		MapKey:             fd.MapKey,
		MapVal:             fd.MapVal,
		Type:               fd.Type,
		IsDecimal:          isDecimal,
		DecimalRound:       fd.DecimalRound,
		YamlName:           fd.YamlName,
	}
	if fd.Map {
		f.MapKeyCS = g.csScalarType(fd.MapKey)
		f.MapValCS = g.csValType(fd.MapVal)
		f.ReadonlyMapValCS = g.csReadonlyValType(fd.MapVal)
		_, f.MapValIsMsg, f.MapValIsEnum = g.ProtoTypeToGo(fd.MapVal, false)
	}
	if fd.Repeated {
		f.ElemIsMsg = fd.IsMsg
	}
	f.IsRecursive = fd.IsRecursive
	f.UseDirectWrapper = fd.IsMsg && !fd.Map && !fd.Repeated && fd.IsRecursive
	return f
}

// buildCSTmpl compiles the C# code template with standard FuncMap.
func buildCSTmpl() (*template.Template, error) {
	fnMap := template.FuncMap{
		"csDefault":  csDefaultValue,
		"upperFirst": protofile.UpperFirst,
		"goTypeName": protofile.GoTypeName,
		// tagSize computes the byte size of a varint-encoded proto field tag at
		// template generation time so the generated code contains a literal integer.
		"tagSize": func(fieldNum, wireType int) int {
			tag := uint64(fieldNum<<3 | wireType)
			return (bits.Len64(tag|1) + 6) / 7
		},
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
		mt := CsMsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment, CommentLineNums: md.CommentLineNums, AsMap: md.AsMap, AsArray: md.AsArray, UrlValues: md.UrlValues, Yaml: md.Yaml}
		for _, fd := range sortedFields {
			mt.Fields = append(mt.Fields, g.buildCSField(fd))
		}
		msgs = append(msgs, mt)
	}

	// Second pass: compute wrapper-related fields.
	// A message type T needs Wrapper classes when T is used as a map value or as a direct recursive field.
	wrapperNeeded := make(map[string]bool)
	for _, mt := range msgs {
		for _, f := range mt.Fields {
			if f.IsMap && f.MapValIsMsg {
				wrapperNeeded[f.MapVal] = true
			}
			if f.UseDirectWrapper {
				wrapperNeeded[f.Type] = true
			}
		}
	}
	for i := range msgs {
		msgs[i].NeedsWrapper = wrapperNeeded[msgs[i].Name]
		for j := range msgs[i].Fields {
			f := &msgs[i].Fields[j]
			if f.IsMap && f.MapValIsMsg {
				f.UseMapValWrapper = true
				f.WrapMapValCS = f.MapValCS + "Wrapper"
				f.WrapReadonlyMapValCS = f.ReadonlyMapValCS + "Wrapper"
				f.EffWriterType = fmt.Sprintf("Dictionary<%s, %s>", f.MapKeyCS, f.WrapMapValCS)
				f.EffReadonlyType = fmt.Sprintf("Dictionary<%s, %s>", f.MapKeyCS, f.WrapReadonlyMapValCS)
				f.EffLocalType = f.EffReadonlyType
				f.MapValIsSelfRef = (f.MapVal == msgs[i].Name)
			} else if f.UseDirectWrapper {
				f.EffWriterType = f.WriterType + "Wrapper"
				f.EffReadonlyType = f.ReadonlyType + "Wrapper"
				f.EffLocalType = f.EffReadonlyType
			} else {
				f.EffWriterType = f.WriterType
				f.EffReadonlyType = f.ReadonlyType
				f.EffLocalType = f.LocalType
			}
		}
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

		// Xx.UrlValues and ReadonlyXx.UrlValues — only when @UrlValues annotation present.
		if mt.UrlValues {
			writerUVPath := filepath.Join(outDir, baseFileName+"."+mt.GoName+".UrlValues.cs")
			wuv, err := os.Create(writerUVPath)
			if err != nil {
				return fmt.Errorf("create %s: %w", writerUVPath, err)
			}
			err = renderCSUrlValuesTmpl(wuv, "CsUrlValuesWriterFile", data)
			wuv.Close()
			if err != nil {
				return fmt.Errorf("render %s: %w", writerUVPath, err)
			}

			readonlyUVPath := filepath.Join(outDir, baseFileName+".Readonly"+mt.GoName+".UrlValues.cs")
			ruv, err := os.Create(readonlyUVPath)
			if err != nil {
				return fmt.Errorf("create %s: %w", readonlyUVPath, err)
			}
			err = renderCSUrlValuesTmpl(ruv, "CsUrlValuesReadonlyFile", data)
			ruv.Close()
			if err != nil {
				return fmt.Errorf("render %s: %w", readonlyUVPath, err)
			}
		}
	}
	return writeCSGitignore(outDir)
}

const csGitignoreContent = `/bin/
/obj/
/Tests/bin/
/Tests/obj/
/Benchmarks/bin/
/Benchmarks/obj/
`

// writeCSGitignore writes a .gitignore to outDir if one does not already exist.
func writeCSGitignore(outDir string) error {
	p := filepath.Join(outDir, ".gitignore")
	if _, err := os.Stat(p); err == nil {
		return nil // already exists
	}
	return os.WriteFile(p, []byte(csGitignoreContent), 0o644)
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
	case "decimal":
		return "1.12345m"
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
			if f.UseMapValWrapper {
				if f.MapValIsSelfRef {
					// self-referential map value: avoid infinite recursion
					valSample = fmt.Sprintf("new %s()", f.WrapMapValCS)
				} else {
					valSample = fmt.Sprintf("new %s { Value = %sTests.MakeSample%s() }", f.WrapMapValCS, f.MapValCS, f.MapValCS)
				}
			} else {
				valSample = fmt.Sprintf("%sTests.MakeSample%s()", f.MapValCS, f.MapValCS)
			}
		} else {
			valSample = primitiveCSLit(f.MapValCS)
			if valSample == "" {
				valSample = fmt.Sprintf("(%s)1", f.MapValCS)
			}
		}
		dictType := f.EffWriterType
		if dictType == "" {
			dictType = f.WriterType
		}
		return fmt.Sprintf("new %s { { %s, %s } }", dictType, keySample, valSample)
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
		if f.UseDirectWrapper {
			// recursive field: avoid infinite recursion, use empty wrapper
			return fmt.Sprintf("new %s()", f.EffWriterType)
		}
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

// csDecimalFields returns all plain (non-map, non-repeated) decimal fields in the list.
func csDecimalFields(fields []CsFieldTpl) []CsFieldTpl {
	var out []CsFieldTpl
	for _, f := range fields {
		if f.IsDecimal && !f.IsMap && !f.IsRepeated {
			out = append(out, f)
		}
	}
	return out
}

// csYamlSpecialStr is a C# string literal whose runtime value contains characters
// that trigger YAML quoting in _strVal: double-quote, brackets, braces, newline,
// tab, and '#' (the inline-comment marker).
const csYamlSpecialStr = `"hello \"world\" [brk] {brc}\nnewline\ttab #comment"`

// csYamlSampleLit is like csSampleLit but substitutes YAML-special C# string
// literals for every string-typed position so that the quoting path in ToYAML
// and the unquote path in FromYAML are exercised.
func csYamlSampleLit(f CsFieldTpl) string {
	if f.IsMap {
		keySample := primitiveCSLit(f.MapKeyCS)
		if keySample == "" {
			keySample = "1"
		}
		if f.MapKeyCS == "string" {
			// backslash triggers quoting; no colon to avoid map-key parser ambiguity
			keySample = `"key\\slash"`
		}
		if f.MapValIsMsg {
			return csSampleLit(f) // delegate for message values
		}
		valSample := primitiveCSLit(f.MapValCS)
		if valSample == "" {
			valSample = fmt.Sprintf("(%s)1", f.MapValCS)
		}
		if f.MapValCS == "string" {
			valSample = `"val:has:colons"` // colon triggers quoting
		}
		dictType := f.EffWriterType
		if dictType == "" {
			dictType = f.WriterType
		}
		return fmt.Sprintf("new %s { { %s, %s } }", dictType, keySample, valSample)
	}
	if f.IsRepeated {
		if f.ElemIsMsg {
			return csSampleLit(f) // delegate for message elements
		}
		var elemSample string
		if f.IsEnum {
			elemSample = fmt.Sprintf("(%s)1", f.ElemTypeCS)
		} else {
			elemSample = primitiveCSLit(f.ElemTypeCS)
			if elemSample == "" {
				elemSample = fmt.Sprintf("(%s)1", f.ElemTypeCS)
			}
			if f.ElemTypeCS == "string" {
				elemSample = `"item \"quoted\""` // double-quote triggers quoting
			}
		}
		return fmt.Sprintf("new %s { %s }", f.WriterType, elemSample)
	}
	if f.IsMsg {
		return csSampleLit(f)
	}
	if f.IsString {
		return csYamlSpecialStr
	}
	return csSampleLit(f)
}

// anyMsgHasYaml reports whether any message in msgs has YAML serialization enabled.
func anyMsgHasYaml(msgs []CsMsgTpl) bool {
	for _, m := range msgs {
		if m.Yaml {
			return true
		}
	}
	return false
}

// safeUnknownFieldNum returns a field number guaranteed to be above any field
// defined in the message, suitable for use as an "unknown field" in tests.
func safeUnknownFieldNum(fields []CsFieldTpl) int {
	max := 0
	for _, f := range fields {
		if f.Number > max {
			max = f.Number
		}
	}
	return max + 100
}

// csTagByteArray encodes the protobuf tag (fieldNum<<3)|wireType as a varint
// and returns a C# byte-array literal, e.g. "new byte[] { 0xD8, 0x0D }".
func csTagByteArray(fieldNum, wireType int) string {
	tag := uint64(fieldNum<<3 | wireType)
	var hexBytes []string
	for tag >= 0x80 {
		hexBytes = append(hexBytes, fmt.Sprintf("0x%02X", byte(tag)|0x80))
		tag >>= 7
	}
	hexBytes = append(hexBytes, fmt.Sprintf("0x%02X", byte(tag)))
	return "new byte[] { " + strings.Join(hexBytes, ", ") + " }"
}

// RenderCSTest renders the unit-test source file for the parsed proto into out.
// namespace is the proto namespace (same as used for the main C# file).
// baseFileName is the PascalCase proto base name (e.g. "DemoServer") used to
// disambiguate generated helper class names when multiple protos share a directory.
func (g *Generator) RenderCSTest(out *os.File, namespace, baseFileName string) error {
	var enums []protofile.EnumDef
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, *ed)
	}

	msgs, _ := g.buildMsgTpls()
	needsYaml := g.csYamlNeedsSet()
	for i := range msgs {
		// YAML methods are also generated for message types referenced by an
		// @yaml root, so their tests must follow the same transitive closure.
		msgs[i].Yaml = needsYaml[msgs[i].Name]
	}

	data := CsRenderData{
		Namespace:    namespace,
		BaseFileName: baseFileName,
		Enums:        enums,
		Messages:     msgs,
	}

	fnMap := template.FuncMap{
		"csDefault":           csDefaultValue,
		"upperFirst":          protofile.UpperFirst,
		"goTypeName":          protofile.GoTypeName,
		"csSampleLit":         csSampleLit,
		"csYamlSampleLit":     csYamlSampleLit,
		"anyMsgHasYaml":       anyMsgHasYaml,
		"firstCsStringField":  firstCsStringField,
		"safeUnknownFieldNum": safeUnknownFieldNum,
		"csTagByteArray":      csTagByteArray,
		"csDecimalFields":     csDecimalFields,
	}

	tmpl, err := template.New("cs_test").Funcs(fnMap).Parse(csTestCodeTemplate)
	if err != nil {
		return fmt.Errorf("parse cs_test template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── C# benchmark helpers ─────────────────────────────────────────────────────

// benchCsMapValLit returns a C# expression for the value side of a map fill loop.
// The loop variable is named `i`. For message values it calls BenchBuild.BuildXxx().
// String values use LargeString (≥100 bytes with escape chars) to match Go benchmark.
func benchCsMapValLit(f CsFieldTpl) string {
	if f.MapValIsMsg {
		return fmt.Sprintf("new %s { Value = Build%s() }", f.WrapMapValCS, f.MapValCS)
	}
	switch f.MapValCS {
	case "bool":
		return "i % 2 == 0"
	case "string":
		return "LargeString"
	case "byte[]":
		return "new byte[] { (byte)(i & 0xFF) }"
	case "float":
		return "(float)i * 0.5f"
	case "double":
		return "(double)i * 0.5"
	case "long":
		return "(long)i"
	case "ulong":
		return "(ulong)i"
	case "uint":
		return "(uint)i"
	case "int":
		return "i"
	default:
		// enum or unknown – cast integer 1 to the enum type
		return fmt.Sprintf("(%s)1", f.MapValCS)
	}
}

// benchCsFixedMapValLit returns a constant C# expression for a map value when
// there is no loop variable available (used for bool-keyed maps).
func benchCsFixedMapValLit(f CsFieldTpl) string {
	if f.MapValIsMsg {
		return fmt.Sprintf("new %s { Value = Build%s() }", f.WrapMapValCS, f.MapValCS)
	}
	switch f.MapValCS {
	case "bool":
		return "true"
	case "string":
		return `"value"`
	case "byte[]":
		return "new byte[] { 0x01 }"
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
	default:
		return fmt.Sprintf("(%s)1", f.MapValCS)
	}
}

// BenchCsMapFill generates a C# statement (or pair of statements) that fills the
// local variable `dict` (type = f.WriterType) with 101 representative entries.
// Bool-keyed maps produce only 2 entries (true/false) since that is all that exist;
// those entries use a fixed literal value (no loop variable `i` in scope).
func BenchCsMapFill(f CsFieldTpl) string {
	switch f.MapKeyCS {
	case "string":
		return fmt.Sprintf("for (var i = 0; i < 101; i++) { dict[i.ToString()] = %s; }", benchCsMapValLit(f))
	case "bool":
		fixed := benchCsFixedMapValLit(f)
		return fmt.Sprintf("dict[true] = %s; dict[false] = %s;", fixed, fixed)
	case "long":
		return fmt.Sprintf("for (var i = 0; i < 101; i++) { dict[(long)i] = %s; }", benchCsMapValLit(f))
	case "ulong":
		return fmt.Sprintf("for (var i = 0; i < 101; i++) { dict[(ulong)i] = %s; }", benchCsMapValLit(f))
	case "uint":
		return fmt.Sprintf("for (var i = 0; i < 101; i++) { dict[(uint)i] = %s; }", benchCsMapValLit(f))
	default: // "int" (covers int32, sint32, fixed32, sfixed32)
		return fmt.Sprintf("for (var i = 0; i < 101; i++) { dict[i] = %s; }", benchCsMapValLit(f))
	}
}

// benchCsElemLit returns a C# expression for a list element in a fill loop.
// The loop variable is named `i`. For message elements it calls BenchBuild.BuildXxx().
// String elements use LargeString (≥100 bytes with escape chars) to match Go benchmark.
func benchCsElemLit(f CsFieldTpl) string {
	if f.ElemIsMsg {
		return fmt.Sprintf("Build%s()", f.ElemTypeCS)
	}
	switch f.ElemTypeCS {
	case "bool":
		return "i % 2 == 0"
	case "string":
		return "LargeString"
	case "byte[]":
		return "new byte[] { (byte)(i & 0xFF) }"
	case "float":
		return "(float)i * 0.5f"
	case "double":
		return "(double)i * 0.5"
	case "long":
		return "(long)i"
	case "ulong":
		return "(ulong)i"
	case "uint":
		return "(uint)i"
	case "int":
		return "i"
	default:
		// enum or unknown
		return fmt.Sprintf("(%s)1", f.ElemTypeCS)
	}
}

// BenchCsSliceFill generates a C# for-loop that fills the local variable `lst`
// (a List<T>) with 101 representative elements.
func BenchCsSliceFill(f CsFieldTpl) string {
	return fmt.Sprintf("for (var i = 0; i < 101; i++) { lst.Add(%s); }", benchCsElemLit(f))
}

// benchCsScalarLit returns a C# literal for a scalar (non-map, non-repeated,
// non-message, non-string, non-bytes) field in the BenchBuild helper.
// Signed integer types (sint32, sint64, sfixed32, sfixed64) use negative values
// to exercise signed-encoding code paths, mirroring the Go benchmark logic.
func benchCsScalarLit(f CsFieldTpl) string {
	if f.IsEnum {
		return fmt.Sprintf("(%s)1", f.WriterType)
	}
	switch f.Type {
	case "bool":
		return "true"
	case "double":
		return "1.5"
	case "float":
		return "1.5f"
	case "sint32", "sfixed32":
		return "-1"
	case "sint64", "sfixed64":
		return "-1L"
	case "int64":
		return "1L"
	case "uint32", "fixed32":
		return "1U"
	case "uint64", "fixed64":
		return "1UL"
	default: // int32 and other integer types
		return "1"
	}
}

// RenderCSBench renders the benchmark source file for the parsed proto into out.
// namespace is the C# namespace (same as used for the main and test files).
// baseFileName is the PascalCase proto base name used to disambiguate static
// helper class names when multiple protos share a benchmark directory.
func (g *Generator) RenderCSBench(out *os.File, namespace, baseFileName string) error {
	msgs, _ := g.buildMsgTpls()

	data := CsRenderData{
		Namespace:    namespace,
		BaseFileName: baseFileName,
		Messages:     msgs,
	}

	fnMap := template.FuncMap{
		"csDefault":        csDefaultValue,
		"upperFirst":       protofile.UpperFirst,
		"goTypeName":       protofile.GoTypeName,
		"csSampleLit":      csSampleLit,
		"benchCsMapFill":   BenchCsMapFill,
		"benchCsSliceFill": BenchCsSliceFill,
		"benchCsScalarLit": benchCsScalarLit,
	}

	tmpl, err := template.New("cs_bench").Funcs(fnMap).Parse(csBenchCodeTemplate)
	if err != nil {
		return fmt.Errorf("parse cs_bench template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── URL values helpers ───────────────────────────────────────────────────────

// csUrlKeyParse returns a C# statement that parses the map-key loop variable _mk
// (string) into a typed local variable _key of the given C# map-key type. Returns
// Error on failure.
func csUrlKeyParse(keyCS, fieldJsonName string) string {
	errExpr := `return Error.WithLoc(1, "bad key ` + fieldJsonName + `");`
	switch keyCS {
	case "string":
		return "var _key = _mk;"
	case "bool":
		return `var _key = _mk == "true" || _mk == "1";`
	case "long":
		return "if (!long.TryParse(_mk, out long _key)) " + errExpr
	case "ulong":
		return "if (!ulong.TryParse(_mk, out ulong _key)) " + errExpr
	case "uint":
		return "if (!uint.TryParse(_mk, out uint _key)) " + errExpr
	default: // int (int32, sint32, sfixed32, fixed32)
		return "if (!int.TryParse(_mk, out int _key)) " + errExpr
	}
}

// buildCSUrlValuesTmpl compiles the URL-values template.
func buildCSUrlValuesTmpl() (*template.Template, error) {
	fnMap := template.FuncMap{
		"csUrlKeyParse": csUrlKeyParse,
		"csDefault":     csDefaultValue,
	}
	tmpl, err := template.New("cs_urlvalues").Funcs(fnMap).Parse(csUrlValuesCodeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse cs_urlvalues template: %w", err)
	}
	return tmpl, nil
}

// renderCSUrlValuesTmpl executes the named URL-values sub-template with data into w.
func renderCSUrlValuesTmpl(w io.Writer, name string, data CsOneTypeData) error {
	tmpl, err := buildCSUrlValuesTmpl()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// ─── C# code template ─────────────────────────────────────────────────────────

//go:embed templates/message.cs.tpl
var csCodeTemplate string

//go:embed templates/test.cs.tpl
var csTestCodeTemplate string

//go:embed templates/benchmark.cs.tpl
var csBenchCodeTemplate string

//go:embed templates/url.values.cs.tpl
var csUrlValuesCodeTemplate string

//go:embed templates/message.yaml.cs.tpl
var csYamlCodeTemplate string

// ─── C# YAML helpers ──────────────────────────────────────────────────────────

// buildCSYamlTmpl compiles the YAML template with its FuncMap.
func buildCSYamlTmpl() (*template.Template, error) {
	fnMap := template.FuncMap{
		"csDefault": csDefaultValue,
		"yamlKey": func(f CsFieldTpl) string {
			if f.YamlName != "" {
				return f.YamlName
			}
			return f.JsonName
		},
		// csYamlCommentExpr returns a C# expression for the block comment text (without '#').
		// When lineNums[idx] > 0 it returns a constant reference (_GoName_YamlComments.Line{N}).
		// The caller writes "#" before this expression. Falls back to an inline string literal.
		"csYamlCommentExpr": func(goName string, lineNums []int, idx int, text string) string {
			if idx < len(lineNums) && lineNums[idx] > 0 {
				return fmt.Sprintf("_%s_YamlComments.Line%d", goName, lineNums[idx])
			}
			return fmt.Sprintf("%q", text+"\n")
		},
		// csYamlInlineCommentExpr returns a C# expression for the field's trailing inline comment
		// (without '#'). When InlineCommentLineNum > 0 it returns a constant reference.
		"csYamlInlineCommentExpr": func(goName string, f CsFieldTpl) string {
			if f.InlineCommentLineNum > 0 {
				return fmt.Sprintf("_%s_YamlComments.Line%d", goName, f.InlineCommentLineNum)
			}
			return fmt.Sprintf("%q", f.InlineComment+"\n")
		},
		// hasAnyCsYamlComments reports whether the message or any of its fields have
		// comment lines or inline trailing comments with known proto file line numbers.
		"hasAnyCsYamlComments": func(msg CsMsgTpl) bool {
			for _, n := range msg.CommentLineNums {
				if n > 0 {
					return true
				}
			}
			for _, f := range msg.Fields {
				for _, n := range f.CommentLineNums {
					if n > 0 {
						return true
					}
				}
				if f.InlineCommentLineNum > 0 {
					return true
				}
			}
			return false
		},
	}
	tmpl, err := template.New("cs_yaml").Funcs(fnMap).Parse(csYamlCodeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse cs_yaml template: %w", err)
	}
	return tmpl, nil
}

// renderCSYamlTmpl executes the named YAML sub-template with data into w.
func renderCSYamlTmpl(w io.Writer, name string, data CsOneTypeData) error {
	tmpl, err := buildCSYamlTmpl()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// csYamlNeedsSet computes the transitive closure of message names that need
// YAML methods generated. It starts with @yaml-annotated messages, then
// expands to all transitively referenced message types.
func (g *Generator) csYamlNeedsSet() map[string]bool {
	needs := make(map[string]bool)
	for name, md := range g.Messages {
		if md.Yaml {
			needs[name] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for name := range needs {
			for _, fd := range g.Messages[name].Fields {
				var ref string
				if fd.Map {
					if _, ok := g.Messages[fd.MapVal]; ok {
						ref = fd.MapVal
					}
				} else if fd.IsMsg {
					ref = fd.Type
				}
				if ref != "" && !needs[ref] {
					needs[ref] = true
					changed = true
				}
			}
		}
	}
	return needs
}

// RenderCSYAMLFiles generates one pair of .cs files per @yaml-annotated message
// (and every message transitively referenced) into outDir:
//   - "{base}.{GoName}.yaml.cs"         — ToYAML on the mutable writer struct
//   - "{base}.Readonly{GoName}.yaml.cs" — FromYAML on the readonly struct
func (g *Generator) RenderCSYAMLFiles(outDir, baseFileName, namespace string) error {
	needsYaml := g.csYamlNeedsSet()
	if len(needsYaml) == 0 {
		return nil
	}

	msgs, _ := g.buildMsgTpls()

	for _, mt := range msgs {
		if !needsYaml[mt.Name] {
			continue
		}
		data := CsOneTypeData{Namespace: namespace, Msg: mt}

		writerPath := filepath.Join(outDir, baseFileName+"."+mt.GoName+".yaml.cs")
		wf, err := os.Create(writerPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", writerPath, err)
		}
		err = renderCSYamlTmpl(wf, "CsYamlWriterFile", data)
		wf.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", writerPath, err)
		}

		readonlyPath := filepath.Join(outDir, baseFileName+".Readonly"+mt.GoName+".yaml.cs")
		rf, err := os.Create(readonlyPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", readonlyPath, err)
		}
		err = renderCSYamlTmpl(rf, "CsYamlReadonlyFile", data)
		rf.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", readonlyPath, err)
		}
	}
	return nil
}
