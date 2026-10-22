package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpVersionAndRemovedArguments(t *testing.T) {
	for _, argument := range []string{"help", "-h", "-help", "--help", "--version"} {
		var output bytes.Buffer
		if err := run([]string{argument}, &output, &output); err != nil || output.Len() == 0 {
			t.Fatalf("%s: %v", argument, err)
		}
	}
	for _, arguments := range [][]string{
		nil, {"generate"}, {"tu"}, {"check"}, {"xi"}, {"-go_out=x"}, {"-csharp_out=x"},
		{"-with.test"}, {"-with.bench"}, {"-with.project"}, {"-no.csproj.file"},
		{"-dry-run"}, {"-types.manifest=x"}, {"-src=a", "-src=b"}, {"-src=a"}, {"unexpected"},
	} {
		var output bytes.Buffer
		if err := run(arguments, &output, &output); err == nil {
			t.Fatalf("accepted %v", arguments)
		}
	}
}

func TestExternalBundleAndCheckBeforeWrite(t *testing.T) {
	root := t.TempDir()
	source, bundle, output := filepath.Join(root, "input.proto"), filepath.Join(root, "bundle"), filepath.Join(root, "output")
	if err := os.Mkdir(bundle, 0755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"tpl_config.yaml": "message: message.txt.tpl\n", "message.txt.tpl": "{{.Message.Name}}:{{len .Message.Fields}}"} {
		if err := os.WriteFile(filepath.Join(bundle, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	arguments := []string{"-src=" + source, "-template.dir=" + bundle, "-out=" + output}
	var messages bytes.Buffer
	if err := os.WriteFile(source, []byte(`syntax = "proto3"; message Item { string name = 1; }`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := run(arguments, &messages, &messages); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(output, "Item.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != "Item:1" {
		t.Fatal(string(before))
	}
	if err := os.WriteFile(source, []byte(`syntax = "proto3"; message Item { optional string name = 1; }`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := run(arguments, &messages, &messages); err == nil || !strings.Contains(err.Error(), "optional") {
		t.Fatalf("expected check error: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(output, "Item.txt"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("output changed: %v", err)
	}
}
