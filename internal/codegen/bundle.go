package codegen

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Message    string            `yaml:"message"`
	Enum       string            `yaml:"enum"`
	JSON       map[string]string `yaml:"json"`
	Protobuf   map[string]string `yaml:"protobuf"`
	YAML       map[string]string `yaml:"yaml"`
	URLValues  map[string]string `yaml:"url_values"`
	Lifecycle  map[string]string `yaml:"lifecycle"`
	Tests      map[string]string `yaml:"tests"`
	Benchmarks map[string]string `yaml:"benchmarks"`
	Project    map[string]string `yaml:"project"`
	Service    string            `yaml:"service"`
	RPC        string            `yaml:"rpc"`
}

type Bundle struct {
	Config   Config
	Files    map[string][]byte
	Identity string
}

func safeRelative(name string) error {
	if name == "" || name == "." || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00") {
		return fmt.Errorf("invalid relative path %q", name)
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("invalid path component %q", part)
		}
	}
	return nil
}

func readResource(resources fs.FS, name string) ([]byte, error) {
	if err := safeRelative(name); err != nil {
		return nil, err
	}
	parts := strings.Split(name, "/")
	for index, part := range parts {
		directory := "."
		if index > 0 {
			directory = strings.Join(parts[:index], "/")
		}
		entries, err := fs.ReadDir(resources, directory)
		if err != nil {
			return nil, err
		}
		found := false
		for _, entry := range entries {
			if entry.Name() != part {
				continue
			}
			found = true
			if entry.Type()&fs.ModeSymlink != 0 {
				return nil, fmt.Errorf("template symlink is not allowed: %s", name)
			}
			if index < len(parts)-1 && !entry.IsDir() {
				return nil, fmt.Errorf("template parent is not a directory: %s", directory)
			}
		}
		if !found {
			return nil, fmt.Errorf("template resource %s: %w", name, fs.ErrNotExist)
		}
	}
	return fs.ReadFile(resources, name)
}

func LoadBundle(resources fs.FS, identity string) (*Bundle, error) {
	contents, err := readResource(resources, "tpl_config.yaml")
	if err != nil {
		return nil, fmt.Errorf("load tpl_config.yaml: %w", err)
	}
	var node yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("tpl_config.yaml: %w", err)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("tpl_config.yaml must contain exactly one document")
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("tpl_config.yaml must be a mapping")
	}
	for _, value := range node.Content[0].Content {
		if value.Tag == "!!null" {
			return nil, fmt.Errorf("tpl_config.yaml:%d: null configuration is not allowed", value.Line)
		}
	}
	bundle := &Bundle{Identity: identity, Files: map[string][]byte{}}
	decoder = yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	if err := decoder.Decode(&bundle.Config); err != nil {
		return nil, fmt.Errorf("tpl_config.yaml: %w", err)
	}
	for _, family := range []struct {
		name    string
		entries map[string]string
		keys    []string
	}{
		{"json", bundle.Config.JSON, []string{"FromJSON", "ToJSON"}},
		{"protobuf", bundle.Config.Protobuf, []string{"FromProtobuf", "ToProtobuf"}},
		{"yaml", bundle.Config.YAML, []string{"FromYAML", "ToYAML"}},
		{"url_values", bundle.Config.URLValues, []string{"FromURLValues", "ToURLValues"}},
		{"lifecycle", bundle.Config.Lifecycle, []string{"Reset", "Clone"}},
	} {
		for key := range family.entries {
			valid := false
			for _, allowed := range family.keys {
				valid = valid || key == allowed
			}
			if !valid {
				return nil, fmt.Errorf("unknown configuration key %s.%s", family.name, key)
			}
		}
	}
	paths := map[string]bool{}
	for _, name := range []string{bundle.Config.Message, bundle.Config.Enum, bundle.Config.Service, bundle.Config.RPC} {
		if name != "" {
			paths[name] = true
		}
	}
	for index := 0; index < len(node.Content[0].Content); index += 2 {
		key, value := node.Content[0].Content[index], node.Content[0].Content[index+1]
		if value.Kind == yaml.ScalarNode && value.Value == "" {
			return nil, fmt.Errorf("empty template for %s", key.Value)
		}
	}
	for _, entries := range []map[string]string{bundle.Config.JSON, bundle.Config.Protobuf, bundle.Config.YAML, bundle.Config.URLValues, bundle.Config.Lifecycle, bundle.Config.Tests, bundle.Config.Benchmarks, bundle.Config.Project} {
		for key, name := range entries {
			if name == "" {
				return nil, fmt.Errorf("empty template for %s", key)
			}
			paths[name] = true
		}
	}
	for _, entries := range []map[string]string{bundle.Config.Tests, bundle.Config.Benchmarks, bundle.Config.Project} {
		for output := range entries {
			expanded := strings.NewReplacer("{message}", "Message", "{file}", "File").Replace(output)
			if strings.ContainsAny(expanded, "{}") {
				return nil, fmt.Errorf("unknown output placeholder: %s", output)
			}
			if err := validateOutputPath(expanded); err != nil {
				return nil, err
			}
		}
	}
	entries, err := fs.ReadDir(resources, ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() != "partials" {
			continue
		}
		if !entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("partials must be a directory, not a symlink")
		}
		err = fs.WalkDir(resources, "partials", func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&fs.ModeSymlink != 0 {
				return fmt.Errorf("template symlink is not allowed: %s", name)
			}
			if !entry.IsDir() && strings.HasSuffix(name, ".tpl") {
				paths[name] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for name := range paths {
		if !strings.HasSuffix(name, ".tpl") {
			return nil, fmt.Errorf("template must end in .tpl: %s", name)
		}
		contents, err := readResource(resources, name)
		if err != nil {
			return nil, err
		}
		bundle.Files[name] = contents
	}
	return bundle, nil
}

func (bundle *Bundle) compile(functions template.FuncMap) (*template.Template, error) {
	root := template.New("bundle").Option("missingkey=error").Funcs(functions)
	names := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	owners := map[string]string{}
	for _, name := range names {
		parsed, err := template.New(name).Option("missingkey=error").Funcs(functions).Parse(string(bundle.Files[name]))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		for _, part := range parsed.Templates() {
			if part.Tree == nil || parse.IsEmptyTree(part.Tree.Root) {
				continue
			}
			if previous, exists := owners[part.Name()]; exists {
				return nil, fmt.Errorf("duplicate template %q in %s and %s", part.Name(), previous, name)
			}
			owners[part.Name()] = name
		}
		for _, part := range parsed.Templates() {
			if part.Tree == nil {
				continue
			}
			if _, err := root.AddParseTree(part.Name(), part.Tree); err != nil {
				return nil, err
			}
		}
	}
	return root, nil
}

func extension(name string) (string, error) {
	extension := path.Ext(strings.TrimSuffix(name, ".tpl"))
	if extension == "" {
		return "", fmt.Errorf("template has no output extension: %s", name)
	}
	return extension, nil
}
