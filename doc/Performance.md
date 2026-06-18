
#### Macos + arm64, M2, golang
* 16kb 长度的 JSON
  - goos: darwin
  - goarch: arm64
  - cpu: Apple M2
  - go version go1.26.1
  - JSON 序列化后长度 16763 字节，包含完整的 19 种 protobuf 的数据类型，以及所有允许的 map key 类型(bool 类型的 key 除外)

| 测试项 | BaoHuLu | encoding/json | bytedance/sonic |
| ---- | ---- | ---- | ---- |
| json encode | 265.08 MB/s<br/>0 allocs/op | 71.41 MB/s(快 3.71 倍)<br/>2245 allocs/op | 72.17 MB/s(快 3.67 倍)<br/>2245 allocs/op |
| json decode | 213.58 MB/s<br/>0 allocs/op | 46.41 MB/s(快 4.60 倍)<br/>1562 allocs/op | 44.61 MB/s(快 4.79 倍)<br/>1562 allocs/op |

| 测试项 | BaoHuLu | google protobuf | github.com/planetscale/vtprotobuf |
| ---- | ---- | ---- | ---- |
| protobuf encode | 665.17 MB/s<br/>0 allocs/op | 157.25 MB/s(快 4.23 倍)<br/>4445 allocs/op | 654.29 MB/s(快 1.66 %)<br/>1 allocs/op |
| protobuf decode | 383.77 MB/s<br/>0 allocs/op | 108.34 MB/s(快 3.54 倍)<br/>2649 allocs/op | 428.57 MB/s(慢 10.45 %)<br/>326 allocs/op |

* 232 字节的 JSON，只有值类型，没有引用类型

| 测试项 | BaoHuLu | encoding/json | bytedance/sonic |
| ---- | ---- | ---- | ---- |
| json encode | 1908.69 MB/s<br/>0 allocs/op | 513.88 MB/s(快 3.71 倍)<br/>2 allocs/op | 515.77 MB/s(快 3.70 倍)<br/>2 allocs/op |
| json decode | 561.07 MB/s<br/>0 allocs/op | 99.13 MB/s(快 5.66 倍)<br/>2 allocs/op | 103.88 MB/s(快 5.40 倍)<br/>2 allocs/op |

| 测试项 | BaoHuLu | google protobuf | github.com/planetscale/vtprotobuf |
| ---- | ---- | ---- | ---- |
| protobuf encode | 1505.78 MB/s<br/>0 allocs/op | 364.38 MB/s(快 4.13 倍)<br/>1 allocs/op | 1815.16 MB/s(慢  17.04 %)<br/>1 allocs/op<br/>MarshalToVT:<br/>2575.42 MB/s(慢  41.53 %)<br/> 0 allocs/op |
| protobuf decode | 837.66 MB/s<br/>0 allocs/op | 507.02 MB/s(快 65.2% )<br/>0 allocs/op | 1305.27 MB/s(慢 35.82 %)<br/>0 allocs/op |

#### Macos + arm64, M2, csharp

* DotNet SDK 10.0.101, 16kb 长度的 JSON

```
BenchmarkDotNet v0.14.0, macOS 26.4.1 (25E253) [Darwin 25.4.0]
Apple M2, 1 CPU, 8 logical and 8 physical cores
.NET SDK 10.0.101
  [Host]     : .NET 10.0.1 (10.0.125.57005), Arm64 RyuJIT AdvSIMD
  DefaultJob : .NET 10.0.1 (10.0.125.57005), Arm64 RyuJIT AdvSIMD
```

| 测试项 | BaoHuLu | System.Text.Json.Serialize/Deserialize |
| ---- | ---- | ---- |
| json encode | 635.5 MB/s<br/>392 B Allocated | 607.8 MB/s(快 4.56% )<br/>50288 B Allocated |
| json decode | 339.8 MB/s<br/>33.02 KB Allocated | 270.8 MB/s(快 25.48% )<br/>163.64 KB Allocated |

| 测试项 | BaoHuLu | Grpc.Tools |
| ---- | ---- | ---- |
| protobuf encode | 1025.0 MB/s<br/> 0 Allocated | 756.7 MB/s(快 35.46% )<br/> 25504 B Allocated |
| protobuf decode | 997.1 MB/s<br/> 33.02 KB Allocated | 424.3 MB/s(快 35.0% )<br/> 203.92 KB Allocated |

* 232 字节的 JSON，只有值类型，没有引用类型

| 测试项 | BaoHuLu | System.Text.Json.Serialize/Deserialize |
| ---- | ---- | ---- |
| json encode | 795.1 MB/s<br/>0 B Allocated | 707.2 MB/s(快 12.43% )<br/>0 B Allocated |
| json decode | 474.5 MB/s<br/>0 B Allocated | 472.7 MB/s(快 0.38% )<br/>96 B Allocated |

| 测试项 | BaoHuLu | Grpc.Tools |
| ---- | ---- | ---- |
| protobuf encode | 845.0 MB/s<br/> 0 Allocated | 788.2 MB/s(快 7.2% )<br/> 152 B Allocated |
| protobuf decode | 885.9 MB/s<br/> 0 B Allocated | 618.1 MB/s(快 43.33% )<br/> 272 B Allocated |

