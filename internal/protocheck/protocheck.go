package protocheck

import (
	"fmt"
	"os"
	"unicode"

	"github.com/ahfuzhang/BaoHuLu/internal/protoextensions"
	"github.com/emicklei/proto"
)

// isValidIdent reports whether s is a valid identifier (letter or _ start,
// followed by letters, digits, or underscores).
func isValidIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

// Check opens the given .proto file, parses it, and reports any syntax errors.
// It returns nil if the file is valid.
func Check(srcFile string) error {
	f, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcFile, err)
	}
	defer f.Close()

	parser := proto.NewParser(f)
	definition, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parse %s: %w", srcFile, err)
	}

	var checkErr error
	messageNames := make(map[string]int) // message name -> first occurrence line
	fieldNames := make(map[string]int)   // field name -> first occurrence line

	proto.Walk(definition,
		// 1. import 语句：打印警告但不报错，忽略后继续处理
		proto.WithImport(func(i *proto.Import) {
			fmt.Fprintf(os.Stderr, "warning: %s:%d: import is not supported and will be ignored\n",
				srcFile, i.Position.Line)
		}),
		// 2. message 字段中的 oneof 不支持
		proto.WithOneof(func(o *proto.Oneof) {
			if checkErr == nil {
				checkErr = fmt.Errorf("%s:%d: oneof is not supported", srcFile, o.Position.Line)
			}
		}),
		// 2. message 字段中 optional/required 修饰符不支持
		proto.WithNormalField(func(nf *proto.NormalField) {
			if checkErr != nil {
				return
			}
			if nf.Optional {
				checkErr = fmt.Errorf("%s:%d: optional field modifier is not supported", srcFile, nf.Position.Line)
				return
			} else if nf.Required {
				checkErr = fmt.Errorf("%s:%d: required field modifier is not supported", srcFile, nf.Position.Line)
				return
			}
			// field name must be a valid identifier
			if !isValidIdent(nf.Name) {
				checkErr = fmt.Errorf("%s:%d: field name %q is not a valid identifier", srcFile, nf.Position.Line, nf.Name)
				return
			}
			// @varName and @jsonName values must also be valid identifiers
			var commentLines []string
			if nf.Comment != nil {
				commentLines = nf.Comment.Lines
			}
			ext, _ := protoextensions.ParseAndStripField(commentLines)
			if ext.VarName != "" && !isValidIdent(ext.VarName) {
				checkErr = fmt.Errorf("%s:%d: @varName value %q is not a valid identifier", srcFile, nf.Position.Line, ext.VarName)
				return
			}
			if ext.JsonName != "" && !isValidIdent(ext.JsonName) {
				checkErr = fmt.Errorf("%s:%d: @jsonName value %q is not a valid identifier", srcFile, nf.Position.Line, ext.JsonName)
				return
			}
			// warn on duplicate field names across all messages
			if firstLine, exists := fieldNames[nf.Name]; exists {
				fmt.Fprintf(os.Stderr, "warning: %s:%d: field name %q already used at line %d in another message\n",
					srcFile, nf.Position.Line, nf.Name, firstLine)
			} else {
				fieldNames[nf.Name] = nf.Position.Line
			}
		}),
		// 3. extensions 语法不支持（WithExtensions 未提供，用自定义 Handler）
		func(v proto.Visitee) {
			e, ok := v.(*proto.Extensions)
			if !ok || checkErr != nil {
				return
			}
			checkErr = fmt.Errorf("%s:%d: extensions is not supported", srcFile, e.Position.Line)
		},
		// 3. extend 语法不支持（表现为 Message.IsExtend == true）
		// also check for duplicate message names
		proto.WithMessage(func(m *proto.Message) {
			if checkErr != nil {
				return
			}
			if m.IsExtend {
				checkErr = fmt.Errorf("%s:%d: extend is not supported", srcFile, m.Position.Line)
				return
			}
			if firstLine, exists := messageNames[m.Name]; exists {
				checkErr = fmt.Errorf("%s:%d: duplicate message name %q (first defined at line %d)",
					srcFile, m.Position.Line, m.Name, firstLine)
				return
			}
			messageNames[m.Name] = m.Position.Line
			// field name must not be identical to the message name
			for _, elem := range m.Elements {
				var fieldName string
				var fieldLine int
				switch f := elem.(type) {
				case *proto.NormalField:
					fieldName = f.Name
					fieldLine = f.Position.Line
				case *proto.MapField:
					fieldName = f.Name
					fieldLine = f.Position.Line
				}
				if fieldName == m.Name {
					checkErr = fmt.Errorf("%s:%d: field name %q must not be identical to its message name",
						srcFile, fieldLine, fieldName)
					return
				}
			}
			// @AsMap constraint: message must contain exactly one map-type field.
			var commentLines []string
			if m.Comment != nil {
				commentLines = m.Comment.Lines
			}
			msgExt, _ := protoextensions.ParseAndStripMessage(commentLines)
			if msgExt.AsMap {
				var fieldCount int
				var hasMapField bool
				for _, elem := range m.Elements {
					switch elem.(type) {
					case *proto.NormalField:
						fieldCount++
					case *proto.MapField:
						fieldCount++
						hasMapField = true
					}
				}
				if fieldCount != 1 {
					checkErr = fmt.Errorf("%s:%d: @AsMap message %q must have exactly one field, got %d",
						srcFile, m.Position.Line, m.Name, fieldCount)
					return
				}
				if !hasMapField {
					checkErr = fmt.Errorf("%s:%d: @AsMap message %q: the single field must be a map type",
						srcFile, m.Position.Line, m.Name)
					return
				}
			}
		}),
		// 4. service method 中 stream 修饰符不支持
		proto.WithRPC(func(r *proto.RPC) {
			if checkErr != nil {
				return
			}
			if r.StreamsRequest {
				checkErr = fmt.Errorf("%s:%d: stream request is not supported", srcFile, r.Position.Line)
			} else if r.StreamsReturns {
				checkErr = fmt.Errorf("%s:%d: stream response is not supported", srcFile, r.Position.Line)
			}
		}),
		// 5. message 和 service 定义中的 option 语句不支持
		proto.WithOption(func(o *proto.Option) {
			if checkErr != nil {
				return
			}
			switch o.Parent.(type) {
			case *proto.Message:
				checkErr = fmt.Errorf("%s:%d: option inside message is not supported", srcFile, o.Position.Line)
			case *proto.Service:
				checkErr = fmt.Errorf("%s:%d: option inside service is not supported", srcFile, o.Position.Line)
			}
		}),
	)
	return checkErr
}
