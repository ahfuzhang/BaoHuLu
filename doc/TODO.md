* proto 文件中， import 一个 public.proto
  - 命令行参数支持 -include=
* csharp 的代码所依赖的公共库，放在哪儿？
  - 是否需要先把 QiWa.framework 先发布出去?   ✅
* 代码覆盖率测试
  - 生成 golang 测试代码 ✅
* benchmark 测试
  - 与各种现有工具的对比测试
  - protobuf 对比 ✅
  - csharp 对比 ✅
* 是否要支持批量输入多个 proto 文件?
  - 是否要支持按照文件夹输入?
* 各个语言的 namespace 如何处理？
  - golang 的 package 名字； golang 的 go.mod ✅
  - csharp 的 namespace  ✅
* 扩展语法
  - 提供文档   ✅
  - 进行支持   ✅
  - rpc 部分的扩展语法
* 性能测试报告  20%
* 安装文档
  - 打上合适的版本
* C# 如何拉取依赖的库?
  - 使用 git clone 的办法解决
  - NuGet 的模式解决    ✅
* golang
  - `var jsonParserPool fastjson.ParserPool`: 同个目录多个 proto 文件时，这里会出问题 ✅
    - 让用户传入 Parser 对象
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
* CSharp
  - 成员上加上 attribute，支持原生的 json 编解码 ✅
  - 生成 test ✅
  - 生成 benchmark ✅
* JSON
  - 数值类型，长度超过 53 bit 的问题
    - golang ✅
* 写文档说明特殊的处理逻辑:
  - bool 类型的 key
  - bytes 类型的支持
* linux + amd64 下的 benchmark
