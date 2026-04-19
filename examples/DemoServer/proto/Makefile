
build_csharp_protogen:
	protogen \
	  demo.proto \
	  --proto_path=./ \
	  --proto_path=/Users/ahfu/code/github.com/googleapis/googleapis/ \
	  --csharp_out=../gen/ --grpc=DemoService

generate:
	mkdir -p ../../../build/buf/golang
	mkdir -p ../../../build/buf/csharp
	buf generate . --path Demo.proto

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

dep:
	buf dep update
	