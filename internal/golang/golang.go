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

// EnumValueGoName converts a protobuf SCREAMING_SNAKE_CASE enum value name to
// Go-style PascalCase. e.g. "STATUS_ACTIVE" → "StatusActive".
func EnumValueGoName(s string) string {
	parts := strings.Split(s, "_")
	var sb strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		sb.WriteString(strings.ToUpper(p[:1]))
		sb.WriteString(strings.ToLower(p[1:]))
	}
	return sb.String()
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
	IsRawBuf   bool   // synthetic rawBuffer []byte field for readonly structs
	StructTag  string // pre-computed struct tag, e.g. `json:"foo,omitempty" yaml:"foo"`
}

type MsgTpl struct {
	Name         string     // proto message name (used as map/lookup key)
	GoName       string     // Go type name: same as proto message name
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

// buildStructTag constructs the full Go struct tag string for a field,
// incorporating json, yaml (@yamlName), and arbitrary extra tags (@tag).
func buildStructTag(fd protofile.FieldDef) string {
	var sb strings.Builder
	sb.WriteByte('`')
	fmt.Fprintf(&sb, `json:"%s,omitempty"`, fd.JsonName)
	if fd.YamlName != "" {
		fmt.Fprintf(&sb, ` yaml:"%s"`, fd.YamlName)
	}
	for _, t := range fd.ExtraTags {
		val := t.Value
		// Strip surrounding double quotes if the user included them as part of
		// the tag value (e.g. @tag=gorm:"col:id" → gorm:"col:id" not gorm:""col:id"").
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		fmt.Fprintf(&sb, ` %s:"%s"`, t.Name, val)
	}
	sb.WriteByte('`')
	return sb.String()
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
		StructTag:  buildStructTag(fd),
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
		"zeroVal":         ZeroVal,
		"readerZero":      ReaderZero,
		"isPackable":      IsPackable,
		"is8ByteNumeric":  Is8ByteNumeric,
		"isSliceType":     func(s string) bool { return strings.HasPrefix(s, "[]") },
		"readFunc":        ReadFuncForType,
		"protoWireType":   ProtoWireType,
		"trimPtr":         func(s string) string { return strings.TrimPrefix(s, "*") },
		"mapKeyGoType":    func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValGoType":    func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValIsMsg":     func(s string) bool { _, isMsg, _ := g.ProtoTypeToGo(s, false); return isMsg },
		"upperFirst":      protofile.UpperFirst,
		"enumValueGoName": EnumValueGoName,
		"readerElemType": func(fd protofile.FieldDef) string {
			return protofile.ReadonlyGoTypeName(fd.Type)
		},
		"readonlyTypeName": protofile.ReadonlyGoTypeName,
		// JSON decoding helpers.
		"jsonMapKeyClass": JsonMapKeyClass,
		"jsonScalarClass": JsonScalarClass,
		"elemType":        func(s string) string { return strings.TrimPrefix(s, "[]") },
		// Extension helpers.
		"hasYamlFields": func(fields []FieldTpl) bool {
			for _, f := range fields {
				if f.YamlName != "" {
					return true
				}
			}
			return false
		},
	}

	tmpl, err := template.New("pb").Funcs(fnMap).Parse(codeTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── test template helpers ───────────────────────────────────────────────────

// SampleScalarLiteral returns a Go source literal for a representative
// non-zero sample value of the given protobuf scalar type.
// goType is used only for enum fallback (e.g. "Status").
func SampleScalarLiteral(protoType, goType string) string {
	switch protoType {
	case "double":
		return "1.5"
	case "float":
		return "float32(1.5)"
	case "int32":
		return "int32(1)"
	case "int64":
		return "int64(2)"
	case "uint32":
		return "uint32(3)"
	case "uint64":
		return "uint64(4)"
	case "sint32":
		return "int32(-1)"
	case "sint64":
		return "int64(-2)"
	case "fixed32":
		return "uint32(5)"
	case "fixed64":
		return "uint64(6)"
	case "sfixed32":
		return "int32(-3)"
	case "sfixed64":
		return "int64(-4)"
	case "bool":
		return "true"
	case "string":
		return `"hello"`
	case "bytes":
		return `utils.UnsafeBytesFromString("data")`
	default:
		// enum or unknown – cast integer 1 to the Go type
		if goType != "" {
			return fmt.Sprintf("%s(1)", goType)
		}
		return "1"
	}
}

// SampleFieldLiteral returns a Go source expression that produces a
// representative non-zero value for field ft, suitable for use inside a
// makeSampleXxx() struct literal.
func SampleFieldLiteral(ft FieldTpl) string {
	if ft.Map {
		// Single-entry map to avoid non-deterministic serialisation order.
		var keyLit string
		switch ft.MapKey {
		case "string":
			keyLit = `"k"`
		case "bool":
			keyLit = "true"
		default:
			keyLit = SampleScalarLiteral(ft.MapKey, "")
		}
		var valLit string
		if ft.MapVal == "bool" {
			valLit = "true"
		} else {
			valLit = SampleScalarLiteral(ft.MapVal, "int32")
		}
		return fmt.Sprintf("%s{%s: %s}", ft.GoType, keyLit, valLit)
	}
	if ft.Repeated {
		elemLit := SampleScalarLiteral(ft.Type, strings.TrimPrefix(ft.GoType, "[]"))
		return fmt.Sprintf("%s{%s}", ft.GoType, elemLit)
	}
	if ft.IsMsg {
		return fmt.Sprintf("makeSample%s()", ft.GoType)
	}
	return SampleScalarLiteral(ft.Type, ft.GoType)
}

// isLargeIntType returns true for proto scalar types whose JSON serialisation
// must use quoted strings when the value exceeds JavaScript's MAX_SAFE_INTEGER
// (2^53 – 1 = 9007199254740991).
func isLargeIntType(protoType string) bool {
	switch protoType {
	case "int64", "uint64", "sint64", "fixed64", "sfixed64":
		return true
	}
	return false
}

// SkipEncodingJSON returns true when the standard encoding/json package cannot
// faithfully round-trip this message's JSON output. This happens when:
//   - any field is a map with a bool key (encoding/json cannot decode them), OR
//   - any field is an embedded message (which might itself contain bool-keyed
//     maps at any depth — conservative but correct).
func SkipEncodingJSON(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.Map && f.MapKey == "bool" {
			return true
		}
		if f.IsMsg {
			return true
		}
	}
	return false
}

// HasLargeIntFields returns true if any field in the list is a direct scalar
// 64-bit integer (not inside a map or repeated slice).
func HasLargeIntFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsRawBuf && isLargeIntType(f.Type) {
			return true
		}
	}
	return false
}

