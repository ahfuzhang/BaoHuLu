
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

# 2026-05-11

```
GOEXPERIMENT=jsonv2 go test -run=^$ -bench='.*AllTypes.*' -v  -benchtime=60s
goos: darwin
goarch: arm64
pkg: github.com/ahfuzhang/BaoHuLu/examples/Demo
cpu: Apple M2
BenchmarkToJSON_AllTypes_Generated
BenchmarkToJSON_AllTypes_Generated-8              480729            148988 ns/op         821.78 MB/s           0 B/op          0 allocs/op
BenchmarkToJSON_AllTypes_EncodingJSON
BenchmarkToJSON_AllTypes_EncodingJSON-8            92494            774092 ns/op         158.17 MB/s      411042 B/op       3260 allocs/op
BenchmarkToJSON_AllTypes_Sonic
BenchmarkToJSON_AllTypes_Sonic-8                  223558            321565 ns/op         380.75 MB/s      127172 B/op         40 allocs/op
BenchmarkToJSON_AllTypes_EncodingJSONv2
BenchmarkToJSON_AllTypes_EncodingJSONv2-8         201386            386256 ns/op         316.98 MB/s      124377 B/op         77 allocs/op
BenchmarkFromJSON_AllTypes_Generated
BenchmarkFromJSON_AllTypes_Generated-8            193580            314195 ns/op         389.68 MB/s       59698 B/op         59 allocs/op
BenchmarkFromJSON_AllTypes_EncodingJSON
BenchmarkFromJSON_AllTypes_EncodingJSON-8          78952            917291 ns/op         133.47 MB/s      226292 B/op       1363 allocs/op
BenchmarkFromJSON_AllTypes_Sonic
BenchmarkFromJSON_AllTypes_Sonic-8                261181            277900 ns/op         440.57 MB/s      275452 B/op        778 allocs/op
BenchmarkFromJSON_AllTypes_EncodingJSONv2
BenchmarkFromJSON_AllTypes_EncodingJSONv2-8        96312            757191 ns/op         161.70 MB/s      226290 B/op       1360 allocs/op
BenchmarkToProtobuf_AllTypes
BenchmarkToProtobuf_AllTypes-8                    756259             96790 ns/op         775.81 MB/s           0 B/op          0 allocs/op
BenchmarkFromProtobuf_AllTypes
BenchmarkFromProtobuf_AllTypes-8                  480618            150222 ns/op         499.87 MB/s       86841 B/op        155 allocs/op
BenchmarkToProtobufVT_AllTypes
BenchmarkToProtobufVT_AllTypes-8                  731880            100051 ns/op         750.53 MB/s           0 B/op          0 allocs/op
BenchmarkFromProtobufVT_AllTypes
BenchmarkFromProtobufVT_AllTypes-8                668502            111442 ns/op         673.81 MB/s       86968 B/op        156 allocs/op
BenchmarkProtobufSizeVT_AllTypes
BenchmarkProtobufSizeVT_AllTypes-8               2688649             26674 ns/op        2815.16 MB/s           0 B/op          0 allocs/op
PASS
ok      github.com/ahfuzhang/BaoHuLu/examples/Demo      1008.093s
```

# 2026-05-11 14:30 负优化, ToJSON

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 746.12 MB/s<br>0 allocs/op | 143.13 MB/s<br>3260 allocs/op<br><span style="color:red">+421.3%</span> | 299.89 MB/s<br>77 allocs/op<br><span style="color:red">+148.8%</span> | 368.06 MB/s<br>40 allocs/op<br><span style="color:red">+102.7%</span> |
| json decode | 402.44 MB/s<br>59 allocs/op | 132.31 MB/s<br>1360 allocs/op<br><span style="color:red">+204.2%</span> | 159.26 MB/s<br>1360 allocs/op<br><span style="color:red">+152.7%</span> | 443.73 MB/s<br>779 allocs/op<br><span style="color:green">-9.3%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 755.80 MB/s<br>0 allocs/op | 786.21 MB/s<br>0 allocs/op<br><span style="color:green">-3.9%</span> |
| pb decode | 710.09 MB/s<br>156 allocs/op | 504.12 MB/s<br>155 allocs/op<br><span style="color:red">+40.9%</span> |

# 2026-05-11 15:00, ToJSON, 改回之前的写法

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 799.95 MB/s<br>0 allocs/op | 148.91 MB/s<br>3260 allocs/op<br><span style="color:red">+437.2%</span> | 342.45 MB/s<br>77 allocs/op<br><span style="color:red">+133.6%</span> | 377.76 MB/s<br>40 allocs/op<br><span style="color:red">+111.8%</span> |
| json decode | 410.88 MB/s<br>59 allocs/op | 136.00 MB/s<br>1359 allocs/op<br><span style="color:red">+202.1%</span> | 165.25 MB/s<br>1359 allocs/op<br><span style="color:red">+148.6%</span> | 450.30 MB/s<br>779 allocs/op<br><span style="color:green">-8.8%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 770.90 MB/s<br>0 allocs/op | 798.64 MB/s<br>0 allocs/op<br><span style="color:green">-3.5%</span> |
| pb decode | 722.04 MB/s<br>156 allocs/op | 512.73 MB/s<br>155 allocs/op<br><span style="color:red">+40.8%</span> |

