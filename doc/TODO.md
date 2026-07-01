* proto 文件中， import 一个 public.proto
  - 命令行参数支持 -include=   => 计划不支持 ❌
  - 跨 namespace 引用类型，如何支持?
* csharp 的代码所依赖的公共库，放在哪儿？
  - 是否需要先把 QiWa.framework 先发布出去?   ✅
* 测试
  - 代码覆盖率测试
    - 生成 golang 测试代码 ✅
    - csharp 代码覆盖率
  - benchmark 测试
    - 与各种现有工具的对比测试
    - protobuf 对比 ✅
    - csharp 对比 ✅
    - linux + amd64 下的 benchmark ✅
  - golang
    - 运行一个 test 以后，自动按照我期望的格式输出 benchmark 的结果  ✅  
    - 对象生命周期的测试， 内存踩踏测试
  - csharp
    - 生成的 csharp 代码，用 lint 工具跑一遍
      - 用 lint 工具跑一遍
      - 把肯定不是错误的 lint 告警进行 disable
  - 进一步提高代码覆盖率
  - 努力达到 100% 的代码覆盖率  
  - 性能测试报告  50%
  - 选中更多的流行库来做性能对比测试
  - 二进制兼容性测试
    - 谷歌官方版本
  - 压测程序，增加 xx times / s  ✅
  - 对比的库：
    - https://github.com/goccy/go-json
    - simd json
* 是否要支持批量输入多个 proto 文件?
  - 是否要支持按照文件夹输入?   => 计划不支持 ❌
* 各个语言的 namespace 如何处理？
  - golang 的 package 名字； golang 的 go.mod ✅
  - csharp 的 namespace  ✅
* 扩展语法 Extensions
  - 提供文档   ✅
  - 进行支持   ✅
  - rpc 部分的扩展语法   ✅
  - 支持 decimal 数据类型 => 无意义 ❌ => 需要认真考虑，否则对金融领域的支持就会有限  ✅
  - @path 支持多次使用  ✅
  - @decimal=round:5   支持 decimal 数据类型，小数位数 5 位  ✅
  - rewrite datatype 选项：
    - 允许在编码阶段把 int64 类型改写为 fixed64，当值大于等于  1<<49，用 fixed64 存储，更加节约空间和性能
    - 后果：其他工具实现的编解码，无法与当前的二进制格式兼容。
    - 本质：通过牺牲二进制格式兼容性来换性能
  - protobuf 无法支持这样的类型：  `repeated map<string, string> data = 2;`   ✅
    - `// @AsMap`   ✅
  - protobuf 无法支持这样的类型： `map<string, repeated string> key_to_many_value = 3;`   ✅
    - 支持扩展：`// @AsArray`
  - json decode: 字符串会使用 fastjson 中的 cache: 这里必须拷贝出来。否则 Parser Reset() 后会出现问题。  
  - 生命周期导致的严重 bug:   ✅  => 通过引入 strArena 来解决，所有的字符串拷贝出来，放在一个 arena 之中
    - json decode 时:
      - json 未转义的字符串：可能因为 request buffer Reset() 而导致字符串被覆盖
      - json 已经转义的字符串：因为 parser Reset() 而导致字符串被覆盖
    - protobuf decode 时:
      - 因为字符串并未复制出来，request buffer 如果 reset，必然导致未来的字符串都被覆盖。 => 只要使用字符串，就会存在这个重大隐患

* golang
  - `var jsonParserPool fastjson.ParserPool`: 同个目录多个 proto 文件时，这里会出问题 ✅
    - 让用户传入 Parser 对象   ✅
  - 不符合 golang 命名规范   ✅

    ```go
    type Status int32

    const (
      STATUS_UNSPECIFIED Status = 0
      STATUS_ACTIVE Status = 1
      STATUS_DISABLED Status = 2
    )
    ```
  - ToJSON() 只序列化有效字段 ✅
  - Clone 方法中: 整数的数组类型，是否按照字节对齐了? ✅
  - 与 golang 1.26 的 json/v2 做性能对比 ✅
  - 把 FromProtobufVT() 做出来 ✅
    - 现在同时使用了两个版本，需要决定最终暴露哪个版本 ✅,  ToProtobuf() 方法，两个版本并存。简单类型使用 ToProtobufVT(), 复杂类型使用 ToProtobufAppend()
  - vtprotobuf
    - 增加单元测试
    - 增加对比测试
    - 增加 benchmark
  - 尝试做 PGO 优化  
  - map 的 value 是 bytes 的情况未处理 ✅
  - map 的 value 是 enum 的情况未处理 ✅
  - map 类型反序列化后，内存踩踏的问题
  - array 类型反序列化后，内存踩踏的问题
