package schema

import "fmt"

type Location struct {
	File   string
	Line   int
	Column int
}

func (location Location) String() string {
	return fmt.Sprintf("%s:%d:%d", location.File, location.Line, location.Column)
}

type Annotation struct {
	Name     string
	Value    string
	Location Location
}

type Switch struct {
	Value    bool
	Explicit bool
}

type File struct {
	Name             string
	Source           string
	Syntax           string
	Package          string
	Options          map[string]string
	OptionValues     map[string]SourceSpan
	Imports          []string
	Annotations      []Annotation
	GenTestCode      Switch
	GenBenchmarkCode Switch
	GenProjectFile   Switch
}

type SourceSpan struct {
	Start int
	End   int
}

type Schema struct {
	File           *File
	Messages       []*Message
	Enums          []*Enum
	Services       []*Service
	Declarations   []Declaration
	MessagesByName map[string]*Message
	EnumsByName    map[string]*Enum
	Analysis       *Analysis
	Warnings       []string
}

type Declaration struct {
	Message *Message
	Enum    *Enum
	Service *Service
}

type TypeKind string

const (
	ScalarKind   TypeKind = "scalar"
	MessageKind  TypeKind = "message"
	EnumKind     TypeKind = "enum"
	ExternalKind TypeKind = "external"
)

type TypeRef struct {
	RawName  string
	Kind     TypeKind
	SymbolID string
	Message  *Message
	Enum     *Enum
}

type Message struct {
	ID           string
	Name         string
	FullName     string
	Location     Location
	Comments     []string
	CommentLines []int
	Annotations  []Annotation
	Fields       []*Field
	Deprecated   bool
	AsMap        bool
	AsArray      bool
	YAML         bool
	URLValues    bool
}

type Field struct {
	Name              string
	Number            int
	Location          Location
	Comments          []string
	CommentLines      []int
	InlineComments    []string
	InlineCommentLine int
	Annotations       []Annotation
	Options           map[string]string
	Type              *TypeRef
	MapKey            *TypeRef
	Repeated          bool
	Map               bool
	Deprecated        bool
	DecimalRound      int
	Recursive         bool
	WireType          int
	TagBytes          []byte
	Packable          bool
}

type Enum struct {
	ID          string
	Name        string
	FullName    string
	Location    Location
	Comments    []string
	Annotations []Annotation
	Values      []*EnumValue
}

type EnumValue struct {
	Name     string
	Number   int
	Location Location
}

type Service struct {
	ID              string
	Name            string
	FullName        string
	Location        Location
	Comments        []string
	Annotations     []Annotation
	Methods         []*RPC
	ReferencedTypes []*TypeRef
}

type RPC struct {
	ID              string
	Name            string
	FullName        string
	Location        Location
	Comments        []string
	Annotations     []Annotation
	Service         *Service
	Request         *TypeRef
	Response        *TypeRef
	Paths           []string
	ClientStreaming bool
	ServerStreaming bool
}

type Analysis struct {
	Dependencies      map[string][]string
	ValueDependencies map[string][]string
	DependencyOrder   []*Message
	YAML              map[string]bool
	URLValues         map[string]bool
}

func (message *Message) ActiveFields() []*Field {
	fields := make([]*Field, 0, len(message.Fields))
	for _, field := range message.Fields {
		if field.Deprecated || field.Type.Message != nil && field.Type.Message.Deprecated {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func (message *Message) Generate() bool {
	return !message.Deprecated && len(message.ActiveFields()) > 0
}

func IsScalar(name string) bool {
	switch name {
	case "double", "float", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool", "string", "bytes":
		return true
	}
	return false
}

func ValidMapKey(name string) bool {
	return IsScalar(name) && name != "double" && name != "float" && name != "bytes"
}

func WireType(reference *TypeRef) int {
	if reference.Kind == MessageKind || reference.Kind == ExternalKind {
		return 2
	}
	switch reference.RawName {
	case "double", "fixed64", "sfixed64":
		return 1
	case "string", "bytes":
		return 2
	case "float", "fixed32", "sfixed32":
		return 5
	}
	return 0
}

func TagBytes(number, wireType int) []byte {
	value := uint64(number)<<3 | uint64(wireType)
	var result []byte
	for value >= 128 {
		result = append(result, byte(value)|128)
		value >>= 7
	}
	return append(result, byte(value))
}
