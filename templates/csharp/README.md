# C# 数据模板包

`hulu -src=input.proto -template.dir=builtin:csharp -out=generated/csharp`。

`types/` 生成 enum、写/读 partial struct 及 Wrapper；`methods/` 按 From/To、Clone、Reset 机械拆分旧正文。字段映射、Wrapper、布局、测试样本等规则全部位于 `partials/`，核心没有 C# 前处理。

测试、benchmark 由 proto 文件级开关启用，各有按类型文件与按任务共享的验证/构造辅助文件。项目默认 net10.0，依赖版本沿用旧定义。`project/` 文件始终 create-only。

Benchmark 的 GrpcGen proto 副本与 namespace 改写也由普通模板完成，生成器不包含源码拼接。`QiWa.Common` 是数据运行库，不等于 RPC 框架依赖；基础包不生成网络代码。

AllTypes 的 372 个类型成员/方法语法 token 与旧版一致；主项目、测试和 benchmark 均编译。完整测试仍有 3 个历史失败，见 [验证记录](../../doc/GenerationValidation.md)。RPC 是独立的 [兼容包](QiWa.rpc/README.md)，尚未完成正文适配。
