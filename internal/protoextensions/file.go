package protoextensions

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/ahfuzhang/BaoHuLu/internal/schema"
)

func ParseFile(source, filename string, syntaxLine int) (*schema.File, error) {
	file := &schema.File{Name: filename, Source: source, Options: map[string]string{}, GenProjectFile: schema.Switch{Value: true}}
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(source))
	lexer.Filename = filename
	lexer.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments
	lexer.Error = func(*scanner.Scanner, string) {}
	for token := lexer.Scan(); token != scanner.EOF; token = lexer.Scan() {
		if token != scanner.Comment {
			continue
		}
		comment := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(lexer.TokenText(), "//"), "/*"), "*/")
		for offset, line := range strings.Split(comment, "\n") {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
			if !strings.HasPrefix(line, "@") {
				continue
			}
			name, value, hasValue := strings.Cut(strings.TrimPrefix(line, "@"), "=")
			name, value = strings.ToLower(strings.TrimSpace(name)), strings.TrimSpace(value)
			location := schema.Location{File: filename, Line: lexer.Position.Line + offset, Column: lexer.Position.Column}
			var target *schema.Switch
			switch name {
			case "gen_test_code":
				target = &file.GenTestCode
			case "gen_benchmark_code":
				target = &file.GenBenchmarkCode
			case "gen_project_file":
				target = &file.GenProjectFile
			}
			if target != nil {
				if location.Line >= syntaxLine {
					return nil, fmt.Errorf("%s: @%s must appear before syntax", location, name)
				}
				if target.Explicit {
					return nil, fmt.Errorf("%s: duplicate @%s", location, name)
				}
				if !hasValue || value != "on" && value != "off" {
					return nil, fmt.Errorf("%s: @%s requires on or off", location, name)
				}
				target.Value, target.Explicit = value == "on", true
			}
			if location.Line < syntaxLine {
				file.Annotations = append(file.Annotations, schema.Annotation{Name: name, Value: value, Location: location})
			}
		}
	}
	return file, nil
}
