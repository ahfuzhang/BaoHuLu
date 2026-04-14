package protofile

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/emicklei/proto"
)

// ─── data model ───────────────────────────────────────────────────────────────

type EnumValue struct {
	Name   string
	Number int
}

type EnumDef struct {
	Name   string
	Values []EnumValue
}

type FieldDef struct {
	Name     string
	JsonName string // original proto field name, used as json tag key
	Number   int
	Type     string // proto type string
	GoType   string // Go type for writer struct
	Repeated bool
	Map      bool
	MapKey   string // proto key type
	MapVal   string // proto val type
	IsMsg    bool   // is embedded message
	IsEnum   bool
	Comment  []string // proto comment lines (without leading //)
}

type MessageDef struct {
	Name    string
	Fields  []FieldDef
	Comment []string // proto comment lines (without leading //)
}

// ─── generator ────────────────────────────────────────────────────────────────

type Generator struct {
	Pkg         string // effective Go package name used in the template
	PackageName string // proto "package" statement value
	GoPackage   string // proto "option go_package" value (full import path)
	Enums       map[string]*EnumDef
	Messages    map[string]*MessageDef
	Order       []string
}

func NewGenerator(pkg string) *Generator {
	return &Generator{
		Pkg:      pkg,
		Enums:    make(map[string]*EnumDef),
		Messages: make(map[string]*MessageDef),
	}
}

// ParseAndCollect parses a proto definition and collects all enums and messages.
// pkg is the fallback package name used when the proto file contains neither a
// "package" statement nor an "option go_package" option.
func ParseAndCollect(r io.Reader, pkg string) (*Generator, error) {
	parser := proto.NewParser(r)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	g := NewGenerator(pkg)
	g.Collect(definition)
	return g, nil
}