# 2026-05-11 15:00, 优化 FromJSON(轻微性能下降)

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 821.47 MB/s<br>0 allocs/op | 157.05 MB/s<br>3260 allocs/op<br><span style="color:red">+423.1%</span> | 343.35 MB/s<br>77 allocs/op<br><span style="color:red">+139.3%</span> | 380.31 MB/s<br>40 allocs/op<br><span style="color:red">+116.0%</span> |
| json decode | 407.00 MB/s<br>59 allocs/op | 135.63 MB/s<br>1362 allocs/op<br><span style="color:red">+200.1%</span> | 164.66 MB/s<br>1362 allocs/op<br><span style="color:red">+147.2%</span> | 450.65 MB/s<br>779 allocs/op<br><span style="color:green">-9.7%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 772.13 MB/s<br>0 allocs/op | 796.72 MB/s<br>0 allocs/op<br><span style="color:green">-3.1%</span> |
| pb decode | 726.75 MB/s<br>156 allocs/op | 503.09 MB/s<br>155 allocs/op<br><span style="color:red">+44.5%</span> |

# 2026-05-11 16:00,  简单的值类型的性能

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 1.88 GB/s<br>0 allocs/op | 344.68 MB/s<br>3 allocs/op<br><span style="color:red">+445.9%</span> | 352.95 MB/s<br>3 allocs/op<br><span style="color:red">+433.1%</span> | 432.80 MB/s<br>3 allocs/op<br><span style="color:red">+334.8%</span> |
| json decode | 530.36 MB/s<br>0 allocs/op | 173.36 MB/s<br>1 allocs/op<br><span style="color:red">+205.9%</span> | 222.42 MB/s<br>1 allocs/op<br><span style="color:red">+138.4%</span> | 403.26 MB/s<br>4 allocs/op<br><span style="color:red">+31.5%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 1.85 GB/s<br>0 allocs/op | 1.20 GB/s<br>0 allocs/op<br><span style="color:red">+53.6%</span> |
| pb decode | 1.16 GB/s<br>0 allocs/op | 717.28 MB/s<br>0 allocs/op<br><span style="color:red">+61.7%</span> |

# 2026-05-12， linux, amd64, intel 4.5GHz

## all types, 64kb

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 1.42 GB/s<br>0 allocs/op | 166.81 MB/s<br>3260 allocs/op<br><span style="color:red">+752.6%</span> | 392.02 MB/s<br>77 allocs/op<br><span style="color:red">+262.8%</span> | 1.15 GB/s<br>40 allocs/op<br><span style="color:red">+23.2%</span> |
| json decode | 502.32 MB/s<br>59 allocs/op | 161.21 MB/s<br>1360 allocs/op<br><span style="color:red">+211.6%</span> | 184.86 MB/s<br>1360 allocs/op<br><span style="color:red">+171.7%</span> | 430.73 MB/s<br>1060 allocs/op<br><span style="color:red">+16.6%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 927.50 MB/s<br>0 allocs/op | 1.12 GB/s<br>0 allocs/op<br><span style="color:green">-17.0%</span> |
| pb decode | 700.54 MB/s<br>156 allocs/op | 551.38 MB/s<br>155 allocs/op<br><span style="color:red">+27.1%</span> |

## value types, 260 bytes

### JSON Performance

| Operation | BaoHuLu | encoding/json | encoding/json/v2 | bytedance/sonic |
|:----------|:-------:|:-------------:|:----------------:|:---------------:|
| json encode | 3.02 GB/s<br>0 allocs/op | 408.92 MB/s<br>3 allocs/op<br><span style="color:red">+638.4%</span> | 384.85 MB/s<br>3 allocs/op<br><span style="color:red">+684.6%</span> | 1.18 GB/s<br>3 allocs/op<br><span style="color:red">+156.5%</span> |
| json decode | 852.70 MB/s<br>0 allocs/op | 268.93 MB/s<br>1 allocs/op<br><span style="color:red">+217.1%</span> | 325.87 MB/s<br>1 allocs/op<br><span style="color:red">+161.7%</span> | 630.11 MB/s<br>3 allocs/op<br><span style="color:red">+35.3%</span> |

### Protobuf Performance

| Operation | BaoHuLu VT (baseline) | BaoHuLu |
|:----------|:---------------------:|:-------:|
| pb encode | 1.98 GB/s<br>0 allocs/op | 1.70 GB/s<br>0 allocs/op<br><span style="color:red">+16.6%</span> |
| pb decode | 1.99 GB/s<br>0 allocs/op | 1.25 GB/s<br>0 allocs/op<br><span style="color:red">+58.6%</span> |


