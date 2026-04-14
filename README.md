
# BaoHuLu (宝葫芦, Magic Calabash)

> 在中国的经典动画片 《金刚葫芦娃》 中，宝葫芦是七娃的法器，可以把妖怪吸进葫芦里。

> The name originates from the classic Chinese animated film "Calabash Brothers". The magic calabash is the magic weapon of the seventh brother, which can suck monsters into the calabash.

![](./doc/images/3.png)

> 今年开始，我准备为公司的直播等平台做一个高性能的 C# 的 RPC 框架，以便在直播打赏、有价礼物消费等场景可以做到类似高频交易的效果。于是我规划了“七娃”这个项目：https://github.com/ahfuzhang/QiWa
>
> 为了让这个框架高性能，我希望能够做到整个处理期间 0 内存分配，从而避免因为 GC 导致的交易延迟不稳定。而类似 grpc 一类的框架封装太厚重，框架内就有大量的内存分配。于是我计划，在 protobuf 序列化和反序列化的阶段就开始自主开发。
> 同时，为了让 QiWa 这个框架更好用，我希望使用者定义好 proto 文件后，框架就生成好基本的处理代码，然后开发者只需要在已经反序列化好的回调函数中填充业务逻辑即可。
>
> 因此，为了配合 QiWa 项目，我以七娃的宝葫芦为名字，计划建设一个命令行代码生成工具：BaoHuLu.

## 目标

* 一个基于 golang 实现的命令行工具
* 根据 *.proto 文件，生成多种语言的数据序列化/反序列化的代码
  * 支持 protobuf 二进制格式的序列化和反序列
  * 支持 JSON 格式的序列化和反序列化
  * (在使用人数很多的情况下可以考虑) 支持 Thrift 二进制格式的序列化和反序列化
  * 考虑支持 FlatBuffers
* 为 QiWa RPC 框架，生成从请求处理 -> 反序列化 -> 触发回调 -> 对响应序列化 这个过程的脚手架代码

## 设计细节

* Proto 文件解析
  * 支持基本的 proto 语法，但是部分容易引起歧义的、性能不友好的、鸡肋的特性不会支持
    * 例如 required / oneof 等
* Proto 语法的扩展
  * 以注释的形式，支持扩展的语法
    * `// @method=GET/POST` 声明某个 service 的 method 对 http method 的支持
    * `// @jsonName=xx` 声明 message 中的某个 field 对应的 JSON key 名字是什么
  * 为了避免工具太复杂，一些经典的外部语法扩展先 **不支持**
    * Google 定义的扩展数据类型
    * 我最喜欢的 gogo proto 中的各种 extension
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
* Serivce 代码生成
  * 短期：仅针对 QiWa 项目生成 RPC 相关的代码
  * 后期：使用者可以自己以 golang template 的语法提供模板，然后生成代码。
  * 暂不支持 stream 关键字
* 开发语言支持：
  * golang
  * csharp
  * 使用者可以自己以 golang template 的语法提供模板，然后生成代码。
* 内存：
  * 重用友好：每个类型都提供 Reset() 方法，便于放到内存池中重用
  * 对于数据序列化，做到 0 内存分配：只需要不断把各个字段追加到目标缓冲区即可
  * 对于数据反序列化：
    * Golang: 保持对原始缓冲区的引用，然后使用 unsafe 代码直接指向原始缓冲区中的数据段。类似 string 类型做到不分配新的内存。
* 性能：
  * 通过减少内存分配 + 对象重用，努力避免 GC 和内存分配的开销
  * 数据序列化：始终采用追加到 buffer 的模式，容易做到极致的性能
  * 数据反序列化：
    * Protobuf: 在代码生成阶段，直接把 tag 和数据类型生成到解析代码中，无任何反射的开销。编译期决定了大多数事情，反序列化会精准的逐个读出 tag，并精准的放到确定的字段。(测试中发现，性能目前在所有 protobuf 的代码中，性能位于第一)
    * JSON: 使用流式解析，而不是 DOM 解析。(测试中发现，性能目前在所有 protobuf 的代码中，性能位于第一)
      * golang: 基于库 https://github.com/valyala/fastjson 来做流式解析
      * Csharp: 基于库 `System.Text.Json.Utf8JsonReader` 来做流式解析
* 其他:
  * 优化 struct 中的成员布局，GC 扫描友好，且节约内存



## 命令行

* 语法检查
  * `hulu xi ./xx.proto`

<img src="./doc/images/1.png" style="zoom:50%;" />



* 生成代码：

  * `hulu tu -src=./xx.proto -go_out=xx_dir -csharp_out=xx_dir _qiwa_out=xx_dir`



  ![](./doc/images/2.png)


## AI 使用声明

本项目 99% 以上的代码由 AI 生成。


## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2026 Fuchun Zhang