func (g *Generator) Collect(def *proto.Proto) {
	proto.Walk(def,
		// Capture the proto "package" statement.
		proto.WithPackage(func(p *proto.Package) {
			g.PackageName = p.Name
		}),
		// Capture file-level options; specifically "go_package".
		proto.WithOption(func(o *proto.Option) {
			if _, ok := o.Parent.(*proto.Proto); !ok {
				return // skip message/service-level options
			}
			if o.Name == "go_package" {
				// Literal.Source carries the raw text including surrounding quotes.
				g.GoPackage = strings.Trim(o.Constant.Source, `"`)
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			ed := &EnumDef{Name: e.Name}
			for _, el := range e.Elements {
				if ev, ok := el.(*proto.EnumField); ok {
					ed.Values = append(ed.Values, EnumValue{Name: ev.Name, Number: ev.Integer})
				}
			}
			g.Enums[e.Name] = ed
		}),
		proto.WithMessage(func(m *proto.Message) {
			g.CollectMessage(m)
		}),
	)
	// Compute the effective Go package name from the proto declarations.
	// Priority: go_package last segment > package name > fallback passed to NewGenerator.
	if g.GoPackage != "" {
		parts := strings.Split(g.GoPackage, "/")
		g.Pkg = parts[len(parts)-1]
	} else if g.PackageName != "" {
		g.Pkg = g.PackageName
	}
}

// SnakeToCamel converts a snake_case proto field name to PascalCase Go name.
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func ExtractCommentLines(c *proto.Comment) []string {
	if c == nil {
		return nil
	}
	return c.Lines
}

func (g *Generator) CollectMessage(m *proto.Message) {
	md := &MessageDef{Name: m.Name, Comment: ExtractCommentLines(m.Comment)}
	for _, el := range m.Elements {
		switch v := el.(type) {
		case *proto.NormalField:
			fd := FieldDef{
				Name:     SnakeToCamel(v.Name),
				JsonName: v.Name,
				Number:   v.Sequence,
				Type:     v.Type,
				Repeated: v.Repeated,
				Comment:  ExtractCommentLines(v.Comment),
			}
			fd.GoType, fd.IsMsg, fd.IsEnum = g.ProtoTypeToGo(v.Type, v.Repeated)
			md.Fields = append(md.Fields, fd)
		case *proto.MapField:
			keyGo, _, _ := g.ProtoTypeToGo(v.KeyType, false)
			valGo, _, _ := g.ProtoTypeToGo(v.Type, false)
			fd := FieldDef{
				Name:     SnakeToCamel(v.Name),
				JsonName: v.Name,
				Number:   v.Sequence,
				Type:     "map",
				GoType:   fmt.Sprintf("map[%s]%s", keyGo, valGo),
				Map:      true,
				MapKey:   v.KeyType,
				MapVal:   v.Type,
				Comment:  ExtractCommentLines(v.Comment),
			}
			md.Fields = append(md.Fields, fd)
		}
	}
	g.Messages[m.Name] = md
	g.Order = append(g.Order, m.Name)
}

var ScalarProtoToGo = map[string]string{
	"double":   "float64",
	"float":    "float32",
	"int32":    "int32",
	"int64":    "int64",
	"uint32":   "uint32",
	"uint64":   "uint64",
	"sint32":   "int32",
	"sint64":   "int64",
	"fixed32":  "uint32",
	"fixed64":  "uint64",
	"sfixed32": "int32",
	"sfixed64": "int64",
	"bool":     "bool",
	"string":   "string",
	"bytes":    "[]byte",
}

func (g *Generator) ProtoTypeToGo(pt string, repeated bool) (goType string, isMsg bool, isEnum bool) {
	if gt, ok := ScalarProtoToGo[pt]; ok {
		if repeated {
			return "[]" + gt, false, false
		}
		return gt, false, false
	}
	if _, ok := g.Enums[pt]; ok {
		if repeated {
			return "[]" + pt, false, true
		}
		return pt, false, true
	}
	// Message type: strip "Message" suffix for a concise Go name.
	goName := GoTypeName(pt)
	if repeated {
		return "[]" + goName, true, false
	}
	return goName, true, false
}

func (g *Generator) EnumOrder() []string {
	var names []string
	for k := range g.Enums {
		names = append(names, k)
	}
	return names
}

// ─── name utilities ───────────────────────────────────────────────────────────

func UpperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// CamelToSnake converts a PascalCase or camelCase name to snake_case.
// e.g. "DemoServer" → "demo_server", "demo_server" → "demo_server".
func CamelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GoTypeName strips the conventional "Message" suffix from a proto message name
// to produce a concise Go type name.
func GoTypeName(protoName string) string {
	return strings.TrimSuffix(protoName, "Message")
}

func ReadonlyGoTypeName(protoName string) string {
	return "Readonly" + GoTypeName(protoName)
}

// ─── struct layout optimisation ──────────────────────────────────────────────
// Sorting rules mirror golang.org/x/tools/go/analysis/passes/fieldalignment:
//  1. Zero-sized fields first.
//  2. Higher alignment before lower alignment (reduces padding).
//  3. Pointer-containing fields before pointer-free fields (shrinks GC scan range).
//  4. Among pointerful fields: fewer trailing non-pointer bytes first.
//  5. Larger fields before smaller fields.
//  6. Original tag number as tiebreaker (deterministic output).

// FieldAlignment returns the memory alignment (bytes) of a field on 64-bit platforms.
func FieldAlignment(fd FieldDef) int {
	if fd.Map || fd.Repeated || fd.IsMsg {
		return 8 // map pointer, slice header, or struct with pointer-sized fields
	}
	switch fd.Type {
	case "double", "int64", "uint64", "sint64", "fixed64", "sfixed64", "string", "bytes":
		return 8
	case "float", "int32", "uint32", "sint32", "fixed32", "sfixed32":
		return 4
	case "bool":
		return 1
	default:
		if fd.IsEnum {
			return 4 // enum is int32
		}
		return 8
	}
}

// FieldGoSize returns the in-memory size (bytes) of a field on 64-bit platforms.
func FieldGoSize(fd FieldDef) int {
	if fd.Map {
		return 8 // map is a single pointer word
	}
	if fd.Repeated {
		return 24 // slice header: ptr(8)+len(8)+cap(8)
	}
	if fd.IsMsg {
		return 24 // fallback; use computed size if available
	}
	switch fd.Type {
	case "double", "int64", "uint64", "sint64", "fixed64", "sfixed64":
		return 8
	case "string":
		return 16 // ptr(8)+len(8)
	case "bytes":
		return 24 // []byte slice header
	case "float", "int32", "uint32", "sint32", "fixed32", "sfixed32":
		return 4
	case "bool":
		return 1
	default:
		if fd.IsEnum {
			return 4
		}
		return 8
	}
}

// FieldPtrdata returns the number of bytes the GC must scan for pointers on 64-bit platforms.
func FieldPtrdata(fd FieldDef) int {
	if fd.Map {
		return 8 // map is a single pointer
	}
	if fd.Repeated {
		return 8 // slice header: only the first word (data ptr) is a GC pointer
	}
	if fd.IsMsg {
		return 8 // fallback
	}
	switch fd.Type {
	case "string":
		return 8 // ptr(8)+len(8); only first word is a GC pointer
	case "bytes":
		return 8 // ptr(8)+len(8)+cap(8); only first word is a GC pointer
	}
	return 0 // all scalar proto types contain no GC pointers
}

// MsgLayoutInfo stores the computed sizeof and ptrdata for a generated struct.
type MsgLayoutInfo struct {
	Size    int
	Ptrdata int
}

// ComputeStructLayout walks sorted fields and returns the exact sizeof and
// ptrdata for the resulting struct on 64-bit platforms.
func ComputeStructLayout(
	fields []FieldDef,
	sizeOf func(FieldDef) int,
	ptrdataOf func(FieldDef) int,
) MsgLayoutInfo {
	offset, maxPtrEnd, maxAlign := 0, 0, 1
	for _, f := range fields {
		sz := sizeOf(f)
		al := FieldAlignment(f)
		offset = (offset + al - 1) &^ (al - 1)
		if al > maxAlign {
			maxAlign = al
		}
		if pd := ptrdataOf(f); pd > 0 {
			if end := offset + pd; end > maxPtrEnd {
				maxPtrEnd = end
			}
		}
		offset += sz
	}
	offset = (offset + maxAlign - 1) &^ (maxAlign - 1)
	return MsgLayoutInfo{Size: offset, Ptrdata: maxPtrEnd}
}

// SortFieldsWithCallbacks reorders fields using the fieldalignment algorithm.
func SortFieldsWithCallbacks(fields []FieldDef, sizeOf, ptrdataOf func(FieldDef) int) []FieldDef {
	out := make([]FieldDef, len(fields))
	copy(out, fields)
	sort.SliceStable(out, func(i, j int) bool {
		fi, fj := out[i], out[j]

		szi, szj := sizeOf(fi), sizeOf(fj)

		// Rule 1: zero-sized fields first.
		if zeroi, zeroj := szi == 0, szj == 0; zeroi != zeroj {
			return zeroi
		}
		// Rule 2: higher alignment first (minimises padding).
		ali, alj := FieldAlignment(fi), FieldAlignment(fj)
		if ali != alj {
			return ali > alj
		}
		// Rule 3: pointer-containing fields before pointer-free fields.
		ptri, ptrj := ptrdataOf(fi), ptrdataOf(fj)
		if noptri, noptrj := ptri == 0, ptrj == 0; noptri != noptrj {
			return noptrj // i has pointers → i goes first
		}
		// Rule 4: among pointerful fields, fewer trailing non-pointer bytes first.
		if ptri > 0 {
			traili, trailj := szi-ptri, szj-ptrj
			if traili != trailj {
				return traili < trailj
			}
		}
		// Rule 5: larger fields first.
		if szi != szj {
			return szi > szj
		}
		// Rule 6: stable tiebreaker — original proto tag number.
		return out[i].Number < out[j].Number
	})
	return out
}

// SortFieldsForLayout sorts fields using static size/ptrdata estimates.
func SortFieldsForLayout(fields []FieldDef) []FieldDef {
	return SortFieldsWithCallbacks(fields, FieldGoSize, FieldPtrdata)
}
