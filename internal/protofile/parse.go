package protofile

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/ahfuzhang/BaoHuLu/internal/protoextensions"
	"github.com/ahfuzhang/BaoHuLu/internal/schema"
	"github.com/emicklei/proto"
)

func Parse(reader io.Reader, filename string) (*schema.Schema, error) {
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}
	parser := proto.NewParser(strings.NewReader(string(source)))
	definition, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	syntaxLine := 0
	for _, element := range definition.Elements {
		if syntax, ok := element.(*proto.Syntax); ok {
			if syntaxLine != 0 {
				return nil, fmt.Errorf("%s: duplicate syntax", filename)
			}
			if syntax.Value != "proto3" {
				return nil, fmt.Errorf("%s: only proto3 syntax is supported", filename)
			}
			syntaxLine = syntax.Position.Line
		}
	}
	if syntaxLine == 0 {
		return nil, fmt.Errorf("%s: syntax = \"proto3\" is required", filename)
	}
	file, err := protoextensions.ParseFile(string(source), filename, syntaxLine)
	if err != nil {
		return nil, err
	}
	file.Syntax = "proto3"
	file.OptionValues = map[string]schema.SourceSpan{}
	document := &schema.Schema{File: file, MessagesByName: map[string]*schema.Message{}, EnumsByName: map[string]*schema.Enum{}}
	for _, element := range definition.Elements {
		switch declaration := element.(type) {
		case *proto.Package:
			if file.Package != "" {
				return nil, fmt.Errorf("%s: duplicate package", filename)
			}
			file.Package = declaration.Name
		case *proto.Option:
			if _, exists := file.Options[declaration.Name]; exists {
				return nil, fmt.Errorf("%s: duplicate option %s", filename, declaration.Name)
			}
			file.Options[declaration.Name] = declaration.Constant.Source
			span, err := optionValueSpan(string(source), declaration.Position.Offset)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, declaration.Position.Line, err)
			}
			file.OptionValues[declaration.Name] = span
		case *proto.Import:
			file.Imports = append(file.Imports, declaration.Filename)
			document.Warnings = append(document.Warnings, fmt.Sprintf("%s:%d: import %q is not resolved; external types must be supplied separately", filename, declaration.Position.Line, declaration.Filename))
		}
	}
	collector := &schemaCollector{document: document, symbols: map[string]schema.Location{}}
	if err := collector.collect(definition.Elements, file.Package); err != nil {
		return nil, err
	}
	if err := document.Resolve(); err != nil {
		return nil, err
	}
	return document, nil
}

func optionValueSpan(source string, offset int) (schema.SourceSpan, error) {
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(source[offset:]))
	lexer.Error = func(*scanner.Scanner, string) {}
	span := schema.SourceSpan{Start: -1}
	assigned, depth := false, 0
	for token := lexer.Scan(); token != scanner.EOF; token = lexer.Scan() {
		if !assigned {
			assigned = token == '='
			continue
		}
		if token == ';' && depth == 0 {
			break
		}
		if span.Start < 0 {
			span.Start = offset + lexer.Position.Offset
		}
		switch token {
		case '{', '[', '(':
			depth++
		case '}', ']', ')':
			depth--
		}
		span.End = offset + lexer.Position.Offset + len(lexer.TokenText())
	}
	if span.Start < 0 || span.End <= span.Start || depth != 0 {
		return schema.SourceSpan{}, fmt.Errorf("cannot locate option value")
	}
	return span, nil
}

type schemaCollector struct {
	document *schema.Schema
	symbols  map[string]schema.Location
}

func (collector *schemaCollector) location(position scanner.Position) schema.Location {
	return schema.Location{File: collector.document.File.Name, Line: position.Line, Column: position.Column}
}

func fullName(scope, name string) string {
	if scope == "" {
		return name
	}
	return scope + "." + name
}

func (collector *schemaCollector) declare(name string, location schema.Location) error {
	if previous, exists := collector.symbols[name]; exists {
		return fmt.Errorf("%s: duplicate declaration %s (first at %s)", location, name, previous)
	}
	collector.symbols[name] = location
	return nil
}

func (collector *schemaCollector) comments(comment *proto.Comment) ([]string, []schema.Annotation) {
	var lines []string
	var annotations []schema.Annotation
	if comment == nil {
		return lines, annotations
	}
	for offset, line := range comment.Lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			lines = append(lines, line)
			continue
		}
		name, value, _ := strings.Cut(strings.TrimPrefix(trimmed, "@"), "=")
		location := collector.location(comment.Position)
		location.Line += offset
		annotations = append(annotations, schema.Annotation{Name: strings.ToLower(strings.TrimSpace(name)), Value: strings.TrimSpace(value), Location: location})
	}
	return lines, annotations
}

