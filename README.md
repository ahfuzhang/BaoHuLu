
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
  * 支持 YAML 格式序列化和反序列化(YAML 序列化时，还会保留 proto 文件中的注释，提高可读性)
  * 支持 URL encode 格式的序列化和反序列化
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
  - 编译期计算：如果某种计算可以在编译期就决定下来，则直接计算成常量放在生成好的代码中
* 分为读对象和写对象，针对读写场景采用了多种优化手段
  - 读对象负责反序列化
  - 写对象负责序列化
* JSON 格式
  - 对于 JSON 序列化
    - 往一个 buffer 中直接追加字符串 (容易做到较好的性能)
  - 对于 JSON 反序列化
    - 使用 JSON 的流式解析库，避免了 DOM 模式的 JSON 解析
      - golang 中使用了：https://github.com/valyala/fastjson
      - csharp 中使用了：`System.Text.Json.Utf8JsonReader`
   - json key 的判断使用 switch-case，编译器会为多个常量字符串生成比 hash 查找更快的方法(eg: perfect hash )
* protobuf 格式
  - 序列化：逐个 tag 追加到一个 buffer 中
  - 反序列化：通过 switch-case 来判断 tag，(部分情况)编译器会编译为 jump-table
* 内存：
  * 重用友好：每个类型都提供 Reset() 方法，便于放到内存池中重用。业务中引入内存池后，可以有效减少 GC 和内存分配。
  * 对于数据序列化：做到 0 内存分配：只需要不断把各个字段追加到目标缓冲区即可
  * 对于数据反序列化：
    * Golang: 对 string 类型进行了 arena 模式的拷贝。(这里的拷贝无法省略，否则引用字符串会因为原始缓冲区失效而发生问题)
* 代码中大量使用了 switch-case 生成 jump table，以及循环展开的技术
  - 避免了分支预测失效
  - 天然内联，顺序执行的长函数对 CPU 的流水线更友好
  - PGO 友好：PGO 场景下，自动调整语句块能够提升性能

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

更多已开发的功能列表，请看：[TODO](./doc/TODO.md)

### Benchmark 数据

#### golang

<table border=1>
<tr>
<td colspan=2> &nbsp; </td>
<td colspan=2 align="center">Linux Amd64<br/>Intel 4.5GHz</td>
<td colspan=2 align="center">MacOS Arm64<br/>macbook m2</td>
</tr>
<tr>
  <td colspan=2> &nbsp; </td>
  <td> 64kb json<br/>(All types)</td>
  <td> 232 bytes json<br/>(Value type only)</td>
  <td> 64kb json<br/>(All types)</td>
  <td> 232 bytes json<br/>(Value type only)</td>
</tr>
<tr>
  <td rowspan=2>JSON<br/></td>
  <td>encode</td>
  <td>1.68 GB/s<br>0 allocs/op<br>13068 times/s</td>
  <td>3.05 GB/s<br>0 allocs/op<br>13157380 times/s</td>
  <td>966.77 MB/s<br>0 allocs/op<br>7539 times/s</td>
  <td>1.89 GB/s<br>0 allocs/op<br>8153138 times/s</td>
</tr>
<tr>
  <td>decode</td>
  <td>466.12 MB/s<br>62 allocs/op<br>3635 times/s</td>
  <td>964.24 MB/s<br>0 allocs/op<br>4156190 times/s</td>
  <td>409.16 MB/s<br>62 allocs/op<br>3191 times/s</td>
  <td>577.07 MB/s<br>0 allocs/op<br>2487367 times/s</td>
</tr>
<tr>
  <td rowspan=2>JSON<br/>(对比 encoding/json)</td>
  <td>encode</td>
  <td><span style="color:red">+948.4%</span></td>
  <td><span style="color:red">+647.1%</span></td>
  <td><span style="color:red">+561.6%</span></td>
  <td><span style="color:red">+450.6%</span></td>
</tr>
<tr>
  <td>decode</td>
  <td><span style="color:red">+222.1%</span></td>
  <td><span style="color:red">+294.7%</span></td>
  <td><span style="color:red">+224.4%</span></td>
  <td><span style="color:red">+231.9%</span></td>
