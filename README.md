
# BaoHuLu (宝葫芦, Magic Calabash)

> 在中国的经典动画片 《金刚葫芦娃》 中，宝葫芦是七娃的法器，可以把妖怪吸进葫芦里。

> The name originates from the classic Chinese animated film "Calabash Brothers". The magic calabash is the magic weapon of the seventh brother, which can suck monsters into the calabash.

![](./doc/images/3.png)

> 今年开始，我准备为公司的直播等平台做一个高性能的 C# 的 RPC 框架，以便在直播打赏、有价礼物消费等场景可以做到类似高频交易的效果。于是我规划了“七娃”这个项目：https://github.com/ahfuzhang/QiWa
>
> 为了让这个框架高性能，我希望能够做到整个处理期间 0 内存分配，从而避免因为 GC 导致的交易延迟不稳定。而类似 grpc 一类的框架封装太厚重，框架内就有大量的内存分配。于是我计划，在 protobuf/JSON 序列化和反序列化的阶段就开始自主开发。
> 同时，为了让 QiWa 这个框架更好用，我希望使用者定义好 proto 文件后，框架就生成好基本的处理代码，然后开发者只需要在已经反序列化好的回调函数中填充业务逻辑即可。
>
> 因此，为了配合 QiWa 项目，我以七娃的宝葫芦为名字，建设一个命令行代码生成工具：BaoHuLu.

## 目标

* 一个基于 golang 实现的命令行工具，用于把 proto 文件生成多种编程语言的高性能的数据序列化代码
* 根据 *.proto 文件，生成多种语言的数据序列化/反序列化的代码
  * 支持 protobuf 二进制格式的序列化和反序列
  * 支持 JSON 格式的序列化和反序列化
  * (在使用人数很多的情况下可以考虑) 支持 Thrift 二进制格式的序列化和反序列化
  * 考虑支持 FlatBuffers
* 为 QiWa RPC 框架，生成从请求处理 -> 反序列化 -> 触发回调 -> 对响应序列化 这个过程的脚手架代码
* 为多种编程语言生成代码：
  - golang
  - csharp

## 性能优势

使用了如下的方法做到了高性能的序列化/反序列化。

* 完全的编译期决定解析流程，避免了在运行期使用反射等功能
  - 在代码生成阶段，直接把 tag 和数据类型生成到解析代码中，无任何反射的开销。编译期决定了大多数事情，反序列化会精准的逐个读出 tag，并精准的放到确定的字段。
* 分为读对象和写对象，针对读写场景采用了多种优化手段
  - 读对象负责反序列化
  - 写对象负责序列化
* 对于 JSON 序列化
  - 往一个 buffer 中直接追加字符串 (容易做到较好的性能)
* 对于 JSON 反序列化
  - 使用 JSON 的流式解析库，避免了 DOM 模式的 JSON 解析
    - golang 中使用了：https://github.com/valyala/fastjson
    - csharp 中使用了：`System.Text.Json.Utf8JsonReader`
  - json key 的判断使用 switch-case，编译器会为多个常量字符串生成 perfect hash  
* 对于 protobuf:
  - 序列化：逐个 tag 追加到一个 buffer 中
  - 反序列化：通过 switch-case 来判断 tag，编译器会编译为 jump-table
* 内存：
  * 重用友好：每个类型都提供 Reset() 方法，便于放到内存池中重用。业务中引入内存池后，可以有效减少 GC 和内存分配。
  * 对于数据序列化：做到 0 内存分配：只需要不断把各个字段追加到目标缓冲区即可
  * 对于数据反序列化：
    * Golang: 保持对原始缓冲区的引用，然后使用 unsafe 代码直接指向原始缓冲区中的数据段。类似 string 类型做到不分配新的内存。
* 代码中大量使用了 switch-case 生成 jump table，以及循环展开的技术
  - 避免了分支预测失效
  - 天然内联，顺序执行的长函数对 CPU 的流水线更友好

## 已实现的功能
* Proto 文件解析
  * 支持基本的 proto 语法，但是部分容易引起歧义的、性能不友好的、鸡肋的特性不会支持
* Proto 语法的扩展
  * 以注释的形式，支持扩展的语法。具体定义请看: [./doc/Extensions.md](./doc/Extensions.md)
    * 扩展语法的例子如下
    * `// @method=GET/POST` 声明某个 service 的 method 对 http method 的支持
    * `// @jsonName=xx` 声明 message 中的某个 field 对应的 JSON key 名字是什么
* Message 的代码生成
  * 每个 message 会生成用于读和用于写的两个类型
    * ReadonlyXX: 提供数据反序列化的方法，以只读的形式访问数据
    * XX: 用于序列化前的数据填充，与传统的 protoc 生成的类型相似。提供了数据序列化的方法
  * 只读类型，提供了如下方法：
    * FromProtobuf(): 输入一段 buffer，对数据进行 protobuf 二进制反序列化
    * FromJSON(): 输入一段 buffer，对数据进行 JSON 文本的反序列化
    * Clone():  把 Readonly 类型复制为可写的类型
    * Reset(): 清空成员，便于这个类型放到内存池中供下次重用
  * 普通类型，提供了如下方法：
    * ToProtobuf(): 提供写缓冲区，把各个字段以 protobuf 二进制格式序列化到缓冲区中
    * ToJSON(): 提供写缓冲区，把各个字段以 JSON 文本格式序列化到缓冲区中
    * Reset(): 清空成员，便于这个类型放到内存池中供下次重用
