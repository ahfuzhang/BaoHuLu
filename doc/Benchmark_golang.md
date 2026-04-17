
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