// LargeIntFields returns only the direct scalar 64-bit integer fields.
func LargeIntFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsRawBuf && isLargeIntType(f.Type) {
			out = append(out, f)
		}
	}
	return out
}

// LargeIntLit returns a Go literal whose magnitude exceeds MAX_SAFE_INTEGER,
// exercising the quoted-string serialisation path in ToJSON.
// Signed types use a negative value to also cover the < -MAX_SAFE_INT branch.
func LargeIntLit(ft FieldTpl) string {
	switch ft.Type {
	case "int64":
		return "int64(9007199254740993)" // 2^53 + 1
	case "uint64", "fixed64":
		return "uint64(9007199254740993)"
	case "sint64", "sfixed64":
		return "int64(-9007199254740993)" // tests the < -MAX_SAFE_INT branch
	}
	return "0"
}

// FirstScalarField returns the first plain (non-map, non-repeated, non-msg,
// non-rawBuffer) field, or nil when no such field exists. The returned field
// is used by the test template to generate JSON-type-error tests.
func FirstScalarField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf {
			return f
		}
	}
	return nil
}

// HasMapsOrSlices returns true if any field is a map, repeated slice, or
// embedded message (all of which have container-specific Clone branches that
// benefit from a double-cycle clone test).
func HasMapsOrSlices(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.Map || f.Repeated || f.IsMsg {
			return true
		}
	}
	return false
}

// FirstMsgField returns the first embedded-message field, or nil.
func FirstMsgField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		if fields[i].IsMsg && !fields[i].Map && !fields[i].Repeated {
			return &fields[i]
		}
	}
	return nil
}

// FirstMapField returns the first map field, or nil.
func FirstMapField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		if fields[i].Map {
			return &fields[i]
		}
	}
	return nil
}

// FirstRepeatedField returns the first repeated (slice, non-map) field, or nil.
func FirstRepeatedField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		if fields[i].Repeated && !fields[i].Map {
			return &fields[i]
		}
	}
	return nil
}

// FirstBytesField returns the first plain bytes field (non-map, non-repeated), or nil.
func FirstBytesField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && f.Type == "bytes" {
			return f
		}
	}
	return nil
}

// isNumericProtoType returns true for proto types whose keys are encoded as
// numeric strings in JSON map keys (int32, int64, uint32, uint64, sint32,
// sint64, fixed32, fixed64, sfixed32, sfixed64).
func isNumericProtoType(pt string) bool {
	switch pt {
	case "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64",
		"sfixed32", "sfixed64":
		return true
	}
	return false
}

