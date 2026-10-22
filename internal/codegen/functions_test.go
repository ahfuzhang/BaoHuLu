package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

func TestTemplateFunctionsReturnTypedValues(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader(`syntax="proto3"; message Alpha { int32 id=1; } message Beta { int32 id=1; }`), "input.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, "message: body.txt.tpl", map[string]string{
		"body.txt.tpl":         `{{with model .}}{{.Name}}:{{if enabled}}on{{else}}off{{end}}:{{range .Values}}{{.}}{{end}}{{end}}`,
		"partials/helpers.tpl": `{{define "fn/model"}}{{return (dict "Name" .Message.Name "Values" (list 1 2 3))}}{{end}}{{define "fn/enabled"}}{{return false}}{{end}}`,
	})
	outputs, err := bundle.Plan(document, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		rendered, err := bundle.Render(outputs)
		if err != nil {
			t.Fatal(err)
		}
		if string(rendered[0].Content) != "Alpha:off:123" || string(rendered[1].Content) != "Beta:off:123" {
			t.Fatalf("typed values/context leaked: %v", rendered)
		}
	}
}

func TestInvalidTemplateFunction(t *testing.T) {
	for _, body := range []string{
		`{{return 1}}text`, `{{return 1}}{{return 2}}`, `{{broken}}`,
	} {
		bundle := fixtureBundle(t, "message: body.txt.tpl", map[string]string{
			"body.txt.tpl": `{{broken}}`, "partials/helpers.tpl": `{{define "fn/broken"}}` + body + `{{end}}`,
		})
		if _, err := bundle.Render([]Output{{Template: "body.txt.tpl", Path: "body.txt"}}); err == nil {
			t.Fatalf("accepted invalid helper: %s", body)
		}
	}
}

func TestMemoUsesFileContextAndDoesNotLeakAcrossRenders(t *testing.T) {
	document, err := protofile.Parse(strings.NewReader(`syntax="proto3"; message Alpha { int32 id=1; } message Beta { int32 id=1; }`), "input.proto")
	if err != nil {
		t.Fatal(err)
	}
	bundle := fixtureBundle(t, "message: body.txt.tpl", map[string]string{
		"body.txt.tpl":         `{{.Message.Name}}:{{(memo "shared").Source}}:{{.Message.Name}}`,
		"partials/helpers.tpl": `{{define "fn/shared"}}{{if or .Message .Enum .Service .RPC .TypeRef .Args}}{{fail "selector leaked"}}{{end}}{{return (dict "Source" .File.Name)}}{{end}}`,
	})
	for _, name := range []string{"first.proto", "second.proto"} {
		document.File.Name = name
		planned, err := bundle.Plan(document, Options{})
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := bundle.Render(planned)
		if err != nil {
			t.Fatal(err)
		}
		if string(rendered[0].Content) != "Alpha:"+name+":Alpha" || string(rendered[1].Content) != "Beta:"+name+":Beta" {
			t.Fatalf("memo context leaked: %v", rendered)
		}
	}
}

func TestMemoAndFunctionNameErrors(t *testing.T) {
	for _, body := range []string{
		`{{memo "missing"}}`,
		`{{define "fn/loop"}}{{memo "loop"}}{{end}}{{memo "loop"}}`,
		`{{define "fn/get"}}conflict{{end}}`,
		`{{define "fn/len"}}conflict{{end}}`,
		`{{define "fn/render"}}conflict{{end}}`,
	} {
		bundle := fixtureBundle(t, "message: body.txt.tpl", map[string]string{"body.txt.tpl": body})
		if _, err := bundle.Render([]Output{{Template: "body.txt.tpl", Path: "body.txt"}}); err == nil {
			t.Fatalf("accepted invalid function: %s", body)
		}
	}
}

func TestMemoryLayout(t *testing.T) {
	fields := []any{
		map[string]any{"Name": "small", "Align": 1, "Size": 1, "Ptrdata": 0, "Number": 3},
		map[string]any{"Name": "slice", "Align": 8, "Size": 24, "Ptrdata": 8, "Number": 2},
		map[string]any{"Name": "pointer", "Align": 8, "Size": 8, "Ptrdata": 8, "Number": 1},
	}
	result, err := memoryLayout(fields)
	if err != nil {
		t.Fatal(err)
	}
	ordered := result["Fields"].([]any)
	if !reflect.DeepEqual(ordered, []any{fields[2], fields[1], fields[0]}) || result["Size"] != 40 || result["Ptrdata"] != 16 {
		t.Fatalf("layout: %v", result)
	}
	for _, invalid := range []any{1, map[string]any{"Size": -1, "Align": 8, "Ptrdata": 0, "Number": 1}, map[string]any{"Size": 1, "Align": 3, "Ptrdata": 0, "Number": 1}} {
		if _, err := memoryLayout([]any{invalid}); err == nil {
			t.Fatalf("accepted %v", invalid)
		}
	}
}
