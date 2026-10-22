package codegen

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ahfuzhang/BaoHuLu/internal/schema"
)

type Output struct {
	Path       string
	Template   string
	Part       string
	Context    RenderContext
	CreateOnly bool
	Content    []byte
}

func validateOutputPath(name string) error {
	if err := safeRelative(name); err != nil {
		return err
	}
	for _, part := range strings.Split(name, "/") {
		stem := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9' {
			return fmt.Errorf("reserved output path: %s", name)
		}
	}
	if strings.HasPrefix(name, ".hulu-") {
		return fmt.Errorf("reserved generator path: %s", name)
	}
	return nil
}

func (bundle *Bundle) Plan(document *schema.Schema, options Options) ([]Output, error) {
	if document == nil || document.File == nil || document.Analysis == nil {
		return nil, fmt.Errorf("a parsed, checked Schema is required")
	}
	for _, capability := range []struct {
		name    string
		setting schema.Switch
		count   int
	}{
		{"tests", document.File.GenTestCode, len(bundle.Config.Tests)},
		{"benchmarks", document.File.GenBenchmarkCode, len(bundle.Config.Benchmarks)},
		{"project", document.File.GenProjectFile, len(bundle.Config.Project)},
	} {
		if capability.setting.Explicit && capability.setting.Value && capability.count == 0 {
			return nil, fmt.Errorf("proto explicitly enables %s but template bundle has no %s mapping", capability.name, capability.name)
		}
	}
	base := RenderContext{Schema: document, File: document.File, Analysis: document.Analysis, Options: options}
	var outputs []Output
	seen := map[string]string{}
	add := func(name, templateFile, part string, context RenderContext) error {
		if templateFile == "" {
			return nil
		}
		if err := validateOutputPath(name); err != nil {
			return err
		}
		folded := strings.ToLower(name)
		if previous, exists := seen[folded]; exists {
			return fmt.Errorf("output collision: %s and %s", previous, name)
		}
		for existing := range seen {
			if strings.HasPrefix(folded, existing+"/") || strings.HasPrefix(existing, folded+"/") {
				return fmt.Errorf("output file/directory collision: %s", name)
			}
		}
		seen[folded] = name
		outputs = append(outputs, Output{Path: name, Template: templateFile, Part: part, Context: context, CreateOnly: part == "project"})
		return nil
	}
	for _, enum := range document.Enums {
		if bundle.Config.Enum == "" || len(enum.Values) == 0 {
			continue
		}
		suffix, err := extension(bundle.Config.Enum)
		if err != nil {
			return nil, err
		}
		context := base
		context.Enum = enum
		if err := add(enum.Name+suffix, bundle.Config.Enum, "enum", context); err != nil {
			return nil, err
		}
	}
	for _, message := range document.Messages {
		if !message.Generate() {
			continue
		}
		context := base
		context.Message = message
		if bundle.Config.Message != "" {
			suffix, err := extension(bundle.Config.Message)
			if err != nil {
				return nil, err
			}
			if err := add(message.Name+suffix, bundle.Config.Message, "message", context); err != nil {
				return nil, err
			}
		}
		for _, family := range []struct {
			name    string
			entries map[string]string
			enabled bool
		}{
			{"json", bundle.Config.JSON, true}, {"protobuf", bundle.Config.Protobuf, true},
			{"yaml", bundle.Config.YAML, document.Analysis.YAML[message.ID]},
			{"url_values", bundle.Config.URLValues, document.Analysis.URLValues[message.ID]},
			{"lifecycle", bundle.Config.Lifecycle, true},
		} {
			if !family.enabled {
				continue
			}
			for _, key := range sortedKeys(family.entries) {
				templateFile := family.entries[key]
				suffix, err := extension(templateFile)
				if err != nil {
					return nil, err
				}
				if err := add(message.Name+"."+key+suffix, templateFile, family.name, context); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, service := range document.Services {
		context := base
		context.Service = service
		if bundle.Config.Service != "" {
			name := service.Name + "." + strings.TrimSuffix(path.Base(bundle.Config.Service), ".tpl")
			if err := add(name, bundle.Config.Service, "service", context); err != nil {
				return nil, err
			}
		}
		if bundle.Config.RPC != "" {
			for _, method := range service.Methods {
				name := service.Name + "." + method.Name + "." + strings.TrimSuffix(path.Base(bundle.Config.RPC), ".tpl")
				if err := add(name, bundle.Config.RPC, "rpc", WithRPC(context, method)); err != nil {
					return nil, err
				}
			}
		}
	}
	basename := strings.TrimSuffix(filepath.Base(document.File.Name), filepath.Ext(document.File.Name))
	for _, family := range []struct {
		name    string
		entries map[string]string
		enabled bool
	}{
		{"tests", bundle.Config.Tests, document.File.GenTestCode.Value},
		{"benchmarks", bundle.Config.Benchmarks, document.File.GenBenchmarkCode.Value},
		{"project", bundle.Config.Project, document.File.GenProjectFile.Value},
	} {
		if !family.enabled {
			continue
		}
		for _, outputPath := range sortedKeys(family.entries) {
			if family.name == "project" && (strings.HasPrefix(outputPath, "Tests/") && !document.File.GenTestCode.Value || strings.HasPrefix(outputPath, "Benchmarks/") && !document.File.GenBenchmarkCode.Value) {
				continue
			}
			name := strings.ReplaceAll(outputPath, "{file}", basename)
			if strings.Contains(name, "{message}") {
				for _, message := range document.Messages {
					if !message.Generate() {
						continue
					}
					context := base
					context.Message = message
					if err := add(strings.ReplaceAll(name, "{message}", message.Name), family.entries[outputPath], family.name, context); err != nil {
						return nil, err
					}
				}
			} else if err := add(name, family.entries[outputPath], family.name, base); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(outputs, func(left, right int) bool { return outputs[left].Path < outputs[right].Path })
	return outputs, nil
}

func sortedKeys(entries map[string]string) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