func commentLines(comment *proto.Comment) []int {
	var lines []int
	if comment != nil {
		for offset, line := range comment.Lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "@") {
				lines = append(lines, comment.Position.Line+offset)
			}
		}
	}
	return lines
}

func (collector *schemaCollector) collect(elements []proto.Visitee, scope string) error {
	for _, element := range elements {
		switch declaration := element.(type) {
		case *proto.Message:
			location := collector.location(declaration.Position)
			if declaration.IsExtend {
				return fmt.Errorf("%s: extend is not supported", location)
			}
			message := &schema.Message{Name: declaration.Name, FullName: fullName(scope, declaration.Name), Location: location}
			message.ID = message.FullName
			if err := collector.declare(message.ID, location); err != nil {
				return err
			}
			message.Comments, message.Annotations = collector.comments(declaration.Comment)
			message.CommentLines = commentLines(declaration.Comment)
			for _, annotation := range message.Annotations {
				switch annotation.Name {
				case "deprecated":
					message.Deprecated = true
				case "asmap":
					message.AsMap = true
				case "asarray":
					message.AsArray = true
				case "yaml":
					message.YAML = true
				case "urlvalues":
					message.URLValues = true
				}
			}
			collector.document.Messages = append(collector.document.Messages, message)
			collector.document.MessagesByName[message.FullName] = message
			collector.document.Declarations = append(collector.document.Declarations, schema.Declaration{Message: message})
			if err := collector.fields(declaration, message); err != nil {
				return err
			}
			if err := collector.collect(declaration.Elements, message.FullName); err != nil {
				return err
			}
		case *proto.Enum:
			enum := &schema.Enum{Name: declaration.Name, FullName: fullName(scope, declaration.Name), Location: collector.location(declaration.Position)}
			enum.ID = enum.FullName
			if err := collector.declare(enum.ID, enum.Location); err != nil {
				return err
			}
			enum.Comments, enum.Annotations = collector.comments(declaration.Comment)
			names, numbers := map[string]bool{}, map[int]bool{}
			allowAlias := false
			for _, child := range declaration.Elements {
				if option, ok := child.(*proto.Option); ok && option.Name == "allow_alias" {
					allowAlias = option.Constant.Source == "true"
				}
			}
			for _, child := range declaration.Elements {
				if value, ok := child.(*proto.EnumField); ok {
					if names[value.Name] || numbers[value.Integer] && !allowAlias {
						return fmt.Errorf("%s: duplicate enum name or number in %s", collector.location(value.Position), enum.Name)
					}
					if value.Integer < -2147483648 || value.Integer > 2147483647 {
						return fmt.Errorf("%s: enum value outside int32 range", collector.location(value.Position))
					}
					names[value.Name], numbers[value.Integer] = true, true
					enum.Values = append(enum.Values, &schema.EnumValue{Name: value.Name, Number: value.Integer, Location: collector.location(value.Position)})
				}
			}
			if len(enum.Values) > 0 && enum.Values[0].Number != 0 {
				return fmt.Errorf("%s: proto3 enum must start at zero", enum.Location)
			}
			collector.document.Enums = append(collector.document.Enums, enum)
			collector.document.EnumsByName[enum.FullName] = enum
			collector.document.Declarations = append(collector.document.Declarations, schema.Declaration{Enum: enum})
		case *proto.Service:
			if err := collector.service(declaration, scope); err != nil {
				return err
			}
		case *proto.Oneof:
			return fmt.Errorf("%s: oneof is not supported", collector.location(declaration.Position))
		case *proto.Extensions:
			return fmt.Errorf("%s: extensions is not supported", collector.location(declaration.Position))
		case *proto.Group:
			return fmt.Errorf("%s: group is not supported", collector.location(declaration.Position))
		case *proto.Option:
			if _, ok := declaration.Parent.(*proto.Message); ok {
				return fmt.Errorf("%s: option inside message is not supported", collector.location(declaration.Position))
			}
		}
	}
	return nil
}