// FirstStringKeyMapField returns the first map whose key type is string, or nil.
// Used to generate a test that passes null as a map value, exercising the inner
// value-parse error branch.
func FirstStringKeyMapField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if f.Map && f.MapKey == "string" {
			return f
		}
	}
	return nil
}

// FirstNumericKeyMapField returns the first map whose key type is numeric, or nil.
// Used to generate a test that passes a non-numeric map key string, exercising
// the strconv.ParseInt / ParseUint error branch.
func FirstNumericKeyMapField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if f.Map && isNumericProtoType(f.MapKey) {
			return f
		}
	}
	return nil
}

// ─── boundary and string-escape test helpers ─────────────────────────────────

// BoundaryCase represents a single boundary-value test case for a numeric field.
type BoundaryCase struct {
	Label     string // human-readable label, e.g. "MyField_max"
	FieldName string // Go field name
	Lit       string // Go literal, e.g. "int32(math.MaxInt32)"
}

type typeBoundary struct {
	label string
	lit   string
}

func boundaryLitsForType(protoType string) []typeBoundary {
	switch protoType {
	case "int32", "sint32", "sfixed32":
		return []typeBoundary{
			{"max", "int32(math.MaxInt32)"},
			{"min", "int32(math.MinInt32)"},
		}
	case "uint32", "fixed32":
		return []typeBoundary{
			{"max", "uint32(math.MaxUint32)"},
		}
	case "int64", "sint64", "sfixed64":
		return []typeBoundary{
			{"max", "int64(math.MaxInt64)"},
			{"min", "int64(math.MinInt64)"},
		}
	case "uint64", "fixed64":
		return []typeBoundary{
			{"max", "uint64(math.MaxUint64)"},
		}
	case "float":
		return []typeBoundary{
			{"max", "float32(math.MaxFloat32)"},
			{"neg_max", "float32(-math.MaxFloat32)"},
			{"smallest", "float32(math.SmallestNonzeroFloat32)"},
		}
	case "double":
		return []typeBoundary{
			{"max", "math.MaxFloat64"},
			{"neg_max", "-math.MaxFloat64"},
			{"smallest", "math.SmallestNonzeroFloat64"},
		}
	}
	return nil
}

// NumericBoundaryCases returns one BoundaryCase per boundary value per direct
// scalar numeric field in fields. Only plain (non-map, non-repeated, non-msg)
// fields are considered.
func NumericBoundaryCases(fields []FieldTpl) []BoundaryCase {
	var out []BoundaryCase
	for _, f := range fields {
		if f.Map || f.Repeated || f.IsMsg || f.IsRawBuf {
			continue
		}
		for _, b := range boundaryLitsForType(f.Type) {
			out = append(out, BoundaryCase{
				Label:     f.Name + "_" + b.label,
				FieldName: f.Name,
				Lit:       b.lit,
			})
		}
	}
	return out
}

// HasNumericBoundaryFields returns true if any direct scalar numeric field
// (int32, uint32, int64, uint64, float, double and their aliases) exists.
func HasNumericBoundaryFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.Map || f.Repeated || f.IsMsg || f.IsRawBuf {
			continue
		}
		if len(boundaryLitsForType(f.Type)) > 0 {
			return true
		}
	}
	return false
}

// HasFloatFields returns true if any direct scalar float or double field exists.
func HasFloatFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf {
			if f.Type == "float" || f.Type == "double" {
				return true
			}
		}
	}
	return false
}

// FloatFields returns all direct scalar float and double fields.
func FloatFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf {
			if f.Type == "float" || f.Type == "double" {
				out = append(out, f)
			}
		}
	}
	return out
}

// FloatIntLit returns a Go literal whose value is exactly an integer
// (e.g. 4.0 for float32, 10.0 for float64), exercising the code path where
// a JSON serialiser may emit "4" instead of "4.0" and the parser must accept both.
func FloatIntLit(ft FieldTpl) string {
	switch ft.Type {
	case "float":
		return "float32(4.0)"
	case "double":
		return "10.0"
	}
	return "0"
}

// FirstStringScalarField returns the first plain string scalar field
// (not map, not repeated, not msg, not rawBuffer), or nil.
func FirstStringScalarField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && f.Type == "string" {
			return f
		}
	}
	return nil
}