* 开发语言支持：
  * golang
    - 反序列化阶段，保持了对原始缓冲区的引用。因此解析 string 和 bytes 类型不会导致内存分配。
      - 如果输入 buffer 可能被覆盖，则可以使用 `FromProtobufWithCopy()` 和 `FromJSONWithCopy()` 代替。这两个方法会把输入缓冲区复制到类型内部。
    - 如果想根据读对象来修改并重新序列化，提供了 `Clone()` 方法来把读对象的各个成员复制到写对象。写对象中使用了 arena 内存分配技巧，避免了分配大量的小对象。
    - 重用对象的情况下：`FromProtobuf()`, `FromJSON()`, `ToProtobuf()`, `ToJSON()` 等方法在运行期间零内存分配  
    - 支持生成每个类型对应的 test 和 benchmark 代码
  * csharp
    - 尽可能减少 utf-8 到 utf-16 的转换
    - 尽可能减少发生 throw exception, 例如使用 TryParse() 代替 Parse()
    - 支持生成每个类型对应的 test 和 benchmark 代码
    - 使用值类型，便于利用栈空间来减少分配
* 其他:
  * 优化 struct 中的成员布局，GC 扫描友好，且节约内存

### Benchmark 数据

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
| ---- | ---- | ---- |
| protobuf encode | 1505.78 MB/s<br/>0 allocs/op | 364.38 MB/s(快 4.13 倍)<br/>1 allocs/op | 1815.16 MB/s(慢  17.04 %)<br/>1 allocs/op<br/>MarshalToVT:<br/>2575.42 MB/s(慢  41.53 %)<br/> 0 allocs/op |
| protobuf decode | 837.66 MB/s<br/>0 allocs/op | 507.02 MB/s(快 65.2% )<br/>0 allocs/op | 1305.27 MB/s(慢 35.82 %)<br/>0 allocs/op |

#### Macos + arm64, M2, csharp

* DotNet SDK 10.0.101

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

## 明确不支持的功能

* proto 文件解析：
  * 不支持关键字： import, required, optional, oneof, extensions, extend, stream, option
  * 为了避免工具太复杂，一些经典的外部语法扩展先 **不支持** (暂不支持通过 import 导入其他 proto 文件)
    * Google 定义的扩展数据类型
    * 我最喜欢的 gogo proto 中的各种 extension


## 开发中的功能

* Serivce 代码生成
  * 短期：仅针对 QiWa 项目生成 RPC 相关的代码
  * 后期：使用者可以自己以 golang template 的语法提供模板，然后生成代码。
  * 暂不支持 stream 关键字
* 开发语言支持：
  * 使用者可以自己以 golang template 的语法提供模板，然后生成代码。

### 远期目标
* 支持 yaml 的序列化和反序列化
* 支持 thrift 二进制格式
* 支持 FlatBuffers 二进制格式
* golang:
  - 使用 plan9 汇编 + AVX2 来优化
  - 反序列化场景：对于 map 类型，key 的个数小于一定数量时，使用连续数组来存储 key，顺序查找代替 hash 查找。
  - 反序列化场景：对于 map 类型，key 超过一定数量后，使用定长的 SwissTable + AVX2 来做 hash 查找
  - JSON 反序列化场景：优化库 https://github.com/valyala/fastjson
* 支持更多编程语言
* 支持更多扩展语法

## How to use, 命令行说明

* 安装
  - `go install github.com/ahfuzhang/BaoHuLu/cmd/hulu@v0.1.1`

* 语法检查
  * `hulu xi ./xx.proto`
  * 或者：`hulu check ./xx.proto`

<img src="./doc/images/1.png" style="zoom:50%;" />



* 生成代码：
  * `hulu tu -src=./xx.proto -go_out=xx_dir -csharp_out=xx_dir _qiwa_out=xx_dir`
    - 或者 `hulu generate ...`
    - `-src=input.proto`
    - `-go_out=$dir`: 把 golang 代码输出到某个目录
      - `-go_out.with.test`: 生成 golang 的测试代码
      - `-go_out.with.bench`: 生成 golang 的 benchmark 代码
    - `-csharp_out=$dir`: 把 csharp 代码输出到某个目录
      - `-csharp_out.with.test`:   生成 csharp 的测试代码
      - `-csharp_out.with.bench`: 生成 csharp 的 benchmark 代码

  ![](./doc/images/2.png)

* 生成测试代码
  - `make gen`

aka:

```bash
hulu tu \
	  -src=./examples/DemoServer/proto/Demo.proto \
	  -go_out=./build/golang/DemoServer/ \
	  -go_out.with.test \
	  -go_out.with.bench \
	  -csharp_out=./build/csharp/DemoServer/ \
	  -csharp_out.with.test \
	  -csharp_out.with.bench
```

## AI 使用声明

本项目 99% 以上的代码由 AI 生成。


## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2026 Fuchun Zhang
