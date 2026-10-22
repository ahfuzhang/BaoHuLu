package codegen

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/ahfuzhang/BaoHuLu/internal/schema"
)

const maxOutputBytes = 64 << 20

type limitedBuffer struct {
	bytes.Buffer
}

func (buffer *limitedBuffer) Write(contents []byte) (int, error) {
	if len(contents) > maxOutputBytes-buffer.Len() {
		return 0, fmt.Errorf("template output exceeds %d bytes", maxOutputBytes)
	}
	return buffer.Buffer.Write(contents)
}

func (bundle *Bundle) Render(outputs []Output) ([]Output, error) {
	var compiled *template.Template
	depth := 0
	var current RenderContext
	type functionResult struct {
		value    any
		returned bool
	}
	var frames []*functionResult
	type memoKey struct {
		Name     string
		Schema   *schema.Schema
		File     *schema.File
		Analysis *schema.Analysis
		Options  Options
	}
	memoized := map[memoKey]any{}
	pending := map[memoKey]bool{}
	helperNames := map[string]bool{}
	functions := template.FuncMap{
		"withRPC":     WithRPC,
		"withTypeRef": WithTypeRef,
		"lower":       strings.ToLower,
		"upper":       strings.ToUpper,
		"trimPrefix":  strings.TrimPrefix,
		"trimSuffix":  strings.TrimSuffix,
		"split":       strings.Split,
		"join":        strings.Join,
		"replace":     strings.ReplaceAll,
		"list":        func(values ...any) []any { return values },
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires key/value pairs")
			}
			result := map[string]any{}
			for index := 0; index < len(values); index += 2 {
				key, ok := values[index].(string)
				if !ok {
					return nil, fmt.Errorf("dict key must be a string")
				}
				result[key] = values[index+1]
			}
			return result, nil
		},
		"fail": func(message string) (string, error) { return "", fmt.Errorf("%s", message) },
	}
	for name, function := range genericFunctions() {
		functions[name] = function
	}
	functions["return"] = func(value any) (string, error) {
		if len(frames) == 0 {
			return "", fmt.Errorf("return is only valid inside a template function")
		}
		frame := frames[len(frames)-1]
		if frame.returned {
			return "", fmt.Errorf("template function returned multiple values")
		}
		frame.value, frame.returned = value, true
		return "", nil
	}
	functions["memo"] = func(name string) (any, error) {
		if !helperNames[name] {
			return nil, fmt.Errorf("memo requires a template function: %s", name)
		}
		key := memoKey{Name: name, Schema: current.Schema, File: current.File, Analysis: current.Analysis, Options: current.Options}
		if value, exists := memoized[key]; exists {
			return value, nil
		}
		if pending[key] {
			return nil, fmt.Errorf("recursive memo computation: %s", name)
		}
		pending[key] = true
		defer delete(pending, key)
		previous := current
		current = RenderContext{Schema: key.Schema, File: key.File, Analysis: key.Analysis, Options: key.Options}
		defer func() { current = previous }()
		value, err := functions[name].(func(...any) (any, error))()
		if err == nil {
			memoized[key] = value
		}
		return value, err
	}
	functionNames := regexp.MustCompile(`\{\{-?\s*define\s+"fn/([A-Za-z][A-Za-z0-9_]*)"`)
	for _, contents := range bundle.Files {
		for _, match := range functionNames.FindAllSubmatch(contents, -1) {
			name := string(match[1])
			if helperNames[name] {
				continue
			}
			if _, exists := functions[name]; exists {
				return nil, fmt.Errorf("template function %s conflicts with a built-in function", name)
			}
			for _, builtin := range strings.Fields("and call html index slice js len not or print printf println urlquery eq ge gt le lt ne render") {
				if name == builtin {
					return nil, fmt.Errorf("template function %s conflicts with a built-in function", name)
				}
			}
			helperNames[name] = true
			functions[name] = func(arguments ...any) (any, error) {
				if len(frames) >= 64 {
					return nil, fmt.Errorf("template function recursion limit exceeded")
				}
				context := current
				context.Args = arguments
				frame := &functionResult{}
				frames = append(frames, frame)
				defer func() { frames = frames[:len(frames)-1] }()
				var buffer limitedBuffer
				if err := compiled.ExecuteTemplate(&buffer, "fn/"+name, context); err != nil {
					return nil, err
				}
				if frame.returned {
					if strings.TrimSpace(buffer.String()) != "" {
						return nil, fmt.Errorf("template function %s mixed returned data and text", name)
					}
					return frame.value, nil
				}
				return buffer.String(), nil
			}
		}
	}
	functions["render"] = func(name string, context RenderContext) (string, error) {
		if depth >= 64 {
			return "", fmt.Errorf("render recursion limit exceeded")
		}
		depth++
		defer func() { depth-- }()
		previous := current
		current = context
		defer func() { current = previous }()
		var buffer limitedBuffer
		if err := compiled.ExecuteTemplate(&buffer, name, context); err != nil {
			return "", err
		}
		return buffer.String(), nil
	}
	var err error
	compiled, err = bundle.compile(functions)
	if err != nil {
		return nil, err
	}
	result := make([]Output, 0, len(outputs))
	for _, output := range outputs {
		current = output.Context
		var buffer limitedBuffer
		if err := compiled.ExecuteTemplate(&buffer, output.Template, output.Context); err != nil {
			return nil, fmt.Errorf("render %s -> %s: %w", output.Template, output.Path, err)
		}
		if strings.TrimSpace(buffer.String()) == "" {
			continue
		}
		output.Content = append([]byte(nil), buffer.Bytes()...)
		result = append(result, output)
	}
	return result, nil
}