// AnyMsgHasNumericBoundary returns true if any message in msgs has numeric
// boundary fields. Used to decide whether to import "math" in the test file.
func AnyMsgHasNumericBoundary(msgs []MsgTpl) bool {
	for _, msg := range msgs {
		if HasNumericBoundaryFields(msg.Fields) {
			return true
		}
	}
	return false
}

// ─── benchmark template helpers ──────────────────────────────────────────────

// benchScalarMapValLit returns a Go literal for the given proto map-value type.
// For message types (not matching any scalar), it returns a benchBuildXxx() call.
func benchScalarMapValLit(mapVal string) string {
	switch mapVal {
	case "double":
		return "1.5"
	case "float":
		return "float32(1.5)"
	case "int32", "sint32", "sfixed32":
		return "int32(1)"
	case "uint32", "fixed32":
		return "uint32(1)"
	case "int64", "sint64", "sfixed64":
		return "int64(1)"
	case "uint64", "fixed64":
		return "uint64(1)"
	case "bool":
		return "true"
	case "string":
		return `"v"`
	case "bytes":
		return `utils.UnsafeBytesFromString("v")`
	default:
		// Enum or message type – use the Go type name.
		return fmt.Sprintf("benchBuild%s()", protofile.GoTypeName(mapVal))
	}
}

// BenchMapFill generates a Go statement (or block of statements) that fills
// the local variable `m` (already declared with the correct map type) with
// 101 representative entries.  For bool keys only two entries are possible.
func BenchMapFill(ft FieldTpl) string {
	valLit := benchScalarMapValLit(ft.MapVal)
	switch ft.MapKey {
	case "string":
		return fmt.Sprintf(`for i := 0; i < 101; i++ { m[strconv.Itoa(i)] = %s }`, valLit)
	case "bool":
		// Only two distinct bool keys exist.
		return fmt.Sprintf("m[false] = %s\n\t\tm[true] = %s", valLit, valLit)
	case "int32", "sint32", "sfixed32":
		return fmt.Sprintf(`for i := int32(0); i < 101; i++ { m[i] = %s }`, valLit)
	case "uint32", "fixed32":
		return fmt.Sprintf(`for i := uint32(0); i < 101; i++ { m[i] = %s }`, valLit)
	case "int64", "sint64", "sfixed64":
		return fmt.Sprintf(`for i := int64(0); i < 101; i++ { m[i] = %s }`, valLit)
	case "uint64", "fixed64":
		return fmt.Sprintf(`for i := uint64(0); i < 101; i++ { m[i] = %s }`, valLit)
	default:
		return fmt.Sprintf(`for i := int32(0); i < 101; i++ { m[i] = %s }`, valLit)
	}
}

// BenchSliceFill generates a Go statement that fills the local variable `s`
// (already declared with the correct slice type and length 101) with
// representative values.
func BenchSliceFill(ft FieldTpl) string {
	switch ft.Type {
	case "double":
		return `for i := 0; i < 101; i++ { s[i] = float64(i) + 0.5 }`
	case "float":
		return `for i := 0; i < 101; i++ { s[i] = float32(i) + 0.5 }`
	case "int32", "sint32", "sfixed32":
		return `for i := 0; i < 101; i++ { s[i] = int32(i) }`
	case "uint32", "fixed32":
		return `for i := 0; i < 101; i++ { s[i] = uint32(i) }`
	case "int64", "sint64", "sfixed64":
		return `for i := 0; i < 101; i++ { s[i] = int64(i) }`
	case "uint64", "fixed64":
		return `for i := 0; i < 101; i++ { s[i] = uint64(i) }`
	case "bool":
		return `for i := 0; i < 101; i++ { s[i] = i%2 == 0 }`
	case "string":
		return `for i := 0; i < 101; i++ { s[i] = "element with escape chars:\nnewline\ttab\"quote\\backslash:0123456789abcdef" }`
	case "bytes":
		return `for i := 0; i < 101; i++ { s[i] = utils.UnsafeBytesFromString("bytes element 0123456789abcdef") }`
	default:
		if ft.IsMsg {
			elemType := strings.TrimPrefix(ft.GoType, "[]")
			return fmt.Sprintf(`for i := 0; i < 101; i++ { s[i] = benchBuild%s() }`, elemType)
		}
		return `for i := 0; i < 101; i++ { s[i] = 1 }`
	}
}

// BenchNeedsStrconv returns true when any message in msgs contains a
// string-keyed map field, which requires strconv.Itoa in the fill loop.
func BenchNeedsStrconv(msgs []MsgTpl) bool {
	for _, msg := range msgs {
		for _, f := range msg.Fields {
			if f.Map && f.MapKey == "string" {
				return true
			}
		}
	}
	return false
}

