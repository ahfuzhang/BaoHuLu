
build:
	protogen \
	  demo.proto \
	  --proto_path=./ \
	  --proto_path=/Users/ahfu/code/github.com/googleapis/googleapis/ \
	  --csharp_out=../gen/ --grpc=DemoService

generate-golang:
	mkdir -p ../build/
	buf generate . --path demo.proto

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

dep:
	buf dep update
	