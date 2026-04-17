
## macos + arm64

### json encode

BenchmarkDotNet v0.14.0, macOS 26.4.1 (25E253) [Darwin 25.4.0]
Apple M2, 1 CPU, 8 logical and 8 physical cores
.NET SDK 10.0.101
  [Host]     : .NET 10.0.1 (10.0.125.57005), Arm64 RyuJIT AdvSIMD
  DefaultJob : .NET 10.0.1 (10.0.125.57005), Arm64 RyuJIT AdvSIMD


| Method                     | Mean     | Error    | StdDev   | Ratio | RatioSD | Rank | MB/s  | Gen0   | Allocated | Alloc Ratio |
|--------------------------- |---------:|---------:|---------:|------:|--------:|-----:|------:|-------:|----------:|------------:|
| BaoHuLu.ToJSON             | 36.65 us | 0.674 us | 0.922 us |  1.00 |    0.04 |    1 | 635.5 |      - |     392 B |        1.00 |
| System.Text.Json.Serialize | 38.32 us | 0.732 us | 0.648 us |  1.05 |    0.03 |    1 | 607.8 | 5.9814 |   50288 B |      128.29 |

### json encode

| Method                       | Mean      | Error    | StdDev   | Ratio | RatioSD | Rank | MB/s  | Gen0    | Gen1   | Allocated | Alloc Ratio |
|----------------------------- |----------:|---------:|---------:|------:|--------:|-----:|------:|--------:|-------:|----------:|------------:|
| System.Text.Json.Deserialize |  88.52 us | 1.732 us | 3.500 us |  0.66 |    0.03 |    1 | 263.1 | 20.0195 | 6.7139 | 163.64 KB |        0.84 |
| BaoHuLu.FromJSON             | 135.10 us | 2.645 us | 4.118 us |  1.00 |    0.04 |    2 | 172.4 | 23.6816 | 3.9063 | 195.31 KB |        1.00 |

### protobuf encode

| Method                 | Mean     | Error    | StdDev   | Ratio | RatioSD | Rank | MB/s   | Gen0   | Allocated | Alloc Ratio |
|----------------------- |---------:|---------:|---------:|------:|--------:|-----:|-------:|-------:|----------:|------------:|
| BaoHuLu.ToProtobuf     | 23.17 us | 0.462 us | 0.474 us |  1.00 |    0.03 |    1 | 1025.0 |      - |         - |          NA |
| Grpc.Tools.ToByteArray | 31.38 us | 0.620 us | 0.828 us |  1.35 |    0.04 |    2 |  756.7 | 2.9907 |   25504 B |          NA |

### protobuf decode

| Method               | Mean     | Error    | StdDev   | Ratio | RatioSD | Rank | MB/s  | Gen0    | Gen1   | Allocated | Alloc Ratio |
|--------------------- |---------:|---------:|---------:|------:|--------:|-----:|------:|--------:|-------:|----------:|------------:|
| Grpc.Tools.ParseFrom | 56.30 us | 1.094 us | 1.497 us |  0.63 |    0.02 |    1 | 421.8 | 24.9023 | 8.2397 | 203.92 KB |        1.04 |
| BaoHuLu.FromProtobuf | 89.93 us | 1.168 us | 0.912 us |  1.00 |    0.01 |    2 | 264.0 | 23.8037 | 4.1504 | 195.31 KB |        1.00 |

