module github.com/ahfuzhang/BaoHuLu/internal/integration/go-contract

go 1.23

require github.com/ahfuzhang/BaoHuLu/examples/Demo v0.0.0

require github.com/ahfuzhang/BaoHuLu v0.16.0 // indirect

replace github.com/ahfuzhang/BaoHuLu/examples/Demo => ../../../build/generated/golang

replace github.com/ahfuzhang/BaoHuLu => ../../..
