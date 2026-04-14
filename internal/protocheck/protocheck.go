package protocheck

import (
	"fmt"
	"os"

	"github.com/emicklei/proto"
)

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
	proto.Walk(definition,
		// 1. import 语句不支持
		proto.WithImport(func(i *proto.Import) {
			if checkErr == nil {
				checkErr = fmt.Errorf("%s:%d: import is not supported", srcFile, i.Position.Line)
			}
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
			} else if nf.Required {
				checkErr = fmt.Errorf("%s:%d: required field modifier is not supported", srcFile, nf.Position.Line)
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
		proto.WithMessage(func(m *proto.Message) {
			if checkErr == nil && m.IsExtend {
				checkErr = fmt.Errorf("%s:%d: extend is not supported", srcFile, m.Position.Line)
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
