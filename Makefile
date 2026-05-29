Version ?= v0.9.0

.PHONY: build run test check check-bce

build:
	go build -ldflags "-X main.Version=$(Version)" -o ./build/hulu ./cmd/hulu/

install:
	go build -ldflags "-X main.Version=$(Version)" ./cmd/hulu/

test: run
	go test -v ./... -coverprofile=./build/coverage.out
	go tool cover -html=./build/coverage.out -o ./build/coverage.html

gen: build
	./build/hulu tu \
	  -src=./examples/DemoServer/proto/Demo.proto \
	  -go_out=./build/golang/DemoServer/ \
	  -go_out.with.test \
	  -go_out.with.bench \
	  -csharp_out=./build/csharp/DemoServer/ \
	  -csharp_out.with.test \
	  -csharp_out.with.bench

# deprecated:
#	  -go_out.with.vtprotobuf 

go: build
	./build/hulu tu \
	  -src=./examples/DemoServer/proto/Demo.proto \
	  -go_out=./build/golang/DemoServer/ \
	  -go_out.with.test \
	  -go_out.with.bench

QiWa.rpc:
	./build/hulu tu \
	  -src=./examples/DemoServer/proto/Demo.proto \
	  -csharp_out=./build/csharp/QiWa.rpc/ \
	  -csharp_out.with.test \
	  -csharp_out.with.bench \
	  -src.csharp_template.dir=./templates/csharp/QiWa.rpc/ \
	  -dst.csharp_template.out_dir=./build/csharp/QiWa.rpc/

util_converage_path=build/golang/utils/coverage

utils_coverage:
	mkdir -p $(util_converage_path)
	go test -coverprofile=$(util_converage_path)/coverage.out -v github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils dependencies/golang/utils/...
	go tool cover -html=$(util_converage_path)/coverage.out -o $(util_converage_path)/coverage.html
