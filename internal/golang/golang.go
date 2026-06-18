package golang

import (
	_ "embed"
	"fmt"
	"math/bits"
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
	case "decimal.Decimal":
		return "decimal.Decimal{}"
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
	case "decimal.Decimal":
		return "decimal.Decimal{}"
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

// ─── alignment helpers ────────────────────────────────────────────────────────

// padRight pads s on the right with spaces so its total length is w.
// If len(s) >= w, s is returned unchanged.
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// maxTagLen returns the max length of "{goName}{field.Name}Tag" across all fields.
func maxTagLen(goName string, fields []FieldTpl) int {
	max := 0
	for _, f := range fields {
		n := len(goName) + len(f.Name) + 3 // "Tag"
		if n > max {
			max = n
		}
	}
	return max
}

// maxNameOfLen returns the max length of "NameOf{goName}{field.Name}" across all fields.
func maxNameOfLen(goName string, fields []FieldTpl) int {
	max := 0
	for _, f := range fields {
		n := 6 + len(goName) + len(f.Name) // "NameOf"
		if n > max {
			max = n
		}
	}
	return max
}

// maxEnumValLen returns the max length of EnumValueGoName(val.Name) across all enum values.
func maxEnumValLen(vals []protofile.EnumValue) int {
	max := 0
	for _, v := range vals {
		n := len(EnumValueGoName(v.Name))
		if n > max {
			max = n
		}
	}
	return max
}

// writerFieldNameMax returns the max length of all field names in a writer struct,
// including the terminal arena field.
func writerFieldNameMax(fields []FieldTpl) int {
	max := len("arena")
	for _, f := range fields {
		if n := len(f.Name); n > max {
			max = n
		}
	}
	return max
}

// readerFieldNameMax returns the max length of all field names in a reader struct,
// including hidden _nameArr, _hasName, and rawBufferLen fields.
func readerFieldNameMax(fields []FieldTpl) int {
	max := len("rawBufferLen")
	for _, f := range fields {
		if f.IsRawBuf {
			continue
		}
		if n := len(f.Name); n > max {
			max = n
		}
		if f.Map && f.MapValIsMsg {
			if n := 1 + len(f.Name) + 3; n > max { // "_" + name + "Arr"
				max = n
			}
		}
		if f.IsMsg && f.IsRecursive {
			if n := 4 + len(f.Name); n > max { // "_has" + name
				max = n
			}
		}
	}
	return max
}

// writerFieldType returns the Go type string for a writer struct field,
// matching what go.tpl generates.
func (g *Generator) writerFieldType(f FieldTpl) string {
	if f.IsDecimal {
		return "decimal.Decimal"
	}
	if f.Map && f.MapValIsMsg {
		kgt, _, _ := g.ProtoTypeToGo(f.MapKey, false)
		vgt, _, _ := g.ProtoTypeToGo(f.MapVal, false)
		return "map[" + kgt + "]*" + vgt
	}
	if f.IsMsg && f.IsRecursive {
		return "*" + f.GoType
	}
	return f.GoType
}

// readerFieldType returns the Go type string for a reader struct field,
// matching what go.tpl generates.
func (g *Generator) readerFieldType(f FieldTpl) string {
	if f.IsRawBuf {
		return "int"
	}
	if f.IsDecimal {
		return "decimal.Decimal"
	}
	if f.Map && f.MapValIsMsg {
		kgt, _, _ := g.ProtoTypeToGo(f.MapKey, false)
		vgt := protofile.ReadonlyGoTypeName(f.MapVal)
		return "map[" + kgt + "]*" + vgt
	}
	if f.IsMsg && f.IsRecursive {
		return "*" + f.ReaderType
	}
	return f.ReaderType
}

// writerFieldTypeMax returns the max type-string length for tagged writer struct fields.
func (g *Generator) writerFieldTypeMax(fields []FieldTpl) int {
	max := 0
	for _, f := range fields {
		if f.StructTag == "" {
			continue
		}
		if n := len(g.writerFieldType(f)); n > max {
			max = n
		}
	}
	return max
}

// readerFieldTypeMax returns the max type-string length for tagged reader struct fields.
func (g *Generator) readerFieldTypeMax(fields []FieldTpl) int {
	max := 0
	for _, f := range fields {
		if f.IsRawBuf || f.StructTag == "" {
			continue
		}
		if n := len(g.readerFieldType(f)); n > max {
			max = n
		}
	}
	return max
}

// ─── per-group struct-field alignment ────────────────────────────────────────

// alignLine is an intermediate field-line descriptor used by computeAlignedFields.
type alignLine struct {
	comment   []string
	name      string
	typeStr   string
	structTag string
}

// AlignedField is a pre-computed struct-field descriptor with per-group column widths.
// The flat list produced by writerAlignedFields / readerAlignedFields contains one element
// per rendered struct line (including hidden _nameArr / _hasName lines and the terminal
// arena field).  NameW and TypeW match what go/printer's tabwriter would compute, so
// the generated code needs no additional gofmt pass.
type AlignedField struct {
	Comment   []string
	Name      string
	TypeStr   string
	StructTag string // empty for untagged fields
	NameW     int    // max name len in this name-column group
	TypeW     int    // max type len in this type-column sub-group (0 for untagged)
}

// computeAlignedFields assigns NameW and TypeW to each line following tabwriter rules:
//   - Name-column groups break when a line has a non-empty comment (the comment line
//     precedes the field and emits a 1-column line that resets tabwriter alignment).
//   - Within each name group, type-column sub-groups break at untagged lines
//     (StructTag == ""), mirroring how gofmt breaks the type-column alignment group
//     at fields with no struct tag.
func computeAlignedFields(lines []alignLine) []AlignedField {
	n := len(lines)
	nameW := make([]int, n)
	typeW := make([]int, n)

	// Pass 1: name-column groups.
	groupStart := 0
	for i := 1; i <= n; i++ {
		if i == n || len(lines[i].comment) > 0 {
			maxLen := 0
			for j := groupStart; j < i; j++ {
				if l := len(lines[j].name); l > maxLen {
					maxLen = l
				}
			}
			for j := groupStart; j < i; j++ {
				nameW[j] = maxLen
			}
			groupStart = i
		}
	}

	// Pass 2: type-column sub-groups within each name group.
	groupStart = 0
	for i := 1; i <= n; i++ {
		if i == n || len(lines[i].comment) > 0 {
			alignTypeSubgroups(lines, typeW, groupStart, i)
			groupStart = i
		}
	}

	out := make([]AlignedField, n)
	for i, l := range lines {
		out[i] = AlignedField{
			Comment:   l.comment,
			Name:      l.name,
			TypeStr:   l.typeStr,
			StructTag: l.structTag,
			NameW:     nameW[i],
			TypeW:     typeW[i],
		}
	}
	return out
}

// alignTypeSubgroups fills typeW[start:end] for the lines in a single name-column group.
func alignTypeSubgroups(lines []alignLine, typeW []int, start, end int) {
	sgStart := start
	for i := start; i <= end; i++ {
		if i == end || lines[i].structTag == "" {
			if i > sgStart {
				maxLen := 0
				for j := sgStart; j < i; j++ {
					if l := len(lines[j].typeStr); l > maxLen {
						maxLen = l
					}
				}
				for j := sgStart; j < i; j++ {
					typeW[j] = maxLen
				}
			}
			if i < end {
				typeW[i] = 0 // untagged
			}
			sgStart = i + 1
		}
	}
}

// writerAlignedFields returns a flat AlignedField list for all lines in the writer struct
// body: visible fields, hidden _nameArr lines (for map-of-msg fields), and the terminal
// arena field.  NameW and TypeW are computed per tabwriter alignment group.
func (g *Generator) writerAlignedFields(fields []FieldTpl) []AlignedField {
	var lines []alignLine
	for _, f := range fields {
		lines = append(lines, alignLine{
			comment:   f.Comment,
			name:      f.Name,
			typeStr:   g.writerFieldType(f),
			structTag: f.StructTag,
		})
	}
	lines = append(lines, alignLine{name: "arena", typeStr: "[]byte"})
	return computeAlignedFields(lines)
}

// readerAlignedFields returns a flat AlignedField list for all lines in the reader
// (Readonly) struct body: rawBufferLen, visible fields, hidden _nameArr / _hasName lines.
// It processes .ReaderFields which already includes the synthetic rawBufferLen entry.
func (g *Generator) readerAlignedFields(fields []FieldTpl) []AlignedField {
	var lines []alignLine
	for _, f := range fields {
		if f.IsRawBuf {
			lines = append(lines, alignLine{name: "rawBufferLen", typeStr: "int"})
			continue
		}
		lines = append(lines, alignLine{
			comment:   f.Comment,
			name:      f.Name,
			typeStr:   g.readerFieldType(f),
			structTag: f.StructTag,
		})
		if f.Map && f.MapValIsMsg {
			valGt := protofile.ReadonlyGoTypeName(f.MapVal)
			lines = append(lines, alignLine{
				name:    "_" + f.Name + "Arr",
				typeStr: "[]" + valGt,
			})
		}
		if f.IsMsg && f.IsRecursive {
			lines = append(lines, alignLine{
				name:    "_has" + f.Name,
				typeStr: "bool",
			})
		}
	}
	return computeAlignedFields(lines)
}

// ─── template data types ──────────────────────────────────────────────────────

type FieldTpl struct {
	protofile.FieldDef
	WireType           WireTypeVal
	ReaderType         string
	IsRawBuf           bool   // synthetic rawBufferLen int field for readonly structs
	StructTag          string // pre-computed struct tag, e.g. `json:"foo,omitempty" yaml:"foo"`
	MapValIsMsg        bool   // true when MapVal is a message type (not scalar/enum)
	MapValMsgIsAsArray bool   // true when MapVal message has @AsArray — JSON value is an array, not an object
	ElemIsRecursive    bool   // true when map/repeated element msg type cycles back to the containing message
	IsDecimal          bool   // true when field has @decimal=round:N annotation (double → decimal.Decimal)
	MsgIsAsArray       bool   // true when IsMsg and the referenced message has @AsArray — JSON value is an array
}

type MsgTpl struct {
	Name          string     // proto message name (used as map/lookup key)
	GoName        string     // Go type name: same as proto message name
	Comment       []string   // proto comment lines (without leading //)
	Fields        []FieldTpl // writer fields, sorted for optimal layout
	ReverseFields []FieldTpl // writer fields in reverse layout order (matches marshalToSizedBufferVT output)
	ReaderFields  []FieldTpl // readonly fields = Fields + rawBufferLen, all sorted
	AsMap         bool       // true when @AsMap annotation is present: single map field, JSON parsed as direct map
	AsArray       bool       // true when @AsArray annotation is present: single repeated field, JSON parsed as direct array
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

// canReachViaAllEdges reports whether message type 'start' can reach 'target'
// following all field types — plain msg, repeated elem, map value — in the
// message graph.  Used to detect runtime infinite-recursion in test builders.
func canReachViaAllEdges(messages map[string]*protofile.MessageDef, start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}
	if visited[start] {
		return false
	}
	visited[start] = true
	md, ok := messages[start]
	if !ok {
		return false
	}
	for _, fd := range md.Fields {
		var next string
		if fd.IsMsg && !fd.Map {
			next = fd.Type
		} else if fd.Map && fd.MapVal != "" {
			if _, ok := messages[fd.MapVal]; ok {
				next = fd.MapVal
			}
		}
		if next != "" && canReachViaAllEdges(messages, next, target, visited) {
			return true
		}
	}
	return false
}

func (g *Generator) makeFieldTpl(fd protofile.FieldDef, containingMsgName string) FieldTpl {
	var wt WireTypeVal
	if fd.Map || fd.Repeated {
		wt = WireTypeLenDelim
	} else {
		wt = WireType(fd.Type, fd.IsMsg)
	}
	var mapValIsMsg bool
	if fd.Map && fd.MapVal != "" {
		_, isMsg, _ := g.ProtoTypeToGo(fd.MapVal, false)
		mapValIsMsg = isMsg
	}
	// Detect whether a map-value or repeated-element message type can reach the
	// containing message through any chain of fields (including maps and slices).
	// Such fields would cause infinite recursion if their builder calls the
	// containing message's builder, so they need an independent "Base" builder.
	var elemIsRecursive bool
	if containingMsgName != "" {
		if fd.Map && fd.MapVal != "" {
			_, isMsg, _ := g.ProtoTypeToGo(fd.MapVal, false)
			if isMsg {
				visited := map[string]bool{}
				elemIsRecursive = canReachViaAllEdges(g.Messages, fd.MapVal, containingMsgName, visited)
			}
		} else if fd.IsMsg && fd.Repeated {
			visited := map[string]bool{}
			elemIsRecursive = canReachViaAllEdges(g.Messages, fd.Type, containingMsgName, visited)
		}
	}
	isDecimal := fd.DecimalRound > 0
	if isDecimal && fd.Type != "double" {
		panic(fmt.Sprintf("@decimal annotation on field %q: only allowed on double fields, got %q", fd.Name, fd.Type))
	}
	var msgIsAsArray bool
	if fd.IsMsg {
		if msgDef, ok := g.Messages[fd.Type]; ok {
			msgIsAsArray = msgDef.AsArray
		}
	}
	var mapValMsgIsAsArray bool
	if fd.Map && fd.MapVal != "" && mapValIsMsg {
		if msgDef, ok := g.Messages[fd.MapVal]; ok {
			mapValMsgIsAsArray = msgDef.AsArray
		}
	}
	return FieldTpl{
		FieldDef:           fd,
		WireType:           wt,
		ReaderType:         g.readerGoType(fd),
		StructTag:          buildStructTag(fd),
		MapValIsMsg:        mapValIsMsg,
		MapValMsgIsAsArray: mapValMsgIsAsArray,
		ElemIsRecursive:    elemIsRecursive,
		IsDecimal:          isDecimal,
		MsgIsAsArray:       msgIsAsArray,
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
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment, AsMap: md.AsMap, AsArray: md.AsArray}

		// --- Writer struct: sort using precomputed writer layouts for IsMsg fields.
		writerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // recursive field stored as a pointer
			}
			if fd.IsMsg {
				if li, ok := writerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		writerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // pointer word
			}
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
			mt.Fields = append(mt.Fields, g.makeFieldTpl(fd, name))
		}

		// --- ReverseFields: writer fields in reverse layout order, matching marshalToSizedBufferVT output.
		for j := len(mt.Fields) - 1; j >= 0; j-- {
			mt.ReverseFields = append(mt.ReverseFields, mt.Fields[j])
		}

		// --- Readonly struct: include rawBufferLen in the sort, and use precomputed
		// readonly layouts for IsMsg fields (readonly types are larger due to rawBufferLen).
		rawBufDef := protofile.FieldDef{Name: "rawBufferLen", Type: "int64", GoType: "int"}
		readerDefs := make([]protofile.FieldDef, 0, len(md.Fields)+1)
		readerDefs = append(readerDefs, rawBufDef)
		readerDefs = append(readerDefs, md.Fields...)

		readerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // recursive field stored as a pointer
			}
			if fd.IsMsg {
				if li, ok := readerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		readerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // pointer word
			}
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
					ReaderType: "int",
					IsRawBuf:   true,
				})
			} else {
				mt.ReaderFields = append(mt.ReaderFields, g.makeFieldTpl(fd, name))
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
		"zeroVal":             ZeroVal,
		"readerZero":          ReaderZero,
		"isPackable":          IsPackable,
		"is8ByteNumeric":      Is8ByteNumeric,
		"isSliceType":         func(s string) bool { return strings.HasPrefix(s, "[]") },
		"readFunc":            ReadFuncForType,
		"protoWireType":       ProtoWireType,
		"trimPtr":             func(s string) string { return strings.TrimPrefix(s, "*") },
		"mapKeyGoType":        func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValGoType":        func(s string) string { gt, _, _ := g.ProtoTypeToGo(s, false); return gt },
		"mapValIsMsg":         func(s string) bool { _, isMsg, _ := g.ProtoTypeToGo(s, false); return isMsg },
		"mapValIsEnum":        func(s string) bool { _, _, isEnum := g.ProtoTypeToGo(s, false); return isEnum },
		"upperFirst":          protofile.UpperFirst,
		"enumValueGoName":     EnumValueGoName,
		"padRight":            padRight,
		"maxTagLen":           maxTagLen,
		"maxNameOfLen":        maxNameOfLen,
		"maxEnumValLen":       maxEnumValLen,
		"writerFieldNameMax":  writerFieldNameMax,
		"readerFieldNameMax":  readerFieldNameMax,
		"writerFieldType":     g.writerFieldType,
		"readerFieldType":     g.readerFieldType,
		"writerFieldTypeMax":  g.writerFieldTypeMax,
		"readerFieldTypeMax":  g.readerFieldTypeMax,
		"writerAlignedFields": g.writerAlignedFields,
		"readerAlignedFields": g.readerAlignedFields,
		"readerElemType": func(fd FieldTpl) string {
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
		"hasMapsOrMsgs":              HasMapsOrMsgs,
		"hasDecimalFields":           HasDecimalFields,
		"anyMsgHasDecimalField":      AnyMsgHasDecimalField,
		"hasStringOrBytesFields":     HasStringOrBytesFields,
		"decimalRound":             func(f FieldTpl) int { return f.DecimalRound },
		// tagSize computes utils.TagSize(fieldNum, wireType) at template generation time,
		// so the generated code contains the literal integer instead of a runtime call.
		"tagSize": func(fieldNum, wireType int) int {
			tag := uint64(fieldNum<<3 | wireType)
			return (bits.Len64(tag|1) + 6) / 7
		},
		// protoWireTypeInt returns the integer wire-type value for a proto scalar type name.
		"protoWireTypeInt": func(pt string) int {
			return int(ProtoWireType(pt))
		},
		// writeTagBytes returns Go statements that write the field tag backward in a buffer.
		"writeTagBytes": func(fieldNum, wireType int) string {
			var wtName string
			switch wireType {
			case 0:
				wtName = "Varint"
			case 1:
				wtName = "64bit"
			case 2:
				wtName = "LenDelim"
			case 5:
				wtName = "32bit"
			default:
				wtName = fmt.Sprintf("wt%d", wireType)
			}
			tag := uint64(fieldNum<<3 | wireType)
			if tag < 0x80 {
				return fmt.Sprintf("i--\n\t\tdAtA[i] = %d /*field=%d, wireType=%s, (%d<<3)|%d (%d)*/",
					byte(tag), fieldNum, wtName, fieldNum, wireType, tag)
			}
			if tag < 0x4000 {
				lo := byte(tag&0x7f | 0x80)
				hi := byte(tag >> 7)
				return fmt.Sprintf(
					"i--\n"+
						"\t\tdAtA[i] = %d /*field=%d, wireType=%s, (%d<<3)|%d=%d, byte[1]=%d>>7 (%d)*/\n"+
						"\t\ti--\n"+
						"\t\tdAtA[i] = %d /*byte[0]=%d&0x7f|0x80 (%d)*/",
					hi, fieldNum, wtName, fieldNum, wireType, tag, tag, hi,
					lo, tag, lo)
			}
			return fmt.Sprintf("i = utils.EncodeVarint(dAtA, i, %d) /*field=%d, wireType=%s, (%d<<3)|%d (%d)*/",
				tag, fieldNum, wtName, fieldNum, wireType, tag)
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
// makeSampleXxxBase() struct literal.  Recursive fields (IsRecursive or
// ElemIsRecursive) return "nil" so the Base function stays recursion-free.
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
		if ft.MapValIsMsg {
			if ft.ElemIsRecursive {
				return "nil" // recursive map value; populated by makeSampleXxx using Base
			}
			// Struct field is map[keyType]*MsgType; use an IIFE to build a pointer entry.
			msgGoName := protofile.GoTypeName(ft.MapVal)
			keyGoType, ok := protofile.ScalarProtoToGo[ft.MapKey]
			if !ok {
				keyGoType = ft.MapKey // string or bool
			}
			return fmt.Sprintf(
				`(func() map[%s]*%s { v := makeSample%s(); return map[%s]*%s{%s: &v} }())`,
				keyGoType, msgGoName, msgGoName, keyGoType, msgGoName, keyLit)
		}
		var valLit string
		if ft.MapVal == "bool" {
			valLit = "true"
		} else {
			valLit = SampleScalarLiteral(ft.MapVal, protofile.GoTypeName(ft.MapVal))
		}
		return fmt.Sprintf("%s{%s: %s}", ft.GoType, keyLit, valLit)
	}
	if ft.Repeated {
		if ft.IsMsg && ft.ElemIsRecursive {
			return "nil" // recursive repeated elem; populated by makeSampleXxx using Base
		}
		if ft.IsMsg {
			elemType := strings.TrimPrefix(ft.GoType, "[]")
			return fmt.Sprintf("%s{makeSample%s()}", ft.GoType, elemType)
		}
		elemLit := SampleScalarLiteral(ft.Type, strings.TrimPrefix(ft.GoType, "[]"))
		return fmt.Sprintf("%s{%s}", ft.GoType, elemLit)
	}
	if ft.IsMsg {
		if ft.IsRecursive {
			return "nil" // recursive pointer field; avoid infinite recursion in sample generation
		}
		return fmt.Sprintf("makeSample%s()", ft.GoType)
	}
	if ft.IsDecimal {
		return "decimal.MustParse(\"1.12345\")"
	}
	return SampleScalarLiteral(ft.Type, ft.GoType)
}

// SampleRecursiveFieldLiteral returns the Go expression for a recursive field's
// value using the independent "Base" builder, to be used in the full
// makeSampleXxx() function after calling makeSampleXxxBase().
func SampleRecursiveFieldLiteral(ft FieldTpl) string {
	if ft.IsMsg && ft.IsRecursive {
		return fmt.Sprintf("(func() *%s { v := makeSample%sBase(); return &v }())", ft.GoType, ft.GoType)
	}
	if ft.Map && ft.MapValIsMsg && ft.ElemIsRecursive {
		msgGoName := protofile.GoTypeName(ft.MapVal)
		var keyLit string
		switch ft.MapKey {
		case "string":
			keyLit = `"k"`
		case "bool":
			return "nil"
		default:
			keyLit = SampleScalarLiteral(ft.MapKey, "")
		}
		keyGoType, ok := protofile.ScalarProtoToGo[ft.MapKey]
		if !ok {
			keyGoType = ft.MapKey
		}
		return fmt.Sprintf(
			`(func() map[%s]*%s { v := makeSample%sBase(); return map[%s]*%s{%s: &v} }())`,
			keyGoType, msgGoName, msgGoName, keyGoType, msgGoName, keyLit)
	}
	if ft.Repeated && ft.IsMsg && ft.ElemIsRecursive {
		elemType := strings.TrimPrefix(ft.GoType, "[]")
		return fmt.Sprintf("%s{makeSample%sBase()}", ft.GoType, elemType)
	}
	return "nil"
}

// HasAnyRecursiveField returns true when any field is a direct recursive
// message pointer (IsRecursive) or a map/repeated whose element type cycles
// back to the containing message (ElemIsRecursive).
func HasAnyRecursiveField(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.IsRecursive || f.ElemIsRecursive {
			return true
		}
	}
	return false
}