* CSharp
  - 成员上加上 attribute，支持原生的 json 编解码 ✅
  - 生成 test ✅
  - 生成 benchmark ✅
  - 是否加了足够多的 readonly ? ✅
  - 代码覆盖率是否足够?
  - 比较快的 protogen 这个工具的对比性能还没做
  - C# 如何拉取依赖的库?
    - 使用 git clone 的办法解决
    - NuGet 的模式解决    ✅
  - csharp rpc
    - 生成 QiWa 框架的 server 端代码  50%
    - 生成 QiWa 框架的 client 端的代码  => WIP
      - 是否需要 client context ?
    - 支持 rpc 上的扩展语法  ✅
    - Clone() 方法不够好：如何对象重用?  ✅
    - 支持 @path=/xx 的扩展  ✅
    - 生成的 Grpc.Tools 文件夹，影响编译
    - 支持代理模式，生成 DemoProxy.cs 文件
    - 支持获取原始请求内容  50%
  - 支持 form 提交的解析  ✅  => 还应该更深入的检查一下实现
  - 模仿 fastjson 实现 fastjson.cs  (实现 fastjson.cs，代替 json utf8 reader)
  - 支持 @AsMap
  - 支持 @AsArray

* JSON
  - 数值类型，长度超过 53 bit 的问题
    - golang ✅
* 写文档说明特殊的处理逻辑:
  - bool 类型的 key ✅
  - bytes 类型的支持 ✅
* 命令行支持传入模板文件，允许自定义的代码生成逻辑 ✅
* proto 文件
  - 递归定义的情况  ✅
  - message 作为各种子类型的情况  ✅
  - map 类型中的 value  是 message 类型，并且 message 存在递归调用的情况  ✅
* golang 性能优化
  - 生成汇编代码
  - 对于读(反序列化)：
    - 当 map key 为 bool 类型时，分配为长度固定为 2 的数组
    - 当 map 的 key 的数量小于 n 个时，所有的 key 放到一个数组中，使用顺序查找代替 hash 查找
      - ? api 该如何设计
      - 使用 simd 来加速查找
    - 使用 arena 内存分配
    - 所有的数组，都放到一大块 arena 当中
    - 实现一个 hashbrwon, 来实现只读 map 的高性能查询
  - 对于 json 反序列化
    - 实现 avx 版本的 fastjson
    - 实现无拷贝版本的 fastjson =>   ❌ 负优化，性能轻微下降
  - 对于 map 类型，要能够数出来 key 的数量
  - 对于 数组类型，要能够数出来 item 的数量
  - json 序列化：
    - 提前计算长度，然后使用数组从后往前赋值的方法来提升性能  =>  ❌ 负优化，无提升
  - FromProtobuf的性能只有 ToProtobuf 的一半不到。相比之下，pb 反序列化的性能优化值得做
  - 数组类型的优化:
    - `r._ValuesArr = append(r._ValuesArr, ReadonlyValueTypes{})`
    - 先计算出元素个数
    - 预先扩容到指定大小
    - 使用切片得到某个成员，而不是 append
* check 功能里：检查 key 不能有特殊字符  ✅
* 支持查看版本号
  - hulu --version  ✅
* 编程语言支持
  - 支持 js 的 pb 序列化/反序列化
  - 支持 typescript
* 格式支持
  - 考虑支持 yaml 格式
    - FromYAML() 时，对 map 使用 perfect hash，让读具备极高的性能
  - yaml 格式  ✅
    - 把 proto 文件中的注释，放到 yaml 中，增加 yaml 的可读性  ✅
    - csharp 代码增加 yaml 格式  ✅
    - golang: @yamlName 要能够影响生成的 yaml  ✅
    - 支持 @AsMap  ✅
    - 支持 @AsArray  ✅
  - url encode 格式      ✅
    - 对 form name 的支持，否则就会直接使用 proto 中的名字      ✅
* 宣传
  - 融入 buf 的生态中，否则很难做到高的使用人数
  - 融入 protoc 的工具链
  - 中文英文双语文档
* 问题
  - 一个 proto 文件，只能有一个 service

