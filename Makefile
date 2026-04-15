.PHONY: build run test 

build:
	go build -o ./build/hulu ./cmd/hulu/

test: run
	go test -v ./... -coverprofile=./build/coverage.out
	go tool cover -html=./build/coverage.out -o ./build/coverage.html

check:
	./build/hulu xi \
	  -src=./examples/DemoServer/proto/Demo.proto

gen:
	./build/hulu tu \
	  -src=./examples/DemoServer/proto/Demo.proto \
	  -go_out=./build/golang/DemoServer/ \
	  -csharp_out=./build/csharp/DemoServer/ \