</tr>
<tr>
  <td rowspan=2>JSON<br/>(对比 encoding/json/v2)</td>
  <td>encode</td>
  <td><span style="color:red">+342.9%</span></td>
  <td><span style="color:red">+707.1%</span></td>
  <td><span style="color:red">+206.5%</span></td>
  <td><span style="color:red">+465.8%</span></td>
</tr>
<tr>
  <td>decode</td>
  <td><span style="color:red">+174.8%</span></td>
  <td><span style="color:red">+239.2%</span></td>
  <td><span style="color:red">+167.1%</span></td>
  <td><span style="color:red">+173.4%</span></td>
</tr>
<tr>
  <td rowspan=2>JSON<br/>(对比 bytedance/sonic)</td>
  <td>encode</td>
  <td><span style="color:red">+58.7%</span></td>
  <td><span style="color:red">+158.2%</span></td>
  <td><span style="color:red">+182.8%</span></td>
  <td><span style="color:red">+343.0%</span></td>
</tr>
<tr>
  <td>decode</td>
  <td><span style="color:red">+20.5%</span></td>
  <td><span style="color:red">+53.0%</span></td>
  <td><span style="color:green">-3.9%</span></td>
  <td><span style="color:red">+50.7%</span></td>
</tr>
<tr>
  <td rowspan=2>Protobuf<br/></td>
  <td>encode</td>
  <td>1.24 GB/s<br>0 allocs/op<br>15368 times/s</td>
  <td>1.95 GB/s<br>0 allocs/op<br>33098481 times/s</td>
  <td>826.31 MB/s<br>0 allocs/op<br>10222 times/s</td>
  <td>1.52 GB/s<br>0 allocs/op<br>25728175 times/s</td>
</tr>
<tr>
  <td>decode</td>
  <td>606.00 MB/s<br>66 allocs/op<br>7497 times/s</td>
  <td>1.38 GB/s<br>0 allocs/op<br>23325147 times/s</td>
  <td>596.86 MB/s<br>66 allocs/op<br>7384 times/s</td>
  <td>788.86 MB/s<br>0 allocs/op<br>13370589 times/s</td>
</tr>
</table>

#### C#, Linux + amd64

<table border=1>
<tr>
  <td colspan=2> &nbsp; </td>
  <td> 64kb json<br/>(All types)</td>
  <td> 232 bytes json<br/>(Value type only)</td>
</tr>
<tr>
  <td rowspan=2>JSON<br/>(对比 StdLib)</td>
  <td>encode</td>
  <td><span style="color:red">+9.07%</span></td>
  <td><span style="color:red">+26.72%</span></td>
</tr>
<tr>
  <td>decode</td>
  <td><span style="color:red">+48.34%</span></td>
  <td><span style="color:red">+18.53%</span></td>
</tr>
<tr>
  <td rowspan=2>Protobuf<br/>(对比Grpc.Tools)</td>
  <td>encode</td>
  <td><span style="color:red">+57.30%</span></td>
  <td><span style="color:red">+31.46%</span></td>
</tr>
<tr>
  <td>decode</td>
  <td><span style="color:red">+258%</span></td>
  <td><span style="color:red">+242%</span></td>
</tr>
</table>

See: [doc/Performance](./doc/Performance.md)

## How to use, 命令行说明

* 安装
  - `go install github.com/ahfuzhang/BaoHuLu/cmd/hulu@v0.14.2`

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
      - `-src.csharp_template.dir=$dir`: 自定义 template, 作为输入的模板
      - `-dst.csharp_template.out_dir=$dir`: 对于自定义输入，设定对应的输出目录

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
* 支持 thrift 二进制格式
* 支持 FlatBuffers 二进制格式
* golang:
  - 使用 plan9 汇编 + AVX2 来优化
  - 反序列化场景：对于 map 类型，key 的个数小于一定数量时，使用连续数组来存储 key，顺序查找代替 hash 查找。
  - 反序列化场景：对于 map 类型，key 超过一定数量后，使用定长的 SwissTable + AVX2 来做 hash 查找
  - JSON 反序列化场景：优化库 https://github.com/valyala/fastjson  ✅
* 支持更多编程语言
* 支持更多扩展语法

See: [TODO 及其完成的情况](./doc/TODO.md)

## AI 使用声明

本项目大多数代码由 AI 生成。


## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2026 Fuchun Zhang
