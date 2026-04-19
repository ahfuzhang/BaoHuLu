.PHONY: build run test check check-bce

build:
	go build -o ./build/hulu ./cmd/hulu/

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