### Linux + amd64, `Intel(R) Core(TM) Ultra 7 265KF(4.5GHz)`, golang 1.26.2

* 16kb 长度的 JSON
  - goos: linux
  - goarch: amd64
  - cpu: Intel(R) Core(TM) Ultra 7 265KF
  - go version 1.26.2
  - JSON 序列化后长度 16763 字节，包含完整的 19 种 protobuf 的数据类型，以及所有允许的 map key 类型(bool 类型的 key 除外)

| 测试项 | BaoHuLu | encoding/json | bytedance/sonic |
| ---- | ---- | ---- | ---- |
| json encode | 781.89 MB/s<br/>0 allocs/op | 100.07 MB/s(快 7.81 倍)<br/>2245 allocs/op | 783.09 MB/s(慢 0.15% )<br/>14 allocs/op |
| json decode | 521.01 MB/s<br/>0 allocs/op | 75.95 MB/s(快 6.86 倍)<br/>1562 allocs/op | 273.18 MB/s(快 1.91 倍)<br/>241 allocs/op |

| 测试项 | BaoHuLu | google protobuf | github.com/planetscale/vtprotobuf | planetscale/vtprotobuf(MarshalToVT，重用内存) |
| ---- | ---- | ---- | ---- | ---- |
| protobuf encode | 1327.84 MB/s<br/>0 allocs/op | 157.16 MB/s(快 8.45 倍)<br/>4445 allocs/op | 618.84 MB/s(快 2.15 倍)<br/>1 allocs/op | 868.54 MB/s(快 52.88% )<br/>0 allocs/op |
| protobuf decode | 1022.86 MB/s<br/>0 allocs/op | 116.40 MB/s(快 8.79 倍)<br/>2649 allocs/op | 385.58 MB/s(快 2.65 倍)<br/>326 allocs/op | - |

* 232 字节的 JSON，只有值类型，没有引用类型

| 测试项 | BaoHuLu | encoding/json | bytedance/sonic |
| ---- | ---- | ---- | ---- |
| json encode | 3280.19 MB/s<br/>0 allocs/op | 594.83 MB/s(快 5.51 倍)<br/>2 allocs/op | 1255.40 MB/s(快  2.61 倍 )<br/>3 allocs/op |
| json decode | 860.12 MB/s<br/>0 allocs/op | 139.99 MB/s(快 6.14 倍)<br/>4 allocs/op | 684.78 MB/s(快 1.26 倍)<br/>2 allocs/op |

| 测试项 | BaoHuLu | google protobuf | github.com/planetscale/vtprotobuf | planetscale/vtprotobuf(MarshalToVT，重用内存) |
| ---- | ---- | ---- | ---- | ---- |
| protobuf encode | 2396.52 MB/s<br/>0 allocs/op | 425.91 MB/s(快 5.63 倍)<br/>1 allocs/op | 1335.21 MB/s(快 79.49 %)<br/>1 allocs/op | 3296.11 MB/s(慢 27.29% )<br/>0 allocs/op |
| protobuf decode | 1490.63 MB/s<br/>0 allocs/op | 705.34 MB/s(快 2.11 倍)<br/>0 allocs/op | 2610.65 MB/s(慢 42.90 %)<br/>0 allocs/op | - |

### Linux + amd64, `Intel(R) Core(TM) Ultra 7 265KF(4.5GHz)`, DotNet 10.0.102

* 16kb 长度的 JSON

| 测试项 | BaoHuLu | System.Text.Json.Serialize/Deserialize |
| ---- | ---- | ---- |
| json encode | 987.2 MB/s<br/>392 B Allocated | 905.1 MB/s(快 9.07% )<br/>50256 B Allocated |
| json decode | 655.5 MB/s<br/>33.02 KB Allocated | 441.9 MB/s(快 48.34% )<br/>163.63 KB Allocated |

| 测试项 | BaoHuLu | Grpc.Tools |
| ---- | ---- | ---- |
| protobuf encode | 2293.5 MB/s<br/> 0 Allocated | 1458.0 MB/s(快 57.30 % )<br/> 25504 B Allocated |
| protobuf decode | 1771.3 MB/s<br/> 33.02 KB Allocated | 686.4 MB/s(快 2.58 倍 )<br/> 203.77 KB Allocated |

* 232 字节的 JSON，只有值类型，没有引用类型

| 测试项 | BaoHuLu | System.Text.Json.Serialize/Deserialize |
| ---- | ---- | ---- |
| json encode | 1305.2 MB/s<br/>0 B Allocated | 1030.0 MB/s(快 26.72% )<br/>584 B Allocated |
| json decode | 912.9 MB/s<br/>0 B Allocated | 770.2 MB/s(快 18.53% )<br/>96 B Allocated |

| 测试项 | BaoHuLu | Grpc.Tools |
| ---- | ---- | ---- |
| protobuf encode | 2315.5 MB/s<br/> 0 Allocated | 1761.4 MB/s(快 31.46 % )<br/> 152 B Allocated |
| protobuf decode | 2365.7 MB/s<br/> 0 B Allocated | 976.8 MB/s(快 2.42 倍 )<br/> 272 B Allocated |

