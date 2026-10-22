package codegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

func fixtureBundle(t *testing.T, config string, files map[string]string) *Bundle {
	t.Helper()
	resources := fstest.MapFS{"tpl_config.yaml": {Data: []byte(config)}}
	for name, contents := range files {
		resources[name] = &fstest.MapFile{Data: []byte(contents)}
	}
	bundle, err := LoadBundle(resources, "test")
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestPlanTypeSelectionAndSwitches(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader(`syntax = "proto3"; package demo;
enum State { NONE = 0; }
message Empty {}
message Thing { State state = 1; }
`), "input.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, `message: type.txt.tpl
enum: enum.txt.tpl
json:
  FromJSON: from.txt.tpl
tests:
  "{message}.test.txt": test.txt.tpl
project:
  "{file}.project": project.txt.tpl
`, map[string]string{
		"type.txt.tpl": "{{.Message.Name}}", "enum.txt.tpl": "{{.Enum.Name}}",
		"from.txt.tpl": "{{.Message.Name}}", "test.txt.tpl": "{{.Message.Name}}", "project.txt.tpl": "{{.File.Package}}",
	})
	outputs, err := bundle.Plan(document, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, output := range outputs {
		paths = append(paths, output.Path)
	}
	if !reflect.DeepEqual(paths, []string{"State.txt", "Thing.FromJSON.txt", "Thing.txt", "input.project"}) {
		t.Fatalf("paths: %v", paths)
	}
	rendered, err := bundle.Render(outputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(rendered) != 4 || string(rendered[0].Content) != "State" {
		t.Fatalf("outputs: %+v", rendered)
	}
	document.File.GenTestCode.Value = true
	document.File.GenProjectFile.Value = false
	outputs, err = bundle.Plan(document, Options{})
	if err != nil {
		t.Fatal(err)
	}
	paths = nil
	for _, output := range outputs {
		paths = append(paths, output.Path)
	}
	if !reflect.DeepEqual(paths, []string{"State.txt", "Thing.FromJSON.txt", "Thing.test.txt", "Thing.txt"}) {
		t.Fatalf("paths: %v", paths)
	}
}

func TestServiceTemplatesUseUnifiedContext(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader(`syntax = "proto3";
message Request { int32 id = 1; }
service Echo {
  rpc First(Request) returns (Request);
  rpc Second(Request) returns (External);
}`), "service.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, "service: server.txt.tpl\nrpc: method.txt.tpl\n", map[string]string{
		"server.txt.tpl":         `{{range .Service.Methods}}{{render "signature" (withRPC $ .)}};{{end}}`,
		"method.txt.tpl":         `{{render "signature" .}}`,
		"partials/signature.tpl": `{{define "signature"}}{{.Service.Name}}.{{.RPC.Name}}:{{render "typename" (withTypeRef . .RPC.Request)}}{{end}}{{define "typename"}}{{.TypeRef.RawName}}{{end}}`,
	})
	outputs, err := bundle.Plan(document, Options{})
	if err != nil {
		t.Fatal(err)
	}
	outputs, err = bundle.Render(outputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(outputs) != 3 || string(outputs[2].Content) != "Echo.First:Request;Echo.Second:Request;" {
		t.Fatalf("outputs: %+v", outputs)
	}
}

func TestStrictBundleValidation(t *testing.T) {
	for _, config := range []string{
		"unknown: body.txt.tpl\n", "message: ''\n", "message: null\n",
		"message: body.txt.tpl\nmessage: other.txt.tpl\n",
		"message: body.txt.tpl\n---\nmessage: body.txt.tpl\n",
		"message: ../body.txt.tpl\n", "message: /body.txt.tpl\n",
		"json:\n  Guess: body.txt.tpl\n", "json:\n  FromJSON: ''\n",
		"tests:\n  ../outside: body.txt.tpl\n", "tests:\n  '{unknown}': body.txt.tpl\n",
	} {
		resources := fstest.MapFS{"tpl_config.yaml": {Data: []byte(config)}, "body.txt.tpl": {Data: []byte("body")}}
		if _, err := LoadBundle(resources, "test"); err == nil {
			t.Errorf("accepted invalid config: %s", config)
		}
	}
}

func TestTemplateFailures(t *testing.T) {
	for _, files := range []map[string]string{
		{"body.txt.tpl": `{{.Missing}}`},
		{"body.txt.tpl": `{{render "loop" .}}`, "partials/loop.tpl": `{{define "loop"}}{{render "loop" .}}{{end}}`},
		{"body.txt.tpl": `{{define "dup"}}first{{end}}`, "partials/duplicate.tpl": `{{define "dup"}}second{{end}}`},
	} {
		bundle := fixtureBundle(t, "message: body.txt.tpl\n", files)
		if _, err := bundle.Render([]Output{{Template: "body.txt.tpl", Path: "body.txt"}}); err == nil {
			t.Fatalf("accepted broken template: %v", files)
		}
	}
}

func TestCapabilityAndOutputConflicts(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader("// @gen_test_code=on\nsyntax = \"proto3\"; message A { int32 id = 1; }"), "input.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, "message: body.txt.tpl\n", map[string]string{"body.txt.tpl": "body"})
	if _, err := bundle.Plan(document, Options{}); err == nil {
		t.Fatal("unsupported explicit tests accepted")
	}
	document.File.GenTestCode.Value = false
	bundle.Config.Project = map[string]string{"A.txt": "body.txt.tpl"}
	if _, err := bundle.Plan(document, Options{}); err == nil {
		t.Fatal("duplicate output accepted")
	}
}

func TestWriteProtectionAndPrune(t *testing.T) {
	root := t.TempDir()
	outputs := []Output{
		{Path: "A.txt", Template: "body.txt.tpl", Content: []byte("old")},
		{Path: "project.txt", CreateOnly: true, Content: []byte("project")},
	}
	if _, err := Write(root, "unit", outputs); err != nil {
		t.Fatal(err)
	}
	manifestBefore, err := os.ReadFile(filepath.Join(root, manifestName))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "A.txt"), []byte("manual"), 0644); err != nil {
		t.Fatal(err)
	}
	outputs[0].Content = []byte("new")
	if _, err := Write(root, "unit", outputs); err == nil {
		t.Fatal("overwrote modified file")
	}
	manifestAfter, err := os.ReadFile(filepath.Join(root, manifestName))
	if err != nil {
		t.Fatal(err)
	}
	if string(manifestBefore) != string(manifestAfter) {
		t.Fatal("failed write changed manifest")
	}
	if _, err := Write(root, "unit", nil); err == nil {
		t.Fatal("pruned modified file")
	}
	if err := os.WriteFile(filepath.Join(root, "A.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := Write(root, "unit", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "A.txt" {
		t.Fatalf("prune: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "project.txt")); err != nil {
		t.Fatal("project was pruned")
	}
}

func TestWriteRejectsUnownedAndSymlinkPaths(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "manual.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, "unit", []Output{{Path: "manual.txt", Content: []byte("replace")}}); err == nil {
		t.Fatal("overwrote unowned file")
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, "unit", []Output{{Path: "escape/file.txt", Content: []byte("escape")}}); err == nil {
		t.Fatal("followed symlink")
	}
	if _, err := Write(root, "unit", []Output{{Path: "../escape.txt"}}); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := Write(root, "first", []Output{{Path: "owned.txt", Content: []byte("first")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, "second", []Output{{Path: "owned.txt", Content: []byte("second")}}); err == nil {
		t.Fatal("overwrote another unit")
	}
}

func TestRecoveryAndConcurrentModification(t *testing.T) {
	root := t.TempDir()
	transaction := journal{Version: 1, Files: []snapshot{{Path: "A.txt", Existed: true, Content: []byte("old"), Mode: 0644, ExpectedExists: true, ExpectedHash: digest([]byte("new"))}}}
	contents, err := json.Marshal(transaction)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, transactionName), contents, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "A.txt"), []byte("manual"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Recover(root); err == nil {
		t.Fatal("recovery overwrote concurrent edit")
	}
	if err := os.WriteFile(filepath.Join(root, "A.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Recover(root); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(root, "A.txt"))
	if err != nil || string(restored) != "old" {
		t.Fatalf("restore: %s, %v", restored, err)
	}
}

func TestAllTypesOutputPlan(t *testing.T) {
	input, err := os.Open("../integration/testdata/AllTypes.proto")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	document, err := protofile.Parse(input, "AllTypes.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, `message: body.txt.tpl
enum: enum.txt.tpl
tests:
  "{message}.test.txt": body.txt.tpl
benchmarks:
  "{message}.bench.txt": body.txt.tpl
project:
  metadata.txt: file.txt.tpl
`, map[string]string{"body.txt.tpl": `{{.Message.FullName}}:{{len .Message.Fields}}`, "enum.txt.tpl": `{{.Enum.FullName}}`, "file.txt.tpl": `{{.File.Package}}`})
	outputs, err := bundle.Plan(document, Options{})
	if err != nil {
		t.Fatal(err)
	}
	outputs, err = bundle.Render(outputs)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if _, err := Write(root, "all-types", outputs); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "AllTypes.txt"))
	if err != nil || !strings.HasPrefix(string(contents), "demo.AllTypes:") {
		t.Fatalf("AllTypes output: %s, %v", contents, err)
	}
	result, err := Write(root, "all-types", outputs)
	if err != nil || len(result.Written) != 0 {
		t.Fatalf("second run: %+v, %v", result, err)
	}
}

func TestWriteRecoversInterruptedRunBeforeGenerating(t *testing.T) {
	root := t.TempDir()
	transaction := journal{Version: 1, Files: []snapshot{{Path: "A.txt", Existed: true, Content: []byte("old"), Mode: 0644, ExpectedExists: true, ExpectedHash: digest([]byte("interrupted"))}}}
	contents, err := json.Marshal(transaction)
	if err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string][]byte{transactionName: contents, "A.txt": []byte("interrupted")} {
		if err := os.WriteFile(filepath.Join(root, name), contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, err = Write(root, "unit", []Output{{Path: "B.txt", Content: []byte("new")}})
	if err == nil || !strings.Contains(err.Error(), "recovered interrupted generation") {
		t.Fatalf("missing recovery diagnostic: %v", err)
	}
	restored, err := os.ReadFile(filepath.Join(root, "A.txt"))
	if err != nil || string(restored) != "old" {
		t.Fatalf("restore: %s, %v", restored, err)
	}
	for _, name := range []string{"B.txt", manifestName, transactionName, lockName} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected output %s: %v", name, err)
		}
	}
}

func TestWritePreservesMetadataAndIsDeterministic(t *testing.T) {
	root := t.TempDir()
	metadata := UnitMetadata{GeneratorVersion: "test", SourceHash: "source", TemplateHash: "template"}
	outputs := []Output{{Path: "A.txt", Content: []byte("stable")}}
	if _, err := Write(root, "unit", outputs, metadata); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, manifestName))
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(before, &manifest); err != nil || manifest.Units["unit"].Metadata != metadata {
		t.Fatalf("metadata lost: %+v, %v", manifest, err)
	}
	result, err := Write(root, "unit", outputs, metadata)
	if err != nil || len(result.Written) != 0 || len(result.Removed) != 0 {
		t.Fatalf("unstable write: %+v, %v", result, err)
	}
	after, err := os.ReadFile(filepath.Join(root, manifestName))
	if err != nil || string(before) != string(after) {
		t.Fatalf("unstable manifest: %v", err)
	}
}