func (collector *schemaCollector) fields(declaration *proto.Message, message *schema.Message) error {
	names, numbers := map[string]bool{}, map[int]bool{}
	for _, element := range declaration.Elements {
		var field *schema.Field
		var original *proto.Field
		switch definition := element.(type) {
		case *proto.NormalField:
			if definition.Optional || definition.Required {
				return fmt.Errorf("%s: optional/required is not supported", collector.location(definition.Position))
			}
			original = definition.Field
			field = &schema.Field{Type: &schema.TypeRef{RawName: definition.Type}, Repeated: definition.Repeated}
		case *proto.MapField:
			original = definition.Field
			if !schema.ValidMapKey(definition.KeyType) {
				return fmt.Errorf("%s: invalid map key %q", collector.location(definition.Position), definition.KeyType)
			}
			field = &schema.Field{Type: &schema.TypeRef{RawName: definition.Type}, MapKey: &schema.TypeRef{RawName: definition.KeyType}, Map: true}
		default:
			continue
		}
		field.Name, field.Number, field.Location = original.Name, original.Sequence, collector.location(original.Position)
		if field.Number <= 0 || field.Number > 536870911 || field.Number >= 19000 && field.Number <= 19999 {
			return fmt.Errorf("%s: invalid field number %d", field.Location, field.Number)
		}
		if names[field.Name] || numbers[field.Number] {
			return fmt.Errorf("%s: duplicate field name or number in %s", field.Location, message.Name)
		}
		names[field.Name], numbers[field.Number] = true, true
		field.Comments, field.Annotations = collector.comments(original.Comment)
		field.CommentLines = commentLines(original.Comment)
		if original.InlineComment != nil {
			field.InlineComments = append([]string{}, original.InlineComment.Lines...)
			field.InlineCommentLine = original.InlineComment.Position.Line
		}
		field.Options = map[string]string{}
		for _, option := range original.Options {
			field.Options[option.Name] = option.Constant.Source
		}
		for _, annotation := range field.Annotations {
			switch annotation.Name {
			case "deprecated":
				field.Deprecated = true
			case "decimal":
				prefix, raw, found := strings.Cut(annotation.Value, ":")
				round, err := strconv.Atoi(raw)
				if !found || prefix != "round" || err != nil || round <= 0 || field.Type.RawName != "double" || field.Map || field.Repeated {
					return fmt.Errorf("%s: @decimal requires round:N (N > 0) on a singular double", annotation.Location)
				}
				field.DecimalRound = round
			}
		}
		message.Fields = append(message.Fields, field)
	}
	if message.AsMap && (message.AsArray || len(message.Fields) != 1 || !message.Fields[0].Map) {
		return fmt.Errorf("%s: @AsMap requires exactly one map field and cannot combine with @AsArray", message.Location)
	}
	if message.AsArray && (len(message.Fields) != 1 || !message.Fields[0].Repeated) {
		return fmt.Errorf("%s: @AsArray requires exactly one repeated field", message.Location)
	}
	return nil
}

func (collector *schemaCollector) service(declaration *proto.Service, scope string) error {
	service := &schema.Service{Name: declaration.Name, FullName: fullName(scope, declaration.Name), Location: collector.location(declaration.Position)}
	service.ID = service.FullName
	if err := collector.declare(service.ID, service.Location); err != nil {
		return err
	}
	service.Comments, service.Annotations = collector.comments(declaration.Comment)
	names, paths := map[string]bool{}, map[string]bool{}
	for _, element := range declaration.Elements {
		definition, ok := element.(*proto.RPC)
		if !ok {
			if option, ok := element.(*proto.Option); ok {
				return fmt.Errorf("%s: option inside service is not supported", collector.location(option.Position))
			}
			continue
		}
		method := &schema.RPC{Name: definition.Name, FullName: service.FullName + "." + definition.Name, Service: service,
			Location: collector.location(definition.Position), Request: &schema.TypeRef{RawName: definition.RequestType}, Response: &schema.TypeRef{RawName: definition.ReturnsType},
			ClientStreaming: definition.StreamsRequest, ServerStreaming: definition.StreamsReturns}
		method.ID = method.FullName
		if method.ClientStreaming || method.ServerStreaming {
			return fmt.Errorf("%s: streaming RPC is not supported", method.Location)
		}
		if names[method.Name] {
			return fmt.Errorf("%s: duplicate RPC %s", method.Location, method.Name)
		}
		names[method.Name] = true
		method.Comments, method.Annotations = collector.comments(definition.Comment)
		for _, annotation := range method.Annotations {
			if annotation.Name == "path" {
				if annotation.Value == "" || paths[annotation.Value] {
					return fmt.Errorf("%s: empty or duplicate RPC path %q", annotation.Location, annotation.Value)
				}
				paths[annotation.Value] = true
				method.Paths = append(method.Paths, annotation.Value)
			}
		}
		service.Methods = append(service.Methods, method)
	}
	for _, method := range service.Methods {
		defaultPath := "/" + service.Name + "/" + method.Name
		if paths[defaultPath] {
			return fmt.Errorf("%s: explicit RPC path conflicts with implicit route %s", method.Location, defaultPath)
		}
	}
	collector.document.Services = append(collector.document.Services, service)
	collector.document.Declarations = append(collector.document.Declarations, schema.Declaration{Service: service})
	return nil
}
