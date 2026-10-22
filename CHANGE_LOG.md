* 2026-10-12（未发布，架构迁移中）：
  - 统一为 `hulu -src -template.dir -out`，生成前自动 check；测试/benchmark/项目改为 proto 文件级开关。
  - 中立 Schema、Service/RPC 引用、通用模板包引擎与受保护输出；Go/C# 规则迁入根目录模板。
  - 类型与方法按计划拆分，移除旧语言生成器及配置草案，运行库路径不变。
  - AllTypes 新旧 API/正文、双向协议/官方 Protobuf、Clone/Reset 及 Go 性能对比见 `doc/GenerationValidation.md`。
  - RPC 业务适配、Go repeated enum 等仍需 D16 review；保留并记录旧测试失败，不宣称全部验收完成。

* 2026-07-22:
  - csharp:
    - bug fix: 解决 url values 中因为使用 JsonSerializer.Serialize 而带来的 AOT 编译告警。
* 2026-07-15: v0.16.0
  - csharp:
    - bug fix: 当某个 message 中包含了不在当前 proto 文件中的 message 时，程序会生成一个空的 message 的处理代码。
    - url values 编码中：父 message 定义了 @UrlValues，但是子 message 未加上 @UrlValues，这时候这个属性应该是传染性的。

* 2026-07-12: v0.15.0
  - golang:
    - bug fix: FromJSON() 中，对 null 的处理有错
    - bug fix: FromYAML() 中，对行内的尾部注释的处理有错
  - csharp:
    - bug fix: FromJSON() 对 @AsArray 的两层嵌套的处理有错
    - bug fix: ToYAML() 对 @AsArray 的两层嵌套的处理有错
