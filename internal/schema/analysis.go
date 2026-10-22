package schema

import (
	"fmt"
	"strings"
)

func (document *Schema) Resolve() error {
	resolve := func(raw, scope string) *TypeRef {
		if IsScalar(raw) {
			return &TypeRef{RawName: raw, Kind: ScalarKind, SymbolID: raw}
		}
		var candidates []string
		if !strings.HasPrefix(raw, ".") {
			for scope != "" {
				candidates = append(candidates, scope+"."+raw)
				last := strings.LastIndex(scope, ".")
				if last < 0 {
					break
				}
				scope = scope[:last]
			}
		}
		candidates = append(candidates, strings.TrimPrefix(raw, "."))
		for _, name := range candidates {
			if message := document.MessagesByName[name]; message != nil {
				return &TypeRef{RawName: raw, Kind: MessageKind, SymbolID: message.ID, Message: message}
			}
			if enum := document.EnumsByName[name]; enum != nil {
				return &TypeRef{RawName: raw, Kind: EnumKind, SymbolID: enum.ID, Enum: enum}
			}
		}
		return &TypeRef{RawName: raw, Kind: ExternalKind, SymbolID: strings.TrimPrefix(raw, ".")}
	}
	for _, message := range document.Messages {
		for _, field := range message.Fields {
			field.Type = resolve(field.Type.RawName, message.FullName)
			if field.Map {
				field.MapKey = resolve(field.MapKey.RawName, message.FullName)
			}
			if field.Deprecated || message.Deprecated {
				continue
			}
			if field.Type.Kind == ExternalKind {
				document.Warnings = append(document.Warnings, fmt.Sprintf("%s: external type %q must be provided separately", field.Location, field.Type.RawName))
			}
			field.WireType = WireType(field.Type)
			field.Packable = field.Type.Kind == EnumKind || field.Type.Kind == ScalarKind && field.Type.RawName != "string" && field.Type.RawName != "bytes"
			if field.Map || field.Repeated && field.Packable {
				field.WireType = 2
			}
			field.TagBytes = TagBytes(field.Number, field.WireType)
		}
	}
	for _, message := range document.Messages {
		for _, field := range message.ActiveFields() {
			if !message.Deprecated && field.Type.Message != nil && !field.Type.Message.Generate() {
				return fmt.Errorf("%s: field references empty message %s, which does not generate code", field.Location, field.Type.RawName)
			}
		}
	}
	for _, service := range document.Services {
		seen := map[string]bool{}
		for _, method := range service.Methods {
			method.Request = resolve(method.Request.RawName, document.File.Package)
			method.Response = resolve(method.Response.RawName, document.File.Package)
			for _, reference := range []*TypeRef{method.Request, method.Response} {
				if reference.Kind == ScalarKind || reference.Kind == EnumKind {
					return fmt.Errorf("%s: RPC %s requires message types, got %s", method.Location, method.Name, reference.RawName)
				}
				if reference.Message != nil && reference.Message.Deprecated {
					return fmt.Errorf("%s: RPC %s references deprecated message %s", method.Location, method.Name, reference.RawName)
				}
				if reference.Message != nil && !reference.Message.Generate() {
					return fmt.Errorf("%s: RPC %s references empty message %s, which does not generate code", method.Location, method.Name, reference.RawName)
				}
				if reference.Kind == ExternalKind {
					document.Warnings = append(document.Warnings, fmt.Sprintf("%s: external RPC type %q must be provided separately", method.Location, reference.RawName))
				}
				if !seen[reference.SymbolID] {
					service.ReferencedTypes = append(service.ReferencedTypes, reference)
					seen[reference.SymbolID] = true
				}
			}
		}
	}
	document.analyze()
	return nil
}

func reachable(graph map[string][]string, start, target string, seen map[string]bool) bool {
	if start == target {
		return true
	}
	if seen[start] {
		return false
	}
	seen[start] = true
	for _, next := range graph[start] {
		if reachable(graph, next, target, seen) {
			return true
		}
	}
	return false
}

func (document *Schema) analyze() {
	analysis := &Analysis{
		Dependencies: map[string][]string{}, ValueDependencies: map[string][]string{},
		YAML: map[string]bool{}, URLValues: map[string]bool{},
	}
	for _, message := range document.Messages {
		if message.Deprecated {
			continue
		}
		analysis.YAML[message.ID] = message.YAML
		analysis.URLValues[message.ID] = message.URLValues
		seen := map[string]bool{}
		for _, field := range message.ActiveFields() {
			if field.Type.Kind != MessageKind && field.Type.Kind != ExternalKind {
				continue
			}
			if !seen[field.Type.SymbolID] {
				analysis.Dependencies[message.ID] = append(analysis.Dependencies[message.ID], field.Type.SymbolID)
				seen[field.Type.SymbolID] = true
			}
			if !field.Map && !field.Repeated {
				analysis.ValueDependencies[message.ID] = append(analysis.ValueDependencies[message.ID], field.Type.SymbolID)
			}
		}
	}
	for _, message := range document.Messages {
		for _, field := range message.ActiveFields() {
			if !field.Map && !field.Repeated && field.Type.Kind == MessageKind {
				field.Recursive = reachable(analysis.ValueDependencies, field.Type.SymbolID, message.ID, map[string]bool{})
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, message := range document.Messages {
			if message.Deprecated {
				continue
			}
			for _, field := range message.ActiveFields() {
				if field.Type.Kind != MessageKind {
					continue
				}
				if analysis.YAML[message.ID] && !analysis.YAML[field.Type.SymbolID] {
					analysis.YAML[field.Type.SymbolID] = true
					changed = true
				}
				if analysis.URLValues[message.ID] && !field.Map && !analysis.URLValues[field.Type.SymbolID] {
					analysis.URLValues[field.Type.SymbolID] = true
					changed = true
				}
			}
		}
	}
	visited := map[string]bool{}
	var visit func(*Message)
	visit = func(message *Message) {
		if visited[message.ID] || message.Deprecated {
			return
		}
		visited[message.ID] = true
		for _, field := range message.ActiveFields() {
			if field.Type.Message != nil && !field.Map && !field.Repeated && !field.Recursive {
				visit(field.Type.Message)
			}
		}
		analysis.DependencyOrder = append(analysis.DependencyOrder, message)
	}
	for _, message := range document.Messages {
		visit(message)
	}
	document.Analysis = analysis
}
