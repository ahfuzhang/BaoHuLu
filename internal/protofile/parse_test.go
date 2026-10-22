package protofile

import (
	"os"
	"strings"
	"testing"

	"github.com/ahfuzhang/BaoHuLu/internal/schema"
)

func TestParseAllTypes(t *testing.T) {
	input, err := os.Open("../integration/testdata/AllTypes.proto")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	document, err := Parse(input, "AllTypes.proto")
	if err != nil {
		t.Fatal(err)
	}
	if !document.File.GenTestCode.Value || !document.File.GenBenchmarkCode.Value || !document.File.GenProjectFile.Value {
		t.Fatal("file switches not preserved")
	}
	message := document.MessagesByName["demo.AllTypes"]
	if message == nil || len(message.Fields) < 30 {
		t.Fatal("AllTypes fields missing")
	}
	for _, field := range message.Fields {
		if field.Name == "next" && (!field.Recursive || field.Type.Message != message) {
			t.Fatal("recursive field does not reference canonical message")
		}
		if field.Name == "Status" && field.Type.Kind != schema.EnumKind {
			t.Fatal("enum classification lost")
		}
	}
	if !document.Analysis.YAML["demo.ValueTypes"] {
		t.Fatal("YAML did not propagate")
	}
}

func TestForwardReferencesAndServiceContext(t *testing.T) {
	document, err := Parse(strings.NewReader(`syntax = "proto3";
package demo;
service Echo {
  // @path=/echo
  rpc First(Request) returns (Response);
  rpc Second(Request) returns (Response);
}
message Request { State state = 1; Response next = 2; }
message Response { Request next = 1; }
enum State { NONE = 0; READY = 1; }
`), "input.proto")
	if err != nil {
		t.Fatal(err)
	}
	request := document.MessagesByName["demo.Request"]
	if request.Fields[0].Type.Kind != schema.EnumKind {
		t.Fatal("forward enum was classified as message")
	}
	if !request.Fields[1].Recursive {
		t.Fatal("mutual recursion not detected")
	}
	service := document.Services[0]
	if len(service.ReferencedTypes) != 2 {
		t.Fatalf("references: %v", service.ReferencedTypes)
	}
	if service.Methods[0].Request.Message != request || service.Methods[1].Request.Message != request {
		t.Fatal("requests were copied instead of sharing Schema message")
	}
	if service.Methods[0].Service != service || service.Methods[0].Paths[0] != "/echo" {
		t.Fatal("service relationship/path missing")
	}
}

func TestFileSwitchDefaultsAndValidation(t *testing.T) {
	for _, testCase := range []struct {
		name, header         string
		test, bench, project bool
	}{
		{"defaults", "", false, false, true},
		{"enabled", "// @gen_test_code=on\n// @gen_benchmark_code=on\n", true, true, true},
		{"disabled", "// @gen_project_file=off\n", false, false, false},
		{"case", "// @GEN_TEST_CODE=on\n", true, false, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			document, err := Parse(strings.NewReader(testCase.header+"syntax = \"proto3\";"), "switches.proto")
			if err != nil {
				t.Fatal(err)
			}
			if document.File.GenTestCode.Value != testCase.test || document.File.GenBenchmarkCode.Value != testCase.bench || document.File.GenProjectFile.Value != testCase.project {
				t.Fatalf("switches: %+v", document.File)
			}
		})
	}
	for _, source := range []string{
		"// @gen_test_code\nsyntax = \"proto3\";",
		"// @gen_test_code=true\nsyntax = \"proto3\";",
		"// @gen_test_code=ON\nsyntax = \"proto3\";",
		"// @gen_test_code=on\n// @gen_test_code=off\nsyntax = \"proto3\";",
		"syntax = \"proto3\";\n// @gen_test_code=on\nmessage A {}",
	} {
		if _, err := Parse(strings.NewReader(source), "invalid.proto"); err == nil {
			t.Errorf("accepted invalid extension: %s", source)
		}
	}
}

func TestInvalidSchemaRejected(t *testing.T) {
	for _, body := range []string{
		"message A { int32 first = 1; int64 second = 1; }",
		"message A { string first = 19000; }",
		"message A { string first = 536870912; }",
		"message A { map<double, int32> values = 1; }",
		"message A { optional int32 first = 1; }",
		"message A { oneof value { string first = 1; } }",
		"message A {} message A {}",
		"enum A { FIRST = 1; }",
		"service A { rpc Call(stream Req) returns (Resp); }",
		"message A {\n// @decimal=round:0\ndouble value = 1;\n}",
		"// @AsMap\nmessage A { int32 value = 1; }",
		"message Empty {} message Used { Empty item = 1; }",
		"message Empty {} service Used { rpc Call(Empty) returns (Empty); }",
		"message Value { int32 item=1; } service Echo {\n// @path=/Echo/Call\nrpc Call(Value) returns (Value); }",
	} {
		if _, err := Parse(strings.NewReader("syntax = \"proto3\";\n"+body), "invalid.proto"); err == nil {
			t.Errorf("accepted invalid schema: %s", body)
		}
	}
}

func TestNestedResolutionAndEmptyMessage(t *testing.T) {
	document, err := Parse(strings.NewReader(`syntax = "proto3";
package demo;
message Leaf { int32 global = 1; }
message Root {
  message Leaf { int32 local = 1; }
  Leaf child = 1;
  .demo.Leaf global = 2;
}
message Empty {}
`), "nested.proto")
	if err != nil {
		t.Fatal(err)
	}
	root := document.MessagesByName["demo.Root"]
	if root.Fields[0].Type.Message.FullName != "demo.Root.Leaf" || root.Fields[1].Type.Message.FullName != "demo.Leaf" {
		t.Fatal("incorrect lexical resolution")
	}
	if document.MessagesByName["demo.Empty"].Generate() {
		t.Fatal("empty message must not generate code")
	}
}