// ExcludeFromCompare returns true for fields that must be skipped during the
// encoding/json vs custom-decoder field-level comparison:
//   - map[bool]T: encoding/json cannot unmarshal bool map keys
//   - plain []byte (proto bytes): comparison requires bytes.Equal; easier to
//     verify correctness via the ToJSON round-trip rather than direct field compare
//   - rawBuffer synthetic field
func ExcludeFromCompare(ft FieldTpl) bool {
	if ft.IsRawBuf {
		return true
	}
	if ft.Map && ft.MapKey == "bool" {
		return true
	}
	if !ft.Map && !ft.Repeated && ft.Type == "bytes" {
		return true
	}
	return false
}

// NeedsDeepEqual returns true when the == operator cannot be used for the
// field type and reflect.DeepEqual must be used instead (maps, slices, embedded
// messages).
func NeedsDeepEqual(ft FieldTpl) bool {
	return ft.Map || ft.Repeated || ft.IsMsg
}

// AnyMsgNeedsReflect returns true when at least one message in the file would
// emit a reflect.DeepEqual call in its JSON roundtrip test. Used to decide
// whether to import "reflect" in the generated test file.
func AnyMsgNeedsReflect(msgs []MsgTpl) bool {
	for _, msg := range msgs {
		if SkipEncodingJSON(msg.Fields) {
			continue
		}
		for _, f := range msg.Fields {
			if !ExcludeFromCompare(f) && NeedsDeepEqual(f) {
				return true
			}
		}
	}
	return false
}

// RenderTest renders the test-file template to out.
func (g *Generator) RenderTest(out *os.File) error {
	var enums []EnumTpl
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, EnumTpl{Name: ed.Name, Values: ed.Values})
	}

	writerLayouts := make(map[string]protofile.MsgLayoutInfo)
	readerLayouts := make(map[string]protofile.MsgLayoutInfo)

	var msgs []MsgTpl
	for _, name := range g.Order {
		md := g.Messages[name]
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment}

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
		"sampleLit":                SampleFieldLiteral,
		"hasLargeIntFields":        HasLargeIntFields,
		"largeIntFields":           LargeIntFields,
		"largeIntLit":              LargeIntLit,
		"skipEncodingJSON":         SkipEncodingJSON,
		"firstScalarField":         FirstScalarField,
		"hasMapsOrSlices":          HasMapsOrSlices,
		"firstMsgField":            FirstMsgField,
		"firstMapField":            FirstMapField,
		"firstRepeatedField":       FirstRepeatedField,
		"firstBytesField":          FirstBytesField,
		"firstStringKeyMapField":   FirstStringKeyMapField,
		"firstNumericKeyMapField":  FirstNumericKeyMapField,
		"hasNumericBoundaryFields": HasNumericBoundaryFields,
		"numericBoundaryCases":     NumericBoundaryCases,
		"firstStringScalarField":   FirstStringScalarField,
		"anyMsgHasNumericBoundary": AnyMsgHasNumericBoundary,
		"hasFloatFields":           HasFloatFields,
		"floatFields":              FloatFields,
		"floatIntLit":              FloatIntLit,
	}
	tmpl, err := template.New("pb_test").Funcs(fnMap).Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("parse test template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// RenderBench renders the timing-benchmark test file template to out.
func (g *Generator) RenderBench(out *os.File) error {
	var enums []EnumTpl
	for _, name := range g.EnumOrder() {
		ed := g.Enums[name]
		enums = append(enums, EnumTpl{Name: ed.Name, Values: ed.Values})
	}

	writerLayouts := make(map[string]protofile.MsgLayoutInfo)

	var msgs []MsgTpl
	for _, name := range g.Order {
		md := g.Messages[name]
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment}

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

		msgs = append(msgs, mt)
	}

	data := RenderData{
		Package:  g.Pkg,
		Enums:    enums,
		Messages: msgs,
	}

	fnMap := template.FuncMap{
		"sampleLit":         SampleFieldLiteral,
		"benchMapFill":      BenchMapFill,
		"benchSliceFill":    BenchSliceFill,
		"benchNeedsStrconv": BenchNeedsStrconv,
	}
	tmpl, err := template.New("pb_timing_test").Funcs(fnMap).Parse(benchTemplate)
	if err != nil {
		return fmt.Errorf("parse bench template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── code template ────────────────────────────────────────────────────────────

//go:embed go.tpl
var codeTemplate string

//go:embed go_test.tpl
var testTemplate string

//go:embed go_timing_test.tpl
var benchTemplate string