// BenchMapFillRecursive generates a single-entry fill for a recursive map
// field using the independent "Base" builder, avoiding infinite recursion.
func BenchMapFillRecursive(ft FieldTpl) string {
	msgGoName := protofile.GoTypeName(ft.MapVal)
	switch ft.MapKey {
	case "string":
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[strconv.Itoa(0)] = &v }`, msgGoName)
	case "int32", "sint32", "sfixed32":
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[int32(0)] = &v }`, msgGoName)
	case "uint32", "fixed32":
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[uint32(0)] = &v }`, msgGoName)
	case "int64", "sint64", "sfixed64":
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[int64(0)] = &v }`, msgGoName)
	case "uint64", "fixed64":
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[uint64(0)] = &v }`, msgGoName)
	default:
		return fmt.Sprintf(`{ v := benchBuild%sBase(); m[0] = &v }`, msgGoName)
	}
}

// BenchSliceFillRecursive generates a single-element fill for a recursive
// repeated field using the independent "Base" builder.
func BenchSliceFillRecursive(ft FieldTpl) string {
	elemType := strings.TrimPrefix(ft.GoType, "[]")
	return fmt.Sprintf(`s[0] = benchBuild%sBase()`, elemType)
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
// non-rawBufferLen, non-decimal) field, or nil when no such field exists. The returned field
// is used by the test template to generate JSON-type-error tests.
func FirstScalarField(fields []FieldTpl) *FieldTpl {
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && !f.IsDecimal {
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

// HasMapsOrMsgs returns true if any field is a map or embedded message.
// Used to decide whether ToProtobuf should delegate to ToProtobufByAppend
// (complex types) or ToProtobufVT (simple scalars and repeated slices).
func HasMapsOrMsgs(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.Map || f.IsMsg {
			return true
		}
	}
	return false
}

// HasDecimalFields returns true when any field in the list uses @decimal annotation.
func HasDecimalFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.IsDecimal {
			return true
		}
	}
	return false
}

// AnyMsgHasDecimalField returns true when any message in the list has a decimal field.
func AnyMsgHasDecimalField(msgs []MsgTpl) bool {
	for _, m := range msgs {
		if HasDecimalFields(m.Fields) {
			return true
		}
	}
	return false
}

// HasStringOrBytesFields returns true when any field would cause ReadString or ReadBytes
// to be called in FromProtobuf — plain string/bytes fields, repeated string/bytes,
// or map fields with a string/bytes key or value.
func HasStringOrBytesFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.IsRawBuf {
			continue
		}
		if f.Map {
			if f.MapKey == "string" || f.MapVal == "string" || f.MapVal == "bytes" {
				return true
			}
			if f.MapValIsMsg {
				return true
			}
			continue
		}
		if f.Type == "string" || f.Type == "bytes" {
			return true
		}
		if f.IsMsg {
			return true
		}
	}
	return false
}

// ScalarStringFields returns non-map, non-repeated, non-msg string fields.
// Used to generate memory-stomping tests that detect unsafe.String aliasing.
func ScalarStringFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && f.Type == "string" {
			out = append(out, *f)
		}
	}
	return out
}

// ScalarBytesFields returns non-map, non-repeated, non-msg bytes fields.
func ScalarBytesFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for i := range fields {
		f := &fields[i]
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && f.Type == "bytes" {
			out = append(out, *f)
		}
	}
	return out
}

// RepeatedStringFields returns repeated (non-map) string fields.
func RepeatedStringFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for i := range fields {
		f := &fields[i]
		if f.Repeated && !f.Map && f.Type == "string" {
			out = append(out, *f)
		}
	}
	return out
}

// RepeatedBytesFields returns repeated (non-map) bytes fields.
func RepeatedBytesFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for i := range fields {
		f := &fields[i]
		if f.Repeated && !f.Map && f.Type == "bytes" {
			out = append(out, *f)
		}
	}
	return out
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
// fields are considered. Decimal fields are excluded (they use a separate test).
func NumericBoundaryCases(fields []FieldTpl) []BoundaryCase {
	var out []BoundaryCase
	for _, f := range fields {
		if f.Map || f.Repeated || f.IsMsg || f.IsRawBuf || f.IsDecimal {
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
// Decimal fields are excluded.
func HasNumericBoundaryFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if f.Map || f.Repeated || f.IsMsg || f.IsRawBuf || f.IsDecimal {
			continue
		}
		if len(boundaryLitsForType(f.Type)) > 0 {
			return true
		}
	}
	return false
}

// HasFloatFields returns true if any direct scalar float or double field exists (excluding decimal).
func HasFloatFields(fields []FieldTpl) bool {
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && !f.IsDecimal {
			if f.Type == "float" || f.Type == "double" {
				return true
			}
		}
	}
	return false
}

// FloatFields returns all direct scalar float and double fields (excluding decimal).
func FloatFields(fields []FieldTpl) []FieldTpl {
	var out []FieldTpl
	for _, f := range fields {
		if !f.Map && !f.Repeated && !f.IsMsg && !f.IsRawBuf && !f.IsDecimal {
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
// (not map, not repeated, not msg, not rawBufferLen), or nil.
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

// AnyMsgHasBytesField returns true if any message contains a field whose sample
// literal requires the utils package (i.e. bytes fields).
func AnyMsgHasBytesField(msgs []MsgTpl) bool {
	for _, msg := range msgs {
		for _, f := range msg.Fields {
			if f.Type == "bytes" {
				return true
			}
			if f.Map && f.MapVal == "bytes" {
				return true
			}
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
		// Enum type – cast integer 1 to the Go enum type.
		// (Message types are handled by BenchMapFill before calling this function.)
		return fmt.Sprintf("%s(1)", protofile.GoTypeName(mapVal))
	}
}

// BenchMapFill generates a Go statement (or block of statements) that fills
// the local variable `m` (already declared with the correct map type) with
// 101 representative entries.  For bool keys only two entries are possible.
func BenchMapFill(ft FieldTpl) string {
	if ft.MapValIsMsg {
		// Map value is a message type; struct field uses *MsgType as value.
		msgGoName := protofile.GoTypeName(ft.MapVal)
		switch ft.MapKey {
		case "string":
			return fmt.Sprintf(`for i := 0; i < 101; i++ { v := benchBuild%s(); m[strconv.Itoa(i)] = &v }`, msgGoName)
		case "bool":
			return fmt.Sprintf("{ v1 := benchBuild%s(); m[false] = &v1; v2 := benchBuild%s(); m[true] = &v2 }", msgGoName, msgGoName)
		case "int32", "sint32", "sfixed32":
			return fmt.Sprintf(`for i := int32(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }`, msgGoName)
		case "uint32", "fixed32":
			return fmt.Sprintf(`for i := uint32(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }`, msgGoName)
		case "int64", "sint64", "sfixed64":
			return fmt.Sprintf(`for i := int64(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }`, msgGoName)
		case "uint64", "fixed64":
			return fmt.Sprintf(`for i := uint64(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }`, msgGoName)
		default:
			return fmt.Sprintf(`for i := int32(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }`, msgGoName)
		}
	}
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

// MapWriterGoType returns the correct Go map type for use in the writer struct
// context. When the map value is a message type, the value is a pointer
// (map[K]*MsgType); for all other types it returns ft.GoType unchanged.
func MapWriterGoType(ft FieldTpl) string {
	if ft.MapValIsMsg {
		keyGoType, ok := protofile.ScalarProtoToGo[ft.MapKey]
		if !ok {
			keyGoType = ft.MapKey // string or bool
		}
		msgGoName := protofile.GoTypeName(ft.MapVal)
		return fmt.Sprintf("map[%s]*%s", keyGoType, msgGoName)
	}
	return ft.GoType
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

// BenchNeedsUtils returns true when any message in msgs has a map field
// with bytes value or a repeated bytes field, both of which call
// utils.UnsafeBytesFromString in the generated fill loops.
func BenchNeedsUtils(msgs []MsgTpl) bool {
	for _, msg := range msgs {
		for _, f := range msg.Fields {
			if f.Map && f.MapVal == "bytes" {
				return true
			}
			if f.Repeated && f.Type == "bytes" {
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
//   - rawBufferLen synthetic field
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
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment, AsMap: md.AsMap, AsArray: md.AsArray}

		writerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // recursive field stored as a pointer
			}
			if fd.IsMsg {
				if li, ok := writerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		writerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // pointer word
			}
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
			mt.Fields = append(mt.Fields, g.makeFieldTpl(fd, name))
		}

		for j := len(mt.Fields) - 1; j >= 0; j-- {
			mt.ReverseFields = append(mt.ReverseFields, mt.Fields[j])
		}

		rawBufDef := protofile.FieldDef{Name: "rawBufferLen", Type: "int64", GoType: "int"}
		readerDefs := make([]protofile.FieldDef, 0, len(md.Fields)+1)
		readerDefs = append(readerDefs, rawBufDef)
		readerDefs = append(readerDefs, md.Fields...)

		readerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // recursive field stored as a pointer
			}
			if fd.IsMsg {
				if li, ok := readerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		readerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // pointer word
			}
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
					ReaderType: "int",
					IsRawBuf:   true,
				})
			} else {
				mt.ReaderFields = append(mt.ReaderFields, g.makeFieldTpl(fd, name))
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
		"sampleLitFull":            SampleRecursiveFieldLiteral,
		"hasAnyRecursiveField":     HasAnyRecursiveField,
		"hasLargeIntFields":        HasLargeIntFields,
		"largeIntFields":           LargeIntFields,
		"largeIntLit":              LargeIntLit,
		"skipEncodingJSON":         SkipEncodingJSON,
		"firstScalarField":         FirstScalarField,
		"hasMapsOrSlices":          HasMapsOrSlices,
		"hasMapsOrMsgs":            HasMapsOrMsgs,
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
		"anyMsgHasBytesField":      AnyMsgHasBytesField,
		"hasFloatFields":           HasFloatFields,
		"floatFields":              FloatFields,
		"floatIntLit":              FloatIntLit,
		"anyMsgHasDecimalField":    AnyMsgHasDecimalField,
		"hasDecimalFields":         HasDecimalFields,
		"decimalFields":            func(fields []FieldTpl) []FieldTpl {
			var out []FieldTpl
			for _, f := range fields {
				if f.IsDecimal {
					out = append(out, f)
				}
			}
			return out
		},
		"scalarStringFields":   ScalarStringFields,
		"scalarBytesFields":    ScalarBytesFields,
		"repeatedStringFields": RepeatedStringFields,
		"repeatedBytesFields":  RepeatedBytesFields,
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
		mt := MsgTpl{Name: md.Name, GoName: protofile.GoTypeName(md.Name), Comment: md.Comment, AsMap: md.AsMap, AsArray: md.AsArray}

		writerSizeOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // recursive field stored as a pointer
			}
			if fd.IsMsg {
				if li, ok := writerLayouts[fd.Type]; ok {
					return li.Size
				}
			}
			return protofile.FieldGoSize(fd)
		}
		writerPtrdataOf := func(fd protofile.FieldDef) int {
			if fd.IsMsg && fd.IsRecursive {
				return 8 // pointer word
			}
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
			mt.Fields = append(mt.Fields, g.makeFieldTpl(fd, name))
		}

		for j := len(mt.Fields) - 1; j >= 0; j-- {
			mt.ReverseFields = append(mt.ReverseFields, mt.Fields[j])
		}

		msgs = append(msgs, mt)
	}

	data := RenderData{
		Package:  g.Pkg,
		Enums:    enums,
		Messages: msgs,
	}

	fnMap := template.FuncMap{
		"sampleLit":               SampleFieldLiteral,
		"benchMapFill":            BenchMapFill,
		"benchSliceFill":          BenchSliceFill,
		"benchNeedsStrconv":       BenchNeedsStrconv,
		"benchNeedsUtils":         BenchNeedsUtils,
		"mapWriterGoType":         MapWriterGoType,
		"hasAnyRecursiveField":    HasAnyRecursiveField,
		"benchMapFillRecursive":   BenchMapFillRecursive,
		"benchSliceFillRecursive": BenchSliceFillRecursive,
		"anyMsgHasDecimalField":   AnyMsgHasDecimalField,
	}
	tmpl, err := template.New("pb_timing_test").Funcs(fnMap).Parse(benchTemplate)
	if err != nil {
		return fmt.Errorf("parse bench template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// CompareRenderData is the template data for the compare test file.
type CompareRenderData struct {
	Package  string
	Messages []MsgTpl
}

// RenderCompare renders the performance-comparison test file to out.
// It generates one Test_<Msg>_with_compare function per message, measuring
// JSON and Protobuf encode/decode throughput against standard library alternatives.
// The generated file depends on the benchBuildXxx helpers from the _timing_test.go file.
func (g *Generator) RenderCompare(out *os.File) error {
	var msgs []MsgTpl
	for _, name := range g.Order {
		md := g.Messages[name]
		msgs = append(msgs, MsgTpl{
			Name:   md.Name,
			GoName: protofile.GoTypeName(md.Name),
			AsMap:   md.AsMap,
			AsArray: md.AsArray,
		})
	}
	data := CompareRenderData{
		Package:  g.Pkg,
		Messages: msgs,
	}
	tmpl, err := template.New("compare_test").Parse(compareTemplate)
	if err != nil {
		return fmt.Errorf("parse compare template: %w", err)
	}
	return tmpl.Execute(out, data)
}

// ─── code template ────────────────────────────────────────────────────────────

//go:embed templates/message.go.tpl
var codeTemplate string

//go:embed templates/test.go.tpl
var testTemplate string

//go:embed templates/timing.test.go.tpl
var benchTemplate string

//go:embed templates/compare.test.go.tpl
var compareTemplate string
