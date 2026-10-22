package integration

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ahfuzhang/BaoHuLu/internal/codegen"
	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
	"github.com/ahfuzhang/BaoHuLu/templates"
)

func TestAllTypesGo(t *testing.T) {
	output := t.TempDir()
	if configured := os.Getenv("BAOHULU_GENERATED_TEST_DIR"); configured != "" {
		output = configured
	}
	bundle, err := templates.Open("builtin:golang")
	if err != nil {
		t.Fatal(err)
	}
	_, err = codegen.Generate(codegen.Options{Source: "testdata/AllTypes.proto", TemplateDir: "builtin:golang", OutputDir: output, Version: "dev"}, bundle)
	if err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") {
			if _, err := parser.ParseFile(token.NewFileSet(), filepath.Join(output, file.Name()), nil, parser.AllErrors); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCapabilityMatrix(t *testing.T) {
	for _, language := range []string{"golang", "csharp"} {
		for _, withTest := range []bool{false, true} {
			for _, withBench := range []bool{false, true} {
				for _, withProject := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/tests=%t/bench=%t/project=%t", language, withTest, withBench, withProject), func(t *testing.T) {
						root := t.TempDir()
						switchValue := func(value bool) string {
							if value {
								return "on"
							}
							return "off"
						}
						source := fmt.Sprintf("// @gen_test_code=%s\n// @gen_benchmark_code=%s\n// @gen_project_file=%s\nsyntax=\"proto3\"; package matrix; message Item { string text=1; int32 count=2; }", switchValue(withTest), switchValue(withBench), switchValue(withProject))
						filename, output := filepath.Join(root, "Matrix.proto"), filepath.Join(root, "output")
						if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
							t.Fatal(err)
						}
						bundle, err := templates.Open("builtin:" + language)
						if err != nil {
							t.Fatal(err)
						}
						result, err := codegen.Generate(codegen.Options{Source: filename, TemplateDir: "builtin:" + language, OutputDir: output}, bundle)
						if err != nil {
							t.Fatal(err)
						}
						projects := 0
						for _, name := range result.Written {
							if strings.HasSuffix(name, ".csproj") || name == "go.mod" {
								projects++
							}
						}
						if (projects > 0) != withProject {
							t.Fatalf("project flag ignored: %v", result.Written)
						}
						if os.Getenv("BAOHULU_BUILD_GENERATED") == "1" && withProject {
							compileGenerated(t, language, output, withTest, withBench)
						}
					})
				}
			}
		}
	}
}

func compileGenerated(t *testing.T, language, directory string, withTest, withBench bool) {
	t.Helper()
	run := func(command string, arguments ...string) {
		process := exec.Command(command, arguments...)
		process.Dir = directory
		process.Env = append(os.Environ(), "GOWORK=off")
		if output, err := process.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", command, arguments, err, output)
		}
	}
	if language == "golang" {
		root, err := filepath.Abs("../..")
		if err != nil {
			t.Fatal(err)
		}
		run("go", "mod", "edit", "-replace=github.com/ahfuzhang/BaoHuLu="+root)
		run("go", "mod", "tidy")
		run("go", "test", "-run", "^$", "./...")
		return
	}
	run("dotnet", "build", "Matrix.csproj", "--nologo")
	if withTest {
		run("dotnet", "build", "Tests/Matrix.Tests.csproj", "--nologo")
	}
	if withBench {
		run("dotnet", "build", "Benchmarks/Matrix.Benchmarks.csproj", "--nologo")
	}
}

func TestBuiltinExternalEquivalence(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader(`syntax="proto3"; package same; message Item { int32 id=1; }`), "Same.proto")
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"golang", "csharp"} {
		var baseline []codegen.Output
		for _, name := range []string{"builtin:" + language, "../../templates/" + language} {
			resources, err := templates.Open(name)
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := codegen.LoadBundle(resources, name)
			if err != nil {
				t.Fatal(err)
			}
			planned, err := bundle.Plan(document, codegen.Options{Version: "dev"})
			if err != nil {
				t.Fatal(err)
			}
			rendered, err := bundle.Render(planned)
			if err != nil {
				t.Fatal(err)
			}
			if baseline == nil {
				baseline = rendered
				continue
			}
			if len(baseline) != len(rendered) {
				t.Fatal("different output count")
			}
			for index := range baseline {
				if baseline[index].Path != rendered[index].Path || string(baseline[index].Content) != string(rendered[index].Content) {
					t.Fatalf("external content differs: %s", rendered[index].Path)
				}
			}
		}
	}
}

