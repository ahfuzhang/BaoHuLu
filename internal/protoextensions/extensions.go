// Package protoextensions parses the custom extension annotations that the
// BaoHuLu code generator reads from proto comment lines.
//
// Extension annotations have the form:
//
//	// @KeyWord=Value
//
// They are written inside ordinary proto comments so that standard protoc
// tooling continues to work unchanged.
package protoextensions

import (
	"strconv"
	"strings"
)

// TagExt holds a single @tag extension: @tag=Name:Value.
type TagExt struct {
	Name  string
	Value string
}

// FieldExtensions collects all extension annotations found on a proto field.
type FieldExtensions struct {
	// When true (@Deprecated annotation present), the field is omitted from generated code.
	Deprecated bool
	// VarName overrides the generated Go / C# struct field name.
	VarName string
	// JsonName overrides the json tag key (and the JSON name constant value).
	JsonName string
	// FormName overrides the form key constant used in FromPostForm. Falls back to JsonName,
	// then to the original proto field name when absent.
	FormName string
	// YamlName, when non-empty, adds a yaml struct tag and a YAML name constant.
	YamlName string
	// Tags holds arbitrary extra struct tags produced by @tag=Name:Value.
	Tags []TagExt
	// DecimalRound, when > 0, marks this field as a decimal type with the given number of
	// decimal places. Only valid on proto double fields. Syntax: @decimal=round:N
	DecimalRound int
}

// MessageExtensions collects all extension annotations found on a proto message.
type MessageExtensions struct {
	// When true (@Deprecated annotation present), the message is omitted from generated code.
	Deprecated bool
	// FromPostForm, when true (@from-post-form annotation present), causes a
	// FromPostForm(ReadOnlySpan<byte>) method to be generated on the C# readonly struct.
	// This annotation is infectious: any message type embedded in an annotated message
	// also gets the method generated.
	FromPostForm bool
	// AsMap, when true (@AsMap annotation present), constrains the message to contain
	// exactly one field of map type. Validated by protocheck and protofile.
	AsMap bool
}

// ParseAndStripField scans comment lines for extension annotations, removes
// those lines from the output, and returns the parsed extensions together with
// the remaining "clean" lines that should appear in generated code comments.
func ParseAndStripField(lines []string) (FieldExtensions, []string) {
	var ext FieldExtensions
	var clean []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			clean = append(clean, line)
			continue
		}
		key, value := splitKeyValue(trimmed[1:]) // strip leading '@'
		switch key {
		case "deprecated":
			ext.Deprecated = true
		case "varname":
			ext.VarName = value
		case "jsonname":
			ext.JsonName = value
		case "formname":
			ext.FormName = value
		case "yamlname":
			ext.YamlName = value
		case "tag":
			// value format: "TagName:tagValue"
			tagName, tagVal, ok := strings.Cut(value, ":")
			if ok {
				ext.Tags = append(ext.Tags, TagExt{
					Name:  strings.TrimSpace(tagName),
					Value: strings.TrimSpace(tagVal),
				})
			}
		case "decimal":
			// value format: "round:N"
			_, roundStr, ok := strings.Cut(value, "round:")
			if ok {
				if n, err := strconv.Atoi(strings.TrimSpace(roundStr)); err == nil && n > 0 {
					ext.DecimalRound = n
				}
			}
		}
	}
	return ext, clean
}

// RpcExtensions collects all extension annotations found on a proto RPC method.
type RpcExtensions struct {
	// Paths holds zero or more additional HTTP path aliases (@path=xxx); multiple
	// @path annotations on the same RPC are all collected.
	Paths []string
}

// ParseAndStripRpc scans comment lines for RPC-level extension annotations,
// removes those lines, and returns the parsed extensions together with the
// remaining clean lines.
func ParseAndStripRpc(lines []string) (RpcExtensions, []string) {
	var ext RpcExtensions
	var clean []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			clean = append(clean, line)
			continue
		}
		key, value := splitKeyValue(trimmed[1:])
		if key == "path" {
			ext.Paths = append(ext.Paths, value)
		}
		// Unknown RPC-level annotations are silently dropped.
	}
	return ext, clean
}

// ParseAndStripMessage scans comment lines for message-level extension
// annotations, removes those lines, and returns the parsed extensions together
// with the remaining clean lines.
func ParseAndStripMessage(lines []string) (MessageExtensions, []string) {
	var ext MessageExtensions
	var clean []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			clean = append(clean, line)
			continue
		}
		key, _ := splitKeyValue(trimmed[1:])
		switch key {
		case "deprecated":
			ext.Deprecated = true
		case "from-post-form":
			ext.FromPostForm = true
		case "asmap":
			ext.AsMap = true
		}
		// Unknown message-level annotations are silently dropped (not added to clean).
	}
	return ext, clean
}

// splitKeyValue splits a raw "Key=Value" string (the part after '@') into a
// lowercase key and a trimmed value. If there is no '=', value is empty.
func splitKeyValue(raw string) (key, value string) {
	k, v, _ := strings.Cut(raw, "=")
	return strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v)
}
