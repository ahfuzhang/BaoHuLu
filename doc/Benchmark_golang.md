
## Macos + arm64

* json 长度 16763 bytes

```text
go test -run=^$ -bench='.*AllTypes.*' -v
goos: darwin
goarch: arm64
pkg: github.com/ahfuzhang/BaoHuLu/examples/Demo
cpu: Apple M2
BenchmarkToJSON_AllTypes_Generated
BenchmarkToJSON_AllTypes_Generated-8               22024             63238 ns/op         265.08 MB/s           0 B/op          0 allocs/op
BenchmarkToJSON_AllTypes_EncodingJSON
BenchmarkToJSON_AllTypes_EncodingJSON-8             5550            234743 ns/op          71.41 MB/s       78011 B/op       2245 allocs/op
BenchmarkToJSON_AllTypes_Sonic
BenchmarkToJSON_AllTypes_Sonic-8                    5589            232264 ns/op          72.17 MB/s       78011 B/op       2245 allocs/op
BenchmarkFromJSON_AllTypes_Generated
BenchmarkFromJSON_AllTypes_Generated-8             15949             78486 ns/op         213.58 MB/s          30 B/op          0 allocs/op
BenchmarkFromJSON_AllTypes_EncodingJSON
BenchmarkFromJSON_AllTypes_EncodingJSON-8           3400            361209 ns/op          46.41 MB/s       74584 B/op       1562 allocs/op
BenchmarkFromJSON_AllTypes_Sonic
BenchmarkFromJSON_AllTypes_Sonic-8                  3028            375791 ns/op          44.61 MB/s       74584 B/op       1562 allocs/op
BenchmarkToProtobuf_AllTypes
BenchmarkToProtobuf_AllTypes-8                     40082             27310 ns/op         665.17 MB/s           0 B/op          0 allocs/op
BenchmarkFromProtobuf_AllTypes
BenchmarkFromProtobuf_AllTypes-8                   24741             47336 ns/op         383.77 MB/s           1 B/op          0 allocs/op
```

# 2026-05-05

## macos + arm64 + golang

```
go test -run=^$ -bench='.*AllTypes.*' -v  -benchtime=60s
goos: darwin
goarch: arm64
pkg: github.com/ahfuzhang/BaoHuLu/examples/Demo
cpu: Apple M2
BenchmarkToJSON_AllTypes_Generated
BenchmarkToJSON_AllTypes_Generated-8              487318            145568 ns/op         841.09 MB/s           0 B/op          0 allocs/op
BenchmarkToJSON_AllTypes_EncodingJSON
BenchmarkToJSON_AllTypes_EncodingJSON-8           115636            620130 ns/op         197.43 MB/s      318262 B/op       7039 allocs/op
BenchmarkToJSON_AllTypes_Sonic
BenchmarkToJSON_AllTypes_Sonic-8                  219277            334594 ns/op         365.92 MB/s      127262 B/op         40 allocs/op
BenchmarkFromJSON_AllTypes_Generated
BenchmarkFromJSON_AllTypes_Generated-8            197461            379330 ns/op         322.77 MB/s      174370 B/op        313 allocs/op
BenchmarkFromJSON_AllTypes_EncodingJSON
BenchmarkFromJSON_AllTypes_EncodingJSON-8          47475           1488364 ns/op          82.26 MB/s      284642 B/op       5598 allocs/op
BenchmarkFromJSON_AllTypes_Sonic
BenchmarkFromJSON_AllTypes_Sonic-8                251452            304255 ns/op         402.41 MB/s      275450 B/op        778 allocs/op
BenchmarkToProtobuf_AllTypes
BenchmarkToProtobuf_AllTypes-8                    709494            105106 ns/op         714.43 MB/s           0 B/op          0 allocs/op
BenchmarkFromProtobuf_AllTypes
BenchmarkFromProtobuf_AllTypes-8                  400658            184964 ns/op         405.98 MB/s      174097 B/op        311 allocs/op
BenchmarkToProtobufVT_AllTypes
BenchmarkToProtobufVT_AllTypes-8                  691990            104545 ns/op         718.26 MB/s           0 B/op          0 allocs/op
BenchmarkFromProtobufVT_AllTypes
BenchmarkFromProtobufVT_AllTypes-8                446850            154568 ns/op       485.81 MB/s       144513 B/op        601 allocs/op

优化后:
BenchmarkFromProtobufVT_AllTypes-8                517660            138442 ns/op         542.40 MB/s      174353 B/op        313 allocs/op

BenchmarkProtobufSizeVT_AllTypes
BenchmarkProtobufSizeVT_AllTypes-8               2524764             26708 ns/op      2811.56 MB/s            0 B/op          0 allocs/op
PASS
ok      github.com/ahfuzhang/BaoHuLu/examples/Demo      863.311s

go test -run=^$ -bench='.*AllTypes.*' -v -benchtime=60s
goos: darwin
goarch: arm64
pkg: github.com/ahfuzhang/BaoHuLu/build/buf/golang/vtprotobuf
cpu: Apple M2
BenchmarkMarshalVT_AllTypes
BenchmarkMarshalVT_AllTypes-8             695641            105851 ns/op         709.40 MB/s       81920 B/op          1 allocs/op
BenchmarkUnmarshalVT_AllTypes
BenchmarkUnmarshalVT_AllTypes-8           396976            177052 ns/op         424.12 MB/s      234050 B/op       1591 allocs/op
PASS
ok      github.com/ahfuzhang/BaoHuLu/build/buf/golang/vtprotobuf        147.354s

```