func TestAllTypesCSharp(t *testing.T) {
	output := t.TempDir()
	if configured := os.Getenv("BAOHULU_GENERATED_TEST_DIR"); configured != "" {
		output = configured
	}
	bundle, err := templates.Open("builtin:csharp")
	if err != nil {
		t.Fatal(err)
	}
	_, err = codegen.Generate(codegen.Options{Source: "testdata/AllTypes.proto", TemplateDir: "builtin:csharp", OutputDir: output, Version: "dev"}, bundle)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoUnsupportedRepeatedEnumDoesNotWrite(t *testing.T) {
	output := filepath.Join(t.TempDir(), "not-created")
	bundle, err := templates.Open("builtin:golang")
	if err != nil {
		t.Fatal(err)
	}
	_, err = codegen.Generate(codegen.Options{Source: "testdata/RepeatedEnum.proto", TemplateDir: "builtin:golang", OutputDir: output, Version: "dev"}, bundle)
	if err == nil || !strings.Contains(err.Error(), "repeated enum") {
		t.Fatalf("expected limitation diagnostic: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output was created: %v", err)
	}
}

func TestTargetValidationBeforeWrite(t *testing.T) {
	for _, test := range []struct {
		name, language, filename, source, diagnostic string
	}{
		{"go-package", "golang", "Valid.proto", `syntax="proto3"; option go_package="example.org/foo;type"; message Item { int32 id=1; }`, "keyword"},
		{"go-enum-filename", "golang", "Valid.proto", `syntax="proto3"; enum State_linux { NONE=0; }`, "platform-specific"},
		{"go-tag-symbol", "golang", "Valid.proto", `syntax="proto3"; message Item { int32 id=1; } message ItemIdTag { int32 id=1; }`, "duplicate generated symbol"},
		{"go-alias", "golang", "Valid.proto", "syntax=\"proto3\"; message Item {\n// @jsonName=bad\"name\nint32 id=1; }", "alias needs escaping"},
		{"cs-alias", "csharp", "Valid.proto", "syntax=\"proto3\"; message Item {\n// @yamlName=bad\\name\nint32 id=1; }", "alias needs escaping"},
		{"cs-namespace", "csharp", "Valid.proto", `syntax="proto3"; option csharp_namespace="Example.class"; message Item { int32 id=1; }`, "keyword"},
		{"cs-enum", "csharp", "Valid.proto", `syntax="proto3"; enum Error { NONE=0; }`, "duplicate generated C# symbol"},
		{"cs-enum-value", "csharp", "Valid.proto", `syntax="proto3"; enum Status { class=0; }`, "keyword"},
		{"cs-basename", "csharp", "bad-name.proto", `syntax="proto3"; package valid; message Item { int32 id=1; }`, "basename"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			filename, output := filepath.Join(root, test.filename), filepath.Join(root, "output")
			if err := os.WriteFile(filename, []byte(test.source), 0644); err != nil {
				t.Fatal(err)
			}
			resources, err := templates.Open("builtin:" + test.language)
			if err != nil {
				t.Fatal(err)
			}
			_, err = codegen.Generate(codegen.Options{Source: filename, TemplateDir: "builtin:" + test.language, OutputDir: output}, resources)
			if err == nil || !strings.Contains(err.Error(), test.diagnostic) {
				t.Fatalf("want %q, got %v", test.diagnostic, err)
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("invalid input created output: %v", err)
			}
		})
	}
}

func TestBenchmarkProtoOnlyChangesNamespaceOption(t *testing.T) {
	source := `// @gen_benchmark_code=on
syntax="proto3";
option java_package="Example.Data";
// option csharp_namespace="Example.Data";
option csharp_namespace =
    'Example.Data';
message Item { int32 id=1; }
`
	document, err := protofile.Parse(strings.NewReader(source), "Example.proto")
	if err != nil {
		t.Fatal(err)
	}
	resources, err := templates.Open("builtin:csharp")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := codegen.LoadBundle(resources, "builtin:csharp")
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := bundle.Plan(document, codegen.Options{})
	if err != nil {
		t.Fatal(err)
	}
	outputs, err = bundle.Render(outputs)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range outputs {
		if strings.HasSuffix(output.Path, ".proto") {
			expected := strings.Replace(source, "'Example.Data'", `"Example.Data.GrpcProto"`, 1)
			if strings.TrimSpace(string(output.Content)) != strings.TrimSpace(expected) {
				t.Fatalf("unrelated source changed: %s", output.Content)
			}
			return
		}
	}
	t.Fatal("benchmark proto missing")
}
