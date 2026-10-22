Version ?= v0.16.0
TestProto = internal/integration/testdata/AllTypes.proto
GeneratedRoot = build/generated

.PHONY: build install test gen go csharp QiWa.rpc generated-build utils_coverage

build:
	go build -ldflags "-X main.Version=$(Version)" -o ./build/hulu ./cmd/hulu/

install:
	go install -ldflags "-X main.Version=$(Version)" ./cmd/hulu/

test:
	go test ./cmd/... ./internal/... ./dependencies/... ./templates/...

gen: go csharp

go: build
	./build/hulu -src=$(TestProto) -template.dir=builtin:golang -out=$(GeneratedRoot)/golang

csharp: build
	./build/hulu -src=$(TestProto) -template.dir=builtin:csharp -out=$(GeneratedRoot)/csharp

QiWa.rpc: build
	./build/hulu -src=examples/DemoServer/proto/Demo.proto -template.dir=builtin:csharp/QiWa.rpc -out=$(GeneratedRoot)/rpc

generated-build: gen
	cd $(GeneratedRoot)/golang && go mod edit -replace=github.com/ahfuzhang/BaoHuLu=$(CURDIR) && go mod tidy && go test -run '^$$' ./...
	cd $(GeneratedRoot)/csharp && dotnet build AllTypes.csproj --nologo
	cd $(GeneratedRoot)/csharp && dotnet build Tests/AllTypes.Tests.csproj --nologo
	cd $(GeneratedRoot)/csharp && dotnet build Benchmarks/AllTypes.Benchmarks.csproj --nologo

util_converage_path=build/golang/utils/coverage

utils_coverage:
	mkdir -p $(util_converage_path)
	go test -coverprofile=$(util_converage_path)/coverage.out -v ./dependencies/golang/utils/...
	go tool cover -html=$(util_converage_path)/coverage.out -o $(util_converage_path)/coverage.html
