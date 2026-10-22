package integration

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoBaselineDeclarations(t *testing.T) {
	baseline, migrated := os.Getenv("BAOHULU_BASELINE_GO"), os.Getenv("BAOHULU_MIGRATED_GO")
	if baseline == "" || migrated == "" {
		t.Skip("set BAOHULU_BASELINE_GO and BAOHULU_MIGRATED_GO for baseline comparison")
	}
	before, after := goDeclarations(t, baseline), goDeclarations(t, migrated)
	for name, declaration := range before {
		if actual, exists := after[name]; !exists {
			t.Errorf("missing declaration %s", name)
		} else if actual != declaration {
			t.Errorf("changed declaration %s", name)
		}
	}
	for name := range after {
		if _, exists := before[name]; !exists {
			t.Errorf("added declaration %s", name)
		}
	}
}

func goDeclarations(t *testing.T, root string) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]string{}
	for _, filename := range files {
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		positions := token.NewFileSet()
		file, err := parser.ParseFile(positions, filename, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			switch node := declaration.(type) {
			case *ast.FuncDecl:
				name := node.Name.Name
				if node.Recv != nil {
					var receiver bytes.Buffer
					if err := format.Node(&receiver, positions, node.Recv.List[0].Type); err != nil {
						t.Fatal(err)
					}
					name = receiver.String() + "." + name
				}
				var content bytes.Buffer
				if err := format.Node(&content, positions, node); err != nil {
					t.Fatal(err)
				}
				result[name] = content.String()
			case *ast.GenDecl:
				if node.Tok != token.TYPE {
					continue
				}
				for _, specification := range node.Specs {
					item := specification.(*ast.TypeSpec)
					var content bytes.Buffer
					if err := format.Node(&content, positions, item); err != nil {
						t.Fatal(err)
					}
					result[item.Name.Name] = content.String()
				}
			}
		}
	}
	return result
}