# 2026-05-10

## 复杂类型

### JSON Performance

| Operation | BaoHuLu | encoding/json | bytedance/sonic |
|:----------|:-------:|:-------------:|:---------------:|
| json encode | 797.14 MB/s<br>0 B/op | 196.69 MB/s<br>318306 B/op<br><span style="color:red">+305.3%</span> | 381.06 MB/s<br>127281 B/op<br><span style="color:red">+109.2%</span> |
| json decode | 341.59 MB/s<br>174353 B/op | 83.80 MB/s<br>285073 B/op<br><span style="color:red">+307.6%</span> | 441.95 MB/s<br>275997 B/op<br><span style="color:green">-22.7%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 761.21 MB/s<br>0 B/op | 799.61 MB/s<br>0 B/op<br><span style="color:green">-4.8%</span> |
| pb decode | 552.99 MB/s<br>174353 B/op | 418.86 MB/s<br>174097 B/op<br><span style="color:red">+32.0%</span> |

## 简单类型

### JSON Performance

| Operation | BaoHuLu | encoding/json | bytedance/sonic |
|:----------|:-------:|:-------------:|:---------------:|
| json encode | 1.81 GB/s<br>0 B/op | 485.76 MB/s<br>352 B/op<br><span style="color:red">+273.3%</span> | 427.58 MB/s<br>374 B/op<br><span style="color:red">+324.1%</span> |
| json decode | 548.61 MB/s<br>0 B/op | 99.27 MB/s<br>328 B/op<br><span style="color:red">+452.7%</span> | 396.74 MB/s<br>571 B/op<br><span style="color:red">+38.3%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 1.83 GB/s<br>0 B/op | 1.15 GB/s<br>0 B/op<br><span style="color:red">+59.8%</span> |
| pb decode | 1.13 GB/s<br>0 B/op | 693.28 MB/s<br>0 B/op<br><span style="color:red">+63.6%</span> |

## AllTypes, 优化了自引用对象的缓存

### JSON Performance

| Operation | BaoHuLu | encoding/json | bytedance/sonic |
|:----------|:-------:|:-------------:|:---------------:|
| json encode | 809.99 MB/s<br>0 B/op | 193.94 MB/s<br>318397 B/op<br><span style="color:red">+317.6%</span> | 361.65 MB/s<br>127442 B/op<br><span style="color:red">+124.0%</span> |
| json decode | 411.54 MB/s<br>59680 B/op | 83.26 MB/s<br>285058 B/op<br><span style="color:red">+394.3%</span> | 438.22 MB/s<br>277483 B/op<br><span style="color:green">-6.1%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 744.06 MB/s<br>0 B/op | 789.26 MB/s<br>0 B/op<br><span style="color:green">-5.7%</span> |
| pb decode | 708.27 MB/s<br>86968 B/op | 510.56 MB/s<br>86840 B/op<br><span style="color:red">+38.7%</span> |
