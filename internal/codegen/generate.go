package codegen

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

type Result struct {
	*WriteResult
	Warnings []string
}

func Generate(options Options, resources fs.FS) (*Result, error) {
	if options.Version == "" {
		options.Version = "dev"
	}
	if options.Source == "" || options.TemplateDir == "" || options.OutputDir == "" {
		return nil, fmt.Errorf("-src, -template.dir and -out are required")
	}
	input, err := os.Open(options.Source)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	document, parseErr := protofile.Parse(input, options.Source)
	closeErr := input.Close()
	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if resources == nil {
		return nil, fmt.Errorf("template resources are required")
	}
	identity, err := generationIdentity(options)
	if err != nil {
		return nil, err
	}
	bundle, err := LoadBundle(resources, options.TemplateDir)
	if err != nil {
		return nil, err
	}
	outputs, err := bundle.Plan(document, options)
	if err != nil {
		return nil, err
	}
	outputs, err = bundle.Render(outputs)
	if err != nil {
		return nil, err
	}
	templateData, err := json.Marshal(struct {
		Config Config
		Files  map[string][]byte
	}{bundle.Config, bundle.Files})
	if err != nil {
		return nil, err
	}
	metadata := UnitMetadata{GeneratorVersion: options.Version, SourceHash: digest([]byte(document.File.Source)), TemplateHash: digest(templateData)}
	written, err := Write(options.OutputDir, identity, outputs, metadata)
	if err != nil {
		return nil, err
	}
	result := &Result{WriteResult: written, Warnings: append([]string(nil), document.Warnings...)}
	for _, name := range written.Preserved {
		result.Warnings = append(result.Warnings, fmt.Sprintf("preserved create-only file %s; update project dependencies manually when enabling features", name))
	}
	return result, nil
}

func generationIdentity(options Options) (string, error) {
	root, err := filepath.Abs(options.OutputDir)
	if err != nil {
		return "", err
	}
	source, err := filepath.Abs(options.Source)
	if err != nil {
		return "", err
	}
	source, err = filepath.Rel(root, source)
	if err != nil {
		return "", err
	}
	bundle := options.TemplateDir
	if !strings.HasPrefix(bundle, "builtin:") {
		absolute, err := filepath.Abs(bundle)
		if err != nil {
			return "", err
		}
		bundle, err = filepath.Rel(root, absolute)
		if err != nil {
			return "", err
		}
		bundle = filepath.ToSlash(bundle)
	}
	identity, err := json.Marshal([]string{filepath.ToSlash(source), bundle})
	if err != nil {
		return "", err
	}
	return digest(identity), nil
}
